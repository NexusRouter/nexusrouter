package main

import (
	"log"

	"go.uber.org/zap"
)

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
