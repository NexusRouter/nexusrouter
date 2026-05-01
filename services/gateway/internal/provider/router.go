package provider

import (
	"net/http"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/handler"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/router"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProvideEngine 构造 Gin 引擎并注册路由与中间件。
func ProvideEngine(log *zap.Logger, cfg *config.Config, ks *keystore.Store) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(router.RequestID())
	e.Use(router.ZapRecovery(log))
	e.Use(router.ErrorJSON(log))
	router.Register(e, router.Deps{Config: cfg, Log: log, KeyStore: ks})
	e.NoRoute(func(c *gin.Context) {
		handler.WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "路由不存在")
	})
	return e
}
