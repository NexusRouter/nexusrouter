package repository

import "time"

// ModelVendor 对应表 model_vendor。
type ModelVendor struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement"`
	VendorName string `gorm:"column:vendor_name;size:64;not null"`
	VendorType int8   `gorm:"column:vendor_type;not null"` // 1 官方 2 第三方
	VendorCode string `gorm:"column:vendor_code;size:32;uniqueIndex"`
	Logo       string `gorm:"column:logo;size:512"`
	Status     int8   `gorm:"column:status;default:1"` // 1 启用 0 禁用
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (ModelVendor) TableName() string { return "model_vendor" }

// ModelBase 对应表 model_base（逻辑模型）。
type ModelBase struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement"`
	ModelName  string `gorm:"column:model_name;size:64;not null"`
	ModelCode  string `gorm:"column:model_code;size:64;not null;uniqueIndex"`
	ModelType  int8   `gorm:"column:model_type;not null"`  // 1 对话 2 Embedding 3 图像 4 语音
	Capability string `gorm:"column:capability;type:text"` // JSON
	Sort       int    `gorm:"column:sort;default:0"`
	Status     int8   `gorm:"column:status;default:1"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (ModelBase) TableName() string { return "model_base" }

// ModelUpstream 对应表 model_upstream。
type ModelUpstream struct {
	ID            int64  `gorm:"column:id;primaryKey;autoIncrement"`
	VendorID      int64  `gorm:"column:vendor_id;not null;index"`
	UpstreamName  string `gorm:"column:upstream_name;size:64;not null"`
	BaseURL       string `gorm:"column:base_url;size:512;not null"`
	APIKey        string `gorm:"column:api_key;size:512"`
	Timeout       int    `gorm:"column:timeout;default:30"`
	MaxConcurrent int    `gorm:"column:max_concurrent;default:100"`
	Status        int8   `gorm:"column:status;default:1"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (ModelUpstream) TableName() string { return "model_upstream" }

// ModelInstance 对应表 model_instance。
type ModelInstance struct {
	ID                int64  `gorm:"column:id;primaryKey;autoIncrement"`
	BaseModelID       int64  `gorm:"column:base_model_id;not null;index"`
	VendorID          int64  `gorm:"column:vendor_id;not null;index"`
	UpstreamID        int64  `gorm:"column:upstream_id;not null;index"` // FK → model_upstream.id
	InstanceName      string `gorm:"column:instance_name;size:64;not null"`
	ProviderModelCode string `gorm:"column:provider_model_code;size:64;not null"`
	Weight            int    `gorm:"column:weight;default:10"`
	Priority          int8   `gorm:"column:priority;default:1"` // 1 高 2 中 3 低
	IsOfficial        int8   `gorm:"column:is_official;default:0"`
	Status            int8   `gorm:"column:status;default:1"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (ModelInstance) TableName() string { return "model_instance" }
