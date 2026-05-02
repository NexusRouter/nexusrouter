package upstream

import (
	"net/url"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
)

// TestNewPicker_PickRoundRobin 在无 active 时按轮询选择合法上游。
func TestNewPicker_PickRoundRobin(t *testing.T) {
	p := NewPicker()
	s := &runtime.Snapshot{
		Upstreams: []runtime.Upstream{
			{ID: "a", BaseURL: "https://a.example.com", Weight: 1},
			{ID: "b", BaseURL: "https://b.example.com", Weight: 1},
		},
		Routing: runtime.Routing{Strategy: "round_robin"},
	}
	u, id, host, err := p.Pick(s)
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if u == nil || id == "" || host == "" {
		t.Fatalf("empty result: url=%v id=%q host=%q", u, id, host)
	}
	if u.Scheme != "https" {
		t.Fatalf("scheme: %s", u.Scheme)
	}
}

// TestNewPicker_PickNilSnapshot 无上游时返回错误。
func TestNewPicker_PickNilSnapshot(t *testing.T) {
	p := NewPicker()
	_, _, _, err := p.Pick(nil)
	if err == nil {
		t.Fatal("want error for nil snapshot")
	}
}

// TestNewPicker_PickActiveUpstream 显式 active_upstream_id 时命中对应条目。
func TestNewPicker_PickActiveUpstream(t *testing.T) {
	p := NewPicker()
	s := &runtime.Snapshot{
		Upstreams: []runtime.Upstream{
			{ID: "u1", BaseURL: "https://one.test/", Weight: 1},
			{ID: "u2", BaseURL: "https://two.test/", Weight: 1},
		},
		Routing: runtime.Routing{ActiveUpstreamID: "u2"},
	}
	got, id, _, err := p.Pick(s)
	if err != nil {
		t.Fatal(err)
	}
	if id != "u2" {
		t.Fatalf("id=%q want u2", id)
	}
	want, _ := url.Parse("https://two.test/")
	if got.String() != want.String() {
		t.Fatalf("url=%s want %s", got, want)
	}
}
