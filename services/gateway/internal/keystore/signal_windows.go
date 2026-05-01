//go:build windows

package keystore

import "go.uber.org/zap"

// ListenSIGHUP Windows 无 SIGHUP，占位不注册。
func (s *Store) ListenSIGHUP(_ *zap.Logger) {}
