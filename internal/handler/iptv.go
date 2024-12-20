package handler

import (
	"io"
	"log"
	"net/http"
	"os"

	"tv-server/internal/cache"
	"tv-server/internal/parser"
	"tv-server/internal/validator"
	"tv-server/internal/writer"

	"github.com/gin-gonic/gin"
)

func HandleIPTV(c *gin.Context) {
	cache.CacheMutex.Lock()
	defer cache.CacheMutex.Unlock()

	m3uURL := c.Query("url")
	if m3uURL == "" {
		c.String(http.StatusBadRequest, "Missing m3u url parameter")
		return
	}

	log.Printf("Processing M3U URL: %s", m3uURL)

	// 清理缓存
	cache.Cleanup()

	// 获取内容
	content, err := fetchContent(m3uURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching content: %v", err)
		return
	}

	// 解析并验证
	entries := parser.ParseM3U(content)
	validEntries := validator.ValidateURLs(entries)

	// 写入缓存文件
	if err := writer.WriteM3U(validEntries, cache.CacheFile); err != nil {
		c.String(http.StatusInternalServerError, "Error writing cache: %v", err)
		return
	}

	// 检查缓存文件是否存在
	if _, err := os.Stat(cache.CacheFile); err != nil {
		c.String(http.StatusInternalServerError, "Cache file not found")
		return
	}

	// 设置header
	c.Header("Content-Type", "application/x-mpegurl")
	c.Header("Content-Disposition", "inline")

	// 从缓存目录读取并返回文件
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
