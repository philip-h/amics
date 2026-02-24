package templates

import "embed"

//go:embed pages/*.gohtml partials/*.gohtml admin/*.gohtml
var TemplateFS embed.FS
