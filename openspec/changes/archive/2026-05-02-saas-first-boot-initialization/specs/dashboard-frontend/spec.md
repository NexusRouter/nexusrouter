## ADDED Requirements

### Requirement: 初始化向导路由

`web/dashboard` MUST 提供独立路由（如 **`/setup`**）承载**首次初始化向导**页面，页面文案与校验 MUST 使用**中文**为主（与现有规范一致）。该路由在未初始化阶段 MUST 可匿名访问。

#### Scenario: 未初始化可打开向导

- **WHEN** 全局未初始化且用户直接访问向导路径
- **THEN** 页面渲染成功且无无限重定向

### Requirement: 全局路由守卫与初始化后禁止访问向导

应用根布局或路由守卫 MUST 在启动或导航时拉取后端初始化状态；当 **`initialized=false`** 时，除向导页与状态轮询所需资源外，访问其他业务路由 MUST **重定向**至向导页。当 **`initialized=true`** 时，访问向导路径 MUST **重定向**至登录页（或已存在之认证入口路径）。

#### Scenario: 未初始化强制进入向导

- **WHEN** **`initialized=false`** 且用户访问非向导内部路径
- **THEN** 浏览器地址栏最终落在向导路径

#### Scenario: 已初始化禁止向导

- **WHEN** **`initialized=true`** 且用户访问向导路径
- **THEN** 浏览器被重定向至登录页
