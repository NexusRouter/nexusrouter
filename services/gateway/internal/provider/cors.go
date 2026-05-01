package provider

import (
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// DynamicCORS 按运行时快照构造 CORS；未启用时直接放行。
// 引擎级顺序：CORS → RequestID → Recovery → ErrorJSON → IP 限流 → …（见 ProvideEngine 注释）。
func DynamicCORS(store *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if store == nil {
			c.Next()
			return
		}
		s := store.Snapshot()
		if s == nil || !s.CORS.Enabled {
			c.Next()
			return
		}
		mx := s.CORS.MaxAgeSeconds
		if mx < 0 {
			mx = 0
		}
		cfg := cors.Config{
			AllowOrigins:     copyTrimNonEmpty(s.CORS.AllowOrigins),
			AllowMethods:     defaultAllowMethods(s.CORS.AllowMethods),
			AllowHeaders:     copyTrimNonEmpty(s.CORS.AllowHeaders),
			ExposeHeaders:    nil,
			AllowCredentials: false,
			MaxAge:           time.Duration(mx) * time.Second,
		}
		cors.New(cfg)(c)
	}
}

func copyTrimNonEmpty(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func defaultAllowMethods(m []string) []string {
	m = copyTrimNonEmpty(m)
	if len(m) == 0 {
		return []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	}
	return m
}
