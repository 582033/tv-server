// internal/router/router_reg.go
package router

import (
	// 导入 static 包
	"net/http"
	"tv-server/internal/assets"
	"tv-server/internal/handler"
	"tv-server/utils/core"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// 设置为发布模式
	gin.SetMode(gin.ReleaseMode)

	//注册中间件
	r.Use(core.Middleware())

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
	r.GET(URLHome, core.WrapHandler(handler.HandleHome))
	r.GET(URLWelcome, core.WrapHandler(handler.HandleHome))
	r.GET(URLCategory, core.WrapHandler(handler.HandleChannelPage))
}

// 注册 API 路由
func registerAPI(r *gin.Engine) {
	r.GET(URLIPTV, core.WrapHandler(handler.HandleM3U))
	r.POST(URLValidate, core.WrapHandler(handler.HandleValidate))
	r.POST(URLUpload, core.WrapHandler(handler.HandleUpload))
	r.GET(URLProcess, core.WrapHandler(handler.HandleProcess))
	r.GET(URLChannels, core.WrapHandler(handler.ListAllChannel))
	r.GET(URLChannelRecordNum, core.WrapHandler(handler.GetRecordNums))
	r.POST(URLChannelValidate, core.WrapHandler(handler.HandleChannelValidate))
}
