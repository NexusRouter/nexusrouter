package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRegister_AudioSpeech_MethodNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	cfg := &config.Config{GatewayAPIKeys: []string{"k"}}
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	Register(e, Deps{
		Config:  cfg,
		Log:     zap.NewNop(),
		Runtime: rt,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/audio/speech", nil)
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "METHOD_NOT_ALLOWED", body["code"])
}
