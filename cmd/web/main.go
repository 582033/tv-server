package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"tv-server/internal/model"
	"tv-server/internal/router"
	"tv-server/utils/cache"
	"tv-server/utils/core"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("c", "config/dev.json", "配置文件路径")
	flag.Parse()

	// 加载配置文件
	if err := core.LoadConfig(*configPath); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 获取配置
	cfg := core.GetConfig()

	// 初始化数据库连接
	if err := model.InitDB(cfg.DB.Type); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer model.CloseDB()

	// 初始化缓存目录
	if err := cache.Init(); err != nil {
		log.Fatalf("初始化缓存失败: %v", err)
	}

	// 初始化路由
	r := router.NewRouter()

	// 启动服务器
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Printf("服务器启动在 http://0.0.0.0%s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")
}
