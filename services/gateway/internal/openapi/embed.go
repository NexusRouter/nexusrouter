package openapi

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gopkg.in/yaml.v3"
)

// openapi.yaml 由 make docs（swag → swagger2openapi）生成，勿手改；单一事实来源为 Go 内 swag 注释。
//
//go:embed openapi.yaml
var specYAML []byte

var specJSON []byte

func init() {
	var root any
	if err := yaml.Unmarshal(specYAML, &root); err != nil {
		panic("openapi: invalid embedded yaml: " + err.Error())
	}
	j, err := json.Marshal(root)
	if err != nil {
		panic("openapi: json marshal: " + err.Error())
	}
	specJSON = j
}

// Register 注册 OpenAPI 规范与 Swagger UI（当 enableSwagger 为 true）。
func Register(r *gin.Engine, enableSwagger bool) {
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml", specYAML)
	})
	r.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", specJSON)
	})
	if !enableSwagger {
		return
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/openapi.yaml"),
	))
}

// SpecYAMLBytes 返回嵌入的 OpenAPI YAML（供测试解析）。
func SpecYAMLBytes() []byte {
	return bytes.Clone(specYAML)
}
