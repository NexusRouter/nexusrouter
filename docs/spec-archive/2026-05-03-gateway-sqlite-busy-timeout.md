# 网关：SQLite 文件连接忙碌等待

在仅使用本地 **SQLite** 文件（未配置 **`NEXUSROUTER_DATABASE_URL`**）时，打开数据库的连接串附带 **`_busy_timeout`**（毫秒），用于在库文件短暂被锁定时等待而非立即返回 **`SQLITE_BUSY`**。可通过 **`NEXUSROUTER_SQLITE_BUSY_TIMEOUT_MS`** 覆盖；未设置或非正数时默认 **3000**；超过 **600000** 时按 **600000** 封顶。路径中若已有查询串（如内存 DSN），则以 **`&`** 追加该参数。**Postgres** 模式不使用此项。
