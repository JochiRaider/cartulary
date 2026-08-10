package revisions

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type candidateSnapshotSource struct {
	value map[string]any
}

func (source candidateSnapshotSource) SnapshotTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error) {
	return source.value, nil
}

func (candidateSnapshotSource) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (candidateSnapshotSource) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.candidate.v1", nil
}

func (candidateSnapshotSource) ValidateDeletePreconditionsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (string, bool, error) {
	return "", false, nil
}

type candidateRowProvider struct{}

func (candidateRowProvider) ValidateRollbackValue(map[string]any) error { return nil }
func (candidateRowProvider) RestoreTx(context.Context, pgx.Tx, rollbackcontract.RestoreRequest) error {
	return nil
}

type candidateNonRowProvider struct{}

func (*candidateNonRowProvider) DescribeTx(context.Context, pgx.Tx, rollbackcontract.DescribeRequest) (rollbackcontract.TargetDescriptor, error) {
	return rollbackcontract.TargetDescriptor{}, nil
}

func (*candidateNonRowProvider) ApplyInverseTx(context.Context, pgx.Tx, rollbackcontract.ApplyInverseRequest) (rollbackcontract.ApplyInverseResult, error) {
	return rollbackcontract.ApplyInverseResult{}, nil
}

func TestCandidateSnapshotCaptureBuildsCanonicalOpaqueEnvelope(t *testing.T) {
	recordID := uuid.New()
	sourceValue := map[string]any{
		"record": map[string]any{
			"record_id":   recordID.String(),
			"record_type": "candidate",
			"row_version": float64(3),
		},
		"source": map[string]any{"title": "source truth"},
	}
	catalog, err := NewRecordSnapshotCaptureCatalogForRequirements([]ProviderContribution{{
		SourceOwnerModule: SourceOwnerModule("candidate_owner"),
		Records: []RecordProviderContribution{{
			SourceOwnerModule: SourceOwnerModule("candidate_owner"),
			RecordType:        "candidate",
			SnapshotSchemaID:  "cartulary.revisions.snapshot.candidate.v1",
			DeleteRestoreSource: candidateSnapshotSource{
				value: sourceValue,
			},
		}},
	}}, []SnapshotSchemaRequirement{{
		RecordType:       "candidate",
		SourceOwner:      SourceOwnerModule("candidate_owner"),
		SnapshotSchemaID: "cartulary.revisions.snapshot.candidate.v1",
	}})
	if err != nil {
		t.Fatalf("build candidate snapshot catalog: %v", err)
	}
	snapshot, err := catalog.captureTx(context.Background(), nil, RecordEnvelope{RecordID: recordID, RecordType: "candidate"})
	if err != nil {
		t.Fatalf("capture canonical snapshot: %v", err)
	}
	if snapshot.RecordID() != recordID || snapshot.SnapshotSchemaID() != "cartulary.revisions.snapshot.candidate.v1" {
		t.Fatalf("captured snapshot identity = %s/%q", snapshot.RecordID(), snapshot.SnapshotSchemaID())
	}
	value, err := capturedSnapshotValue(&snapshot, recordID)
	if err != nil {
		t.Fatalf("read captured snapshot internally: %v", err)
	}
	if got := value["snapshot_schema_id"]; got != "cartulary.revisions.snapshot.candidate.v1" {
		t.Fatalf("snapshot schema id = %#v", got)
	}
	sourceValue["source"].(map[string]any)["title"] = "mutated"
	if got := value["source"].(map[string]any)["title"]; got != "source truth" {
		t.Fatalf("captured snapshot retained mutable source input: %#v", got)
	}

	facts, err := recordRevisionConflictFacts(LiveRecordChange{
		BeforeValue: map[string]any{"cells": map[string]any{"candidate.tags": map[string]any{"value": []any{"before"}}}},
		AfterValue:  map[string]any{"cells": map[string]any{"candidate.tags": map[string]any{"value": []any{"after"}}}},
	})
	if err != nil || len(facts) != 1 || facts[0].FieldKey != "candidate.tags" || !facts[0].BeforePresent || !facts[0].AfterPresent {
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

func TestCandidateSnapshotCaptureRejectsInvalidShapeAndMissingRegistration(t *testing.T) {
	recordID := uuid.New()
	if _, err := NewRecordSnapshotCaptureCatalog(nil); !errors.Is(err, ErrMissingSnapshotCapture) {
		t.Fatalf("incomplete catalog error = %v", err)
	}

	invalid, err := NewRecordSnapshotCaptureCatalogForRequirements([]ProviderContribution{{
		SourceOwnerModule: SourceOwnerModule("candidate_owner"),
		Records: []RecordProviderContribution{{
			SourceOwnerModule:   SourceOwnerModule("candidate_owner"),
			RecordType:          "candidate",
			SnapshotSchemaID:    "cartulary.revisions.snapshot.candidate.v1",
			DeleteRestoreSource: candidateSnapshotSource{value: map[string]any{"record": map[string]any{}, "source": map[string]any{}, "projection": map[string]any{}}},
		}},
	}}, []SnapshotSchemaRequirement{{
		RecordType:       "candidate",
		SourceOwner:      SourceOwnerModule("candidate_owner"),
		SnapshotSchemaID: "cartulary.revisions.snapshot.candidate.v1",
	}})
	if err != nil {
		t.Fatalf("build invalid-shape source catalog: %v", err)
	}
	if _, err := invalid.captureTx(context.Background(), nil, RecordEnvelope{RecordID: recordID, RecordType: "candidate"}); !errors.Is(err, ErrInvalidCapturedSnapshot) {
		t.Fatalf("invalid capture error = %v", err)
	}
}

func TestTargetSemanticsCatalogCompilesGenericRowAndOwnerNonRowEntries(t *testing.T) {
	provider := &candidateNonRowProvider{}
	requirements := candidateTargetRequirements()
	contributions := candidateTargetContributions(provider)
	catalog, err := NewTargetSemanticsCatalog(requirements, contributions)
	if err != nil {
		t.Fatalf("compile target semantics: %v", err)
	}
	descriptors := catalog.Descriptors()
	if len(descriptors) != 2 || descriptors[0].TargetKind != "child" || descriptors[1].TargetKind != "record" {
		t.Fatalf("compiled descriptors = %#v", descriptors)
	}
	descriptors[1].AdmittedRecordTypes[0] = "mutated"
	if next := catalog.Descriptors(); next[1].AdmittedRecordTypes[0] != "candidate" {
		t.Fatalf("catalog descriptor output was mutable: %#v", next)
	}
	if targetKind, err := catalog.defaultRowTargetKind("candidate"); err != nil || targetKind != "record" {
		t.Fatalf("default row target = %q, %v", targetKind, err)
	}
	if dispatch, err := catalog.dispatchClass("record"); err != nil || dispatch != RollbackDispatchRow {
		t.Fatalf("row dispatch = %q, %v", dispatch, err)
	}
	if _, err := catalog.rowProvider("record", "candidate"); err != nil {
		t.Fatalf("resolve candidate row provider: %v", err)
	}
	if dispatch, err := catalog.dispatchClass("child"); err != nil || dispatch != RollbackDispatchNonRow {
		t.Fatalf("non-row dispatch = %q, %v", dispatch, err)
	}
	if _, err := catalog.nonRowProvider("child"); err != nil {
		t.Fatalf("resolve child provider: %v", err)
	}
	if _, err := catalog.rowProvider("child", "candidate"); !errors.Is(err, ErrInvalidTargetSemantics) {
		t.Fatalf("non-row target admitted as row: %v", err)
	}

	first := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	second := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	description, err := catalog.DescribeMutation(StoredMutation{
		TargetKind: "child",
		TargetID:   "child-1",
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
}

func TestTargetSemanticsCatalogFailsClosed(t *testing.T) {
	provider := &candidateNonRowProvider{}
	tests := []struct {
		name          string
		requirements  []TargetSemanticsRequirement
		contributions []ProviderContribution
		want          error
	}{
		{name: "missing", requirements: append(candidateTargetRequirements(), TargetSemanticsRequirement{TargetKind: "absent", SourceOwnerID: "candidate_owner", DispatchClass: RollbackDispatchNonRow, Addressability: HistorySingleEntry}), contributions: candidateTargetContributions(provider), want: ErrMissingTargetSemantics},
		{name: "duplicate requirement", requirements: append(candidateTargetRequirements(), candidateTargetRequirements()[0]), contributions: candidateTargetContributions(provider), want: ErrDuplicateTargetSemantics},
		{name: "unknown contribution", requirements: candidateTargetRequirements()[1:], contributions: candidateTargetContributions(provider), want: ErrUnexpectedTargetSemantics},
		{name: "missing history", requirements: candidateTargetRequirements(), contributions: candidateTargetContributionsWithoutHistory(provider), want: ErrInvalidTargetSemantics},
		{name: "typed nil provider", requirements: candidateTargetRequirements(), contributions: candidateTargetContributions((*candidateNonRowProvider)(nil)), want: ErrInvalidTargetSemantics},
		{name: "field mismatch", requirements: candidateTargetRequirementsWithFields([]string{"other_record_id"}), contributions: candidateTargetContributions(provider), want: ErrInvalidTargetSemantics},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewTargetSemanticsCatalog(test.requirements, test.contributions)
			if !errors.Is(err, test.want) {
				t.Fatalf("catalog error = %v, want %v", err, test.want)
			}
		})
	}
}

func candidateTargetRequirements() []TargetSemanticsRequirement {
	return []TargetSemanticsRequirement{
		{TargetKind: "child", SourceOwnerID: "candidate_owner", DispatchClass: RollbackDispatchNonRow, HistoryRecordIDFields: []string{"record_id"}, Addressability: HistorySingleEntry},
		{TargetKind: "record", SourceOwnerID: "record_source_owner", DispatchClass: RollbackDispatchRow, AdmittedRecordTypes: []string{"candidate"}, Addressability: HistorySingleEntry},
	}
}

func candidateTargetRequirementsWithFields(fields []string) []TargetSemanticsRequirement {
	values := candidateTargetRequirements()
	values[0].HistoryRecordIDFields = fields
	return values
}

func candidateTargetContributions(provider rollbackcontract.NonRowTargetProvider) []ProviderContribution {
	return []ProviderContribution{{
		SourceOwnerModule: SourceOwnerModule("candidate_owner"),
		Records: []RecordProviderContribution{{
			SourceOwnerModule:   SourceOwnerModule("candidate_owner"),
			RecordType:          "candidate",
			RowRollbackProvider: candidateRowProvider{},
		}},
		NonRowTargets: []NonRowProviderContribution{{
			SourceOwnerModule: SourceOwnerModule("candidate_owner"),
			TargetKind:        "child",
			HistorySemantics:  NewFieldHistoryTargetSemantics([]string{"record_id"}, HistorySingleEntry),
			RollbackProvider:  provider,
		}},
	}}
}

func candidateTargetContributionsWithoutHistory(provider rollbackcontract.NonRowTargetProvider) []ProviderContribution {
	values := candidateTargetContributions(provider)
	values[0].NonRowTargets[0].HistorySemantics = nil
	return values
}
