#!/usr/bin/env sh
# 提交前尽量对齐 .github/workflows/ci.yml（未改动的子项目会跳过）。
# 未包含：Gateway 的 make docs（耗时长，推送前请在改 swag 注释时本地执行）。
set -eu
ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

STAGED="$(git diff --cached --name-only --diff-filter=ACMRTUXB)"
if [ -z "$STAGED" ]; then
  exit 0
fi

need_gw=0
need_fe=0
need_spec=0
while IFS= read -r f; do
  [ -z "$f" ] && continue
  case "$f" in
    services/gateway/*|services/gateway/go.mod|services/gateway/go.sum) need_gw=1 ;;
    web/dashboard/*|web/dashboard/package.json|web/dashboard/pnpm-lock.yaml) need_fe=1 ;;
    openspec/*) need_spec=1 ;;
    scripts/*|.husky/*|.github/workflows/*) need_gw=1; need_fe=1; need_spec=1 ;;
  esac
done <<EOF
$STAGED
EOF

if [ "$need_gw" = 1 ]; then
  echo ">>> pre-commit: Gateway (Go)"
  cd "$ROOT/services/gateway"
  if [ -n "$(gofmt -l .)" ]; then
    echo "gofmt: 以下文件未格式化，请在 services/gateway 下执行 gofmt -w 后重新暂存。"
    gofmt -l . | sed 's/^/  /'
    exit 1
  fi
  if ! command -v golangci-lint >/dev/null 2>&1; then
    echo "未找到 golangci-lint。请安装与 CI 相同版本，例如:"
    echo "  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4"
    exit 1
  fi
  golangci-lint run
  go vet ./...
  go test ./... -count=1
  go build -o /dev/null ./cmd/api
  echo ">>> Gateway OK"
fi

if [ "$need_fe" = 1 ]; then
  echo ">>> pre-commit: Dashboard (pnpm)"
  cd "$ROOT/web/dashboard"
  if ! command -v pnpm >/dev/null 2>&1; then
    echo "未找到 pnpm。请: corepack enable && corepack prepare pnpm@9.15.9 --activate"
    exit 1
  fi
  pnpm install --frozen-lockfile
  pnpm run lint
  pnpm run test:coverage
  pnpm run build
  echo ">>> Dashboard OK"
fi

if [ "$need_spec" = 1 ]; then
  echo ">>> pre-commit: OpenSpec"
  if ! command -v openspec >/dev/null 2>&1; then
    echo "未找到 openspec CLI。请: npm install -g @fission-ai/openspec@latest"
    exit 1
  fi
  openspec validate --all --no-interactive
  echo ">>> OpenSpec OK"
fi

exit 0
