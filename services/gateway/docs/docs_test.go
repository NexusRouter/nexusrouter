package docs

import "testing"

// TestSwaggerInfoPresent swag 生成物已注册且含标题（make docs 后与本测试同源）。
func TestSwaggerInfoPresent(t *testing.T) {
	if SwaggerInfo == nil {
		t.Fatal("SwaggerInfo 为 nil")
	}
	if SwaggerInfo.Title == "" {
		t.Fatal("SwaggerInfo.Title 为空")
	}
}
