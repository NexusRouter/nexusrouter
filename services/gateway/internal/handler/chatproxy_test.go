package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func testKeyStore(t *testing.T, keys ...string) *keystore.Store {
	t.Helper()
	s, err := keystore.New(&config.Config{GatewayAPIKeys: keys}, zap.NewNop(), nil)
	require.NoError(t, err)
	return s
}

func testRuntimeStore(t *testing.T, cfg *config.Config) *runtime.Store {
	t.Helper()
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	return rt
}

func TestMergeChatStreamIncludeUsage_StreamTrueAddsOption(t *testing.T) {
	in := []byte(`{"model":"m","messages":[{"role":"user","content":"x"}],"stream":true}`)
	out := mergeChatStreamIncludeUsage(in)
	var obj map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(out, &obj))
	require.Contains(t, obj, "stream_options")
	var so map[string]any
	require.NoError(t, json.Unmarshal(obj["stream_options"], &so))
	require.Equal(t, true, so["include_usage"])
}

func TestMergeChatStreamIncludeUsage_PreservesOtherStreamOptions(t *testing.T) {
	in := []byte(`{"model":"m","stream":true,"stream_options":{"custom_flag":true}}`)
	out := mergeChatStreamIncludeUsage(in)
	var obj map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(out, &obj))
	var so map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(obj["stream_options"], &so))
	require.Equal(t, json.RawMessage("true"), so["custom_flag"])
	require.Equal(t, json.RawMessage("true"), so["include_usage"])
}

func TestMergeChatStreamIncludeUsage_NonStreamUnchanged(t *testing.T) {
	in := []byte(`{"model":"m","messages":[{"role":"user","content":"x"}]}`)
	assert.Equal(t, string(in), string(mergeChatStreamIncludeUsage(in)))
}

func TestValidateChatCompletionsMaxTokens(t *testing.T) {
	ok := func(raw string) {
		t.Helper()
		err := validateChatCompletionsMaxTokens([]byte(raw))
		assert.NoError(t, err)
	}
	bad := func(raw string, subs string) {
		t.Helper()
		err := validateChatCompletionsMaxTokens([]byte(raw))
		require.Error(t, err)
		assert.Contains(t, err.Error(), subs)
	}

	ok(`{"model":"x","messages":[]}`)
	ok(`{"model":"x","messages":[],"max_tokens":null}`)
	ok(`{"model":"x","messages":[],"max_tokens":0}`)
	ok(`{"model":"x","messages":[],"max_tokens":100}`)
	maxOK := int64(math.MaxInt32 / 2)
	ok(`{"model":"x","messages":[],"max_tokens":` + strconv.FormatInt(maxOK, 10) + `}`)

	bad(`{"model":"x","messages":[],"max_tokens":-1}`, "范围")
	bad(`{"model":"x","messages":[],"max_tokens":`+strconv.FormatInt(maxOK+1, 10)+`}`, "范围")
	bad(`{"model":"x","messages":[],"max_tokens":1.5}`, "整数")
	bad(`{"model":"x","messages":[],"max_tokens":"10"}`, "整数")
	bad(`{"model":"x","messages":[],"max_tokens":false}`, "整数")
}

func TestValidateChatCompletionsMessages(t *testing.T) {
	ok := func(raw string) {
		t.Helper()
		assert.NoError(t, validateChatCompletionsMessages([]byte(raw)))
	}
	bad := func(raw string, subs string) {
		t.Helper()
		err := validateChatCompletionsMessages([]byte(raw))
		require.Error(t, err)
		assert.Contains(t, err.Error(), subs)
	}

	ok(`{"model":"x","messages":[{"role":"user","content":"h"}]}`)
	ok(`not-json`) // 非对象 JSON 不因本规则拦截
	bad(`{}`, "缺少")
	bad(`{"model":"x"}`, "缺少")
	bad(`{"messages":null}`, "null")
	bad(`{"messages":[]}`, "至少")
	ok(`{"messages":[{}]}`)
	ok(`{"messages":[{"role":"user","content":"h"}]}`)
	bad(`{"messages":"x"}`, "数组")
}

func TestChatProxy_InvalidMaxTokensDoesNotHitUpstream(t *testing.T) {
	var hits atomic.Int32
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[],"max_tokens":-1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, int32(0), hits.Load())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "INVALID_REQUEST", body["code"])
}

func TestChatProxy_EmptyMessagesDoesNotHitUpstream(t *testing.T) {
	var hits atomic.Int32
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, int32(0), hits.Load())
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "INVALID_REQUEST", body["code"])
}

func TestChatProxy_StreamMergesIncludeUsageForUpstream(t *testing.T) {
	var gotBody string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		gotBody = string(b)
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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[{"role":"user","content":"x"}],"stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(gotBody), &parsed))
	so, ok := parsed["stream_options"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, so["include_usage"])
}

func TestChatProxy_StreamIncludeUsageDisabledNoBodyChange(t *testing.T) {
	var gotBody string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	off := false
	cfg := &config.Config{
		UpstreamBaseURL:        up.URL,
		GatewayAPIKeys:         []string{"sk-gw"},
		UpstreamTimeout:        5 * time.Second,
		ChatStreamIncludeUsage: &off,
	}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	raw := `{"model":"m","messages":[{"role":"user","content":"x"}],"stream":true}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, raw, gotBody)
}

func TestNewChatUpstreamTransport_ExplicitProxy(t *testing.T) {
	tr := newChatUpstreamTransport(&config.Config{UpstreamHTTPProxy: "http://127.0.0.1:8888"}, time.Minute)
	require.NotNil(t, tr.Proxy)
	u, err := tr.Proxy(&http.Request{})
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "127.0.0.1:8888", u.Host)
}

func TestNewChatUpstreamTransport_NoProxyWhenUnset(t *testing.T) {
	tr := newChatUpstreamTransport(&config.Config{}, time.Minute)
	assert.Nil(t, tr.Proxy)
	trNilCfg := newChatUpstreamTransport(nil, time.Minute)
	assert.Nil(t, trNilCfg.Proxy)
}

func TestNewChatUpstreamTransport_InvalidProxyIgnored(t *testing.T) {
	for _, raw := range []string{"not-a-url", "http:///nohost", ":missing-scheme"} {
		tr := newChatUpstreamTransport(&config.Config{UpstreamHTTPProxy: raw}, time.Second)
		assert.Nil(t, tr.Proxy, "raw=%q", raw)
	}
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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	body := `{"model":"m","messages":[{"role":"user","content":"hi"}]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"id":"x","object":"chat.completion"}`, rec.Body.String())
	assert.Empty(t, rec.Result().Header.Get("X-Accel-Buffering"))
}

func TestChatProxy_UpstreamRequestOmitsAcceptEncoding(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Accept-Encoding"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[{"role":"user","content":"h"}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestChatProxy_SSEAddsDownstreamBufferingHint(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data: {}\n\n"))
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[{"role":"user","content":"x"}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "no", rec.Result().Header.Get("X-Accel-Buffering"))
	require.Equal(t, "no-cache", rec.Result().Header.Get("Cache-Control"))
	require.Equal(t, "keep-alive", rec.Result().Header.Get("Connection"))
	assert.Contains(t, rec.Body.String(), "data:")
}

func TestChatProxy_SSEPreservesUpstreamCacheControl(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "private, no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data: {}\n\n"))
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[{"role":"user","content":"x"}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "private, no-store", rec.Result().Header.Get("Cache-Control"))
	require.Equal(t, "no", rec.Result().Header.Get("X-Accel-Buffering"))
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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"m","messages":[{"role":"user","content":"x"}]}`))
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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(`{"model":"m","messages":[{"role":"user","content":"x"}]}`))
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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), rt, nil, nil))

	for i := 0; i < 4; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[{"role":"user","content":"x"}]}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk-gw")
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}
	require.Equal(t, int32(2), countA.Load())
	require.Equal(t, int32(2), countB.Load())
}

func TestChatProxy_RewritesModelFromModelLibraryBinding(t *testing.T) {
	var gotBody string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer up.Close()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, repository.AutoMigrate(db))
	require.NoError(t, repository.CreateModelCatalogEntry(db, &repository.ModelCatalogEntry{ID: "logical-m", DisplayName: "L"}))
	am := "upstream-real"
	require.NoError(t, repository.CreateModelUpstreamBinding(db, &repository.ModelUpstreamBinding{
		CatalogEntryID: "logical-m",
		UpstreamID:     "default",
		Enabled:        true,
		ActualModel:    &am,
	}))

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	r := gin.New()
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, db))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"logical-m","messages":[{"role":"user","content":"x"}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(gotBody), &parsed))
	var mid string
	require.NoError(t, json.Unmarshal(parsed["model"], &mid))
	require.Equal(t, "upstream-real", mid)
}

func TestChatProxy_StreamTrueAddsAcceptWhenAbsent(t *testing.T) {
	var gotAccept string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[{"role":"user","content":"x"}],"stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "text/event-stream", gotAccept)
}

func TestChatProxy_StreamTrueDoesNotOverrideClientAccept(t *testing.T) {
	var gotAccept string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[{"role":"user","content":"x"}],"stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	req.Header.Set("Accept", "application/json")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", gotAccept)
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
	r.POST("/v1/chat/completions", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ChatProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"m","messages":[{"role":"user","content":"x"}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	req.Header.Set("User-Agent", "nexus-test-ua")
	req.Header.Set("X-Custom-Client", "foo")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "nexus-test-ua", gotUA)
	require.Equal(t, "foo", gotCustom)
}
