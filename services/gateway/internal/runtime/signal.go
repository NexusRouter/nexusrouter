//go:build !windows

package runtime

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

// ListenSIGHUP 在后台监听 SIGHUP 并重载 gateway.yaml；path 为空时不注册。
func (s *Store) ListenSIGHUP(log *zap.Logger) {
	if log == nil {
		log = zap.NewNop()
	}
	if s == nil || s.Path() == "" {
		return
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP)
	go func() {
		for range ch {
			if err := s.Reload(); err != nil {
				log.Error("SIGHUP 重载网关配置失败", zap.Error(err), zap.String("path", s.Path()))
				continue
			}
			log.Info("已通过 SIGHUP 重载网关配置文件", zap.String("path", s.Path()))
		}
	}()
}
