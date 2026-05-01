package provider

import (
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/google/wire"
)

// Set 为 Wire 提供的 Provider 集合。
var Set = wire.NewSet(
	ProvideLogger,
	ProvideConfig,
	ProvideKeyStore,
	ProvideRuntimeStore,
	ProvideEngine,
)

// ProvideRuntimeStore 加载 gateway.yaml（若配置）并与 env 合并为初始快照。
func ProvideRuntimeStore(cfg *config.Config) (*runtime.Store, error) {
	return runtime.NewStore(cfg)
}
