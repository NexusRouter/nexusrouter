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

func TestMergeImagesGenerationsDefaultModel_InjectsWhenMissing(t *testing.T) {
	in := []byte(`{"prompt":"a red dot"}`)
	out := mergeImagesGenerationsDefaultModel(in)
	var o map[string]any
	require.NoError(t, json.Unmarshal(out, &o))
	assert.Equal(t, defaultImagesGenerationModel, o["model"])
	assert.Equal(t, "a red dot", o["prompt"])
}

func TestMergeImagesGenerationsDefaultModel_PreservesExisting(t *testing.T) {
	in := []byte(`{"model":"dall-e-3","prompt":"x"}`)
	out := mergeImagesGenerationsDefaultModel(in)
	assert.JSONEq(t, string(in), string(out))
}

func TestValidateImagesGenerationsRequestBody(t *testing.T) {
	require.NoError(t, validateImagesGenerationsRequestBody([]byte(`{"model":"m","prompt":"x"}`)))
	require.NoError(t, validateImagesGenerationsRequestBody([]byte(`{"prompt":"hello"}`)))

	err := validateImagesGenerationsRequestBody([]byte(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "prompt")

	err = validateImagesGenerationsRequestBody([]byte(`{"prompt":null}`))
	require.Error(t, err)

	err = validateImagesGenerationsRequestBody([]byte(`{"prompt":""}`))
	require.Error(t, err)

	err = validateImagesGenerationsRequestBody([]byte(`{"prompt":1}`))
	require.Error(t, err)

	err = validateImagesGenerationsRequestBody([]byte(`not-json`))
	require.Error(t, err)
}

func TestImagesGenerationsProxy_MissingPromptNoUpstream(t *testing.T) {
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
	r.POST("/v1/images/generations", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ImagesGenerationsProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/images/generations", strings.NewReader(`{"model":"dall-e-2"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, int32(0), hits.Load())
}

func TestImagesGenerationsProxy_ForwardsPathAndDefaultModel(t *testing.T) {
	var gotPath, gotBody string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"created":1,"data":[]}`))
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	e := gin.New()
	e.POST("/v1/images/generations", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ImagesGenerationsProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/images/generations", strings.NewReader(`{"prompt":"a blue square"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/v1/images/generations", gotPath)
	assert.Contains(t, gotBody, `"model":"dall-e-2"`)
	assert.Contains(t, gotBody, `"prompt":"a blue square"`)
}
