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
		var n int64
		if err := db.Model(&repository.AdminUserModel{}).Where("role = ?", "admin").Count(&n).Error; err == nil && n > 0 {
			return true
		}
	}
	return strings.TrimSpace(cfg.AdminUsername) != "" && strings.TrimSpace(cfg.AdminPasswordBcrypt) != ""
}
