package revisions

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	recorddeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"
)

type commandServiceTestDB struct{}

func (commandServiceTestDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("SELECT 0"), nil
}

func (commandServiceTestDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (commandServiceTestDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

func (commandServiceTestDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, nil
}

type commandServiceTestProjection struct{}

func (commandServiceTestProjection) RebuildIncidentTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func TestCommandServiceRequiresEveryExplicitDependency(t *testing.T) {
	t.Parallel()
	dependencies := validCommandServiceDependencies(t)
	if _, err := NewCommandService(dependencies); err != nil {
		t.Fatalf("complete command service dependencies rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*CommandServiceDependencies)
	}{
		{name: "database", mutate: func(value *CommandServiceDependencies) { value.Database = nil }},
		{name: "attribution", mutate: func(value *CommandServiceDependencies) { value.ImportedAttributionResolver = nil }},
		{name: "projection", mutate: func(value *CommandServiceDependencies) { value.ProjectionRebuilder = nil }},
		{name: "delete restore", mutate: func(value *CommandServiceDependencies) { value.DeleteRestoreProviders = nil }},
		{name: "row rollback", mutate: func(value *CommandServiceDependencies) { value.RowRollbackProviders = nil }},
		{name: "non-row rollback", mutate: func(value *CommandServiceDependencies) { value.NonRowRollbackProviders = nil }},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			invalid := dependencies
			test.mutate(&invalid)
			if _, err := NewCommandService(invalid); !errors.Is(err, ErrInvalidCommandServiceDependency) {
				t.Fatalf("dependency error = %v", err)
			}
		})
	}
}

func validCommandServiceDependencies(t testing.TB) CommandServiceDependencies {
	t.Helper()
	deleteRestore, err := NewDeleteRestoreProviderCatalog([]string{"host"}, DeleteRestoreProviderRegistration{RecordType: "host", Provider: recorddeleterestore.TableProvider{}})
	if err != nil {
		t.Fatalf("delete/restore catalog: %v", err)
	}
	rows, err := NewRowProviderCatalog([]string{"host"}, RowProviderRegistration{RecordType: "host", Provider: catalogRowProvider{}})
	if err != nil {
		t.Fatalf("row catalog: %v", err)
	}
	nonRows, err := NewNonRowProviderCatalog([]string{"record_link"}, NonRowProviderRegistration{TargetKind: "record_link", Provider: stubNonRowProvider{}})
	if err != nil {
		t.Fatalf("non-row catalog: %v", err)
	}
	return CommandServiceDependencies{
		Database:                    commandServiceTestDB{},
		ImportedAttributionResolver: fakeImportedAttributionResolver{},
		ProjectionRebuilder:         commandServiceTestProjection{},
		DeleteRestoreProviders:      deleteRestore,
		RowRollbackProviders:        rows,
		NonRowRollbackProviders:     nonRows,
	}
}
