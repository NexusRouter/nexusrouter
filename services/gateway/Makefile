.PHONY: docs wire lint test build-api

# 可选：从 Go 内 swag 注释重新生成 docs/ 与 internal/openapi/openapi.yaml（需 Node.js npx）。
# 日常构建与 CI 不执行；更新注释后如需刷新嵌入的 OpenAPI，可本地运行 make docs 并提交变更。
docs:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d ./cmd/api,./internal/handler --parseInternal -o ./docs
	npx --yes swagger2openapi@7.0.0 docs/swagger.yaml -o internal/openapi/openapi.yaml

test:
	go test ./... -count=1 -cover

build-api:
	go build -o /tmp/nexusrouter-gateway ./cmd/api

wire:
	cd cmd/api && go run github.com/google/wire/cmd/wire@v0.7.0

# 需已安装 golangci-lint（与 CI 同版本）：go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4
lint:
	golangci-lint run
