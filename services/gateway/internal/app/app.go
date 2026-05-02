// Package app 聚合 HTTP 引擎与日志，作为进程入口组合根。
package app

import (
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/alerts"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Application 网关应用实例。
type Application struct {
	Log      *zap.Logger
	Engine   *gin.Engine
	KeyStore *keystore.Store
	Runtime  *runtime.Store
	Config   *config.Config
	Metrics  *metrics.Collector
}

// NewApplication 由 Wire 注入构造。
func NewApplication(log *zap.Logger, engine *gin.Engine, ks *keystore.Store, rt *runtime.Store, cfg *config.Config, col *metrics.Collector) *Application {
	return &Application{Log: log, Engine: engine, KeyStore: ks, Runtime: rt, Config: cfg, Metrics: col}
}

// Run 启动 HTTP 服务；监听地址来自 `NEXUSROUTER_HTTP_LISTEN_ADDR`（默认 `:8080`）。若密钥库或网关配置基于文件则注册 SIGHUP 热加载（非 Windows）。
func (a *Application) Run() error {
	if a.KeyStore != nil {
		a.KeyStore.ListenSIGHUP(a.Log)
	}
	if a.Runtime != nil {
		a.Runtime.ListenSIGHUP(a.Log)
	}
	if a.Metrics != nil && a.Runtime != nil {
		alerts.StartBackground(a.Runtime, a.Metrics, a.Log)
	}
	addr := ":8080"
	if a.Config != nil {
		if s := strings.TrimSpace(a.Config.HTTPListenAddr); s != "" {
			addr = s
		}
	}
	return a.Engine.Run(addr)
}
