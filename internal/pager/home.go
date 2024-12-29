package pager

import (
	"tv-server/utils/core"

	"github.com/gin-gonic/gin"
)

func PageHome(c *core.Context) {
	c.WebRender("template/index.html", gin.H{
		"title":  "IPTV 服务器",
		"active": "home",
	}, nil)
}
