package router

import (
	"net/http"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/handler"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/openapi"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Deps 注册路由所需依赖。
type Deps struct {
	Config *config.Config
	Log    *zap.Logger
}

// Register 注册业务路由、OpenAPI 与 Chat 代理。
func Register(r *gin.Engine, d Deps) {
	if d.Log == nil {
		d.Log = zap.NewNop()
	}
	if d.Config == nil {
		d.Config = config.Load()
	}
	cfg, log := d.Config, d.Log

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	openapi.Register(r, cfg.EnableSwaggerUI)

	disallow := func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    "METHOD_NOT_ALLOWED",
			"message": "仅支持 POST",
		})
	}
	for _, method := range []string{
		http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead,
	} {
		r.Handle(method, "/v1/chat/completions", disallow)
	}
	r.Handle(http.MethodOptions, "/v1/chat/completions", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	r.POST("/v1/chat/completions", handler.GatewayAuth(cfg), handler.ChatProxy(cfg, log))
}
