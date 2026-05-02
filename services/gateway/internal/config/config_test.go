package config

import "testing"

// TestLoad_ReturnsNonNil 默认环境变量下 Load 返回可用配置对象。
func TestLoad_ReturnsNonNil(t *testing.T) {
	cfg := Load()
	if cfg == nil {
		t.Fatal("Load() 返回 nil")
	}
	if cfg.HTTPListenAddr == "" {
		t.Fatal("HTTPListenAddr 为空")
	}
}
