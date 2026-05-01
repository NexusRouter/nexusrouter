package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const openAIOverviewURL = "https://developers.openai.com/api/reference/overview"

// 对应 change tasks 1.1–1.5：OpenAPI 3.0 与 Swagger UI 契约。
func TestOpenAPI_Swagger_Contract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	cfg := &config.Config{
		EnableSwaggerUI: true,
		GatewayAPIKeys:  []string{"test-key"},
	}
	Register(e, Deps{Config: cfg, Log: zap.NewNop()})

	t.Run("GET /openapi.yaml 200 且 openapi 3.0", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		require.Contains(t, body, "openapi: 3.0")
	})

	t.Run("YAML 路径与安全", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var root map[string]any
		require.NoError(t, yaml.Unmarshal(rec.Body.Bytes(), &root))
		paths, ok := root["paths"].(map[string]any)
		require.True(t, ok)
		cc, ok := paths["/v1/chat/completions"].(map[string]any)
		require.True(t, ok)
		post, ok := cc["post"].(map[string]any)
		require.True(t, ok)
		_, hasRB := post["requestBody"]
		require.True(t, hasRB, "post 应含 requestBody")

		secRoot, _ := root["security"].([]any)
		secPost, _ := post["security"].([]any)
		require.True(t, len(secRoot) > 0 || len(secPost) > 0, "根级或操作级 security 至少一处")

		comps, ok := root["components"].(map[string]any)
		require.True(t, ok)
		schemes, ok := comps["securitySchemes"].(map[string]any)
		require.True(t, ok)
		var bearer map[string]any
		for name, v := range schemes {
			if strings.Contains(strings.ToLower(name), "bearer") {
				bearer, _ = v.(map[string]any)
				break
			}
		}
		require.NotNil(t, bearer)
		require.Equal(t, "http", bearer["type"])
		require.Equal(t, "bearer", bearer["scheme"])
	})

	t.Run("含 OpenAI overview 链接", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
		e.ServeHTTP(rec, req)
		require.Contains(t, rec.Body.String(), openAIOverviewURL)
	})

	t.Run("GET /swagger/index.html 引用 OAS3", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		html := rec.Body.String()
		require.True(t, strings.Contains(html, "url:") || strings.Contains(html, "configUrl"),
			"Swagger UI 应引用 spec URL")
	})

	t.Run("GET /openapi.json 合法", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
		e.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		var root map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &root))
		ver, _ := root["openapi"].(string)
		require.True(t, strings.HasPrefix(ver, "3.0."), ver)
	})
}
