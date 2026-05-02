package repository

import (
	"fmt"

	"gorm.io/gorm"
)

// ReplaceAllAPIKeyModels 事务内清空并写入 api_keys。
func ReplaceAllAPIKeyModels(db *gorm.DB, models []APIKeyModel) error {
	if db == nil {
		return fmt.Errorf("repository: db 为空")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&APIKeyModel{}).Error; err != nil {
			return err
		}
		for i := range models {
			if err := tx.Create(&models[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
