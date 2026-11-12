package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distEmbed embed.FS

// GetFS returns the embedded web assets filesystem.
func GetFS() fs.FS {
	sub, err := fs.Sub(distEmbed, "dist")
	if err != nil {
		return distEmbed
	}
	return sub
}
