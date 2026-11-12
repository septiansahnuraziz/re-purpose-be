package docs

import _ "embed"

// OpenAPISpec is served by GET /docs/openapi.yaml.
//
//go:embed openapi.yaml
var OpenAPISpec []byte
