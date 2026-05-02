package handler

import (
	"net/http"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/adminauth"
	"github.com/gin-gonic/gin"
)

// adminOperatorWriteGuard 阻止 operator 对除登出外的写方法调用。
func adminOperatorWriteGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		m := c.Request.Method
		if m == http.MethodGet || m == http.MethodHead {
			c.Next()
			return
		}
		p := c.Request.URL.Path
		if m == http.MethodPost && strings.HasSuffix(p, "/auth/logout") {
			c.Next()
			return
		}
		raw, ok := c.Get(adminauth.CtxClaims)
		if !ok {
			c.Next()
			return
		}
		cl, ok := raw.(*adminauth.Claims)
		if !ok || cl == nil {
			c.Next()
			return
		}
		if strings.EqualFold(strings.TrimSpace(cl.Role), "admin") {
			c.Next()
			return
		}
		if strings.EqualFold(strings.TrimSpace(cl.Role), "operator") {
			WriteGatewayError(c, http.StatusForbidden, "FORBIDDEN", "操作员无权执行此写操作")
			c.Abort()
			return
		}
		c.Next()
	}
}
