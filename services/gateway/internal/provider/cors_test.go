package provider

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestProvideEngine_CORS_PreflightAllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "gateway.yaml")
	err := os.WriteFile(yamlPath, []byte(`cors:
  enabled: true
  allow_origins:
    - https://app.example.com
  allow_methods: [GET, POST, OPTIONS]
  allow_headers: [Authorization, Content-Type]
`), 0o600)
	require.NoError(t, err)

	log := zap.NewNop()
	cfg := &config.Config{
		EnableSwaggerUI:   false,
		GatewayAPIKeys:    []string{"k"},
		GatewayConfigFile: yamlPath,
	}
	ks, err := keystore.New(cfg, log)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg)
	require.NoError(t, err)
	e := ProvideEngine(log, cfg, ks, rt)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "https://app.example.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestProvideEngine_CORS_DisallowedOriginNoAllowOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "gateway.yaml")
	err := os.WriteFile(yamlPath, []byte(`cors:
  enabled: true
  allow_origins:
    - https://trusted.example.com
`), 0o600)
	require.NoError(t, err)

	log := zap.NewNop()
	cfg := &config.Config{
		EnableSwaggerUI:   false,
		GatewayAPIKeys:    []string{"k"},
		GatewayConfigFile: yamlPath,
	}
	ks, err := keystore.New(cfg, log)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg)
	require.NoError(t, err)
	e := ProvideEngine(log, cfg, ks, rt)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	e.ServeHTTP(rec, req)
	require.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestProvideEngine_EnvOnlyNoGatewayYAML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	cfg := &config.Config{
		EnableSwaggerUI: false,
		GatewayAPIKeys:  []string{"k"},
	}
	ks, err := keystore.New(cfg, log)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg)
	require.NoError(t, err)
	e := ProvideEngine(log, cfg, ks, rt)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
