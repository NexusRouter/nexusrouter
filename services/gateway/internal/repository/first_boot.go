package repository

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// InitInProgressTTL 初始化进行中标记的超时时间，超时后允许新的初始化尝试。
	InitInProgressTTL = 2 * time.Minute
)

var (
	// ErrBootstrapAlreadyCompleted 系统已完成首次初始化。
	ErrBootstrapAlreadyCompleted = errors.New("bootstrap: 已完成初始化")
	// ErrBootstrapInProgress 另一请求正在执行初始化。
	ErrBootstrapInProgress = errors.New("bootstrap: 初始化进行中")
)

// BootstrapPhase 供前端展示的阶段枚举。
type BootstrapPhase string

const (
	BootstrapPhaseReady        BootstrapPhase = "ready"
	BootstrapPhaseInitializing BootstrapPhase = "initializing"
	BootstrapPhaseCompleted    BootstrapPhase = "completed"
)

// BootstrapStatusDTO 初始化状态查询结果。
type BootstrapStatusDTO struct {
	Initialized bool
	Phase       BootstrapPhase
}

// EnsureSystemBootstrap 确保存在默认引导行，并在遗留数据（已有管理员）时同步为已初始化。
func EnsureSystemBootstrap(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	var row SystemBootstrapRow
	err := db.First(&row, SystemBootstrapSingletonPK).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = SystemBootstrapRow{ID: SystemBootstrapSingletonPK, Initialized: false}
		if err := db.Create(&row).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	var adminCount int64
	if err := db.Model(&AdminUserModel{}).Where("role = ?", "admin").Count(&adminCount).Error; err != nil {
		return err
	}
	if adminCount > 0 && !row.Initialized {
		if err := db.Model(&SystemBootstrapRow{}).Where("id = ?", SystemBootstrapSingletonPK).Updates(map[string]any{
			"initialized":      true,
			"init_in_progress": false,
			"init_started_at":  nil,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// IsSystemInitialized 快速读取是否已完成首次初始化（无事务）。
func IsSystemInitialized(db *gorm.DB) (bool, error) {
	if db == nil {
		return true, nil
	}
	var row SystemBootstrapRow
	if err := db.Select("initialized").First(&row, SystemBootstrapSingletonPK).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, err
	}
	return row.Initialized, nil
}

// GetBootstrapStatus 返回初始化状态与阶段，并在进行中标记过期时惰性清理。
func GetBootstrapStatus(db *gorm.DB, now time.Time) (BootstrapStatusDTO, error) {
	if db == nil {
		return BootstrapStatusDTO{Initialized: true, Phase: BootstrapPhaseCompleted}, nil
	}
	return getBootstrapStatusTx(db, now)
}

func getBootstrapStatusTx(db *gorm.DB, now time.Time) (BootstrapStatusDTO, error) {
	var out BootstrapStatusDTO
	err := db.Transaction(func(tx *gorm.DB) error {
		var row SystemBootstrapRow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, SystemBootstrapSingletonPK).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				out = BootstrapStatusDTO{Initialized: true, Phase: BootstrapPhaseCompleted}
				return nil
			}
			return err
		}
		if row.Initialized {
			out = BootstrapStatusDTO{Initialized: true, Phase: BootstrapPhaseCompleted}
			return nil
		}
		if row.InitInProgress && row.InitStartedAt != nil && now.Sub(*row.InitStartedAt) < InitInProgressTTL {
			out = BootstrapStatusDTO{Initialized: false, Phase: BootstrapPhaseInitializing}
			return nil
		}
		if row.InitInProgress {
			if err := tx.Model(&SystemBootstrapRow{}).Where("id = ?", SystemBootstrapSingletonPK).Updates(map[string]any{
				"init_in_progress": false,
				"init_started_at":  nil,
			}).Error; err != nil {
				return err
			}
		}
		out = BootstrapStatusDTO{Initialized: false, Phase: BootstrapPhaseReady}
		return nil
	})
	return out, err
}

// CompleteFirstBootInput 向导提交载荷。
type CompleteFirstBootInput struct {
	AdminUsername   string
	AdminPassword   string
	SiteDisplayName string
	BcryptCost      int // 测试可调低；0 表示默认
}

// CompleteFirstBoot 在单事务内完成首次初始化（至多一次成功）。
func CompleteFirstBoot(db *gorm.DB, now time.Time, in CompleteFirstBootInput) error {
	if db == nil {
		return errors.New("bootstrap: db 为空")
	}
	u := strings.TrimSpace(in.AdminUsername)
	p := in.AdminPassword
	if u == "" || len(u) > 191 {
		return errors.New("bootstrap: 管理员用户名无效")
	}
	if len(p) < 8 {
		return errors.New("bootstrap: 密码过短")
	}
	site := strings.TrimSpace(in.SiteDisplayName)
	if len(site) > 255 {
		return errors.New("bootstrap: 站点显示名过长")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var row SystemBootstrapRow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, SystemBootstrapSingletonPK).Error; err != nil {
			return err
		}
		if row.Initialized {
			return ErrBootstrapAlreadyCompleted
		}
		if row.InitInProgress && row.InitStartedAt != nil && now.Sub(*row.InitStartedAt) < InitInProgressTTL {
			return ErrBootstrapInProgress
		}
		if row.InitInProgress {
			if err := tx.Model(&SystemBootstrapRow{}).Where("id = ?", SystemBootstrapSingletonPK).Updates(map[string]any{
				"init_in_progress": false,
				"init_started_at":  nil,
			}).Error; err != nil {
				return err
			}
		}
		started := now
		if err := tx.Model(&SystemBootstrapRow{}).Where("id = ?", SystemBootstrapSingletonPK).Updates(map[string]any{
			"init_in_progress": true,
			"init_started_at":  started,
		}).Error; err != nil {
			return err
		}
		var n int64
		if err := tx.Model(&AdminUserModel{}).Where("username = ?", u).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return errors.New("bootstrap: 用户名已存在")
		}
		cost := bcrypt.DefaultCost
		if in.BcryptCost >= bcrypt.MinCost && in.BcryptCost <= bcrypt.MaxCost {
			cost = in.BcryptCost
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(p), cost)
		if err != nil {
			return err
		}
		admin := AdminUserModel{
			Username:       u,
			Role:           "admin",
			PasswordBcrypt: string(hash),
		}
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}
		if err := tx.Model(&SystemBootstrapRow{}).Where("id = ?", SystemBootstrapSingletonPK).Updates(map[string]any{
			"initialized":       true,
			"init_in_progress":  false,
			"init_started_at":   nil,
			"site_display_name": site,
		}).Error; err != nil {
			return err
		}
		return nil
	})
}

// ResetFirstBoot 将系统恢复为未初始化并删除全部管理用户（需由 handler 校验超管身份）。
func ResetFirstBoot(db *gorm.DB) error {
	if db == nil {
		return errors.New("bootstrap: db 为空")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&AdminUserModel{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&SystemBootstrapRow{}).Where("id = ?", SystemBootstrapSingletonPK).Updates(map[string]any{
			"initialized":       false,
			"init_in_progress":  false,
			"init_started_at":   nil,
			"site_display_name": "",
		}).Error; err != nil {
			return err
		}
		return nil
	})
}
