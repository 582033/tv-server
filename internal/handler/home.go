package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleHome(c *gin.Context) {
	c.HTML(http.StatusOK, "template/index.html", gin.H{
		"title":  "IPTV 服务器",
		"active": "home",
	})
}
