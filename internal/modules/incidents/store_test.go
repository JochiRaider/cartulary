package incidents_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/mutationtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/modules/workbook/startup/bootstrapport"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestStoreCreateIncidentCommitsBootstrapAdminAndWorkbookPreferences_Unit(t *testing.T) {
	harness := storetest.StartStore(t, "incident_membership-u-2-02")
	store := harness.Incidents
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"incident_membership-u202@example.test",
		"Incident administration U202",
		"IncidentMembershipU202Pass!",
		false,
		false,
		true,
	)

	result := storetest.CreateIncidentInStore(t, harness.Incidents, actor, incidents.CreateIncidentRequest{
		ClientTxnID: "txn-incident_membership-u-2-02-create",
		IncidentKey: "IR-U202",
		Title:       "Incident administration incident-storage",
	})

	if !result.Created {
		t.Fatalf("fresh incident create must be marked created: %#v", result)
	}

	membership, err := store.GetMembership(context.Background(), result.Incident.ID, actor.ID)
	if err != nil {
		t.Fatalf("lookup bootstrap membership: %v", err)
	}
	if membership.Role != "admin" || membership.UserID != actor.ID || membership.AddedByUserID != actor.ID {
		t.Fatalf("unexpected bootstrap membership: %#v", membership)
	}

	startupStore := workbookassembly.NewStartupStore(
		harness.DB,
		workbookstartup.NewWorkspaceRegistryFromPublication(nil),
	)
	defaultPrefs, err := startupStore.GetDefaultPreferences(context.Background(), result.Incident.ID)
	if err != nil {
		t.Fatalf("lookup incident workbook preferences: %v", err)
	}
	if defaultPrefs.IncidentID != result.Incident.ID || defaultPrefs.UpdatedByUserID == nil || *defaultPrefs.UpdatedByUserID != actor.ID {
		t.Fatalf("unexpected incident workbook preferences: %#v", defaultPrefs)
	}
	if len(defaultPrefs.DefaultSheetRef) != 0 || !defaultPrefs.CreatedAt.Equal(result.Incident.CreatedAt) || !defaultPrefs.UpdatedAt.Equal(result.Incident.CreatedAt) {
		t.Fatalf("incident bootstrap preferences must preserve null ref and commit timestamp: %#v", defaultPrefs)
	}

	userPrefs, err := startupStore.GetUserPreferences(context.Background(), result.Incident.ID, actor.ID)
	if err != nil {
		t.Fatalf("lookup user workbook preferences: %v", err)
	}
	if userPrefs.IncidentID != result.Incident.ID || userPrefs.UserID != actor.ID {
		t.Fatalf("unexpected user workbook preferences: %#v", userPrefs)
	}
	if len(userPrefs.HomeSheetRef) != 0 || !userPrefs.CreatedAt.Equal(result.Incident.CreatedAt) || !userPrefs.UpdatedAt.Equal(result.Incident.CreatedAt) {
		t.Fatalf("user bootstrap preferences must preserve null ref and commit timestamp: %#v", userPrefs)
	}
}

func TestStoreCreateIncidentPreferenceBootstrapFailureRollsBack(t *testing.T) {
	harness := storetest.StartStore(t, "incident_membership-support-bootstrap-port-rollback")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"incident_membership-support-bootstrap-port@example.test",
		"Incident administration Bootstrap Port",
		"IncidentMembershipBootstrapPortPass!",
		false,
		false,
		true,
	)
	bootstrapErr := errors.New("bootstrap port failed")
	store := incidents.NewApplicationWithOptions(harness.DB, incidents.ApplicationOptions{
		PreferenceBootstrap: failingPreferenceBootstrap{err: bootstrapErr},
	})
	request := incidents.CreateIncidentRequest{
		ClientTxnID: "txn-incident_membership-support-bootstrap-port-rollback",
		IncidentKey: "IR-SUPPORT-BOOTSTRAP-PORT-ROLLBACK",
		Title:       "Incident administration support bootstrap port rollback",
	}
	_, err := store.CreateIncident(
		context.Background(),
		actor,
		request,
		incidents.IncidentCreateRequestHash(request),
		"req-"+request.ClientTxnID,
		time.Now().UTC(),
	)
	if !errors.Is(err, bootstrapErr) {
		t.Fatalf("expected bootstrap port failure, got %T %[1]v", err)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM incidents WHERE incident_key_canonical = $1`, request.IncidentKey); got != 0 {
		t.Fatalf("bootstrap failure must leave no incident row, got %d", got)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM incident_memberships WHERE user_id = $1`, actor.ID); got != 0 {
		t.Fatalf("bootstrap failure must leave no membership row, got %d", got)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM incident_workbook_preferences`); got != 0 {
		t.Fatalf("bootstrap failure must leave no incident workbook preferences, got %d", got)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM user_workbook_preferences WHERE user_id = $1`, actor.ID); got != 0 {
		t.Fatalf("bootstrap failure must leave no user workbook preferences, got %d", got)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = 'incidents.create' AND actor_user_id = $1`, actor.ID); got != 0 {
		t.Fatalf("bootstrap failure must leave no idempotency commit, got %d", got)
	}
	if got := mutationtest.CountMutationArtifacts(t, mutationtest.PostgresDatabase(harness.DB), mutationtest.MutationSelector{ClientTxnID: request.ClientTxnID}); got != 0 {
		t.Fatalf("bootstrap failure must leave no mutation artifacts, got %d", got)
	}
}

func TestTerminalMutationCommitResultsUseCommittedAuditEventIdentity_Unit(t *testing.T) {
	harness := storetest.StartStore(t, "incident-terminal-effect-identity")
	ctx := context.Background()
	admin := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"incident-terminal-effect-admin@example.test",
		"Incident terminal effect admin",
		"IncidentTerminalEffectAdmin1!",
		false,
		false,
		true,
	)
	member := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"incident-terminal-effect-member@example.test",
		"Incident terminal effect member",
		"IncidentTerminalEffectMember1!",
		false,
		false,
		true,
	)
	incident := storetest.CreateIncidentInStore(t, harness.Incidents, admin, incidents.CreateIncidentRequest{
		ClientTxnID: "txn-terminal-effect-create",
		IncidentKey: "IR-TERMINAL-EFFECT",
		Title:       "Terminal effect identity",
	}).Incident
	membership := storetest.CreateMembershipInStore(
		t,
		harness.DB,
		admin,
		incident.ID,
		member,
		incidents.MembershipCreateRequest{
			ClientTxnID: "txn-terminal-effect-membership",
			UserID:      &member.ID,
			Role:        "viewer",
		},
	).Membership

	closeRequest := incidents.IncidentLifecycleRequest{
		BaseIncidentVersion: incident.IncidentVersion,
		ClientTxnID:         "txn-terminal-effect-close",
		Reason:              "Terminal effect identity validation.",
	}
	closeResult, err := harness.Incidents.TransitionIncidentLifecycle(
		ctx,
		admin,
		incident.ID,
		"close",
		closeRequest,
		incidents.IncidentLifecycleRequestHash("close", closeRequest),
		"req-terminal-effect-close",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("close incident: %v", err)
	}
	requireCommittedAuditIdentity(t, harness, closeResult.Commit, "req-terminal-effect-close")

	replay, err := harness.Incidents.TransitionIncidentLifecycle(
		ctx,
		admin,
		incident.ID,
		"close",
		closeRequest,
		incidents.IncidentLifecycleRequestHash("close", closeRequest),
		"req-terminal-effect-close-replay",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("replay close incident: %v", err)
	}
	if replay.Commit.Disposition != incidents.TerminalMutationReplay || replay.Commit.EffectKey != uuid.Nil {
		t.Fatalf("close replay commit = %#v, want replay without effect key", replay.Commit)
	}
	if err := replay.Commit.Validate(); err != nil {
		t.Fatalf("validate replay commit: %v", err)
	}

	deleteResult, err := harness.Incidents.DeleteMembership(
		ctx,
		admin,
		incident.ID,
		member.ID,
		incidents.MembershipDeleteRequest{BaseMembershipVersion: membership.MembershipVersion},
		"req-terminal-effect-membership-delete",
	)
	if err != nil {
		t.Fatalf("delete membership: %v", err)
	}
	requireCommittedAuditIdentity(t, harness, deleteResult.Commit, "req-terminal-effect-membership-delete")
}

func requireCommittedAuditIdentity(
	t testing.TB,
	harness *storetest.StoreHarness,
	commit incidents.TerminalMutationCommit,
	requestID string,
) {
	t.Helper()
	if err := commit.Validate(); err != nil {
		t.Fatalf("validate terminal mutation commit: %v", err)
	}
	if commit.Disposition != incidents.TerminalMutationNewCommit {
		t.Fatalf("terminal mutation disposition = %q, want new_commit", commit.Disposition)
	}
	var auditEventID uuid.UUID
	if err := harness.DB.QueryRow(context.Background(), `
SELECT id
  FROM deployment_admin_audit_events
 WHERE request_id = $1
`, requestID).Scan(&auditEventID); err != nil {
		t.Fatalf("query committed audit identity for %s: %v", requestID, err)
	}
	if commit.EffectKey != auditEventID {
		t.Fatalf("effect key = %s, want committed audit event %s", commit.EffectKey, auditEventID)
	}
}

func TestIncidentBundleImportFinalizationCommitsBootstrapState(t *testing.T) {
	harness := storetest.StartStore(t, "incident_membership-support-incident-bundle-finalize")
	finalizer := incidents.NewIncidentBundleImportFinalizer(workbookstartuppostgres.NewWriter())
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"incident_membership-support-bundle-finalize@example.test",
		"Incident administration Bundle Finalize",
		"IncidentMembershipBundleFinalizePass!",
		false,
		true,
		true,
	)
	incidentID := uuid.New()
	publishedAt := time.Date(2026, 5, 25, 18, 0, 0, 0, time.UTC)
	requestID := incidents.ImportBundleRequestID(uuid.New())

	tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin finalization transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	insertImportedIncidentTx(t, tx, incidentID, actor.ID, "IR-BUNDLE-FINALIZE")
	if err := finalizer.FinalizeIncidentBundleImportTx(context.Background(), tx, incidents.IncidentBundleImportFinalizationParams{
		IncidentID:        incidentID,
		SubmittedByUserID: actor.ID,
		PublishedAt:       publishedAt,
		RequestID:         &requestID,
	}); err != nil {
		t.Fatalf("finalize incident bundle import: %v", err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit finalization transaction: %v", err)
	}

	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id = $1 AND user_id = $2 AND role = 'admin' AND joined_at = $3 AND updated_at = $3 AND membership_version = 1`, incidentID, actor.ID, publishedAt); got != 1 {
		t.Fatalf("import finalization membership rows: got %d want 1", got)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM incident_workbook_preferences WHERE incident_id = $1 AND default_sheet_ref IS NULL AND updated_by_user_id = $2 AND created_at = $3 AND updated_at = $3`, incidentID, actor.ID, publishedAt); got != 1 {
		t.Fatalf("import finalization default preference rows: got %d want 1", got)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM user_workbook_preferences WHERE incident_id = $1 AND user_id = $2 AND home_sheet_ref IS NULL AND created_at = $3 AND updated_at = $3`, incidentID, actor.ID, publishedAt); got != 1 {
		t.Fatalf("import finalization user preference rows: got %d want 1", got)
	}
	events := mutationtest.LookupOwnerMutations(
		t, mutationtest.PostgresDatabase(

			harness.DB),

		mutationtest.MutationSelector{IncidentID: incidentID.String()},
		mutationtest.MutationOwnerIncidentMembership)

	event := mutationtest.RequireOwnerMutationEvent(
		t,
		events,
		mutationtest.MutationOwnerIncidentMembership,
		"incident_membership_created",
		actor.ID.String(),
		actor.ID.String(),
	)
	if event.RequestID != requestID || event.ClientTxnID != "" {
		t.Fatalf("unexpected import finalization audit attribution: %#v", event)
	}
}

func TestIncidentBundleImportFinalizationRejectsMissingSubmitter(t *testing.T) {
	harness := storetest.StartStore(t, "incident_membership-support-incident-bundle-finalize-missing")
	finalizer := incidents.NewIncidentBundleImportFinalizer(workbookstartuppostgres.NewWriter())
	creator := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"incident_membership-support-bundle-finalize-creator@example.test",
		"Incident administration Bundle Finalize Creator",
		"IncidentMembershipBundleFinalizeCreatorPass!",
		false,
		true,
		true,
	)
	incidentID := uuid.New()

	tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin finalization transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	insertImportedIncidentTx(t, tx, incidentID, creator.ID, "IR-BUNDLE-FINALIZE-MISSING")
	err = finalizer.FinalizeIncidentBundleImportTx(context.Background(), tx, incidents.IncidentBundleImportFinalizationParams{
		IncidentID:        incidentID,
		SubmittedByUserID: uuid.New(),
		PublishedAt:       time.Now().UTC(),
	})
	if !errors.Is(err, incidents.ErrInitialAdminUnavailable) {
		t.Fatalf("expected ErrInitialAdminUnavailable, got %T %[1]v", err)
	}
	_ = tx.Rollback(context.Background())

	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM incidents WHERE id = $1`, incidentID); got != 0 {
		t.Fatalf("failed finalization transaction left incident row, got %d", got)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id = $1`, incidentID); got != 0 {
		t.Fatalf("failed finalization transaction left membership row, got %d", got)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM incident_workbook_preferences WHERE incident_id = $1`, incidentID); got != 0 {
		t.Fatalf("failed finalization transaction left default preference row, got %d", got)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM user_workbook_preferences WHERE incident_id = $1`, incidentID); got != 0 {
		t.Fatalf("failed finalization transaction left user preference row, got %d", got)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM deployment_admin_audit_events WHERE incident_id = $1`, incidentID); got != 0 {
		t.Fatalf("failed finalization transaction left audit row, got %d", got)
	}
}

type failingPreferenceBootstrap struct {
	err error
}

func (p failingPreferenceBootstrap) InsertInitialTx(context.Context, pgx.Tx, bootstrapport.InitialPreferenceInput) error {
	return p.err
}

func insertImportedIncidentTx(t testing.TB, tx pgx.Tx, incidentID uuid.UUID, actorID uuid.UUID, incidentKey string) {
	t.Helper()
	if _, err := tx.Exec(context.Background(), `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
)
VALUES ($1, $2, $2, $3, 'active', $4, $4)
`, incidentID, incidentKey, incidentKey+" title", actorID); err != nil {
		t.Fatalf("seed imported incident row: %v", err)
	}
}

func TestStoreCreateIncidentReplayPreservesDurableSideEffectsAndScopesByActor_Unit(t *testing.T) {
	harness := storetest.StartStore(t, "incident_membership-u-2-04")
	store := harness.Incidents
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"incident_membership-u204@example.test",
		"Incident administration U204",
		"IncidentMembershipU204Pass!",
		false,
		false,
		true,
	)
	secondActor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"incident_membership-u204-second@example.test",
		"Incident administration U204 Second",
		"IncidentMembershipU204SecondPass!",
		false,
		false,
		true,
	)

	firstRequest := incidents.CreateIncidentRequest{
		ClientTxnID: "txn-u-2-04",
		IncidentKey: "IR-U204",
		Title:       "Replay Incident",
	}

	firstResult, err := store.CreateIncident(
		context.Background(),
		actor,
		firstRequest,
		incidents.IncidentCreateRequestHash(firstRequest),
		"req-txn-u-2-04-create",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("create first incident: %v", err)
	}
	if !firstResult.Created {
		t.Fatalf("fresh create must be marked created: %#v", firstResult)
	}

	selector := storetest.IncidentCreateReplaySelector{
		ActorUserID: actor.ID,
		ClientTxnID: firstRequest.ClientTxnID,
		IncidentID:  firstResult.Incident.ID,
	}
	stableBefore := storetest.SnapshotIncidentCreateReplaySideEffects(t, storetest.PostgresReplayDatabase(harness.DB), selector)

	replayRequest := firstRequest
	replayResult, err := store.CreateIncident(
		context.Background(),
		actor,
		replayRequest,
		incidents.IncidentCreateRequestHash(replayRequest),
		"req-txn-u-2-04-replay",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("replay incident create: %v", err)
	}
	if replayResult.Created || replayResult.Incident.ID != firstResult.Incident.ID {
		t.Fatalf("unexpected replay result: first=%#v replay=%#v", firstResult, replayResult)
	}
	if canonicalFirst := canonicalJSONMap(t, firstResult.Payload); !reflect.DeepEqual(canonicalFirst, replayResult.Payload) {
		t.Fatalf("expected replayed payload to match original create payload: first=%#v replay=%#v", canonicalFirst, replayResult.Payload)
	}
	if stableAfter := storetest.SnapshotIncidentCreateReplaySideEffects(t, storetest.PostgresReplayDatabase(harness.DB), selector); stableAfter != stableBefore {
		t.Fatalf("replay must keep durable side effects stable: before=%+v after=%+v", stableBefore, stableAfter)
	}

	divergentRequest := firstRequest
	divergentRequest.Title = "Different title"
	if _, err := store.CreateIncident(
		context.Background(),
		actor,
		divergentRequest,
		incidents.IncidentCreateRequestHash(divergentRequest),
		"req-txn-u-2-04-divergent",
		time.Now().UTC(),
	); !errors.Is(err, authn.ErrClientTxnConflict) {
		t.Fatalf("divergent replay must return client transaction conflict: %v", err)
	}
	if stableAfterConflict := storetest.SnapshotIncidentCreateReplaySideEffects(t, storetest.PostgresReplayDatabase(harness.DB), selector); stableAfterConflict != stableBefore {
		t.Fatalf("divergent replay must not change durable side effects: before=%+v after=%+v", stableBefore, stableAfterConflict)
	}

	secondActorRequest := incidents.CreateIncidentRequest{
		ClientTxnID: firstRequest.ClientTxnID,
		IncidentKey: "IR-U204-ACTOR2",
		Title:       "Second Actor Incident",
	}
	secondActorResult, err := store.CreateIncident(
		context.Background(),
		secondActor,
		secondActorRequest,
		incidents.IncidentCreateRequestHash(secondActorRequest),
		"req-txn-u-2-04-actor-two",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("create second-actor incident: %v", err)
	}
	if !secondActorResult.Created || secondActorResult.Incident.ID == firstResult.Incident.ID {
		t.Fatalf("actor-scoped idempotency must allow a distinct create for a different actor: %#v", secondActorResult)
	}

	wantSecondActorSideEffects := storetest.IncidentCreateReplaySideEffects{
		BootstrapMembershipRows:        1,
		IncidentRows:                   1,
		IncidentWorkbookPreferenceRows: 1,
		OwnerMutationRows:              2,
		RouteIdempotencyRows:           1,
		UserWorkbookPreferenceRows:     1,
	}
	if got := storetest.SnapshotIncidentCreateReplaySideEffects(t, storetest.PostgresReplayDatabase(harness.DB), storetest.IncidentCreateReplaySelector{
		ActorUserID: secondActor.ID,
		ClientTxnID: secondActorRequest.ClientTxnID,
		IncidentID:  secondActorResult.Incident.ID,
	}); got != wantSecondActorSideEffects {
		t.Fatalf("unexpected second-actor durable side effects: got=%+v want=%+v", got, wantSecondActorSideEffects)
	}
}

func TestStoreIncidentPatchReturnsTypedVersionConflictDetails_Unit(t *testing.T) {
	harness := storetest.StartStore(t, "incident_membership-u-2-14")
	store := incidents.NewApplication(harness.DB, workbookstartuppostgres.NewWriter())
	admin := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"incident_membership-u214-admin@example.test",
		"Incident administration U214 Admin",
		"IncidentMembershipU214AdminPass!",
		false,
		false,
		true,
	)

	incidentResult := storetest.CreateIncidentInStore(t, harness.Incidents, admin, incidents.CreateIncidentRequest{
		ClientTxnID: "txn-incident_membership-u-2-14-incident",
		IncidentKey: "IR-U214",
		Title:       "Incident administration incident-storage",
	})

	ctx := context.Background()
	tlp := "TLP:AMBER"
	updated, changed, err := store.UpdateIncident(
		ctx,
		admin,
		incidentResult.Incident.ID,
		incidents.IncidentPatchRequest{
			BaseIncidentVersion: incidentResult.Incident.IncidentVersion,
			TLP:                 incidents.OptionalNullableString{Present: true, Value: &tlp},
		},
		"req-incident_membership-u-2-14-update",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("update incident before stale patch: %v", err)
	}
	if !changed || updated.IncidentVersion != incidentResult.Incident.IncidentVersion+1 {
		t.Fatalf("expected material update to advance incident version: changed=%v updated=%#v", changed, updated)
	}

	staleTLP := "TLP:GREEN"
	_, _, err = store.UpdateIncident(
		ctx,
		admin,
		incidentResult.Incident.ID,
		incidents.IncidentPatchRequest{
			BaseIncidentVersion: incidentResult.Incident.IncidentVersion,
			TLP:                 incidents.OptionalNullableString{Present: true, Value: &staleTLP},
		},
		"req-incident_membership-u-2-14-stale",
		time.Now().UTC(),
	)
	if !errors.Is(err, incidents.ErrIncidentVersionConflict) {
		t.Fatalf("stale incident patch must reject with version conflict: %v", err)
	}
	var conflict *incidents.IncidentVersionConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("stale incident patch must return typed version conflict: %T %[1]v", err)
	}
	if conflict.IncidentID != incidentResult.Incident.ID ||
		conflict.BaseIncidentVersion != incidentResult.Incident.IncidentVersion ||
		conflict.CurrentIncidentVersion != updated.IncidentVersion {
		t.Fatalf("unexpected incident version conflict: %#v", conflict)
	}
	details := conflict.Details()
	if details["incident_id"] != incidentResult.Incident.ID.String() ||
		details["base_incident_version"] != incidentResult.Incident.IncidentVersion ||
		details["current_incident_version"] != updated.IncidentVersion {
		t.Fatalf("unexpected incident version conflict details: %#v", details)
	}

	current, err := store.GetVisibleIncident(ctx, incidentResult.Incident.ID, admin.ID)
	if err != nil {
		t.Fatalf("lookup incident after stale patch: %v", err)
	}
	if current.IncidentVersion != updated.IncidentVersion || current.TLP == nil || *current.TLP != tlp {
		t.Fatalf("stale incident patch must not mutate incident state: before=%#v after=%#v", updated, current)
	}
}

func TestStoreMembershipPatchAndDeleteRejectStaleBaseVersion_Unit(t *testing.T) {
	harness := storetest.StartStore(t, "incident_membership-u-2-07")
	store := incidents.NewApplication(harness.DB, workbookstartuppostgres.NewWriter())
	admin := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"incident_membership-u207-admin@example.test",
		"Incident administration U207 Admin",
		"IncidentMembershipU207AdminPass!",
		false,
		false,
		true,
	)
	target := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"incident_membership-u207-target@example.test",
		"Incident administration U207 Target",
		"IncidentMembershipU207TargetPass!",
		false,
		false,
		true,
	)

	incidentResult := storetest.CreateIncidentInStore(t, harness.Incidents, admin, incidents.CreateIncidentRequest{
		ClientTxnID: "txn-incident_membership-u-2-07-incident",
		IncidentKey: "IR-U207",
		Title:       "Incident administration incident-storage",
	})

	membershipResult := storetest.CreateMembershipInStore(
		t,
		harness.DB,
		admin,
		incidentResult.Incident.ID,
		target,
		incidents.MembershipCreateRequest{
			ClientTxnID: "txn-incident_membership-u-2-07-membership",
			UserID:      &target.ID,
			Role:        "viewer",
		},
	)
	staleVersion := membershipResult.Membership.MembershipVersion + 1

	if _, _, err := store.UpdateMembership(
		context.Background(),
		admin,
		incidentResult.Incident.ID,
		target.ID,
		incidents.MembershipPatchRequest{
			BaseMembershipVersion: staleVersion,
			Role:                  "reviewer",
		},
		"req-incident_membership-u-2-07-patch",
		time.Now().UTC(),
	); !errors.Is(err, incidents.ErrMembershipVersionConflict) {
		t.Fatalf("stale membership patch must reject with version conflict: %v", err)
	}

	current, err := store.GetMembership(context.Background(), incidentResult.Incident.ID, target.ID)
	if err != nil {
		t.Fatalf("lookup membership after stale patch: %v", err)
	}
	if current.Role != membershipResult.Membership.Role || current.MembershipVersion != membershipResult.Membership.MembershipVersion {
		t.Fatalf("stale membership patch must not mutate membership state: before=%#v after=%#v", membershipResult.Membership, current)
	}

	noOp, changed, err := store.UpdateMembership(
		context.Background(),
		admin,
		incidentResult.Incident.ID,
		target.ID,
		incidents.MembershipPatchRequest{
			BaseMembershipVersion: membershipResult.Membership.MembershipVersion,
			Role:                  membershipResult.Membership.Role,
		},
		"req-incident_membership-u-2-07-no-op",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("same-role membership patch must succeed: %v", err)
	}
	if changed {
		t.Fatalf("same-role membership patch must report no material change: %#v", noOp)
	}
	if noOp.Role != membershipResult.Membership.Role || noOp.MembershipVersion != membershipResult.Membership.MembershipVersion {
		t.Fatalf("same-role membership patch must keep role and version stable: before=%#v after=%#v", membershipResult.Membership, noOp)
	}

	current, err = store.GetMembership(context.Background(), incidentResult.Incident.ID, target.ID)
	if err != nil {
		t.Fatalf("lookup membership after same-role patch: %v", err)
	}
	if current.Role != membershipResult.Membership.Role || current.MembershipVersion != membershipResult.Membership.MembershipVersion {
		t.Fatalf("same-role membership patch must not mutate durable state: before=%#v after=%#v", membershipResult.Membership, current)
	}

	if _, err := store.DeleteMembership(
		context.Background(),
		admin,
		incidentResult.Incident.ID,
		target.ID,
		incidents.MembershipDeleteRequest{BaseMembershipVersion: staleVersion},
		"req-incident_membership-u-2-07-delete",
	); !errors.Is(err, incidents.ErrMembershipVersionConflict) {
		t.Fatalf("stale membership delete must reject with version conflict: %v", err)
	}

	current, err = store.GetMembership(context.Background(), incidentResult.Incident.ID, target.ID)
	if err != nil {
		t.Fatalf("lookup membership after stale delete: %v", err)
	}
	if current.Role != membershipResult.Membership.Role || current.MembershipVersion != membershipResult.Membership.MembershipVersion {
		t.Fatalf("stale membership delete must not mutate membership state: before=%#v after=%#v", membershipResult.Membership, current)
	}
}

func canonicalJSONMap(t testing.TB, value map[string]any) map[string]any {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal canonical json map: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode canonical json map: %v", err)
	}
	return decoded
}
