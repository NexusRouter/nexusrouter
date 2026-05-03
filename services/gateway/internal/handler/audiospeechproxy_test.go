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

func TestMergeAudioSpeechDefaultModel_InjectsWhenMissing(t *testing.T) {
	in := []byte(`{"input":"hi","voice":"alloy"}`)
	out := mergeAudioSpeechDefaultModel(in)
	var o map[string]any
	require.NoError(t, json.Unmarshal(out, &o))
	assert.Equal(t, defaultAudioSpeechModel, o["model"])
	assert.Equal(t, "hi", o["input"])
	assert.Equal(t, "alloy", o["voice"])
}

func TestMergeAudioSpeechDefaultModel_PreservesExisting(t *testing.T) {
	in := []byte(`{"model":"tts-1-hd","input":"x","voice":"nova"}`)
	out := mergeAudioSpeechDefaultModel(in)
	assert.JSONEq(t, string(in), string(out))
}

func TestValidateAudioSpeechRequestBody(t *testing.T) {
	require.NoError(t, validateAudioSpeechRequestBody([]byte(`{"model":"tts-1","input":"a","voice":"alloy"}`)))
	require.NoError(t, validateAudioSpeechRequestBody([]byte(`{"input":"hello","voice":"echo"}`)))

	err := validateAudioSpeechRequestBody([]byte(`{"voice":"alloy"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input")

	err = validateAudioSpeechRequestBody([]byte(`{"input":"x"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "voice")

	err = validateAudioSpeechRequestBody([]byte(`{"input":"","voice":"a"}`))
	require.Error(t, err)

	err = validateAudioSpeechRequestBody([]byte(`not-json`))
	require.Error(t, err)
}

func TestAudioSpeechProxy_MissingVoiceNoUpstream(t *testing.T) {
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
	r.POST("/v1/audio/speech", GatewayAuth(testKeyStore(t, "sk-gw"), nil), AudioSpeechProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/speech", strings.NewReader(`{"model":"tts-1","input":"hi"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, int32(0), hits.Load())
}

func TestAudioSpeechProxy_ForwardsPathAndDefaultModel(t *testing.T) {
	var gotPath, gotBody string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		gotBody = string(b)
		w.Header().Set("Content-Type", "audio/mpeg")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte{0, 0, 0})
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		UpstreamBaseURL: up.URL,
		GatewayAPIKeys:  []string{"sk-gw"},
		UpstreamTimeout: 5 * time.Second,
	}
	e := gin.New()
	e.POST("/v1/audio/speech", GatewayAuth(testKeyStore(t, "sk-gw"), nil), AudioSpeechProxy(cfg, zap.NewNop(), testRuntimeStore(t, cfg), nil, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/audio/speech", strings.NewReader(`{"input":"hello","voice":"alloy"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-gw")
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/v1/audio/speech", gotPath)
	assert.Contains(t, gotBody, `"model":"tts-1"`)
	assert.Contains(t, gotBody, `"input":"hello"`)
	assert.Contains(t, gotBody, `"voice":"alloy"`)
}
