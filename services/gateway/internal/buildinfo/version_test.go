package buildinfo

import (
	"testing"
	"time"
)

// TestVersionNonEmpty 版本字符串用于健康检查等，须非空。
func TestVersionNonEmpty(t *testing.T) {
	if Version == "" {
		t.Fatal("Version 为空")
	}
}

func TestProcessStartSet(t *testing.T) {
	if ProcessStart.IsZero() {
		t.Fatal("ProcessStart 未初始化")
	}
	if time.Since(ProcessStart) < 0 {
		t.Fatal("ProcessStart 在未来")
	}
}
