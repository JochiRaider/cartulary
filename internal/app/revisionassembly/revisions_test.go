package revisionassembly

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type revisionsCompositionTestDB struct{}

func (revisionsCompositionTestDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("SELECT 0"), nil
}

func (revisionsCompositionTestDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (revisionsCompositionTestDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

func (revisionsCompositionTestDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, nil
}

type revisionsCompositionTestAttribution struct{}

func (revisionsCompositionTestAttribution) ResolveImportedSourceActorsTx(context.Context, pgx.Tx, uuid.UUID, string, string, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

type revisionsCompositionTestProjection struct{}

func (revisionsCompositionTestProjection) RebuildIncidentTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (revisionsCompositionTestProjection) Supports(string) bool {
	return false
}

func (revisionsCompositionTestProjection) LoadRowTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error) {
	return nil, pgx.ErrNoRows
}

func TestRevisionsCompositionRegistersEveryRequiredProvider(t *testing.T) {
	t.Parallel()
	service, err := NewCommandService(revisionsCompositionTestDB{}, revisionsCompositionTestAttribution{}, revisionsCompositionTestProjection{})
	if err != nil {
		t.Fatalf("compose revisions command service: %v", err)
	}
	if service == nil {
		t.Fatal("compose revisions command service returned nil")
	}
}
