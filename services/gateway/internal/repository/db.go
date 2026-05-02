package repository

import (
	"fmt"
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
		path = "gateway.db"
	}
	return gorm.Open(sqlite.Open(path), gcfg)
}

// AutoMigrate 创建或演进全部领域表。
func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("repository: db 为空")
	}
	return db.AutoMigrate(&GatewaySnapshotRow{}, &APIKeyModel{}, &AdminUserModel{})
}
