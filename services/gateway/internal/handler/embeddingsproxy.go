package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/accesslog"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/bodyreuse"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/upstream"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// mergeEmbeddingsModelFromPathParam 当 URL 含 :model 且 JSON 未提供非空 model 时，写入 path 中的模型名（与常见 OpenAI 兼容栈对 /v1/engines/{id}/embeddings 的语义一致）。
func mergeEmbeddingsModelFromPathParam(body []byte, pathModel string) []byte {
	pathModel = strings.TrimSpace(pathModel)
	if pathModel == "" {
		return body
	}
	var probe struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return body
	}
	if strings.TrimSpace(probe.Model) != "" {
		return body
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return body
	}
	b, err := json.Marshal(pathModel)
	if err != nil {
		return body
	}
	obj["model"] = json.RawMessage(b)
	out, err := json.Marshal(obj)
	if err != nil {
		return body
	}
	return out
}

func validateEmbeddingsRequestBody(body []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return errors.New("请求体须为 JSON 对象")
	}
	raw, ok := obj["input"]
	if !ok {
		return errors.New("请求体缺少 input 字段")
	}
	if len(bytes.TrimSpace(raw)) == 0 || string(raw) == "null" {
		return errors.New("input 不可为 null 或空")
	}
	return nil
}

// EmbeddingsProxy 将 POST /v1/embeddings 与 POST /v1/engines/:model/embeddings 反向代理至上游；鉴权、限流、上游选择与改写语义与 Chat 代理对齐（不含 Chat 专用校验与流式 Accept 改写）。
func EmbeddingsProxy(cfg *config.Config, log *zap.Logger, rt *runtime.Store, col *metrics.Collector, db *gorm.DB) gin.HandlerFunc {
	pick := upstream.NewPicker()
	baseTransport := newChatUpstreamTransport(cfg, cfg.UpstreamTimeout)

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

		body, err := bodyreuse.GetRequestBody(c)
		if err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "无法读取请求体")
			return
		}
		body = mergeEmbeddingsModelFromPathParam(body, c.Param("model"))
		if err := validateEmbeddingsRequestBody(body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
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
					log.Debug("embeddings model pick failed", zap.String("model", modelName), zap.Error(err))
				}
				WriteGatewayError(c, http.StatusNotFound, "MODEL_UNAVAILABLE", "未找到可用的模型实例")
				return
			}
			body = repository.RewriteChatBodyToProvider(body, tgt.ProviderModelCode)
			base = newChatUpstreamTransport(cfg, tgt.Timeout)
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
					sanitizeChatOutboundRequestHeaders(pr.Out.Header)
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
					sanitizeChatOutboundRequestHeaders(pr.Out.Header)
				},
			}
		}

		bodyreuse.ResetRequestBody(c, body)

		rev.Transport = base
		rev.ModifyResponse = func(resp *http.Response) error {
			for _, h := range hopByHopHeaders {
				resp.Header.Del(h)
			}
			ensureSSEProxyResponseHeaders(resp.Header)
			return nil
		}
		rev.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			if log != nil {
				log.Warn("embeddings proxy upstream error",
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
					log.Error("embeddings proxy panic",
						zap.Any("error", rec),
						zap.String("request_id", c.GetString("request_id")),
						zap.String("stack", string(debug.Stack())),
					)
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
