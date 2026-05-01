.PHONY: docs wire
# 根据 swag 注释重新生成 docs/（swagger.json / swagger.yaml / docs.go）；OpenAPI 3 机器入口为 internal/openapi/openapi.yaml，需手工维护或与上游契约对齐后提交。
docs:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d ./cmd/api,./internal/handler --parseInternal -o ./docs

wire:
	cd cmd/api && go run github.com/google/wire/cmd/wire@v0.7.0
