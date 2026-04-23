package webassets

import (
	"embed"
	"io/fs"
)

//go:embed fallback/index.html all:dist
var embeddedFiles embed.FS

func hasDistIndex() bool {
	file, err := embeddedFiles.Open("dist/index.html")
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

func ReadIndexHTML() ([]byte, error) {
	if hasDistIndex() {
		return embeddedFiles.ReadFile("dist/index.html")
	}
	return embeddedFiles.ReadFile("fallback/index.html")
}

func StaticFS() (fs.FS, error) {
	if hasDistIndex() {
		return fs.Sub(embeddedFiles, "dist")
	}
	return fs.Sub(embeddedFiles, "fallback")
}
