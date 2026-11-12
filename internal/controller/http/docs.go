package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/docs"
)

const swaggerHTML = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>RePurpose.ai API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/docs/openapi.yaml",
      dom_id: "#swagger-ui",
      deepLinking: true,
      persistAuthorization: true
    });
  </script>
</body>
</html>`

// RegisterDocs mounts the interactive API documentation and raw OpenAPI spec.
func RegisterDocs(app *gin.Engine) {
	app.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
	})

	app.GET("/docs/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml", docs.OpenAPISpec)
	})

	app.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs")
	})
}
