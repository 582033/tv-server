package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestList(t *testing.T) {
	List(&gin.Context{})
}

func TestListAllChannel(t *testing.T) {
	ListAllChannel(&gin.Context{})
}