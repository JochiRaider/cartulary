//go:build linux

package rootedfs

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRootedFSPrivateStage_Unit(t *testing.T) {
	path := t.TempDir()
	root, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	assertEmpty := func() {
		t.Helper()
		entries, err := os.ReadDir(path)
		if err != nil || len(entries) != 0 {
			t.Fatalf("named staging data: %v %v", entries, err)
		}
	}
	var retained io.Writer
	reader, err := root.Stage(context.Background(), 6, func(w io.Writer) error {
		retained = w
		assertEmpty()
		_, err := io.WriteString(w, "secret")
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reader.(io.Writer); ok {
		t.Fatal("read capability exposes writing")
	}
	if _, err := retained.Write([]byte("x")); err == nil {
		t.Fatal("retained writer mutated completed stage")
	}
	body, err := io.ReadAll(reader)
	if err != nil || string(body) != "secret" {
		t.Fatalf("stage: %q %v", body, err)
	}
	assertEmpty()
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := reader.Read(make([]byte, 1)); err == nil {
		t.Fatal("closed stage readable")
	}
	assertEmpty()
	for _, test := range []struct {
		name  string
		max   int64
		write WriteFunc
	}{
		{"bound", 1, func(w io.Writer) error { _, _ = w.Write([]byte("too big")); return nil }},
		{"failure", 10, func(w io.Writer) error { _, _ = w.Write([]byte("secret")); return errors.New("rejected") }},
		{"invalid bound", 0, func(io.Writer) error { t.Fatal("invalid bound invoked callback"); return nil }},
	} {
		t.Run(test.name, func(t *testing.T) {
			if reader, err := root.Stage(context.Background(), test.max, test.write); reader != nil || err == nil {
				t.Fatalf("failed stage exposed: %v %v", reader, err)
			}
			assertEmpty()
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	if reader, err := root.Stage(ctx, 10, func(w io.Writer) error {
		_, _ = w.Write([]byte("secret"))
		cancel()
		return nil
	}); reader != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v %v", reader, err)
	}
	assertEmpty()
	if reader, err := root.Stage(ctx, 10, func(io.Writer) error { t.Fatal("canceled callback"); return nil }); reader != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("pre-cancellation: %v %v", reader, err)
	}

	base := t.TempDir()
	admitted := filepath.Join(base, "admitted")
	if err := os.Mkdir(admitted, 0o700); err != nil {
		t.Fatal(err)
	}
	replaced, err := Open(admitted)
	if err != nil {
		t.Fatal(err)
	}
	defer replaced.Close()
	if reader, err := replaced.Stage(context.Background(), 10, func(w io.Writer) error {
		if err := os.Rename(admitted, filepath.Join(base, "moved")); err != nil {
			return err
		}
		if err := os.Mkdir(admitted, 0o700); err != nil {
			return err
		}
		_, err := w.Write([]byte("secret"))
		return err
	}); reader != nil || err == nil {
		t.Fatalf("root replacement admitted: %v %v", reader, err)
	}
	for _, name := range []string{"admitted", "moved"} {
		entries, err := os.ReadDir(filepath.Join(base, name))
		if err != nil || len(entries) != 0 {
			t.Fatalf("replacement leaked staging: %v %v", entries, err)
		}
	}
}
