package main

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteArchiveIsDeterministicAndServesRootEntries(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	mustWriteFile(t, filepath.Join(sourceDir, "index.html"), `<script src="/assets/app.js"></script>`)
	mustWriteFile(t, filepath.Join(sourceDir, "assets", "app.js"), "console.log('ok');")

	first := filepath.Join(t.TempDir(), "web-assets.zip")
	second := filepath.Join(t.TempDir(), "web-assets.zip")
	if err := writeArchive(sourceDir, filepath.Join(sourceDir, "index.html"), first, t.TempDir()); err != nil {
		t.Fatalf("write first archive: %v", err)
	}
	if err := writeArchive(sourceDir, filepath.Join(sourceDir, "index.html"), second, t.TempDir()); err != nil {
		t.Fatalf("write second archive: %v", err)
	}

	firstBytes, err := os.ReadFile(first)
	if err != nil {
		t.Fatalf("read first archive: %v", err)
	}
	secondBytes, err := os.ReadFile(second)
	if err != nil {
		t.Fatalf("read second archive: %v", err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("archive output is not deterministic")
	}

	reader, err := zip.NewReader(bytes.NewReader(firstBytes), int64(len(firstBytes)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	indexBytes, err := reader.Open("index.html")
	if err != nil {
		t.Fatalf("open index.html from zip: %v", err)
	}
	defer indexBytes.Close()
	assetBytes, err := reader.Open("assets/app.js")
	if err != nil {
		t.Fatalf("open asset from zip: %v", err)
	}
	defer assetBytes.Close()
}

func TestWriteArchiveRejectsSymlinks(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	mustWriteFile(t, filepath.Join(sourceDir, "index.html"), "ok")
	if err := os.Symlink(filepath.Join(sourceDir, "index.html"), filepath.Join(sourceDir, "linked.html")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	err := writeArchive(sourceDir, filepath.Join(sourceDir, "index.html"), filepath.Join(t.TempDir(), "web-assets.zip"), t.TempDir())
	if err == nil {
		t.Fatalf("expected symlink rejection")
	}
	if !strings.Contains(err.Error(), "must not be a symlink") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteArchiveRequiresIndexInsideSource(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	otherDir := t.TempDir()
	mustWriteFile(t, filepath.Join(sourceDir, "index.html"), "ok")
	mustWriteFile(t, filepath.Join(otherDir, "index.html"), "wrong")

	err := writeArchive(sourceDir, filepath.Join(otherDir, "index.html"), filepath.Join(t.TempDir(), "web-assets.zip"), t.TempDir())
	if err == nil {
		t.Fatalf("expected external index rejection")
	}
	if !strings.Contains(err.Error(), "source index must be inside source dir") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
