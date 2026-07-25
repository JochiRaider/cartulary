//go:build linux

package securefile

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestSecureFileRead_Unit(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "manifest.json")
	if err := os.WriteFile(path, []byte(`{"schema":"v1"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	document, err := Read(path, 1024)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	first := document.Bytes()
	first[0] = 'X'
	if got := string(document.Bytes()); got != `{"schema":"v1"}` {
		t.Fatalf("document bytes were mutable: %q", got)
	}
	if document.Size() != int64(len(document.Bytes())) || document.Mode().Perm() != 0o600 {
		t.Fatalf("metadata = size %d mode %v", document.Size(), document.Mode())
	}
	if _, err := Read(path, 4); err == nil {
		t.Fatal("oversized manifest succeeded")
	} else {
		var typed *Error
		if !errors.As(err, &typed) || typed.Kind != FailureTooLarge {
			t.Fatalf("oversized manifest error = %v; want FailureTooLarge", err)
		}
	}
	if _, err := Read("relative.json", 1024); err == nil {
		t.Fatal("relative path succeeded")
	}
}

func TestSecureFileRejectsUnsafeObjectsAndRedactsPaths_Unit(t *testing.T) {
	root := t.TempDir()
	regular := filepath.Join(root, "regular")
	if err := os.WriteFile(regular, []byte("safe"), 0o600); err != nil {
		t.Fatal(err)
	}
	finalLink := filepath.Join(root, "final-link")
	if err := os.Symlink(regular, finalLink); err != nil {
		t.Fatal(err)
	}
	parentLink := filepath.Join(root, "parent-link")
	if err := os.Symlink(root, parentLink); err != nil {
		t.Fatal(err)
	}
	fifo := filepath.Join(root, "fifo")
	if err := unix.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	socket := filepath.Join(root, "socket")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	for name, path := range map[string]string{
		"final symlink":  finalLink,
		"parent symlink": filepath.Join(parentLink, "regular"),
		"directory":      root,
		"FIFO":           fifo,
		"socket":         socket,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Read(path, 1024)
			if err == nil {
				t.Fatal("unsafe object succeeded")
			}
			if strings.Contains(err.Error(), root) || strings.Contains(err.Error(), path) {
				t.Fatalf("error disclosed host path: %v", err)
			}
			var typed *Error
			if !errors.As(err, &typed) {
				t.Fatalf("error type = %T; want *Error", err)
			}
		})
	}
}
