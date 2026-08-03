package revisionassembly

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
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

func TestRevisionsRuntimeBuildsExactImmutableRecordViewCatalog(t *testing.T) {
	t.Parallel()
	contributions := CurrentProviderContributions()
	runtime, err := Build(
		Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         collaboration.NewIntentAppender(),
		},
		contributions...,
	)
	if err != nil {
		t.Fatalf("build Revisions runtime: %v", err)
	}

	type route struct {
		recordType   string
		variant      string
		viewSchemaID string
	}
	want := []route{
		{recordType: "artifact", variant: "comm_log", viewSchemaID: "cartulary.view.comm_log.v1"},
		{recordType: "artifact", variant: "finding", viewSchemaID: "cartulary.view.findings.v1"},
		{recordType: "artifact", variant: "forensic_keyword", viewSchemaID: "cartulary.view.forensic_keywords.v1"},
		{recordType: "artifact", variant: "handoff", viewSchemaID: "cartulary.view.handoff.v1"},
		{recordType: "artifact", variant: "investigative_query", viewSchemaID: "cartulary.view.investigative_queries.v1"},
		{recordType: "artifact", variant: "lesson", viewSchemaID: "cartulary.view.lesson.v1"},
		{recordType: "artifact", variant: "note", viewSchemaID: "cartulary.view.notes.v1"},
		{recordType: "artifact", variant: "status_review", viewSchemaID: "cartulary.view.status_review.v1"},
		{recordType: "assessment", viewSchemaID: "cartulary.view.assessments.v1"},
		{recordType: "decision", viewSchemaID: "cartulary.view.decisions.v1"},
		{recordType: "evidence", viewSchemaID: "cartulary.view.evidence.v1"},
		{recordType: "host", viewSchemaID: "cartulary.view.hosts.v1"},
		{recordType: "identity", viewSchemaID: "cartulary.view.identities.v1"},
		{recordType: "indicator", viewSchemaID: "cartulary.view.indicators.v1"},
		{recordType: "party", viewSchemaID: "cartulary.view.parties.v1"},
		{recordType: "task_request", viewSchemaID: "cartulary.view.task_requests.v1"},
		{recordType: "timeline_event", viewSchemaID: "cartulary.view.timeline.v2"},
	}
	gotDescriptors := runtime.RecordViewCatalog().Descriptors()
	got := make([]route, 0, len(gotDescriptors))
	for _, descriptor := range gotDescriptors {
		variant := ""
		if descriptor.Variant != nil {
			if descriptor.Variant.Kind != "artifact_type" {
				t.Fatalf("unexpected variant kind in %#v", descriptor)
			}
			variant = descriptor.Variant.Value
		}
		if len(descriptor.ViewSchemaIDs) != 1 {
			t.Fatalf("route must identify exactly one view: %#v", descriptor)
		}
		got = append(got, route{
			recordType:   descriptor.RecordType,
			variant:      variant,
			viewSchemaID: descriptor.ViewSchemaIDs[0],
		})
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("record/view routes = %#v, want %#v", got, want)
	}

	contributions[0].Records[0].RecordViewRoutes[0].ViewSchemaIDs[0] = "mutated.input"
	gotDescriptors[0].Variant.Value = "mutated.output"
	gotDescriptors[0].ViewSchemaIDs[0] = "mutated.output"
	next := runtime.RecordViewCatalog().Descriptors()
	if next[0].Variant.Value != "comm_log" ||
		next[0].ViewSchemaIDs[0] != "cartulary.view.comm_log.v1" {
		t.Fatalf("record/view catalog was mutated through an input or output slice: %#v", next[0])
	}

	if got, err := runtime.RecordViewCatalog().Resolve("artifact", map[string]any{
		"cells": map[string]any{
			"note.title":    map[string]any{"value": "later"},
			"comm_log.time": map[string]any{"value": "earlier"},
		},
		"source": map[string]any{"artifact_type": "note"},
	}); err != nil || got != "cartulary.view.comm_log.v1" {
		t.Fatalf("resolve sorted artifact prefix = %q, %v", got, err)
	}
	if got, err := runtime.RecordViewCatalog().Resolve("artifact", map[string]any{
		"source": map[string]any{"artifact_type": "note"},
	}); err != nil || got != "cartulary.view.notes.v1" {
		t.Fatalf("resolve deleted artifact source type = %q, %v", got, err)
	}
}

func TestRevisionsRuntimeRejectsIncompleteOrAmbiguousRecordViewCatalogs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func([]revisions.ProviderContribution) []revisions.ProviderContribution
		want   error
	}{
		{
			name: "missing conflict field provider",
			mutate: func(values []revisions.ProviderContribution) []revisions.ProviderContribution {
				values[0].ConflictFieldProvider = nil
				return values
			},
			want: conflicts.ErrMissingFieldResolver,
		},
		{
			name: "missing route",
			mutate: func(values []revisions.ProviderContribution) []revisions.ProviderContribution {
				values[0].Records[0].RecordViewRoutes = values[0].Records[0].RecordViewRoutes[1:]
				return values
			},
			want: revisions.ErrMissingRecordViewRoute,
		},
		{
			name: "duplicate contribution id",
			mutate: func(values []revisions.ProviderContribution) []revisions.ProviderContribution {
				duplicate := values[0].Records[0].RecordViewRoutes[0]
				duplicate.Variant = &revisions.RecordVariant{Kind: "artifact_type", Value: "new_type"}
				values[0].Records[0].RecordViewRoutes = append(values[0].Records[0].RecordViewRoutes, duplicate)
				return values
			},
			want: revisions.ErrDuplicateRecordViewRoute,
		},
		{
			name: "duplicate view",
			mutate: func(values []revisions.ProviderContribution) []revisions.ProviderContribution {
				duplicate := values[0].Records[0].RecordViewRoutes[0]
				duplicate.ContributionID = "artifacts.duplicate_view"
				duplicate.Variant = &revisions.RecordVariant{Kind: "artifact_type", Value: "new_type"}
				values[0].Records[0].RecordViewRoutes = append(values[0].Records[0].RecordViewRoutes, duplicate)
				return values
			},
			want: revisions.ErrDuplicateRecordViewRoute,
		},
		{
			name: "unexpected view for record",
			mutate: func(values []revisions.ProviderContribution) []revisions.ProviderContribution {
				values[1].Records[0].RecordViewRoutes[0].ViewSchemaIDs[0] = "cartulary.view.parties.v1"
				return values
			},
			want: revisions.ErrUnexpectedRecordViewRoute,
		},
		{
			name: "unknown view",
			mutate: func(values []revisions.ProviderContribution) []revisions.ProviderContribution {
				values[1].Records[0].RecordViewRoutes[0].ViewSchemaIDs[0] = "cartulary.view.unknown.v1"
				return values
			},
			want: revisions.ErrUnknownRecordViewSchema,
		},
		{
			name: "unsupported variant",
			mutate: func(values []revisions.ProviderContribution) []revisions.ProviderContribution {
				values[1].Records[0].RecordViewRoutes[0].Variant = &revisions.RecordVariant{
					Kind:  "artifact_type",
					Value: "assessment",
				}
				return values
			},
			want: revisions.ErrUnsupportedRecordVariant,
		},
		{
			name: "ambiguous selector",
			mutate: func(values []revisions.ProviderContribution) []revisions.ProviderContribution {
				duplicate := values[0].Records[0].RecordViewRoutes[0]
				duplicate.ContributionID = "artifacts.ambiguous"
				values[0].Records[0].RecordViewRoutes = append(values[0].Records[0].RecordViewRoutes, duplicate)
				return values
			},
			want: revisions.ErrAmbiguousRecordViewRoute,
		},
		{
			name: "required policy without routes",
			mutate: func(values []revisions.ProviderContribution) []revisions.ProviderContribution {
				values[1].Records[0].RecordViewRoutes = nil
				return values
			},
			want: revisions.ErrMissingRecordViewRoute,
		},
		{
			name: "none policy with routes",
			mutate: func(values []revisions.ProviderContribution) []revisions.ProviderContribution {
				values[1].Records[0].LiveRecordChangePolicy = revisions.LiveRecordChangeNone
				return values
			},
			want: revisions.ErrUnexpectedRecordViewRoute,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := Build(
				Dependencies{
					HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
					IntentAppender:         collaboration.NewIntentAppender(),
				},
				test.mutate(cloneProviderContributions(CurrentProviderContributions()))...,
			)
			if !errors.Is(err, test.want) {
				t.Fatalf("build error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestRevisionsRuntimeReusesAppenderForCommandService(t *testing.T) {
	t.Parallel()
	runtime, err := Build(
		Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         collaboration.NewIntentAppender(),
		},
		CurrentProviderContributions()...,
	)
	if err != nil {
		t.Fatalf("build Revisions runtime: %v", err)
	}
	service, err := runtime.NewCommandService(
		revisionsCompositionTestDB{},
		revisionsCompositionTestAttribution{},
		revisionsCompositionTestProjection{},
		func() time.Time { return time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC) },
	)
	if err != nil {
		t.Fatalf("compose Revisions command service: %v", err)
	}
	if service == nil || runtime.Appender() == nil {
		t.Fatal("composition returned a nil service or appender")
	}
}

func TestRevisionsRuntimeBuildsOwnerComposedConflictFieldResolver(t *testing.T) {
	t.Parallel()
	runtime, err := Build(
		Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         collaboration.NewIntentAppender(),
		},
		CurrentProviderContributions()...,
	)
	if err != nil {
		t.Fatalf("build Revisions runtime: %v", err)
	}
	resolver := runtime.ConflictFieldResolver()
	field, err := resolver.ResolveWritableField("cartulary.view.notes.v1", "note.body")
	if err != nil {
		t.Fatalf("resolve Notes body field: %v", err)
	}
	if field.FieldKey != "note.body" || field.ValueKind != "direct_value" || field.ConflictResolutionClass != "text_compare_merge" {
		t.Fatalf("Notes body field = %#v", field)
	}
	if _, err := resolver.ResolveViewSchema("cartulary.view.unknown.v1"); err == nil {
		t.Fatal("unknown view schema resolved")
	}
}
