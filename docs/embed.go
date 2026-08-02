package docs

import _ "embed"

// SwaggerYAML contiene la especificación OpenAPI incluida en el binario.
//
//go:embed swagger.yaml
var SwaggerYAML []byte
