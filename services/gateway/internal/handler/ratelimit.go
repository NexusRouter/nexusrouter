package handler

import (
	"net/http"
	"strings"
	"sync"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func rateSkipPath(path string) bool {
	switch {
	case path == "/health", strings.HasPrefix(path, "/openapi"), strings.HasPrefix(path, "/swagger"):
		return true
	case strings.HasPrefix(path, "/internal"):
		return true
	default:
		return false
	}
}

// IPRateLimit 在鉴权前按客户端 IP 限流。
func IPRateLimit(store *runtime.Store, log *zap.Logger) gin.HandlerFunc {
	if store == nil {
		return func(c *gin.Context) { c.Next() }
	}
	var mu sync.Mutex
	lim := map[string]*rate.Limiter{}
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
		if s.RateLimit.RPSPerIP <= 0 {
			c.Next()
			return
		}
		ip := c.ClientIP()
		rps := s.RateLimit.RPSPerIP
		burst := int(rps + 0.999)
		if burst < 1 {
			burst = 1
		}
		mu.Lock()
		l, ok := lim[ip]
		if !ok {
			l = rate.NewLimiter(rate.Limit(rps), burst)
			lim[ip] = l
		}
		mu.Unlock()
		if !l.Allow() {
			if log != nil {
				log.Warn("rate limit ip",
					zap.String("request_id", c.GetString("request_id")),
					zap.String("client_ip", ip),
					zap.String("reason", "RATE_LIMIT_IP"),
				)
			}
			WriteGatewayError(c, http.StatusTooManyRequests, "RATE_LIMITED", "请求过于频繁")
			c.Abort()
			return
		}
		c.Next()
	}
}

// KeyRateLimit 在鉴权后按上下文 rate_limit_key 限流。
func KeyRateLimit(store *runtime.Store, log *zap.Logger) gin.HandlerFunc {
	if store == nil {
		return func(c *gin.Context) { c.Next() }
	}
	var mu sync.Mutex
	lim := map[string]*rate.Limiter{}
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		s := store.Snapshot()
		if s.RateLimit.RPSPerKey <= 0 {
			c.Next()
			return
		}
		key := c.GetString("rate_limit_key")
		if key == "" {
			c.Next()
			return
		}
		rps := s.RateLimit.RPSPerKey
		burst := int(rps + 0.999)
		if burst < 1 {
			burst = 1
		}
		mu.Lock()
		l, ok := lim[key]
		if !ok {
			l = rate.NewLimiter(rate.Limit(rps), burst)
			lim[key] = l
		}
		mu.Unlock()
		if !l.Allow() {
			if log != nil {
				log.Warn("rate limit key",
					zap.String("request_id", c.GetString("request_id")),
					zap.String("reason", "RATE_LIMIT_KEY"),
				)
			}
			WriteGatewayError(c, http.StatusTooManyRequests, "RATE_LIMITED", "请求过于频繁")
			c.Abort()
			return
		}
		c.Next()
	}
}
