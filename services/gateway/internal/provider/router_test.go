package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// 接口测试：与 Wire 装配一致的引擎行为（spec：/health 与未知路由 JSON）。
func TestApplyGinModeFromEnv_DebugAndRelease(t *testing.T) {
	t.Run("debug", func(t *testing.T) {
		t.Setenv("GIN_MODE", gin.DebugMode)
		gin.SetMode(gin.ReleaseMode)
		applyGinModeFromEnv()
		require.Equal(t, gin.DebugMode, gin.Mode())
	})
	t.Run("release when unset", func(t *testing.T) {
		t.Setenv("GIN_MODE", "")
		gin.SetMode(gin.DebugMode)
		applyGinModeFromEnv()
		require.Equal(t, gin.ReleaseMode, gin.Mode())
	})
	t.Run("release for other values", func(t *testing.T) {
		t.Setenv("GIN_MODE", "release")
		gin.SetMode(gin.DebugMode)
		applyGinModeFromEnv()
		require.Equal(t, gin.ReleaseMode, gin.Mode())
	})
}

func TestProvideEngine_HealthAndNotFound(t *testing.T) {
	t.Setenv("GIN_MODE", "")
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	cfg := &config.Config{}
	ks, err := keystore.New(cfg, log, nil)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	e := ProvideEngine(log, cfg, ks, rt, ProvideMetrics(), nil)

	t.Run("GET /health", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "ok", body["status"])
		assert.NotEmpty(t, body["version"])
		assert.NotEmpty(t, body["server_time"])
		assert.NotEmpty(t, body["start_time"])
		assert.Contains(t, body, "uptime_seconds")
	})

	t.Run("HEAD /health", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodHead, "/health", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Empty(t, rec.Body.String())
		assert.NotEmpty(t, rec.Header().Get("Content-Length"))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	})

	t.Run("GET /unknown JSON 404", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/unknown-route-xyz", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "NOT_FOUND", body["code"])
		assert.Equal(t, "路由不存在", body["message"])
		assert.NotEmpty(t, body["request_id"])
	})

	t.Run("GET /v1/unregistered OpenAI 形 404", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/no-such-endpoint", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		errObj, ok := body["error"].(map[string]any)
		require.True(t, ok, "body: %s", rec.Body.String())
		assert.Equal(t, "invalid_request_error", errObj["type"])
		assert.Contains(t, errObj["message"], "GET")
		assert.Contains(t, errObj["message"], "/v1/no-such-endpoint")
		assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
	})

	t.Run("GET / Cache-Control no-cache", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("GET /?q Cache-Control not from RootStrictNoCache", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/?q=1", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Code)
		assert.NotEqual(t, "no-cache", rec.Header().Get("Cache-Control"))
	})
}
