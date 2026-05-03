package router

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/handler"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/locale"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/requestid"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
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
	assert.Equal(t, got, rec.Header().Get(headerRequestIDLegacy), "legacy request id header mirrors primary")
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
	assert.Equal(t, "client-rid-1", rec.Header().Get(headerRequestID))
	assert.Equal(t, "client-rid-1", rec.Header().Get(headerRequestIDLegacy))
}

func TestRequestID_LegacyInboundWhenPrimaryAbsent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/t", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("request_id"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.Header.Set(headerRequestIDLegacy, "  legacy-rid-1  ")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "legacy-rid-1", rec.Body.String())
	assert.Equal(t, "legacy-rid-1", rec.Header().Get(headerRequestID))
	assert.Equal(t, "legacy-rid-1", rec.Header().Get(headerRequestIDLegacy))
}

func TestRequestID_PrimaryInboundBeatsLegacy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/t", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("request_id"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.Header.Set(headerRequestID, "main-pick")
	req.Header.Set(headerRequestIDLegacy, "other-legacy")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "main-pick", rec.Body.String())
	assert.Equal(t, "main-pick", rec.Header().Get(headerRequestID))
	assert.Equal(t, "main-pick", rec.Header().Get(headerRequestIDLegacy))
}

func TestRequestID_ContextCarriesID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/t", func(c *gin.Context) {
		child, cancel := context.WithTimeout(c.Request.Context(), time.Minute)
		defer cancel()
		c.String(http.StatusOK, requestid.FromContext(child))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.Header.Set(headerRequestID, "ctx-rid-1")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ctx-rid-1", rec.Body.String())
}

func TestAcceptLanguage_SetsGinAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.Use(AcceptLanguage())
	r.GET("/t", func(c *gin.Context) {
		child, cancel := context.WithTimeout(c.Request.Context(), time.Minute)
		defer cancel()
		c.String(http.StatusOK, c.GetString(locale.GinKey)+"|"+locale.FromContext(child)+"|"+requestid.FromContext(child))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.Header.Set(headerRequestID, "rid-loc")
	req.Header.Set("Accept-Language", "zh-CN,en;q=0.9")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, locale.TagZH+"|"+locale.TagZH+"|rid-loc", rec.Body.String())
}

func TestZapHTTPAccessLog_StructuredFields(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	log := zap.New(core)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.Use(ZapHTTPAccessLog(log))
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(headerRequestID, "access-rid-1")
	req.RemoteAddr = "192.0.2.1:1234"
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	entries := logs.All()
	require.Len(t, entries, 1)
	assert.Equal(t, "http_access", entries[0].Message)
	m := entries[0].ContextMap()
	assert.Equal(t, "access-rid-1", m["request_id"])
	assert.EqualValues(t, 204, m["status"])
	assert.Equal(t, http.MethodGet, m["method"])
	assert.Equal(t, "/health", m["path"])
	assert.NotNil(t, m["latency_ms"])
}

func TestAcceptLanguage_DefaultEnglish(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AcceptLanguage())
	r.GET("/t", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString(locale.GinKey))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, locale.TagEN, rec.Body.String())
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
	var panicBody map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &panicBody))
	assert.NotEmpty(t, panicBody["request_id"])
	assert.Equal(t, panicBody["request_id"], rec.Header().Get(headerRequestID))
}

func TestZapRecovery_LogsMethodAndPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	core, observed := observer.New(zap.ErrorLevel)
	log := zap.New(core)
	r := gin.New()
	r.Use(RequestID())
	r.Use(ZapRecovery(log))
	r.GET("/panic", func(c *gin.Context) { panic("boom") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	entries := observed.All()
	require.Len(t, entries, 1)
	assert.Equal(t, "panic recovered", entries[0].Message)
	fields := entries[0].ContextMap()
	assert.Equal(t, http.MethodGet, fields["method"])
	assert.Equal(t, "/panic", fields["path"])
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
	var errBody map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errBody))
	assert.NotEmpty(t, errBody["request_id"])
	assert.Equal(t, errBody["request_id"], rec.Header().Get(headerRequestID))
}

func TestGzipRequestDecode_DecompressesAndStripsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	plain := []byte(`{"hello":"world"}`)
	var gzbuf bytes.Buffer
	gw := gzip.NewWriter(&gzbuf)
	_, err := gw.Write(plain)
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	r := gin.New()
	r.Use(RequestID())
	r.Use(GzipRequestDecode())
	r.POST("/echo", func(c *gin.Context) {
		assert.Equal(t, "", c.GetHeader("Content-Encoding"))
		assert.Equal(t, int64(len(plain)), c.Request.ContentLength)
		b, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		c.Data(http.StatusOK, "application/json", b)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader(gzbuf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, string(plain), rec.Body.String())
}

func TestGzipRequestDecode_HeaderCaseInsensitive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	plain := []byte(`ok`)
	var gzbuf bytes.Buffer
	gw := gzip.NewWriter(&gzbuf)
	_, _ = gw.Write(plain)
	require.NoError(t, gw.Close())

	r := gin.New()
	r.Use(RequestID())
	r.Use(GzipRequestDecode())
	r.POST("/t", func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		c.String(http.StatusOK, string(b))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/t", bytes.NewReader(gzbuf.Bytes()))
	req.Header.Set("Content-Encoding", "GZIP")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestGzipRequestDecode_InvalidGzipReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.Use(GzipRequestDecode())
	r.POST("/t", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/t", bytes.NewReader([]byte("not-gzip")))
	req.Header.Set("Content-Encoding", "gzip")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "INVALID_REQUEST", body["code"])
}

func TestGzipRequestDecode_NoEncodingPassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	raw := []byte(`plain`)
	r := gin.New()
	r.Use(GzipRequestDecode())
	r.POST("/t", func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		c.Data(http.StatusOK, "text/plain", b)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/t", bytes.NewReader(raw))
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "plain", rec.Body.String())
}

func TestWriteGatewayError_ClientRequestIDPreserved(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/x", func(c *gin.Context) {
		handler.WriteGatewayError(c, http.StatusTeapot, "TEAPOT", "no")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(headerRequestID, "rid-xyz")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusTeapot, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "rid-xyz", body["request_id"])
	assert.Equal(t, "rid-xyz", rec.Header().Get(headerRequestID))
}

func TestRootStrictNoCache_SetsOnBareSlash(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RootStrictNoCache())
	r.NoRoute(func(c *gin.Context) { c.Status(http.StatusTeapot) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, rootStrictNoCacheControl, rec.Header().Get("Cache-Control"))
}

func TestRootStrictNoCache_SkipsWhenQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RootStrictNoCache())
	r.NoRoute(func(c *gin.Context) { c.Status(http.StatusTeapot) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?x=1", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Empty(t, rec.Header().Get("Cache-Control"))
}

func TestUploadsStaticCache_AddsHeaderOn200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(UploadsStaticCache())
	r.GET("/uploads/vendor-logos/x.svg", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", []byte("<svg/>"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/uploads/vendor-logos/x.svg", nil)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, uploadsStaticCacheControl, rec.Header().Get("Cache-Control"))
}

func TestUploadsStaticCache_SkipsWhenAlreadySet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(UploadsStaticCache())
	r.GET("/uploads/a.png", func(c *gin.Context) {
		c.Header("Cache-Control", "private, max-age=60")
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/uploads/a.png", nil)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "private, max-age=60", rec.Header().Get("Cache-Control"))
}

func TestUploadsStaticCache_Skips404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(UploadsStaticCache())
	r.GET("/uploads/missing", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/uploads/missing", nil)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get("Cache-Control"))
}

func TestUploadsStaticCache_SkipsNonUploadsPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(UploadsStaticCache())
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Cache-Control"))
}
