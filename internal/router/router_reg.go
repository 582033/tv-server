// internal/router/router_reg.go
package router

import (
	// 导入 static 包
	"net/http"
	"tv-server/internal/assets"
	"tv-server/internal/handler"
	"tv-server/internal/pager"
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
	r.GET(URLHome, core.WrapHandler(pager.PageHome))
	r.GET(URLWelcome, core.WrapHandler(pager.PageHome))
	r.GET(URLChannel, core.WrapHandler(pager.PageChannel))
	r.GET(URLChannelDetail, core.WrapHandler(pager.PageChannelDetail))
}

// 注册 API 路由
func registerAPI(r *gin.Engine) {
	r.GET(URLAPIIPTV, core.WrapHandler(handler.HandleM3U))
	r.POST(URLAPIValidate, core.WrapHandler(handler.HandleValidate))
	r.POST(URLAPIUpload, core.WrapHandler(handler.HandleUpload))
	r.GET(URLAPIProcess, core.WrapHandler(handler.HandleProcess))
	r.GET(URLAPIChannels, core.WrapHandler(handler.HandleListAllChannel))
	r.GET(URLAPIChannelRecordNum, core.WrapHandler(handler.HandleGetRecordNums))
	r.POST(URLAPIChannelValidate, core.WrapHandler(handler.HandleChannelValidate))
	r.GET(URLAPIChannelDetail, core.WrapHandler(handler.HandleChannelDetail))
}
