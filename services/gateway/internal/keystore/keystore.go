// Package keystore 负责网关 API Key 的加载、热更新与校验。
package keystore

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Record 为单条密钥的运行时表示（由 JSON 或遗留 env 列表构造）。
type Record struct {
	ID        string
	Secret    string
	Disabled  bool
	ExpiresAt *time.Time
	CreatedAt *time.Time
}

// PublicRecord 对外展示（脱敏 secret）。
type PublicRecord struct {
	ID           string     `json:"id"`
	MaskedSecret string     `json:"masked_secret"`
	Disabled     bool       `json:"disabled"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
}

// Store 在内存中保存密钥快照，支持原子替换以便热加载。
type Store struct {
	log  *zap.Logger
	path string // 遗留密钥 JSON 路径（导入源或日志）；数据库模式下可为空
	db   *gorm.DB

	records atomic.Pointer[[]Record]
}

// New 根据配置构造密钥库：若提供 db 则从数据库加载；否则回退为 JSON 文件或遗留 env 密钥（测试用）。
func New(cfg *config.Config, log *zap.Logger, db *gorm.DB) (*Store, error) {
	if log == nil {
		log = zap.NewNop()
	}
	s := &Store{log: log, path: strings.TrimSpace(cfg.GatewayKeysFile), db: db}
	if db != nil {
		var rows []repository.APIKeyModel
		if err := db.Order("key_id").Find(&rows).Error; err != nil {
			return nil, err
		}
		s.replace(recordsFromModels(rows))
		return s, nil
	}
	if s.path != "" {
		recs, err := LoadRecordsFromFile(s.path)
		if err != nil {
			return nil, err
		}
		s.replace(recs)
		return s, nil
	}
	s.replace(RecordsFromLegacy(cfg.GatewayAPIKeys))
	return s, nil
}

// RecordsFromLegacy 将逗号分隔遗留密钥转为记录列表。
func RecordsFromLegacy(keys []string) []Record {
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
	CreatedAt *string `json:"created_at,omitempty"`
}

// LoadRecordsFromFile 从 JSON 文件加载密钥记录（供启动导入与测试）。
func LoadRecordsFromFile(path string) ([]Record, error) {
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
		if r.CreatedAt != nil && strings.TrimSpace(*r.CreatedAt) != "" {
			t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(*r.CreatedAt))
			if err != nil {
				t, err = time.Parse(time.RFC3339, strings.TrimSpace(*r.CreatedAt))
			}
			if err == nil {
				tu := t.UTC()
				rec.CreatedAt = &tu
			}
		}
		out = append(out, rec)
	}
	return out, nil
}

func (s *Store) replace(recs []Record) {
	cp := make([]Record, len(recs))
	copy(cp, recs)
	s.records.Store(&cp)
}

// Reload 从数据库或 JSON 文件重新加载。
func (s *Store) Reload() error {
	if s.db != nil {
		var rows []repository.APIKeyModel
		if err := s.db.Order("key_id").Find(&rows).Error; err != nil {
			return err
		}
		s.replace(recordsFromModels(rows))
		return nil
	}
	if strings.TrimSpace(s.path) == "" {
		return nil
	}
	recs, err := LoadRecordsFromFile(s.path)
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

// MaskSecret 脱敏展示 API Key。
func MaskSecret(secret string) string {
	s := strings.TrimSpace(secret)
	if s == "" {
		return ""
	}
	r := []rune(s)
	if len(r) <= 8 {
		return "****"
	}
	return string(r[:4]) + "…" + string(r[len(r)-4:])
}

// ListPublic 返回脱敏后的密钥列表（数据库模式或已配置密钥文件）。
func (s *Store) ListPublic() ([]PublicRecord, error) {
	if s == nil || (s.db == nil && strings.TrimSpace(s.path) == "") {
		return nil, fmt.Errorf("keystore: 未配置数据库或 NEXUSROUTER_GATEWAY_KEYS_FILE")
	}
	recs := s.snapshot()
	out := make([]PublicRecord, 0, len(recs))
	for _, r := range recs {
		id := strings.TrimSpace(r.ID)
		if id == "" {
			id = "(legacy)"
		}
		out = append(out, PublicRecord{
			ID:           id,
			MaskedSecret: MaskSecret(r.Secret),
			Disabled:     r.Disabled,
			ExpiresAt:    r.ExpiresAt,
			CreatedAt:    r.CreatedAt,
		})
	}
	return out, nil
}

// ReplaceAllRecords 写回数据库或 JSON 并替换内存快照。
func (s *Store) ReplaceAllRecords(recs []Record) error {
	if s == nil {
		return fmt.Errorf("keystore: Store 为空")
	}
	now := time.Now().UTC()
	norm := make([]Record, 0, len(recs))
	for _, r := range recs {
		sec := strings.TrimSpace(r.Secret)
		if sec == "" {
			continue
		}
		if strings.TrimSpace(r.ID) == "" {
			r.ID = newRandomID()
		}
		if r.CreatedAt == nil {
			t := now
			r.CreatedAt = &t
		}
		norm = append(norm, r)
	}
	if s.db != nil {
		if err := repository.ReplaceAllAPIKeyModels(s.db, toAPIKeyModels(norm)); err != nil {
			return err
		}
		s.replace(norm)
		return nil
	}
	if strings.TrimSpace(s.path) == "" {
		return fmt.Errorf("keystore: 未配置密钥文件")
	}
	raw := make([]fileRecord, 0, len(norm))
	for _, r := range norm {
		raw = append(raw, recordToFile(r))
	}
	b, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := writeJSONAtomic(s.path, b); err != nil {
		return err
	}
	s.replace(norm)
	return nil
}

func recordToFile(r Record) fileRecord {
	fr := fileRecord{
		ID:       strings.TrimSpace(r.ID),
		Secret:   r.Secret,
		Disabled: r.Disabled,
	}
	if r.ExpiresAt != nil {
		s := r.ExpiresAt.UTC().Format(time.RFC3339Nano)
		fr.ExpiresAt = &s
	}
	if r.CreatedAt != nil {
		s := r.CreatedAt.UTC().Format(time.RFC3339Nano)
		fr.CreatedAt = &s
	}
	return fr
}

func writeJSONAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".nexus-keys-*.json")
	if err != nil {
		return fmt.Errorf("keystore: 临时文件: %w", err)
	}
	tmp := f.Name()
	ok := false
	defer func() {
		if !ok {
			_ = os.Remove(tmp)
		}
	}()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("keystore: 替换密钥文件: %w", err)
	}
	ok = true
	return nil
}

func newRandomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "key_" + hex.EncodeToString(b)
}

// NewRandomSecret 生成随机 API Key 明文（hex）。
func NewRandomSecret() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// SnapshotRecords 返回当前密钥快照副本（仅供管理端修改前读取）。
func (s *Store) SnapshotRecords() []Record {
	recs := s.snapshot()
	out := make([]Record, len(recs))
	copy(out, recs)
	return out
}
