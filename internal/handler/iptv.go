package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tv-server/internal/logic/m3u"
	db "tv-server/internal/model/mongodb"
	"tv-server/utils/cache"

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

// HandleValidate 处理验证请求
func HandleValidate(c *gin.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ValidateResponse{
			Success: false,
			Message: "Invalid request",
		})
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
	validEntries := make([]m3u.Entry, 0)

	//将allEntries写入mongodb
	if err := saveEntries(c, allEntries); err != nil {
		fmt.Printf("写入MongoDB失败: %v\n", err)
	}

	// 使用带缓冲的通道进行并发控制
	workers := 50
	tasks := make(chan m3u.Entry, len(allEntries))
	results := make(chan m3u.Entry, len(allEntries))
	done := make(chan bool)

	// 启动工作协程
	for i := 0; i < workers; i++ {
		go func(entryChan <-chan m3u.Entry, resultChan chan<- m3u.Entry, maxLatency int) {
			for entry := range entryChan {
				valid, err := m3u.ValidateURL(entry.URL, maxLatency)
				if valid && err == nil {
					resultChan <- entry
				}
			}
		}(tasks, results, req.MaxLatency)
	}

	// 发送任务
	go func() {
		for _, entry := range allEntries {
			tasks <- entry
		}
		close(tasks)
	}()

	// 收集结果
	go func() {
		for entry := range results {
			validEntries = append(validEntries, entry)
		}
		done <- true
	}()

	// 设置超时控制
	timeout := time.After(time.Duration(req.MaxLatency*len(allEntries)/workers) * time.Millisecond)

	var finalValidEntries []m3u.Entry
	select {
	case <-done:
		// 在验证完成后进行去重
		urlMap := make(map[string]m3u.Entry)
		for _, entry := range validEntries {
			urlMap[entry.URL] = entry
		}
		for _, entry := range urlMap {
			finalValidEntries = append(finalValidEntries, entry)
		}
		fmt.Printf("验证完成，原始链接 %d 个，验证通过 %d 个，去重后有效链接 %d 个\n",
			len(allEntries), len(validEntries), len(finalValidEntries))

	case <-timeout:
		// 超时时也进行去重
		urlMap := make(map[string]m3u.Entry)
		for _, entry := range validEntries {
			urlMap[entry.URL] = entry
		}
		for _, entry := range urlMap {
			finalValidEntries = append(finalValidEntries, entry)
		}
		fmt.Printf("验证超时，原始链接 %d 个，验证通过 %d 个，去重后有效链接 %d 个\n",
			len(allEntries), len(validEntries), len(finalValidEntries))
	}

	// 写入缓存文件
	if len(finalValidEntries) > 0 {
		tempFile := cache.CacheFile + ".temp"
		if err := m3u.WriteToFile(finalValidEntries, tempFile); err != nil {
			c.JSON(http.StatusInternalServerError, ValidateResponse{
				Success: false,
				Message: "写入缓存失败",
			})
			return
		}

		if err := os.Rename(tempFile, cache.CacheFile); err != nil {
			os.Remove(tempFile)
			c.JSON(http.StatusInternalServerError, ValidateResponse{
				Success: false,
				Message: "更新缓存文件失败",
			})
			return
		}
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
func HandleM3U(c *gin.Context) {
	if _, err := os.Stat(cache.CacheFile); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "No M3U file available. Please validate M3U URLs first.")
		return
	}

	c.Header("Content-Type", "application/x-mpegurl")
	c.Header("Content-Disposition", "inline")
	c.File(cache.CacheFile)
}

func fetchContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// HandleUpload 处理文件上传
func HandleUpload(c *gin.Context) {
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

// 添加新的路由处理函数
func RegisterRoutes(r *gin.Engine) {
	r.GET("/iptv.m3u", HandleM3U)
	r.POST("/api/validate", HandleValidate)
	r.POST("/api/upload", HandleUpload) // 添加上传路由
}

func saveEntries(c *gin.Context, entries []m3u.Entry) error {
	if debugBytes, _ := json.Marshal(entries); len(debugBytes) > 0 {
		fmt.Printf("RequestID:%v DebugMessage:%s Value:%s", nil, "entries", string(debugBytes))
	}
	//问题出在这里!!
	parsedEntries := m3u.ParseEntry(entries)
	msList := make([]*db.MediaStream, 0, len(entries))

	if debugBytes, _ := json.Marshal(msList); len(debugBytes) > 0 {
		fmt.Printf("RequestID:%v DebugMessage:%s Value:%s", nil, "msList", string(debugBytes))
	}
	//todo 这里的数据，如果StreamName和ChannelName在数据库中，就更新，不在，就插入

	for _, parsedEntry := range parsedEntries {
		ms := &db.MediaStream{
			StreamName:  parsedEntry.Title,
			ChannelName: parsedEntry.Channel,
			StreamUrl:   []string{parsedEntry.URL},
			StreamLogo:  parsedEntry.Logo,
		}
		msList = append(msList, ms)
	}
	return db.BatchSave(c, msList)
}
