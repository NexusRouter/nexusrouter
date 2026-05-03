# 网关：持久化日志目录

## 功能说明

支持通过环境变量 **`NEXUSROUTER_LOG_DIR`** 或命令行 **`-log-dir <路径>`**（在已合并 **`.env`** 且环境中尚未设置 **`NEXUSROUTER_LOG_DIR`** 时生效）指定日志目录。目录下追加写入 **`gateway.log`**（JSON 行，**`Info`** 及以上级别），标准错误仍为可读控制台输出；未配置时行为与此前仅开发控制台日志一致。

## 实现要点

- **`internal/config`**：解析 **`NEXUSROUTER_LOG_DIR`** 至 **`LogDir`**。
- **`internal/provider`**：**`ProvideLogger`** 在 **`LogDir`** 非空时 **`Tee`** 文件与标准错误。
- **`cmd/api`**：**`applyCLILogDir`** 在 **`LoadDotEnv`** 之后设置环境变量，供 **`config.Load`** 读取。

## 兼容性

- 未设置 **`NEXUSROUTER_LOG_DIR`** 且未传 **`-log-dir`** 时，与仅 **`zap.NewDevelopment()`** 行为一致。
