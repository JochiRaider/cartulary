package webassets

import (
	"archive/zip"
	"bytes"
	"embed"
	"errors"
	"io/fs"
)

//go:embed fallback/index.html all:dist
var embeddedFiles embed.FS

const distArchivePath = "dist/web-assets.zip"

func ReadIndexHTML() ([]byte, error) {
	return readIndexHTML(embeddedFiles)
}

func StaticFS() (fs.FS, error) {
	return staticFS(embeddedFiles)
}

func readIndexHTML(files fs.FS) ([]byte, error) {
	archive, ok, err := openDistArchive(files)
	if err != nil {
		return nil, err
	}
	if ok {
		return fs.ReadFile(archive, "index.html")
	}
	return fs.ReadFile(files, "fallback/index.html")
}

func staticFS(files fs.FS) (fs.FS, error) {
	archive, ok, err := openDistArchive(files)
	if err != nil {
		return nil, err
	}
	if ok {
		return archive, nil
	}
	return fs.Sub(files, "fallback")
}

func openDistArchive(files fs.FS) (*zip.Reader, bool, error) {
	data, err := fs.ReadFile(files, distArchivePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, false, err
	}
	return reader, true, nil
}
