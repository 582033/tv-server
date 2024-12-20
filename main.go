package main

import (
	"log"

	"tv-server/internal/cache"
	"tv-server/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化缓存目录
	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to create cache directory: %v", err)
	}

	// 输出缓存文件位置
	log.Printf("Cache file location: %s", cache.CacheFile)

	r := gin.Default()
	r.GET("/iptv.m3u", handler.HandleIPTV)
	r.Run(":8080")
}
