// Package accesslog 提供与主 Zap 分离的代理访问日志（JSON 行）。
package accesslog

import (
	"os"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 访问日志；未启用时 log 为 nil。
type Logger struct {
	log   *zap.Logger
	level string
}

// New 基于快照构造；未启用时返回空 Logger。
func New(s *runtime.Snapshot) *Logger {
	if s == nil || !s.ProxyAccessLog.Enabled {
		return &Logger{}
	}
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	enc := zapcore.NewJSONEncoder(encCfg)
	ws := zapcore.AddSync(os.Stdout)
	if p := strings.TrimSpace(s.ProxyAccessLog.Path); p != "" {
		lj := &lumberjack.Logger{
			Filename:   p,
			MaxSize:    maxInt(1, s.ProxyAccessLog.MaxSizeMB),
			MaxBackups: maxInt(1, s.ProxyAccessLog.MaxBackups),
		}
		ws = zapcore.AddSync(lj)
	}
	core := zapcore.NewCore(enc, ws, zap.NewAtomicLevelAt(zapcore.InfoLevel))
	zl := zap.New(core)
	lv := strings.ToLower(strings.TrimSpace(s.ProxyAccessLog.Level))
	if lv == "" {
		lv = "info"
	}
	return &Logger{log: zl, level: lv}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ShouldLog 按级别与状态判断是否记录。
func ShouldLog(level string, status int, gatewayErr bool) bool {
	lv := strings.ToLower(strings.TrimSpace(level))
	if lv == "error" {
		return status >= 500 || gatewayErr
	}
	return true
}

// Write 写入一条代理访问日志。
func (l *Logger) Write(status int, gatewayErr bool, fields ...zap.Field) {
	if l.log == nil {
		return
	}
	if !ShouldLog(l.level, status, gatewayErr) {
		return
	}
	l.log.Info("proxy_access", append([]zap.Field{zap.Time("ts", time.Now().UTC())}, fields...)...)
}
