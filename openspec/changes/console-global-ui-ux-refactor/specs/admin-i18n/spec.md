## ADDED Requirements

### Requirement: 控制台核心术语表

系统 MUST 在仪表盘 i18n 资源中提供**控制台核心术语**的键集合（建议使用统一前缀如 `consoleTerms.*` 或 `glossary.*`），覆盖侧栏域标题、侧栏项、以及与用户任务强相关的跨页概念（如网关策略、上游、API Key、访问日志等）。中文与英文 MUST 表达同一含义；MUST NOT 在同一界面混用未定义的英文缩写作为主标签而无首次解释。

#### Scenario: 双语对照一致

- **WHEN** 用户在中文与英文之间切换
- **THEN** 术语表所覆盖的文案同步切换且语义一致

#### Scenario: 侧栏引用术语键

- **WHEN** 审查者检查侧栏域标题与菜单项文案来源
- **THEN** 其用户可见字符串来自 i18n 键而非 JSX 硬编码终态文案（开发期占位除外）

### Requirement: 与全局 UX 规格对齐

`admin-i18n` 的术语与文案组织 MUST 支持 `admin-console-global-ux` 中「术语与主副文案层级」及「空态、加载与错误基线」对 i18n 的要求；新增键 MUST 同时提供 `zh` 与 `en`。

#### Scenario: 空态与错误可走 i18n

- **WHEN** 某页展示全局基线要求的空态或错误提示
- **THEN** 该提示文案可通过 i18n 解析为中英之一
