## Context

网关已具备：`gateway.yaml` 中的 **`rate_limit`**（`rps_per_ip`、`rps_per_key`）、**`cors`** 段、**`proxy_access_log`**（JSON 行，字段含 `request_id`、`path`、`status`、`client_ip`、`api_key_fp` 等，见 `internal/accesslog`）。管理控制台（`/api/admin/v1`）与仪表盘已存在。进阶功能需在**不破坏** OpenAI 兼容路径与既有单值限流语义的前提下，扩展**可检索日志**、**多规则限流**、**CORS 批量编辑**与 **IP 名单**。

## Goals / Non-Goals

**Goals:**

- 提供管理端**日志查询** API 与 UI：多条件筛选、分页、**CSV 导出**。
- 提供**多条限流规则**的配置与生效顺序；支持按 **IP** / **API Key（指纹）** 维度及阈值、路径范围。
- 提供 **CORS** 的可视化编辑：域名/方法/头、**批量域名**导入、写回配置并热生效。
- 提供 **IP 黑/白名单**：类型区分、批量增删、与请求链整合、实时生效。

**Non-Goals：**

- 完整 ELK/Splunk 级日志平台；本阶段以**单机可部署**、低外部依赖为优先。
- 替换现有 `POST /internal/*` 运维路径；与之并存。
- 多区域全局速率同步（需分布式存储）——除非在设计评审后明确纳入。

## Decisions

1. **日志数据源**  
   - **选定**：以 **`proxy_access_log`** 输出路径下的 **JSON 行文件**为主数据源；管理 API **反向时间顺序**扫描与过滤（可配置最大扫描字节/行数以保护延迟）。  
   - **备选**：旁路写入 SQLite——作为第二阶段若扫描性能不足再引入。

2. **「响应」日志范围**  
   - **选定**：首版查询字段与现有 **`proxy_access`** 行一致；若需响应体片段，**不**纳入默认导出（合规与体积）；在 spec 中声明「请求/响应」为**元数据级**（状态、耗时、错误标记），响应体另议。

3. **多规则限流模型**  
   - **选定**：在 `gateway.yaml`（或等价运行时快照）中新增 **`rate_limit_rules`** 数组（`id`、`dimension`：`ip`|`api_key_fp`、`match_path_prefix`、`rps`、`burst`、`enabled`、`priority`）；中间件按 **priority 降序** 匹配首条命中；未命中则回退现有全局 `rps_per_ip` / `rps_per_key`。  
   - **备选**：仅 UI 编辑仍写回两个全局标量——拒绝，不满足「多条规则」。

4. **CORS 批量域名**  
   - **选定**：API 接受 **`allow_origins`** 字符串数组或换行/逗号分隔的 **bulk 文本** 一次提交；服务端拆分去重后合并入快照并 **Persist + Reload**（与现有 `core-admin-management` 写盘模式一致）。

5. **IP 名单语义**  
   - **选定**：快照中 **`ip_access`**：`mode`（`off`|`allowlist`|`denylist`）、`cidrs: []string`（支持单 IP 与 **CIDR** 字符串）；**白名单模式**：仅允许列表内 IP 访问受保护路由（至少 `POST /v1/chat/completions`）；**黑名单模式**：列表内拒绝；**off**：不启用。黑名单与白名单**互斥**，`mode` 单选。  
   - **Mitigation**：默认 `off`；误锁时运维可通过本地 **`reload-config`** 或控制台（需已登录）切回。

6. **实时生效**  
   - **选定**：与现网一致——**写配置 + `Runtime.Reload()`** 或等价内存替换；名单与 CORS 与限流规则同一快照版本号（可选 `etag` 返回给前端做乐观锁）。

## Risks / Trade-offs

- **[Risk] 大文件扫描拖慢管理 API** → **Mitigation**：硬上限（时间窗 + 最大行数）、异步导出任务（二期）或索引。  
- **[Risk] 名单与 CIDR 解析错误** → **Mitigation**：服务端校验、拒绝无效项并返回逐条错误。  
- **[Risk] CSV 导出含可关联指纹** → **Mitigation**：导出列不包含完整 API Key；文档声明留存周期。

## Migration Plan

1. 扩展 `gateway.yaml` schema（向后兼容：无新段时行为与今一致）。  
2. 发布后端 → 发布前端；旧配置无需迁移即可启动。  
3. 回滚：移除新 YAML 段并 reload；前端隐藏进阶菜单可由 feature flag 控制（可选）。

## Open Questions

- 日志查询是否必须支持**多文件**（rotated logs）联合查询——建议首版支持 `MaxBackups` 范围内文件名模式。  
- 限流规则是否需 **per-route** 除 `path_prefix` 外再支持 **method**——首版可仅 `path_prefix` 空表示全局。
