.PHONY: wire lint test build-api

test:
	go test ./... -count=1 -cover

build-api:
	go build -o /tmp/nexusrouter-gateway ./cmd/api

wire:
	cd cmd/api && go run github.com/google/wire/cmd/wire@v0.7.0

# 需已安装 golangci-lint（与 CI 同版本）：go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4
lint:
	golangci-lint run
