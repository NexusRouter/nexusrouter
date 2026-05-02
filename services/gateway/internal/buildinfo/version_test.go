package buildinfo

import "testing"

// TestVersionNonEmpty 版本字符串用于健康检查等，须非空。
func TestVersionNonEmpty(t *testing.T) {
	if Version == "" {
		t.Fatal("Version 为空")
	}
}
