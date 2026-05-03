package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBootstrapGate_BlocksWhenUninitialized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "g.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, repository.AutoMigrate(db))
	require.NoError(t, repository.EnsureSystemBootstrap(db))

	r := gin.New()
	r.Use(RequestID())
	r.Use(BootstrapGate(db, zap.NewNop()))
	r.GET("/v1/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/x", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "BOOTSTRAP_REQUIRED", body["code"])
}

func TestBootstrapGate_AllowsWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "g2.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, repository.AutoMigrate(db))
	require.NoError(t, repository.EnsureSystemBootstrap(db))

	r := gin.New()
	r.Use(RequestID())
	r.Use(BootstrapGate(db, zap.NewNop()))
	r.GET("/api/bootstrap/v1/status", func(c *gin.Context) { c.Status(http.StatusTeapot) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/bootstrap/v1/status", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusTeapot, rec.Code)
}

func TestBootstrapGate_AllowsAPIStatusWhenUninitialized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "g3.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, repository.AutoMigrate(db))
	require.NoError(t, repository.EnsureSystemBootstrap(db))

	r := gin.New()
	r.Use(RequestID())
	r.Use(BootstrapGate(db, zap.NewNop()))
	r.GET("/api/status", func(c *gin.Context) { c.Status(http.StatusTeapot) })
	r.GET("/v1/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusTeapot, rec.Code)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/v1/x", nil)
	r.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusForbidden, rec2.Code)
}
