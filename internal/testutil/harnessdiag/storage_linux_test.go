//go:build linux

package harnessdiag

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

func TestDiagnosticStorageFaults(t *testing.T) {
	root := filepath.Join(t.TempDir(), "private")
	dir, err := harnesslocal.PrivateDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	dir.Close()
	leaf := filepath.Join(root, "sample")
	if err := harnesslocal.PublishNew(leaf, []byte("original")); err != nil {
		t.Fatal(err)
	}
	if err := harnesslocal.PublishNew(leaf, []byte("replacement")); err == nil {
		t.Fatal("overwrote published artifact")
	}
	original, _ := os.ReadFile(leaf)
	if string(original) != "original" {
		t.Fatal("changed immutable bytes")
	}
	stream, err := harnesslocal.NewStream(leaf, 16)
	if err != nil {
		t.Fatal(err)
	}
	if err = stream.Write([]byte("stream overwrite")); err != nil {
		t.Fatal(err)
	}
	if stream.Close() == nil {
		t.Fatal("stream overwrote published artifact")
	}
	if err := os.Symlink(leaf, filepath.Join(root, "symlink")); err != nil {
		t.Fatal(err)
	}
	if harnesslocal.ReplaceFile(filepath.Join(root, "symlink"), []byte("mutated")) == nil {
		t.Fatal("followed symlink")
	}
	if err := os.Link(leaf, filepath.Join(root, "hardlink")); err != nil {
		t.Fatal(err)
	}
	if harnesslocal.ReplaceFile(leaf, []byte("mutated")) == nil {
		t.Fatal("replaced multiply linked state")
	}
	os.Remove(filepath.Join(root, "hardlink"))
	entries, _ := os.ReadDir(root)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".publishing-") {
			t.Fatal("failed publication left staging file")
		}
	}
	stream, err = harnesslocal.NewStream(filepath.Join(root, "moved"), 16)
	if err != nil {
		t.Fatal(err)
	}
	if err = stream.Write([]byte("must not publish")); err != nil {
		t.Fatal(err)
	}
	old := root + "-old"
	if err = os.Rename(root, old); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if stream.Close() == nil {
		t.Fatal("published through replaced directory")
	}
	for _, directory := range []string{root, old} {
		entries, _ = os.ReadDir(directory)
		for _, entry := range entries {
			if entry.Name() == "moved" || strings.HasPrefix(entry.Name(), ".publishing-") {
				t.Fatal("replaced directory retained unauthorized publication")
			}
		}
	}
}
