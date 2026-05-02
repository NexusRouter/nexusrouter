# gateway-backend Specification Delta

## MODIFIED Requirements

### Requirement: 后端依赖矩阵

模块 MUST 声明并可通过构建解析以下依赖（主模块或间接，以 `go.mod` 为准，版本约束与规范一致）：**gorm.io/gorm v1.25.x**、**gorm.io/driver/postgres v1.5.x**、**gorm.io/driver/sqlite v1.x**（或与 GORM 兼容的当前主版本）、**github.com/redis/go-redis/v9 v9.x**、**github.com/google/wire v0.6.x**、**github.com/spf13/viper v1.19.x**、**go.uber.org/zap v1.27.x**、**github.com/golang-jwt/jwt/v5 v5.x**、**github.com/go-playground/validator/v10 v10.x**、**github.com/golang-migrate/migrate/v4 v4.x**、**github.com/swaggo/swag v1.16.x**、**github.com/stretchr/testify v1.10.x**。

#### Scenario: 编译通过

- **WHEN** 在 `services/gateway` 执行 `go build ./...`
- **THEN** 以零退出码完成
