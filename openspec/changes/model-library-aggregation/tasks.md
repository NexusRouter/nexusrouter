## 1. 数据库与领域模型

- [x] 1.1 按 `design.md` DDL 定义 GORM 实体：`model_vendor`、`model_base`、`model_upstream`、`model_instance`（字段名、索引、UNIQUE、`upstream_id` → `model_upstream.id`）
- [x] 1.2 添加迁移，**仅创建四张新表**（SQLite + Postgres），**不包含**任何旧表数据迁移
- [x] 1.3 实现 repository：CRUD 与按 `model_code`、`vendor_id`、`status` 查询

## 2. 网关运行时

- [x] 2.1 实现 `InstancePicker`：按 `priority` ASC、`is_official`、`weight`、各表 `status=1` 及可选健康选择实例
- [x] 2.2 在 **POST `/v1/chat/completions`** 链中：选中实例后向 **`model_upstream.base_url`** 转发，使用 **`api_key`**；将 body 中 **`model`** 改写为 **`provider_model_code`**
- [x] 2.3 更新 **GET `/v1/models`**：仅聚合满足启用链的逻辑模型，`id` = **`model_code`**
- [ ] 2.4 （可选）对 `model_upstream` 做探针，临时剔除不可达实例 — **未实现**（可后续迭代）

## 3. 管理 API

- [x] 3.1 注册 `/api/admin/v1/model-library/**` 下四资源 REST，JWT + RBAC 与现网一致
- [x] 3.2 提供「对指定 **`model_upstream`** 拉取 **`{base_url}/v1/models`**」同步接口，输出导入预览结构
- [x] 3.3 机读 OpenAPI：产品决策为**不暴露** HTTP OpenAPI 端点（无 `internal/openapi`、无 `/openapi.json`）

## 4. 控制台

- [x] 4.1 扩展模型库页：厂商 / 逻辑模型 / 上游 / 实例 CRUD 与筛选，字段对齐 DDL
- [x] 4.2 补充 i18n（`pages.modelLibrary.*` 或约定前缀）

## 5. 验证与文档

- [x] 5.1 后端测试：选择算法、`provider_model_code` 改写、`/v1/models` 过滤、**`api_key`** 不落日志
- [x] 5.2 更新 README：四表说明与**无迁移**部署前提
