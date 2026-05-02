package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestOperatorCannotPutGatewayConfig(t *testing.T) {
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
		EnableSwaggerUI:             false,
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
	require.NotNil(t, auth)

	e := gin.New()
	RegisterAdminConsole(e, cfg, auth, metrics.NewCollector(), rt, ks, log, nil)

	tok, _, role, err := auth.Login("operator1", opPass, false)
	require.NoError(t, err)
	require.Equal(t, "operator", role)

	body := `{"upstreams":[{"id":"a","base_url":"https://example.com","weight":1}],"routing":{"strategy":"round_robin","default_upstream_id":"a","active_upstream_id":""},"persist":false}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/v1/gateway/config", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestOperatorCanGetMetrics(t *testing.T) {
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
		EnableSwaggerUI:             false,
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
	col := metrics.NewCollector()
	RegisterAdminConsole(e, cfg, auth, col, rt, ks, log, nil)
	tok, _, _, err := auth.Login("operator1", opPass, false)
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/v1/metrics/summary", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}
