# 网关：入口命令行 `-version`

## 功能说明

API 进程在加载 **`.env`**、初始化依赖与监听端口之前，若命令行包含 **`-version`**，则向标准输出打印一行构建版本标识（与 **`internal/buildinfo.Version`**、**`GET /health`** 的 **`version`** 字段一致），并以退出码 **0** 结束，不启动 HTTP 服务。

解析失败（例如无法识别的标志）时向标准错误输出说明并以非零码退出，与 Go **`flag`** 包惯例一致。

## 实现要点

- 在 **`cmd/api/main.go`** 中于 **`config.LoadDotEnv()`** 之前调用 **`exitIfVersionRequested`**；若需退出则 **`os.Exit(0)`** 或 **`os.Exit(2)`**。
- 版本字符串来源为 **`internal/buildinfo.Version`**（可由 **`-ldflags -X ...`** 注入）。

## 兼容性

- 未传递 **`-version`** 时启动流程与此前一致。
