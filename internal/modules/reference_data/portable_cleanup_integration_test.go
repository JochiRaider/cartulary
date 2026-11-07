package reference_data_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/testutil/referencetest"
)

type portableCleanupStorage struct {
	reference_data.ArtifactStorage
	afterWrite func()
	writeErr   error
}

func (s portableCleanupStorage) NewWorkspace(ctx context.Context) (reference_data.VerificationWorkspace, error) {
	w, err := s.ArtifactStorage.NewWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	return portableCleanupWorkspace{w, s.afterWrite, s.writeErr}, nil
}

type portableCleanupWorkspace struct {
	reference_data.VerificationWorkspace
	afterWrite func()
	writeErr   error
}

func (w portableCleanupWorkspace) Write(ctx context.Context, name string, write func(io.Writer) error) error {
	err := w.VerificationWorkspace.Write(ctx, name, write)
	if err == nil && w.afterWrite != nil {
		w.afterWrite()
	}
	if err == nil && w.writeErr != nil {
		return w.writeErr
	}
	return err
}

type portableClosingReader struct {
	io.Reader
	close func() error
}

func (r portableClosingReader) Close() error { return r.close() }

func TestPortableRootedCompletion_Integration(t *testing.T) {
	t.Run("two containers share disposal grace", testPortableDisposalGrace)
	for _, mode := range []string{"success", "optional rejection", "required rejection", "required rejection with cleanup failure", "cancel", "parent timeout", "owner timeout", "partial write failure", "source close failure"} {
		t.Run(mode, func(t *testing.T) {
			f := referencetest.NewPortable(t)
			request := f.Request()
			if mode == "owner timeout" {
				// The parent's earlier work consumes the same 60-second budget.
				// Expire it before this isolated fixture's unrenewed Job lease.
				request.ExecutionStarted = time.Now().Add(-58 * time.Second)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			storage := portableCleanupStorage{ArtifactStorage: f.Root}
			var once sync.Once
			writes := 0
			storage.afterWrite = func() {
				writes++
				once.Do(func() {
					switch mode {
					case "cancel":
						cancel()
					case "parent timeout", "owner timeout":
						<-ctx.Done()
					}
				})
			}
			if mode == "partial write failure" {
				storage.writeErr = io.ErrClosedPipe
			}
			if mode == "parent timeout" {
				var stop context.CancelFunc
				ctx, stop = context.WithTimeout(ctx, 500*time.Millisecond)
				defer stop()
			}
			if mode == "optional rejection" || mode == "required rejection" || mode == "required rejection with cleanup failure" {
				f.At = f.At.Add(2 * 365 * 24 * time.Hour)
			}
			if mode == "optional rejection" {
				var content map[string]json.RawMessage
				if err := json.Unmarshal(*request.ContentManifest, &content); err != nil {
					t.Fatal(err)
				}
				content["required_members"] = json.RawMessage("[]")
				data, err := json.Marshal(content)
				if err != nil {
					t.Fatal(err)
				}
				request.ContentManifest = &data
			}
			closed := 0
			request.OpenContainer = func(context.Context, string) (io.ReadCloser, error) {
				return portableClosingReader{bytes.NewReader(f.Pack.Container), func() error {
					closed++
					if mode == "source close failure" {
						return io.ErrClosedPipe
					}
					return nil
				}}, nil
			}
			service := f.Service(storage, referencetest.Lifetime(t))
			err := service.RunImportExecution(ctx, request.Execution, request.ExecutionStarted, func(run context.Context, scope *reference_data.IncidentReferenceExecution) error {
				// The owner budget cancels the callback context, not the outer test context.
				ctx = run
				request.ExecutionScope = scope
				prepared, err := service.PrepareImport(run, request)
				switch mode {
				case "owner timeout", "parent timeout":
					if !errors.Is(context.Cause(run), context.DeadlineExceeded) {
						t.Error("timeout replaced by another interruption", context.Cause(run))
					}
				case "cancel":
					if !errors.Is(context.Cause(run), context.Canceled) {
						t.Error("cancellation replaced by another interruption", context.Cause(run))
					}
				}
				if err != nil && prepared != nil {
					t.Error("failed preparation returned content")
				}
				if mode == "required rejection with cleanup failure" {
					var rejected *reference_data.IncidentBundleReferenceValidationError
					if !errors.As(err, &rejected) {
						t.Error("required rejection was not exercised", err)
					}
					cleanupErr := scope.CompleteCleanup(func(context.Context) error { return io.ErrClosedPipe })
					outcome, _, _, classifyErr := scope.ClassifyAbsent(context.Background(), cleanupErr)
					if classifyErr != nil || outcome != "execution_failed" {
						t.Error("cleanup failure authorized content verdict", outcome, classifyErr)
					}
					return cleanupErr
				}
				return err
			})
			succeeds := mode == "success" || mode == "optional rejection"
			if succeeds && err != nil {
				t.Fatal(err)
			}
			if !succeeds && err == nil {
				t.Fatal("interrupted/rejected preparation succeeded")
			}
			if mode == "required rejection" {
				var rejected *reference_data.IncidentBundleReferenceValidationError
				if !errors.As(err, &rejected) {
					t.Fatal("missing required verdict", err)
				}
			}
			if writes == 0 || closed != 1 {
				t.Fatal("real portable resources were not exercised", writes, closed)
			}
			entries, readErr := os.ReadDir(filepath.Join(f.Temporary, "reference-packs", "verification"))
			if readErr != nil || len(entries) != 0 {
				t.Fatal("returned with scratch", entries, readErr)
			}
			if mode == "cancel" || mode == "parent timeout" || mode == "owner timeout" || mode == "partial write failure" || mode == "source close failure" {
				var count int
				if err := f.Pool.QueryRow(t.Context(), `SELECT count(*) FROM reference_pack_operations WHERE job_id=$1`, request.Execution.JobID()).Scan(&count); err != nil || count != 0 {
					t.Fatal("failed probe admitted an operation", count, err)
				}
			}
		})
	}
}

type portableDisposalStorage struct {
	reference_data.ArtifactStorage
	beforeRemove func()
}

func (s portableDisposalStorage) RemovePublished(ref reference_data.StorageRef) error {
	s.beforeRemove()
	return s.ArtifactStorage.RemovePublished(ref)
}

func testPortableDisposalGrace(t *testing.T) {
	f := referencetest.NewPortable(t)
	request := f.CohortRequest()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	fatal := make(chan error, 1)
	lifetime, err := reference_data.NewVerificationLifetime(100*time.Millisecond, func(err error) { fatal <- err })
	if err != nil {
		t.Fatal(err)
	}
	open := request.OpenContainer
	closed := 0
	request.OpenContainer = func(ctx context.Context, path string) (io.ReadCloser, error) {
		reader, err := open(ctx, path)
		if err != nil {
			return nil, err
		}
		return portableClosingReader{reader, func() error {
			closed++
			if err := reader.Close(); err != nil {
				return err
			}
			if closed == 2 {
				cancel()
				return io.ErrClosedPipe
			}
			return nil
		}}, nil
	}
	release := make(chan struct{})
	removed := 0
	storage := portableDisposalStorage{f.Root, func() {
		removed++
		if removed == 1 {
			// Both container disposals belong to the original operation. A late
			// successful deletion must not clear exhaustion before the next one.
			<-release
		}
	}}
	service := f.Service(storage, lifetime)
	done := make(chan error, 1)
	go func() {
		done <- service.RunImportExecution(ctx, request.Execution, request.ExecutionStarted, func(run context.Context, scope *reference_data.IncidentReferenceExecution) error {
			request.ExecutionScope = scope
			prepared, err := service.PrepareImport(run, request)
			if prepared != nil || !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
				t.Error("disposal cleared fatal", prepared != nil, err)
			}
			// This represents the later source-owner/parent completion stage.
			if err := scope.CompleteCleanup(func(context.Context) error { return nil }); !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
				t.Error("later stage renewed grace", err)
			}
			return err
		})
	}()
	select {
	case <-fatal:
	case err := <-done:
		t.Fatal("finished before blocked disposal", err)
	case <-time.After(10 * time.Second):
		t.Fatal("missing fatal during disposal")
	}
	select {
	case <-done:
		t.Fatal("returned with disposal active")
	default:
	}
	close(release)
	if err := <-done; !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
		t.Fatal(err)
	}
	if closed != 2 || removed != 2 {
		t.Fatal("cohort resources not completed", closed, removed)
	}
	entries, err := os.ReadDir(filepath.Join(f.Temporary, "reference-packs", "verification"))
	if err != nil || len(entries) != 0 {
		t.Fatal("scratch escaped disposal", entries, err)
	}
	var n int
	if err := f.Pool.QueryRow(t.Context(), `SELECT count(*) FROM reference_pack_operations WHERE job_id=$1`, request.Execution.JobID()).Scan(&n); err != nil || n != 0 {
		t.Fatal("fatal probe admitted an operation", n, err)
	}
}
