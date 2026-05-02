package router

import (
	"net/http"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/handler"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BootstrapGate 在未完成首次初始化时拦截非白名单请求。
func BootstrapGate(db *gorm.DB, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.Next()
			return
		}
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if bootstrapPathAllowlisted(c.Request.Method, path) {
			c.Next()
			return
		}
		ok, err := repository.IsSystemInitialized(db)
		if err != nil {
			if log != nil {
				log.Warn("bootstrap gate: read state failed", zap.Error(err))
			}
			handler.WriteGatewayError(c, http.StatusServiceUnavailable, "BOOTSTRAP_STATE_UNAVAILABLE", "无法读取初始化状态")
			c.Abort()
			return
		}
		if ok {
			c.Next()
			return
		}
		handler.WriteGatewayError(c, http.StatusForbidden, "BOOTSTRAP_REQUIRED", "系统尚未完成首次初始化")
		c.Abort()
	}
}

func bootstrapPathAllowlisted(method, path string) bool {
	switch method {
	case http.MethodGet:
		if path == "/health" {
			return true
		}
		if path == "/api/bootstrap/v1/status" {
			return true
		}
		if path == "/api/admin/v1/auth/password-reset-info" {
			return true
		}
	case http.MethodPost:
		if path == "/api/bootstrap/v1/complete" {
			return true
		}
		if path == "/api/admin/v1/auth/login" {
			return true
		}
	case http.MethodHead:
		if path == "/health" {
			return true
		}
	}
	return false
}
