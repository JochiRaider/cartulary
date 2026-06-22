package webassets

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestReadIndexHTMLUsesArchiveWhenPresent(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{
		"fallback/index.html": {Data: []byte("fallback")},
		distArchivePath:       {Data: testArchive(t, map[string]string{"index.html": "dist"})},
	}

	got, err := readIndexHTML(files)
	if err != nil {
		t.Fatalf("readIndexHTML(): %v", err)
	}
	if string(got) != "dist" {
		t.Fatalf("readIndexHTML()=%q, want dist", string(got))
	}
}

func TestStaticFSUsesArchiveWhenPresent(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{
		"fallback/index.html": {Data: []byte("fallback")},
		distArchivePath: {
			Data: testArchive(t, map[string]string{
				"index.html":    "dist",
				"assets/app.js": "console.log('ok');",
			}),
		},
	}

	static, err := staticFS(files)
	if err != nil {
		t.Fatalf("staticFS(): %v", err)
	}
	got, err := fs.ReadFile(static, "assets/app.js")
	if err != nil {
		t.Fatalf("read static asset: %v", err)
	}
	if string(got) != "console.log('ok');" {
		t.Fatalf("asset=%q", string(got))
	}
}

func TestReadIndexHTMLFallsBackWithoutArchive(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{
		"fallback/index.html": {Data: []byte("fallback")},
		"dist/index.html":     {Data: []byte("legacy loose dist is ignored")},
	}

	got, err := readIndexHTML(files)
	if err != nil {
		t.Fatalf("readIndexHTML(): %v", err)
	}
	if string(got) != "fallback" {
		t.Fatalf("readIndexHTML()=%q, want fallback", string(got))
	}
}

func testArchive(t *testing.T, entries map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for name, content := range entries {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := file.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}
