package handler

import (
	"net/http"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
)

const headerAPIKey = "X-API-Key"

// GatewayAuth 校验网关层凭证：优先 Authorization Bearer；可选兼容 X-API-Key（deprecated，见 README）。
func GatewayAuth(ks *keystore.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if ks == nil || !ks.HasKeys() {
			WriteGatewayError(c, http.StatusUnauthorized, "UNAUTHORIZED", "网关未配置允许的 API 密钥")
			return
		}

		authz := strings.TrimSpace(c.GetHeader("Authorization"))
		if len(authz) >= 7 && strings.EqualFold(authz[:7], "bearer ") {
			tok := strings.TrimSpace(authz[7:])
			if ks.ValidateBearer(tok) {
				c.Set("rate_limit_key", runtime.FingerprintBearer(tok))
				c.Next()
				return
			}
		}

		if k := strings.TrimSpace(c.GetHeader(headerAPIKey)); k != "" && ks.ValidateXAPIKey(k) {
			c.Set("rate_limit_key", runtime.FingerprintBearer(k))
			c.Next()
			return
		}

		WriteGatewayError(c, http.StatusUnauthorized, "UNAUTHORIZED", "凭证无效或缺失")
	}
}
