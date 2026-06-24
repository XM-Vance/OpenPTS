// API 文档：OpenAPI 3.1 规范 + Swagger UI 静态托管。
// /docs/openapi.yaml → 文档源（YAML）
// /docs              → Swagger UI（HTML 内嵌从 CDN 加载 swagger-ui，无外部依赖）
package handler

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed swagger_ui.html
var swaggerUIHTML string

//go:embed openapi.yaml
var openapiYAML string

// OpenAPISpec 返回 YAML 格式的 OpenAPI 规范。
func OpenAPISpec(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", []byte(openapiYAML))
}

// SwaggerUI 返回 Swagger UI 容器页面。
func SwaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIHTML))
}
