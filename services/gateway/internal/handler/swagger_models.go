package handler

// ChatCompletionRequest Chat Completions 请求体（文档用子集）。
type ChatCompletionRequest struct {
	Model    string        `json:"model" example:"gpt-4o-mini"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
}

// ChatMessage 单条对话消息。
type ChatMessage struct {
	Role    string `json:"role" example:"user"`
	Content string `json:"content" example:"hello"`
}

// GatewayErrorBody 网关自产错误 JSON（与运行时 WriteGatewayError 一致）。
type GatewayErrorBody struct {
	Code       string `json:"code" example:"UNAUTHORIZED"`
	Message    string `json:"message" example:"凭证无效或缺失"`
	RequestID  string `json:"request_id" example:"a1b2c3d4e5f6g7h8"`
}

// ChatCompletionResponse 非流式成功响应（文档用子集）。
type ChatCompletionResponse struct {
	ID      string `json:"id" example:"chatcmpl-xxx"`
	Object  string `json:"object" example:"chat.completion"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
}
