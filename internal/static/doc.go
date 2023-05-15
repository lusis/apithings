// Package static contains embedded htmx templates
package static

import "embed"

// StaticFS is the filesystem storing static html contents
//
//go:embed files/*
var StaticFS embed.FS
