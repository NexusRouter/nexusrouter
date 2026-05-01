package handler

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var hopByHopHeaders = []string{
	"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
	"Te", "Trailer", "Transfer-Encoding", "Upgrade",
}

// ChatProxy 将 POST /v1/chat/completions 反向代理至配置的上游（支持多上游轮询）。
// 中间件顺序由路由注册保证：RequestID → Recovery → GatewayAuth → ChatProxy。
func ChatProxy(cfg *config.Config, log *zap.Logger) gin.HandlerFunc {
	bases := cfg.EffectiveUpstreamBases()
	if len(bases) == 0 {
		return func(c *gin.Context) {
			WriteGatewayError(c, http.StatusServiceUnavailable, "UPSTREAM_NOT_CONFIGURED", "上游服务未配置")
		}
	}
	parsed := make([]*url.URL, 0, len(bases))
	for _, b := range bases {
		u, err := url.Parse(b)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return func(c *gin.Context) {
				WriteGatewayError(c, http.StatusServiceUnavailable, "UPSTREAM_INVALID", "上游基址无效")
			}
		}
		parsed = append(parsed, u)
	}

	var rr atomic.Uint32
	transport := &http.Transport{
		ResponseHeaderTimeout: cfg.UpstreamTimeout,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
	}

	return func(c *gin.Context) {
		if c.Request.Body != nil {
			body, err := io.ReadAll(c.Request.Body)
			_ = c.Request.Body.Close()
			if err != nil {
				WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "无法读取请求体")
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			c.Request.ContentLength = int64(len(body))
		}

		n := uint32(len(parsed))
		i := rr.Add(1)
		base := parsed[i%n]

		c.Request.Form = nil
		c.Request.PostForm = nil

		proxy := httputil.NewSingleHostReverseProxy(base)
		orig := proxy.Director
		proxy.Director = func(req *http.Request) {
			orig(req)
			if !cfg.ForwardClientAuthorization {
				req.Header.Del("Authorization")
				if cfg.UpstreamAPIKey != "" {
					req.Header.Set("Authorization", "Bearer "+cfg.UpstreamAPIKey)
				}
			}
			for _, h := range hopByHopHeaders {
				req.Header.Del(h)
			}
		}
		proxy.Transport = transport
		proxy.ModifyResponse = func(resp *http.Response) error {
			for _, h := range hopByHopHeaders {
				resp.Header.Del(h)
			}
			return nil
		}
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			if log != nil {
				log.Warn("chat proxy upstream error",
					zap.Error(err),
					zap.String("request_id", r.Header.Get(headerRequestID)),
				)
			}
			if w.Header().Get("Content-Type") != "" {
				return
			}
			WriteGatewayErrorHTTP(w, r, http.StatusBadGateway, "BAD_GATEWAY", "上游不可用")
		}

		defer func() {
			if rec := recover(); rec != nil {
				if log != nil {
					log.Error("chat proxy panic", zap.Any("error", rec), zap.String("request_id", c.GetString("request_id")))
				}
				if !c.Writer.Written() {
					WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误")
				}
			}
		}()

		rw := &closeNotifyResponseWriter{ResponseWriter: c.Writer}
		proxy.ServeHTTP(rw, c.Request)
		c.Abort()
	}
}
