package referencetest

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BlockedImportCleanup enters real rooted deletion with its parent lease held.
// Only the supplied production fatal process sink can complete this scenario.
func BlockedImportCleanup(t testing.TB, lifetime *reference_data.VerificationLifetime) {
	t.Helper()
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
	data, err := base64.StdEncoding.DecodeString(fixture.Container)
	if err != nil {
		t.Fatal(err)
	}
	temporary := t.TempDir()
	storage, err := referenceassembly.NewRootStorage(temporary, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	root, err := rootedfs.Open(temporary)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var once sync.Once
	var lease io.Closer
	defer func() {
		if lease != nil {
			_ = lease.Close()
		}
	}()
	blocked := blockedStorage{ArtifactStorage: storage, afterWrite: func() {
		once.Do(func() {
			var err error
			lease, err = root.LockDirectory(context.Background(), rootedfs.MustParseReference("reference-packs/verification"))
			if err != nil {
				panic(err)
			}
			cancel()
		})
	}}
	admission, err := reference_data.NewImportAdmission(reference_data.ImportAdmissionOptions{Lifetime: lifetime, Postgres: new(pgxpool.Pool), Storage: blocked, Limits: reference_data.DefaultLimits(), JobAdmission: noAdmission{}, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	_, err = admission.PrepareImport(ctx, bytes.NewReader(data))
	t.Fatalf("blocked cleanup returned instead of terminating process: %v", err)
}

type noAdmission struct{}

func (noAdmission) CreateQueuedTx(context.Context, pgx.Tx, jobs.EnqueueParams, time.Time) (jobs.Resource, error) {
	return jobs.Resource{}, errors.New("unexpected admission")
}

type blockedStorage struct {
	reference_data.ArtifactStorage
	afterWrite func()
}

func (s blockedStorage) NewWorkspace(ctx context.Context) (reference_data.VerificationWorkspace, error) {
	w, err := s.ArtifactStorage.NewWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	return blockedWorkspace{w, s.afterWrite}, nil
}

type blockedWorkspace struct {
	reference_data.VerificationWorkspace
	afterWrite func()
}

func (w blockedWorkspace) Write(ctx context.Context, name string, write func(io.Writer) error) error {
	err := w.VerificationWorkspace.Write(ctx, name, write)
	if err == nil {
		w.afterWrite()
	}
	return err
}
