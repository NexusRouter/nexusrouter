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

func TestHealth_JSONFieldsAndRFC3339(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", Health())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "ok", body["status"])
	require.NotEmpty(t, body["version"])
	serverTimeStr, ok := body["server_time"].(string)
	require.True(t, ok)
	_, err := time.Parse(time.RFC3339Nano, serverTimeStr)
	require.NoError(t, err)
	startTimeStr, ok := body["start_time"].(string)
	require.True(t, ok)
	_, err = time.Parse(time.RFC3339Nano, startTimeStr)
	require.NoError(t, err)
	require.Contains(t, body, "uptime_seconds")
	require.IsType(t, float64(0), body["uptime_seconds"])
}
