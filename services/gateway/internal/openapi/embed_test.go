package openapi

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestSpecYAMLBytes_OpenAPI3 校验 go:embed 的 openapi.yaml 非空且为 OpenAPI 3.0.x。
func TestSpecYAMLBytes_OpenAPI3(t *testing.T) {
	raw := SpecYAMLBytes()
	if len(raw) == 0 {
		t.Fatal("嵌入的 openapi.yaml 为空，请检查 internal/openapi/openapi.yaml 是否已提交")
	}
	if !strings.Contains(string(raw), "openapi:") {
		t.Fatal("YAML 缺少 openapi 根字段")
	}
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		t.Fatalf("解析嵌入 YAML: %v", err)
	}
	ver, ok := root["openapi"].(string)
	if !ok || ver == "" || !strings.HasPrefix(ver, "3.0.") {
		t.Fatalf("openapi 版本应为 3.0.x 前缀，得到 %q (ok=%v)", ver, ok)
	}
}
