package handler

// HealthCheckDoc 仅用于 swag 生成 OpenAPI 注释（运行时由 Health 处理）。
//
//	@Summary		健康检查
//	@Description	返回服务状态、版本号与当前时间（UTC RFC3339Nano），无需鉴权。
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/health [get]
func HealthCheckDoc() {
}

// HealthResponse 为 GET /health 的 JSON 形状（swag 模型）。
type HealthResponse struct {
	Status     string `json:"status" example:"ok"`
	Version    string `json:"version" example:"dev"`
	ServerTime string `json:"server_time" example:"2026-05-01T12:00:00.000000001Z"`
}
