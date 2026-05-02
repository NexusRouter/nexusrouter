package repository

import "time"

// GatewaySnapshotRow 单行持久化网关配置片段（与 runtime marshalFileYAML 格式一致）。
type GatewaySnapshotRow struct {
	ID        uint `gorm:"primaryKey"`
	YAMLBody  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 显式表名。
func (GatewaySnapshotRow) TableName() string { return "gateway_snapshots" }

// APIKeyModel API Key 表行。
type APIKeyModel struct {
	KeyID     string `gorm:"column:key_id;primaryKey;size:191"`
	Secret    string `gorm:"size:512"`
	Disabled  bool
	ExpiresAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 显式表名。
func (APIKeyModel) TableName() string { return "api_keys" }

// AdminUserModel 管理控制台登录用户（admin / operator）。
type AdminUserModel struct {
	Username       string `gorm:"primaryKey;size:191"`
	Role           string `gorm:"size:32"` // admin | operator
	PasswordBcrypt string `gorm:"size:255"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TableName 显式表名。
func (AdminUserModel) TableName() string { return "admin_users" }

// SystemBootstrapSingletonPK 全局引导状态行主键，固定为 1。
const SystemBootstrapSingletonPK uint = 1

// SystemBootstrapRow 全局首次初始化状态（单行，主键固定为 1）。
type SystemBootstrapRow struct {
	ID              uint `gorm:"primaryKey"`
	Initialized     bool
	InitInProgress  bool
	InitStartedAt   *time.Time
	SiteDisplayName string `gorm:"size:255"`
	UpdatedAt       time.Time
}

// TableName 显式表名。
func (SystemBootstrapRow) TableName() string { return "system_bootstrap" }

// ModelCatalogEntry 模型库目录项（逻辑模型 id，与客户端请求 model 字段对齐）。
type ModelCatalogEntry struct {
	ID          string `gorm:"column:id;primaryKey;size:191"`
	DisplayName string `gorm:"size:255"`
	OwnedBy     string `gorm:"size:191"`
	Metadata    string `gorm:"type:text"` // 可选 JSON
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName 显式表名。
func (ModelCatalogEntry) TableName() string { return "model_catalog_entries" }

// ModelUpstreamBinding 目录项与上游 id 的绑定；同一 (catalog_entry_id, upstream_id) 唯一。
type ModelUpstreamBinding struct {
	ID             uint   `gorm:"primaryKey"`
	CatalogEntryID string `gorm:"size:191;uniqueIndex:ux_binding_cat_up"`
	UpstreamID     string `gorm:"size:191;uniqueIndex:ux_binding_cat_up"`
	Enabled        bool
	Priority       int64
	ActualModel    *string `gorm:"size:191"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TableName 显式表名。
func (ModelUpstreamBinding) TableName() string { return "model_upstream_bindings" }
