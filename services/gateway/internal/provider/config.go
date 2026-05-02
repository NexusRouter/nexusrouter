package provider

import (
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProvideConfig 加载网关配置（环境变量）。
func ProvideConfig() *config.Config {
	return config.Load()
}

// ProvideKeyStore 基于配置与数据库构造密钥库。
func ProvideKeyStore(cfg *config.Config, log *zap.Logger, db *gorm.DB) (*keystore.Store, error) {
	return keystore.New(cfg, log, db)
}
