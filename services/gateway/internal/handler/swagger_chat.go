package handler

// CreateChatCompletionDoc 仅用于 swag 生成 OpenAPI/Swagger 注释（运行时由 ChatProxy 处理）。
//
//	@Summary		Chat Completions
//	@Description	与 OpenAI Chat Completions 兼容子集；官方概览 https://developers.openai.com/api/reference/overview
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Produce		text/event-stream
//	@Param			X-API-Key		header		string					false	"可选，与 Bearer 二选一"
//	@Param			body			body		ChatCompletionRequest	true	"请求体"
//	@Success		200				{object}	ChatCompletionResponse
//	@Failure		401				{object}	GatewayErrorBody
//	@Failure		502				{object}	GatewayErrorBody
//	@Security		BearerAuth
//	@Router			/v1/chat/completions [post]
func CreateChatCompletionDoc() {
}
