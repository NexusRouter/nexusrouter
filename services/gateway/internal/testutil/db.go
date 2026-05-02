package testutil

import (
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MustSQLiteMemory 返回内存 SQLite 并完成迁移与 Bootstrap（与生产启动顺序一致）。
func MustSQLiteMemory(t *testing.T, cfg *config.Config) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, repository.AutoMigrate(db))
	require.NoError(t, repository.BootstrapFromConfig(cfg, db, nil))
	require.NoError(t, keystore.BootstrapKeysIfEmpty(cfg, db, nil))
	return db
}
