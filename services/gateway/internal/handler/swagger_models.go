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

// ChatCompletionResponse 非流式成功响应（文档用子集）。
type ChatCompletionResponse struct {
	ID      string `json:"id" example:"chatcmpl-xxx"`
	Object  string `json:"object" example:"chat.completion"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
}
