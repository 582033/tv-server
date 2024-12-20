package handler

import (
	"io"
	"net/http"
	"os"

	"tv-server/internal/logic/m3u"
	"tv-server/utils/cache"

	"github.com/gin-gonic/gin"
)

type ValidateRequest struct {
	URL string `json:"url"`
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

	// 获取内容
	content, err := fetchContent(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ValidateResponse{
			Success: false,
			Message: "Error fetching content: " + err.Error(),
		})
		return
	}

	// 解析并验证
	entries := m3u.Parse(content)
	validEntries := m3u.ValidateURLs(entries)

	// 写入缓存文件
	if err := m3u.WriteToFile(validEntries, cache.CacheFile); err != nil {
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
		c.String(http.StatusNotFound, "No M3U file available. Please validate a M3U URL first.")
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
