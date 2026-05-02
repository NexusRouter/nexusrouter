# NexusRouter 管理控制台（`web/dashboard`）

基于 **React + TypeScript + Vite** 的 Web 管理界面，通过网关 HTTP API 管理 NexusRouter（上游、密钥、限流、CORS、IP 名单、代理访问日志、模型库等）。网关侧说明见同仓库 [`services/gateway/README.md`](../services/gateway/README.md)。

## 环境要求

- **Node.js ≥ 22**（与 `package.json` 中 `engines` 一致）
- **pnpm**（推荐与仓库一致的 9.x，见 `packageManager` 字段）

## 安装依赖

```bash
cd web/dashboard
pnpm install
```

## 本地开发

1. **先启动网关**（默认监听 `:8080`），例如：

   ```bash
   cd services/gateway
   go run ./cmd/api
   ```

   监听地址、数据库、管理员等环境变量说明见 [Gateway README § 环境要求 / 环境变量一览](../services/gateway/README.md)。

2. **启动前端开发服务器**：

   ```bash
   cd web/dashboard
   pnpm dev
   ```

   默认由 Vite 提供本地开发地址（一般为 `http://127.0.0.1:5173`，以终端输出为准）。

### 前端如何访问网关 API

`src/services/api.ts` 中 Axios 的 `baseURL` 由环境变量 **`VITE_API_BASE_URL`** 决定：

| 配置方式 | 行为 |
| -------- | ---- |
| **未设置**（推荐本地快速试） | 默认 **`http://127.0.0.1:8080`**，浏览器**直连**网关，可能受网关 CORS 策略影响。 |
| **在 `web/dashboard/.env` 中设为空** `VITE_API_BASE_URL=` | `baseURL` 为空串，请求走**相对路径**，由 Vite 开发服务器**代理**到网关，便于避免跨域。 |

`vite.config.ts` 中已将下列路径代理到 **`http://127.0.0.1:8080`**：

- `/api`、`/health`

若网关未跑在本机 8080，请同时修改 **`vite.config.ts`** 里 `server.proxy` 的 `target`，或改用直连并在网关中配置好 **CORS**（参见 Gateway README「中间件与限流顺序」与 CORS 管理 API）。

### 使用流程说明

- **首次部署 / 空库**：网关可能处于 **Bootstrap** 阶段，未完成向导时多数页面会返回 `403`（`BOOTSTRAP_REQUIRED`）。控制台会引导进入 **初始化向导**（`/setup`），完成后才能正常登录管理功能。
- **已初始化**：使用管理员（或操作员）账号在 **登录页**登录；令牌保存在浏览器 **localStorage**（键名由应用维护），后续请求通过 Axios 拦截器附加 **`Authorization: Bearer`**。
- **操作员（operator）**：部分写操作会被拒绝（403），与网关策略一致。

具体白名单路径、Bootstrap API、管理 API 列表见 [Gateway README § 首次初始化 / 管理控制台 API](../services/gateway/README.md)。

## 操作手册

以下说明面向**在浏览器中使用本控制台**的运维与管理员，按页面与常用操作组织。路径以控制台内路由为准（开发环境通常为 `http://127.0.0.1:5173` 加下列路径）。

### 全局布局与通用操作

- **侧栏导航（桌面端）**：左侧为功能入口；当前页高亮。点击顶部品牌区域可回到 **概览**（`/dashboard`）。
- **移动设备**：顶部左侧 **菜单** 按钮打开抽屉，内含与侧栏相同的导航；选中一项后抽屉自动关闭。
- **顶栏**：右侧依次为 **深色/浅色主题**、**语言（中/英）**、**退出登录**。退出后需重新登录；令牌保存在浏览器本地（`localStorage`）。
- **告警条**：若网关上报非「正常」状态，内容区顶部可能出现**红色（严重）或黄色（警告）**横幅，提示需要关注的运行态原因；约每 30 秒刷新一次。

### 首次初始化（`/setup`）

网关数据库为空或未完成引导时，访问任意受保护页会被重定向到 **初始化向导**。

1. 填写 **管理员用户名**、**密码**（至少 8 位）。
2. 可选填写 **站点显示名**。
3. 提交成功后跳转到 **登录页**；之后按正常登录流程使用。

若网关已初始化，直接访问 `/setup` 会被重定向到登录页。

### 登录（`/login`）

1. 输入用户名、密码。
2. 可选勾选 **记住用户名**，下次打开登录页会预填用户名（不保存密码）。
3. 登录成功后进入上次要访问的页面，或默认进入 **概览**。

### 角色与权限

- **管理员（admin）**：可执行全部写操作（保存配置、创建密钥等）。
- **操作员（operator）**：多为只读；写操作（保存、编辑、删除等）在界面上隐藏或会被接口拒绝（403），与网关策略一致。

下文凡提到「保存」「编辑」等，操作员账号可能不可用。

### 概览（`/dashboard`）

- 展示网关 **在线状态**、**当前 RPS 估算**、**成功率**、**平均延迟**，以及 **今日/昨日请求量**、**今日按错误码统计**。
- 数据约每 **15 秒** 自动刷新一次；也可刷新浏览器页面强制拉取最新数据。

### 上游（`/upstreams`）

用于查看与调整 **上游列表**、**路由策略**及 **当前固定上游**。

- **只读信息**：表格展示各上游 `id`、`base_url`、`weight`；标签区展示策略（如 `round_robin`）、默认上游、**当前固定（pin）的上游**、是否已配置 `gateway.yaml` 等。
- **编辑并写入**（管理员）：打开弹窗，可增删改上游行（`id` / `base_url` / `weight`），并编辑 `strategy`、`default_upstream_id`、`active_upstream_id`（可空）。提交后 **持久化** 到网关配置。
- **取消固定上游**：确认对话框后清空「当前固定」并持久化，恢复按策略调度。

持久化行为与环境变量 `NEXUSROUTER_GATEWAY_CONFIG_FILE` 等见页内说明与 Gateway README。

### 模型库（`/model-library`）

四张表：**厂商**、**逻辑模型**（`model_code`）、**上游**（`base_url` + `api_key`）、**实例**（逻辑模型 + 厂商 + 上游行，`provider_model_code` / 优先级 / 权重 / 是否官方）。页内按 Tab 管理；厂商可填 **Logo URL**，常见 `vendor_code` 会显示默认图标（Simple Icons 风格）。

- **从上游同步**：选择 **`model_upstream` 行 id**，可选 Bearer，调用该行的 `{base_url}/v1/models`，仅拉取模型 id 列表供参考。

### API 密钥（`/api-keys`）

- **创建**：生成新密钥；**完整密钥仅创建时通过弹窗展示一次**，请立即复制保存。
- **启用/禁用**：表格中开关切换；禁用后该密钥不可用。
- **删除单条**：操作列删除（需确认）。
- **批量禁用 / 批量删除**：勾选多行后使用对应按钮（需确认）。

密钥文件与环境变量 `NEXUSROUTER_GATEWAY_KEYS_FILE` 见页内说明。

### 代理访问日志（`/logs`）

查询网关记录的代理访问日志（需网关在 **系统设置** 中启用 `proxy_access_log` 且路径可写，否则可能无数据或页内有提示）。

1. 可选填 **时间范围**、**路径前缀**、**HTTP 状态码范围**、**API Key 指纹**、**客户端 IP**、**单页条数**（上限以网关为准）。
2. 点击 **查询** 加载第一页结果。
3. **下一页**：使用游标分页，基于上次结果的 `next_cursor` 继续加载。
4. **导出 CSV**：按**当前筛选条件**导出（先执行过一次查询以确定条件）；文件名为 `access_logs.csv`。

若结果提示 **扫描截断** 等，说明本次查询未完整扫描，可缩小时间范围或联系运维调整网关侧限制。

### 网关策略（`/gateway`）

侧栏进入 **网关策略** 后，顶部 **Tab** 切换三类配置；访问 `/gateway` 会默认进入 **限流**（`/gateway/rate-limits`）。

**限流（`/gateway/rate-limits`）**

- 页面展示 **全局** 每 IP / 每 Key 的 RPS 提示（只读参考）。
- 在表单表格中维护多条规则：**优先级**、**路径前缀**、**维度**（如 `ip`、`api_key` 等，以网关支持为准）、**RPS**、**Burst**、**是否启用**。
- **添加规则** / **删除行** 后，点击 **保存并落盘**（或等价文案）一次性提交并持久化。

**CORS（`/gateway/cors`）**

- 开关 **启用 CORS**，配置 **允许的 Origin**（支持标签输入与 **批量文本** 多行/逗号分隔合并）、**允许的方法与请求头**、**预检缓存秒数**。
- **持久化** 开关控制是否写入配置文件。
- 保存后生效策略以网关为准。

**IP 访问控制（`/gateway/ip-access`）**

- **模式**：关闭 / 仅允许名单 / 仅拒绝名单。
- **CIDR 列表**：用标签或 **批量添加** 文本框录入；与 **持久化到 YAML** 选项一起 **完整保存**。
- **增量 PATCH**：第二块表单可单独 **批量增加**、**批量移除** CIDR，不必重填整张表；同样有 **持久化** 选项。

### 系统设置（`/settings`）

- **上方只读列表**：展示网关聚合的系统项（键、值、可变性与提示），便于核对运行配置。
- **代理访问日志** 表单（管理员可写）：`enabled`、`path`、`level`（如 `info`/`error`）、滚动 **单文件大小 MB**、**保留备份个数**，以及是否 **持久化**。保存后日志是否落盘、路径是否有效需结合网关部署环境确认。

操作员仅可查看只读列表，不能保存下方表单。

### 常见问题提示

| 现象 | 建议 |
| ---- | ---- |
| 页面大量 403 / `BOOTSTRAP_REQUIRED` | 先完成 `/setup` 初始化。 |
| API 请求跨域失败 | 开发环境可将 `VITE_API_BASE_URL` 置空走 Vite 代理，或在网关配置 CORS（见上文「前端如何访问网关 API」）。 |
| 登录后立刻退出或 401 | 检查网关时间、JWT 配置与浏览器是否禁用本地存储。 |
| 操作员无法保存 | 预期行为；换管理员账号或调整网关角色。 |

## 构建与预览

```bash
pnpm build    # tsc -b && vite build，产物在 dist/
pnpm preview  # 本地预览生产构建（默认仍可通过环境变量指向网关）
```

## 测试与代码质量

```bash
pnpm test            # Vitest 单测
pnpm test:coverage   # 带覆盖率
pnpm lint            # ESLint
pnpm format          # Prettier 格式化
```

## 生产部署

- 将 **`dist/`** 静态资源与**网关同域**部署（由网关或反向统一托管），**或**
- 构建前设置 **`VITE_API_BASE_URL`** 为网关对外根 URL（例如 `https://router.example.com`），使浏览器请求指向正确后端。

**安全建议**：管理控制台与 JWT 仅应在 **HTTPS**、可信网络或 VPN 后暴露；勿将真实密钥写入前端仓库。生产可将网关的 `NEXUSROUTER_ENABLE_SWAGGER_UI` 设为 `false` 等，详见 Gateway README。

---

## 附录：Vite 模板与 ESLint

以下为创建项目时自带的 Vite 说明，便于贡献者扩展工具链。

本模板提供在 Vite 中运行 React 的最小配置，包含 HMR 与部分 ESLint 规则。

目前官方提供两个插件：

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) 使用 [Oxc](https://oxc.rs)
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) 使用 [SWC](https://swc.rs/)

### React Compiler

本模板默认未启用 React Compiler，因其会影响开发与构建性能。如需接入，请参阅[官方文档](https://react.dev/learn/react-compiler/installation)。

### 扩展 ESLint 配置

若在开发生产级应用，建议更新配置以启用「类型感知」的 lint 规则：

```js
export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // 其他配置…

      // 移除 tseslint.configs.recommended，并替换为下方之一
      tseslint.configs.recommendedTypeChecked,
      // 或使用更严格的规则集
      tseslint.configs.strictTypeChecked,
      // 可选：增加风格类规则
      tseslint.configs.stylisticTypeChecked,

      // 其他配置…
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // 其他选项…
    },
  },
])
```

也可安装 [eslint-plugin-react-x](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-x) 与 [eslint-plugin-react-dom](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-dom)，以启用面向 React 的 lint 规则：

```js
// eslint.config.js
import reactX from 'eslint-plugin-react-x'
import reactDom from 'eslint-plugin-react-dom'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // 其他配置…
      // 启用面向 React 的规则
      reactX.configs['recommended-typescript'],
      // 启用面向 React DOM 的规则
      reactDom.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // 其他选项…
    },
  },
])
```
