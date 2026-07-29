package application

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func TestRestoreRuntimeFailureClosesCreatedResourcesInReverseOrder_Unit(t *testing.T) {
	var closed []string
	source := Deployment{
		OpenPostgres: func(context.Context) (PostgresPool, error) {
			return &lifetimePool{name: "source-postgres", closed: &closed}, nil
		},
		OpenObjectStore: func(context.Context) (objectstore.Store, error) {
			return &lifetimeObjectStore{name: "source-object", closed: &closed}, nil
		},
		OpenBackup: func() (recovery.BackupStorage, error) {
			return nil, errors.New("injected backup open failure")
		},
	}
	target := Deployment{
		OpenPostgres: func(context.Context) (PostgresPool, error) {
			return &lifetimePool{name: "target-postgres", closed: &closed}, nil
		},
		OpenObjectStore: func(context.Context) (objectstore.Store, error) {
			return &lifetimeObjectStore{name: "target-object", closed: &closed}, nil
		},
	}
	service := Service{LoadDeployment: func(path string) (Deployment, error) {
		if path == "source" {
			return source, nil
		}
		return target, nil
	}}

	_, _, _, _, _, _, _, err := service.openRestoreRuntime(context.Background(), operationRequest{
		SourceConfigPath: "source",
		TargetConfigPath: "target",
	})
	if err == nil {
		t.Fatal("restore runtime accepted injected backup open failure")
	}
	want := []string{"target-object", "source-object", "target-postgres", "source-postgres"}
	if !reflect.DeepEqual(closed, want) {
		t.Fatalf("restore runtime close order got %v want %v", closed, want)
	}
}

func TestSourceRuntimeFailureClosesCreatedPostgres_Unit(t *testing.T) {
	var closed []string
	service := Service{LoadDeployment: func(string) (Deployment, error) {
		return Deployment{
			OpenPostgres: func(context.Context) (PostgresPool, error) {
				return &lifetimePool{name: "source-postgres", closed: &closed}, nil
			},
			OpenBackup: func() (recovery.BackupStorage, error) {
				return nil, errors.New("injected backup open failure")
			},
		}, nil
	}}

	_, _, _, _, err := service.openSourceRuntime(context.Background(), "source")
	if err == nil {
		t.Fatal("source runtime accepted injected backup open failure")
	}
	if want := []string{"source-postgres"}; !reflect.DeepEqual(closed, want) {
		t.Fatalf("source runtime close order got %v want %v", closed, want)
	}
}

type lifetimePool struct {
	name   string
	closed *[]string
}

func (pool *lifetimePool) Close() {
	*pool.closed = append(*pool.closed, pool.name)
}

func (*lifetimePool) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (*lifetimePool) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (*lifetimePool) QueryRow(context.Context, string, ...any) pgx.Row {
	return lifetimeRow{}
}

func (*lifetimePool) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, errors.New("unexpected BeginTx")
}

type lifetimeRow struct{}

func (lifetimeRow) Scan(...any) error {
	return errors.New("unexpected Scan")
}

type lifetimeObjectStore struct {
	name   string
	closed *[]string
}

func (*lifetimeObjectStore) UploadTarget(context.Context, string, time.Time) (objectstore.UploadTarget, error) {
	return objectstore.UploadTarget{}, errors.New("unexpected UploadTarget")
}

func (*lifetimeObjectStore) CompleteUploadTarget(context.Context, string, io.Reader, string) error {
	return errors.New("unexpected CompleteUploadTarget")
}

func (*lifetimeObjectStore) PutObject(context.Context, string, io.Reader, int64, string) error {
	return errors.New("unexpected PutObject")
}

func (*lifetimeObjectStore) ReadObject(context.Context, string, objectstore.ReadOptions) (io.ReadCloser, objectstore.ObjectInfo, error) {
	return nil, objectstore.ObjectInfo{}, errors.New("unexpected ReadObject")
}

func (*lifetimeObjectStore) StatObject(context.Context, string) (objectstore.ObjectInfo, error) {
	return objectstore.ObjectInfo{}, errors.New("unexpected StatObject")
}

func (*lifetimeObjectStore) ListObjects(context.Context, string) ([]objectstore.ObjectInfo, error) {
	return nil, errors.New("unexpected ListObjects")
}

func (*lifetimeObjectStore) DeleteObject(context.Context, string) error {
	return errors.New("unexpected DeleteObject")
}

func (store *lifetimeObjectStore) Close() error {
	*store.closed = append(*store.closed, store.name)
	return nil
}
