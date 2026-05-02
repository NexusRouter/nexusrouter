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

func TestMergeModerationsDefaultModel_InjectsWhenMissing(t *testing.T) {
	in := []byte(`{"input":"hello"}`)
	out := mergeModerationsDefaultModel(in)
	var o map[string]any
	require.NoError(t, json.Unmarshal(out, &o))
	assert.Equal(t, defaultModerationModel, o["model"])
	assert.Equal(t, "hello", o["input"])
}

func TestMergeModerationsDefaultModel_PreservesExisting(t *testing.T) {
	in := []byte(`{"model":"omni-moderation-latest","input":"x"}`)
	out := mergeModerationsDefaultModel(in)
	assert.JSONEq(t, string(in), string(out))
}

func TestValidateModerationsRequestBody(t *testing.T) {
	require.NoError(t, validateModerationsRequestBody([]byte(`{"model":"m","input":"x"}`)))
	require.NoError(t, validateModerationsRequestBody([]byte(`{"input":["a","b"]}`)))
	require.NoError(t, validateModerationsRequestBody([]byte(`{"input":{"a":1}}`)))

	err := validateModerationsRequestBody([]byte(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input")

	err = validateModerationsRequestBody([]byte(`{"input":null}`))
	require.Error(t, err)

	err = validateModerationsRequestBody([]byte(`{"input":""}`))
	require.Error(t, err)

	err = validateModerationsRequestBody([]byte(`{"input":[]}`))
	require.Error(t, err)

	err = validateModerationsRequestBody([]byte(`not-json`))
	require.Error(t, err)
}

func TestModerationsProxy_MissingInputNoUpstream(t *testing.T) {
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
	r.POST("/v1/moderations", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ModerationsProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/moderations", strings.NewReader(`{"model":"m"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, int32(0), hits.Load())
}

func TestModerationsProxy_ForwardsPathAndDefaultModel(t *testing.T) {
	var gotPath, gotBody string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"modr_1","model":"text-moderation-latest","results":[]}`))
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	e := gin.New()
	e.POST("/v1/moderations", GatewayAuth(testKeyStore(t, "sk-gw"), nil), ModerationsProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/moderations", strings.NewReader(`{"input":"check this"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/v1/moderations", gotPath)
	assert.Contains(t, gotBody, `"model":"text-moderation-latest"`)
	assert.Contains(t, gotBody, `"input":"check this"`)
}
