package revisions

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type catalogAdmissionSnapshotSource struct {
	value map[string]any
}

func (source catalogAdmissionSnapshotSource) SnapshotTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error) {
	return source.value, nil
}

func (catalogAdmissionSnapshotSource) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (catalogAdmissionSnapshotSource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.catalog_admission.v1", nil
}

func (catalogAdmissionSnapshotSource) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

type catalogAdmissionRowProvider struct{}

func (catalogAdmissionRowProvider) ValidateRollbackValue(map[string]any) error { return nil }
func (catalogAdmissionRowProvider) RestoreTx(context.Context, pgx.Tx, rollbackcontract.RestoreRequest) error {
	return nil
}

type catalogAdmissionNonRowProvider struct{}

func (*catalogAdmissionNonRowProvider) DescribeTx(context.Context, pgx.Tx, rollbackcontract.DescribeRequest) (rollbackcontract.TargetDescriptor, error) {
	return rollbackcontract.TargetDescriptor{}, nil
}

func (*catalogAdmissionNonRowProvider) ApplyInverseTx(context.Context, pgx.Tx, rollbackcontract.ApplyInverseRequest) (rollbackcontract.ApplyInverseResult, error) {
	return rollbackcontract.ApplyInverseResult{}, nil
}

type catalogAdmissionHistoryValidator struct {
	mutation StoredMutation
	err      error
}

func (validator *catalogAdmissionHistoryValidator) ValidateHistoryMutation(mutation StoredMutation) error {
	validator.mutation = mutation
	return validator.err
}

func TestCatalogAdmissionSnapshotCaptureBuildsCanonicalOpaqueEnvelope(t *testing.T) {
	recordID := uuid.New()
	sourceValue := map[string]any{
		"record": map[string]any{
			"record_id":   recordID.String(),
			"record_type": "catalog_admission",
			"row_version": float64(3),
		},
		"source": map[string]any{"title": "source truth"},
	}
	catalog, err := compileRecordSnapshotCaptureCatalog([]ProviderContribution{{
		SourceOwnerModule: SourceOwnerModule("catalog_admission_owner"),
		Records: []RecordProviderContribution{{
			SourceOwnerModule: SourceOwnerModule("catalog_admission_owner"),
			RecordType:        "catalog_admission",
			SnapshotSchemaID:  "cartulary.revisions.snapshot.catalog_admission.v1",
			DeleteRestoreSource: catalogAdmissionSnapshotSource{
				value: sourceValue,
			},
		}},
	}}, []snapshotSchemaRequirement{{
		RecordType:       "catalog_admission",
		SourceOwner:      SourceOwnerModule("catalog_admission_owner"),
		SnapshotSchemaID: "cartulary.revisions.snapshot.catalog_admission.v1",
	}})
	if err != nil {
		t.Fatalf("build catalog-admission snapshot catalog: %v", err)
	}
	snapshot, err := catalog.captureTx(context.Background(), nil, RecordEnvelope{RecordID: recordID, RecordType: "catalog_admission"})
	if err != nil {
		t.Fatalf("capture canonical snapshot: %v", err)
	}
	if snapshot.RecordID() != recordID || snapshot.SnapshotSchemaID() != "cartulary.revisions.snapshot.catalog_admission.v1" {
		t.Fatalf("catalog-admission snapshot identity = %s/%q", snapshot.RecordID(), snapshot.SnapshotSchemaID())
	}
	value, err := recordSnapshotValue(&snapshot, recordID)
	if err != nil {
		t.Fatalf("read record snapshot internally: %v", err)
	}
	if got := value["snapshot_schema_id"]; got != "cartulary.revisions.snapshot.catalog_admission.v1" {
		t.Fatalf("snapshot schema id = %#v", got)
	}
	sourceValue["source"].(map[string]any)["title"] = "mutated"
	if got := value["source"].(map[string]any)["title"]; got != "source truth" {
		t.Fatalf("record snapshot retained mutable source input: %#v", got)
	}

	facts, err := recordRevisionConflictFacts(LiveRecordChange{
		BeforeValue: map[string]any{"cells": map[string]any{"catalog_admission.tags": map[string]any{"value": []any{"before"}}}},
		AfterValue:  map[string]any{"cells": map[string]any{"catalog_admission.tags": map[string]any{"value": []any{"after"}}}},
	})
	if err != nil || len(facts) != 1 || facts[0].FieldKey != "catalog_admission.tags" || !facts[0].BeforePresent || !facts[0].AfterPresent {
		t.Fatalf("derived live conflict facts = %#v, err = %v", facts, err)
	}
	nullPayload, err := revisionConflictFactValue(nil, true)
	if err != nil || string(nullPayload.([]byte)) != "null" {
		t.Fatalf("present JSON null conflict fact = %#v, err = %v", nullPayload, err)
	}
	absentPayload, err := revisionConflictFactValue(nil, false)
	if err != nil || absentPayload != nil {
		t.Fatalf("absent conflict fact = %#v, err = %v", absentPayload, err)
	}
}

func TestCatalogAdmissionSnapshotCaptureRejectsInvalidShapeAndMissingRegistration(t *testing.T) {
	recordID := uuid.New()
	if _, err := NewRecordSnapshotCaptureCatalog(nil); !errors.Is(err, ErrMissingSnapshotCapture) {
		t.Fatalf("incomplete catalog error = %v", err)
	}
	if _, err := parseSnapshotSchemaRequirements([]byte(`{"schema_id":"wrong","registry_version":1,"schemas":[{}]}`)); !errors.Is(err, ErrInvalidRecordSnapshot) {
		t.Fatalf("invalid snapshot registry identity error = %v", err)
	}

	invalid, err := compileRecordSnapshotCaptureCatalog([]ProviderContribution{{
		SourceOwnerModule: SourceOwnerModule("catalog_admission_owner"),
		Records: []RecordProviderContribution{{
			SourceOwnerModule:   SourceOwnerModule("catalog_admission_owner"),
			RecordType:          "catalog_admission",
			SnapshotSchemaID:    "cartulary.revisions.snapshot.catalog_admission.v1",
			DeleteRestoreSource: catalogAdmissionSnapshotSource{value: map[string]any{"record": map[string]any{}, "source": map[string]any{}, "projection": map[string]any{}}},
		}},
	}}, []snapshotSchemaRequirement{{
		RecordType:       "catalog_admission",
		SourceOwner:      SourceOwnerModule("catalog_admission_owner"),
		SnapshotSchemaID: "cartulary.revisions.snapshot.catalog_admission.v1",
	}})
	if err != nil {
		t.Fatalf("build invalid-shape source catalog: %v", err)
	}
	if _, err := invalid.captureTx(context.Background(), nil, RecordEnvelope{RecordID: recordID, RecordType: "catalog_admission"}); !errors.Is(err, ErrInvalidRecordSnapshot) {
		t.Fatalf("invalid capture error = %v", err)
	}
	duplicate := catalogAdmissionSnapshotContribution(map[string]any{
		"record": map[string]any{"record_id": recordID.String(), "record_type": "catalog_admission"},
		"source": map[string]any{},
	})
	duplicate.Records = append(duplicate.Records, duplicate.Records[0])
	if _, err := compileRecordSnapshotCaptureCatalog(
		[]ProviderContribution{duplicate},
		[]snapshotSchemaRequirement{{
			RecordType:       "catalog_admission",
			SourceOwner:      SourceOwnerModule("catalog_admission_owner"),
			SnapshotSchemaID: "cartulary.revisions.snapshot.catalog_admission.v1",
		}},
	); !errors.Is(err, ErrDuplicateSnapshotCapture) {
		t.Fatalf("duplicate snapshot provider error = %v", err)
	}
}

func TestTargetSemanticsCatalogCompilesGenericRowAndOwnerNonRowEntries(t *testing.T) {
	provider := &catalogAdmissionNonRowProvider{}
	validator := &catalogAdmissionHistoryValidator{}
	requirements := catalogAdmissionTargetRequirements()
	contributions := catalogAdmissionTargetContributions(provider)
	contributions[0].NonRowTargets[0].HistoryValidator = validator
	catalog, err := compileTargetSemanticsCatalog(requirements, contributions)
	if err != nil {
		t.Fatalf("compile target semantics: %v", err)
	}
	targetKinds := catalog.targetKinds()
	if !reflect.DeepEqual(targetKinds, []string{"child", "record"}) {
		t.Fatalf("compiled target kinds = %#v", targetKinds)
	}
	targetKinds[1] = "mutated"
	requirements[0].HistoryRecordIDFields[0] = "mutated"
	if next := catalog.targetKinds(); !reflect.DeepEqual(next, []string{"child", "record"}) {
		t.Fatalf("catalog target-kind output was mutable: %#v", next)
	}
	if targetKind, err := catalog.defaultRowTargetKind("catalog_admission"); err != nil || targetKind != "record" {
		t.Fatalf("default row target = %q, %v", targetKind, err)
	}
	if dispatch, err := catalog.dispatchClass("record"); err != nil || dispatch != rollbackcontract.DispatchRow {
		t.Fatalf("row dispatch = %q, %v", dispatch, err)
	}
	if _, err := catalog.rowProvider("record", "catalog_admission"); err != nil {
		t.Fatalf("resolve catalog-admission row provider: %v", err)
	}
	if dispatch, err := catalog.dispatchClass("child"); err != nil || dispatch != rollbackcontract.DispatchNonRow {
		t.Fatalf("non-row dispatch = %q, %v", dispatch, err)
	}
	if _, err := catalog.nonRowProvider("child"); err != nil {
		t.Fatalf("resolve child provider: %v", err)
	}
	if _, err := catalog.rowProvider("child", "catalog_admission"); !errors.Is(err, ErrInvalidTargetSemantics) {
		t.Fatalf("non-row target admitted as row: %v", err)
	}

	first := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	second := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	description, err := catalog.DescribeMutation(StoredMutation{
		TargetKind:    "child",
		TargetID:      "child-1",
		OperationKind: "patch",
		BeforeValue: map[string]any{
			"record_id": second.String(),
		},
		AfterValue: map[string]any{
			"record_id": first.String(),
		},
	})
	if err != nil {
		t.Fatalf("describe child mutation: %v", err)
	}
	want := []uuid.UUID{first, second}
	if !reflect.DeepEqual(description.HistoryRecordIDs, want) || !reflect.DeepEqual(description.HistoryEntryRecordIDs, want) {
		t.Fatalf("history description = %#v, want ids %#v", description, want)
	}
	if validator.mutation.OperationKind != "patch" || validator.mutation.TargetID != "child-1" {
		t.Fatalf("owner validator did not receive stored mutation: %#v", validator.mutation)
	}
	validator.err = errors.New("owner detail must not escape")
	if _, err := catalog.DescribeMutation(StoredMutation{TargetKind: "child", TargetID: "child-2", OperationKind: "create"}); !errors.Is(err, ErrInvalidTargetSemantics) || errors.Is(err, validator.err) {
		t.Fatalf("owner validation error was not normalized: %v", err)
	}
}

func TestTargetSemanticsCatalogFailsClosed(t *testing.T) {
	provider := &catalogAdmissionNonRowProvider{}
	if _, err := parseTargetSemanticsRequirements([]byte(`{"schema_id":"wrong","registry_version":1,"targets":[{}]}`)); !errors.Is(err, ErrInvalidTargetSemantics) {
		t.Fatalf("invalid target registry identity error = %v", err)
	}
	tests := []struct {
		name          string
		requirements  []targetSemanticsRequirement
		contributions []ProviderContribution
		want          error
	}{
		{name: "missing", requirements: append(catalogAdmissionTargetRequirements(), targetSemanticsRequirement{TargetKind: "absent", SourceOwnerID: "catalog_admission_owner", DispatchClass: rollbackcontract.DispatchNonRow, Addressability: HistorySingleEntry}), contributions: catalogAdmissionTargetContributions(provider), want: ErrMissingTargetSemantics},
		{name: "duplicate requirement", requirements: append(catalogAdmissionTargetRequirements(), catalogAdmissionTargetRequirements()[0]), contributions: catalogAdmissionTargetContributions(provider), want: ErrDuplicateTargetSemantics},
		{name: "unknown contribution", requirements: catalogAdmissionTargetRequirements()[1:], contributions: catalogAdmissionTargetContributions(provider), want: ErrUnexpectedTargetSemantics},
		{name: "missing history", requirements: catalogAdmissionTargetRequirements(), contributions: catalogAdmissionTargetContributionsWithoutHistory(provider), want: ErrInvalidTargetSemantics},
		{name: "typed nil provider", requirements: catalogAdmissionTargetRequirements(), contributions: catalogAdmissionTargetContributions((*catalogAdmissionNonRowProvider)(nil)), want: ErrInvalidTargetSemantics},
		{name: "field mismatch", requirements: catalogAdmissionTargetRequirementsWithFields([]string{"other_record_id"}), contributions: catalogAdmissionTargetContributions(provider), want: ErrInvalidTargetSemantics},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := compileTargetSemanticsCatalog(test.requirements, test.contributions)
			if !errors.Is(err, test.want) {
				t.Fatalf("catalog error = %v, want %v", err, test.want)
			}
		})
	}
}

func catalogAdmissionSnapshotContribution(value map[string]any) ProviderContribution {
	return ProviderContribution{
		SourceOwnerModule: SourceOwnerModule("catalog_admission_owner"),
		Records: []RecordProviderContribution{{
			SourceOwnerModule: SourceOwnerModule("catalog_admission_owner"),
			RecordType:        "catalog_admission",
			SnapshotSchemaID:  "cartulary.revisions.snapshot.catalog_admission.v1",
			DeleteRestoreSource: catalogAdmissionSnapshotSource{
				value: value,
			},
		}},
	}
}

func catalogAdmissionTargetRequirements() []targetSemanticsRequirement {
	return []targetSemanticsRequirement{
		{TargetKind: "child", SourceOwnerID: "catalog_admission_owner", DispatchClass: rollbackcontract.DispatchNonRow, HistoryRecordIDFields: []string{"record_id"}, Addressability: HistorySingleEntry},
		{TargetKind: "record", SourceOwnerID: "record_source_owner", DispatchClass: rollbackcontract.DispatchRow, AdmittedRecordTypes: []string{"catalog_admission"}, Addressability: HistorySingleEntry},
	}
}

func catalogAdmissionTargetRequirementsWithFields(fields []string) []targetSemanticsRequirement {
	values := catalogAdmissionTargetRequirements()
	values[0].HistoryRecordIDFields = fields
	return values
}

func catalogAdmissionTargetContributions(provider rollbackcontract.NonRowTargetProvider) []ProviderContribution {
	return []ProviderContribution{{
		SourceOwnerModule: SourceOwnerModule("catalog_admission_owner"),
		Records: []RecordProviderContribution{{
			SourceOwnerModule:   SourceOwnerModule("catalog_admission_owner"),
			RecordType:          "catalog_admission",
			RowRollbackProvider: catalogAdmissionRowProvider{},
		}},
		NonRowTargets: []NonRowProviderContribution{{
			SourceOwnerModule: SourceOwnerModule("catalog_admission_owner"),
			TargetKind:        "child",
			HistoryFacet:      NewFieldAssociationHistoryFacet([]string{"record_id"}, HistorySingleEntry),
			RollbackProvider:  provider,
		}},
	}}
}

func catalogAdmissionTargetContributionsWithoutHistory(provider rollbackcontract.NonRowTargetProvider) []ProviderContribution {
	values := catalogAdmissionTargetContributions(provider)
	values[0].NonRowTargets[0].HistoryFacet = HistoryFacet{}
	return values
}
