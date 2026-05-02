package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/accesslog"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/upstream"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

func extractChatModelField(body []byte) string {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return ""
	}
	raw, ok := obj["model"]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	return strings.TrimSpace(s)
}

// ChatProxy 将 POST /v1/chat/completions 反向代理至上游。
// 若库中存在 model_instance 记录，则仅使用四表 + model_upstream.base_url（不与 gateway.yaml 混用）。
func ChatProxy(cfg *config.Config, log *zap.Logger, rt *runtime.Store, col *metrics.Collector, db *gorm.DB) gin.HandlerFunc {
	pick := upstream.NewPicker()
	baseTransport := &http.Transport{
		ResponseHeaderTimeout: cfg.UpstreamTimeout,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
	}

	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			if col == nil {
				return
			}
			st := c.Writer.Status()
			if st == 0 {
				st = http.StatusOK
			}
			code := c.GetString("gateway_error_code")
			col.RecordChat(st, time.Since(start).Milliseconds(), code)
		}()

		var body []byte
		if c.Request.Body != nil {
			var err error
			body, err = io.ReadAll(c.Request.Body)
			_ = c.Request.Body.Close()
			if err != nil {
				WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "无法读取请求体")
				return
			}
		}

		var (
			base   *http.Transport
			rev    *httputil.ReverseProxy
			upID   string
			upHost string
		)

		if db != nil && repository.UseDatabaseModelLibrary(db) {
			modelName := extractChatModelField(body)
			if modelName == "" {
				WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求体缺少 model 字段")
				return
			}
			tgt, err := repository.PickChatTarget(db, modelName)
			if err != nil {
				if log != nil {
					log.Debug("chat model pick failed", zap.String("model", modelName), zap.Error(err))
				}
				WriteGatewayError(c, http.StatusNotFound, "MODEL_UNAVAILABLE", "未找到可用的模型实例")
				return
			}
			body = repository.RewriteChatBodyToProvider(body, tgt.ProviderModelCode)
			base = baseTransport.Clone()
			base.ResponseHeaderTimeout = tgt.Timeout
			upID = "inst:" + strconv.FormatInt(tgt.InstanceID, 10)
			upHost = tgt.UpstreamHost
			target := tgt.BaseURL
			rev = &httputil.ReverseProxy{
				Rewrite: func(pr *httputil.ProxyRequest) {
					pr.SetURL(target)
					pr.Out.Host = pr.In.Host
					if tgt.APIKey != "" {
						pr.Out.Header.Set("Authorization", "Bearer "+tgt.APIKey)
					} else {
						pr.Out.Header.Del("Authorization")
					}
					for _, h := range hopByHopHeaders {
						pr.Out.Header.Del(h)
					}
				},
			}
		} else {
			snap := rt.Snapshot()
			var err error
			var pu *url.URL
			pu, upID, upHost, err = pick.Pick(snap)
			if err != nil || pu == nil {
				WriteGatewayError(c, http.StatusServiceUnavailable, "UPSTREAM_NOT_CONFIGURED", "上游服务未配置")
				return
			}
			body = repository.RewriteChatCompletionsModelBody(body, db, upID)
			base = baseTransport
			target := pu
			rev = &httputil.ReverseProxy{
				Rewrite: func(pr *httputil.ProxyRequest) {
					pr.SetURL(target)
					pr.Out.Host = pr.In.Host
					if !cfg.ForwardClientAuthorization {
						pr.Out.Header.Del("Authorization")
						if cfg.UpstreamAPIKey != "" {
							pr.Out.Header.Set("Authorization", "Bearer "+cfg.UpstreamAPIKey)
						}
					}
					for _, h := range hopByHopHeaders {
						pr.Out.Header.Del(h)
					}
				},
			}
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Request.ContentLength = int64(len(body))
		c.Request.Form = nil
		c.Request.PostForm = nil

		rev.Transport = base
		rev.ModifyResponse = func(resp *http.Response) error {
			for _, h := range hopByHopHeaders {
				resp.Header.Del(h)
			}
			return nil
		}
		rev.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
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
			snap := rt.Snapshot()
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

		rev.ServeHTTP(cw, c.Request)
		c.Abort()
	}
}
