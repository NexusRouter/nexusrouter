# official-vendor-seed Specification

## Purpose

定义网关在启动阶段向 **`model_vendor`** 表**幂等预置**已知官方原厂（及可选的第三方聚合商策略）的行为，与 **`model-library`** / **`model-library-aggregation`** 中 **`model_vendor`** 字段定义一致；不定义上游与实例。

## ADDED Requirements

### Requirement: 启动时预置官方厂商

网关 SHALL 在完成 **`model_vendor` 表迁移后的启动路径中**调用官方厂商种子逻辑，向 **`model_vendor`** 写入一组**产品维护的**记录，每条至少包含 **`vendor_name`、`vendor_type`、`vendor_code`、`status`**（**`status` SHALL 为启用态 1**），并符合现有 **`vendor_code` 唯一**约束。

#### Scenario: 首次部署存在缺失编码

- **WHEN** 数据库中不存在某条预置的 **`vendor_code`**
- **THEN** 启动过程 SHALL 插入该条记录且进程不因该插入失败而忽略唯一约束以外的错误

#### Scenario: 已存在相同 vendor_code

- **WHEN** 数据库中已存在与预置相同的 **`vendor_code`**
- **THEN** 启动过程 SHALL NOT 覆盖或删除该行的管理员可写字段（如 **`vendor_name`、`logo`、`status`**）；SHALL 跳过插入或等效于 no-op

### Requirement: 官方类型与清单可追溯

预置记录中代表「原厂/官方 API 提供方」的行 **`vendor_type` SHALL 为 1**（与现有模型一致：**1=官方，2=第三方**）。产品 SHALL 在实现代码或文档中维护与 **PangaeaHub `relay/channeltype`** 的**对照说明**（常量名或渠道语义），以便评审与扩展；**MUST NOT** 在运行时依赖 PangaeaHub 仓库。

#### Scenario: 类型一致

- **WHEN** 审查者检查种子数据定义
- **THEN** 每条官方原厂记录的 **`vendor_type`** 为 **1**，且 **`vendor_code`** 在表内唯一

### Requirement: 失败语义

若官方厂商种子在**非重复键**原因下失败（例如数据库不可用、约束违反非预期），网关 SHALL 使启动失败并返回/记录明确错误，**MUST NOT** 在部分厂商写入成功后静默吞掉剩余失败而导致不一致状态无日志。

#### Scenario: 持久化错误

- **WHEN** 插入过程中发生非唯一约束的数据库错误
- **THEN** 启动流程 SHALL 失败且错误可被运维从日志或进程退出码感知
