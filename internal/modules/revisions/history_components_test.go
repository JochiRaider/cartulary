package revisions

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestHistoryComponentsMaterializeDecorateAndOrder_Unit(t *testing.T) {
	record := RecordHistoryRecord{IncidentID: uuid.New(), RecordID: uuid.New(), RecordType: "host", RowVersion: 3}
	actorID := uuid.New()
	olderChangeSet := uuid.New()
	newerChangeSet := uuid.New()
	base := time.Date(2026, 8, 3, 12, 0, 0, 0, time.FixedZone("fixture", -4*60*60))
	materializer := historyRowMaterializer{}
	older := materializer.Mutation(record, mutationHistoryRow{
		ChangeSetID:   olderChangeSet,
		ActorUserID:   actorID,
		CommittedAt:   base,
		Source:        "workbook.records.patch",
		SequenceNo:    1,
		TargetKind:    "host",
		TargetID:      record.RecordID.String(),
		OperationKind: "field_update",
		BeforeValue:   []byte(`{"name":"before"}`),
		AfterValue:    []byte(`{"name":"after"}`),
	})
	newer := materializer.Mutation(record, mutationHistoryRow{
		ChangeSetID:   newerChangeSet,
		ActorUserID:   actorID,
		CommittedAt:   base.Add(time.Minute),
		Source:        "workbook.records.patch",
		SequenceNo:    1,
		TargetKind:    "host",
		TargetID:      record.RecordID.String(),
		OperationKind: "field_update",
	})

	revisionOnlyChangeSet := uuid.New()
	revisionItems := materializer.Revisions(record, []revisionHistoryRow{
		{ChangeSetID: olderChangeSet, ActorUserID: actorID, CommittedAt: base, RevisionNo: 2},
		{ChangeSetID: revisionOnlyChangeSet, ActorUserID: actorID, CommittedAt: base.Add(-time.Minute), RevisionNo: 1},
	}, []RecordHistoryItem{older, newer})
	if len(revisionItems) != 1 || revisionItems[0].ChangeSetID != revisionOnlyChangeSet {
		t.Fatalf("revision-only materialization = %#v", revisionItems)
	}
	if got := older.DiffSummary["summary"]; got != "field_update host" {
		t.Fatalf("mutation summary = %#v", got)
	}

	resolver := &historyAttributionResolverFake{sourceActors: map[string]string{olderChangeSet.String(): "source-user-7"}}
	items := []RecordHistoryItem{older, newer, revisionItems[0]}
	if err := (importedHistoryAttributionDecorator{resolver: resolver}).DecorateTx(context.Background(), nil, record.IncidentID, items); err != nil {
		t.Fatalf("decorate attribution: %v", err)
	}
	if items[0].SourceActorID == nil || *items[0].SourceActorID != "source-user-7" {
		t.Fatalf("source attribution = %#v", items[0].SourceActorID)
	}
	if resolver.sourceTable != "change_sets" || resolver.sourceColumn != "actor_user_id" {
		t.Fatalf("attribution owner address = %s.%s", resolver.sourceTable, resolver.sourceColumn)
	}

	resources := (historyPageAssembler{}).Resources(items)
	if len(resources) != 3 || resources[0]["change_set_id"] != newerChangeSet.String() || resources[2]["change_set_id"] != revisionOnlyChangeSet.String() {
		t.Fatalf("history page order = %#v", resources)
	}
	if got := resources[0]["available_rollback_actions"]; !reflect.DeepEqual(got, []string{}) {
		t.Fatalf("empty actions must remain a JSON array, got %#v", got)
	}
	if got := resources[0]["committed_at"]; got != base.Add(time.Minute).UTC().Format(time.RFC3339Nano) {
		t.Fatalf("UTC committed_at = %#v", got)
	}
}

func TestHistoryDecompositionBoundaries_Unit(t *testing.T) {
	materializer, err := os.ReadFile("history_materializer.go")
	if err != nil {
		t.Fatalf("read history materializer: %v", err)
	}
	for _, sqlToken := range []string{"SELECT ", "INSERT INTO ", "UPDATE ", "DELETE FROM "} {
		if strings.Contains(string(materializer), sqlToken) {
			t.Fatalf("history materializer contains persistence token %q", sqlToken)
		}
	}
	service, err := os.ReadFile("history_service.go")
	if err != nil {
		t.Fatalf("read history service: %v", err)
	}
	for _, transportToken := range []string{"net/http", "internal/platform/authn", "pagination."} {
		if strings.Contains(string(service), transportToken) {
			t.Fatalf("history application service contains transport token %q", transportToken)
		}
	}
	if _, err := os.Stat("store.go"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("mixed history store must remain removed, stat error = %v", err)
	}
}

func TestHistoryQueryRepositoryMapsPersistenceRows_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	var database postgres.DB
	if pgtest.ExplicitPostgresFixturePolicyT(t) == pgtest.PostgresFixturePolicyTemplateClone {
		testDatabase := harness.PrepareIsolatedDatabaseT(t, "revisions-history-repository")
		pool, err := pgxpool.New(context.Background(), testDatabase.DSN)
		if err != nil {
			t.Fatalf("open history repository database: %v", err)
		}
		t.Cleanup(pool.Close)
		database = pool
	} else {
		database = harness.BeginRollbackDBT(t, "revisions-history-repository")
	}
	actor := authstoretest.SeedLocalUserRecord(
		t,
		database,
		"history-repository@example.test",
		"History Repository",
		"HistoryRepositoryPass1!",
		false,
		false,
		true,
	)
	now := time.Date(2026, 8, 3, 16, 45, 0, 123456000, time.UTC)
	incidentResult, err := incidents.NewApplication(database).CreateIncident(context.Background(), actor, incidents.CreateIncidentRequest{
		ClientTxnID: "history-repository-incident",
		IncidentKey: "IR-HISTORY-REPOSITORY",
		Title:       "History repository component",
	}, []byte("history-repository-incident"), "req-history-repository-incident", now)
	if err != nil {
		t.Fatalf("create incident: %v", err)
	}
	recordID := uuid.New()
	envelopetest.SeedRecordEnvelope(t, database, incidentResult.Incident.ID, actor.ID, recordID, "host")
	changeSetID := uuid.New()
	if _, err := database.Exec(context.Background(), `
INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source, created_at)
VALUES ($1, $2, $3, 'history_repository_test', $4)
`, changeSetID, incidentResult.Incident.ID, actor.ID, now); err != nil {
		t.Fatalf("seed change set: %v", err)
	}
	if _, err := database.Exec(context.Background(), `
INSERT INTO change_set_mutations (
    change_set_id, sequence_no, target_kind, target_id, operation_kind, before_value, after_value
) VALUES ($1, 1, 'host', $2, 'field_update', '{"name":"before"}', '{"name":"after"}')
`, changeSetID, recordID.String()); err != nil {
		t.Fatalf("seed change-set mutation: %v", err)
	}
	if _, err := database.Exec(context.Background(), `
INSERT INTO record_revisions (change_set_id, record_id, row_version, before_json, after_json, created_at)
VALUES ($1, $2, 2, '{"cells":{}}', '{"cells":{"host.name":{"value":"after"}}}', $3)
`, changeSetID, recordID, now); err != nil {
		t.Fatalf("seed record revision: %v", err)
	}

	tx, err := database.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin repository transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	repository := historyQueryRepository{}
	record := RecordHistoryRecord{IncidentID: incidentResult.Incident.ID, RecordID: recordID, RecordType: "host", RowVersion: 2}
	mutationRows, err := repository.LoadMutationRowsTx(context.Background(), tx, record)
	if err != nil {
		t.Fatalf("load mutation rows: %v", err)
	}
	if len(mutationRows) != 1 || mutationRows[0].ChangeSetID != changeSetID || mutationRows[0].RevisionNo == nil || *mutationRows[0].RevisionNo != 2 {
		t.Fatalf("mutation repository mapping = %#v", mutationRows)
	}
	revisionRows, err := repository.LoadRevisionRowsTx(context.Background(), tx, record)
	if err != nil {
		t.Fatalf("load revision rows: %v", err)
	}
	if len(revisionRows) != 1 || revisionRows[0].ChangeSetID != changeSetID || revisionRows[0].RevisionNo != 2 {
		t.Fatalf("revision repository mapping = %#v", revisionRows)
	}
	ref, err := repository.EnsureHistoryEntryRefTx(context.Background(), tx, recordID, changeSetID, 1)
	if err != nil {
		t.Fatalf("ensure history entry ref: %v", err)
	}
	repeated, err := repository.EnsureHistoryEntryRefTx(context.Background(), tx, recordID, changeSetID, 1)
	if err != nil || repeated != ref {
		t.Fatalf("stable history entry ref = %q, %v; want %q", repeated, err, ref)
	}
}

type historyAttributionResolverFake struct {
	sourceActors map[string]string
	sourceTable  string
	sourceColumn string
}

func (f *historyAttributionResolverFake) ResolveImportedSourceActorsTx(_ context.Context, _ pgx.Tx, _ uuid.UUID, sourceTable string, sourceColumn string, _ []string) (map[string]string, error) {
	f.sourceTable = sourceTable
	f.sourceColumn = sourceColumn
	return f.sourceActors, nil
}
