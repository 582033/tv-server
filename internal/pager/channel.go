package pager

import (
	"tv-server/utils/core"

	"github.com/gin-gonic/gin"
)

// PageChannel 处理频道分类页面
func PageChannel(c *core.Context) {
	c.WebRender("template/channels.html", gin.H{
		"title":  "频道分类",
		"active": "category",
	}, nil)
}
