package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

const headerRequestID = "X-Request-ID"

func randomRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// WriteGatewayError 写入网关统一 JSON 错误（含 request_id，与 X-Request-ID 响应头一致）。
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
