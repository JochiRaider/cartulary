//go:build linux

package rootedfs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestRootedFSReadOnlyInspection_Unit(t *testing.T) {
	base := t.TempDir()
	absent := filepath.Join(base, "absent", "nested")
	if empty, err := InspectEmpty(absent); err != nil || !empty {
		t.Fatalf("absent: %t %v", empty, err)
	}
	if _, err := os.Stat(filepath.Dir(absent)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("inspection created root: %v", err)
	}
	if empty, err := InspectEmpty(base); err != nil || !empty {
		t.Fatalf("empty: %t %v", empty, err)
	}
	for _, kind := range []string{"file", "directory", "link"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			entry := filepath.Join(root, "entry")
			var err error
			switch kind {
			case "file":
				err = os.WriteFile(entry, nil, 0o600)
			case "directory":
				err = os.Mkdir(entry, 0o700)
			case "link":
				err = os.Symlink("missing", entry)
			}
			if err != nil {
				t.Fatal(err)
			}
			if empty, err := InspectEmpty(root); err != nil || empty {
				t.Fatalf("retained: %t %v", empty, err)
			}
		})
	}
	link := filepath.Join(base, "unsafe")
	if err := os.Symlink(t.TempDir(), link); err != nil {
		t.Fatal(err)
	}
	if empty, err := InspectEmpty(filepath.Join(link, "absent")); err == nil || empty {
		t.Fatalf("unsafe: %t %v", empty, err)
	}
}
