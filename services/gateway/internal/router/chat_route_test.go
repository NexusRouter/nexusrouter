package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// 非法方法返回 405 与统一 JSON（spec 7.2）。
func TestRegister_ChatCompletions_MethodNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	Register(e, Deps{
		Config: &config.Config{EnableSwaggerUI: false, GatewayAPIKeys: []string{"k"}},
		Log:    zap.NewNop(),
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "METHOD_NOT_ALLOWED", body["code"])
}
