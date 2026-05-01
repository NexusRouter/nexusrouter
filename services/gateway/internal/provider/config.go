package provider

import (
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/keystore"
	"go.uber.org/zap"
)

// ProvideConfig 加载网关配置（环境变量）。
func ProvideConfig() *config.Config {
	return config.Load()
}

// ProvideKeyStore 基于配置构造密钥库；密钥文件配置非法时返回错误以阻止启动。
func ProvideKeyStore(cfg *config.Config, log *zap.Logger) (*keystore.Store, error) {
	return keystore.New(cfg, log)
}
