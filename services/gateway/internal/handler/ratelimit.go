package handler

import (
	"net/http"
	"strings"
	"sync"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
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
	case strings.HasPrefix(path, "/api/admin"):
		return true
	default:
		return false
	}
}

func burstFromRPS(rps float64, burst int) int {
	if burst > 0 {
		return burst
	}
	b := int(rps + 0.999)
	if b < 1 {
		b = 1
	}
	return b
}

// IPRateLimit 在鉴权前按客户端 IP 限流（规则表 + 全局回退）。
func IPRateLimit(store *runtime.Store, log *zap.Logger, col *metrics.Collector) gin.HandlerFunc {
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
		path := c.Request.URL.Path
		var rps float64
		var burst int
		var limKey string
		if rule := runtime.SelectIPRateRule(s.RateLimitRules, path); rule != nil {
			rps = rule.RPS
			burst = burstFromRPS(rule.RPS, rule.Burst)
			limKey = "iprule:" + rule.ID + ":" + c.ClientIP()
		} else if s.RateLimit.RPSPerIP > 0 {
			rps = s.RateLimit.RPSPerIP
			burst = burstFromRPS(rps, 0)
			limKey = "ipglobal:" + c.ClientIP()
		} else {
			c.Next()
			return
		}
		ip := c.ClientIP()
		mu.Lock()
		l, ok := lim[limKey]
		if !ok {
			l = rate.NewLimiter(rate.Limit(rps), burst)
			lim[limKey] = l
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
			if col != nil {
				col.RecordGatewayError("RATE_LIMITED")
			}
			WriteGatewayError(c, http.StatusTooManyRequests, "RATE_LIMITED", "请求过于频繁")
			c.Abort()
			return
		}
		c.Next()
	}
}

// KeyRateLimit 在鉴权后按上下文 rate_limit_key 限流（规则表 + 全局回退）。
func KeyRateLimit(store *runtime.Store, log *zap.Logger, col *metrics.Collector) gin.HandlerFunc {
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
		path := c.Request.URL.Path
		key := c.GetString("rate_limit_key")
		if key == "" {
			c.Next()
			return
		}
		var rps float64
		var burst int
		var limKey string
		if rule := runtime.SelectKeyRateRule(s.RateLimitRules, path); rule != nil {
			rps = rule.RPS
			burst = burstFromRPS(rule.RPS, rule.Burst)
			limKey = "keyrule:" + rule.ID + ":" + key
		} else if s.RateLimit.RPSPerKey > 0 {
			rps = s.RateLimit.RPSPerKey
			burst = burstFromRPS(rps, 0)
			limKey = "keyglobal:" + key
		} else {
			c.Next()
			return
		}
		mu.Lock()
		l, ok := lim[limKey]
		if !ok {
			l = rate.NewLimiter(rate.Limit(rps), burst)
			lim[limKey] = l
		}
		mu.Unlock()
		if !l.Allow() {
			if log != nil {
				log.Warn("rate limit key",
					zap.String("request_id", c.GetString("request_id")),
					zap.String("reason", "RATE_LIMIT_KEY"),
				)
			}
			if col != nil {
				col.RecordGatewayError("RATE_LIMITED")
			}
			WriteGatewayError(c, http.StatusTooManyRequests, "RATE_LIMITED", "请求过于频繁")
			c.Abort()
			return
		}
		c.Next()
	}
}
