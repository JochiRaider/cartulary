package revisions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type testDeleteRestoreSource struct{}

func (testDeleteRestoreSource) SnapshotTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error) {
	return map[string]any{}, nil
}

func (testDeleteRestoreSource) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (testDeleteRestoreSource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.test.v1", nil
}

func (testDeleteRestoreSource) ValidateDeletePreconditionsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (string, bool, error) {
	return "", false, nil
}

func TestDeleteRestoreSourceCatalogFailsClosed(t *testing.T) {
	t.Parallel()
	source := testDeleteRestoreSource{}
	if _, err := NewDeleteRestoreSourceCatalog([]string{"host", "party"},
		DeleteRestoreSourceRegistration{RecordType: "host", Source: source},
	); !errors.Is(err, ErrMissingDeleteRestoreSource) {
		t.Fatalf("missing source error = %v", err)
	}
	if _, err := NewDeleteRestoreSourceCatalog([]string{"host"},
		DeleteRestoreSourceRegistration{RecordType: "host", Source: source},
		DeleteRestoreSourceRegistration{RecordType: "host", Source: source},
	); !errors.Is(err, ErrDuplicateDeleteRestoreSource) {
		t.Fatalf("duplicate source error = %v", err)
	}
	if _, err := NewDeleteRestoreSourceCatalog([]string{"host"},
		DeleteRestoreSourceRegistration{RecordType: "party", Source: source},
	); !errors.Is(err, ErrUnexpectedDeleteRestoreSource) {
		t.Fatalf("unexpected source error = %v", err)
	}
	var typedNil *testDeleteRestoreSource
	if _, err := NewDeleteRestoreSourceCatalog([]string{"host"},
		DeleteRestoreSourceRegistration{RecordType: "host", Source: typedNil},
	); !errors.Is(err, ErrMissingDeleteRestoreSource) {
		t.Fatalf("typed nil source error = %v", err)
	}
}
