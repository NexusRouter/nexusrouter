package main

import (
	"log"

	_ "github.com/NexusRouter/nexusrouter/services/gateway/docs" // swag 生成文档注册
	"go.uber.org/zap"
)

// NexusRouter API 网关（OpenAI 兼容 Chat Completions 反向代理 + OpenAPI 3 文档）。
//
//	@title			NexusRouter Gateway API
//	@version		1.0.0
//	@description	OpenAI 兼容 Chat Completions 网关。REST 概览与认证约定参见 https://developers.openai.com/api/reference/overview
//	@BasePath		/
//	@host			localhost:8080
//	@externalDocs.description	OpenAI API Overview
//	@externalDocs.url			https://developers.openai.com/api/reference/overview
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Bearer 令牌，与 OpenAI 一致（见官方概览）。
func main() {
	app, err := InitializeApp()
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer func(l *zap.Logger) { _ = l.Sync() }(app.Log)

	app.Log.Info("网关启动", zap.String("addr", ":8080"))
	if err := app.Run(); err != nil {
		app.Log.Fatal("服务退出", zap.Error(err))
	}
}
