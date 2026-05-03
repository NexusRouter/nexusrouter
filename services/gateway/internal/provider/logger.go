package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ProvideLogger 提供 Zap 日志：未配置 LogDir 时为开发用控制台输出；配置 LogDir 时同时向标准错误与目录下 JSON 日志文件写入（默认 gateway.log；LogDailyFile 时为 gateway-YYYYMMDD.log）。
func ProvideLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg == nil || strings.TrimSpace(cfg.LogDir) == "" {
		return zap.NewDevelopment()
	}
	dir := strings.TrimSpace(cfg.LogDir)
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("log dir: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("log dir: %w", err)
	}
	base := "gateway.log"
	if cfg.LogDailyFile {
		base = "gateway-" + time.Now().Format("20060102") + ".log"
	}
	logPath := filepath.Join(abs, base)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	devEncCfg := zap.NewDevelopmentEncoderConfig()
	devEncCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEnc := zapcore.NewConsoleEncoder(devEncCfg)
	fileEnc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stderr), zapcore.DebugLevel),
		zapcore.NewCore(fileEnc, zapcore.AddSync(f), zapcore.InfoLevel),
	)
	return zap.New(core, zap.AddCaller()), nil
}
