package router

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRequestID_GeneratesHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/t", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	got := rec.Header().Get(headerRequestID)
	require.NotEmpty(t, got)
}

func TestRequestID_PreservesClientHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	// 客户端传入的 ID 写入上下文；未强制写回响应头（与仅生成时写头一致）。
	r.GET("/t", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("request_id"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.Header.Set(headerRequestID, "client-rid-1")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "client-rid-1", rec.Body.String())
}

func TestZapRecovery_ReturnsJSONOnPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	r := gin.New()
	r.Use(RequestID())
	r.Use(ZapRecovery(log))
	r.GET("/panic", func(c *gin.Context) { panic("boom") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "INTERNAL_ERROR")
	assert.Contains(t, rec.Body.String(), "服务器内部错误")
}

func TestErrorJSON_HandlerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	r := gin.New()
	r.Use(RequestID())
	r.Use(ErrorJSON(log))
	r.GET("/bad", func(c *gin.Context) {
		_ = c.Error(errors.New("bad request"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "REQUEST_ERROR")
}
