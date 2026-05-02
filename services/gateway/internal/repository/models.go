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
