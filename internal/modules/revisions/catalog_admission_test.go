package revisions

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

// Synthetic record families exercise catalog mechanics, not Records semantics.
func catalogAdmissionEnvelopeValidator(uuid.UUID, uuid.UUID, string, map[string]any) error {
	return nil
}

func catalogAdmissionSourceValidator(map[string]any, map[string]any) error { return nil }

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

type catalogTestHistoryValidator struct{}

func (catalogTestHistoryValidator) ValidateHistoryMutation(StoredMutation) error { return nil }

func catalogTestHistoryReferences(StoredMutation, uuid.UUID) ([]HistoryRecordReference, error) {
	return nil, nil
}

type catalogTestHistoryValidatorFunc func(StoredMutation) error

func (f catalogTestHistoryValidatorFunc) ValidateHistoryMutation(m StoredMutation) error { return f(m) }

func (validator *catalogAdmissionHistoryValidator) ValidateHistoryMutation(mutation StoredMutation) error {
	validator.mutation = mutation
	return validator.err
}

func TestCatalogAdmissionSnapshotCaptureBuildsCanonicalOpaqueEnvelope(t *testing.T) {
	recordID := uuid.New()
	incidentID := uuid.New()
	validationCalls := 0
	sourceCalls := 0
	var sourceError error
	validateSource := func(record, source map[string]any) error {
		sourceCalls++
		if record["record_id"] != recordID.String() || source["title"] != "source truth" {
			t.Fatalf("source validator context = %#v / %#v", record, source)
		}
		return sourceError
	}
	var validationError error
	validateEnvelope := func(gotIncident, gotRecord uuid.UUID, gotType string, value map[string]any) error {
		validationCalls++
		if gotIncident != incidentID || gotRecord != recordID || gotType != "catalog_admission" || value["record_id"] != recordID.String() {
			t.Fatalf("envelope validator context = %s/%s/%q/%#v", gotIncident, gotRecord, gotType, value)
		}
		return validationError
	}
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
			HistoryProjector:       catalogTestHistoryProjector,
			SourceOwnerModule:      SourceOwnerModule("catalog_admission_owner"),
			RecordType:             "catalog_admission",
			SnapshotSchemaID:       "cartulary.revisions.snapshot.catalog_admission.v1",
			ValidateSnapshotSource: validateSource,
			DeleteRestoreSource: catalogAdmissionSnapshotSource{
				value: sourceValue,
			},
		}},
	}}, []snapshotSchemaRequirement{{
		RecordType:       "catalog_admission",
		SourceOwner:      SourceOwnerModule("catalog_admission_owner"),
		SnapshotSchemaID: "cartulary.revisions.snapshot.catalog_admission.v1",
	}}, validateEnvelope)
	if err != nil {
		t.Fatalf("build catalog-admission snapshot catalog: %v", err)
	}
	snapshot, err := catalog.captureTx(context.Background(), nil, RecordEnvelope{IncidentID: incidentID, RecordID: recordID, RecordType: "catalog_admission"})
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
	if err := catalog.validatePersisted(incidentID, recordID, "catalog_admission", value); err != nil || validationCalls != 2 {
		t.Fatalf("capture and persisted validation: calls %d, err %v", validationCalls, err)
	}
	validationError = errors.New("owner rejects envelope")
	if _, err := catalog.captureTx(t.Context(), nil, RecordEnvelope{IncidentID: incidentID, RecordID: recordID, RecordType: "catalog_admission"}); !errors.Is(err, validationError) || !errors.Is(err, ErrInvalidRecordSnapshot) {
		t.Fatalf("capture lost owner rejection: %v", err)
	}
	if err := catalog.validatePersisted(incidentID, recordID, "catalog_admission", value); !errors.Is(err, validationError) || !errors.Is(err, ErrInvalidRecordSnapshot) {
		t.Fatalf("persisted value lost owner rejection: %v", err)
	}
	if validationCalls != 4 {
		t.Fatalf("owner validation calls = %d", validationCalls)
	}
	if sourceCalls != 2 {
		t.Fatalf("source validator calls after Records rejection = %d, want 2", sourceCalls)
	}
	validationError = nil
	sourceError = errors.New("owner rejects source")
	if _, err := catalog.captureTx(t.Context(), nil, RecordEnvelope{IncidentID: incidentID, RecordID: recordID, RecordType: "catalog_admission"}); !errors.Is(err, sourceError) || !errors.Is(err, ErrInvalidRecordSnapshot) {
		t.Fatalf("capture lost source rejection: %v", err)
	}
	if err := catalog.validatePersisted(incidentID, recordID, "catalog_admission", value); !errors.Is(err, sourceError) || !errors.Is(err, ErrInvalidRecordSnapshot) {
		t.Fatalf("persisted value lost source rejection: %v", err)
	}
	if sourceCalls != 4 {
		t.Fatalf("source validator calls = %d, want 4", sourceCalls)
	}
	sourceValue["source"].(map[string]any)["title"] = "mutated"
	if got := value["source"].(map[string]any)["title"]; got != "source truth" {
		t.Fatalf("record snapshot retained mutable source input: %#v", got)
	}

	facts, err := canonicalRevisionConflictFacts([]RevisionConflictFact{{
		FieldKey: "catalog_admission.title", BeforePresent: true, BeforeValue: "before title",
		AfterPresent: true, AfterValue: "after title",
	}, {
		FieldKey: "catalog_admission.tags", BeforePresent: true, BeforeValue: []any{"before"},
		AfterPresent: true, AfterValue: []any{"after"},
	}})
	if err != nil || len(facts) != 2 || facts[0].FieldKey != "catalog_admission.tags" || facts[1].FieldKey != "catalog_admission.title" || !facts[0].BeforePresent || !facts[0].AfterPresent {
		t.Fatalf("canonical explicit live conflict facts = %#v, err = %v", facts, err)
	}
	if _, err := canonicalRevisionConflictFacts([]RevisionConflictFact{{FieldKey: "duplicate"}, {FieldKey: "duplicate"}}); err == nil {
		t.Fatal("duplicate explicit conflict facts must be rejected")
	}
	if _, err := canonicalRevisionConflictFacts([]RevisionConflictFact{{FieldKey: " invalid"}}); err == nil {
		t.Fatal("non-canonical explicit conflict field key must be rejected")
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
	missingValidator := validProviderContributions()
	missingValidator[0].Records[0].ValidateSnapshotSource = nil
	if _, err := NewRecordSnapshotCaptureCatalog(missingValidator, catalogAdmissionEnvelopeValidator); !errors.Is(err, ErrMissingSnapshotCapture) {
		t.Fatalf("missing source validator admitted: %v", err)
	}
	if _, err := NewRecordSnapshotCaptureCatalog(validProviderContributions(), nil); !errors.Is(err, ErrMissingSnapshotCapture) {
		t.Fatalf("missing Records validator admitted: %v", err)
	}
	if _, err := NewRecordSnapshotCaptureCatalog(nil, catalogAdmissionEnvelopeValidator); !errors.Is(err, ErrMissingSnapshotCapture) {
		t.Fatalf("incomplete catalog error = %v", err)
	}
	if _, err := parseSnapshotSchemaRequirements([]byte(`{"schema_id":"wrong","registry_version":1,"schemas":[{}]}`)); !errors.Is(err, ErrInvalidRecordSnapshot) {
		t.Fatalf("invalid snapshot registry identity error = %v", err)
	}

	invalid, err := compileRecordSnapshotCaptureCatalog([]ProviderContribution{{
		SourceOwnerModule: SourceOwnerModule("catalog_admission_owner"),
		Records: []RecordProviderContribution{{
			HistoryProjector:       catalogTestHistoryProjector,
			SourceOwnerModule:      SourceOwnerModule("catalog_admission_owner"),
			RecordType:             "catalog_admission",
			SnapshotSchemaID:       "cartulary.revisions.snapshot.catalog_admission.v1",
			ValidateSnapshotSource: catalogAdmissionSourceValidator,
			DeleteRestoreSource:    catalogAdmissionSnapshotSource{value: map[string]any{"record": map[string]any{}, "source": map[string]any{}, "projection": map[string]any{}}},
		}},
	}}, []snapshotSchemaRequirement{{
		RecordType:       "catalog_admission",
		SourceOwner:      SourceOwnerModule("catalog_admission_owner"),
		SnapshotSchemaID: "cartulary.revisions.snapshot.catalog_admission.v1",
	}}, catalogAdmissionEnvelopeValidator)
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
		catalogAdmissionEnvelopeValidator,
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
	exerciseRequiredHistoryValidatorAdmission(t)
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

func exerciseRequiredHistoryValidatorAdmission(t *testing.T) {
	t.Helper()
	for _, required := range []bool{false, true} {
		for _, test := range []struct {
			name      string
			validator HistoryValidator
			present   bool
		}{
			{"absent", nil, false},
			{"typed_nil_pointer", (*catalogAdmissionHistoryValidator)(nil), false},
			{"typed_nil_function", catalogTestHistoryValidatorFunc(nil), false},
			{"value", catalogTestHistoryValidator{}, true},
			{"pointer", &catalogAdmissionHistoryValidator{}, true},
			{"function", catalogTestHistoryValidatorFunc(func(StoredMutation) error { return nil }), true},
		} {
			name := "optional/"
			if required {
				name = "required/"
			}
			t.Run(name+test.name, func(t *testing.T) {
				requirements := catalogAdmissionTargetRequirements()
				requirements[0].HistoryValidatorRequired = required
				contributions := catalogAdmissionTargetContributions(&catalogAdmissionNonRowProvider{})
				contributions[0].NonRowTargets[0].HistoryValidator = test.validator
				_, err := compileTargetSemanticsCatalog(requirements, contributions)
				admitted := test.present || (!required && test.validator == nil)
				if admitted {
					if err != nil {
						t.Fatalf("admitted validator rejected: %v", err)
					}
				} else if !errors.Is(err, ErrInvalidTargetSemantics) {
					t.Fatalf("absent/nil validator admitted: %v", err)
				}
			})
		}
	}
	t.Run("required_policy_cannot_bind_row_dispatch", func(t *testing.T) {
		requirements := catalogAdmissionTargetRequirements()
		requirements[1].HistoryValidatorRequired = true
		if _, err := compileTargetSemanticsCatalog(requirements, catalogAdmissionTargetContributions(&catalogAdmissionNonRowProvider{})); !errors.Is(err, ErrInvalidTargetSemantics) {
			t.Fatalf("non-row validator policy admitted on row dispatch: %v", err)
		}
	})
	for _, test := range []struct {
		name     string
		value    any
		present  bool
		admitted bool
	}{
		{"required", true, true, true}, {"optional", false, true, true},
		{"missing", nil, false, false}, {"null", nil, true, false},
		{"string", "true", true, false}, {"number", 1, true, false},
	} {
		t.Run("registry_policy/"+test.name, func(t *testing.T) {
			entry := map[string]any{"target_kind": "child", "source_owner": "catalog_admission_owner", "dispatch_class": "non_row", "admitted_record_types": []string{}, "history_record_id_fields": []string{"record_id"}, "addressability": "single_history_entry"}
			if test.present {
				entry["history_validator_required"] = test.value
			}
			payload, err := json.Marshal(map[string]any{"schema_id": "cartulary.revisions_target_semantics_registry.v1", "registry_version": 1, "targets": []any{entry}})
			if err != nil {
				t.Fatal(err)
			}
			got, err := parseTargetSemanticsRequirements(payload)
			if !test.admitted {
				if !errors.Is(err, ErrInvalidTargetSemantics) {
					t.Fatalf("missing or mistyped policy admitted: %v", err)
				}
			} else if err != nil || len(got) != 1 || got[0].HistoryValidatorRequired != test.value.(bool) {
				t.Fatalf("registry policy lost: %#v, %v", got, err)
			}
		})
	}
	t.Run("current_required_targets", func(t *testing.T) {
		requirements, err := currentTargetSemanticsRequirements()
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, requirement := range requirements {
			if requirement.HistoryValidatorRequired {
				got = append(got, requirement.TargetKind)
			}
		}
		if !reflect.DeepEqual(got, []string{"entity_alias", "entity_mention", "entity_preserved_identifier", "indicator_observation", "indicator_state_interval", "record_link", "record_tag"}) {
			t.Fatalf("required validators=%v", got)
		}
	})
}

func catalogAdmissionSnapshotContribution(value map[string]any) ProviderContribution {
	return ProviderContribution{
		SourceOwnerModule: SourceOwnerModule("catalog_admission_owner"),
		Records: []RecordProviderContribution{{
			HistoryProjector:       catalogTestHistoryProjector,
			SourceOwnerModule:      SourceOwnerModule("catalog_admission_owner"),
			RecordType:             "catalog_admission",
			SnapshotSchemaID:       "cartulary.revisions.snapshot.catalog_admission.v1",
			ValidateSnapshotSource: catalogAdmissionSourceValidator,
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
			HistoryProjector:       catalogTestHistoryProjector,
			SourceOwnerModule:      SourceOwnerModule("catalog_admission_owner"),
			RecordType:             "catalog_admission",
			SnapshotSchemaID:       "cartulary.revisions.snapshot.catalog_admission.v1",
			ValidateSnapshotSource: catalogAdmissionSourceValidator,
			RowRollbackProvider:    catalogAdmissionRowProvider{},
		}},
		NonRowTargets: []NonRowProviderContribution{{
			HistoryProjector:  catalogTestHistoryProjector,
			SourceOwnerModule: SourceOwnerModule("catalog_admission_owner"),
			TargetKind:        "child",
			HistoryReferences: catalogTestHistoryReferences,
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

func (catalogAdmissionSnapshotSource) ValidateStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}
