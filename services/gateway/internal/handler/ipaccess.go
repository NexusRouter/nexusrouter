package handler

import (
	"net/http"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/ipaccess"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IPAccessControl 在 IP 限流之后、业务路由（含鉴权）之前检查 ip_access 名单。
func IPAccessControl(store *runtime.Store, log *zap.Logger, col *metrics.Collector) gin.HandlerFunc {
	if store == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		if rateSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		s := store.Snapshot()
		mode := strings.ToLower(strings.TrimSpace(s.IPAccess.Mode))
		if mode == "" {
			mode = "off"
		}
		if mode == "off" {
			c.Next()
			return
		}
		nets, err := ipaccess.Compile(s.IPAccess.CIDRs)
		if err != nil {
			if log != nil {
				log.Error("ip_access compile", zap.Error(err))
			}
			c.Next()
			return
		}
		ip := c.ClientIP()
		hit := ipaccess.Match(ip, nets)
		switch mode {
		case "denylist":
			if !hit {
				c.Next()
				return
			}
		case "allowlist":
			if hit {
				c.Next()
				return
			}
		default:
			c.Next()
			return
		}
		if log != nil {
			log.Warn("ip access denied",
				zap.String("request_id", c.GetString("request_id")),
				zap.String("client_ip", ip),
				zap.String("mode", mode),
				zap.String("reason", "IP_BLOCKED"),
			)
		}
		if col != nil {
			col.RecordGatewayError("IP_BLOCKED")
		}
		WriteGatewayError(c, http.StatusForbidden, "IP_BLOCKED", "客户端 IP 不在允许范围或被拒绝")
		c.Abort()
	}
}
