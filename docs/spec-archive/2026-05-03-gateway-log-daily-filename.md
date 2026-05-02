# 网关：按日期的持久化日志文件名

## 功能说明

在已配置持久化日志目录的前提下，支持通过环境变量 **`NEXUSROUTER_LOG_DAILY_FILE=true`**（或其它 Viper 认可的布尔真值）将 JSON 日志文件从固定的 **`gateway.log`** 切换为 **`gateway-YYYYMMDD.log`**（**`YYYYMMDD`** 取进程初始化日志组件时的本地日期）。未启用时行为与此前仅写入 **`gateway.log`** 一致；跨自然日不自动切换文件名（与按进程生命周期固定文件名语义一致）。

## 实现要点

- **`internal/config`**：解析 **`NEXUSROUTER_LOG_DAILY_FILE`** 至 **`LogDailyFile`**。
- **`internal/provider`**：**`ProvideLogger`** 根据 **`LogDailyFile`** 选择文件名。

## 兼容性

- 未设置或设为假时，仍为 **`gateway.log`**，不影响现有部署。
