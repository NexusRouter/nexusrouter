package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UseDatabaseModelLibrary 当存在至少一条 model_instance 记录时，Chat/List 仅走库表上游（不与 gateway.yaml 混用）。
func UseDatabaseModelLibrary(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	var n int64
	if err := db.Model(&ModelInstance{}).Where("status = ?", 1).Count(&n).Error; err != nil {
		return false
	}
	return n > 0
}

// ChatPickResult 选中实例后的转发参数。
type ChatPickResult struct {
	BaseURL           *url.URL
	APIKey            string
	ProviderModelCode string
	InstanceID        int64
	UpstreamHost      string
	Timeout           time.Duration
}

type instanceCandidate struct {
	Inst     ModelInstance
	Upstream ModelUpstream
}

// PickChatTarget 按 model_code（逻辑名）选择一条实例；无可用时返回 err。
func PickChatTarget(db *gorm.DB, logicalModelCode string) (*ChatPickResult, error) {
	if db == nil {
		return nil, errors.New("repository: db 为空")
	}
	logicalModelCode = strings.TrimSpace(logicalModelCode)
	if logicalModelCode == "" {
		return nil, errors.New("repository: model 为空")
	}
	var base ModelBase
	if err := db.Where("model_code = ? AND status = ?", logicalModelCode, 1).First(&base).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("repository: 逻辑模型不存在: %w", err)
		}
		return nil, err
	}
	var rows []ModelInstance
	if err := db.Where("base_model_id = ? AND status = ?", base.ID, 1).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("repository: 无可用实例")
	}
	var candidates []instanceCandidate
	for _, ins := range rows {
		var v ModelVendor
		if err := db.Where("id = ? AND status = ?", ins.VendorID, 1).First(&v).Error; err != nil {
			continue
		}
		var up ModelUpstream
		if err := db.Where("id = ? AND status = ?", ins.UpstreamID, 1).First(&up).Error; err != nil {
			continue
		}
		candidates = append(candidates, instanceCandidate{Inst: ins, Upstream: up})
	}
	if len(candidates) == 0 {
		return nil, errors.New("repository: 无启用链路的实例")
	}
	chosen := pickWeightedInstance(candidates)
	if chosen == nil {
		return nil, errors.New("repository: 选择实例失败")
	}
	pu, err := url.Parse(strings.TrimSpace(chosen.Upstream.BaseURL))
	if err != nil || pu.Scheme == "" || pu.Host == "" {
		return nil, errors.New("repository: 上游 base_url 非法")
	}
	to := time.Duration(chosen.Upstream.Timeout) * time.Second
	if to <= 0 {
		to = 30 * time.Second
	}
	return &ChatPickResult{
		BaseURL:           pu,
		APIKey:            strings.TrimSpace(chosen.Upstream.APIKey),
		ProviderModelCode: strings.TrimSpace(chosen.Inst.ProviderModelCode),
		InstanceID:        chosen.Inst.ID,
		UpstreamHost:      pu.Host,
		Timeout:           to,
	}, nil
}

// pickWeightedInstance：先按 priority 数值升序取最小档；该档内 is_official=1 优先于 0；同档同官方标志下按 weight 加权随机。
func pickWeightedInstance(cands []instanceCandidate) *instanceCandidate {
	if len(cands) == 0 {
		return nil
	}
	minP := cands[0].Inst.Priority
	for _, c := range cands[1:] {
		if c.Inst.Priority < minP {
			minP = c.Inst.Priority
		}
	}
	var tier []instanceCandidate
	for _, c := range cands {
		if c.Inst.Priority == minP {
			tier = append(tier, c)
		}
	}
	official := make([]instanceCandidate, 0)
	other := make([]instanceCandidate, 0)
	for _, c := range tier {
		if c.Inst.IsOfficial == 1 {
			official = append(official, c)
		} else {
			other = append(other, c)
		}
	}
	pool := official
	if len(pool) == 0 {
		pool = other
	}
	if len(pool) == 0 {
		return nil
	}
	return weightedPick(pool, func(c instanceCandidate) int {
		w := c.Inst.Weight
		if w <= 0 {
			return 1
		}
		return w
	})
}

func weightedPick(cands []instanceCandidate, weight func(instanceCandidate) int) *instanceCandidate {
	if len(cands) == 1 {
		return &cands[0]
	}
	total := 0
	for _, c := range cands {
		total += weight(c)
	}
	if total <= 0 {
		return &cands[0]
	}
	r := rand.IntN(total)
	acc := 0
	for i := range cands {
		acc += weight(cands[i])
		if r < acc {
			return &cands[i]
		}
	}
	return &cands[len(cands)-1]
}

// PublishedAggRow GET /v1/models 单行。
type PublishedAggRow struct {
	ModelCode string
	OwnedBy   string
	CreatedAt int64
}

// ListPublishedModelsAggregation 启用链路上的逻辑模型。
func ListPublishedModelsAggregation(db *gorm.DB) ([]PublishedAggRow, error) {
	if db == nil {
		return nil, errors.New("repository: db 为空")
	}
	var bases []ModelBase
	err := db.Raw(`
SELECT DISTINCT b.id, b.model_name, b.model_code, b.model_type, b.capability, b.sort, b.status, b.created_at, b.updated_at
FROM model_base b
INNER JOIN model_instance i ON i.base_model_id = b.id AND i.status = 1
INNER JOIN model_vendor v ON v.id = i.vendor_id AND v.status = 1
INNER JOIN model_upstream u ON u.id = i.upstream_id AND u.status = 1
WHERE b.status = 1
ORDER BY b.sort ASC, b.id ASC
`).Scan(&bases).Error
	if err != nil {
		return nil, err
	}
	out := make([]PublishedAggRow, 0, len(bases))
	for _, b := range bases {
		ob := ""
		var inst0 ModelInstance
		if err := db.Where("base_model_id = ? AND status = ?", b.ID, 1).First(&inst0).Error; err == nil {
			var v ModelVendor
			if err := db.First(&v, inst0.VendorID).Error; err == nil {
				ob = v.VendorName
			}
		}
		out = append(out, PublishedAggRow{
			ModelCode: b.ModelCode,
			OwnedBy:   ob,
			CreatedAt: b.CreatedAt.Unix(),
		})
	}
	return out, nil
}

// GetPublishedModelBase 若 model_code 在公开列表中则返回 base。
func GetPublishedModelBase(db *gorm.DB, modelCode string) (*ModelBase, error) {
	if db == nil {
		return nil, errors.New("repository: db 为空")
	}
	modelCode = strings.TrimSpace(modelCode)
	var base ModelBase
	if err := db.Where("model_code = ? AND status = ?", modelCode, 1).First(&base).Error; err != nil {
		return nil, err
	}
	var n int64
	err := db.Raw(`
SELECT COUNT(*) FROM model_instance i
INNER JOIN model_vendor v ON v.id = i.vendor_id AND v.status = 1
INNER JOIN model_upstream u ON u.id = i.upstream_id AND u.status = 1
WHERE i.base_model_id = ? AND i.status = 1
`, base.ID).Scan(&n).Error
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &base, nil
}

// RewriteChatBodyToProvider 将 JSON body 中 model 替换为 providerModelCode。
func RewriteChatBodyToProvider(body []byte, providerModelCode string) []byte {
	providerModelCode = strings.TrimSpace(providerModelCode)
	if len(body) == 0 || providerModelCode == "" {
		return body
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return body
	}
	if _, ok := obj["model"]; !ok {
		return body
	}
	newRaw, err := json.Marshal(providerModelCode)
	if err != nil {
		return body
	}
	obj["model"] = newRaw
	b, err := json.Marshal(obj)
	if err != nil {
		return body
	}
	return b
}
