package router

// URL 路径常量定义
const (
	// 页面路由
	URLHome     = "/"
	URLWelcome  = "/home"
	URLCategory = "/category"

	// API 路由
	URLIPTV             = "/iptv.m3u"
	URLValidate         = "/api/validate"
	URLUpload           = "/api/upload"
	URLProcess          = "/api/process"
	URLChannels         = "/api/channels"
	URLChannelRecordNum = "/api/channel/get_record_num"
	URLChannelValidate  = "/api/channel/validate"

	// 其他路由分类可以在这里继续添加
	// 例如：
	// URLAdmin  = "/admin"
	// URLUser   = "/user"
	// URLSystem = "/system"
)
