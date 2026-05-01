package provider

import "github.com/google/wire"

// Set 为 Wire 提供的 Provider 集合。
var Set = wire.NewSet(
	ProvideLogger,
	ProvideConfig,
	ProvideKeyStore,
	ProvideEngine,
)
