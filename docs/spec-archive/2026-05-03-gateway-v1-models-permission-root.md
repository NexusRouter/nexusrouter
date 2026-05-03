# 网关：`/v1/models` 与单条检索的 permission、root 字段

## 功能说明

**`GET /v1/models`** 与 **`GET /v1/models/:model`** 成功响应中的每个模型对象在既有 **`id`、`object`、`created`、`owned_by`** 之外，补充 **`permission`**（非空 JSON 数组，元素类型为 **`model_permission`**，布尔与组织等字段与常见 OpenAI 兼容列表形状一致）及 **`root`**（字符串，与该项 **`id`** 相同）。无父级模型时不输出 **`parent`** 字段。

## 实现要点

- 构造逻辑集中在 **`handler.newOpenAIModelItem`**，列表聚合路径与目录路径、单条检索路径均复用。

## 兼容性

- 仅扩展 JSON 字段；未依赖新字段的客户端行为不变。
