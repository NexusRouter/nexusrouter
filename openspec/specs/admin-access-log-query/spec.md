# admin-access-log-query Specification

## Purpose
TBD - created by archiving change advanced-admin-management. Update Purpose after archive.
## Requirements
### Requirement: 管理端日志条件查询

系统 MUST 提供受管理权限保护的 **HTTP API**，支持按以下条件对代理访问日志进行**组合筛选**（AND 语义，未传条件则不参与过滤）：**时间范围**（UTC）、**请求路径**（前缀或精确匹配以实现为准）、**HTTP 状态码**（单值或范围）、**API Key 指纹**（与日志字段 **`api_key_fp`** 对齐；MUST NOT 接受完整明文 Key 作为查询输入）、**客户端 IP**。结果 MUST 支持**分页**（`limit`/`cursor` 或 `page`/`page_size`，在实现中固定一种）。

#### Scenario: 仅时间范围查询返回有序结果

- **WHEN** 管理员提交合法时间窗与分页参数且无其它筛选项
- **THEN** 响应包含按时间**逆序**排列的日志条目列表及下一页游标或总数（与实现一致），且 MUST NOT 包含完整 **`Authorization`** 原文

#### Scenario: 无日志文件或功能未启用时明确错误

- **WHEN** `proxy_access_log` 未启用或路径不可读
- **THEN** API MUST 返回可理解错误（如 **503** 或 **400**），且 MUST NOT 返回伪造记录

### Requirement: CSV 导出

管理端 MUST 提供将**当前筛选条件下**的日志结果导出为 **CSV** 的能力（同步流式下载或生成临时文件下载，由实现选定）。CSV MUST 使用 UTF-8；列集合 MUST 在 `design.md` 列出且默认**不包含**可还原用户密钥的列。

#### Scenario: 导出空结果

- **WHEN** 筛选条件命中 0 条
- **THEN** 导出文件仍 MUST 成功生成且至少包含表头行（或等价空表），且 HTTP 状态为成功类

