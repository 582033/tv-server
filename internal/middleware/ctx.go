package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func WithContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background() // 或者使用 context.WithCancel, context.WithValue 等
		// 你可以在这里添加一些初始化的上下文值，例如请求 ID，用户 ID 等
		c.Request = c.Request.WithContext(ctx)

		// 使用uuid生成一个request_id
		requestID := uuid.New().String()
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}
