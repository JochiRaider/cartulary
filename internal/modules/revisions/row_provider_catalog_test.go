package revisions

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type catalogRowProvider struct{}

func (catalogRowProvider) ValidateRollbackValue(map[string]any) error { return nil }
func (catalogRowProvider) RestoreTx(context.Context, pgx.Tx, rollbackcontract.RestoreRequest) error {
	return nil
}
func (catalogRowProvider) TouchTx(context.Context, pgx.Tx, rollbackcontract.TouchRequest) error {
	return nil
}

func TestRowProviderCatalogRejectsDuplicateAndMissingRegistrations(t *testing.T) {
	t.Parallel()

	provider := catalogRowProvider{}
	if _, err := NewRowProviderCatalog([]string{"host"},
		RowProviderRegistration{RecordType: "host", Provider: provider},
		RowProviderRegistration{RecordType: "host", Provider: provider},
	); !errors.Is(err, ErrDuplicateRowRollbackProvider) {
		t.Fatalf("duplicate registration error = %v", err)
	}

	if _, err := NewRowProviderCatalog([]string{"identity", "host"}, RowProviderRegistration{RecordType: "host", Provider: provider}); !errors.Is(err, ErrMissingRowRollbackProvider) {
		t.Fatalf("missing registration error = %v", err)
	}
	if _, err := NewRowProviderCatalog([]string{"host"}, RowProviderRegistration{RecordType: "identity", Provider: provider}); !errors.Is(err, ErrUnexpectedRowRollbackProvider) {
		t.Fatalf("unexpected registration error = %v", err)
	}
	var typedNil *catalogRowProvider
	if _, err := NewRowProviderCatalog([]string{"host"}, RowProviderRegistration{RecordType: "host", Provider: typedNil}); !errors.Is(err, ErrMissingRowRollbackProvider) {
		t.Fatalf("typed nil provider error = %v", err)
	}
}
