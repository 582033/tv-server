package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleHome(c *gin.Context) {
	c.HTML(http.StatusOK, "static/index.html", gin.H{
		"title": "TV Server",
	})
}
