package router

// URL 路径常量定义
const (
	// 页面路由
	URLHome          = "/"
	URLWelcome       = "/home"
	URLChannel       = "/channel"
	URLChannelDetail = "/channel/:channel_name"

	// API 路由
	URLAPIIPTV             = "/iptv.m3u"
	URLAPIValidate         = "/api/validate"
	URLAPIUpload           = "/api/upload"
	URLAPIProcess          = "/api/process"
	URLAPIChannels         = "/api/channels"
	URLAPIChannelRecordNum = "/api/channel/get_record_num"
	URLAPIChannelValidate  = "/api/channel/validate"
	URLAPIChannelDetail    = "/api/channel/detail"

	// 其他路由分类可以在这里继续添加
	// 例如：
	// URLAdmin  = "/admin"
	// URLUser   = "/user"
	// URLSystem = "/system"
)
