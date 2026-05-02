package provider

import (
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProvideDB 打开数据库、迁移并在空库时从遗留文件/环境变量导入。
func ProvideDB(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	db, err := repository.OpenDB(cfg, log)
	if err != nil {
		return nil, err
	}
	if err := repository.AutoMigrate(db); err != nil {
		return nil, err
	}
	if err := repository.EnsureSystemBootstrap(db); err != nil {
		return nil, err
	}
	if err := repository.BootstrapFromConfig(cfg, db, log); err != nil {
		return nil, err
	}
	if err := repository.SeedOfficialVendors(db, log); err != nil {
		return nil, err
	}
	if err := keystore.BootstrapKeysIfEmpty(cfg, db, log); err != nil {
		return nil, err
	}
	return db, nil
}
