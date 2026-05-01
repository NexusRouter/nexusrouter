// Package upstream 根据运行时快照选择上游 URL。
package upstream

import (
	"fmt"
	"math/rand/v2"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
)

// Picker 封装加权随机与轮询。
type Picker struct {
	rr  atomic.Uint32
	rng *rand.Rand
}

// NewPicker 构造选择器（独立随机流）。
func NewPicker() *Picker {
	return &Picker{
		rng: rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().Unix()))),
	}
}

// Pick 返回用于 ReverseProxy 的基址 URL、上游 id 与 host。
func (p *Picker) Pick(s *runtime.Snapshot) (*url.URL, string, string, error) {
	if s == nil || len(s.Upstreams) == 0 {
		return nil, "", "", fmt.Errorf("upstream: 无上游")
	}
	byID := make(map[string]runtime.Upstream, len(s.Upstreams))
	for _, u := range s.Upstreams {
		byID[u.ID] = u
	}
	if aid := strings.TrimSpace(s.Routing.ActiveUpstreamID); aid != "" {
		u, ok := byID[aid]
		if !ok {
			return nil, "", "", fmt.Errorf("upstream: active id 不存在")
		}
		pu, err := url.Parse(u.BaseURL)
		if err != nil || pu.Scheme == "" || pu.Host == "" {
			return nil, "", "", fmt.Errorf("upstream: 非法 base_url")
		}
		return pu, u.ID, pu.Host, nil
	}
	strat := strings.ToLower(strings.TrimSpace(s.Routing.Strategy))
	if strat == "round_robin" || strat == "" {
		n := uint32(len(s.Upstreams))
		i := p.rr.Add(1) % n
		u := s.Upstreams[i]
		pu, err := url.Parse(u.BaseURL)
		if err != nil || pu.Scheme == "" || pu.Host == "" {
			return nil, "", "", err
		}
		return pu, u.ID, pu.Host, nil
	}
	// weighted_random
	type item struct {
		u  runtime.Upstream
		pu *url.URL
	}
	var items []item
	total := 0
	for _, u := range s.Upstreams {
		w := u.Weight
		if w <= 0 {
			continue
		}
		pu, err := url.Parse(u.BaseURL)
		if err != nil || pu.Scheme == "" || pu.Host == "" {
			continue
		}
		items = append(items, item{u: u, pu: pu})
		total += w
	}
	if total == 0 {
		def := strings.TrimSpace(s.Routing.DefaultUpstreamID)
		if def != "" {
			if u, ok := byID[def]; ok {
				pu, err := url.Parse(u.BaseURL)
				if err == nil && pu.Scheme != "" && pu.Host != "" {
					return pu, u.ID, pu.Host, nil
				}
			}
		}
		u := s.Upstreams[0]
		pu, err := url.Parse(u.BaseURL)
		if err != nil {
			return nil, "", "", err
		}
		return pu, u.ID, pu.Host, nil
	}
	r := p.rng.IntN(total)
	acc := 0
	for _, it := range items {
		acc += it.u.Weight
		if r < acc {
			return it.pu, it.u.ID, it.pu.Host, nil
		}
	}
	u := items[len(items)-1]
	return u.pu, u.u.ID, u.pu.Host, nil
}
