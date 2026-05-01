package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
}
