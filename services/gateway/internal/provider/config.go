package provider

import "github.com/NexusRouter/nexusrouter/services/gateway/internal/config"

// ProvideConfig 加载网关配置（环境变量）。
func ProvideConfig() *config.Config {
	return config.Load()
}
