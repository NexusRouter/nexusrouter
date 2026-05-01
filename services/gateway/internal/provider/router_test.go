package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// 接口测试：与 Wire 装配一致的引擎行为（spec：/health 与未知路由 JSON）。
func TestProvideEngine_HealthAndNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	cfg := &config.Config{EnableSwaggerUI: false}
	ks, err := keystore.New(cfg, log)
	require.NoError(t, err)
	e := ProvideEngine(log, cfg, ks)

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
}
