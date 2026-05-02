## Why

全新部署时 **`model_vendor` 为空**，管理员在模型库向导或厂商管理里需要手工逐个添加常见原厂（OpenAI、Anthropic、Google 等），与同类网关（如 PangaeaHub 在 **`relay/channeltype`** 中枚举的渠道类型及默认 Base URL）相比，首启体验差且易与社区惯用的 **`vendor_code` 命名不一致**。在数据库迁移完成后的启动链路中**幂等预置**已知官方厂商行，可让 UI 下拉与后续上游配置直接对齐。

## What Changes

- 在网关 **DB 就绪且表已迁移** 后增加一步：**当 `model_vendor` 中尚无任何行（或按约定仅对缺失的 `vendor_code` 补全）时**，插入一组**固定的官方原厂**（`vendor_type = 1`）记录，包含稳定的 **`vendor_code`**、展示用 **`vendor_name`**，可选 **`logo` 占位**。
- 厂商清单与命名**参考**本地仓库 `~/VisualStudioCodeProjects/PangaeaHub` 的 **`relay/channeltype/define.go`**（渠道枚举）与 **`relay/channeltype/url.go`**（默认 Base URL 对应关系），在 NexusRouter 侧**择优收录**面向公开 API 的原厂（排除明显代理/占位渠道），并在 **`design.md`** 中给出明确表列与代码常量位置。
- **非 BREAKING**：不删除管理员已有数据；预置为**插入缺失**或**仅空表时批量插入**（具体策略见设计），不自动创建 **`model_upstream` / `model_instance`**（仍由向导或管理员配置密钥与端点）。

## Capabilities

### New Capabilities

- `official-vendor-seed`: 定义启动时预置官方厂商的**触发条件**、**幂等规则**、**厂商清单字段约束**（与现有 **`model_vendor`** DDL 一致），以及与 PangaeaHub 渠道枚举的**对照说明**（仅文档化映射，不引入对该仓库的运行时依赖）。

### Modified Capabilities

- （无）现有 **`model-library`** / **`model-library-aggregation`** 表结构与 CRUD 语义不变；本变更仅增加**数据初始化**行为，可通过新增能力规范描述，无需改写既有需求正文。

## Impact

- **后端**：`services/gateway/internal/repository`（或 `provider` 启动链）新增种子函数并在 **`ProvideDB` / `BootstrapFromConfig`** 附近按顺序调用；可选单元测试（内存 SQLite）验证幂等。
- **前端**：无强制变更；模型库向导/厂商列表将**天然出现**预置厂商（若当前从 API 拉取全量厂商）。
- **依赖**：不增加 Go 模块依赖；PangaeaHub 仅作**人工对照**来源。
