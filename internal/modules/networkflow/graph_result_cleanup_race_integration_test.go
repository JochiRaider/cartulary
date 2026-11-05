package networkflow_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/graphfixture"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestNetworkFlowGraphCleanupPublicationAndAdmissionRaces_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "network-flow-graph-cleanup-races")
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	actor := authstoretest.SeedActor(t, pool, authstoretest.ActorAttributes{Email: "network-flow-cleanup-races@example.test", DisplayName: "Cleanup Race Tester"})
	incident := appsupport.CreateIncidentInStore(t, pool, actor, "txn-network-flow-cleanup-races", "IR-NFCLEANRACE", "Network Flow cleanup races")
	store := NewStore(pool, DefaultEffectiveLimits(), WithResourceIntentAppender(networkflowsupport.NewResourceIntentAppender()))
	now := time.Date(2026, 8, 16, 18, 0, 0, 0, time.UTC)

	graphViewID := "nfgv_dddddddddddddddddddddddddddddddd"
	jobID := uuid.New()
	seedGraphCleanupJob(t, ctx, pool, jobID, incident.ID, actor.ID, now)
	declaration := graphViewDeclarationFixture(graphViewID, incident.ID, actor.ID, now)
	declaration.DesiredSourceSnapshotID = "nfsnap_" + testSHA1
	declaration.LatestJobID = &jobID
	insertGraphCleanupDeclaration(t, ctx, pool, store, declaration)
	first := graphResultForReportingSource(t, graphViewID, now)
	publishGraphCleanupResult(t, ctx, pool, first)

	cleanupTx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupTx.Rollback(ctx)
	cleaner, _ := postgresresult.NewCleaner(cleanupTx)
	firstCandidate, err := cleaner.LockCleanupCandidate(ctx, ProfileID, nil)
	if err != nil || firstCandidate == nil || firstCandidate.ProjectionResultID != first.Binding.ProjectionResultID {
		t.Fatalf("lock publication-race cleanup candidate: candidate=%#v err=%v", firstCandidate, err)
	}

	publicationStarted := make(chan struct{})
	publicationDone := make(chan error, 1)
	go func() {
		publicationTx, beginErr := pool.BeginTx(ctx, pgx.TxOptions{})
		if beginErr != nil {
			publicationDone <- beginErr
			return
		}
		defer func() { _ = publicationTx.Rollback(ctx) }()
		publisher, publisherErr := postgresresult.NewPublisher(publicationTx)
		if publisherErr != nil {
			publicationDone <- publisherErr
			return
		}
		close(publicationStarted)
		if publisherErr = publisher.PublishResult(ctx, first); publisherErr != nil {
			publicationDone <- publisherErr
			return
		}
		if _, publisherErr = store.PublishGraphViewResultTx(
			ctx, publicationTx, incident.ID, graphViewID, declaration.MaterializationGeneration,
			declaration.DesiredSourceSnapshotID, jobID, selectedGraphCleanupBinding(first.Binding), now.Add(time.Minute),
		); publisherErr != nil {
			publicationDone <- publisherErr
			return
		}
		publicationDone <- publicationTx.Commit(ctx)
	}()
	<-publicationStarted
	select {
	case err := <-publicationDone:
		t.Fatalf("publication crossed a cleanup-held result lock: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	selected, err := store.LockGraphViewDeclarationsSelectingResultTx(ctx, cleanupTx, first.Binding.ProjectionResultID)
	if err != nil || selected {
		t.Fatalf("pre-publication selected check = %t err=%v", selected, err)
	}
	if deleted, err := cleaner.DeleteLockedResult(ctx, first.Binding.ProjectionResultID); err != nil || !deleted {
		t.Fatalf("delete cleanup-held result: deleted=%t err=%v", deleted, err)
	}
	if err := cleanupTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-publicationDone:
		if err != nil {
			t.Fatalf("publication replay after cleanup failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("publication replay did not resume after cleanup commit")
	}
	requireGraphResultCount(t, pool, first.Binding.ProjectionResultID, 1)
	requireSelectedGraphCleanupResult(t, ctx, store, incident.ID, graphViewID, first.Binding.ProjectionResultID)

	second := graphfixture.Result(t, graphfixture.ResultOptions{EdgeStringProperties: map[string]string{"direction_kind": "source_to_destination", "counter_semantics": "source_reported_flow_total"}, GraphViewID: graphViewID, SourceOwnerID: ProfileID, SourceSnapshotID: "nfsnap_" + testSHA1, ProjectionVersion: "network_flow_activity.v2", VertexKind: "network_flow.ip_endpoint.v1", EdgeKind: "network_flow.flow_edge.v2", VertexCount: 2, VertexLabels: []string{"changed"}, PublishedAt: now.Add(time.Second)})
	publishGraphCleanupResult(t, ctx, pool, second)
	publicationTx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer publicationTx.Rollback(ctx)
	publisher, _ := postgresresult.NewPublisher(publicationTx)
	if err := publisher.PublishResult(ctx, second); err != nil {
		t.Fatalf("lock replayed result before declaration: %v", err)
	}
	concurrentCleanupTx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer concurrentCleanupTx.Rollback(ctx)
	concurrentCleaner, _ := postgresresult.NewCleaner(concurrentCleanupTx)
	lockedWhilePublishing, err := concurrentCleaner.LockCleanupCandidate(ctx, ProfileID, firstCandidate)
	if err != nil || lockedWhilePublishing != nil {
		t.Fatalf("cleanup did not skip publication-held result: candidate=%#v err=%v", lockedWhilePublishing, err)
	}
	if err := concurrentCleanupTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PublishGraphViewResultTx(
		ctx, publicationTx, incident.ID, graphViewID, declaration.MaterializationGeneration,
		declaration.DesiredSourceSnapshotID, jobID, selectedGraphCleanupBinding(second.Binding), now.Add(2*time.Minute),
	); err != nil {
		t.Fatalf("select publication-held result: %v", err)
	}
	if err := publicationTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	requireSelectedGraphCleanupResult(t, ctx, store, incident.ID, graphViewID, second.Binding.ProjectionResultID)

	source, err := NewReportingGraphSource(store, func() bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	admissionTx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer admissionTx.Rollback(ctx)
	retained, err := source.ReadSelectedResultTx(ctx, admissionTx, incident.ID, second.Binding)
	if err != nil {
		t.Fatalf("admit Reporting input: %v", err)
	}
	admissionRaceCleanupTx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer admissionRaceCleanupTx.Rollback(ctx)
	admissionRaceCleaner, _ := postgresresult.NewCleaner(admissionRaceCleanupTx)
	lockedWhileAdmitting, err := admissionRaceCleaner.LockCleanupCandidate(ctx, ProfileID, firstCandidate)
	if err != nil || lockedWhileAdmitting != nil {
		t.Fatalf("cleanup did not skip input-admission-held result: candidate=%#v err=%v", lockedWhileAdmitting, err)
	}
	if err := admissionRaceCleanupTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := admissionTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if retained.Projection.Binding != second.Binding || len(retained.Projection.Vertices) != 2 {
		t.Fatal("input admission omitted selected contribution")
	}

	thirdGraphViewID := "nfgv_eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	third := graphResultForReportingSource(t, thirdGraphViewID, now.Add(2*time.Second))
	publishGraphCleanupResult(t, ctx, pool, third)
	thirdDeclaration := graphViewDeclarationFixture(thirdGraphViewID, incident.ID, actor.ID, now)
	thirdDeclaration.SelectedResult = pointerSelectedGraphCleanupBinding(third.Binding)
	insertGraphCleanupDeclaration(t, ctx, pool, store, thirdDeclaration)

	cleanupFirstTx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupFirstTx.Rollback(ctx)
	cleanupFirstCleaner, _ := postgresresult.NewCleaner(cleanupFirstTx)
	secondCursor := &graphprojection.ResultCleanupCandidateV2{ProjectionResultID: second.Binding.ProjectionResultID, PublishedAt: second.PublishedAt}
	thirdCandidate, err := cleanupFirstCleaner.LockCleanupCandidate(ctx, ProfileID, secondCursor)
	if err != nil || thirdCandidate == nil || thirdCandidate.ProjectionResultID != third.Binding.ProjectionResultID {
		t.Fatalf("lock selected admission-race candidate: candidate=%#v err=%v", thirdCandidate, err)
	}
	inputAdmissionStarted := make(chan struct{})
	inputAdmissionDone := make(chan error, 1)
	go func() {
		tx, beginErr := pool.BeginTx(ctx, pgx.TxOptions{})
		if beginErr != nil {
			inputAdmissionDone <- beginErr
			return
		}
		defer func() { _ = tx.Rollback(ctx) }()
		close(inputAdmissionStarted)
		_, admissionErr := source.ReadSelectedResultTx(ctx, tx, incident.ID, third.Binding)
		if admissionErr != nil {
			inputAdmissionDone <- admissionErr
			return
		}
		inputAdmissionDone <- tx.Commit(ctx)
	}()
	<-inputAdmissionStarted
	select {
	case err := <-inputAdmissionDone:
		t.Fatalf("input admission crossed a cleanup-held result lock: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	selected, err = store.LockGraphViewDeclarationsSelectingResultTx(ctx, cleanupFirstTx, third.Binding.ProjectionResultID)
	if err != nil || !selected {
		t.Fatalf("cleanup selected-binding recheck = %t err=%v", selected, err)
	}
	if err := cleanupFirstTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-inputAdmissionDone:
		if err != nil {
			t.Fatalf("input admission after selected preservation failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("input admission did not resume after cleanup commit")
	}
	requireGraphResultCount(t, pool, third.Binding.ProjectionResultID, 1)
}

func selectedGraphCleanupBinding(binding graphprojection.ResultBindingV2) GraphViewSelectedResultBinding {
	return GraphViewSelectedResultBinding{
		ProjectionResultID:            binding.ProjectionResultID,
		SourceSnapshotID:              binding.SourceSnapshotID,
		ProjectionSchemaID:            binding.ProjectionSchemaID,
		ProjectionVersion:             binding.ProjectionVersion,
		NormalizedConfigurationSHA256: binding.NormalizedConfigurationSHA256,
		NormalizedSourceSHA256:        binding.NormalizedSourceSHA256,
		CanonicalOutputSHA256:         binding.CanonicalOutputSHA256,
	}
}

func pointerSelectedGraphCleanupBinding(binding graphprojection.ResultBindingV2) *GraphViewSelectedResultBinding {
	selected := selectedGraphCleanupBinding(binding)
	return &selected
}

func publishGraphCleanupResult(t testing.TB, ctx context.Context, pool *pgxpool.Pool, result graphprojection.CompletedResultV2) {
	t.Helper()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	publisher, err := postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatal(err)
	}
	if err := publisher.PublishResult(ctx, result); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func insertGraphCleanupDeclaration(t testing.TB, ctx context.Context, pool *pgxpool.Pool, store *Store, declaration GraphViewDeclaration) {
	t.Helper()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := store.InsertGraphViewDeclarationTx(ctx, tx, declaration); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func seedGraphCleanupJob(t testing.TB, ctx context.Context, pool *pgxpool.Pool, jobID, incidentID, actorID uuid.UUID, now time.Time) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
INSERT INTO jobs (
    job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, auth_policy, handler_name,
    job_kind, progress_unit_id
) VALUES (
    $1, 'incident', $2, 'queued', true, $3, $4, $4, 0,
    'incident_membership', $5, $6,
    'network_flow_activity.graph_view_materialize.projection_result.v1'
)
`, jobID, incidentID, actorID, now, GraphViewWorkerKind, GraphViewMaterializationJobKind); err != nil {
		t.Fatal(err)
	}
}

func requireSelectedGraphCleanupResult(t testing.TB, ctx context.Context, store *Store, incidentID uuid.UUID, graphViewID, resultID string) {
	t.Helper()
	declaration, err := store.GetGraphViewDeclaration(ctx, incidentID, graphViewID)
	if err != nil || declaration.SelectedResult == nil || declaration.SelectedResult.ProjectionResultID != resultID {
		t.Fatalf("selected result = %#v err=%v; want %s", declaration.SelectedResult, err, resultID)
	}
}
