// Package config 负责网关运行时配置（环境变量）。
package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 网关与 Chat 代理、文档 UI 相关配置。
type Config struct {
	// UpstreamBaseURL 单上游基址（遗留）；当 UpstreamBaseURLs 为空时作为回退。
	UpstreamBaseURL string
	// UpstreamBaseURLs 多上游基址列表（逗号分隔环境变量）；非空时优先于 UpstreamBaseURL。
	UpstreamBaseURLs []string
	// UpstreamTimeout 等待上游响应头的上限（非流式整响应仍受 Transport 约束）。
	UpstreamTimeout time.Duration
	// GatewayAPIKeys 遗留：逗号分隔明文密钥；当 GatewayKeysFile 为空时用于构造密钥库。
	GatewayAPIKeys []string
	// GatewayKeysFile JSON 密钥文件路径；非空时优先于 GatewayAPIKeys。
	GatewayKeysFile string
	// AdminReloadToken 非空时启用 POST /internal/reload-keys，并以 Bearer 该值作为管理凭据。
	AdminReloadToken string
	// UpstreamAPIKey 发往上游的 Bearer token（当未开启 ForwardClientAuthorization 时注入）；可为空。
	UpstreamAPIKey string
	// ForwardClientAuthorization 为 true 时将客户端 Authorization 原样转发上游。
	ForwardClientAuthorization bool
	// EnableSwaggerUI 是否暴露 /swagger/* 与相关静态资源。
	EnableSwaggerUI bool
	// GatewayConfigFile 可选 gateway.yaml 路径；非空时与 env 合并并由运行时热加载。
	GatewayConfigFile string
	// HTTPListenAddr Gin 监听地址，如 `:8080`；默认 `:8080`。
	HTTPListenAddr string

	// --- 管理控制台（可选）---
	// EnableAdminConsole 为 true 时注册 /api/admin/*（仍需 JWT 密钥与账号配置完整）。
	EnableAdminConsole bool
	// AdminJWTSecret 管理端 JWT HMAC 密钥；非空且启用控制台时用于签发/校验。
	AdminJWTSecret string
	// AdminJWTExpire 管理端访问令牌有效期。
	AdminJWTExpire time.Duration
	// AdminUsername 管理员登录用户名。
	AdminUsername string
	// AdminPasswordBcrypt 管理员密码的 bcrypt 哈希（字符串，含 $2a$ 前缀）。
	AdminPasswordBcrypt string
	// AdminOperatorUsername 可选操作员用户名（与管理员分离）；须同时配置 OperatorPasswordBcrypt。
	AdminOperatorUsername string
	// AdminOperatorPasswordBcrypt 操作员 bcrypt；与 AdminOperatorUsername 配对。
	AdminOperatorPasswordBcrypt string
	// AdminRefreshExpire 「记住我」刷新令牌/长会话相对过期时间（若实现双令牌则使用；当前可与 JWT 过期合并配置）。
	AdminRefreshExpire time.Duration
	// AdminPasswordResetSMTP 预留：邮件重置发件配置（首版可不解析，仅占位）。
	AdminPasswordResetSMTP string
}

// EffectiveUpstreamBases 返回非空的上游基址列表（多上游优先，否则回退单键）。
func (c *Config) EffectiveUpstreamBases() []string {
	out := make([]string, 0, len(c.UpstreamBaseURLs)+1)
	for _, u := range c.UpstreamBaseURLs {
		u = strings.TrimSpace(u)
		if u != "" {
			out = append(out, u)
		}
	}
	if len(out) > 0 {
		return out
	}
	if strings.TrimSpace(c.UpstreamBaseURL) != "" {
		return []string{strings.TrimSpace(c.UpstreamBaseURL)}
	}
	return nil
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

	adminConsole := false
	if v.IsSet("NEXUSROUTER_ENABLE_ADMIN_CONSOLE") {
		adminConsole = v.GetBool("NEXUSROUTER_ENABLE_ADMIN_CONSOLE")
	}
	adminJWTExp := 24 * time.Hour
	if v.IsSet("NEXUSROUTER_ADMIN_JWT_EXPIRE") {
		if d, err := time.ParseDuration(v.GetString("NEXUSROUTER_ADMIN_JWT_EXPIRE")); err == nil {
			adminJWTExp = d
		}
	}
	adminRefreshExp := 7 * 24 * time.Hour
	if v.IsSet("NEXUSROUTER_ADMIN_REFRESH_EXPIRE") {
		if d, err := time.ParseDuration(v.GetString("NEXUSROUTER_ADMIN_REFRESH_EXPIRE")); err == nil {
			adminRefreshExp = d
		}
	}

	urlsCSV := strings.TrimSpace(v.GetString("NEXUSROUTER_UPSTREAM_BASE_URLS"))
	var urls []string
	if urlsCSV != "" {
		urls = splitCommaNonEmpty(urlsCSV)
	}

	httpAddr := strings.TrimSpace(v.GetString("NEXUSROUTER_HTTP_LISTEN_ADDR"))
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	return &Config{
		UpstreamBaseURL:             strings.TrimSpace(v.GetString("NEXUSROUTER_UPSTREAM_BASE_URL")),
		UpstreamBaseURLs:            urls,
		UpstreamTimeout:             timeout,
		GatewayAPIKeys:              splitKeys(v.GetString("NEXUSROUTER_GATEWAY_API_KEYS")),
		GatewayKeysFile:             strings.TrimSpace(v.GetString("NEXUSROUTER_GATEWAY_KEYS_FILE")),
		AdminReloadToken:            strings.TrimSpace(v.GetString("NEXUSROUTER_ADMIN_RELOAD_TOKEN")),
		UpstreamAPIKey:              strings.TrimSpace(v.GetString("NEXUSROUTER_UPSTREAM_API_KEY")),
		ForwardClientAuthorization:  fwd,
		EnableSwaggerUI:             swagger,
		GatewayConfigFile:           strings.TrimSpace(v.GetString("NEXUSROUTER_GATEWAY_CONFIG_FILE")),
		HTTPListenAddr:              httpAddr,
		EnableAdminConsole:          adminConsole,
		AdminJWTSecret:              strings.TrimSpace(v.GetString("NEXUSROUTER_ADMIN_JWT_SECRET")),
		AdminJWTExpire:              adminJWTExp,
		AdminUsername:               strings.TrimSpace(v.GetString("NEXUSROUTER_ADMIN_USERNAME")),
		AdminPasswordBcrypt:         strings.TrimSpace(v.GetString("NEXUSROUTER_ADMIN_PASSWORD_BCRYPT")),
		AdminOperatorUsername:       strings.TrimSpace(v.GetString("NEXUSROUTER_ADMIN_OPERATOR_USERNAME")),
		AdminOperatorPasswordBcrypt: strings.TrimSpace(v.GetString("NEXUSROUTER_ADMIN_OPERATOR_PASSWORD_BCRYPT")),
		AdminRefreshExpire:          adminRefreshExp,
		AdminPasswordResetSMTP:      strings.TrimSpace(v.GetString("NEXUSROUTER_ADMIN_PASSWORD_RESET_SMTP")),
	}
}

// AdminConsoleConfigured 管理控制台是否具备最小可登录配置。
func (c *Config) AdminConsoleConfigured() bool {
	if c == nil || !c.EnableAdminConsole {
		return false
	}
	if strings.TrimSpace(c.AdminJWTSecret) == "" || strings.TrimSpace(c.AdminUsername) == "" || strings.TrimSpace(c.AdminPasswordBcrypt) == "" {
		return false
	}
	return true
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

func splitCommaNonEmpty(s string) []string {
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
