package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMergeEmbeddingsModelFromPathParam_InjectsWhenMissing(t *testing.T) {
	in := []byte(`{"input":"hi"}`)
	out := mergeEmbeddingsModelFromPathParam(in, "ada")
	var o map[string]any
	require.NoError(t, json.Unmarshal(out, &o))
	assert.Equal(t, "ada", o["model"])
	assert.Equal(t, "hi", o["input"])
}

func TestMergeEmbeddingsModelFromPathParam_PreservesExisting(t *testing.T) {
	in := []byte(`{"model":"babbage","input":"x"}`)
	out := mergeEmbeddingsModelFromPathParam(in, "ada")
	assert.JSONEq(t, string(in), string(out))
}

func TestValidateEmbeddingsRequestBody(t *testing.T) {
	require.NoError(t, validateEmbeddingsRequestBody([]byte(`{"model":"m","input":"x"}`)))
	require.NoError(t, validateEmbeddingsRequestBody([]byte(`{"input":["a","b"]}`)))

	err := validateEmbeddingsRequestBody([]byte(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input")

	err = validateEmbeddingsRequestBody([]byte(`{"input":null}`))
	require.Error(t, err)

	err = validateEmbeddingsRequestBody([]byte(`not-json`))
	require.Error(t, err)
}

func TestEmbeddingsProxy_MissingInputNoUpstream(t *testing.T) {
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
	r.POST("/v1/embeddings", GatewayAuth(testKeyStore(t, "sk-gw"), nil), EmbeddingsProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", strings.NewReader(`{"model":"m"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, int32(0), hits.Load())
}

func TestEmbeddingsProxy_EnginePathInjectsModel(t *testing.T) {
	var gotPath, gotBody string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list"}`))
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	e := gin.New()
	e.POST("/v1/engines/:model/embeddings", GatewayAuth(testKeyStore(t, "sk-gw"), nil), EmbeddingsProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/engines/text-embedding-ada-002/embeddings", strings.NewReader(`{"input":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/v1/engines/text-embedding-ada-002/embeddings", gotPath)
	assert.Contains(t, gotBody, `"model":"text-embedding-ada-002"`)
	assert.Contains(t, gotBody, `"input":"hello"`)
}
