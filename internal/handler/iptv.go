package handler

import (
	"io"
	"net/http"
	"os"
	"sync"

	"tv-server/internal/logic/m3u"
	"tv-server/utils/cache"

	"github.com/gin-gonic/gin"
)

type ValidateRequest struct {
	URLs []string `json:"urls"`
}

type ValidateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

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

	if len(req.URLs) == 0 {
		c.JSON(http.StatusBadRequest, ValidateResponse{
			Success: false,
			Message: "No URLs provided",
		})
		return
	}

	// 并发获取所有M3U内容
	var wg sync.WaitGroup
	var mu sync.Mutex
	allEntries := make([]m3u.Entry, 0)

	for _, url := range req.URLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			content, err := fetchContent(url)
			if err != nil {
				return // 跳过错误的URL
			}

			entries := m3u.Parse(content)
			validEntries := m3u.ValidateURLs(entries)

			mu.Lock()
			allEntries = append(allEntries, validEntries...)
			mu.Unlock()
		}(url)
	}

	wg.Wait()

	if len(allEntries) == 0 {
		c.JSON(http.StatusBadRequest, ValidateResponse{
			Success: false,
			Message: "No valid entries found",
		})
		return
	}

	// 写入合并后的缓存文件
	if err := m3u.WriteToFile(allEntries, cache.CacheFile); err != nil {
		c.JSON(http.StatusInternalServerError, ValidateResponse{
			Success: false,
			Message: "Error writing cache: " + err.Error(),
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
