package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tv-server/internal/logic/m3u"
	"tv-server/internal/model"
	"tv-server/internal/model/types"
	"tv-server/utils/cache"
	"tv-server/utils/core"
	"tv-server/utils/msg"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ValidateRequest struct {
	URLs       []string `json:"urls"`
	MaxLatency int      `json:"maxLatency"`
	Token      string   `json:"token"`
}

type ValidateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Stats   struct {
		Total  int `json:"total"`  // 原始链接数
		Unique int `json:"unique"` // 去重后数量
		Valid  int `json:"valid"`  // 有效链接数
	} `json:"stats"`
	M3ULink string `json:"m3uLink"` // M3U 文件链接
}

type UploadResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Token    string `json:"token"`
	FileName string `json:"fileName"`
}

// 定义结构体
type ChannelInfo struct {
	Metadata string `json:"Metadata"`
	URL      string `json:"URL"`
}

var (
	tempFiles = struct {
		sync.RWMutex
		files map[string]string
	}{
		files: make(map[string]string),
	}
)

// HandleProcess 获取验证进度
// ** 需要先触发验证，在请求进度才能从0开始，否则可能获取到的进度是100
func HandleProcess(c *core.Context) {
	c.WebResponse(msg.CodeOK, gin.H{
		"success": true,
		"message": "获取进度成功",
		"process": m3u.GetProcess(),
	}, nil)
}

// SaveValidatedEntries 保存验证后的条目到缓存文件
func SaveValidatedEntries(entries []m3u.Entry) error {
	if len(entries) == 0 {
		return nil
	}

	tempFile := cache.CacheFile + ".temp"
	if err := m3u.WriteToFile(entries, tempFile); err != nil {
		return fmt.Errorf("写入缓存失败: %v", err)
	}

	if err := os.Rename(tempFile, cache.CacheFile); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("更新缓存文件失败: %v", err)
	}

	return nil
}

// HandleValidate 处理验证请求
func HandleValidate(c *core.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.WebResponse(msg.CodeBadRequest, nil, err)
		return
	}

	// 获取所有链接
	var allEntries []m3u.Entry

	// 处理上传的文件和URLs
	if req.Token != "" {
		filePath := filepath.Join(cache.CacheDir, req.Token)
		fmt.Printf("处理上传的文件，Token: %s, 文件路径: %s\n", req.Token, filePath)

		// 确保缓存目录存在
		if err := os.MkdirAll(cache.CacheDir, 0755); err != nil {
			fmt.Printf("创建缓存目录失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, ValidateResponse{
				Success: false,
				Message: "系统错误：无法访问缓存目录",
			})
			return
		}

		if _, err := os.Stat(filePath); err != nil {
			fmt.Printf("文件不存在或无法访问: %v\n", err)
		} else {
			entries, err := m3u.ParseFile(filePath)
			if err != nil {
				fmt.Printf("解析文件失败: %v\n", err)
			} else {
				fmt.Printf("成功解析文件，获取到 %d 个条目\n", len(entries))
				allEntries = append(allEntries, entries...)
			}
		}
	}
	for _, url := range req.URLs {
		if entries, err := m3u.ParseURL(url); err == nil {
			allEntries = append(allEntries, entries...)
		}
	}

	// 验证链接
	fmt.Printf("开始验证 %d 个链接\n", len(allEntries))

	//将allEntries写入mongodb
	if err := saveEntries(c, allEntries); err != nil {
		fmt.Printf("写入MongoDB失败: %v\n", err)
	}

	//req.MaxLatency单位是ms
	maxLatency := time.Duration(req.MaxLatency) * time.Millisecond
	//开始验证并去重
	validEntries, finalValidEntries, err := m3u.ValidateAndUnique(allEntries, maxLatency, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ValidateResponse{
			Success: false,
			Message: "验证失败",
		})
		return
	}

	// 使用新的公共函数
	if err := SaveValidatedEntries(finalValidEntries); err != nil {
		c.JSON(http.StatusInternalServerError, ValidateResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, ValidateResponse{
		Success: true,
		Message: "验证完成！",
		Stats: struct {
			Total  int `json:"total"`
			Unique int `json:"unique"`
			Valid  int `json:"valid"`
		}{
			Total:  len(allEntries),
			Unique: len(validEntries),
			Valid:  len(finalValidEntries),
		},
		M3ULink: fmt.Sprintf("http://%s/iptv.m3u", c.Request.Host),
	})
}

// 返回缓存的M3U文件
func HandleM3U(c *core.Context) {
	if _, err := os.Stat(cache.CacheFile); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "No M3U file available. Please validate M3U URLs first.")
		return
	}

	c.Header("Content-Type", "application/x-mpegurl")
	c.Header("Content-Disposition", "inline")
	c.File(cache.CacheFile)
}

// HandleUpload 处理文件上传
func HandleUpload(c *core.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "无效的文件上传",
		})
		return
	}

	// 确保缓存目录存在
	if err := os.MkdirAll(cache.CacheDir, 0755); err != nil {
		fmt.Printf("创建缓存目录失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "创建缓存目录失败",
		})
		return
	}

	token := uuid.New().String()
	tempFilePath := filepath.Join(cache.CacheDir, token)

	fmt.Printf("准备保存文件到: %s\n", tempFilePath)

	// 保存上传的文件
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		fmt.Printf("保存文件失败: %v\n", err)
		os.Remove(tempFilePath)
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "保存文件失败",
		})
		return
	}

	fmt.Printf("文件已成功保存到: %s\n", tempFilePath)

	// 设置定时清理（比如1小时后）
	go func() {
		time.Sleep(1 * time.Hour)
		if err := os.Remove(tempFilePath); err != nil {
			fmt.Printf("清理文件失败: %v\n", err)
		}
	}()

	c.JSON(http.StatusOK, UploadResponse{
		Success:  true,
		Message:  "文件上传成功",
		Token:    token,
		FileName: file.Filename,
	})
}

func saveEntries(ctx *core.Context, entries []m3u.Entry) error {
	parsedEntries := m3u.ParseEntry(entries)
	msList := make([]*types.MediaStream, 0, len(entries))

	for _, parsedEntry := range parsedEntries {
		ms := &types.MediaStream{
			StreamName:  parsedEntry.Title,
			ChannelName: parsedEntry.Channel,
			StreamUrl:   []string{parsedEntry.URL},
			StreamLogo:  parsedEntry.Logo,
		}
		msList = append(msList, ms)
	}

	db := model.GetDB()
	return db.M3U().BatchSave(ctx, msList)
}
