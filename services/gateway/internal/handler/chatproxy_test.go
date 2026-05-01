package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGatewayAuth_XAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{GatewayAPIKeys: []string{"gw-secret"}}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(cfg), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set(headerAPIKey, "gw-secret")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGatewayAuth_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{GatewayAPIKeys: []string{"secret"}}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(cfg), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestChatProxy_Upstream200_JSON(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Contains(t, string(b), `"model"`)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"x","object":"chat.completion"}`))
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(cfg), ChatProxy(cfg, zap.NewNop()))

	body := `{"model":"m","messages":[{"role":"user","content":"hi"}]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"id":"x","object":"chat.completion"}`, rec.Body.String())
}

func TestChatProxy_Upstream4xx_Passthrough(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad"}`))
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(cfg), ChatProxy(cfg, zap.NewNop()))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"m","messages":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"error":"bad"}`, rec.Body.String())
}

func TestChatProxy_UpstreamUnreachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: "http://127.0.0.1:1",
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 200 * time.Millisecond,
	}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(cfg), ChatProxy(cfg, zap.NewNop()))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"m","messages":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadGateway, rec.Code)
}
