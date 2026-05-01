//go:build !windows

package keystore

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

// ListenSIGHUP 在后台监听 SIGHUP 并热加载密钥文件；path 为空时不注册信号。
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
				log.Error("SIGHUP 重载密钥失败", zap.Error(err))
				continue
			}
			log.Info("已通过 SIGHUP 重载 API 密钥文件", zap.String("path", s.Path()))
		}
	}()
}
