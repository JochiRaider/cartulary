package referencetest

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

func BlockedPortableCleanup(t testing.TB, lifetime *reference_data.VerificationLifetime, sourceClose bool) {
	t.Helper()
	f := NewPortable(t)
	request := f.Request()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var storage reference_data.ArtifactStorage = f.Root
	if sourceClose {
		request.OpenContainer = func(context.Context, string) (io.ReadCloser, error) {
			return blockedPortableReader{Reader: bytes.NewReader(f.Pack.Container), cancel: cancel}, nil
		}
	} else {
		root, err := rootedfs.Open(f.Temporary)
		if err != nil {
			t.Fatal(err)
		}
		defer root.Close()
		var once sync.Once
		var lease io.Closer
		defer func() {
			if lease != nil {
				_ = lease.Close()
			}
		}()
		storage = blockedStorage{ArtifactStorage: f.Root, afterWrite: func() {
			once.Do(func() {
				var err error
				lease, err = root.LockDirectory(context.Background(), rootedfs.MustParseReference("reference-packs/verification"))
				if err != nil {
					panic(err)
				}
				cancel()
			})
		}}
	}
	service := f.Service(storage, lifetime)
	err := service.RunImportExecution(ctx, request.Execution, request.ExecutionStarted, func(run context.Context, scope *reference_data.IncidentReferenceExecution) error {
		request.ExecutionScope = scope
		_, err := service.PrepareImport(run, request)
		return err
	})
	t.Fatalf("blocked portable cleanup returned instead of process exit: %v", err)
}

type blockedPortableReader struct {
	io.Reader
	cancel context.CancelFunc
}

func (r blockedPortableReader) Close() error { r.cancel(); select {} }
