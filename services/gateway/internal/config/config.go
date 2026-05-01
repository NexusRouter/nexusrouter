// Package config 负责网关运行时配置（环境变量）。
package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 网关与 Chat 代理、文档 UI 相关配置。
type Config struct {
	// UpstreamBaseURL 上游 OpenAI 兼容服务基址，须含 scheme 与 host；为空时 POST /v1/chat/completions 返回 503。
	UpstreamBaseURL string
	// UpstreamTimeout 等待上游响应头的上限（非流式整响应仍受 Transport 约束）。
	UpstreamTimeout time.Duration
	// GatewayAPIKeys 允许的网关 API 密钥列表；请求须携带 Bearer 与该值之一一致，或 X-API-Key 头与之之一一致。为空则任何请求均 401（除健康检查等）。
	GatewayAPIKeys []string
	// UpstreamAPIKey 发往上游的 Bearer token（当未开启 ForwardClientAuthorization 时注入）；可为空。
	UpstreamAPIKey string
	// ForwardClientAuthorization 为 true 时将客户端 Authorization 原样转发上游。
	ForwardClientAuthorization bool
	// EnableSwaggerUI 是否暴露 /swagger/* 与相关静态资源。
	EnableSwaggerUI bool
}

// Load 通过 Viper 读取环境变量（键名与 README 中 NEXUSROUTER_* 一致）。
func Load() *Config {
	v := viper.New()
	v.AutomaticEnv()

	timeout := 120 * time.Second
	if v.IsSet("NEXUSROUTER_UPSTREAM_TIMEOUT") {
		if d, err := time.ParseDuration(v.GetString("NEXUSROUTER_UPSTREAM_TIMEOUT")); err == nil {
			timeout = d
		}
	}
	fwd := false
	if v.IsSet("NEXUSROUTER_FORWARD_CLIENT_AUTHORIZATION") {
		fwd = v.GetBool("NEXUSROUTER_FORWARD_CLIENT_AUTHORIZATION")
	}
	swagger := true
	if v.IsSet("NEXUSROUTER_ENABLE_SWAGGER_UI") {
		swagger = v.GetBool("NEXUSROUTER_ENABLE_SWAGGER_UI")
	}

	return &Config{
		UpstreamBaseURL:            strings.TrimSpace(v.GetString("NEXUSROUTER_UPSTREAM_BASE_URL")),
		UpstreamTimeout:            timeout,
		GatewayAPIKeys:             splitKeys(v.GetString("NEXUSROUTER_GATEWAY_API_KEYS")),
		UpstreamAPIKey:             strings.TrimSpace(v.GetString("NEXUSROUTER_UPSTREAM_API_KEY")),
		ForwardClientAuthorization: fwd,
		EnableSwaggerUI:            swagger,
	}
}

func splitKeys(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
