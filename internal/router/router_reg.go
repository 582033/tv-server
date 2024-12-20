package router

import (
	"log"
	"os"

	"tv-server/internal/handler"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	log.Printf("Current working directory: %s", workDir)

	// 设置为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 使用项目根目录的 static 文件夹
	r.Static("/static", "static")
	r.LoadHTMLFiles("static/index.html")

	// 注册路由
	registerPages(r)
	registerAPI(r)

	return r
}

// 注册页面路由
func registerPages(r *gin.Engine) {
	r.GET(URLHome, handler.HandleHome)
	r.GET(URLWelcome, handler.HandleHome)
}

// 注册 API 路由
func registerAPI(r *gin.Engine) {
	r.GET(URLIPTV, handler.HandleM3U)
	r.POST(URLValidate, handler.HandleValidate)
	r.POST(URLUpload, handler.HandleUpload)
}
