package runtime

import (
	"strings"
	"testing"
)

func TestSelectIPRateRule_PriorityAndPrefix(t *testing.T) {
	rules := []RateLimitRule{
		{ID: "low", Priority: 1, MatchPathPrefix: "/v1", Dimension: "ip", RPS: 1, Burst: 1, Enabled: true},
		{ID: "high", Priority: 10, MatchPathPrefix: "/v1/chat", Dimension: "ip", RPS: 9, Burst: 2, Enabled: true},
		{ID: "disabled", Priority: 99, MatchPathPrefix: "", Dimension: "ip", RPS: 99, Burst: 1, Enabled: false},
	}
	r := SelectIPRateRule(rules, "/v1/chat/completions")
	if r == nil || r.ID != "high" {
		t.Fatalf("expected high, got %+v", r)
	}
	r2 := SelectIPRateRule(rules, "/v1/models")
	if r2 == nil || r2.ID != "low" {
		t.Fatalf("expected low for /v1/models, got %+v", r2)
	}
}

func TestValidateSnapshot_AutoFillsEmptyRateLimitRuleID(t *testing.T) {
	s := &Snapshot{
		RateLimitRules: []RateLimitRule{
			{ID: "", Priority: 0, Dimension: "ip", RPS: 1, Burst: 1, Enabled: true},
		},
	}
	if err := ValidateSnapshot(s); err != nil {
		t.Fatal(err)
	}
	id := strings.TrimSpace(s.RateLimitRules[0].ID)
	if id == "" || !strings.HasPrefix(id, "rl-") {
		t.Fatalf("expected assigned id with prefix rl-, got %q", id)
	}
}

func TestSelectKeyRateRule_GlobalPrefix(t *testing.T) {
	rules := []RateLimitRule{
		{ID: "k", Priority: 1, MatchPathPrefix: "", Dimension: "api_key_fp", RPS: 3, Burst: 1, Enabled: true},
	}
	r := SelectKeyRateRule(rules, "/any")
	if r == nil || r.ID != "k" {
		t.Fatalf("expected rule k, got %+v", r)
	}
}
