//go:build linux

package rootedfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestRootedFSReferenceValidation_Unit(t *testing.T) {
	for name, raw := range map[string]string{
		"empty":                "",
		"absolute":             "/archive/file",
		"dot":                  ".",
		"dot component":        "archive/./file",
		"parent":               "archive/../file",
		"empty component":      "archive//file",
		"trailing separator":   "archive/file/",
		"backslash":            `archive\file`,
		"NUL":                  "archive/\x00file",
		"noncanonical Unicode": "cafe\u0301/file",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseReference(raw); !errors.Is(err, ErrInvalidReference) {
				t.Fatalf("ParseReference(%q) error = %v; want ErrInvalidReference", raw, err)
			}
		})
	}
	reference, err := ParseReference("archive/private/file.json")
	if err != nil || reference.String() != "archive/private/file.json" {
		t.Fatalf("canonical reference = %q, %v", reference.String(), err)
	}

	for name, references := range map[string][]string{
		"duplicate":               {"a/b", "a/b"},
		"prefix":                  {"a/b", "a/b/c"},
		"normalization collision": {"café/file", "cafe\u0301/file"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ValidateReferenceSet(references); !errors.Is(err, ErrReferenceCollision) {
				t.Fatalf("ValidateReferenceSet(%q) error = %v; want ErrReferenceCollision", references, err)
			}
		})
	}
}

func TestRootedFSOperationContainment_Unit(t *testing.T) {
	rootPath := filepath.Join(t.TempDir(), "storage")
	if err := os.Mkdir(rootPath, 0o700); err != nil {
		t.Fatal(err)
	}
	root, err := Open(rootPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = root.Close() })

	directory := MustParseReference("private/workspace")
	if err := root.MakePrivateDir(directory); err != nil {
		t.Fatalf("MakePrivateDir: %v", err)
	}
	info, err := os.Stat(filepath.Join(rootPath, "private", "workspace"))
	if err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("private directory mode = %v, %v; want 0700", info.Mode().Perm(), err)
	}

	reference := MustParseReference("private/workspace/bundle.json")
	if err := root.CreateExclusive(context.Background(), reference, func(writer io.Writer) error {
		_, writeErr := writer.Write([]byte(`{"ok":true}`))
		return writeErr
	}); err != nil {
		t.Fatalf("CreateExclusive: %v", err)
	}
	fileInfo, err := os.Stat(filepath.Join(rootPath, filepath.FromSlash(reference.String())))
	if err != nil || fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf("private file mode = %v, %v; want 0600", fileInfo.Mode().Perm(), err)
	}
	payload, metadata, err := root.ReadRegular(reference, 1024)
	if err != nil || string(payload) != `{"ok":true}` || metadata.Size != int64(len(payload)) {
		t.Fatalf("ReadRegular = %q, %+v, %v", payload, metadata, err)
	}
	payload[0] = 'X'
	again, _, err := root.ReadRegular(reference, 1024)
	if err != nil || string(again) != `{"ok":true}` {
		t.Fatalf("read result was not immutable: %q, %v", again, err)
	}
	handle, openedMetadata, err := root.OpenRegular(reference)
	if err != nil {
		t.Fatalf("OpenRegular: %v", err)
	}
	streamed, err := io.ReadAll(handle)
	closeErr := handle.Close()
	if err != nil || closeErr != nil || string(streamed) != `{"ok":true}` || openedMetadata.Size != int64(len(streamed)) {
		t.Fatalf("OpenRegular payload = %q metadata=%+v read=%v close=%v", streamed, openedMetadata, err, closeErr)
	}
	entries, err := root.ListRegular()
	if err != nil {
		t.Fatalf("ListRegular: %v", err)
	}
	if len(entries) != 1 || entries[0].Reference.String() != reference.String() || entries[0].Metadata.Size != int64(len(streamed)) {
		t.Fatalf("ListRegular = %#v", entries)
	}
	if err := root.CreateExclusive(context.Background(), reference, func(io.Writer) error { return nil }); err == nil {
		t.Fatal("duplicate exclusive creation succeeded")
	}

	if err := os.Symlink(filepath.Join(rootPath, "private"), filepath.Join(rootPath, "linked-parent")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := root.ReadRegular(MustParseReference("linked-parent/workspace/bundle.json"), 1024); err == nil {
		t.Fatal("parent symlink read succeeded")
	}
	if err := os.Symlink(filepath.Join(rootPath, "private", "workspace", "bundle.json"), filepath.Join(rootPath, "final-link")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := root.ReadRegular(MustParseReference("final-link"), 1024); err == nil {
		t.Fatal("final symlink read succeeded")
	}
	if err := os.Symlink("loop", filepath.Join(rootPath, "loop")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := root.ReadRegular(MustParseReference("loop/file"), 1024); err == nil {
		t.Fatal("symlink loop read succeeded")
	}

	if err := os.Link(filepath.Join(rootPath, "private", "workspace", "bundle.json"), filepath.Join(rootPath, "hard-link")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := root.ReadRegular(MustParseReference("hard-link"), 1024); err == nil {
		t.Fatal("hard-link read succeeded")
	}
	if _, err := root.ListRegular(); err == nil {
		t.Fatal("listing with a hard link succeeded")
	}
	if err := unix.Mkfifo(filepath.Join(rootPath, "fifo"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := root.ReadRegular(MustParseReference("fifo"), 1024); err == nil {
		t.Fatal("FIFO read succeeded")
	}
	socketPath := filepath.Join(rootPath, "socket")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	if _, _, err := root.ReadRegular(MustParseReference("socket"), 1024); err == nil {
		t.Fatal("socket read succeeded")
	}
	deviceRoot, err := Open("/dev")
	if err != nil {
		t.Fatalf("open device root: %v", err)
	}
	t.Cleanup(func() { _ = deviceRoot.Close() })
	if _, _, err := deviceRoot.ReadRegular(MustParseReference("null"), 1024); err == nil {
		t.Fatal("device read succeeded")
	}
}

func TestRootedFSAtomicLifecycleAndRootIdentity_Unit(t *testing.T) {
	parent := t.TempDir()
	rootPath := filepath.Join(parent, "private", "nested", "storage")
	root, err := OpenOrCreate(rootPath)
	if err != nil {
		t.Fatalf("OpenOrCreate: %v", err)
	}
	t.Cleanup(func() { _ = root.Close() })
	for current := rootPath; current != parent; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil {
			t.Fatalf("stat created root component: %v", err)
		}
		if !info.IsDir() || info.Mode().Perm() != 0o700 {
			t.Fatalf("created root component %s mode = %v; want private directory", current, info.Mode())
		}
	}
	reference := MustParseReference("artifact")

	write := func(value string) WriteFunc {
		return func(writer io.Writer) error {
			_, err := io.WriteString(writer, value)
			return err
		}
	}
	if err := root.AtomicReplace(context.Background(), reference, write("v1")); err != nil {
		t.Fatalf("initial AtomicReplace: %v", err)
	}
	failedExclusive := MustParseReference("failed-exclusive")
	if err := root.CreateExclusive(context.Background(), failedExclusive, func(writer io.Writer) error {
		_, _ = io.WriteString(writer, "partial")
		return errors.New("stop")
	}); err == nil {
		t.Fatal("failed exclusive creation succeeded")
	}
	if _, err := os.Stat(filepath.Join(rootPath, failedExclusive.String())); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed exclusive creation leaked partial file: %v", err)
	}
	exclusive := MustParseReference("exclusive")
	if err := root.CreateExclusive(context.Background(), exclusive, func(writer io.Writer) error {
		if _, err := os.Stat(filepath.Join(rootPath, exclusive.String())); !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("exclusive destination became visible before publication: %w", err)
		}
		_, err := io.WriteString(writer, "complete")
		return err
	}); err != nil {
		t.Fatalf("exclusive atomic publication: %v", err)
	}
	payload, _, err := root.ReadRegular(exclusive, 32)
	if err != nil || string(payload) != "complete" {
		t.Fatalf("exclusive publication bytes = %q, %v", payload, err)
	}
	if err := root.CreateExclusive(context.Background(), exclusive, write("replacement")); !errors.Is(err, os.ErrExist) {
		t.Fatalf("duplicate exclusive creation error = %v; want fs.ErrExist", err)
	}
	if err := root.AtomicReplace(context.Background(), reference, func(writer io.Writer) error {
		_, _ = io.WriteString(writer, "partial")
		return errors.New("stop")
	}); err == nil {
		t.Fatal("failed replacement succeeded")
	}
	payload, _, err = root.ReadRegular(reference, 32)
	if err != nil || string(payload) != "v1" {
		t.Fatalf("failed replacement changed publication: %q, %v", payload, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := root.AtomicReplace(ctx, reference, func(writer io.Writer) error {
		_, _ = io.WriteString(writer, "canceled")
		cancel()
		return nil
	}); err == nil {
		t.Fatal("canceled replacement succeeded")
	}
	payload, _, err = root.ReadRegular(reference, 32)
	if err != nil || string(payload) != "v1" {
		t.Fatalf("canceled replacement changed publication: %q, %v", payload, err)
	}
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".cartulary-tmp-") {
			t.Fatalf("temporary file leaked: %s", entry.Name())
		}
	}

	destination := MustParseReference("renamed")
	if err := root.RenameExclusive(reference, destination); err != nil {
		t.Fatalf("RenameExclusive: %v", err)
	}
	if err := root.RemoveRegular(destination); err != nil {
		t.Fatalf("RemoveRegular: %v", err)
	}
	if _, _, err := root.ReadRegular(destination, 32); err == nil {
		t.Fatal("removed file remained readable")
	}

	replacementPath := filepath.Join(parent, "replacement")
	if err := os.Rename(rootPath, replacementPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(rootPath, 0o700); err != nil {
		t.Fatal(err)
	}
	err = root.CreateExclusive(context.Background(), MustParseReference("must-not-publish"), write("secret"))
	if !errors.Is(err, ErrRootIdentityChanged) {
		t.Fatalf("operation after root replacement error = %v; want ErrRootIdentityChanged", err)
	}
	if _, err := os.Stat(filepath.Join(rootPath, "must-not-publish")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("replacement root received publication: %v", err)
	}
}

func TestRootedFSDetectsComponentAndRootReplacementRaces_Unit(t *testing.T) {
	parent := t.TempDir()
	rootPath := filepath.Join(parent, "storage")
	if err := os.Mkdir(rootPath, 0o700); err != nil {
		t.Fatal(err)
	}
	root, err := Open(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	if err := root.MakePrivateDir(MustParseReference("workspace/private")); err != nil {
		t.Fatal(err)
	}

	outside := filepath.Join(parent, "outside")
	if err := os.Mkdir(outside, 0o700); err != nil {
		t.Fatal(err)
	}
	moved := filepath.Join(rootPath, "workspace", "moved")
	reference := MustParseReference("workspace/private/artifact")
	err = root.CreateExclusive(context.Background(), reference, func(writer io.Writer) error {
		if _, writeErr := io.WriteString(writer, "partial"); writeErr != nil {
			return writeErr
		}
		if renameErr := os.Rename(filepath.Join(rootPath, "workspace", "private"), moved); renameErr != nil {
			return renameErr
		}
		return os.Symlink(outside, filepath.Join(rootPath, "workspace", "private"))
	})
	if !errors.Is(err, ErrRootIdentityChanged) {
		t.Fatalf("component replacement error = %v; want ErrRootIdentityChanged", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "artifact")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("component replacement escaped write: %v", err)
	}
	if _, err := os.Stat(filepath.Join(moved, "artifact")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("component replacement retained partial write: %v", err)
	}

	if err := os.Remove(filepath.Join(rootPath, "workspace", "private")); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(moved, filepath.Join(rootPath, "workspace", "private")); err != nil {
		t.Fatal(err)
	}
	movedRoot := filepath.Join(parent, "moved-root")
	err = root.AtomicReplace(context.Background(), MustParseReference("workspace/private/published"), func(writer io.Writer) error {
		if _, writeErr := io.WriteString(writer, "partial"); writeErr != nil {
			return writeErr
		}
		if renameErr := os.Rename(rootPath, movedRoot); renameErr != nil {
			return renameErr
		}
		return os.Mkdir(rootPath, 0o700)
	})
	if !errors.Is(err, ErrRootIdentityChanged) {
		t.Fatalf("root replacement race error = %v; want ErrRootIdentityChanged", err)
	}
	if _, err := os.Stat(filepath.Join(rootPath, "workspace", "private", "published")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("replacement root received publication: %v", err)
	}
	if _, err := os.Stat(filepath.Join(movedRoot, "workspace", "private", "published")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("moved root retained partial publication: %v", err)
	}
}

func TestRootedFSErrorsDoNotDiscloseHostRoot_Unit(t *testing.T) {
	rootPath := filepath.Join(t.TempDir(), "secret-deployment-root")
	if err := os.Mkdir(rootPath, 0o700); err != nil {
		t.Fatal(err)
	}
	root, err := Open(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	_, _, err = root.ReadRegular(MustParseReference("missing"), 16)
	if err == nil || strings.Contains(err.Error(), rootPath) {
		t.Fatalf("operation error disclosed root: %v", err)
	}
	if _, err := Open(filepath.Join(rootPath, "missing")); err == nil || strings.Contains(err.Error(), rootPath) {
		t.Fatalf("constructor error disclosed root: %v", err)
	}
	if err := root.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := root.Close(); err != nil {
		t.Fatalf("idempotent Close: %v", err)
	}
	if err := root.Check(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Check after close = %v; want ErrClosed", err)
	}
}
