package provider

import (
	"net/http"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/handler"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/router"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProvideEngine 构造 Gin 引擎并注册路由与中间件。
// 中间件顺序：CORS（含 OPTIONS）→ RequestID → ZapRecovery → ErrorJSON → per-IP 限流（鉴权前）→
// IP 名单 → 业务路由内 Chat 链：GatewayAuth → per-Key 限流 → ChatProxy。
func ProvideEngine(log *zap.Logger, cfg *config.Config, ks *keystore.Store, rt *runtime.Store, col *metrics.Collector, db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(DynamicCORS(rt))
	e.Use(router.RequestID())
	e.Use(router.ZapRecovery(log))
	e.Use(router.ErrorJSON(log))
	e.Use(handler.IPRateLimit(rt, log, col))
	e.Use(handler.IPAccessControl(rt, log, col))
	router.Register(e, router.Deps{Config: cfg, Log: log, KeyStore: ks, Runtime: rt, Metrics: col, DB: db})
	e.NoRoute(func(c *gin.Context) {
		handler.WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "路由不存在")
	})
	return e
}
