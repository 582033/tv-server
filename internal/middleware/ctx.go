package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 自定义类型，嵌入 gin.Context
type SrContext struct {
	*gin.Context
}

// 实现 RequestIDGetter 接口
func (c *SrContext) GetRequestID() string {
	// 假设请求的唯一标识符是 X-Request-ID
	return c.Request.Header.Get("X-Request-ID")
}

func WithContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("context", &SrContext{c})
		// 设置请求唯一标识符
		c.Request.Header.Set("X-Request-ID", uuid.New().String())
		c.Next()
	}
}
