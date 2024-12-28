package core

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
)

// Context 自定义上下文，扩展gin.Context
type Context struct {
	*gin.Context
	StdCtx context.Context
}

// NewContext 创建基础上下文
func NewContext() *Context {
	return &Context{
		Context: &gin.Context{},
		StdCtx:  context.Background(),
	}
}

// GetRequestID 获取请求ID
func (c *Context) GetRequestID() string {
	if id, exists := c.Get("X-Request-ID"); exists {
		return id.(string)
	}
	return ""
}

// Middleware Gin中间件，为每个请求创建上下文
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID
		requestID := uuid.New().String()

		// 创建自定义上下文
		ctx := &Context{
			Context: c,
			StdCtx:  context.WithValue(c.Request.Context(), RequestIDKey, requestID),
		}

		// 设置请求ID到header
		c.Request.Header.Set("X-Request-ID", requestID)

		// 将上下文保存到gin的上下文中
		c.Set("context", ctx)

		c.Next()
	}
}

// FromGinContext 从gin.Context中获取自定义上下文
func FromGinContext(c *gin.Context) *Context {
	if ctx, exists := c.Get("context"); exists {
		if customCtx, ok := ctx.(*Context); ok {
			return customCtx
		}
	}
	return nil
}

// WrapHandler 包装处理函数，将core.Context转换为gin.HandlerFunc
func WrapHandler(handler func(*Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := FromGinContext(c)
		if ctx == nil {
			ctx = &Context{
				Context: c,
				StdCtx:  c.Request.Context(),
			}
		}
		handler(ctx)
	}
}
