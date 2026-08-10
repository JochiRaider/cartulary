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

func TestTargetSemanticsCatalogRejectsDuplicateAndMissingRowProviders(t *testing.T) {
	t.Parallel()
	requirement := []targetSemanticsRequirement{{
		TargetKind:          "record",
		SourceOwnerID:       "record_source_owner",
		DispatchClass:       rollbackcontract.DispatchRow,
		AdmittedRecordTypes: []string{"host"},
		Addressability:      HistorySingleEntry,
	}}
	record := func(recordType string, provider rollbackcontract.RowSourceProvider) RecordProviderContribution {
		return RecordProviderContribution{
			SourceOwnerModule:   SourceOwnerEntities,
			RecordType:          recordType,
			RowRollbackProvider: provider,
		}
	}

	if _, err := compileTargetSemanticsCatalog(requirement, []ProviderContribution{{
		SourceOwnerModule: SourceOwnerEntities,
		Records:           []RecordProviderContribution{record("host", catalogRowProvider{}), record("host", catalogRowProvider{})},
	}}); !errors.Is(err, ErrDuplicateTargetSemantics) {
		t.Fatalf("duplicate row provider error = %v", err)
	}

	missingRequirement := append([]targetSemanticsRequirement(nil), requirement...)
	missingRequirement[0].AdmittedRecordTypes = []string{"host", "identity"}
	if _, err := compileTargetSemanticsCatalog(missingRequirement, []ProviderContribution{{
		SourceOwnerModule: SourceOwnerEntities,
		Records:           []RecordProviderContribution{record("host", catalogRowProvider{})},
	}}); !errors.Is(err, ErrInvalidTargetSemantics) {
		t.Fatalf("missing admitted row provider error = %v", err)
	}

	var typedNil *catalogRowProvider
	if _, err := compileTargetSemanticsCatalog(requirement, []ProviderContribution{{
		SourceOwnerModule: SourceOwnerEntities,
		Records:           []RecordProviderContribution{record("host", typedNil)},
	}}); !errors.Is(err, ErrInvalidTargetSemantics) {
		t.Fatalf("typed nil row provider error = %v", err)
	}
}
