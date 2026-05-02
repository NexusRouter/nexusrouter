package repository

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenDB 按配置打开 GORM：非空 DatabaseURL 使用 Postgres，否则使用 SQLite 文件。
func OpenDB(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("repository: 配置为空")
	}
	gcfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	if strings.TrimSpace(cfg.DatabaseURL) != "" {
		return gorm.Open(postgres.Open(strings.TrimSpace(cfg.DatabaseURL)), gcfg)
	}
	path := strings.TrimSpace(cfg.SQLitePath)
	if path == "" {
		path = defaultSQLitePathUnderGatewayModule()
	}
	return gorm.Open(sqlite.Open(path), gcfg)
}

// defaultSQLitePathUnderGatewayModule 返回 services/gateway/gateway.db 的绝对路径（相对本包源码定位模块根，不依赖进程 cwd）。
// 若解析失败则回退为 "gateway.db"（与历史行为一致）。
func defaultSQLitePathUnderGatewayModule() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "gateway.db"
	}
	// file = .../internal/repository/db.go → 模块根为 .../（即 services/gateway）
	gwRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	joined := filepath.Join(gwRoot, "gateway.db")
	abs, err := filepath.Abs(joined)
	if err != nil {
		return "gateway.db"
	}
	return abs
}

// AutoMigrate 创建或演进全部领域表。
func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("repository: db 为空")
	}
	return db.AutoMigrate(
		&GatewaySnapshotRow{},
		&APIKeyModel{},
		&AdminUserModel{},
		&SystemBootstrapRow{},
		&ModelCatalogEntry{},
		&ModelUpstreamBinding{},
	)
}
