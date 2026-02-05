package templates

import "embed"

// EmbeddedFS contains the default template files shipped with the binary.
//
//go:embed SOUL.md USER.md skills/*.md
var EmbeddedFS embed.FS
