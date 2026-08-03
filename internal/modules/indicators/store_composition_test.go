package indicators

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func TestIndicatorStoreCompositionAndRepositoryBoundaries(t *testing.T) {
	t.Parallel()

	if _, err := NewStore(StoreDependencies{}); err == nil || !strings.Contains(err.Error(), "Postgres is required") {
		t.Fatalf("missing Postgres dependency error = %v", err)
	}
	if _, err := NewStore(StoreDependencies{Postgres: inertIndicatorDB{}}); err == nil || !strings.Contains(err.Error(), "Revisions is required") {
		t.Fatalf("missing Revisions dependency error = %v", err)
	}
	owner, err := NewStore(StoreDependencies{
		Postgres:  inertIndicatorDB{},
		Revisions: &revisions.Appender{},
	})
	if err != nil || owner == nil {
		t.Fatalf("compose complete Indicators owner: owner=%#v err=%v", owner, err)
	}

	for _, path := range []string{
		"source_repository.go",
		"observation_repository.go",
		"lifecycle_repository.go",
	} {
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		content := string(body)
		for _, forbidden := range []string{
			"BeginTx(",
			".Commit(",
			".Rollback(",
			"AppendRevision",
			"projection",
			"idempotency",
		} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s owns forbidden workflow concern %q", path, forbidden)
			}
		}
	}
}

type inertIndicatorDB struct{}

func (inertIndicatorDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("unexpected Exec")
}

func (inertIndicatorDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected Query")
}

func (inertIndicatorDB) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow")
}

func (inertIndicatorDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	panic("unexpected BeginTx")
}
