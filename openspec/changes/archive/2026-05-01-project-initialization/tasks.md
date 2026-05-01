## 1. 后端（services/gateway）

- [x] 1.1 将 go.mod 中 Go 版本改为 1.24.x，并对齐 CI 的 setup-go
- [x] 1.2 将 Gin 固定为 v1.10.0 并 go mod tidy
- [x] 1.3 安装规范所列其余 Go 依赖（GORM、Redis、Wire、Viper、Zap、JWT、validator、migrate、swag、testify）
- [x] 1.4 创建 cmd/api 与 internal 分层目录及占位包
- [x] 1.5 实现 cmd/api 入口：Zap + Gin :8080，移除根目录 main.go
- [x] 1.6 接入 Wire：wire.go / wire_gen.go 与 main 调用
- [x] 1.7 添加 services/gateway README：air 安装与本地运行说明
- [x] 1.8 统一 JSON 错误响应与 Zap 记录（骨架中间件）

## 2. 前端（web/dashboard）

- [x] 2.1 使用 pnpm 在 web/dashboard 初始化 Vite React TS 工程
- [x] 2.2 安装规范版本矩阵依赖（antd、Tailwind v4、React Query、Zustand 等）
- [x] 2.3 配置 Tailwind v4 与 @tailwindcss/vite
- [x] 2.4 创建 src/components、pages、stores、services、utils 目录
- [x] 2.5 配置 ESLint 9 与 Prettier 及 package.json 脚本

## 3. 仓库与 CI

- [x] 3.1 更新 openspec/project.md 与 README 本地开发说明
- [x] 3.2 更新 .github/workflows/ci.yml（Node 22、pnpm、gateway Go 1.24、dashboard 构建/测试）
- [x] 3.3 按需更新 .gitignore
- [x] 3.4 运行 openspec validate --all --no-interactive

## 4. 收尾

- [x] 4.1 在 services/gateway 与 web/dashboard 分别跑通 go test / pnpm build / pnpm lint
- [x] 4.2 将本 tasks.md 中已完成项勾选为 [x]
