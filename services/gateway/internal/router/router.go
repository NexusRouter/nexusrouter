package router

import (
	"net/http"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/adminauth"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/handler"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/openapi"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Deps 注册路由所需依赖。
type Deps struct {
	Config   *config.Config
	Log      *zap.Logger
	KeyStore *keystore.Store
	Runtime  *runtime.Store
	Metrics  *metrics.Collector
	DB       *gorm.DB
}

// Register 注册业务路由、OpenAPI 与 Chat 代理。
// 引擎级顺序见 provider：CORS → RequestID → Recovery → ErrorJSON → IP 限流 → IP 名单 → 本处 GatewayAuth → Key 限流 → ChatProxy。
func Register(r *gin.Engine, d Deps) {
	if d.Log == nil {
		d.Log = zap.NewNop()
	}
	if d.Config == nil {
		d.Config = config.Load()
	}
	cfg, log := d.Config, d.Log

	adm := adminauth.New(cfg, d.DB)
	handler.RegisterFirstBootRoutes(r, cfg, d.DB, adm, log)

	r.GET("/health", handler.Health())

	openapi.Register(r, cfg.EnableSwaggerUI)

	if strings.TrimSpace(cfg.AdminReloadToken) != "" {
		if d.KeyStore != nil {
			r.POST("/internal/reload-keys", handler.AdminReloadKeys(cfg, d.KeyStore, log))
		}
		if d.Runtime != nil {
			r.POST("/internal/reload-config", handler.AdminReloadConfig(cfg, d.Runtime, log))
			r.PUT("/internal/upstream/active", handler.AdminSetActiveUpstream(cfg, d.Runtime, log))
		}
	}

	disallow := func(c *gin.Context) {
		handler.WriteGatewayError(c, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "仅支持 POST")
	}
	for _, method := range []string{
		http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead,
	} {
		r.Handle(method, "/v1/chat/completions", disallow)
	}
	r.Handle(http.MethodOptions, "/v1/chat/completions", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	r.POST("/v1/chat/completions",
		handler.GatewayAuth(d.KeyStore, d.Metrics),
		handler.KeyRateLimit(d.Runtime, log, d.Metrics),
		handler.ChatProxy(cfg, log, d.Runtime, d.Metrics),
	)

	handler.RegisterAdminConsole(r, cfg, adm, d.Metrics, d.Runtime, d.KeyStore, log, d.DB)
}
