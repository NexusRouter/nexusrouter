## 1. 信息架构与组件拆分

- [x] 1.1 在 `web/dashboard/src/pages/SystemSettings.tsx`（或拆分为 `components/system-settings/*`）实现三大模块：运行状态 Card、代理日志 Card、高级配置 Collapse
- [x] 1.2 将 `settings` 数组映射为「业务标题 + 值 + 生效方式 + 说明」行组件，避免以 `key` 作为主标题
- [x] 1.3 对齐现有布局宽度、暗色模式与 Ant Design 用法，与网关策略等页视觉一致

## 2. 国际化与术语表落地

- [x] 2.1 在 `web/dashboard/src/locales/zh.ts` 与 `en.ts` 增加/调整 `settings.*` 键：模块标题、各配置项业务名、生效方式长说明、`max_size_mb`/`max_backups` 与 GET 列表不对齐时的提示句
- [x] 2.2 确保中英文对同一配置项描述同一修改路径，无混用「热更新」与英文 slug

## 3. 表单与数据联动

- [x] 3.1 保持 `PUT` body 形状不变；确认保存成功后 `invalidateQueries(['system-settings'])` 已触发且运行状态区展示更新
- [x] 3.2 为日志表单区补充与运行状态的文案联动说明（副标题或 Alert），并保留 Operator 禁用与 `persist` 开关说明

## 4. 风险提示与高级区

- [x] 4.1 在折叠「高级配置」中展示键名与环境变量提示（来自现有 `hint` / `key`），支持排障场景
- [x] 4.2 写回 `gateway.yaml` 旁保留或增强一行风险/作用说明；服务端错误 message 原样或可映射展示

## 5. 验证

- [x] 5.1 本地执行 `pnpm exec tsc --noEmit`（在 `web/dashboard`）确保无类型错误
- [x] 5.2 手测：Admin 保存日志、Operator 只读、中英文切换、保存后只读区与表单一致
