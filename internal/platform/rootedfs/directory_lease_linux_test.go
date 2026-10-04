//go:build linux

package rootedfs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func testDirectoryLeases(t *testing.T) {
	if path := os.Getenv("CARTULARY_TEST_DIRECTORY_LEASE_CHILD"); path != "" {
		root, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		lease, ok, err := root.TryLockDirectory(MustParseReference("work"))
		if err != nil || !ok {
			t.Fatal("child lease", err)
		}
		defer lease.Close()
		if _, err := fmt.Fprintln(os.Stdout, "lease-ready"); err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, os.Stdin)
		os.Exit(0)
	}
	ctx := context.Background()
	path := t.TempDir()
	a, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	dir := MustParseReference("work")
	if err := a.MakePrivateDir(dir); err != nil {
		t.Fatal(err)
	}
	first, acquired, err := a.TryLockDirectory(dir)
	if err != nil || !acquired {
		t.Fatal("initial lease", err)
	}
	defer first.Close()
	flags, err := unix.FcntlInt(uintptr(first.(*directoryLease).fd), unix.F_GETFD, 0)
	if err != nil || flags&unix.FD_CLOEXEC == 0 {
		t.Fatal("directory lease can escape into an executed child", err)
	}
	if lock, acquired, err := b.TryLockDirectory(dir); err != nil || acquired || lock != nil {
		t.Fatal("second capability bypassed lease", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := b.LockDirectory(canceled, dir); !errors.Is(err, context.Canceled) {
		t.Fatal("lease wait ignored cancellation", err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal("lease close not idempotent", err)
	}
	sharedA, err := a.LockSharedDirectory(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	sharedB, err := b.LockSharedDirectory(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if lease, ok, err := a.TryLockDirectory(dir); err != nil || ok || lease != nil {
		t.Fatal("shared publisher lease bypassed", err)
	}
	if err := sharedA.Close(); err != nil {
		t.Fatal(err)
	}
	if lease, ok, err := a.TryLockDirectory(dir); err != nil || ok || lease != nil {
		t.Fatal("second shared lease lost", err)
	}
	if err := sharedB.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := b.LockDirectory(ctx, dir)
	if err != nil {
		t.Fatal("released lease unavailable", err)
	}
	defer second.Close()
	for i := range 300 {
		ref := MustParseReference(fmt.Sprintf("work/%03d", i))
		if err := a.CreateExclusive(ctx, ref, func(io.Writer) error { return nil }); err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	if err := a.VisitDirectory(ctx, dir, func(entry DirectoryEntry) error {
		if entry.Directory || seen[entry.Reference.String()] {
			t.Fatal("duplicate or wrong child", entry)
		}
		seen[entry.Reference.String()] = true
		return a.RemoveRegular(entry.Reference)
	}); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 300 {
		t.Fatal("streamed traversal lost entries", len(seen))
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(path, "link")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := a.TryLockDirectory(MustParseReference("link")); err == nil {
		t.Fatal("leased directory symlink")
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(path, "work", "link")); err != nil {
		t.Fatal(err)
	}
	if err := a.VisitDirectory(ctx, dir, func(DirectoryEntry) error { t.Fatal("unsafe child reached visitor"); return nil }); err == nil {
		t.Fatal("traversal admitted symlink")
	}
	if _, _, err := a.TryLockDirectory(Reference{}); !errors.Is(err, ErrInvalidReference) {
		t.Fatal("zero reference acquired root lease", err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
	t.Run("process death releases ownership", func(t *testing.T) {
		childCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		command := exec.CommandContext(childCtx, os.Args[0], "-test.run=^TestRootedFSOperationContainment_Unit$/^directory_ownership_and_bounded_traversal$")
		command.Env = append(os.Environ(), "CARTULARY_TEST_DIRECTORY_LEASE_CHILD="+path)
		output, err := command.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		input, err := command.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		defer input.Close()
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = command.Process.Kill(); _ = command.Wait() }()
		ready, err := bufio.NewReader(output).ReadString('\n')
		if err != nil || ready != "lease-ready\n" {
			t.Fatal("child did not establish lease", ready, err)
		}
		if lease, ok, err := a.TryLockDirectory(dir); err != nil || ok || lease != nil {
			t.Fatal("cross-process lease bypass", err)
		}
		if err := command.Process.Kill(); err != nil {
			t.Fatal(err)
		}
		if err := command.Wait(); err == nil {
			t.Fatal("child exited without interruption")
		}
		lease, ok, err := a.TryLockDirectory(dir)
		if err != nil || !ok {
			t.Fatal("dead process retained lease", err)
		}
		if err := lease.Close(); err != nil {
			t.Fatal(err)
		}
	})
}
