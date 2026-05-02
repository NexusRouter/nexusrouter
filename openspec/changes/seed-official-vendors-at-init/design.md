## Context

NexusRouter 网关使用 GORM **`AutoMigrate`** 创建 **`model_vendor`** 等四表；启动链为 **`repository.OpenDB` → `AutoMigrate` → `EnsureSystemBootstrap` → `BootstrapFromConfig` → `keystore.BootstrapKeysIfEmpty`**（见 `services/gateway/internal/provider/database.go`）。当前 **`BootstrapFromConfig`** 仅处理 gateway YAML 与管理员账号，**不**预置模型厂商。参考项目 **PangaeaHub** 在 **`relay/channeltype/define.go`** 用常量枚举所有对接渠道，在 **`relay/channeltype/url.go`** 用与下标对齐的 **`ChannelBaseURLs`** 给出默认 **`https` 根地址**（用于中继路由）。本变更将其中**面向公开 API 的原厂**沉淀为 NexusRouter 的 **`model_vendor`** 行（仅元数据，不含密钥）。

## Goals / Non-Goals

**Goals:**

- 在进程启动、迁移完成后**自动**写入一组稳定的官方原厂 **`model_vendor`**（**`vendor_type = 1`**），包含 **`vendor_name`、`vendor_code`（UNIQUE）**，**`logo`** 可空或统一空字符串。
- **幂等**：重复启动、多副本先后启动不得因唯一约束失败而崩溃；升级后新版本可增加**新**原厂行而不破坏已有数据。
- **`vendor_code`** 命名与 Pangaea 渠道概念**可对照**（见下表），便于运维与文档一致。

**Non-Goals:**

- 不预置 **`model_upstream` / `model_base` / `model_instance`**（无默认密钥、无默认可路由模型）。
- 不在运行时依赖 PangaeaHub 源码或二进制；对照表仅维护在 NexusRouter 仓库内。
- 不把代理/二方转发商（如部分 **API2D** 类）纳入「官方原厂」种子，除非产品明确归类为 **`vendor_type = 2`** 并单独列表（首版可省略）。

## Decisions

### 1. 触发与幂等策略

- **采用按 `vendor_code` 的「存在则跳过」插入**（GORM **`FirstOrCreate`** 或等价 **`WHERE vendor_code = ?` 后无则 Create**），在**每次**网关启动完成迁移后执行一次轻量种子循环。
- **理由**：仅 **`COUNT(*)=0` 时插入**无法在后续版本向已有部署**追加**新原厂；按编码幂等可支持升级增项。
- **备选**：仅空表插入 — 实现更简单但升级需手工迁移或一次性 job；否决。

### 2. 厂商清单来源与收录规则

- 以 PangaeaHub **`channeltype`** 中**有明确公开 HTTPS 根地址**且代表**模型原厂或公有云模型服务**的渠道为主，映射为 NexusRouter 的 **`vendor_code`（小写 snake_case）** 与展示名 **`vendor_name`**。
- **不收录**：**`Unknown`、`Custom`、`Dummy`、`Proxy`** 及明显个人代理站；**`OpenAICompatible`** 为协议形态而非单一厂商，不单独成行。
- **聚合/平台类**（如 **OpenRouter**、国内部分聚合）若纳入，**`vendor_type` MUST 为 2**；首版可**仅收录 `vendor_type=1`** 以降低歧义，聚合类留给管理员自建或后续变更。

**对照表（实现时以代码内常量为最终来源；Base URL 仅帮助选型，不写入 `model_vendor` 表）：**

| vendor_code | vendor_name（示例） | Pangaea 常量（define.go） | 备注 |
|-------------|----------------------|---------------------------|------|
| openai | OpenAI | OpenAI | `ChannelBaseURLs[OpenAI]` |
| anthropic | Anthropic | Anthropic | |
| google_gemini | Google Gemini | Gemini | 与 Generative Language API 对齐 |
| azure_openai | Azure OpenAI | Azure | Pangaea 中 BaseURL 为空，仍作为原厂云保留 |
| baidu | Baidu 千帆 | Baidu / BaiduV2 | 二选一编码统一在实现中固定 |
| zhipu | 智谱 AI | Zhipu | |
| aliyun_dashscope | 阿里云 DashScope | Ali / AliBailian | 实现定一条主 `vendor_code` |
| moonshot | Moonshot | Moonshot | |
| baichuan | 百川智能 | Baichuan | |
| minimax | MiniMax | Minimax | |
| mistral | Mistral AI | Mistral | |
| groq | Groq | Groq | |
| deepseek | DeepSeek | DeepSeek | |
| cohere | Cohere | Cohere | |
| xai | xAI | XAI | |
| together | Together AI | TogetherAI | |
| cloudflare | Cloudflare Workers AI | Cloudflare | |
| doubao | 火山方舟 Doubao | Doubao | |
| novita | Novita | Novita | |
| replicate | Replicate | Replicate | |
| hunyuan | 腾讯混元 | Tencent | |
| siliconflow | SiliconFlow | SiliconFlow | 若视为平台则改为 type 2 或首版不收录 |

（实现阶段可删减列数以控制首版范围，但须在 **`design` 与 spec 中**说明「最小集」原则。）

### 3. 代码落点

- 新建 **`services/gateway/internal/repository/official_vendors_seed.go`**（或 **`seed_official_vendors.go`**）导出 **`SeedOfficialVendors(db *gorm.DB, log *zap.Logger) error`**。
- 在 **`ProvideDB`** 中于 **`BootstrapFromConfig`** 之后（或紧接 **`AutoMigrate`** 之后、与 **`BootstrapFromConfig`** 并列）调用，确保表已存在。
- **理由**：与现有 **`bootstrap.go`** 并列，职责清晰；避免 `keystore` 循环依赖。

### 4. 日志与错误

- 每个新插入的厂商 **DEBUG 或 INFO** 一条；冲突跳过不视为错误。
- 任一条插入失败（非唯一约束）应 **返回 error** 使启动失败，避免半初始化。

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| 与管理员手工创建的 **`vendor_code` 冲突** | 种子使用稳定编码；冲突时 **跳过**（视为已存在） |
| 多副本同时首启竞态 | 依赖 DB UNIQUE + 「先查后插」或 upsert；接受偶发唯一错误时重试或忽略重复 |
| 「官方」边界争议 | 首版 **`vendor_type=1` 仅收录公认原厂**；聚合商后续单独列表 |
| 国际化展示名 | 首版 **`vendor_name` 用英文或中英择一**，与 Dashboard i18n 解耦（仅种子数据） |

## Migration Plan

- **部署**：常规发版；无单独迁移脚本，依赖应用启动时种子。
- **回滚**：回滚二进制；已插入的 **`model_vendor`** 行保留（无害）。若需清理，由运维 SQL 按 `vendor_code` 删除（非自动化）。

## Open Questions

- **`siliconflow` / `openrouter`** 是否纳入首版及对应 **`vendor_type`**（若纳入 OpenRouter，强烈建议 **type 2**）。
- **`baidu` / `baidu_v2`** 是否在种子中合并为单一 **`vendor_code`**。
