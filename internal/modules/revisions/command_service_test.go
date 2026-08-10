package revisions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (commandServiceTestProjection) Supports(string) bool {
	return false
}

func (commandServiceTestProjection) LoadRowTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error) {
	return nil, pgx.ErrNoRows
}

type commandServiceTestHistoricalPolicy struct{}

func (commandServiceTestHistoricalPolicy) IsSuppressedTx(context.Context, pgx.Tx) (bool, error) {
	return false, nil
}

type commandServiceTestAuthorizer struct{}

func (commandServiceTestAuthorizer) AuthorizeCommandTx(context.Context, pgx.Tx, uuid.UUID, ActorID, CommandKind) error {
	return nil
}

type commandServiceTestIdempotency struct{}

func (commandServiceTestIdempotency) Get(context.Context, IdempotencyKey) (IdempotencyRecord, error) {
	return IdempotencyRecord{}, ErrIdempotencyNotFound
}

func (commandServiceTestIdempotency) PutSuccessTx(context.Context, pgx.Tx, IdempotencyKey, []byte, map[string]any) error {
	return nil
}

type commandServiceTestEnvelopes struct{}

func (commandServiceTestEnvelopes) LoadEnvelope(context.Context, uuid.UUID) (RecordEnvelope, error) {
	return RecordEnvelope{}, ErrEnvelopeNotFound
}

func (commandServiceTestEnvelopes) LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID, bool) (RecordEnvelope, error) {
	return RecordEnvelope{}, ErrEnvelopeNotFound
}

func (commandServiceTestEnvelopes) AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error) {
	return 0, nil
}

func (commandServiceTestEnvelopes) SetDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) (int64, error) {
	return 0, nil
}

func (commandServiceTestEnvelopes) LockDestructiveRecordsNowaitTx(context.Context, pgx.Tx, []uuid.UUID) error {
	return nil
}

type incidentBundleEnvelopeReaderStub struct{}

func (incidentBundleEnvelopeReaderStub) RecordTypeTx(
	context.Context,
	pgx.Tx,
	uuid.UUID,
	uuid.UUID,
) (string, error) {
	return "host", nil
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
		{name: "transactions", mutate: func(value *CommandServiceDependencies) { value.Transactions = nil }},
		{name: "authorization", mutate: func(value *CommandServiceDependencies) { value.Authorization = nil }},
		{name: "idempotency", mutate: func(value *CommandServiceDependencies) { value.Idempotency = nil }},
		{name: "attribution", mutate: func(value *CommandServiceDependencies) { value.ImportedAttributionResolver = nil }},
		{name: "projection", mutate: func(value *CommandServiceDependencies) { value.Projections = nil }},
		{name: "live records", mutate: func(value *CommandServiceDependencies) { value.LiveRecords = nil }},
		{name: "delete restore sources", mutate: func(value *CommandServiceDependencies) { value.DeleteRestoreSources = nil }},
		{name: "target semantics", mutate: func(value *CommandServiceDependencies) { value.TargetSemantics = nil }},
		{name: "appender", mutate: func(value *CommandServiceDependencies) { value.Appender = nil }},
		{name: "record envelopes", mutate: func(value *CommandServiceDependencies) { value.RecordEnvelopes = nil }},
		{name: "clock", mutate: func(value *CommandServiceDependencies) { value.Clock = nil }},
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
	database := commandServiceTestDB{}
	contributions := validProviderContributions()
	deleteRestoreSources, err := buildDeleteRestoreSourceCatalog(contributions)
	if err != nil {
		t.Fatalf("build delete/restore sources: %v", err)
	}
	targetSemantics := validTargetSemanticsCatalog(t, contributions)
	return CommandServiceDependencies{
		Transactions:                database,
		Authorization:               commandServiceTestAuthorizer{},
		Idempotency:                 commandServiceTestIdempotency{},
		ImportedAttributionResolver: fakeImportedAttributionResolver{},
		Projections:                 commandServiceTestProjection{},
		LiveRecords:                 commandServiceTestProjection{},
		DeleteRestoreSources:        deleteRestoreSources,
		TargetSemantics:             targetSemantics,
		Appender: &Appender{
			recordViews:      &RecordViewCatalog{},
			historicalPolicy: commandServiceTestHistoricalPolicy{},
		},
		RecordEnvelopes: commandServiceTestEnvelopes{},
		Clock:           func() time.Time { return time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC) },
	}
}

func validProviderContributions() []ProviderContribution {
	record := func(owner SourceOwnerModule, recordType string, historyTargetKinds ...string) RecordProviderContribution {
		return RecordProviderContribution{
			SourceOwnerModule:   owner,
			RecordType:          recordType,
			SnapshotSchemaID:    "cartulary.revisions.snapshot." + recordType + ".v1",
			HistoryTargetKinds:  append([]string(nil), historyTargetKinds...),
			DeleteRestoreSource: testDeleteRestoreSource{},
			RowRollbackProvider: catalogRowProvider{},
		}
	}
	nonRow := func(owner SourceOwnerModule, targetKind string) NonRowProviderContribution {
		fields := map[string][]string{
			"entity_alias":                {"record_id"},
			"entity_mention":              {"source_record_id"},
			"entity_preserved_identifier": {"record_id"},
			"indicator_observation":       {"source_record_id", "resolved_indicator_record_id"},
			"indicator_state_interval":    {"indicator_record_id"},
			"record_link":                 {"src_record_id", "dst_record_id"},
			"record_tag":                  {"record_id"},
		}
		addressability := HistorySingleEntry
		if targetKind == "entity_alias" || targetKind == "entity_preserved_identifier" {
			addressability = HistoryNotIndividuallyAddressable
		}
		return NonRowProviderContribution{
			SourceOwnerModule: owner,
			TargetKind:        targetKind,
			HistorySemantics:  NewFieldHistoryTargetSemantics(fields[targetKind], addressability),
			RollbackProvider:  stubNonRowProvider{},
		}
	}
	return []ProviderContribution{
		{SourceOwnerModule: SourceOwnerArtifacts, Records: []RecordProviderContribution{record(SourceOwnerArtifacts, "artifact")}},
		{SourceOwnerModule: SourceOwnerAssessments, Records: []RecordProviderContribution{record(SourceOwnerAssessments, "assessment", "assessment")}},
		{SourceOwnerModule: SourceOwnerEntities, Records: []RecordProviderContribution{record(SourceOwnerEntities, "host", "host"), record(SourceOwnerEntities, "identity", "identity")}, NonRowTargets: []NonRowProviderContribution{nonRow(SourceOwnerEntities, "entity_alias"), nonRow(SourceOwnerEntities, "entity_mention"), nonRow(SourceOwnerEntities, "entity_preserved_identifier")}},
		{SourceOwnerModule: SourceOwnerEvidence, Records: []RecordProviderContribution{record(SourceOwnerEvidence, "evidence", "evidence")}},
		{SourceOwnerModule: SourceOwnerIndicators, Records: []RecordProviderContribution{record(SourceOwnerIndicators, "indicator", "indicator")}, NonRowTargets: []NonRowProviderContribution{nonRow(SourceOwnerIndicators, "indicator_observation"), nonRow(SourceOwnerIndicators, "indicator_state_interval")}},
		{SourceOwnerModule: SourceOwnerLinks, NonRowTargets: []NonRowProviderContribution{nonRow(SourceOwnerLinks, "record_link"), nonRow(SourceOwnerLinks, "record_tag")}},
		{SourceOwnerModule: SourceOwnerParties, Records: []RecordProviderContribution{record(SourceOwnerParties, "party")}},
		{SourceOwnerModule: SourceOwnerTasksDecisions, Records: []RecordProviderContribution{record(SourceOwnerTasksDecisions, "task_request"), record(SourceOwnerTasksDecisions, "decision")}},
		{SourceOwnerModule: SourceOwnerTimeline, Records: []RecordProviderContribution{record(SourceOwnerTimeline, "timeline_event", "timeline_record")}},
	}
}

func validTargetSemanticsRequirements() []TargetSemanticsRequirement {
	return []TargetSemanticsRequirement{
		{TargetKind: "assessment", SourceOwnerID: "assessments", DispatchClass: RollbackDispatchRow, AdmittedRecordTypes: []string{"assessment"}, Addressability: HistorySingleEntry},
		{TargetKind: "entity_alias", SourceOwnerID: "entities", DispatchClass: RollbackDispatchNonRow, HistoryRecordIDFields: []string{"record_id"}, Addressability: HistoryNotIndividuallyAddressable},
		{TargetKind: "entity_mention", SourceOwnerID: "entities", DispatchClass: RollbackDispatchNonRow, HistoryRecordIDFields: []string{"source_record_id"}, Addressability: HistorySingleEntry},
		{TargetKind: "entity_preserved_identifier", SourceOwnerID: "entities", DispatchClass: RollbackDispatchNonRow, HistoryRecordIDFields: []string{"record_id"}, Addressability: HistoryNotIndividuallyAddressable},
		{TargetKind: "evidence", SourceOwnerID: "evidence", DispatchClass: RollbackDispatchRow, AdmittedRecordTypes: []string{"evidence"}, Addressability: HistorySingleEntry},
		{TargetKind: "host", SourceOwnerID: "entities", DispatchClass: RollbackDispatchRow, AdmittedRecordTypes: []string{"host"}, Addressability: HistorySingleEntry},
		{TargetKind: "identity", SourceOwnerID: "entities", DispatchClass: RollbackDispatchRow, AdmittedRecordTypes: []string{"identity"}, Addressability: HistorySingleEntry},
		{TargetKind: "indicator", SourceOwnerID: "indicators", DispatchClass: RollbackDispatchRow, AdmittedRecordTypes: []string{"indicator"}, Addressability: HistorySingleEntry},
		{TargetKind: "indicator_observation", SourceOwnerID: "indicators", DispatchClass: RollbackDispatchNonRow, HistoryRecordIDFields: []string{"source_record_id", "resolved_indicator_record_id"}, Addressability: HistorySingleEntry},
		{TargetKind: "indicator_state_interval", SourceOwnerID: "indicators", DispatchClass: RollbackDispatchNonRow, HistoryRecordIDFields: []string{"indicator_record_id"}, Addressability: HistorySingleEntry},
		{TargetKind: "record", SourceOwnerID: "record_source_owner", DispatchClass: RollbackDispatchRow, AdmittedRecordTypes: []string{"artifact", "assessment", "decision", "evidence", "host", "identity", "indicator", "party", "task_request", "timeline_event"}, Addressability: HistorySingleEntry},
		{TargetKind: "record_link", SourceOwnerID: "links", DispatchClass: RollbackDispatchNonRow, HistoryRecordIDFields: []string{"dst_record_id", "src_record_id"}, Addressability: HistorySingleEntry},
		{TargetKind: "record_tag", SourceOwnerID: "links", DispatchClass: RollbackDispatchNonRow, HistoryRecordIDFields: []string{"record_id"}, Addressability: HistorySingleEntry},
		{TargetKind: "timeline_record", SourceOwnerID: "timeline", DispatchClass: RollbackDispatchRow, AdmittedRecordTypes: []string{"timeline_event"}, Addressability: HistorySingleEntry},
	}
}

func validTargetSemanticsCatalog(t testing.TB, contributions []ProviderContribution) *TargetSemanticsCatalog {
	t.Helper()
	catalog, err := NewTargetSemanticsCatalog(validTargetSemanticsRequirements(), contributions)
	if err != nil {
		t.Fatalf("build target-semantics catalog: %v", err)
	}
	return catalog
}

func TestCommandServiceRejectsInvalidProviderContributionSets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func([]ProviderContribution) []ProviderContribution
		want   error
	}{
		{name: "missing owner", mutate: func(values []ProviderContribution) []ProviderContribution { return values[1:] }, want: ErrMissingProviderContribution},
		{name: "duplicate owner", mutate: func(values []ProviderContribution) []ProviderContribution { return append(values, values[0]) }, want: ErrDuplicateProviderContribution},
		{name: "unexpected owner", mutate: func(values []ProviderContribution) []ProviderContribution {
			values[0].SourceOwnerModule = "unknown"
			return values
		}, want: ErrUnexpectedProviderContribution},
		{name: "wrong owner", mutate: func(values []ProviderContribution) []ProviderContribution {
			values[0].Records[0].RecordType = "party"
			return values
		}, want: ErrUnexpectedProviderContribution},
		{name: "mismatched record owner declaration", mutate: func(values []ProviderContribution) []ProviderContribution {
			values[0].Records[0].SourceOwnerModule = SourceOwnerParties
			return values
		}, want: ErrUnexpectedProviderContribution},
		{name: "mismatched non-row owner declaration", mutate: func(values []ProviderContribution) []ProviderContribution {
			values[2].NonRowTargets[0].SourceOwnerModule = SourceOwnerLinks
			return values
		}, want: ErrUnexpectedProviderContribution},
		{name: "nil delete restore", mutate: func(values []ProviderContribution) []ProviderContribution {
			values[0].Records[0].DeleteRestoreSource = nil
			return values
		}, want: ErrMissingDeleteRestoreSource},
		{name: "nil row rollback", mutate: func(values []ProviderContribution) []ProviderContribution {
			values[0].Records[0].RowRollbackProvider = nil
			return values
		}, want: ErrInvalidTargetSemantics},
		{name: "nil non-row rollback", mutate: func(values []ProviderContribution) []ProviderContribution {
			values[2].NonRowTargets[0].RollbackProvider = nil
			return values
		}, want: ErrInvalidTargetSemantics},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			contributions := test.mutate(validProviderContributions())
			if err := ValidateProviderContributions(contributions); !errors.Is(err, test.want) {
				t.Fatalf("provider contribution error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestIncidentBundleValidationCatalogFailsClosed(t *testing.T) {
	t.Parallel()
	contributions := validProviderContributions()
	targets := validTargetSemanticsCatalog(t, contributions)
	catalog, err := NewIncidentBundleValidationCatalog(incidentBundleEnvelopeReaderStub{}, targets, contributions)
	if err != nil {
		t.Fatalf("build valid incident-bundle validation catalog: %v", err)
	}
	if !catalog.resolvesTargetKind("record") ||
		!catalog.resolvesTargetKind("host") ||
		!catalog.resolvesTargetKind("record_link") {
		t.Fatalf("validation target catalog is incomplete: %#v", catalog.targetKinds())
	}
	contributions[0].Records[0].RecordType = "mutated-after-build"
	if catalog.resolvesTargetKind("mutated-after-build") {
		t.Fatal("caller mutation escaped the immutable validation catalog")
	}

	if _, err := NewIncidentBundleValidationCatalog(incidentBundleEnvelopeReaderStub{}, nil, contributions); !errors.Is(err, ErrMissingHistoryTargetProvider) {
		t.Fatalf("missing target semantics error = %v", err)
	}
}
