package revisions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
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

func (testDeleteRestoreSource) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

func TestDeleteRestoreSourceCatalogFailsClosed(t *testing.T) {
	t.Parallel()
	source := testDeleteRestoreSource{}
	if _, err := newDeleteRestoreSourceCatalog([]string{"host", "party"},
		deleteRestoreSourceRegistration{RecordType: "host", Source: source},
	); !errors.Is(err, ErrMissingDeleteRestoreSource) {
		t.Fatalf("missing source error = %v", err)
	}
	if _, err := newDeleteRestoreSourceCatalog([]string{"host"},
		deleteRestoreSourceRegistration{RecordType: "host", Source: source},
		deleteRestoreSourceRegistration{RecordType: "host", Source: source},
	); !errors.Is(err, ErrDuplicateDeleteRestoreSource) {
		t.Fatalf("duplicate source error = %v", err)
	}
	if _, err := newDeleteRestoreSourceCatalog([]string{"host"},
		deleteRestoreSourceRegistration{RecordType: "party", Source: source},
	); !errors.Is(err, ErrUnexpectedDeleteRestoreSource) {
		t.Fatalf("unexpected source error = %v", err)
	}
	var typedNil *testDeleteRestoreSource
	if _, err := newDeleteRestoreSourceCatalog([]string{"host"},
		deleteRestoreSourceRegistration{RecordType: "host", Source: typedNil},
	); !errors.Is(err, ErrMissingDeleteRestoreSource) {
		t.Fatalf("typed nil source error = %v", err)
	}
}
