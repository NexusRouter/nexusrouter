// Package keystore 负责网关 API Key 的加载、热更新与校验。
package keystore

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"go.uber.org/zap"
)

// Record 为单条密钥的运行时表示（由 JSON 或遗留 env 列表构造）。
type Record struct {
	ID        string
	Secret    string
	Disabled  bool
	ExpiresAt *time.Time
}

// Store 在内存中保存密钥快照，支持原子替换以便热加载。
type Store struct {
	log  *zap.Logger
	path string // 非空表示从文件加载并可 Reload

	records atomic.Pointer[[]Record]
}

// New 根据配置构造密钥库：优先 JSON 文件，否则使用逗号分隔的遗留 env 密钥。
func New(cfg *config.Config, log *zap.Logger) (*Store, error) {
	if log == nil {
		log = zap.NewNop()
	}
	s := &Store{log: log, path: strings.TrimSpace(cfg.GatewayKeysFile)}
	if s.path != "" {
		recs, err := loadRecordsFromFile(s.path)
		if err != nil {
			return nil, err
		}
		s.replace(recs)
		return s, nil
	}
	s.replace(recordsFromLegacy(cfg.GatewayAPIKeys))
	return s, nil
}

func recordsFromLegacy(keys []string) []Record {
	out := make([]Record, 0, len(keys))
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		out = append(out, Record{
			ID:       "",
			Secret:   k,
			Disabled: false,
		})
	}
	return out
}

type fileRecord struct {
	ID        string  `json:"id"`
	Secret    string  `json:"secret"`
	Disabled  bool    `json:"disabled"`
	ExpiresAt *string `json:"expires_at"`
}

func loadRecordsFromFile(path string) ([]Record, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw []fileRecord
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	out := make([]Record, 0, len(raw))
	for _, r := range raw {
		sec := strings.TrimSpace(r.Secret)
		if sec == "" {
			continue
		}
		rec := Record{
			ID:       strings.TrimSpace(r.ID),
			Secret:   sec,
			Disabled: r.Disabled,
		}
		if r.ExpiresAt != nil && strings.TrimSpace(*r.ExpiresAt) != "" {
			t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(*r.ExpiresAt))
			if err != nil {
				t, err = time.Parse(time.RFC3339, strings.TrimSpace(*r.ExpiresAt))
			}
			if err != nil {
				return nil, err
			}
			tu := t.UTC()
			rec.ExpiresAt = &tu
		}
		out = append(out, rec)
	}
	if len(out) == 0 {
		return nil, errors.New("keystore: 密钥文件无有效记录")
	}
	return out, nil
}

func (s *Store) replace(recs []Record) {
	cp := make([]Record, len(recs))
	copy(cp, recs)
	s.records.Store(&cp)
}

// Reload 从配置文件重新加载（路径为空时为 no-op）。
func (s *Store) Reload() error {
	if strings.TrimSpace(s.path) == "" {
		return nil
	}
	recs, err := loadRecordsFromFile(s.path)
	if err != nil {
		return err
	}
	s.replace(recs)
	return nil
}

// Path 返回配置的密钥文件路径（可能为空）。
func (s *Store) Path() string { return s.path }

// HasKeys 是否存在至少一条可用于鉴权的密钥（启用且未过期）。
func (s *Store) HasKeys() bool {
	recs := s.snapshot()
	for i := range recs {
		if s.recordActive(&recs[i], time.Now().UTC()) {
			return true
		}
	}
	return false
}

func (s *Store) snapshot() []Record {
	p := s.records.Load()
	if p == nil {
		return nil
	}
	return *p
}

func (s *Store) recordActive(r *Record, now time.Time) bool {
	if r.Disabled {
		return false
	}
	if r.ExpiresAt != nil && !now.Before(*r.ExpiresAt) {
		return false
	}
	return strings.TrimSpace(r.Secret) != ""
}

// ValidateBearer 校验 Bearer 令牌；token 为已去掉 Bearer 前缀的明文。
func (s *Store) ValidateBearer(token string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}
	now := time.Now().UTC()
	for _, r := range s.snapshot() {
		if !s.recordActive(&r, now) {
			continue
		}
		if r.Secret == token {
			return true
		}
	}
	return false
}

// ValidateXAPIKey 校验 X-API-Key（deprecated 兼容路径）。
func (s *Store) ValidateXAPIKey(key string) bool {
	return s.ValidateBearer(key)
}
