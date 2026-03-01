package templates

import "embed"

//go:embed pages/*.html partials/*.html admin/*.html
var TemplateFS embed.FS
