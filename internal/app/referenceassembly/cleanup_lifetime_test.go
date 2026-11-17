package referenceassembly_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	verificationtest "github.com/JochiRaider/cartulary/internal/testutil/referencetest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type uncalledAdmission struct{}

func (uncalledAdmission) CreateQueuedTx(context.Context, pgx.Tx, jobs.EnqueueParams, time.Time) (jobs.Resource, error) {
	return jobs.Resource{}, errors.New("unexpected Job admission")
}

type interruptedStorage struct {
	reference_data.ArtifactStorage
	write    func()
	close    func()
	writeErr error
}

func (s interruptedStorage) NewWorkspace(ctx context.Context) (reference_data.VerificationWorkspace, error) {
	w, err := s.ArtifactStorage.NewWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	return interruptedWorkspace{VerificationWorkspace: w, write: s.write, close: s.close, writeErr: s.writeErr}, nil
}

type interruptedWorkspace struct {
	reference_data.VerificationWorkspace
	write    func()
	close    func()
	writeErr error
}

func (w interruptedWorkspace) Write(ctx context.Context, name string, fn func(io.Writer) error) error {
	err := w.VerificationWorkspace.Write(ctx, name, fn)
	if err == nil && w.write != nil {
		w.write()
	}
	if err == nil && w.writeErr != nil {
		return w.writeErr
	}
	return err
}
func (w interruptedWorkspace) Close(ctx context.Context) error {
	if w.close != nil {
		w.close()
	}
	return w.VerificationWorkspace.Close(ctx)
}

func TestImportProbeSupervisedRootedCleanup_Unit(t *testing.T) {
	raw, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/signed-container.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Container string `json:"container_base64"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	container, err := base64.StdEncoding.DecodeString(fixture.Container)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"success", "cancellation", "deadline", "write failure", "cancellation waiting for lease"} {
		run := func(t *testing.T) {
			started := time.Now()
			temporary := t.TempDir()
			root, err := referenceassembly.NewRootStorage(temporary, t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			storage := interruptedStorage{ArtifactStorage: root}
			wrote := false
			switch mode {
			case "cancellation":
				storage.write = cancel
			case "deadline":
				var deadlineCancel context.CancelFunc
				ctx, deadlineCancel = context.WithTimeout(ctx, 40*time.Millisecond)
				defer deadlineCancel()
				storage.write = func() { wrote = true; <-ctx.Done() }
			case "cancellation waiting for lease":
				leaseRoot, err := rootedfs.Open(temporary)
				if err != nil {
					t.Fatal(err)
				}
				defer leaseRoot.Close()
				storage.close = func() {
					lease, err := leaseRoot.LockDirectory(context.Background(), rootedfs.MustParseReference("reference-packs/verification"))
					if err != nil {
						panic(err)
					}
					go func() {
						time.Sleep(20 * time.Millisecond)
						cancel()
						time.Sleep(20 * time.Millisecond)
						_ = lease.Close()
					}()
				}
			case "write failure":
				storage.writeErr = io.ErrClosedPipe
			}
			application, err := reference_data.NewImportAdmission(reference_data.ImportAdmissionOptions{Lifetime: verificationtest.Lifetime(t), Postgres: new(pgxpool.Pool), Storage: storage, Limits: reference_data.DefaultLimits(), JobAdmission: uncalledAdmission{}, Now: time.Now})
			if err != nil {
				t.Fatal(err)
			}
			pending, err := application.PrepareImport(ctx, bytes.NewReader(container))
			if mode == "deadline" && (!wrote || !errors.Is(ctx.Err(), context.DeadlineExceeded)) {
				t.Fatal("deadline must interrupt an existing workspace after a write", wrote, ctx.Err())
			}
			// In fake time the lease is released at 40 ms; cleanup cannot finish earlier.
			if mode == "cancellation waiting for lease" && time.Since(started) < 40*time.Millisecond {
				t.Fatal("cleanup completed before the parent lease was released", time.Since(started))
			}
			if mode == "success" {
				if err != nil {
					t.Fatal(err)
				}
				if err := pending.Close(); err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("interrupted preparation succeeded")
			}
			entries, err := os.ReadDir(filepath.Join(temporary, "reference-packs", "verification"))
			if err != nil || len(entries) != 0 {
				t.Fatal("operation returned with scratch", entries, err)
			}
		}
		t.Run(mode, func(t *testing.T) {
			if mode == "deadline" || mode == "cancellation waiting for lease" {
				// Advance fake time only after every goroutine blocks: the
				// deadline fires after the write hook blocks, and the lease
				// cancellation and release land while cleanup polls for the
				// lease. Host scheduling cannot reorder either sequence.
				synctest.Test(t, run)
			} else {
				run(t)
			}
		})
	}
}
