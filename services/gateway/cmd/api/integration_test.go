package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 接口测试：Wire 装配后的完整应用（与 spec「入口 + /health」一致）。
func TestInitializeApp_HealthEndpoint(t *testing.T) {
	app, err := InitializeApp()
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotNil(t, app.Engine)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	app.Engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
	assert.NotEmpty(t, body["version"])
	assert.NotEmpty(t, body["server_time"])
	assert.NotEmpty(t, body["start_time"])
	assert.Contains(t, body, "uptime_seconds")

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodHead, "/health", nil)
	app.Engine.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusOK, rec2.Code)
	assert.Empty(t, rec2.Body.String())
	assert.NotEmpty(t, rec2.Header().Get("Content-Length"))
}

// 公开 GET /api/status：Wire 装配后 success/data 封装与 200。
func TestInitializeApp_APIStatusEndpoint(t *testing.T) {
	app, err := InitializeApp()
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotNil(t, app.Engine)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	app.Engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, true, body["success"])
	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	assert.NotEmpty(t, data["version"])
	assert.NotEmpty(t, data["start_time"])
}

func completeFirstBootViaAPI(t *testing.T, appEngine http.Handler, username, password string) {
	t.Helper()
	body := fmt.Sprintf(
		`{"admin_username":%q,"admin_password":%q,"site_display_name":"integration"}`,
		username, password,
	)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/bootstrap/v1/complete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	appEngine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

// 管理控制台：错误密码、无令牌访问受保护接口、登录后拉取指标。
func TestInitializeApp_AdminConsoleAuth(t *testing.T) {
	t.Setenv("NEXUSROUTER_SQLITE_PATH", filepath.Join(t.TempDir(), "admin-console.db"))
	t.Setenv("NEXUSROUTER_ENABLE_ADMIN_CONSOLE", "true")
	t.Setenv("NEXUSROUTER_ADMIN_JWT_SECRET", "unit-test-jwt-secret-min-32-chars!!")

	app, err := InitializeApp()
	require.NoError(t, err)
	completeFirstBootViaAPI(t, app.Engine, "admin", "correct-password")

	// 无令牌访问指标
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/v1/metrics/summary", nil)
	app.Engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	// 错误密码
	body := `{"username":"admin","password":"wrong"}`
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/admin/v1/auth/login", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	app.Engine.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusUnauthorized, rec2.Code)

	// 正确登录
	bodyOK := `{"username":"admin","password":"correct-password"}`
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/api/admin/v1/auth/login", strings.NewReader(bodyOK))
	req3.Header.Set("Content-Type", "application/json")
	app.Engine.ServeHTTP(rec3, req3)
	require.Equal(t, http.StatusOK, rec3.Code)
	var login map[string]any
	require.NoError(t, json.Unmarshal(rec3.Body.Bytes(), &login))
	tok, tokOK := login["access_token"].(string)
	require.True(t, tokOK, "access_token 类型应为 string")
	require.NotEmpty(t, tok)

	rec4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodGet, "/api/admin/v1/metrics/summary", nil)
	req4.Header.Set("Authorization", "Bearer "+tok)
	app.Engine.ServeHTTP(rec4, req4)
	require.Equal(t, http.StatusOK, rec4.Code)
	var metrics map[string]any
	require.NoError(t, json.Unmarshal(rec4.Body.Bytes(), &metrics))
	assert.Contains(t, metrics, "requests_total")
}

// 未匹配路由：配置外置前端基址时 301 到基址与请求 URI 拼接。
func TestInitializeApp_NoRouteFrontendRedirect(t *testing.T) {
	t.Setenv("NEXUSROUTER_SQLITE_PATH", filepath.Join(t.TempDir(), "fe-redirect.db"))
	t.Setenv("NEXUSROUTER_ENABLE_ADMIN_CONSOLE", "true")
	t.Setenv("NEXUSROUTER_ADMIN_JWT_SECRET", "unit-test-jwt-secret-min-32-chars!!")
	t.Setenv("NEXUSROUTER_FRONTEND_BASE_URL", "https://ui.example.test")

	app, err := InitializeApp()
	require.NoError(t, err)
	completeFirstBootViaAPI(t, app.Engine, "admin", "correct-password")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/app/dashboard?x=1", nil)
	app.Engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "https://ui.example.test/app/dashboard?x=1", rec.Header().Get("Location"))
}

// 忘记密码说明接口无需登录。
func TestInitializeApp_AdminPasswordResetInfo(t *testing.T) {
	t.Setenv("NEXUSROUTER_SQLITE_PATH", filepath.Join(t.TempDir(), "pw-reset.db"))
	t.Setenv("NEXUSROUTER_ENABLE_ADMIN_CONSOLE", "true")
	t.Setenv("NEXUSROUTER_ADMIN_JWT_SECRET", "unit-test-jwt-secret-min-32-chars!!")

	app, err := InitializeApp()
	require.NoError(t, err)
	completeFirstBootViaAPI(t, app.Engine, "admin", "bootstrap-pass-8")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/v1/auth/password-reset-info", nil)
	app.Engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Contains(t, body, "hint")
}

// 无效 JWT 访问管理接口。
func TestInitializeApp_AdminInvalidToken(t *testing.T) {
	t.Setenv("NEXUSROUTER_SQLITE_PATH", filepath.Join(t.TempDir(), "invalid-jwt.db"))
	t.Setenv("NEXUSROUTER_ENABLE_ADMIN_CONSOLE", "true")
	t.Setenv("NEXUSROUTER_ADMIN_JWT_SECRET", "unit-test-jwt-secret-min-32-chars!!")

	app, err := InitializeApp()
	require.NoError(t, err)
	completeFirstBootViaAPI(t, app.Engine, "admin", "some-pass-word-8")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/v1/metrics/summary", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-jwt")
	app.Engine.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
