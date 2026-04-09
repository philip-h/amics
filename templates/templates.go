package templates

import "embed"

//go:embed pages/*.html partials/*.html layouts/*.html
var TemplateFS embed.FS
