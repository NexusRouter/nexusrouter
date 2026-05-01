package handler

import (
	"net/http"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminReloadKeys 在提供 NEXUSROUTER_ADMIN_RELOAD_TOKEN 时注册；用于热加载密钥文件。
//
//	@Summary		热加载 API 密钥文件
//	@Description	需携带与 NEXUSROUTER_ADMIN_RELOAD_TOKEN 相同的 Bearer 管理令牌；成功时重新读取密钥 JSON。
//	@Tags			admin
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/internal/reload-keys [post]
func AdminReloadKeys(cfg *config.Config, ks *keystore.Store, log *zap.Logger) gin.HandlerFunc {
	want := strings.TrimSpace(cfg.AdminReloadToken)
	return func(c *gin.Context) {
		if want == "" {
			c.Status(http.StatusNotFound)
			return
		}
		h := c.GetHeader("Authorization")
		tok := ""
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(h)), "bearer ") {
			tok = strings.TrimSpace(h[7:])
		}
		if tok != want {
			WriteGatewayError(c, http.StatusUnauthorized, "UNAUTHORIZED", "管理令牌无效或缺失")
			return
		}
		if ks == nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "密钥库未初始化")
			return
		}
		if err := ks.Reload(); err != nil {
			if log != nil {
				log.Error("reload keys failed", zap.Error(err))
			}
			WriteGatewayError(c, http.StatusInternalServerError, "RELOAD_FAILED", "密钥重载失败")
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "keys reloaded"})
	}
}
