// Package app 聚合 HTTP 引擎与日志，作为进程入口组合根。
package app

import (
	"go.uber.org/zap"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/gin-gonic/gin"
)

// Application 网关应用实例。
type Application struct {
	Log      *zap.Logger
	Engine   *gin.Engine
	KeyStore *keystore.Store
}

// NewApplication 由 Wire 注入构造。
func NewApplication(log *zap.Logger, engine *gin.Engine, ks *keystore.Store) *Application {
	return &Application{Log: log, Engine: engine, KeyStore: ks}
}

// Run 启动 HTTP 服务，默认监听 :8080；若密钥库基于文件则注册 SIGHUP 热加载（非 Windows）。
func (a *Application) Run() error {
	if a.KeyStore != nil {
		a.KeyStore.ListenSIGHUP(a.Log)
	}
	return a.Engine.Run(":8080")
}
