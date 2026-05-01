package router

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const headerRequestID = "X-Request-ID"

func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// RequestID 注入请求 ID，便于日志关联。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = randomID()
			c.Writer.Header().Set(headerRequestID, rid)
		}
		c.Set("request_id", rid)
		c.Next()
	}
}

// ZapRecovery 捕获 panic 并写入 Zap，返回 JSON 错误。
func ZapRecovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered",
					zap.Any("error", err),
					zap.String("request_id", c.GetString("request_id")),
					zap.String("stack", string(debug.Stack())),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "服务器内部错误",
				})
			}
		}()
		c.Next()
	}
}

// ErrorJSON 将 handler 中 c.Error 链上的最后一个错误转为 JSON（骨架）。
func ErrorJSON(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last()
		log.Warn("handler error",
			zap.String("request_id", c.GetString("request_id")),
			zap.String("path", c.FullPath()),
			zap.Error(err),
		)
		if c.Writer.Written() {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "REQUEST_ERROR",
			"message": err.Error(),
		})
	}
}
