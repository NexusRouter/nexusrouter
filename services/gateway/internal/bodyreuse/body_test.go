package bodyreuse

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetRequestBody_CachesAndSecondReadMatches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/t", func(c *gin.Context) {
		b1, err := GetRequestBody(c)
		require.NoError(t, err)
		b2, err := GetRequestBody(c)
		require.NoError(t, err)
		require.Equal(t, b1, b2)
		require.Equal(t, `{"a":1}`, string(b1))
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/t", bytes.NewBufferString(`{"a":1}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestUnmarshalBodyReusable_RestoresBodyForReadAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	type payload struct {
		Model string `json:"model"`
	}
	r := gin.New()
	r.POST("/t", func(c *gin.Context) {
		var p payload
		require.NoError(t, UnmarshalBodyReusable(c, &p))
		require.Equal(t, "x", p.Model)
		again, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"model":"x"}`, string(again))
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/t", bytes.NewBufferString(`{"model":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestResetRequestBody_UpdatesCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/t", func(c *gin.Context) {
		_, err := GetRequestBody(c)
		require.NoError(t, err)
		ResetRequestBody(c, []byte(`{"k":2}`))
		b, err := GetRequestBody(c)
		require.NoError(t, err)
		require.JSONEq(t, `{"k":2}`, string(b))
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/t", bytes.NewBufferString(`{"k":1}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
