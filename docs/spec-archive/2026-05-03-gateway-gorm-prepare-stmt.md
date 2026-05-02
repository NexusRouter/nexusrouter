# 网关 OpenDB 启用 GORM 预编译语句

## 行为

**`internal/repository.OpenDB`** 在打开 **Postgres** 或 **SQLite** 时，向 **`gorm.Open`** 传入的 **`gorm.Config`** MUST 将 **`PrepareStmt`** 设为 **`true`**，并保持既有静默日志配置；从而在进程内复用预编译语句、降低重复 SQL 解析成本。测试代码使用独立内存 DSN 时可不强制相同配置。
