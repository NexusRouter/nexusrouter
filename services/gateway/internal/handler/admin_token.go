package handler

import (
	"net/http"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/gin-gonic/gin"
)

// requireAdminReloadToken 校验与 NEXUSROUTER_ADMIN_RELOAD_TOKEN 相同的 Bearer；失败时已写入 JSON 错误。
func requireAdminReloadToken(cfg *config.Config, c *gin.Context) bool {
	want := strings.TrimSpace(cfg.AdminReloadToken)
	if want == "" {
		c.Status(http.StatusNotFound)
		return false
	}
	h := c.GetHeader("Authorization")
	tok := ""
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(h)), "bearer ") {
		tok = strings.TrimSpace(h[7:])
	}
	if tok != want {
		WriteGatewayError(c, http.StatusUnauthorized, "UNAUTHORIZED", "管理令牌无效或缺失")
		return false
	}
	return true
}
