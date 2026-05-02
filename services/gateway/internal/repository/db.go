package repository

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const sqliteBusyTimeoutMaxMS = 600_000

// gatewayGormConfig 返回网关进程内持久化库使用的 GORM 配置（预编译语句与静默日志）。
func gatewayGormConfig() *gorm.Config {
	return &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Silent),
		PrepareStmt: true,
	}
}

// effectiveSQLiteBusyMS 将配置中的毫秒数归一为合法 _busy_timeout：未设置或非正数时为 3000，过大时封顶。
func effectiveSQLiteBusyMS(ms int) int {
	if ms <= 0 {
		return 3000
	}
	if ms > sqliteBusyTimeoutMaxMS {
		return sqliteBusyTimeoutMaxMS
	}
	return ms
}

// sqliteOpenDSN 为 SQLite 驱动连接串附加 _busy_timeout（与路径中已有 query 时用 & 拼接）。
func sqliteOpenDSN(path string, busyMS int) string {
	busy := effectiveSQLiteBusyMS(busyMS)
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + "_busy_timeout=" + strconv.Itoa(busy)
}

// OpenDB 按配置打开 GORM：非空 DatabaseURL 使用 Postgres，否则使用 SQLite 文件。
func OpenDB(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("repository: 配置为空")
	}
	gcfg := gatewayGormConfig()
	if strings.TrimSpace(cfg.DatabaseURL) != "" {
		return gorm.Open(postgres.Open(strings.TrimSpace(cfg.DatabaseURL)), gcfg)
	}
	path := strings.TrimSpace(cfg.SQLitePath)
	if path == "" {
		path = defaultSQLitePathUnderGatewayModule()
	}
	dsn := sqliteOpenDSN(path, cfg.SQLiteBusyTimeoutMS)
	return gorm.Open(sqlite.Open(dsn), gcfg)
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
		&ModelVendor{},
		&ModelBase{},
		&ModelUpstream{},
		&ModelInstance{},
	)
}
