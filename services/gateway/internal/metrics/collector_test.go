package metrics

import "testing"

// TestNewCollector_SummaryJSON 采集器可构造并输出管理端摘要。
func TestNewCollector_SummaryJSON(t *testing.T) {
	c := NewCollector()
	c.RecordGatewayError("TEST_CODE")
	m := c.SummaryJSON()
	if m["requests_total"] != uint64(1) {
		t.Fatalf("requests_total=%v want 1", m["requests_total"])
	}
}

// TestCollectorNilMethods 空指针调用须安全 no-op。
func TestCollectorNilMethods(t *testing.T) {
	var c *Collector
	c.RecordChat(200, 1, "")
	c.RecordGatewayError("X")
	if len(c.SummaryJSON()) != 0 {
		t.Fatal("nil SummaryJSON should be empty map")
	}
}
