package adminauth

import (
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
)

// TestIsConsoleConfigured_falseWhenDisabled 未启用控制台时不视为已配置。
func TestIsConsoleConfigured_falseWhenDisabled(t *testing.T) {
	cfg := &config.Config{EnableAdminConsole: false}
	if IsConsoleConfigured(cfg, nil) {
		t.Fatal("未启用控制台时应为 false")
	}
}
