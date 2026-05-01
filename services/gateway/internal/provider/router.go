package provider

import (
	"net/http"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/router"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProvideEngine 构造 Gin 引擎并注册路由与中间件。
func ProvideEngine(log *zap.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(router.RequestID())
	e.Use(router.ZapRecovery(log))
	e.Use(router.ErrorJSON(log))
	router.Register(e)
	e.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "路由不存在",
		})
	})
	return e
}
