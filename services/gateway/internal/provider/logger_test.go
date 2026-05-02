package provider

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/stretchr/testify/require"
)

func TestProvideLogger_fileSink(t *testing.T) {
	dir := t.TempDir()
	log, err := ProvideLogger(&config.Config{LogDir: dir})
	require.NoError(t, err)
	require.NotNil(t, log)
	log.Info("probe")
	_ = log.Sync()
	b, err := os.ReadFile(filepath.Join(dir, "gateway.log"))
	require.NoError(t, err)
	require.Contains(t, string(b), "probe")
}

func TestProvideLogger_dailyFileSink(t *testing.T) {
	dir := t.TempDir()
	wantName := "gateway-" + time.Now().Format("20060102") + ".log"
	log, err := ProvideLogger(&config.Config{LogDir: dir, LogDailyFile: true})
	require.NoError(t, err)
	require.NotNil(t, log)
	log.Info("daily")
	_ = log.Sync()
	b, err := os.ReadFile(filepath.Join(dir, wantName))
	require.NoError(t, err)
	require.Contains(t, string(b), "daily")
}

func TestProvideLogger_nilConfigDevelopment(t *testing.T) {
	log, err := ProvideLogger(nil)
	require.NoError(t, err)
	require.NotNil(t, log)
}
