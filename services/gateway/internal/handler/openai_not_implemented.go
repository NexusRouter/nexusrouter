package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OpenAINotImplemented 返回 **501 Not Implemented** 与网关统一 JSON 错误体，用于已识别但尚未实现的 OpenAI 兼容子路径。
func OpenAINotImplemented() gin.HandlerFunc {
	return func(c *gin.Context) {
		WriteGatewayError(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "该 OpenAI 兼容接口尚未实现")
	}
}
