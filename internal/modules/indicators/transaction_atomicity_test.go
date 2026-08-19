package indicators

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	indicatorprovider "github.com/JochiRaider/cartulary/internal/modules/indicators/projectionprovider"
	indicatorcontract "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	projectionfixture "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport/fixturewriter"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestIndicatorWorkflowRollsBackRepositoryWritesOnRevisionFailure_Integration(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	db := postgresHarness.BeginRollbackDBT(t, "indicator-repository-atomicity")
	owner, err := NewStore(StoreDependencies{
		Postgres:    db,
		Revisions:   &revisions.Appender{},
		Projections: newTransactionTestProjectionPort(t, db),
		SourceText:  transactionTestSourceTextPort{},
	})
	if err != nil {
		t.Fatalf("compose Indicators owner: %v", err)
	}
	actor := authstoretest.SeedLocalUserRecord(t, db, "indicator-atomicity@example.test", "Indicator Atomicity", "IndicatorAtomicity1!", false, false, true)
	now := time.Date(2026, 8, 3, 19, 0, 0, 0, time.UTC)
	incidentResult, err := incidents.NewApplication(db, workbookstartuppostgres.NewWriter()).CreateIncident(ctx, actor, incidents.CreateIncidentRequest{
		ClientTxnID: "txn-indicator-atomicity-incident",
		IncidentKey: "IR-IND-ATOMICITY",
		Title:       "Indicator repository atomicity",
	}, []byte("txn-indicator-atomicity-incident"), "request-indicator-atomicity-incident", now)
	if err != nil {
		t.Fatalf("create incident: %v", err)
	}
	incidentID := incidentResult.Incident.ID

	owner.revisionsStore = &failingIndicatorRevisionPort{failAt: 3}
	failedCommand := CreateCommand{
		ClientTxnID:   "txn-indicator-atomicity-failed-create",
		IndicatorType: "domain_name",
		ValueKind:     "atomic",
		DisplayValue:  "failed.example",
	}
	if _, createErr := owner.CreateIndicatorRow(ctx, actor, incidentID, failedCommand, []byte("failed-create"), "request-failed-create", now); !errors.Is(createErr, errInjectedIndicatorRevision) {
		t.Fatalf("create revision failure = %v", createErr)
	}
	requireIndicatorAtomicCounts(t, db, incidentID, failedCommand.ClientTxnID, 0, 0, 0, 0, 0, 0)

	owner.projections = failingIndicatorProjectionPort{}
	projectionFailureCommand := CreateCommand{
		ClientTxnID:   "txn-indicator-atomicity-projection-failure",
		IndicatorType: "domain_name",
		ValueKind:     "atomic",
		DisplayValue:  "projection-failure.example",
	}
	if _, projectionErr := owner.CreateIndicatorRow(ctx, actor, incidentID, projectionFailureCommand, []byte("projection-failure"), "request-projection-failure", now.Add(30*time.Second)); !errors.Is(projectionErr, errInjectedIndicatorProjection) {
		t.Fatalf("create projection failure = %v", projectionErr)
	}
	requireIndicatorAtomicCounts(t, db, incidentID, projectionFailureCommand.ClientTxnID, 0, 0, 0, 0, 0, 0)
	owner.projections = newTransactionTestProjectionPort(t, db)

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin baseline source transaction: %v", err)
	}
	created, _, _, _, err := owner.upsertIndicatorTx(ctx, tx, actor, incidentID, CreateCommand{
		ClientTxnID:   "txn-indicator-atomicity-source",
		IndicatorType: "domain_name",
		ValueKind:     "atomic",
		DisplayValue:  "source.example",
	}, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("create baseline Indicator: %v", err)
	}
	if err := owner.refreshProjectionRowTx(ctx, tx, created.RecordID); err != nil {
		t.Fatalf("refresh baseline Indicator projection: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit baseline Indicator: %v", err)
	}

	owner.revisionsStore = &failingIndicatorRevisionPort{failAt: 1}
	if _, observationErr := owner.CreateIndicatorObservation(ctx, actor, IndicatorObservationCreateParams{
		IncidentID:                incidentID,
		SourceRecordID:            created.RecordID,
		BaseRowVersion:            1,
		SourceFieldKey:            "indicator.display_value",
		SpanStartByte:             0,
		SpanEndByte:               len("source.example"),
		ResolvedIndicatorRecordID: &created.RecordID,
		ClientTxnID:               "txn-indicator-atomicity-observation",
		RequestID:                 "request-indicator-atomicity-observation",
		RequestHash:               []byte("indicator-atomicity-observation"),
	}); !errors.Is(observationErr, errInjectedIndicatorRevision) {
		t.Fatalf("observation revision failure = %v", observationErr)
	}
	if got := countIndicatorRows(t, db, `SELECT count(*) FROM indicator_observations WHERE incident_id = $1`, incidentID); got != 0 {
		t.Fatalf("failed observation left %d rows", got)
	}

	owner.revisionsStore = &failingIndicatorRevisionPort{failAt: 1}
	if _, lifecycleErr := owner.AppendIndicatorLifecycleInterval(ctx, actor, IndicatorLifecycleAppendParams{
		IncidentID:        incidentID,
		IndicatorRecordID: created.RecordID,
		BaseRowVersion:    1,
		LifecycleState:    "active",
		ValidFrom:         now.Add(3 * time.Minute),
		SupportRefs:       []uuid.UUID{},
		ClientTxnID:       "txn-indicator-atomicity-lifecycle",
		RequestID:         "request-indicator-atomicity-lifecycle",
		RequestHash:       []byte("indicator-atomicity-lifecycle"),
	}); !errors.Is(lifecycleErr, errInjectedIndicatorRevision) {
		t.Fatalf("lifecycle revision failure = %v", lifecycleErr)
	}
	if got := countIndicatorRows(t, db, `SELECT count(*) FROM indicator_state_intervals WHERE incident_id = $1`, incidentID); got != 0 {
		t.Fatalf("failed lifecycle append left %d rows", got)
	}
}

var errInjectedIndicatorRevision = errors.New("injected Indicator revision failure")
var errInjectedIndicatorProjection = errors.New("injected Indicator projection failure")

type failingIndicatorProjectionPort struct{}

func (failingIndicatorProjectionPort) RefreshIndicatorTx(context.Context, pgx.Tx, uuid.UUID) error {
	return errInjectedIndicatorProjection
}

func (failingIndicatorProjectionPort) LoadIndicatorTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error) {
	panic("LoadIndicatorTx must not follow a failed projection refresh")
}

func (failingIndicatorProjectionPort) DeleteIndicatorTx(context.Context, pgx.Tx, uuid.UUID) error {
	panic("unexpected DeleteIndicatorTx")
}

func (failingIndicatorProjectionPort) RebuildIndicatorsTx(context.Context, pgx.Tx, uuid.UUID) error {
	panic("unexpected RebuildIndicatorsTx")
}

type transactionTestSourceTextPort struct{}

func (transactionTestSourceTextPort) LoadTextTx(context.Context, pgx.Tx, uuid.UUID, string, string) (SourceTextValue, error) {
	row := map[string]any{"cells": map[string]any{}}
	return SourceTextValue{ViewSchemaID: ViewSchemaID, Text: "source.example", Row: row}, nil
}

func (transactionTestSourceTextPort) LoadRowTx(context.Context, pgx.Tx, uuid.UUID, string, string) (map[string]any, error) {
	return map[string]any{"view_schema_id": "cartulary.view.timeline.v2", "cells": map[string]any{}}, nil
}

func (transactionTestSourceTextPort) RefreshAndLoadRowTx(context.Context, pgx.Tx, uuid.UUID, string, string) (map[string]any, error) {
	return map[string]any{"view_schema_id": "cartulary.view.timeline.v2", "cells": map[string]any{}}, nil
}

func newTransactionTestProjectionPort(t testing.TB, db postgres.DB) indicatorcontract.Rows {
	t.Helper()
	contribution, err := indicatorprovider.NewContribution()
	if err != nil {
		t.Fatalf("construct Indicator projection contribution: %v", err)
	}
	return projectionfixture.NewIndicatorFixtureWriter(t, db, contribution.Source())
}

type failingIndicatorRevisionPort struct {
	calls  int
	failAt int
}

func (port *failingIndicatorRevisionPort) shouldFail() error {
	port.calls++
	if port.calls == port.failAt {
		return errInjectedIndicatorRevision
	}
	return nil
}

func (port *failingIndicatorRevisionPort) AppendChangeSetTx(_ context.Context, _ pgx.Tx, _ revisions.AppendChangeSetParams) (uuid.UUID, error) {
	if err := port.shouldFail(); err != nil {
		return uuid.Nil, err
	}
	return uuid.New(), nil
}

func (port *failingIndicatorRevisionPort) CaptureRecordSnapshotTx(_ context.Context, _ pgx.Tx, _ uuid.UUID) (revisions.RecordSnapshot, error) {
	if err := port.shouldFail(); err != nil {
		return revisions.RecordSnapshot{}, err
	}
	return revisions.RecordSnapshot{}, nil
}

func (port *failingIndicatorRevisionPort) AppendMutationTx(_ context.Context, _ pgx.Tx, _ revisions.AppendNonRowMutationParams) error {
	return port.shouldFail()
}

func (port *failingIndicatorRevisionPort) AppendRecordMutationTx(_ context.Context, _ pgx.Tx, _ revisions.AppendRecordMutationParams) error {
	return port.shouldFail()
}

func (port *failingIndicatorRevisionPort) AppendRecordRevisionAndIntentTx(_ context.Context, _ pgx.Tx, _ revisions.AppendRecordRevisionParams) error {
	return port.shouldFail()
}

func requireIndicatorAtomicCounts(
	t testing.TB,
	db postgres.DB,
	incidentID uuid.UUID,
	clientTxnID string,
	recordsWant int,
	sourcesWant int,
	projectionsWant int,
	changeSetsWant int,
	mutationsWant int,
	idempotencyWant int,
) {
	t.Helper()
	checks := []struct {
		query string
		args  []any
		want  int
	}{
		{`SELECT count(*) FROM records WHERE incident_id = $1 AND record_type = 'indicator'`, []any{incidentID}, recordsWant},
		{`SELECT count(*) FROM indicators WHERE incident_id = $1`, []any{incidentID}, sourcesWant},
		{`SELECT count(*) FROM indicator_active_identities WHERE incident_id = $1`, []any{incidentID}, sourcesWant},
		{`SELECT count(*) FROM indicator_grid_projection WHERE incident_id = $1`, []any{incidentID}, projectionsWant},
		{`SELECT count(*) FROM change_sets WHERE incident_id = $1`, []any{incidentID}, changeSetsWant},
		{`SELECT count(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id = $1`, []any{incidentID}, mutationsWant},
		{`SELECT count(*) FROM route_idempotency WHERE client_txn_id = $1`, []any{clientTxnID}, idempotencyWant},
	}
	for _, check := range checks {
		if got := countIndicatorRows(t, db, check.query, check.args...); got != check.want {
			t.Fatalf("count for %q = %d, want %d", check.query, got, check.want)
		}
	}
}

func countIndicatorRows(t testing.TB, db postgres.DB, query string, args ...any) int {
	t.Helper()
	var got int
	if err := db.QueryRow(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("query Indicator count: %v", err)
	}
	return got
}
