package repository

import (
	"errors"
	"os"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BootstrapFromConfig 在空库时从遗留 gateway.yaml 与环境变量管理员导入（幂等）。
// API Key 导入由 keystore.BootstrapKeysIfEmpty 单独完成，避免包循环依赖。
func BootstrapFromConfig(cfg *config.Config, db *gorm.DB, log *zap.Logger) error {
	if cfg == nil || db == nil {
		return nil
	}
	if log == nil {
		log = zap.NewNop()
	}
	if err := seedGatewayYAML(cfg, db, log); err != nil {
		return err
	}
	if err := seedAdminUsers(cfg, db, log); err != nil {
		return err
	}
	return nil
}

func seedGatewayYAML(cfg *config.Config, db *gorm.DB, log *zap.Logger) error {
	var row GatewaySnapshotRow
	err := db.First(&row, 1).Error
	if err == nil && strings.TrimSpace(row.YAMLBody) != "" {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
		return err
	}
	path := strings.TrimSpace(cfg.GatewayConfigFile)
	if path == "" {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		log.Warn("跳过网关 YAML 导入：文件不可读", zap.String("path", path), zap.Error(err))
		return nil
	}
	body := string(b)
	row = GatewaySnapshotRow{ID: 1, YAMLBody: body}
	if err := db.Save(&row).Error; err != nil {
		return err
	}
	log.Info("已从 gateway.yaml 导入到数据库", zap.String("path", path))
	return nil
}

func seedAdminUsers(cfg *config.Config, db *gorm.DB, log *zap.Logger) error {
	var n int64
	if err := db.Model(&AdminUserModel{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	u := strings.TrimSpace(cfg.AdminUsername)
	h := strings.TrimSpace(cfg.AdminPasswordBcrypt)
	if u != "" && h != "" {
		if err := db.Create(&AdminUserModel{Username: u, Role: "admin", PasswordBcrypt: h}).Error; err != nil {
			return err
		}
		log.Info("已从环境变量导入管理员账号到数据库", zap.String("username", u))
	}
	ou := strings.TrimSpace(cfg.AdminOperatorUsername)
	oh := strings.TrimSpace(cfg.AdminOperatorPasswordBcrypt)
	if ou != "" && oh != "" {
		if err := db.Create(&AdminUserModel{Username: ou, Role: "operator", PasswordBcrypt: oh}).Error; err != nil {
			return err
		}
		log.Info("已从环境变量导入操作员账号到数据库", zap.String("username", ou))
	}
	return nil
}
