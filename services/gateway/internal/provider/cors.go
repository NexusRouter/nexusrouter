package provider

import (
	"net/url"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// DynamicCORS 按运行时快照构造 CORS；未在 gateway.yaml 启用 CORS 时，对常见本机开发 Origin（localhost / 127.0.0.1 / ::1）回显 Access-Control-Allow-Origin，避免仪表盘直连网关时出现浏览器预检失败。
// 引擎级顺序：CORS → RequestID → AcceptLanguage → ZapHTTPAccessLog → Recovery → ErrorJSON → RootStrictNoCache → UploadsStaticCache → IP 限流 → …（见 ProvideEngine 注释）。
func DynamicCORS(store *runtime.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if store == nil {
			c.Next()
			return
		}
		s := store.Snapshot()
		if s == nil || !s.CORS.Enabled {
			applyLocalDevCORSIfNeeded(c)
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

// applyLocalDevCORSIfNeeded 在无全局 CORS 配置时，仅当请求带 Origin 且为本机开发地址时注入 CORS 头；同源或 curl 无 Origin 时保持原样。
func applyLocalDevCORSIfNeeded(c *gin.Context) {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin == "" || !isLocalDevOrigin(origin) {
		c.Next()
		return
	}
	cfg := cors.Config{
		AllowOriginFunc: isLocalDevOrigin,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			"X-Requested-With", "X-Request-ID", "X-Oneapi-Request-Id",
		},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	cors.New(cfg)(c)
}

func isLocalDevOrigin(origin string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	switch u.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
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
