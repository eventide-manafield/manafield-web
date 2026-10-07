package webassets

import "embed"

// FS contains the Web Shell templates and static assets.
//
//go:embed templates/*.html static/*
var FS embed.FS
