package handler

import (
	"net/http"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminReloadConfig 从磁盘重新加载 gateway.yaml（路径为空时无操作并返回 ok）。
//
//	@Summary		热加载网关配置文件
//	@Description	需携带与 NEXUSROUTER_ADMIN_RELOAD_TOKEN 相同的 Bearer；失败时保留旧快照。
//	@Tags			admin
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/internal/reload-config [post]
func AdminReloadConfig(cfg *config.Config, rt *runtime.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireAdminReloadToken(cfg, c) {
			return
		}
		if rt == nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "运行时未初始化")
			return
		}
		if err := rt.Reload(); err != nil {
			if log != nil {
				log.Error("reload gateway config failed", zap.Error(err), zap.String("path", rt.Path()))
			}
			WriteGatewayError(c, http.StatusInternalServerError, "RELOAD_FAILED", "网关配置重载失败")
			return
		}
		if log != nil {
			log.Info("gateway config reloaded", zap.String("path", rt.Path()))
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "gateway config reloaded"})
	}
}

// SetActiveUpstreamBody PUT /internal/upstream/active 请求体（供 swag 引用）。
type SetActiveUpstreamBody struct {
	ActiveUpstreamID string `json:"active_upstream_id"`
}

// AdminSetActiveUpstream 仅更新内存中的 active_upstream_id；不自动写回磁盘（见 README）。
//
//	@Summary		固定当前上游 id
//	@Description	传空字符串可解除 pin；需 Bearer 管理令牌。
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			body	body	SetActiveUpstreamBody	true	"active_upstream_id 可为空表示解除 pin"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/internal/upstream/active [put]
func AdminSetActiveUpstream(cfg *config.Config, rt *runtime.Store, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireAdminReloadToken(cfg, c) {
			return
		}
		if rt == nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "运行时未初始化")
			return
		}
		var body SetActiveUpstreamBody
		if err := c.ShouldBindJSON(&body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体须为 JSON：active_upstream_id")
			return
		}
		if err := rt.SetActiveUpstream(body.ActiveUpstreamID); err != nil {
			if log != nil {
				log.Warn("set active upstream failed", zap.Error(err))
			}
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_UPSTREAM", err.Error())
			return
		}
		if log != nil {
			log.Info("active upstream updated", zap.String("active_upstream_id", body.ActiveUpstreamID))
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "active_upstream_id": body.ActiveUpstreamID})
	}
}
