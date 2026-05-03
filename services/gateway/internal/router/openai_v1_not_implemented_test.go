package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRegister_OpenAIV1NotImplemented_501(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.Config{GatewayAPIKeys: []string{"sk-gw"}}
	ks, err := keystore.New(cfg, zap.NewNop(), nil)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	Register(r, Deps{
		Config:   cfg,
		Log:      zap.NewNop(),
		KeyStore: ks,
		Runtime:  rt,
	})

	for _, path := range []string{"/v1/completions", "/v1/images/edits"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Authorization", "Bearer sk-gw")
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotImplemented, rec.Code, path)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "NOT_IMPLEMENTED", body["code"], path)
	}
}

func TestRegister_OpenAIV1NotImplemented_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.Config{GatewayAPIKeys: []string{"sk-gw"}}
	ks, err := keystore.New(cfg, zap.NewNop(), nil)
	require.NoError(t, err)
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	Register(r, Deps{
		Config:   cfg,
		Log:      zap.NewNop(),
		KeyStore: ks,
		Runtime:  rt,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/images/edits", nil)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
