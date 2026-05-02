package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// 对应 model-library：GET /v1/models 鉴权与空列表；非法方法 405。
func TestRegister_Models_Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	cfg := &config.Config{GatewayAPIKeys: []string{"gw-secret"}}
	ks, err := keystore.New(cfg, zap.NewNop(), nil)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	Register(e, Deps{Config: cfg, Log: zap.NewNop(), KeyStore: ks, Runtime: rt})

	t.Run("POST /v1/models 405", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/models", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.Equal(t, "METHOD_NOT_ALLOWED", body["code"])
	})

	t.Run("GET /v1/models 无凭证 401", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("GET /v1/models 有凭证 200 空列表", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		req.Header.Set("Authorization", "Bearer gw-secret")
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.Equal(t, "list", body["object"])
		data, ok := body["data"].([]any)
		require.True(t, ok)
		require.Len(t, data, 0)
	})

	t.Run("GET /v1/models/unknown 404", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/models/does-not-exist", nil)
		req.Header.Set("Authorization", "Bearer gw-secret")
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}
