package handler

import (
	"net/http"

	"tv-server/utils/core"

	"github.com/gin-gonic/gin"
)

func HandleHome(c *core.Context) {
	c.HTML(http.StatusOK, "template/index.html", gin.H{
		"title":  "IPTV 服务器",
		"active": "home",
	})
}
