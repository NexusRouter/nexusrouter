package accesslog

import "testing"

// TestDefaultQueryLimits 默认查询上限为正。
func TestDefaultQueryLimits(t *testing.T) {
	l := DefaultQueryLimits()
	if l.MaxScanBytes <= 0 || l.MaxLines <= 0 {
		t.Fatalf("limits: %+v", l)
	}
}
