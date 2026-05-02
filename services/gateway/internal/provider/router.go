package provider

import (
	"net/http"
	"net/url"
	"os"
	"strings"

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

// applyGinModeFromEnv 根据环境变量 GIN_MODE 设置 Gin 模式：值为 debug（忽略首尾空白，大小写敏感，与 Gin 常量一致）时为调试模式，否则为发布模式（含未设置）。
func applyGinModeFromEnv() {
	if strings.TrimSpace(os.Getenv("GIN_MODE")) == gin.DebugMode {
		gin.SetMode(gin.DebugMode)
		return
	}
	gin.SetMode(gin.ReleaseMode)
}

// ProvideEngine 构造 Gin 引擎并注册路由与中间件。
// 中间件顺序：CORS（含 OPTIONS）→ RequestID → AcceptLanguage → ZapHTTPAccessLog → GzipRequestDecode → ZapRecovery → ErrorJSON → RootStrictNoCache → UploadsStaticCache → per-IP 限流（鉴权前）→
// IP 名单 → 业务路由内 OpenAI 兼容链：GatewayAuth → per-Key 限流 → ChatProxy / EmbeddingsProxy / ModerationsProxy / ImagesGenerationsProxy / AudioSpeechProxy。
func ProvideEngine(log *zap.Logger, cfg *config.Config, ks *keystore.Store, rt *runtime.Store, col *metrics.Collector, db *gorm.DB) *gin.Engine {
	applyGinModeFromEnv()
	e := gin.New()
	e.Use(DynamicCORS(rt))
	e.Use(router.RequestID())
	e.Use(router.AcceptLanguage())
	e.Use(router.ZapHTTPAccessLog(log))
	e.Use(router.GzipRequestDecode())
	e.Use(router.ZapRecovery(log))
	e.Use(router.ErrorJSON(log))
	e.Use(router.RootStrictNoCache())
	e.Use(router.UploadsStaticCache())
	e.Use(router.BootstrapGate(db, log))
	e.Use(handler.IPRateLimit(rt, log, col))
	e.Use(handler.IPAccessControl(rt, log, col))
	router.Register(e, router.Deps{Config: cfg, Log: log, KeyStore: ks, Runtime: rt, Metrics: col, DB: db})
	e.NoRoute(noRouteHandler(cfg))
	return e
}

func noRouteHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		base := ""
		if cfg != nil {
			base = strings.TrimSuffix(strings.TrimSpace(cfg.FrontendBaseURL), "/")
		}
		if base != "" {
			if u, err := url.Parse(base); err == nil && u.Host != "" {
				scheme := strings.ToLower(u.Scheme)
				if scheme == "http" || scheme == "https" {
					c.Redirect(http.StatusMovedPermanently, base+c.Request.URL.RequestURI())
					return
				}
			}
		}
		p := c.Request.URL.Path
		if p == "/v1" || strings.HasPrefix(p, "/v1/") {
			handler.WriteOpenAINotFoundPath(c)
			return
		}
		handler.WriteGatewayError(c, http.StatusNotFound, "NOT_FOUND", "路由不存在")
	}
}
