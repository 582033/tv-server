package main

import (
	"flag"
	"fmt"
	"log"

	"tv-server/internal/router"
	"tv-server/utils/cache"
	"tv-server/utils/core"
)

func main() {
	// 定义命令行参数
	configPath := flag.String("c", "config/dev.json", "配置文件路径")
	flag.Parse()

	// 加载配置
	if err := core.LoadConfig(*configPath); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 获取配置
	cfg := core.GetConfig()

	// 初始化缓存目录
	if err := cache.Init(); err != nil {
		log.Fatalf("初始化缓存失败: %v", err)
	}

	// 创建路由
	r := router.NewRouter()

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("服务器启动在 http://0.0.0.0%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
