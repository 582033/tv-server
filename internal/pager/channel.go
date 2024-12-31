package pager

import (
	"tv-server/utils/core"

	"github.com/gin-gonic/gin"
)

// PageChannel 处理频道分类页面
func PageChannel(c *core.Context) {
	c.WebRender("template/channels.html", gin.H{
		"title": "频道分类",
	}, nil)
}

func PageChannelDetail(c *core.Context) {
	channelName := c.Param("channel_name")
	c.WebRender("template/channel_detail.html", gin.H{
		"title":       "频道详情",
		"channelName": channelName,
		"channelUrl":  "/channel",
	}, nil)
}
