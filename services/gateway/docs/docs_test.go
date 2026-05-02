package docs

import "testing"

// TestSwaggerInfoPresent docs 包内 swag 生成物已注册且含标题。
func TestSwaggerInfoPresent(t *testing.T) {
	if SwaggerInfo == nil {
		t.Fatal("SwaggerInfo 为 nil")
	}
	if SwaggerInfo.Title == "" {
		t.Fatal("SwaggerInfo.Title 为空")
	}
}
