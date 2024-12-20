package handler

import (
	"fmt"
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
		if entries, err := m3u.ParseFile(filepath.Join(cache.CacheDir, req.Token)); err == nil {
			allEntries = append(allEntries, entries...)
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

	// 使用带缓冲的通道进行并发控制
	workers := 50
	tasks := make(chan m3u.Entry, len(allEntries))
	results := make(chan m3u.Entry, len(allEntries))
	done := make(chan bool)

	// 启动工作协程
	for i := 0; i < workers; i++ {
		go func() {
			for entry := range tasks {
				if m3u.ValidateURL(entry.URL, req.MaxLatency) {
					results <- entry
				}
			}
		}()
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
