// Package runtime 承载可从 gateway.yaml 热更新的运行时快照（与 env 静态配置分离）。
package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/ipaccess"
	"gopkg.in/yaml.v3"
)

// Upstream 单条上游定义。
type Upstream struct {
	ID      string `yaml:"id"`
	BaseURL string `yaml:"base_url"`
	Weight  int    `yaml:"weight"`
}

// CORS 运行时 CORS 段。
type CORS struct {
	Enabled       bool     `yaml:"enabled"`
	AllowOrigins  []string `yaml:"allow_origins"`
	AllowMethods  []string `yaml:"allow_methods"`
	AllowHeaders  []string `yaml:"allow_headers"`
	MaxAgeSeconds int      `yaml:"max_age_seconds"`
}

// RateLimit 限流段。
type RateLimit struct {
	RPSPerIP  float64 `yaml:"rps_per_ip"`
	RPSPerKey float64 `yaml:"rps_per_key"`
}

// RateLimitRule 单条限流规则（priority 越大越优先；match_path_prefix 空表示匹配所有路径）。
type RateLimitRule struct {
	ID               string  `yaml:"id" json:"id"`
	Priority         int     `yaml:"priority" json:"priority"`
	MatchPathPrefix  string  `yaml:"match_path_prefix" json:"match_path_prefix"`
	Dimension        string  `yaml:"dimension" json:"dimension"` // ip | api_key_fp
	RPS              float64 `yaml:"rps" json:"rps"`
	Burst            int     `yaml:"burst" json:"burst"`
	Enabled          bool    `yaml:"enabled" json:"enabled"`
}

// IPAccess IP 白/黑名单段。
type IPAccess struct {
	Mode  string   `yaml:"mode" json:"mode"` // off | allowlist | denylist
	CIDRs []string `yaml:"cidrs" json:"cidrs"`
}

// ProxyAccessLog 代理访问日志段。
type ProxyAccessLog struct {
	Enabled    bool   `yaml:"enabled"`
	Path       string `yaml:"path"`
	Level      string `yaml:"level"` // info | error
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
}

// AdminAlerts 管理面板运行态告警阈值（错误率基于进程内累计成功率近似）。
type AdminAlerts struct {
	Enabled              bool    `yaml:"enabled" json:"enabled"`
	ErrorRateThreshold   float64 `yaml:"error_rate_threshold" json:"error_rate_threshold"` // 0~1，如 0.2 表示错误率 ≥20% 触发
	WindowSeconds        int     `yaml:"window_seconds" json:"window_seconds"`               // 预留；当前评估用固定 tick
	ConsecutivePeriods   int     `yaml:"consecutive_periods" json:"consecutive_periods"`     // 连续超阈值评估周期数后升为 critical
}

// Routing 上游路由策略。
type Routing struct {
	Strategy          string `yaml:"strategy"` // weighted_random | round_robin
	DefaultUpstreamID string `yaml:"default_upstream_id"`
	ActiveUpstreamID  string `yaml:"active_upstream_id"`
}

type fileYAML struct {
	Upstreams        []Upstream        `yaml:"upstreams"`
	Routing          Routing           `yaml:"routing"`
	CORS             CORS              `yaml:"cors"`
	RateLimit        RateLimit         `yaml:"rate_limit"`
	RateLimitRules   []RateLimitRule   `yaml:"rate_limit_rules"`
	IPAccess         IPAccess          `yaml:"ip_access"`
	ProxyAccessLog   ProxyAccessLog    `yaml:"proxy_access_log"`
	AdminAlerts      AdminAlerts       `yaml:"admin_alerts"`
}

// Snapshot 为原子替换的运行时视图。
type Snapshot struct {
	Upstreams        []Upstream
	Routing          Routing
	CORS             CORS
	RateLimit        RateLimit
	RateLimitRules   []RateLimitRule
	IPAccess         IPAccess
	ProxyAccessLog   ProxyAccessLog
	AdminAlerts      AdminAlerts
}

// Store 保存当前快照并可从文件重载。
type Store struct {
	path string
	cfg  *config.Config
	v    atomic.Value // *Snapshot
}

// NewStore 基于静态 env 配置构造初始快照；若 gateway.yaml 存在则合并。
func NewStore(cfg *config.Config) (*Store, error) {
	if cfg == nil {
		cfg = config.Load()
	}
	s := &Store{path: strings.TrimSpace(cfg.GatewayConfigFile), cfg: cfg}
	snap, err := buildSnapshot(cfg, s.path)
	if err != nil {
		return nil, err
	}
	s.v.Store(snap)
	return s, nil
}

// Snapshot 返回当前指针（只读使用）。
func (s *Store) Snapshot() *Snapshot {
	v := s.v.Load()
	if v == nil {
		return &Snapshot{}
	}
	return v.(*Snapshot)
}

// Path 返回配置文件路径（可能为空）。
func (s *Store) Path() string { return s.path }

// Reload 从磁盘重新加载；失败时保留旧快照并返回错误。
func (s *Store) Reload() error {
	if strings.TrimSpace(s.path) == "" {
		return nil
	}
	snap, err := buildSnapshot(s.cfg, s.path)
	if err != nil {
		return err
	}
	s.v.Store(snap)
	return nil
}

// SetActiveUpstream 仅更新内存中的 active_upstream_id（用于管理 API）。
func (s *Store) SetActiveUpstream(id string) error {
	cur := s.Snapshot()
	next := cloneSnapshot(cur)
	next.Routing.ActiveUpstreamID = strings.TrimSpace(id)
	if err := validateSnapshot(next); err != nil {
		return err
	}
	s.v.Store(next)
	return nil
}

func cloneSnapshot(cur *Snapshot) *Snapshot {
	if cur == nil {
		return &Snapshot{}
	}
	out := *cur
	out.Upstreams = append([]Upstream(nil), cur.Upstreams...)
	out.RateLimitRules = append([]RateLimitRule(nil), cur.RateLimitRules...)
	out.IPAccess.CIDRs = append([]string(nil), cur.IPAccess.CIDRs...)
	out.CORS.AllowOrigins = append([]string(nil), cur.CORS.AllowOrigins...)
	out.CORS.AllowMethods = append([]string(nil), cur.CORS.AllowMethods...)
	out.CORS.AllowHeaders = append([]string(nil), cur.CORS.AllowHeaders...)
	out.AdminAlerts = cur.AdminAlerts
	return &out
}

func buildSnapshot(cfg *config.Config, path string) (*Snapshot, error) {
	base := snapshotFromEnv(cfg)
	if path == "" {
		return base, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("runtime: 读取网关配置: %w", err)
	}
	var f fileYAML
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("runtime: 解析 YAML: %w", err)
	}
	mergeFile(base, &f)
	if err := validateSnapshot(base); err != nil {
		return nil, err
	}
	return base, nil
}

func snapshotFromEnv(cfg *config.Config) *Snapshot {
	bases := cfg.EffectiveUpstreamBases()
	up := make([]Upstream, 0, len(bases))
	for i, b := range bases {
		id := fmt.Sprintf("e%d", i)
		w := 1
		if len(bases) == 1 {
			id = "default"
		}
		up = append(up, Upstream{ID: id, BaseURL: b, Weight: w})
	}
	def := ""
	if len(up) > 0 {
		def = up[0].ID
	}
	return &Snapshot{
		Upstreams: up,
		Routing: Routing{
			Strategy:          "round_robin",
			DefaultUpstreamID: def,
			ActiveUpstreamID:  "",
		},
		CORS:           CORS{Enabled: false},
		RateLimit:      RateLimit{},
		IPAccess:       IPAccess{Mode: "off"},
		ProxyAccessLog: ProxyAccessLog{Enabled: false, Level: "info"},
		AdminAlerts:    AdminAlerts{Enabled: false},
	}
}

func mergeFile(dst *Snapshot, f *fileYAML) {
	if len(f.Upstreams) > 0 {
		dst.Upstreams = f.Upstreams
	}
	if f.Routing.Strategy != "" {
		dst.Routing.Strategy = f.Routing.Strategy
	}
	if f.Routing.DefaultUpstreamID != "" {
		dst.Routing.DefaultUpstreamID = f.Routing.DefaultUpstreamID
	}
	dst.Routing.ActiveUpstreamID = f.Routing.ActiveUpstreamID
	dst.CORS = f.CORS
	dst.RateLimit = f.RateLimit
	if f.RateLimitRules != nil {
		dst.RateLimitRules = append([]RateLimitRule(nil), f.RateLimitRules...)
	}
	if strings.TrimSpace(f.IPAccess.Mode) != "" || len(f.IPAccess.CIDRs) > 0 {
		dst.IPAccess = f.IPAccess
	}
	dst.ProxyAccessLog = f.ProxyAccessLog
	if dst.ProxyAccessLog.Level == "" {
		dst.ProxyAccessLog.Level = "info"
	}
	dst.AdminAlerts = f.AdminAlerts
}

func validateSnapshot(s *Snapshot) error {
	if err := validateAdminAlerts(s); err != nil {
		return err
	}
	if err := validateIPAccess(s); err != nil {
		return err
	}
	if err := validateRateLimitRules(s); err != nil {
		return err
	}
	ids := map[string]bool{}
	for _, u := range s.Upstreams {
		if strings.TrimSpace(u.ID) == "" || strings.TrimSpace(u.BaseURL) == "" {
			return fmt.Errorf("runtime: upstream 缺少 id 或 base_url")
		}
		if _, err := url.Parse(u.BaseURL); err != nil {
			return fmt.Errorf("runtime: 非法 base_url %q: %w", u.BaseURL, err)
		}
		ids[u.ID] = true
	}
	if len(s.Upstreams) == 0 {
		return nil
	}
	if s.Routing.DefaultUpstreamID != "" && !ids[s.Routing.DefaultUpstreamID] {
		return fmt.Errorf("runtime: default_upstream_id 不存在于 upstreams")
	}
	if s.Routing.ActiveUpstreamID != "" && !ids[s.Routing.ActiveUpstreamID] {
		return fmt.Errorf("runtime: active_upstream_id 不存在于 upstreams")
	}
	return nil
}

func validateAdminAlerts(s *Snapshot) error {
	a := s.AdminAlerts
	if !a.Enabled {
		return nil
	}
	if a.ErrorRateThreshold <= 0 || a.ErrorRateThreshold > 1 {
		return fmt.Errorf("runtime: admin_alerts.error_rate_threshold 须在 (0,1]")
	}
	if a.WindowSeconds != 0 && a.WindowSeconds < 10 {
		return fmt.Errorf("runtime: admin_alerts.window_seconds 须 >= 10 或 0")
	}
	if a.ConsecutivePeriods < 0 {
		return fmt.Errorf("runtime: admin_alerts.consecutive_periods 不能为负数")
	}
	return nil
}

func validateIPAccess(s *Snapshot) error {
	mode := strings.ToLower(strings.TrimSpace(s.IPAccess.Mode))
	if mode == "" {
		mode = "off"
	}
	switch mode {
	case "off", "allowlist", "denylist":
	default:
		return fmt.Errorf("runtime: ip_access.mode 非法，须为 off|allowlist|denylist")
	}
	if mode == "off" {
		return nil
	}
	if _, err := ipaccess.Compile(s.IPAccess.CIDRs); err != nil {
		return fmt.Errorf("runtime: ip_access.cidrs: %w", err)
	}
	return nil
}

func validateRateLimitRules(s *Snapshot) error {
	seen := map[string]bool{}
	for _, r := range s.RateLimitRules {
		id := strings.TrimSpace(r.ID)
		if id == "" {
			return fmt.Errorf("runtime: rate_limit_rules 缺少 id")
		}
		if seen[id] {
			return fmt.Errorf("runtime: rate_limit_rules 重复 id %q", id)
		}
		seen[id] = true
		d := strings.ToLower(strings.TrimSpace(r.Dimension))
		if d != "ip" && d != "api_key_fp" {
			return fmt.Errorf("runtime: 规则 %q dimension 须为 ip 或 api_key_fp", id)
		}
		if r.RPS <= 0 {
			return fmt.Errorf("runtime: 规则 %q rps 须为正数", id)
		}
		if r.Burst < 0 {
			return fmt.Errorf("runtime: 规则 %q burst 不能为负数", id)
		}
		burst := r.Burst
		if burst == 0 {
			burst = int(r.RPS + 0.999)
		}
		if burst < 1 {
			return fmt.Errorf("runtime: 规则 %q burst 须 >= 1", id)
		}
	}
	return nil
}

// SelectIPRateRule 返回命中的首条已启用 IP 维度规则（按 priority 降序）。
func SelectIPRateRule(rules []RateLimitRule, path string) *RateLimitRule {
	type item struct {
		r RateLimitRule
	}
	var cand []item
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		if strings.ToLower(strings.TrimSpace(r.Dimension)) != "ip" {
			continue
		}
		pfx := strings.TrimSpace(r.MatchPathPrefix)
		if pfx != "" && !strings.HasPrefix(path, pfx) {
			continue
		}
		cand = append(cand, item{r: r})
	}
	if len(cand) == 0 {
		return nil
	}
	sort.SliceStable(cand, func(i, j int) bool {
		if cand[i].r.Priority != cand[j].r.Priority {
			return cand[i].r.Priority > cand[j].r.Priority
		}
		return cand[i].r.ID < cand[j].r.ID
	})
	cp := cand[0].r
	return &cp
}

// SelectKeyRateRule 返回命中的首条已启用 api_key_fp 维度规则（按 priority 降序）。
func SelectKeyRateRule(rules []RateLimitRule, path string) *RateLimitRule {
	type item struct {
		r RateLimitRule
	}
	var cand []item
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		if strings.ToLower(strings.TrimSpace(r.Dimension)) != "api_key_fp" {
			continue
		}
		pfx := strings.TrimSpace(r.MatchPathPrefix)
		if pfx != "" && !strings.HasPrefix(path, pfx) {
			continue
		}
		cand = append(cand, item{r: r})
	}
	if len(cand) == 0 {
		return nil
	}
	sort.SliceStable(cand, func(i, j int) bool {
		if cand[i].r.Priority != cand[j].r.Priority {
			return cand[i].r.Priority > cand[j].r.Priority
		}
		return cand[i].r.ID < cand[j].r.ID
	})
	cp := cand[0].r
	return &cp
}

// FingerprintBearer 用于限流维度的匿名 key 指纹（非日志明文）。
func FingerprintBearer(token string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(h[:8])
}

// CloneSnapshot 返回深拷贝（供管理 API 修改前克隆）。
func CloneSnapshot(cur *Snapshot) *Snapshot {
	return cloneSnapshot(cur)
}

// ValidateSnapshot 校验快照（供管理 API）。
func ValidateSnapshot(s *Snapshot) error {
	return validateSnapshot(s)
}
