package adminauth

import (
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"gorm.io/gorm"
)

// IsConsoleConfigured 管理控制台是否具备最小可登录配置（JWT + 管理员来源：数据库或环境变量）。
func IsConsoleConfigured(cfg *config.Config, db *gorm.DB) bool {
	if cfg == nil || !cfg.EnableAdminConsole {
		return false
	}
	if strings.TrimSpace(cfg.AdminJWTSecret) == "" {
		return false
	}
	if db != nil {
		var boot repository.SystemBootstrapRow
		if err := db.Select("initialized").First(&boot, repository.SystemBootstrapSingletonPK).Error; err == nil && !boot.Initialized {
			// 未完成首次向导：仍注册管理路由（登录/说明）以便向导完成后立即可用；业务接口由门闸拦截。
			return true
		}
		var n int64
		if err := db.Model(&repository.AdminUserModel{}).Where("role = ?", "admin").Count(&n).Error; err == nil && n > 0 {
			return true
		}
	}
	return strings.TrimSpace(cfg.AdminUsername) != "" && strings.TrimSpace(cfg.AdminPasswordBcrypt) != ""
}
