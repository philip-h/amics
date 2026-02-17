package templates

import "embed"

//go:embed pages/*.gohtml partials/*.gohtml
var TemplateFS embed.FS
