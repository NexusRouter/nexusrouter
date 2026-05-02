package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EffectiveUploadsDir 返回上传文件根目录绝对路径。
func (c *Config) EffectiveUploadsDir() string {
	if c != nil {
		if s := strings.TrimSpace(c.UploadsDir); s != "" {
			if filepath.IsAbs(s) {
				return filepath.Clean(s)
			}
			if wd, err := os.Getwd(); err == nil {
				return filepath.Clean(filepath.Join(wd, s))
			}
			return filepath.Clean(s)
		}
	}
	return defaultUploadsUnderGatewayModule()
}

func defaultUploadsUnderGatewayModule() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Clean("data/uploads")
	}
	// .../internal/config/uploads.go → 模块根 services/gateway
	gwRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(gwRoot, "data", "uploads")
}
