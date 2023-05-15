// Package templates contains embedded htmx templates
package templates

import "embed"

// UITemplateFS is the filesystem storing htmx templates
//
//go:embed *
var UITemplateFS embed.FS
