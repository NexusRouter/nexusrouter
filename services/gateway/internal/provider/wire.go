package provider

import (
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// Set 为 Wire 提供的 Provider 集合。
var Set = wire.NewSet(
	ProvideLogger,
	ProvideConfig,
	ProvideDB,
	ProvideKeyStore,
	ProvideRuntimeStore,
	ProvideMetrics,
	ProvideEngine,
)

// ProvideRuntimeStore 从数据库（或遗留 gateway.yaml）与 env 合并加载初始快照。
func ProvideRuntimeStore(cfg *config.Config, db *gorm.DB) (*runtime.Store, error) {
	return runtime.NewStore(cfg, db)
}

// ProvideMetrics 进程内指标采集器（单例）。
func ProvideMetrics() *metrics.Collector {
	return metrics.NewCollector()
}
