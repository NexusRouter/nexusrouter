package openapi

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestSpecYAMLBytes_OpenAPI3 校验 go:embed 的 openapi.yaml 非空且为 OpenAPI 3.0.x（须先 make docs）。
func TestSpecYAMLBytes_OpenAPI3(t *testing.T) {
	raw := SpecYAMLBytes()
	if len(raw) == 0 {
		t.Fatal("嵌入的 openapi.yaml 为空，请在 services/gateway 下执行 make docs")
	}
	if !strings.Contains(string(raw), "openapi:") {
		t.Fatal("YAML 缺少 openapi 根字段")
	}
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		t.Fatalf("解析嵌入 YAML: %v", err)
	}
	ver, _ := root["openapi"].(string)
	if ver == "" || !strings.HasPrefix(ver, "3.0.") {
		t.Fatalf("openapi 版本应为 3.0.x 前缀，得到 %q", ver)
	}
}
