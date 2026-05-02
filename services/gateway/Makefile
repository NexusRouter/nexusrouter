.PHONY: docs wire lint test build-api

# 单一事实来源：Go 源码中的 swag 注释。
# - docs/：swag init 生成的 Swagger 2（docs.go、swagger.json、swagger.yaml）
# - internal/openapi/openapi.yaml：由 swagger2openapi 从 docs/swagger.yaml 转换的 OAS3，go:embed 后运行时提供 /openapi.*
#   该文件已 .gitignore；go build / go test 前须成功执行过 docs（见 test、build-api 目标）。
# 本地执行 make docs 需已安装 Node.js（用于 npx）。
docs:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d ./cmd/api,./internal/handler --parseInternal -o ./docs
	npx --yes swagger2openapi@7.0.0 docs/swagger.yaml -o internal/openapi/openapi.yaml

test: docs
	go test ./... -count=1 -cover

build-api: docs
	go build -o /tmp/nexusrouter-gateway ./cmd/api

wire:
	cd cmd/api && go run github.com/google/wire/cmd/wire@v0.7.0

# 需已安装 golangci-lint（与 CI 同版本）：go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4
lint:
	golangci-lint run
