package main

import (
	"io"
	"net/http"
	"sync"
	"time"

	"tv-server/internal/parser"
	"tv-server/internal/validator"
	"tv-server/internal/writer"

	"github.com/gin-gonic/gin"
)

var (
	lastValidation time.Time
	cacheMutex     sync.Mutex
	cacheFile      = "validated.m3u"
)

func main() {
	r := gin.Default()

	// IPTV route
	r.GET("/iptv.m3u", handleIPTV)

	r.Run(":8080")
}

func handleIPTV(c *gin.Context) {
	m3uURL := c.Query("url")
	if m3uURL == "" {
		c.String(http.StatusBadRequest, "Missing m3u url parameter")
		return
	}

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
	if err := writer.WriteM3U(validEntries, cacheFile); err != nil {
		c.String(http.StatusInternalServerError, "Error writing cache: %v", err)
		return
	}

	// 返回验证后的内容
	c.File(cacheFile)
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
