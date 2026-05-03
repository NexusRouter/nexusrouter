package router

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/handler"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/locale"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const headerRequestID = "X-Request-ID"

// headerRequestIDLegacy 为部分既有 OpenAI 兼容代理客户端使用的请求关联头名；与 headerRequestID 同值回写，且仅在主头缺省时作为入站回退读取。
const headerRequestIDLegacy = "X-Oneapi-Request-Id"

// uploadsStaticCacheControl 与常见静态资源长期缓存一致（秒级 max-age）；仅用于 `/uploads/` 下成功响应且未已有 Cache-Control 时。
const uploadsStaticCacheControl = "public, max-age=604800"

const rootStrictNoCacheControl = "no-cache"

func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// GzipRequestDecode 当客户端声明 Content-Encoding: gzip 时解压请求体，并更新 Content-Length、移除 Content-Encoding，
// 供后续处理器与上游转发按明文 body 处理（与未压缩请求一致）。
func GzipRequestDecode() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.EqualFold(strings.TrimSpace(c.GetHeader("Content-Encoding")), "gzip") {
			c.Next()
			return
		}
		if c.Request.Body == nil {
			handler.WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解压请求体")
			return
		}
		gr, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			handler.WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解压请求体")
			return
		}
		raw, err := io.ReadAll(gr)
		_ = gr.Close()
		_ = c.Request.Body.Close()
		if err != nil {
			handler.WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解压请求体")
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(raw))
		c.Request.ContentLength = int64(len(raw))
		c.Request.Header.Del("Content-Encoding")
		c.Request.Form = nil
		c.Request.PostForm = nil
		c.Next()
	}
}

// RequestID 注入请求 ID，便于日志关联；响应头始终携带与 body 一致的 X-Request-ID，并回写与之一致的备用响应头（见 headerRequestIDLegacy），便于仅识别备用头的客户端；入站时主头优先，主头为空则读取备用头。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := strings.TrimSpace(c.GetHeader(headerRequestID))
		if rid == "" {
			rid = strings.TrimSpace(c.GetHeader(headerRequestIDLegacy))
		}
		if rid == "" {
			rid = randomID()
		}
		c.Writer.Header().Set(headerRequestID, rid)
		c.Writer.Header().Set(headerRequestIDLegacy, rid)
		c.Set("request_id", rid)
		c.Request = c.Request.WithContext(requestid.WithID(c.Request.Context(), rid))
		c.Next()
	}
}

// AcceptLanguage 根据 Accept-Language 归约语言标签（zh-CN / en），写入 Gin 与 Request.Context，供后续处理器或日志使用。
func AcceptLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		tag := locale.NormalizeFromAcceptLanguage(c.GetHeader("Accept-Language"))
		c.Set(locale.GinKey, tag)
		c.Request = c.Request.WithContext(locale.WithLocale(c.Request.Context(), tag))
		c.Next()
	}
}

// ZapHTTPAccessLog 在请求处理结束后写入一条 Zap 结构化访问日志（状态码、耗时、方法、路径、客户端 IP、request_id）；不记录 query、body 或鉴权头。
func ZapHTTPAccessLog(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if log == nil {
			return
		}
		st := c.Writer.Status()
		if st == 0 {
			st = http.StatusOK
		}
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		log.Info("http_access",
			zap.String("request_id", c.GetString("request_id")),
			zap.Int("status", st),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("client_ip", c.ClientIP()),
			zap.Int64("latency_ms", time.Since(start).Milliseconds()),
		)
	}
}

// RootStrictNoCache 当请求的 RequestURI 严格等于 "/" 时，在调用后续处理链之前写入 Cache-Control: no-cache，
// 避免中间层或浏览器将根路径响应当作可长期复用资源；带查询串的根请求（如 "/?x=1"）不在此列。若后续处理器改写 Cache-Control，以最终响应为准。
func RootStrictNoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/" {
			c.Header("Cache-Control", rootStrictNoCacheControl)
		}
		c.Next()
	}
}

// UploadsStaticCache 在 `/uploads/` 前缀的 GET、HEAD 成功响应上补充 Cache-Control，便于浏览器复用已拉取的公开静态文件；若响应已含非空 Cache-Control、或状态非成功（2xx）且非 304，则不写入。
func UploadsStaticCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		m := c.Request.Method
		if m != http.MethodGet && m != http.MethodHead {
			return
		}
		path := c.Request.URL.Path
		if len(path) < len("/uploads/") || !strings.HasPrefix(path, "/uploads/") {
			return
		}
		if strings.TrimSpace(c.Writer.Header().Get("Cache-Control")) != "" {
			return
		}
		st := c.Writer.Status()
		if st == 0 {
			st = http.StatusOK
		}
		if st != http.StatusNotModified && (st < http.StatusOK || st >= http.StatusMultipleChoices) {
			return
		}
		c.Writer.Header().Set("Cache-Control", uploadsStaticCacheControl)
	}
}

// ZapRecovery 捕获 panic 并写入 Zap，返回 JSON 错误。
func ZapRecovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				path := c.FullPath()
				if path == "" {
					path = c.Request.URL.Path
				}
				log.Error("panic recovered",
					zap.Any("error", err),
					zap.String("request_id", c.GetString("request_id")),
					zap.String("method", c.Request.Method),
					zap.String("path", path),
					zap.String("stack", string(debug.Stack())),
				)
				if !c.Writer.Written() {
					handler.WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误")
				}
			}
		}()
		c.Next()
	}
}

// ErrorJSON 将 handler 中 c.Error 链上的最后一个错误转为 JSON。
func ErrorJSON(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last()
		log.Warn("handler error",
			zap.String("request_id", c.GetString("request_id")),
			zap.String("path", c.FullPath()),
			zap.Error(err),
		)
		if c.Writer.Written() {
			return
		}
		handler.WriteGatewayError(c, http.StatusBadRequest, "REQUEST_ERROR", err.Error())
	}
}
