// internal/router/router_reg.go
package router

import (

	// 导入 static 包
	"net/http"
	"tv-server/internal/assets"
	"tv-server/internal/handler"
	"tv-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// 设置为发布模式
	gin.SetMode(gin.ReleaseMode)

	//注册中间件
	r.Use(middleware.WithContext())

	//加载静态文件
	r.StaticFS("/static", http.FS(assets.StaticFS))
	//加载模板
	assets.LoadHTMLFromEmbedFS(r, assets.TemplateFS, "template/*.html")

	// 注册路由
	registerPages(r)
	registerAPI(r)

	return r
}

// 注册页面路由
func registerPages(r *gin.Engine) {
	r.GET(URLHome, handler.HandleHome)
	r.GET(URLWelcome, handler.HandleHome)
	r.GET(URLCategory, handler.HandleChannelPage)
}

// 注册 API 路由
func registerAPI(r *gin.Engine) {
	r.GET(URLIPTV, handler.HandleM3U)
	r.POST(URLValidate, handler.HandleValidate)
	r.POST(URLUpload, handler.HandleUpload)
	r.GET(URLProcess, handler.HandleProcess)
	r.GET(URLChannels, handler.ListAllChannel)
	r.GET(URLChannelRecordNum, handler.GetRecordNums)
	r.POST(URLChannelValidate, handler.HandleChannelValidate)
}
