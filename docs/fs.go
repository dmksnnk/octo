package docs

import "embed"

// Swagger contains the embedded swagger page for rendering OpenAPI docs and docs themselves.
//
//go:embed swagger.html openapi.yaml
var Swagger embed.FS
