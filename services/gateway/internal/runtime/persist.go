package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PersistSnapshot 将给定快照校验后原子写入网关配置文件并 Reload；失败时保留旧快照。
func (s *Store) PersistSnapshot(next *Snapshot) error {
	if s == nil {
		return fmt.Errorf("runtime: Store 为空")
	}
	if next == nil {
		return fmt.Errorf("runtime: 快照为空")
	}
	if err := validateSnapshot(next); err != nil {
		return err
	}
	p := strings.TrimSpace(s.path)
	if p == "" {
		return fmt.Errorf("runtime: 未配置 NEXUSROUTER_GATEWAY_CONFIG_FILE，无法持久化")
	}
	b, err := marshalFileYAML(next)
	if err != nil {
		return err
	}
	if err := writeFileAtomic(p, b); err != nil {
		return err
	}
	return s.Reload()
}

// PersistCurrent 将当前内存快照写回磁盘并 Reload。
func (s *Store) PersistCurrent() error {
	cur := s.Snapshot()
	return s.PersistSnapshot(cloneSnapshot(cur))
}

func marshalFileYAML(s *Snapshot) ([]byte, error) {
	f := fileYAML{
		Upstreams:      append([]Upstream(nil), s.Upstreams...),
		Routing:        s.Routing,
		CORS:           s.CORS,
		RateLimit:      s.RateLimit,
		RateLimitRules: append([]RateLimitRule(nil), s.RateLimitRules...),
		IPAccess:       IPAccess{Mode: s.IPAccess.Mode, CIDRs: append([]string(nil), s.IPAccess.CIDRs...)},
		ProxyAccessLog: s.ProxyAccessLog,
		AdminAlerts:    s.AdminAlerts,
	}
	return yaml.Marshal(&f)
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".nexus-gw-*.yaml")
	if err != nil {
		return fmt.Errorf("runtime: 创建临时文件: %w", err)
	}
	tmp := f.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmp)
		}
	}()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return fmt.Errorf("runtime: 写入临时文件: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("runtime: 同步临时文件: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("runtime: 关闭临时文件: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("runtime: 替换配置文件: %w", err)
	}
	cleanup = false
	return nil
}
