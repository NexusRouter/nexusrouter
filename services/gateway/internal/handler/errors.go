package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const headerRequestID = "X-Request-ID"

// GatewayErrorBody 网关统一错误 JSON 形状（与 WriteGatewayError 一致）。
type GatewayErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func randomRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// WriteGatewayError 写入网关统一 JSON 错误（含 request_id，与 X-Request-ID 响应头一致）。
// WriteOpenAINotFoundPath 对未匹配任何已注册路由、且路径属于 OpenAI 兼容 v1 命名空间（**`/v1`** 或 **`/v1/…`**）的请求写入 **404**，body 为常见 **`error`** 对象（**`type`** 为 **`invalid_request_error`**），便于沿用 OpenAI 客户端错误解析逻辑；**`X-Request-ID`** 与 Gin 中的 **`request_id`** 在已存在时保持一致。
func WriteOpenAINotFoundPath(c *gin.Context) {
	rid := c.GetString("request_id")
	if rid == "" {
		rid = c.GetHeader(headerRequestID)
	}
	if rid == "" {
		rid = randomRequestID()
	}
	c.Writer.Header().Set(headerRequestID, rid)
	c.Set("request_id", rid)
	msg := fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path)
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
		"error": gin.H{
			"message": msg,
			"type":    "invalid_request_error",
			"param":   "",
			"code":    "",
		},
	})
}

func WriteGatewayError(c *gin.Context, status int, code, message string) {
	rid := c.GetString("request_id")
	if rid == "" {
		rid = c.GetHeader(headerRequestID)
	}
	if rid == "" {
		rid = randomRequestID()
	}
	c.Writer.Header().Set(headerRequestID, rid)
	c.Set("gateway_error_code", code)
	c.AbortWithStatusJSON(status, gin.H{
		"code":       code,
		"message":    message,
		"request_id": rid,
	})
}

// WriteGatewayErrorHTTP 在无 Gin 上下文场景（如 ReverseProxy.ErrorHandler）写入统一 JSON。
func WriteGatewayErrorHTTP(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	rid := r.Header.Get(headerRequestID)
	if rid == "" {
		rid = randomRequestID()
	}
	w.Header().Set(headerRequestID, rid)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":       code,
		"message":    message,
		"request_id": rid,
	})
}
