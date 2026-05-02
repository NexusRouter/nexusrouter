package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBootstrapFromConfig_ImportsGatewayYAMLWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "gw.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte(`upstreams:
  - id: u1
    base_url: https://example.com
    weight: 1
routing:
  strategy: round_robin
  default_upstream_id: u1
`), 0o600))

	cfg := &config.Config{GatewayConfigFile: yamlPath}
	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "t.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, AutoMigrate(db))
	require.NoError(t, BootstrapFromConfig(cfg, db, zap.NewNop()))

	var row GatewaySnapshotRow
	require.NoError(t, db.First(&row, 1).Error)
	require.Contains(t, row.YAMLBody, "u1")
}
