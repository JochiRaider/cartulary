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
		{name: "provider contributions", mutate: func(value *CommandServiceDependencies) { value.ProviderContributions = nil }},
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
	return CommandServiceDependencies{
		Database:                    commandServiceTestDB{},
		ImportedAttributionResolver: fakeImportedAttributionResolver{},
		ProjectionRebuilder:         commandServiceTestProjection{},
		ProviderContributions:       validProviderContributions(),
	}
}

func validProviderContributions() []ProviderContribution {
	record := func(owner SourceOwnerModule, recordType string) RecordProviderContribution {
		return RecordProviderContribution{
			SourceOwnerModule:     owner,
			RecordType:            recordType,
			DeleteRestoreProvider: recorddeleterestore.TableProvider{},
			RowRollbackProvider:   catalogRowProvider{},
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
			values[0].Records[0].DeleteRestoreProvider = nil
			return values
		}, want: ErrMissingDeleteRestoreProvider},
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
