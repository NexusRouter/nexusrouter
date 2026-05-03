# 网关：入口命令行 `-port` 推导监听端口

## 功能说明

API 进程支持在命令行传入 **`-port <端口>`**（十进制 **`1`**–**`65535`**）。在已加载当前工作目录 **`.env`**（若存在）之后，若 **`NEXUSROUTER_HTTP_LISTEN_ADDR`** 与 **`PORT`** 在环境中仍均为空，则将 **`PORT`** 设为该值，供 **`config.Load`** 按既有规则推导 **`HTTPListenAddr`**；若上述任一环境变量已非空，则忽略 **`-port`**。非法端口导致向标准错误输出说明并以非零退出码终止，不启动 HTTP。**`-version`**、**`-help`**、**`-h`** 仍在加载 **`.env`** 之前处理。

## 实现要点

- **`cmd/api/main.go`**：先 **`parseStartupCLI`** 处理版本与帮助并退出；再 **`LoadDotEnv`**；再 **`applyCLIPort`**；再 **`InitializeApp`**。
- **`cmd/api/cli.go`**：与版本、帮助共用同一 **`flag.FlagSet`** 定义 **`-port`**。

## 兼容性

- 未使用 **`-port`** 时行为与此前一致。
- **`NEXUSROUTER_HTTP_LISTEN_ADDR`**、操作系统或 **`.env`** 提供的 **`PORT`** 始终优先于 **`-port`**。
