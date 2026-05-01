package handler

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/accesslog"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/upstream"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var hopByHopHeaders = []string{
	"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
	"Te", "Trailer", "Transfer-Encoding", "Upgrade",
}

type captureWriter struct {
	gin.ResponseWriter
	status int
}

func (w *captureWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *captureWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// ChatProxy 将 POST /v1/chat/completions 反向代理至运行时选中的上游。
// 引擎级中间件顺序见 provider：CORS → RequestID → Recovery → ErrorJSON → IP 限流 →（本链）鉴权 → Key 限流 → ChatProxy。
func ChatProxy(cfg *config.Config, log *zap.Logger, rt *runtime.Store) gin.HandlerFunc {
	pick := upstream.NewPicker()
	transport := &http.Transport{
		ResponseHeaderTimeout: cfg.UpstreamTimeout,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
	}

	return func(c *gin.Context) {
		snap := rt.Snapshot()
		base, upID, upHost, err := pick.Pick(snap)
		if err != nil || base == nil {
			WriteGatewayError(c, http.StatusServiceUnavailable, "UPSTREAM_NOT_CONFIGURED", "上游服务未配置")
			return
		}

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

		start := time.Now()
		baseW := &closeNotifyResponseWriter{ResponseWriter: c.Writer}
		cw := &captureWriter{ResponseWriter: baseW}
		defer func() {
			if rec := recover(); rec != nil {
				if log != nil {
					log.Error("chat proxy panic", zap.Any("error", rec), zap.String("request_id", c.GetString("request_id")))
				}
				if !c.Writer.Written() {
					WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误")
				}
			}
			st := cw.status
			if st == 0 {
				st = c.Writer.Status()
			}
			gwErr := st == http.StatusBadGateway || st == http.StatusGatewayTimeout || st == http.StatusInternalServerError
			dur := time.Since(start).Milliseconds()
			fields := []zap.Field{
				zap.String("request_id", c.GetString("request_id")),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.String("upstream_id", upID),
				zap.String("upstream_host", upHost),
				zap.Int("status", st),
				zap.Int64("duration_ms", dur),
			}
			if fp := c.GetString("rate_limit_key"); fp != "" {
				fields = append(fields, zap.String("api_key_fp", fp))
			}
			accesslog.New(snap).Write(st, gwErr, fields...)
		}()

		proxy.ServeHTTP(cw, c.Request)
		c.Abort()
	}
}
