package handler

import (
	"net/http"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/gin-gonic/gin"
)

const headerAPIKey = "X-API-Key"

// GatewayAuth 校验网关层凭证（Bearer 或 X-API-Key），与配置中 NEXUSROUTER_GATEWAY_API_KEYS 之一匹配。
func GatewayAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(cfg.GatewayAPIKeys) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "网关未配置允许的 API 密钥",
			})
			return
		}
		if matchGatewayAuth(c, cfg.GatewayAPIKeys) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "凭证无效或缺失",
		})
	}
}

func matchGatewayAuth(c *gin.Context, keys []string) bool {
	if h := c.GetHeader("Authorization"); strings.HasPrefix(strings.ToLower(h), "bearer ") {
		token := strings.TrimSpace(h[7:])
		for _, k := range keys {
			if k == token {
				return true
			}
		}
	}
	if k := c.GetHeader(headerAPIKey); k != "" {
		for _, want := range keys {
			if want == k {
				return true
			}
		}
	}
	return false
}
