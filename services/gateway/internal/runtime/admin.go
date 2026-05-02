package runtime

import "fmt"

// NewSnapshotFromBase 基于当前快照克隆并替换 upstreams 与 routing 段，供管理 API 使用。
func NewSnapshotFromBase(base *Snapshot, upstreams []Upstream, r Routing) (*Snapshot, error) {
	if base == nil {
		base = &Snapshot{}
	}
	next := cloneSnapshot(base)
	next.Upstreams = append([]Upstream(nil), upstreams...)
	next.Routing = r
	if err := validateSnapshot(next); err != nil {
		return nil, err
	}
	return next, nil
}

// ApplySnapshot 校验并以内存快照替换当前视图（不写磁盘）。
func (s *Store) ApplySnapshot(next *Snapshot) error {
	if s == nil {
		return fmt.Errorf("runtime: Store 为空")
	}
	if next == nil {
		return fmt.Errorf("runtime: 快照为空")
	}
	if err := validateSnapshot(next); err != nil {
		return err
	}
	cp := cloneSnapshot(next)
	s.v.Store(cp)
	return nil
}
