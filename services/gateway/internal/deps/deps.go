// Package deps 将规范约定的第三方库固定进 go.mod；业务代码落地后可逐步移除对应空白导入。
package deps

import (
	_ "github.com/go-playground/validator/v10"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/redis/go-redis/v9"
	_ "github.com/spf13/viper"
	_ "github.com/swaggo/swag"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/gorm"
)
