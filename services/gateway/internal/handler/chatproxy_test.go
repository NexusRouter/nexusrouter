package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testKeyStore(t *testing.T, keys ...string) *keystore.Store {
	t.Helper()
	s, err := keystore.New(&config.Config{GatewayAPIKeys: keys}, zap.NewNop())
	require.NoError(t, err)
	return s
}

func testRuntimeStore(t *testing.T, cfg *config.Config) *runtime.Store {
	t.Helper()
	rt, err := runtime.NewStore(cfg)
	require.NoError(t, err)
	return rt
}

func TestGatewayAuth_XAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "gw-secret"), nil), func(c *gin.Context) {
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
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "secret"), nil), func(c *gin.Context) {
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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil))

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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil))

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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"m","messages":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadGateway, rec.Code)
	var ge GatewayErrorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &ge))
	require.Equal(t, "BAD_GATEWAY", ge.Code)
	require.NotEmpty(t, ge.RequestID)
}

func TestChatProxy_RoundRobinTwoUpstreams(t *testing.T) {
	var countA, countB atomic.Int32
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		countA.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"from":"a"}`))
	}))
	defer s1.Close()
	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		countB.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"from":"b"}`))
	}))
	defer s2.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURLs: []string{s1.URL, s2.URL},
		GatewayAPIKeys:   []string{"sk-gw"},
		UpstreamTimeout:  5 * time.Second,
	}
	r := gin.New()
	rt := testRuntimeStore(t, cfg)
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), rt, nil))

	for i := 0; i < 4; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[]}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk-gw")
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}
	require.Equal(t, int32(2), countA.Load())
	require.Equal(t, int32(2), countB.Load())
}

func TestChatProxy_ForwardsCustomHeader(t *testing.T) {
	var gotUA, gotCustom string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotCustom = r.Header.Get("X-Custom-Client")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	req.Header.Set("User-Agent", "nexus-test-ua")
	req.Header.Set("X-Custom-Client", "foo")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "nexus-test-ua", gotUA)
	require.Equal(t, "foo", gotCustom)
}
