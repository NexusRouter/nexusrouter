package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"net"
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

func newChatUpstreamTransport(cfg *config.Config, responseHeaderTimeout time.Duration) *http.Transport {
	tr := &http.Transport{
		ResponseHeaderTimeout: responseHeaderTimeout,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
		// 不向出站请求自动附加 Accept-Encoding: gzip，与出站头剔除一致，避免上游按压缩响应体协商。
		DisableCompression: true,
	}
	if cfg == nil {
		return tr
	}
	p := strings.TrimSpace(cfg.UpstreamHTTPProxy)
	if p == "" {
		return tr
	}
	u, err := url.Parse(p)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return tr
	}
	tr.Proxy = http.ProxyURL(u)
	return tr
}

var hopByHopHeaders = []string{
	"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
	"Te", "Trailer", "Transfer-Encoding", "Upgrade",
}

// sanitizeChatOutboundRequestHeaders 在发往 Chat 上游的出站请求上剔除 hop-by-hop 头，并移除 Accept-Encoding，
// 避免将客户端对响应体的压缩协商原样带给上游（与常见 OpenAI 兼容代理行为一致）。
func sanitizeChatOutboundRequestHeaders(h http.Header) {
	for _, name := range hopByHopHeaders {
		h.Del(name)
	}
	h.Del("Accept-Encoding")
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

// ensureSSEProxyResponseHeaders 在反向代理回写客户端前为 SSE 响应补充提示头，减少中间层整段缓冲导致的首包与增量延迟。
// 若上游未提供 Cache-Control，则设为 no-cache，避免中间层误缓存流式响应。
// 若 Connection 缺省，则设为 keep-alive，便于 HTTP/1.x 客户端与部分中间层保持长连接语义。
func ensureSSEProxyResponseHeaders(h http.Header) {
	ct := strings.ToLower(strings.TrimSpace(h.Get("Content-Type")))
	if ct == "" || !strings.Contains(ct, "text/event-stream") {
		return
	}
	h.Set("X-Accel-Buffering", "no")
	if strings.TrimSpace(h.Get("Cache-Control")) == "" {
		h.Set("Cache-Control", "no-cache")
	}
	if strings.TrimSpace(h.Get("Connection")) == "" {
		h.Set("Connection", "keep-alive")
	}
}

func chatRequestStreamEnabled(body []byte) bool {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return false
	}
	raw, ok := obj["stream"]
	if !ok {
		return false
	}
	var v bool
	if err := json.Unmarshal(raw, &v); err != nil {
		return false
	}
	return v
}

// applyChatUpstreamDefaultHeaders 在转发前按请求体与入站头补充上游协商所需头（不修改 body）。
// 当 JSON 顶层 stream 为 true 且客户端未提供非空 Accept 时，设置 Accept: text/event-stream，与常见 OpenAI 兼容上游对流式的期望一致。
func applyChatUpstreamDefaultHeaders(pr *httputil.ProxyRequest, body []byte) {
	if strings.TrimSpace(pr.In.Header.Get("Accept")) != "" {
		return
	}
	if !chatRequestStreamEnabled(body) {
		return
	}
	pr.Out.Header.Set("Accept", "text/event-stream")
}

// mergeChatStreamIncludeUsage 在顶层 stream 为 true 时合并或写入 stream_options.include_usage=true，保留 stream_options 中其它键。
// 若 body 非对象 JSON、stream 非 true、或 stream_options 已存在且非 JSON 对象，则返回原样 body。
func mergeChatStreamIncludeUsage(body []byte) []byte {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return body
	}
	rawStream, ok := obj["stream"]
	if !ok {
		return body
	}
	var streamVal bool
	if err := json.Unmarshal(rawStream, &streamVal); err != nil || !streamVal {
		return body
	}
	opts := map[string]json.RawMessage{}
	if rawSO, ok := obj["stream_options"]; ok && len(rawSO) > 0 {
		if err := json.Unmarshal(rawSO, &opts); err != nil {
			return body
		}
	}
	opts["include_usage"] = json.RawMessage("true")
	merged, err := json.Marshal(opts)
	if err != nil {
		return body
	}
	obj["stream_options"] = json.RawMessage(merged)
	out, err := json.Marshal(obj)
	if err != nil {
		return body
	}
	return out
}

// validateChatCompletionsMaxTokens 在 body 为 JSON 对象且含 max_tokens 时校验其值：
// 须为整数，范围 [0, MaxInt32/2]；与常见 OpenAI 兼容栈对非法 max_tokens 的拒绝语义一致。
func validateChatCompletionsMaxTokens(body []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil
	}
	raw, ok := obj["max_tokens"]
	if !ok {
		return nil
	}
	if len(bytes.TrimSpace(raw)) == 0 || string(raw) == "null" {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return errors.New("max_tokens 须为整数")
	}
	if v == nil {
		return nil
	}
	f, ok := v.(float64)
	if !ok {
		return errors.New("max_tokens 须为整数")
	}
	if f != math.Trunc(f) {
		return errors.New("max_tokens 须为整数")
	}
	if f < 0 || f > float64(math.MaxInt32/2) {
		return errors.New("max_tokens 超出允许范围")
	}
	return nil
}

// validateChatCompletionsMessages 在 body 为 JSON 对象时要求顶层 messages 为至少含一项的 JSON 数组；
// 缺省、为 null、类型非数组或长度为 0 时拒绝，与常见 Chat Completions 必填语义一致。
func validateChatCompletionsMessages(body []byte) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil
	}
	raw, ok := obj["messages"]
	if !ok {
		return errors.New("请求体缺少 messages 字段")
	}
	if len(bytes.TrimSpace(raw)) == 0 || string(raw) == "null" {
		return errors.New("messages 不可为 null")
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return errors.New("messages 须为数组")
	}
	if len(arr) == 0 {
		return errors.New("messages 须至少包含一条消息")
	}
	return nil
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
// 发往上游的 HTTP 客户端可在配置中显式设置代理 URL；未设置时沿用 net/http 默认（含 HTTP_PROXY 等环境变量）。
func ChatProxy(cfg *config.Config, log *zap.Logger, rt *runtime.Store, col *metrics.Collector, db *gorm.DB) gin.HandlerFunc {
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
		if err := validateChatCompletionsMaxTokens(body); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		if err := validateChatCompletionsMessages(body); err != nil {
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
					log.Debug("chat model pick failed", zap.String("model", modelName), zap.Error(err))
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
					applyChatUpstreamDefaultHeaders(pr, body)
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
					applyChatUpstreamDefaultHeaders(pr, body)
				},
			}
		}

		if cfg.ChatStreamIncludeUsageEffective() {
			body = mergeChatStreamIncludeUsage(body)
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
					log.Error("chat proxy panic",
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
