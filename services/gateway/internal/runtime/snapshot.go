// Package runtime 承载可从 gateway.yaml 热更新的运行时快照（与 env 静态配置分离）。
package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync/atomic"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
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

// ProxyAccessLog 代理访问日志段。
type ProxyAccessLog struct {
	Enabled    bool   `yaml:"enabled"`
	Path       string `yaml:"path"`
	Level      string `yaml:"level"` // info | error
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
}

// Routing 上游路由策略。
type Routing struct {
	Strategy          string `yaml:"strategy"` // weighted_random | round_robin
	DefaultUpstreamID string `yaml:"default_upstream_id"`
	ActiveUpstreamID  string `yaml:"active_upstream_id"`
}

type fileYAML struct {
	Upstreams      []Upstream     `yaml:"upstreams"`
	Routing        Routing        `yaml:"routing"`
	CORS           CORS           `yaml:"cors"`
	RateLimit      RateLimit      `yaml:"rate_limit"`
	ProxyAccessLog ProxyAccessLog `yaml:"proxy_access_log"`
}

// Snapshot 为原子替换的运行时视图。
type Snapshot struct {
	Upstreams      []Upstream
	Routing        Routing
	CORS           CORS
	RateLimit      RateLimit
	ProxyAccessLog ProxyAccessLog
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
		ProxyAccessLog: ProxyAccessLog{Enabled: false, Level: "info"},
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
	dst.ProxyAccessLog = f.ProxyAccessLog
	if dst.ProxyAccessLog.Level == "" {
		dst.ProxyAccessLog.Level = "info"
	}
}

func validateSnapshot(s *Snapshot) error {
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

// FingerprintBearer 用于限流维度的匿名 key 指纹（非日志明文）。
func FingerprintBearer(token string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(h[:8])
}
