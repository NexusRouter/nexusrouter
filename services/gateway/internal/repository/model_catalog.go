package repository

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// ListModelCatalogEntries 分页列出目录项（按 id 升序）。
func ListModelCatalogEntries(db *gorm.DB, offset, limit int) ([]ModelCatalogEntry, int64, error) {
	if db == nil {
		return nil, 0, errors.New("repository: db 为空")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	var total int64
	if err := db.Model(&ModelCatalogEntry{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ModelCatalogEntry
	err := db.Order("id asc").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

// GetModelCatalogEntry 按主键读取。
func GetModelCatalogEntry(db *gorm.DB, id string) (*ModelCatalogEntry, error) {
	if db == nil {
		return nil, errors.New("repository: db 为空")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var row ModelCatalogEntry
	err := db.Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// CreateModelCatalogEntry 插入目录项；主键冲突返回 err。
func CreateModelCatalogEntry(db *gorm.DB, row *ModelCatalogEntry) error {
	if db == nil || row == nil {
		return errors.New("repository: 参数无效")
	}
	return db.Create(row).Error
}

// UpdateModelCatalogEntry 更新展示字段。
func UpdateModelCatalogEntry(db *gorm.DB, id string, displayName, ownedBy, metadata string) error {
	if db == nil {
		return errors.New("repository: db 为空")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return gorm.ErrRecordNotFound
	}
	res := db.Model(&ModelCatalogEntry{}).Where("id = ?", id).Updates(map[string]any{
		"display_name": displayName,
		"owned_by":     ownedBy,
		"metadata":     metadata,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteModelCatalogEntry 删除目录项及其绑定。
func DeleteModelCatalogEntry(db *gorm.DB, id string) error {
	if db == nil {
		return errors.New("repository: db 为空")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return gorm.ErrRecordNotFound
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("catalog_entry_id = ?", id).Delete(&ModelUpstreamBinding{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ?", id).Delete(&ModelCatalogEntry{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

// ListBindingsForCatalog 某目录项下全部绑定。
func ListBindingsForCatalog(db *gorm.DB, catalogID string) ([]ModelUpstreamBinding, error) {
	if db == nil {
		return nil, errors.New("repository: db 为空")
	}
	catalogID = strings.TrimSpace(catalogID)
	var rows []ModelUpstreamBinding
	err := db.Where("catalog_entry_id = ?", catalogID).Order("id asc").Find(&rows).Error
	return rows, err
}

// CreateModelUpstreamBinding 插入绑定；唯一约束冲突返回 err。
func CreateModelUpstreamBinding(db *gorm.DB, row *ModelUpstreamBinding) error {
	if db == nil || row == nil {
		return errors.New("repository: 参数无效")
	}
	return db.Create(row).Error
}

// UpdateModelUpstreamBinding 按绑定 id 更新。
func UpdateModelUpstreamBinding(db *gorm.DB, id uint, enabled bool, priority int64, actual *string) error {
	if db == nil {
		return errors.New("repository: db 为空")
	}
	res := db.Model(&ModelUpstreamBinding{}).Where("id = ?", id).Updates(map[string]any{
		"enabled":      enabled,
		"priority":     priority,
		"actual_model": actual,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteModelUpstreamBinding 删除单条绑定。
func DeleteModelUpstreamBinding(db *gorm.DB, id uint) error {
	if db == nil {
		return errors.New("repository: db 为空")
	}
	res := db.Delete(&ModelUpstreamBinding{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetModelUpstreamBinding 读取单条绑定。
func GetModelUpstreamBinding(db *gorm.DB, id uint) (*ModelUpstreamBinding, error) {
	if db == nil {
		return nil, errors.New("repository: db 为空")
	}
	var row ModelUpstreamBinding
	err := db.First(&row, id).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// PublishedModelRow 聚合后的公开模型行（JOIN 结果）。
type PublishedModelRow struct {
	CatalogID   string
	DisplayName string
	OwnedBy     string
	CreatedAt   int64
}

// ListPublishedModels 返回：启用绑定且 upstream_id 落在 validUpstream 集合中的目录项（按 catalog id 去重，取最小 binding id 优先）。
func ListPublishedModels(db *gorm.DB, validUpstream map[string]struct{}) ([]PublishedModelRow, error) {
	if db == nil {
		return nil, errors.New("repository: db 为空")
	}
	if len(validUpstream) == 0 {
		return nil, nil
	}
	var bindings []ModelUpstreamBinding
	if err := db.Where("enabled = ?", true).Find(&bindings).Error; err != nil {
		return nil, err
	}
	seenCat := make(map[string]struct{})
	var catalogIDs []string
	for _, b := range bindings {
		if _, ok := validUpstream[b.UpstreamID]; !ok {
			continue
		}
		if _, ok := seenCat[b.CatalogEntryID]; ok {
			continue
		}
		seenCat[b.CatalogEntryID] = struct{}{}
		catalogIDs = append(catalogIDs, b.CatalogEntryID)
	}
	if len(catalogIDs) == 0 {
		return nil, nil
	}
	var entries []ModelCatalogEntry
	if err := db.Where("id IN ?", catalogIDs).Order("id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	out := make([]PublishedModelRow, 0, len(entries))
	for _, e := range entries {
		out = append(out, PublishedModelRow{
			CatalogID:   e.ID,
			DisplayName: e.DisplayName,
			OwnedBy:     e.OwnedBy,
			CreatedAt:   e.CreatedAt.Unix(),
		})
	}
	return out, nil
}

// IsCatalogPublished 若存在指向有效上游的启用绑定则 true。
func IsCatalogPublished(db *gorm.DB, catalogID string, validUpstream map[string]struct{}) (bool, *ModelCatalogEntry, error) {
	if db == nil {
		return false, nil, errors.New("repository: db 为空")
	}
	catalogID = strings.TrimSpace(catalogID)
	ent, err := GetModelCatalogEntry(db, catalogID)
	if err != nil {
		return false, nil, err
	}
	var n int64
	q := db.Model(&ModelUpstreamBinding{}).Where("catalog_entry_id = ? AND enabled = ?", catalogID, true)
	ids := make([]string, 0, len(validUpstream))
	for uid := range validUpstream {
		ids = append(ids, uid)
	}
	if len(ids) == 0 {
		return false, ent, nil
	}
	if err := q.Where("upstream_id IN ?", ids).Count(&n).Error; err != nil {
		return false, nil, err
	}
	return n > 0, ent, nil
}
