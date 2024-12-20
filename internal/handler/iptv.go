package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tv-server/internal/logic/m3u"
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
}

type UploadResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Token    string `json:"token"`
	FileName string `json:"fileName"`
}

var (
	tempFiles = struct {
		sync.RWMutex
		files map[string]string
	}{
		files: make(map[string]string),
	}
)

// 处理验证请求
func HandleValidate(c *gin.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ValidateResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	var allEntries []m3u.Entry

	// 处理上传的文件
	if req.Token != "" {
		tempFiles.RLock()
		tempFile, exists := tempFiles.files[req.Token]
		tempFiles.RUnlock()

		if exists {
			// 读取临时文件内容
			content, err := os.ReadFile(tempFile)
			if err == nil {
				entries := m3u.Parse(string(content))
				allEntries = append(allEntries, entries...)
			}

			// 处理完后删除临时文件
			tempFiles.Lock()
			delete(tempFiles.files, req.Token)
			tempFiles.Unlock()
			os.Remove(tempFile)
		}
	}

	// 处理URL列表
	for _, url := range req.URLs {
		allEntries = append(allEntries, m3u.Entry{URL: url})
	}

	// 验证所有链接
	validEntries := m3u.ValidateURLsWithLatency(allEntries, req.MaxLatency)

	// 更新缓存文件
	if err := m3u.WriteToFile(validEntries, cache.CacheFile); err != nil {
		c.JSON(http.StatusInternalServerError, ValidateResponse{
			Success: false,
			Message: "Error writing cache",
		})
		return
	}

	c.JSON(http.StatusOK, ValidateResponse{
		Success: true,
		Message: "Validation successful",
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

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "m3u-*"+filepath.Ext(file.Filename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "创建临时文件失败",
		})
		return
	}

	// 保存上传的文件
	if err := c.SaveUploadedFile(file, tempFile.Name()); err != nil {
		os.Remove(tempFile.Name())
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Message: "保存文件失败",
		})
		return
	}

	// 生成随机token
	token := uuid.New().String()

	// 保存临时文件信息
	tempFiles.Lock()
	tempFiles.files[token] = tempFile.Name()
	tempFiles.Unlock()

	// 设置定时清理（比如1小时后）
	go func() {
		time.Sleep(1 * time.Hour)
		tempFiles.Lock()
		if path, exists := tempFiles.files[token]; exists {
			delete(tempFiles.files, token)
			os.Remove(path)
		}
		tempFiles.Unlock()
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
