// Package app 聚合 HTTP 引擎与日志，作为进程入口组合根。
package app

import (
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// Application 网关应用实例。
type Application struct {
	Log    *zap.Logger
	Engine *gin.Engine
}

// NewApplication 由 Wire 注入构造。
func NewApplication(log *zap.Logger, engine *gin.Engine) *Application {
	return &Application{Log: log, Engine: engine}
}

// Run 启动 HTTP 服务，默认监听 :8080。
func (a *Application) Run() error {
	return a.Engine.Run(":8080")
}
