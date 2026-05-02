# 网关监听地址与 PORT 环境变量

## 行为

当未设置 **`NEXUSROUTER_HTTP_LISTEN_ADDR`**（或仅为空白）时，若存在非空通用环境变量 **`PORT`**，则将其规范为 Gin 可用的监听地址：已为 **`host:port`** 形式时原样使用；否则视为端口数字并加前导 **`:`**。

当 **`NEXUSROUTER_HTTP_LISTEN_ADDR`** 已非空时，**`PORT`** 不参与覆盖。

二者皆缺省时默认 **`HTTPListenAddr`** 为 **`:8080`**。
