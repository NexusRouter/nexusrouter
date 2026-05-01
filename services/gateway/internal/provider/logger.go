package provider

import "go.uber.org/zap"

// ProvideLogger 提供开发友好的 Zap 日志（后续可切换为生产配置）。
func ProvideLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}
