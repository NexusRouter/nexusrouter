package alerts

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/metrics"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"github.com/stretchr/testify/require"
)

func TestEvalOnce_GoesCriticalAfterConsecutive(t *testing.T) {
	atomic.StoreUint32(&consecutive, 0)
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "gw.yaml")
	err := os.WriteFile(yamlPath, []byte(`upstreams:
  - id: a
    base_url: https://example.com
    weight: 1
routing:
  strategy: round_robin
  default_upstream_id: a
admin_alerts:
  enabled: true
  error_rate_threshold: 0.05
  consecutive_periods: 2
`), 0o600)
	require.NoError(t, err)
	cfg := &config.Config{
		GatewayAPIKeys:    []string{"k"},
		GatewayConfigFile: yamlPath,
	}
	rt, err := runtime.NewStore(cfg, nil)
	require.NoError(t, err)
	col := metrics.NewCollector()
	for i := 0; i < 50; i++ {
		col.RecordGatewayError("E")
	}
	evalOnce(rt, col, nil)
	require.Equal(t, "warning", Current().Level)
	evalOnce(rt, col, nil)
	require.Equal(t, "critical", Current().Level)
}
