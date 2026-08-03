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
		{name: "provider contributions", mutate: func(value *CommandServiceDependencies) { value.ProviderContributions = nil }},
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
	return CommandServiceDependencies{
		Transactions:                database,
		Authorization:               commandServiceTestAuthorizer{},
		Idempotency:                 commandServiceTestIdempotency{},
		ImportedAttributionResolver: fakeImportedAttributionResolver{},
		Projections:                 commandServiceTestProjection{},
		ProviderContributions:       validProviderContributions(),
		Appender: &Appender{
			recordViews:      &RecordViewCatalog{},
			historicalPolicy: commandServiceTestHistoricalPolicy{},
		},
		RecordEnvelopes: commandServiceTestEnvelopes{},
		Clock:           func() time.Time { return time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC) },
	}
}

func validProviderContributions() []ProviderContribution {
	record := func(owner SourceOwnerModule, recordType string) RecordProviderContribution {
		return RecordProviderContribution{
			SourceOwnerModule:   owner,
			RecordType:          recordType,
			DeleteRestoreSource: testDeleteRestoreSource{},
			RowRollbackProvider: catalogRowProvider{},
		}
	}
	nonRow := func(owner SourceOwnerModule, targetKind string) NonRowProviderContribution {
		return NonRowProviderContribution{SourceOwnerModule: owner, TargetKind: targetKind, RollbackProvider: stubNonRowProvider{}}
	}
	return []ProviderContribution{
		{SourceOwnerModule: SourceOwnerArtifacts, Records: []RecordProviderContribution{record(SourceOwnerArtifacts, "artifact")}},
		{SourceOwnerModule: SourceOwnerAssessments, Records: []RecordProviderContribution{record(SourceOwnerAssessments, "assessment")}},
		{SourceOwnerModule: SourceOwnerEntities, Records: []RecordProviderContribution{record(SourceOwnerEntities, "host"), record(SourceOwnerEntities, "identity")}, NonRowTargets: []NonRowProviderContribution{nonRow(SourceOwnerEntities, "entity_alias"), nonRow(SourceOwnerEntities, "entity_mention"), nonRow(SourceOwnerEntities, "entity_preserved_identifier")}},
		{SourceOwnerModule: SourceOwnerEvidence, Records: []RecordProviderContribution{record(SourceOwnerEvidence, "evidence")}},
		{SourceOwnerModule: SourceOwnerIndicators, Records: []RecordProviderContribution{record(SourceOwnerIndicators, "indicator")}, NonRowTargets: []NonRowProviderContribution{nonRow(SourceOwnerIndicators, "indicator_observation"), nonRow(SourceOwnerIndicators, "indicator_state_interval")}},
		{SourceOwnerModule: SourceOwnerLinks, NonRowTargets: []NonRowProviderContribution{nonRow(SourceOwnerLinks, "record_link"), nonRow(SourceOwnerLinks, "record_tag")}},
		{SourceOwnerModule: SourceOwnerParties, Records: []RecordProviderContribution{record(SourceOwnerParties, "party")}},
		{SourceOwnerModule: SourceOwnerTasksDecisions, Records: []RecordProviderContribution{record(SourceOwnerTasksDecisions, "task_request"), record(SourceOwnerTasksDecisions, "decision")}},
		{SourceOwnerModule: SourceOwnerTimeline, Records: []RecordProviderContribution{record(SourceOwnerTimeline, "timeline_event")}},
	}
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
		}, want: ErrMissingRowRollbackProvider},
		{name: "nil non-row rollback", mutate: func(values []ProviderContribution) []ProviderContribution {
			values[2].NonRowTargets[0].RollbackProvider = nil
			return values
		}, want: ErrMissingNonRowRollbackProvider},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dependencies := validCommandServiceDependencies(t)
			dependencies.ProviderContributions = test.mutate(dependencies.ProviderContributions)
			if _, err := NewCommandService(dependencies); !errors.Is(err, test.want) {
				t.Fatalf("provider contribution error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestIncidentBundleValidationCatalogFailsClosed(t *testing.T) {
	t.Parallel()
	contributions := validProviderContributions()
	catalog, err := NewIncidentBundleValidationCatalog(incidentBundleEnvelopeReaderStub{}, contributions)
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

	duplicate := validProviderContributions()
	duplicate[0].Records[0].HistoryTargetKinds = []string{"shared_target"}
	duplicate[1].Records[0].HistoryTargetKinds = []string{"shared_target"}
	if _, err := NewIncidentBundleValidationCatalog(incidentBundleEnvelopeReaderStub{}, duplicate); !errors.Is(err, ErrDuplicateHistoryTargetProvider) {
		t.Fatalf("duplicate target provider error = %v", err)
	}

	empty := validProviderContributions()
	empty[0].Records[0].HistoryTargetKinds = []string{""}
	if _, err := NewIncidentBundleValidationCatalog(incidentBundleEnvelopeReaderStub{}, empty); !errors.Is(err, ErrUnexpectedProviderContribution) {
		t.Fatalf("empty target provider error = %v", err)
	}
}
