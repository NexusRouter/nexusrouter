package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAPIStatus_GET_SuccessEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/status", APIStatus())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, true, body["success"])
	require.Equal(t, "", body["message"])
	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	require.NotEmpty(t, data["version"])
	startStr, ok := data["start_time"].(string)
	require.True(t, ok)
	_, err := time.Parse(time.RFC3339Nano, startStr)
	require.NoError(t, err)
}
