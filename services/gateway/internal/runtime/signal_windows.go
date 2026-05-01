//go:build windows

package runtime

import "go.uber.org/zap"

// ListenSIGHUP Windows 无 SIGHUP。
func (s *Store) ListenSIGHUP(_ *zap.Logger) {}
