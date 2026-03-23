// Package web embeds the static frontend assets (CSS, JS) so the binary
// remains a single, self-contained executable.
package web

import "embed"

//go:embed static
var Static embed.FS
