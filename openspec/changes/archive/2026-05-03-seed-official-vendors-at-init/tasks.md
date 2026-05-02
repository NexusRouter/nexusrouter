## 1. 种子数据与仓储

- [x] 1.1 在 `services/gateway/internal/repository/` 新增官方厂商常量表（`vendor_code`、`vendor_name`、`vendor_type`，默认 `status=1`），并与 `design.md` 中 PangaeaHub 对照表对齐、定稿首版收录范围（含百度/阿里单编码等待定项）
- [x] 1.2 实现 `SeedOfficialVendors(db *gorm.DB, log *zap.Logger) error`：按 `vendor_code` 幂等插入（`FirstOrCreate` 或先查后插），已存在则跳过且不更新字段
- [x] 1.3 在 `ProvideDB`（或等价启动链）于迁移与 `BootstrapFromConfig` 之后调用种子函数；非唯一约束错误时返回 error 使启动失败

## 2. 验证与文档

- [x] 2.1 添加仓库内测试：内存 SQLite 迁移后调用种子两次，第二次无重复行且行数与首版常量数一致；可选断言若干 `vendor_code` 存在
- [x] 2.2 将本变更的 `specs/official-vendor-seed/spec.md` 在归档流程中合并入 `openspec/specs/`（随 `/opsx:apply` 或发布前完成）；更新 `services/gateway/README.md` 中「模型库」相关节简述启动预置厂商（若已有该节）
