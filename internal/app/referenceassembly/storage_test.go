package referenceassembly_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/google/uuid"
)

func testWorkspaceRecovery(t *testing.T) {
	ctx := context.Background()
	temporary, published := t.TempDir(), t.TempDir()
	first, err := referenceassembly.NewRootStorage(temporary, published)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	live, err := first.NewWorkspace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	if err := live.Write(ctx, "payload/entries.ndjson", func(w io.Writer) error { _, err := io.WriteString(w, "live"); return err }); err != nil {
		t.Fatal(err)
	}
	staged, _, _, err := first.StageStream(ctx, strings.NewReader("staged"), 6)
	if err != nil {
		t.Fatal(err)
	}
	defer first.RemoveStaged(staged)
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte("authoritative")))
	retained, retainedLease, err := first.PublishStream(ctx, digest, 13, strings.NewReader("authoritative"))
	if err != nil {
		t.Fatal(err)
	}
	defer first.RemovePublished(retained)
	defer retainedLease.Close()

	parent := filepath.Join(temporary, "reference-packs", "verification")
	abandoned := filepath.Join(parent, uuid.NewString())
	if err := os.Mkdir(abandoned, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{strings.Repeat("a", 64), ".cartulary-tmp-" + strings.Repeat("b", 32)} {
		if err := os.WriteFile(filepath.Join(abandoned, name), []byte("abandoned"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	second, err := referenceassembly.NewRootStorage(temporary, published)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	blocked, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	err = second.WithPublishedCollection(blocked, func(reference_data.PublishedObjects) error { t.Error("live publisher collected"); return nil })
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("publisher lease did not block collection", err)
	}
	if err := retainedLease.Close(); err != nil {
		t.Fatal(err)
	}
	if err := second.WithPublishedCollection(ctx, func(objects reference_data.PublishedObjects) error {
		seen := false
		err := objects.Visit(ctx, func(ref reference_data.StorageRef) error {
			if ref == retained {
				seen = true
			}
			return nil
		})
		if !seen {
			t.Error("published inventory lost object")
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}

	for range 2 {
		if err := second.ReconcileWorkspaces(ctx); err != nil {
			t.Fatal(err)
		}
		reader, err := live.Open(ctx, "payload/entries.ndjson")
		if err != nil {
			t.Fatal("live workspace collected", err)
		}
		data, err := io.ReadAll(reader)
		closeErr := reader.Close()
		if err != nil || closeErr != nil || string(data) != "live" {
			t.Fatal("live bytes changed", err, closeErr)
		}
	}
	if _, err := os.Stat(abandoned); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("abandoned workspace retained", err)
	}
	for _, open := range []func() (reference_data.ContainerReader, int64, error){
		func() (reference_data.ContainerReader, int64, error) { return second.OpenStaged(ctx, staged) },
		func() (reference_data.ContainerReader, int64, error) { return second.OpenPublished(ctx, retained) },
	} {
		reader, _, err := open()
		if err != nil {
			t.Fatal("scratch sweep crossed namespace", err)
		}
		if err := reader.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := live.Close(); err != nil {
		t.Fatal(err)
	}
	if err := second.ReconcileWorkspaces(ctx); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 0 {
		t.Fatal("scratch remains after completion", entries, err)
	}
	unsafe := filepath.Join(parent, uuid.NewString())
	if err := os.Mkdir(unsafe, 0700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "retained")
	if err := os.WriteFile(outside, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(unsafe, strings.Repeat("c", 64))); err != nil {
		t.Fatal(err)
	}
	if err := second.ReconcileWorkspaces(ctx); err == nil {
		t.Fatal("unsafe scratch admitted")
	}
	if data, err := os.ReadFile(outside); err != nil || string(data) != "keep" {
		t.Fatal("sweep escaped root", err)
	}
}

func TestReferencePackStreamingWorkspaceIsolationAndCleanup_Unit(t *testing.T) {
	t.Run("restart sweep preserves live work across capabilities", testWorkspaceRecovery)
	t.Run("operator incoming boundary", func(t *testing.T) {
		root := t.TempDir()
		storage, err := referenceassembly.NewRootStorage(t.TempDir(), root)
		if err != nil {
			t.Fatal(err)
		}
		defer storage.Close()
		if err := os.Mkdir(filepath.Join(root, "incoming"), 0700); err != nil {
			t.Fatal(err)
		}
		name := strings.Repeat("a", 128)
		if err := os.WriteFile(filepath.Join(root, "incoming", name), []byte("admitted"), 0600); err != nil {
			t.Fatal(err)
		}
		reader, err := storage.OpenIncoming(context.Background(), name)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil || string(data) != "admitted" {
			t.Fatal("incoming byte mismatch")
		}
		if err := os.Symlink(name, filepath.Join(root, "incoming", "link")); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(root, "incoming", "directory"), 0700); err != nil {
			t.Fatal(err)
		}
		for _, invalid := range []string{"", ".", "..", "../" + name, "a/b", "a\\b", "a\x00b", "é", strings.Repeat("a", 129), "link", "directory"} {
			if r, err := storage.OpenIncoming(context.Background(), invalid); err == nil {
				r.Close()
				t.Fatalf("unsafe incoming admitted %q", invalid)
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := storage.OpenIncoming(ctx, name); !errors.Is(err, context.Canceled) {
			t.Fatal("canceled incoming read")
		}
	})
	temporary := filepath.Join(t.TempDir(), "temporary")
	storage, err := referenceassembly.NewRootStorage(temporary, filepath.Join(t.TempDir(), "published"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	ctx := context.Background()
	a, err := storage.NewWorkspace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := storage.NewWorkspace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	write := func(value string) func(io.Writer) error {
		return func(w io.Writer) error { _, err := io.WriteString(w, value); return err }
	}
	if err := a.Write(ctx, "payload/entries.ndjson", write("a\n")); err != nil {
		t.Fatal(err)
	}
	if err := b.Write(ctx, "payload/entries.ndjson", write("b\n")); err != nil {
		t.Fatal(err)
	}
	if err := a.Write(ctx, "payload/entries.ndjson", write("overwritten")); err == nil {
		t.Fatal("duplicate workspace member replaced")
	}
	if err := a.Write(ctx, "../escape", write("escape")); err == nil {
		t.Fatal("unsafe name admitted")
	}
	reader, err := a.Open(ctx, "payload/entries.ndjson")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(reader)
	reader.Close()
	if err != nil || string(data) != "a\n" {
		t.Fatal("workspace content changed")
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Open(ctx, "payload/entries.ndjson"); err == nil {
		t.Fatal("closed workspace reopened")
	}
	reader, err = b.Open(ctx, "payload/entries.ndjson")
	if err != nil {
		t.Fatal(err)
	}
	data, err = io.ReadAll(reader)
	reader.Close()
	if err != nil || string(data) != "b\n" {
		t.Fatal("cleanup crossed workspace boundary")
	}
	staged, _, _, err := storage.StageStream(ctx, strings.NewReader("0123456789"), int64(len("0123456789")))
	if err != nil {
		t.Fatal(err)
	}
	container, size, err := storage.OpenStaged(ctx, staged)
	if err != nil || size != 10 {
		t.Fatal("streaming container open failed")
	}
	buf := make([]byte, 3)
	if _, err := container.ReadAt(buf, 5); err != nil || string(buf) != "567" {
		t.Fatal("random-access container read failed")
	}
	container.Close()
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(temporary, "reference-packs", "verification"))
	if err != nil || len(entries) != 0 {
		t.Fatal("workspace cleanup left extracted files or directories")
	}
}

func TestReferencePackRootStorageEnforcesReferencesAndLifecycle_Unit(t *testing.T) {
	temporaryRoot := filepath.Join(t.TempDir(), "temporary")
	publishedRoot := filepath.Join(t.TempDir(), "reference-packs")
	storage, err := referenceassembly.NewRootStorage(temporaryRoot, publishedRoot)
	if err != nil {
		t.Fatalf("NewRootStorage: %v", err)
	}
	t.Cleanup(storage.Close)

	staged, _, _, err := storage.StageStream(context.Background(), strings.NewReader("staged pack"), int64(len("staged pack")))
	if err != nil {
		t.Fatalf("Stage: %v", err)
	}
	if filepath.IsAbs(staged.String()) || strings.Contains(staged.String(), temporaryRoot) {
		t.Fatalf("Stage returned host path %q", staged.String())
	}
	stagedPath := filepath.Join(temporaryRoot, filepath.FromSlash(staged.String()))
	assertPrivateRegularFile(t, stagedPath, []byte("staged pack"))
	reader, size, err := storage.OpenStaged(context.Background(), staged)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(io.NewSectionReader(reader, 0, size))
	reader.Close()
	if err != nil || string(data) != "staged pack" {
		t.Fatalf("OpenStaged: %q %v", data, err)
	}
	if _, _, _, err := storage.StageStream(context.Background(), strings.NewReader("staged pack"), int64(len("staged pack")-1)); err == nil {
		t.Fatal("over-limit stream admitted")
	}
	if err := storage.RemoveStaged(staged); err != nil {
		t.Fatalf("RemoveStaged: %v", err)
	}
	if err := storage.RemoveStaged(staged); err != nil {
		t.Fatalf("RemoveStaged idempotent retry: %v", err)
	}

	if _, _, err := storage.PublishStream(context.Background(), "invalid", 7, strings.NewReader("invalid")); err == nil {
		t.Fatal("Publish accepted an invalid bundle digest")
	}
	bundleSHA := fmt.Sprintf("%x", sha256.Sum256([]byte("published pack")))
	published, publishedLease, err := storage.PublishStream(context.Background(), bundleSHA, int64(len("published pack")), strings.NewReader("published pack"))
	if err == nil {
		defer publishedLease.Close()
	}
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if filepath.IsAbs(published.String()) || strings.Contains(published.String(), publishedRoot) {
		t.Fatalf("Publish returned host path %q", published.String())
	}
	publishedPath := filepath.Join(publishedRoot, filepath.FromSlash(published.String()))
	assertPrivateRegularFile(t, publishedPath, []byte("published pack"))
	reader, size, err = storage.OpenPublished(context.Background(), published)
	if err != nil {
		t.Fatal(err)
	}
	data, err = io.ReadAll(io.NewSectionReader(reader, 0, size))
	reader.Close()
	if err != nil || string(data) != "published pack" {
		t.Fatalf("OpenPublished: %q %v", data, err)
	}
	if _, _, err := storage.PublishStream(context.Background(), bundleSHA, int64(len("published pack")), strings.NewReader("different pack")); err == nil {
		t.Fatal("digest mismatch admitted")
	}
	second, secondLease, err := storage.PublishStream(context.Background(), bundleSHA, int64(len("published pack")), strings.NewReader("published pack"))
	if err == nil {
		defer secondLease.Close()
	}
	if err != nil {
		t.Fatalf("second Publish: %v", err)
	}
	if second.String() == published.String() {
		t.Fatal("separate publications reused a mutable destination")
	}
	assertPrivateRegularFile(t, publishedPath, []byte("published pack"))
	if err := storage.RemovePublished(second); err != nil {
		t.Fatalf("RemovePublished: %v", err)
	}
	t.Run("recovery preserves confined object identity", func(t *testing.T) {
		ctx := context.Background()
		source, err := referenceassembly.NewRecoveryStorage(temporaryRoot, publishedRoot, reference_data.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		defer source.Close()
		targetRoot := filepath.Join(t.TempDir(), "restored")
		target, err := referenceassembly.NewRecoveryStorage(filepath.Join(t.TempDir(), "scratch"), targetRoot, reference_data.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		defer target.Close()
		if err := target.RequireEmpty(ctx); err != nil {
			t.Fatal(err)
		}
		info, err := source.StatRecoveryObject(ctx, published.String())
		if err != nil || info.PlaintextBytes != int64(len("published pack")) {
			t.Fatal("backup stat", info, err)
		}
		r, err := source.OpenRecoveryObject(ctx, published.String())
		if err != nil {
			t.Fatal(err)
		}
		err = target.RestoreMember(ctx, published.String(), bundleSHA, info.PlaintextBytes, r)
		r.Close()
		if err != nil {
			t.Fatal("restore exact reference", err)
		}
		assertPrivateRegularFile(t, filepath.Join(targetRoot, published.String()), []byte("published pack"))
		if target.RequireEmpty(ctx) == nil {
			t.Fatal("nonempty root admitted")
		}
		if target.RestoreMember(ctx, published.String(), bundleSHA, info.PlaintextBytes, strings.NewReader("published pack")) == nil {
			t.Fatal("restored object overwritten")
		}
		for _, bad := range []string{"../escape", "objects/" + bundleSHA, published.String() + "/child"} {
			if target.RestoreMember(ctx, bad, bundleSHA, info.PlaintextBytes, strings.NewReader("published pack")) == nil {
				t.Fatal("invalid recovery reference accepted")
			}
		}
		if err := target.ResetVerificationTarget(ctx); err != nil {
			t.Fatal(err)
		}
		if err := target.ResetVerificationTarget(ctx); err != nil {
			t.Fatal("repeat cleanup", err)
		}
		if err := target.RequireEmpty(ctx); err != nil {
			t.Fatal(err)
		}
		if target.RestoreMember(ctx, published.String(), bundleSHA, info.PlaintextBytes, strings.NewReader("tampered bytes")) == nil {
			t.Fatal("altered restored bytes admitted")
		}
		if err := target.RequireEmpty(ctx); err != nil {
			t.Fatal("failed restore left bytes", err)
		}
		assertPrivateRegularFile(t, publishedPath, []byte("published pack"))
	})
}

func TestReferencePackRootStorageFailsClosedOnCancellationSymlinkAndRootReplacement_Unit(t *testing.T) {
	t.Run("cancellation leaves no partial publication", func(t *testing.T) {
		storage, err := referenceassembly.NewRootStorage(
			filepath.Join(t.TempDir(), "temporary"),
			filepath.Join(t.TempDir(), "published"),
		)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(storage.Close)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, _, _, err := storage.StageStream(ctx, strings.NewReader("must not publish"), int64(len("must not publish"))); err == nil {
			t.Fatal("canceled Stage succeeded")
		}
	})

	t.Run("symlinked component", func(t *testing.T) {
		temporaryRoot := filepath.Join(t.TempDir(), "temporary")
		publishedRoot := filepath.Join(t.TempDir(), "published")
		outside := t.TempDir()
		for _, root := range []string{temporaryRoot, publishedRoot} {
			if err := os.Mkdir(root, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.Mkdir(filepath.Join(temporaryRoot, "reference-packs"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(temporaryRoot, "reference-packs", "imports")); err != nil {
			t.Fatal(err)
		}
		storage, err := referenceassembly.NewRootStorage(temporaryRoot, publishedRoot)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(storage.Close)
		if _, _, _, err := storage.StageStream(context.Background(), strings.NewReader("escape"), int64(len("escape"))); err == nil {
			t.Fatal("Stage followed a symlinked component")
		} else if strings.Contains(err.Error(), temporaryRoot) || strings.Contains(err.Error(), outside) {
			t.Fatalf("Stage error disclosed host root: %v", err)
		}
		entries, err := os.ReadDir(outside)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("symlink escape wrote outside root: %#v", entries)
		}
	})

	t.Run("root replacement", func(t *testing.T) {
		base := t.TempDir()
		temporaryRoot := filepath.Join(base, "temporary")
		originalRoot := filepath.Join(base, "temporary-original")
		publishedRoot := filepath.Join(base, "published")
		for _, root := range []string{temporaryRoot, publishedRoot} {
			if err := os.Mkdir(root, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		storage, err := referenceassembly.NewRootStorage(temporaryRoot, publishedRoot)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(storage.Close)
		if err := os.Rename(temporaryRoot, originalRoot); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(temporaryRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		if _, _, _, err := storage.StageStream(context.Background(), strings.NewReader("replacement"), int64(len("replacement"))); !errors.Is(err, rootedfs.ErrRootIdentityChanged) {
			t.Fatalf("Stage after root replacement error = %v; want ErrRootIdentityChanged", err)
		}
	})
}

func assertPrivateRegularFile(t testing.TB, path string, want []byte) {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		t.Fatalf("%s mode = %v; want private regular file", path, info.Mode())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(want) {
		t.Fatalf("%s bytes = %q want %q", path, data, want)
	}
}
