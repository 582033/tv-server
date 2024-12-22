package main

import (
	"log"
	"os"
	"time"

	"tv-server/internal/router"
	"tv-server/utils/cache"
)

func main() {
	// 初始化缓存目录
	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to create cache directory: %v", err)
	}

	// 启动定期清理缓存的goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
		defer ticker.Stop()

		for range ticker.C {
			cache.Cleanup()
		}
	}()

	// 输出缓存文件位置
	log.Printf("Cache file location: %s", cache.CacheFile)

	// 添加调试信息
	log.Printf("Working directory: %s", getCurrentDirectory())

	// 设置并启动路由
	r := router.NewRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return err.Error()
	}
	return dir
}
