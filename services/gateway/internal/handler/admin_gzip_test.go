package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/adminauth"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func TestAdminAPICompressesJSONWhenAcceptGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "gateway.yaml")
	err := os.WriteFile(yamlPath, []byte(`upstreams:
  - id: a
    base_url: https://example.com
    weight: 1
routing:
  strategy: round_robin
  default_upstream_id: a
`), 0o600)
	require.NoError(t, err)
	opPass := "op-secret-9"
	opHash, err := bcrypt.GenerateFromPassword([]byte(opPass), bcrypt.MinCost)
	require.NoError(t, err)
	admHash, err := bcrypt.GenerateFromPassword([]byte("adminpass"), bcrypt.MinCost)
	require.NoError(t, err)
	cfg := &config.Config{
		GatewayAPIKeys:              []string{"k"},
		GatewayConfigFile:           yamlPath,
		EnableAdminConsole:          true,
		AdminJWTSecret:              "test-secret-test-secret-test-32",
		AdminJWTExpire:              24 * time.Hour,
		AdminUsername:               "admin",
		AdminPasswordBcrypt:         string(admHash),
		AdminOperatorUsername:       "operator1",
		AdminOperatorPasswordBcrypt: string(opHash),
	}
	log := zap.NewNop()
	ks, err := keystore.New(cfg, log, nil)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	auth := adminauth.New(cfg, nil)
	e := gin.New()
	RegisterAdminConsole(e, cfg, auth, metrics.NewCollector(), rt, ks, log, nil)
	tok, _, _, err := auth.Login("operator1", opPass, false)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/v1/metrics/summary", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept-Encoding", "gzip")
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	gr, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	t.Cleanup(func() { _ = gr.Close() })
	raw, err := io.ReadAll(gr)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"requests_total"`)
}

func TestAdminAPINoGzipWhenClientDoesNotAccept(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "gateway.yaml")
	err := os.WriteFile(yamlPath, []byte(`upstreams:
  - id: a
    base_url: https://example.com
    weight: 1
routing:
  strategy: round_robin
  default_upstream_id: a
`), 0o600)
	require.NoError(t, err)
	opPass := "op-secret-9"
	opHash, err := bcrypt.GenerateFromPassword([]byte(opPass), bcrypt.MinCost)
	require.NoError(t, err)
	admHash, err := bcrypt.GenerateFromPassword([]byte("adminpass"), bcrypt.MinCost)
	require.NoError(t, err)
	cfg := &config.Config{
		GatewayAPIKeys:              []string{"k"},
		GatewayConfigFile:           yamlPath,
		EnableAdminConsole:          true,
		AdminJWTSecret:              "test-secret-test-secret-test-32",
		AdminJWTExpire:              24 * time.Hour,
		AdminUsername:               "admin",
		AdminPasswordBcrypt:         string(admHash),
		AdminOperatorUsername:       "operator1",
		AdminOperatorPasswordBcrypt: string(opHash),
	}
	log := zap.NewNop()
	ks, err := keystore.New(cfg, log, nil)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	auth := adminauth.New(cfg, nil)
	e := gin.New()
	RegisterAdminConsole(e, cfg, auth, metrics.NewCollector(), rt, ks, log, nil)
	tok, _, _, err := auth.Login("operator1", opPass, false)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/v1/metrics/summary", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Empty(t, rec.Header().Get("Content-Encoding"))
	require.Contains(t, rec.Body.String(), `"requests_total"`)
}
