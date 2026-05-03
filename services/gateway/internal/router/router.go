package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/adminauth"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/handler"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
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

// Register 注册业务路由与 OpenAI 兼容代理（Chat、Embeddings、Moderations、Images Generations、Audio Speech）。
// 引擎级顺序见 provider：CORS → RequestID → AcceptLanguage → ZapHTTPAccessLog → GzipRequestDecode → Recovery → ErrorJSON → RootStrictNoCache → UploadsStaticCache → IP 限流 → IP 名单 → 本处 GatewayAuth → Key 限流 → ChatProxy / EmbeddingsProxy / ModerationsProxy / ImagesGenerationsProxy / AudioSpeechProxy。（GET /api/status 无需鉴权。）
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

	health := handler.Health()
	r.GET("/health", health)
	r.HEAD("/health", health)

	r.GET("/api/status", handler.APIStatus())

	uploadRoot := cfg.EffectiveUploadsDir()
	if err := os.MkdirAll(filepath.Join(uploadRoot, "vendor-logos"), 0755); err != nil {
		log.Warn("创建上传目录失败", zap.String("path", uploadRoot), zap.Error(err))
	} else {
		r.Static("/uploads", uploadRoot)
	}

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
	disallowModels := func(c *gin.Context) {
		handler.WriteGatewayError(c, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "仅支持 GET")
	}
	for _, method := range []string{
		http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead,
	} {
		r.Handle(method, "/v1/chat/completions", disallow)
	}
	for _, method := range []string{
		http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead,
	} {
		r.Handle(method, "/v1/embeddings", disallow)
		r.Handle(method, "/v1/engines/:model/embeddings", disallow)
		r.Handle(method, "/v1/moderations", disallow)
		r.Handle(method, "/v1/images/generations", disallow)
		r.Handle(method, "/v1/audio/speech", disallow)
	}
	for _, method := range []string{
		http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead,
	} {
		r.Handle(method, "/v1/models", disallowModels)
	}
	// DELETE /v1/models/:model 由显式未实现路由返回 501，不在此处以 405 占位。
	for _, method := range []string{
		http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodHead,
	} {
		r.Handle(method, "/v1/models/:model", disallowModels)
	}
	r.Handle(http.MethodOptions, "/v1/chat/completions", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.Handle(http.MethodOptions, "/v1/models", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.Handle(http.MethodOptions, "/v1/models/:model", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.Handle(http.MethodOptions, "/v1/embeddings", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.Handle(http.MethodOptions, "/v1/engines/:model/embeddings", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.Handle(http.MethodOptions, "/v1/moderations", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.Handle(http.MethodOptions, "/v1/images/generations", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.Handle(http.MethodOptions, "/v1/audio/speech", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	r.GET("/v1/models",
		handler.GatewayAuth(d.KeyStore, d.Metrics),
		handler.ListModels(d.DB, d.Runtime),
	)
	r.GET("/v1/models/:model",
		handler.GatewayAuth(d.KeyStore, d.Metrics),
		handler.RetrieveModel(d.DB, d.Runtime),
	)

	r.POST("/v1/chat/completions",
		handler.GatewayAuth(d.KeyStore, d.Metrics),
		handler.KeyRateLimit(d.Runtime, log, d.Metrics),
		handler.ChatProxy(cfg, log, d.Runtime, d.Metrics, d.DB),
	)
	r.POST("/v1/embeddings",
		handler.GatewayAuth(d.KeyStore, d.Metrics),
		handler.KeyRateLimit(d.Runtime, log, d.Metrics),
		handler.EmbeddingsProxy(cfg, log, d.Runtime, d.Metrics, d.DB),
	)
	r.POST("/v1/engines/:model/embeddings",
		handler.GatewayAuth(d.KeyStore, d.Metrics),
		handler.KeyRateLimit(d.Runtime, log, d.Metrics),
		handler.EmbeddingsProxy(cfg, log, d.Runtime, d.Metrics, d.DB),
	)
	r.POST("/v1/moderations",
		handler.GatewayAuth(d.KeyStore, d.Metrics),
		handler.KeyRateLimit(d.Runtime, log, d.Metrics),
		handler.ModerationsProxy(cfg, log, d.Runtime, d.Metrics, d.DB),
	)
	r.POST("/v1/images/generations",
		handler.GatewayAuth(d.KeyStore, d.Metrics),
		handler.KeyRateLimit(d.Runtime, log, d.Metrics),
		handler.ImagesGenerationsProxy(cfg, log, d.Runtime, d.Metrics, d.DB),
	)
	r.POST("/v1/audio/speech",
		handler.GatewayAuth(d.KeyStore, d.Metrics),
		handler.KeyRateLimit(d.Runtime, log, d.Metrics),
		handler.AudioSpeechProxy(cfg, log, d.Runtime, d.Metrics, d.DB),
	)

	registerOpenAIV1NotImplementedRoutes(r, d)

	handler.RegisterAdminConsole(r, cfg, adm, d.Metrics, d.Runtime, d.KeyStore, log, d.DB)
}
