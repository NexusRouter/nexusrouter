package keystore

import (
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BootstrapKeysIfEmpty 在 api_keys 为空时从密钥文件或遗留环境变量导入。
func BootstrapKeysIfEmpty(cfg *config.Config, db *gorm.DB, log *zap.Logger) error {
	if cfg == nil || db == nil {
		return nil
	}
	if log == nil {
		log = zap.NewNop()
	}
	var n int64
	if err := db.Model(&repository.APIKeyModel{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	path := strings.TrimSpace(cfg.GatewayKeysFile)
	if path != "" {
		recs, err := LoadRecordsFromFile(path)
		if err != nil {
			log.Warn("跳过密钥 JSON 导入：解析失败", zap.String("path", path), zap.Error(err))
			return nil
		}
		if err := repository.ReplaceAllAPIKeyModels(db, toAPIKeyModels(recs)); err != nil {
			return err
		}
		log.Info("已从密钥文件导入到数据库", zap.String("path", path), zap.Int("count", len(recs)))
		return nil
	}
	recs := RecordsFromLegacy(cfg.GatewayAPIKeys)
	if len(recs) == 0 {
		return nil
	}
	if err := repository.ReplaceAllAPIKeyModels(db, toAPIKeyModels(recs)); err != nil {
		return err
	}
	log.Info("已从环境变量遗留密钥导入到数据库", zap.Int("count", len(recs)))
	return nil
}
