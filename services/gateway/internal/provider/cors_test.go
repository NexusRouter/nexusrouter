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
	ks, err := keystore.New(cfg, log, nil)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	e := ProvideEngine(log, cfg, ks, rt, ProvideMetrics(), nil)

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
	ks, err := keystore.New(cfg, log, nil)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	e := ProvideEngine(log, cfg, ks, rt, ProvideMetrics(), nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	e.ServeHTTP(rec, req)
	require.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestProvideEngine_CORS_PreflightAfterYAMLReload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "gateway.yaml")
	err := os.WriteFile(yamlPath, []byte(`cors:
  enabled: true
  allow_origins:
    - https://a.example.com
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
	ks, err := keystore.New(cfg, log, nil)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	e := ProvideEngine(log, cfg, ks, rt, ProvideMetrics(), nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	req.Header.Set("Origin", "https://a.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	e.ServeHTTP(rec, req)
	require.Equal(t, "https://a.example.com", rec.Header().Get("Access-Control-Allow-Origin"))

	err = os.WriteFile(yamlPath, []byte(`cors:
  enabled: true
  allow_origins:
    - https://b.example.com
  allow_methods: [GET, POST, OPTIONS]
  allow_headers: [Authorization, Content-Type]
`), 0o600)
	require.NoError(t, err)
	require.NoError(t, rt.Reload())

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	req2.Header.Set("Origin", "https://b.example.com")
	req2.Header.Set("Access-Control-Request-Method", "POST")
	e.ServeHTTP(rec2, req2)
	require.Equal(t, "https://b.example.com", rec2.Header().Get("Access-Control-Allow-Origin"))

	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	req3.Header.Set("Origin", "https://a.example.com")
	req3.Header.Set("Access-Control-Request-Method", "POST")
	e.ServeHTTP(rec3, req3)
	require.Empty(t, rec3.Header().Get("Access-Control-Allow-Origin"))
}

func TestProvideEngine_EnvOnlyNoGatewayYAML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	cfg := &config.Config{
		EnableSwaggerUI: false,
		GatewayAPIKeys:  []string{"k"},
	}
	ks, err := keystore.New(cfg, log, nil)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	e := ProvideEngine(log, cfg, ks, rt, ProvideMetrics(), nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestProvideEngine_CORS_Disabled_LocalhostPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	cfg := &config.Config{
		EnableSwaggerUI: false,
		GatewayAPIKeys:  []string{"k"},
	}
	ks, err := keystore.New(cfg, log, nil)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	e := ProvideEngine(log, cfg, ks, rt, ProvideMetrics(), nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/bootstrap/v1/complete", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "http://localhost:5173", rec.Header().Get("Access-Control-Allow-Origin"))

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodOptions, "/api/bootstrap/v1/complete", nil)
	req2.Header.Set("Origin", "https://evil.example.com")
	req2.Header.Set("Access-Control-Request-Method", "POST")
	e.ServeHTTP(rec2, req2)
	require.Empty(t, rec2.Header().Get("Access-Control-Allow-Origin"))
}
