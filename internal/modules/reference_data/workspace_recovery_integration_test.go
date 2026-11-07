package reference_data_test

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
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	verificationtest "github.com/JochiRaider/cartulary/internal/testutil/referencetest"
	"github.com/jackc/pgx/v5/pgxpool"
)

type maintenanceObserver struct {
	t         *testing.T
	completed chan string
}

func (o maintenanceObserver) BeginReferenceOperation(ctx context.Context, operation string) (context.Context, func(string)) {
	if operation != "reference_pack.collection" {
		o.t.Error("unsafe maintenance observation", operation)
	}
	return ctx, func(outcome string) { o.completed <- outcome }
}

func TestReferenceWorkspaceAutomaticRecovery_Integration(t *testing.T) {
	ctx := t.Context()
	db := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "reference-workspace-recovery")
	pool, err := pgxpool.New(ctx, db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	temporary, published := t.TempDir(), t.TempDir()
	storage, err := referenceassembly.NewRootStorage(temporary, published)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	// Unclaimed optional administration must not disable required maintenance.
	if err := reference_data.ReconcileBaseRelease(ctx, pool, storage, reference_data.BaseReleaseOptions{Finalizer: consumerIntegrityOptions(t).Finalizer, Limits: reference_data.DefaultLimits()}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	failed, err := storage.NewWorkspace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := failed.Write(ctx, "payload", func(w io.Writer) error { _, err := io.WriteString(w, "private"); return err }); err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(temporary, "reference-packs", "verification")
	dirs, err := os.ReadDir(parent)
	if err != nil || len(dirs) != 1 {
		t.Fatal(dirs, err)
	}
	residue := filepath.Join(parent, dirs[0].Name())
	members, err := os.ReadDir(residue)
	if err != nil || len(members) != 1 {
		t.Fatal(members, err)
	}
	member := filepath.Join(residue, members[0].Name())
	if err := os.Remove(member); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(member, 0700); err != nil {
		t.Fatal(err)
	}
	closeErr := failed.Close(ctx)
	if closeErr == nil {
		t.Fatal("real confined deletion fault did not fail")
	}
	second, err := referenceassembly.NewRootStorage(temporary, published)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	live, err := second.NewWorkspace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close(ctx)
	if err := live.Write(ctx, "live", func(w io.Writer) error { _, err := io.WriteString(w, "keep"); return err }); err != nil {
		t.Fatal(err)
	}
	staged, _, _, err := second.StageStream(ctx, strings.NewReader("input"), 5)
	if err != nil {
		t.Fatal(err)
	}
	defer second.RemoveStaged(staged)
	orphan, lease, err := storage.PublishStream(ctx, fmt.Sprintf("%x", sha256.Sum256([]byte("orphan"))), 6, strings.NewReader("orphan"))
	if err != nil {
		t.Fatal(err)
	}
	if err := lease.Close(); err != nil {
		t.Fatal(err)
	}
	ticks := make(chan time.Time, 1)
	observed := make(chan string, 2)
	collector := reference_data.NewTestObjectCollector(t, pool, storage, storage, maintenanceObserver{t, observed}, ticks)
	serving := make(chan struct{})
	if err := collector.Prepare(ctx, serving); err != nil {
		t.Fatal(err)
	}
	defer collector.Close(context.Background())
	close(serving)
	sweep := func(want string) {
		t.Helper()
		ticks <- time.Now()
		select {
		case got := <-observed:
			if got != want {
				t.Fatal("maintenance outcome", got, want)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("maintenance did not finish")
		}
	}
	sweep("failed")
	if _, err := os.Stat(residue); err != nil {
		t.Fatal("unsafe residue not retained", err)
	}
	if _, err := os.Stat(filepath.Join(published, filepath.FromSlash(orphan.String()))); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("scratch failure skipped published phase", err)
	}
	if err := os.Remove(member); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(member, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	// Creation shares the collector's parent lease, including the ownership
	// handoff. Exercise this concurrently using a separate rooted instance.
	created := make(chan error, 1)
	go func() {
		for range 25 {
			w, err := second.NewWorkspace(ctx)
			if err != nil {
				created <- err
				return
			}
			err = w.Write(ctx, "concurrent", func(w io.Writer) error { _, err := io.WriteString(w, "keep"); return err })
			err = errors.Join(err, w.Close(ctx))
			if err != nil {
				created <- err
				return
			}
		}
		created <- nil
	}()
	sweep("success")
	if err := <-created; err != nil {
		t.Fatal("creation raced collection", err)
	}
	if _, err := os.Stat(residue); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("automatic loop retained abandoned scratch", err)
	}
	if err := failed.Close(ctx); err != closeErr {
		t.Fatal("released handle changed cached outcome", err)
	}
	reader, err := live.Open(ctx, "live")
	if err != nil {
		t.Fatal("live workspace collected", err)
	}
	data, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil || string(data) != "keep" {
		t.Fatal("live workspace changed", err)
	}
	input, _, err := storage.OpenStaged(ctx, staged)
	if err != nil {
		t.Fatal("input namespace changed", err)
	}
	_ = input.Close()
	if err := reference_data.ValidateRequiredState(ctx, pool, storage, reference_data.DefaultLimits(), verificationtest.Lifetime(t)); err != nil {
		t.Fatal("authoritative Base bytes changed", err)
	}
}
