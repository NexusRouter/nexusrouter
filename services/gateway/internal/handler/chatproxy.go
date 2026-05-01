package handler

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var hopByHopHeaders = []string{
	"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
	"Te", "Trailers", "Transfer-Encoding", "Upgrade",
}

// ChatProxy 将 POST /v1/chat/completions 反向代理至配置的上游。
func ChatProxy(cfg *config.Config, log *zap.Logger) gin.HandlerFunc {
	transport := &http.Transport{
		ResponseHeaderTimeout: cfg.UpstreamTimeout,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
	}

	return func(c *gin.Context) {
		if strings.TrimSpace(cfg.UpstreamBaseURL) == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    "UPSTREAM_NOT_CONFIGURED",
				"message": "上游服务未配置",
			})
			return
		}

		// 缓冲请求体，避免 Gin / 中间件已读导致 ReverseProxy 转发空 body。
		if c.Request.Body != nil {
			body, err := io.ReadAll(c.Request.Body)
			_ = c.Request.Body.Close()
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"code":    "INVALID_REQUEST",
					"message": "无法读取请求体",
				})
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			c.Request.ContentLength = int64(len(body))
		}

		base, err := url.Parse(cfg.UpstreamBaseURL)
		if err != nil || base.Scheme == "" || base.Host == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    "UPSTREAM_INVALID",
				"message": "上游基址无效",
			})
			return
		}

		// 避免 ReverseProxy 在检测到已解析 Form 时丢弃非法 query 导致异常。
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
			log.Warn("chat proxy upstream error",
				zap.Error(err),
				zap.String("request_id", r.Header.Get("X-Request-ID")),
			)
			if w.Header().Get("Content-Type") != "" {
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"code":"BAD_GATEWAY","message":"上游不可用"}`))
		}

		defer func() {
			if rec := recover(); rec != nil {
				log.Error("chat proxy panic", zap.Any("error", rec), zap.String("request_id", c.GetString("request_id")))
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "服务器内部错误",
					})
				}
			}
		}()

		rw := &closeNotifyResponseWriter{ResponseWriter: c.Writer}
		proxy.ServeHTTP(rw, c.Request)
		c.Abort()
	}
}
