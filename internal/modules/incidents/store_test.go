package incidents_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/importfinalizerport"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/admissiontest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/mutationtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/modules/workbook/startup/bootstrapport"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestStoreCreateIncidentCommitsBootstrapAdminAndWorkbookPreferences_Unit(t *testing.T) {
	harness := storetest.StartStore(t, "incident_membership-u-2-02")
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

	result := storetest.CreateIncidentInStore(
		t, harness.Incidents, actor,
		"txn-incident_membership-u-2-02-create", "IR-U202", "Incident administration incident-storage",
	)

	if !result.Created {
		t.Fatalf("fresh incident create must be marked created: %#v", result)
	}

	membership := storetest.GetMembership(t, harness.DB, result.Incident.ID, actor.ID)
	if membership.Role != "admin" || membership.UserID != actor.ID || membership.AddedByUserID != actor.ID {
		t.Fatalf("unexpected bootstrap membership: %#v", membership)
	}

	startupStore, err := workbookassembly.NewStartupStore(
		harness.DB,
		workbookstartup.NewWorkspaceRegistryFromPublication(nil),
	)
	if err != nil {
		t.Fatalf("construct startup store: %v", err)
	}
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

func TestMutationCommandTimeIsSharedAcrossDomainRawAndProjectedAuditRows_Unit(t *testing.T) {
	location := time.FixedZone("command-input", -5*60*60)
	createTime := time.Date(2026, 8, 28, 8, 0, 0, 0, location)
	currentTime := createTime
	clockCalls := 0
	harness := storetest.StartStoreWithClock(t, "incident-command-time", func() time.Time {
		clockCalls++
		return currentTime
	})
	ctx := context.Background()
	admin := authstoretest.SeedLocalUserRecord(
		t, harness.DB, "incident-command-time-admin@example.test", "Incident command time admin",
		"IncidentCommandTimeAdmin1!", false, false, true,
	)
	member := authstoretest.SeedLocalUserRecord(
		t, harness.DB, "incident-command-time-member@example.test", "Incident command time member",
		"IncidentCommandTimeMember1!", false, false, true,
	)
	createRequestID := "req-command-time-create"
	createRequest := admissiontest.IncidentCreate(t, admissiontest.IncidentCreateInput{
		ClientTxnID: "txn-command-time-create",
		IncidentKey: "IR-COMMAND-TIME",
		Title:       "Shared command time",
	})
	created, err := harness.Incidents.CreateIncident(ctx, admin, createRequest, createRequestID)
	if err != nil {
		t.Fatalf("create command-time incident: %v", err)
	}
	requireMutationClockCalls(t, clockCalls, 1)
	wantCreateTime := createTime.UTC()
	if !created.Incident.CreatedAt.Equal(wantCreateTime) || !created.Incident.UpdatedAt.Equal(wantCreateTime) {
		t.Fatalf("incident create timestamps = (%s, %s), want %s", created.Incident.CreatedAt, created.Incident.UpdatedAt, wantCreateTime)
	}
	if got := storetest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM incident_memberships
 WHERE incident_id = $1 AND user_id = $2 AND joined_at = $3 AND updated_at = $3
`, created.Incident.ID, admin.ID, wantCreateTime); got != 1 {
		t.Fatalf("bootstrap membership command-time rows = %d, want 1", got)
	}
	if got := storetest.QueryCount(t, harness.DB, `
SELECT (SELECT COUNT(*) FROM incident_workbook_preferences WHERE incident_id = $1 AND created_at = $3 AND updated_at = $3)
     + (SELECT COUNT(*) FROM user_workbook_preferences WHERE incident_id = $1 AND user_id = $2 AND created_at = $3 AND updated_at = $3)
`, created.Incident.ID, admin.ID, wantCreateTime); got != 2 {
		t.Fatalf("workbook preference command-time rows = %d, want 2", got)
	}
	requireMutationAuditTime(t, harness, createRequestID, wantCreateTime, 2, 1)

	membershipTime := createTime.Add(time.Minute)
	currentTime = membershipTime
	membershipRequestID := "req-command-time-membership-create"
	membershipRequest := admissiontest.MembershipCreate(t, admissiontest.MembershipCreateInput{
		ClientTxnID: "txn-command-time-membership-create",
		UserID:      &member.ID,
		Role:        "viewer",
	})
	membershipResult, err := harness.Incidents.CreateMembership(
		ctx, admin, created.Incident.ID, member, membershipRequest, membershipRequestID,
	)
	if err != nil {
		t.Fatalf("create command-time membership: %v", err)
	}
	requireMutationClockCalls(t, clockCalls, 2)
	wantMembershipTime := membershipTime.UTC()
	if !membershipResult.Membership.JoinedAt.Equal(wantMembershipTime) || !membershipResult.Membership.UpdatedAt.Equal(wantMembershipTime) {
		t.Fatalf("membership create timestamps = (%s, %s), want %s", membershipResult.Membership.JoinedAt, membershipResult.Membership.UpdatedAt, wantMembershipTime)
	}
	requireMutationAuditTime(t, harness, membershipRequestID, wantMembershipTime, 1, 1)

	incidentPatchTime := createTime.Add(2 * time.Minute)
	currentTime = incidentPatchTime
	incidentPatchRequestID := "req-command-time-incident-patch"
	severity := "high"
	updatedIncident, changed, err := harness.Incidents.UpdateIncident(
		ctx,
		admin,
		created.Incident.ID,
		admissiontest.IncidentPatch(t, admissiontest.IncidentPatchInput{
			BaseIncidentVersion: created.Incident.IncidentVersion,
			Severity:            admissiontest.OptionalStringInput{Present: true, Value: &severity},
		}),
		incidentPatchRequestID,
	)
	if err != nil || !changed {
		t.Fatalf("patch command-time incident: changed=%t err=%v", changed, err)
	}
	requireMutationClockCalls(t, clockCalls, 3)
	if !updatedIncident.UpdatedAt.Equal(incidentPatchTime.UTC()) {
		t.Fatalf("incident patch timestamp = %s, want %s", updatedIncident.UpdatedAt, incidentPatchTime.UTC())
	}
	requireMutationAuditTime(t, harness, incidentPatchRequestID, incidentPatchTime.UTC(), 1, 0)

	membershipPatchTime := createTime.Add(3 * time.Minute)
	currentTime = membershipPatchTime
	membershipPatchRequestID := "req-command-time-membership-patch"
	updatedMembership, changed, err := harness.Incidents.UpdateMembership(
		ctx,
		admin,
		created.Incident.ID,
		member.ID,
		admissiontest.MembershipPatch(t, admissiontest.MembershipPatchInput{
			BaseMembershipVersion: membershipResult.Membership.MembershipVersion,
			Role:                  "reviewer",
		}),
		membershipPatchRequestID,
	)
	if err != nil || !changed {
		t.Fatalf("patch command-time membership: changed=%t err=%v", changed, err)
	}
	requireMutationClockCalls(t, clockCalls, 4)
	if !updatedMembership.UpdatedAt.Equal(membershipPatchTime.UTC()) {
		t.Fatalf("membership patch timestamp = %s, want %s", updatedMembership.UpdatedAt, membershipPatchTime.UTC())
	}
	requireMutationAuditTime(t, harness, membershipPatchRequestID, membershipPatchTime.UTC(), 1, 1)

	lifecycleTime := createTime.Add(4 * time.Minute)
	currentTime = lifecycleTime
	lifecycleRequestID := "req-command-time-lifecycle"
	lifecycle, err := harness.Incidents.TransitionIncidentLifecycle(
		ctx,
		admin,
		created.Incident.ID,
		admissiontest.IncidentLifecycle(t, incidents.LifecycleActionClose, admissiontest.IncidentLifecycleInput{
			BaseIncidentVersion: updatedIncident.IncidentVersion,
			ClientTxnID:         "txn-command-time-lifecycle",
			Reason:              "resolved",
		}),
		lifecycleRequestID,
	)
	if err != nil {
		t.Fatalf("close command-time incident: %v", err)
	}
	requireMutationClockCalls(t, clockCalls, 5)
	if lifecycle.Incident.ClosedAt == nil || !lifecycle.Incident.ClosedAt.Equal(lifecycleTime.UTC()) || !lifecycle.Incident.UpdatedAt.Equal(lifecycleTime.UTC()) {
		t.Fatalf("lifecycle timestamps = closed=%v updated=%s, want %s", lifecycle.Incident.ClosedAt, lifecycle.Incident.UpdatedAt, lifecycleTime.UTC())
	}
	requireMutationAuditTime(t, harness, lifecycleRequestID, lifecycleTime.UTC(), 1, 0)
	if replay, err := harness.Incidents.TransitionIncidentLifecycle(
		ctx,
		admin,
		created.Incident.ID,
		admissiontest.IncidentLifecycle(t, incidents.LifecycleActionClose, admissiontest.IncidentLifecycleInput{
			BaseIncidentVersion: updatedIncident.IncidentVersion,
			ClientTxnID:         "txn-command-time-lifecycle",
			Reason:              "resolved",
		}),
		"req-command-time-lifecycle-replay",
	); err != nil || !replay.Commit.IsReplay() {
		t.Fatalf("replay command-time lifecycle: replay=%#v err=%v", replay, err)
	}
	requireMutationClockCalls(t, clockCalls, 5)
	if got := mutationtest.CountMutationArtifacts(
		t,
		mutationtest.PostgresDatabase(harness.DB),
		mutationtest.MutationSelector{IncidentID: created.Incident.ID.String()},
		mutationtest.MutationOwnerIncidentResource,
	); got != 3 {
		t.Fatalf("lifecycle replay changed incident audit count: got %d want 3", got)
	}

	deleteTime := createTime.Add(5 * time.Minute)
	currentTime = deleteTime
	deleteRequestID := "req-command-time-membership-delete"
	if _, err := harness.Incidents.DeleteMembership(
		ctx,
		admin,
		created.Incident.ID,
		member.ID,
		admissiontest.MembershipDelete(t, updatedMembership.MembershipVersion),
		deleteRequestID,
	); err != nil {
		t.Fatalf("delete command-time membership: %v", err)
	}
	requireMutationClockCalls(t, clockCalls, 6)
	requireMutationAuditTime(t, harness, deleteRequestID, deleteTime.UTC(), 1, 1)
}

func requireMutationClockCalls(t testing.TB, got int, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("mutation clock calls = %d, want %d", got, want)
	}
}

func requireMutationAuditTime(
	t testing.TB,
	harness *storetest.StoreHarness,
	requestID string,
	want time.Time,
	wantRaw int,
	wantProjected int,
) {
	t.Helper()
	if got := storetest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM deployment_admin_audit_events
 WHERE request_id = $1 AND created_at = $2
`, requestID, want); got != wantRaw {
		t.Fatalf("raw audit rows at command time for %s = %d, want %d", requestID, got, wantRaw)
	}
	if got := storetest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM administrative_audit_projections projection
  JOIN deployment_admin_audit_events raw ON raw.id = projection.audit_event_id
 WHERE raw.request_id = $1 AND raw.created_at = $2 AND projection.occurred_at = $2
`, requestID, want); got != wantProjected {
		t.Fatalf("projected audit rows at command time for %s = %d, want %d", requestID, got, wantProjected)
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
	store, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            harness.DB,
		PreferenceBootstrap: failingPreferenceBootstrap{err: bootstrapErr},
		Now:                 time.Now,
	})
	if err != nil {
		t.Fatalf("construct Incidents application: %v", err)
	}
	input := admissiontest.IncidentCreateInput{
		ClientTxnID: "txn-incident_membership-support-bootstrap-port-rollback",
		IncidentKey: "IR-SUPPORT-BOOTSTRAP-PORT-ROLLBACK",
		Title:       "Incident administration support bootstrap port rollback",
	}
	request := admissiontest.IncidentCreate(t, input)
	_, err = store.CreateIncident(
		context.Background(),
		actor,
		request,
		"req-"+input.ClientTxnID,
	)
	if !errors.Is(err, bootstrapErr) {
		t.Fatalf("expected bootstrap port failure, got %T %[1]v", err)
	}
	if got := storetest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM incidents WHERE incident_key_canonical = $1`, input.IncidentKey); got != 0 {
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
	if got := mutationtest.CountMutationArtifacts(t, mutationtest.PostgresDatabase(harness.DB), mutationtest.MutationSelector{ClientTxnID: input.ClientTxnID}); got != 0 {
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
	incident := storetest.CreateIncidentInStore(
		t, harness.Incidents, admin,
		"txn-terminal-effect-create", "IR-TERMINAL-EFFECT", "Terminal effect identity",
	).Incident
	membership := storetest.CreateMembershipInStore(
		t,
		harness.DB,
		admin,
		incident.ID,
		member,
		"txn-terminal-effect-membership",
		"viewer",
	).Membership

	closeRequest := admissiontest.IncidentLifecycle(t, incidents.LifecycleActionClose, admissiontest.IncidentLifecycleInput{
		BaseIncidentVersion: incident.IncidentVersion,
		ClientTxnID:         "txn-terminal-effect-close",
		Reason:              "Terminal effect identity validation.",
	})
	closeResult, err := harness.Incidents.TransitionIncidentLifecycle(
		ctx,
		admin,
		incident.ID,
		closeRequest,
		"req-terminal-effect-close",
	)
	if err != nil {
		t.Fatalf("close incident: %v", err)
	}
	requireCommittedAuditIdentity(t, harness, closeResult.Commit, "req-terminal-effect-close")

	replay, err := harness.Incidents.TransitionIncidentLifecycle(
		ctx,
		admin,
		incident.ID,
		closeRequest,
		"req-terminal-effect-close-replay",
	)
	if err != nil {
		t.Fatalf("replay close incident: %v", err)
	}
	if !replay.Commit.IsReplay() {
		t.Fatalf("close replay commit = %#v, want replay without effect key", replay.Commit)
	}
	if _, present := replay.Commit.EffectKey(); present {
		t.Fatalf("close replay exposed an effect key: %#v", replay.Commit)
	}

	deleteResult, err := harness.Incidents.DeleteMembership(
		ctx,
		admin,
		incident.ID,
		member.ID,
		admissiontest.MembershipDelete(t, membership.MembershipVersion),
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
	if !commit.IsNewCommit() {
		t.Fatalf("terminal mutation commit = %#v, want new commit", commit)
	}
	effectKey, present := commit.EffectKey()
	if !present {
		t.Fatal("new terminal mutation commit omitted effect key")
	}
	var auditEventID uuid.UUID
	if err := harness.DB.QueryRow(context.Background(), `
SELECT id
  FROM deployment_admin_audit_events
 WHERE request_id = $1
`, requestID).Scan(&auditEventID); err != nil {
		t.Fatalf("query committed audit identity for %s: %v", requestID, err)
	}
	if effectKey != auditEventID {
		t.Fatalf("effect key = %s, want committed audit event %s", effectKey, auditEventID)
	}
}

func TestIncidentBundleImportFinalizationCommitsBootstrapState(t *testing.T) {
	harness := storetest.StartStore(t, "incident_membership-support-incident-bundle-finalize")
	finalizer, err := incidents.NewIncidentBundleImportFinalizer(workbookstartuppostgres.NewWriter())
	if err != nil {
		t.Fatalf("construct incident bundle import finalizer: %v", err)
	}
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
	requestID := "incident-bundle-finalization-request"

	tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin finalization transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	insertImportedIncidentTx(t, tx, incidentID, actor.ID, "IR-BUNDLE-FINALIZE")
	if err := finalizer.FinalizeIncidentBundleImportTx(context.Background(), tx, importfinalizerport.Params{
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
	requireMutationAuditTime(t, harness, requestID, publishedAt, 1, 1)
}

func TestIncidentBundleImportFinalizationRejectsMissingSubmitter(t *testing.T) {
	harness := storetest.StartStore(t, "incident_membership-support-incident-bundle-finalize-missing")
	finalizer, err := incidents.NewIncidentBundleImportFinalizer(workbookstartuppostgres.NewWriter())
	if err != nil {
		t.Fatalf("construct incident bundle import finalizer: %v", err)
	}
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
	err = finalizer.FinalizeIncidentBundleImportTx(context.Background(), tx, importfinalizerport.Params{
		IncidentID:        incidentID,
		SubmittedByUserID: uuid.New(),
		PublishedAt:       time.Now().UTC(),
	})
	if !errors.Is(err, importfinalizerport.ErrInitialAdminUnavailable) {
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

func TestIncidentBundleImportFinalizerRejectsInvalidDependenciesAndParamsBeforeDatabaseAccess(t *testing.T) {
	if finalizer, err := incidents.NewIncidentBundleImportFinalizer(nil); err == nil || finalizer != nil {
		t.Fatalf("nil bootstrap constructed finalizer: finalizer=%v err=%v", finalizer, err)
	}
	var typedNilBootstrap *workbookstartuppostgres.Writer
	if finalizer, err := incidents.NewIncidentBundleImportFinalizer(typedNilBootstrap); err == nil || finalizer != nil {
		t.Fatalf("typed-nil bootstrap constructed finalizer: finalizer=%v err=%v", finalizer, err)
	}

	finalizer, err := incidents.NewIncidentBundleImportFinalizer(workbookstartuppostgres.NewWriter())
	if err != nil {
		t.Fatalf("construct valid finalizer: %v", err)
	}
	validParams := importfinalizerport.Params{
		IncidentID:        uuid.New(),
		SubmittedByUserID: uuid.New(),
		PublishedAt:       time.Date(2026, 8, 19, 5, 0, 0, 0, time.UTC),
	}
	var typedNilTx *pgxpool.Tx
	tests := []struct {
		name      string
		tx        pgx.Tx
		params    importfinalizerport.Params
		wantError string
	}{
		{name: "nil transaction", params: validParams, wantError: "transaction is required"},
		{name: "typed nil transaction", tx: typedNilTx, params: validParams, wantError: "transaction is required"},
		{name: "zero incident ID", tx: &pgxpool.Tx{}, params: importfinalizerport.Params{SubmittedByUserID: validParams.SubmittedByUserID, PublishedAt: validParams.PublishedAt}, wantError: "incident ID is required"},
		{name: "zero submitter ID", tx: &pgxpool.Tx{}, params: importfinalizerport.Params{IncidentID: validParams.IncidentID, PublishedAt: validParams.PublishedAt}, wantError: "submitter ID is required"},
		{name: "zero publication time", tx: &pgxpool.Tx{}, params: importfinalizerport.Params{IncidentID: validParams.IncidentID, SubmittedByUserID: validParams.SubmittedByUserID}, wantError: "publication time is required"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := finalizer.FinalizeIncidentBundleImportTx(context.Background(), test.tx, test.params)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("finalization error = %v, want %q", err, test.wantError)
			}
		})
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

	firstInput := admissiontest.IncidentCreateInput{
		ClientTxnID: "txn-u-2-04",
		IncidentKey: "IR-U204",
		Title:       "Replay Incident",
	}
	firstRequest := admissiontest.IncidentCreate(t, firstInput)

	firstResult, err := store.CreateIncident(
		context.Background(),
		actor,
		firstRequest,
		"req-txn-u-2-04-create",
	)
	if err != nil {
		t.Fatalf("create first incident: %v", err)
	}
	if !firstResult.Created {
		t.Fatalf("fresh create must be marked created: %#v", firstResult)
	}

	selector := storetest.IncidentCreateReplaySelector{
		ActorUserID: actor.ID,
		ClientTxnID: firstInput.ClientTxnID,
		IncidentID:  firstResult.Incident.ID,
	}
	stableBefore := storetest.SnapshotIncidentCreateReplaySideEffects(t, storetest.PostgresReplayDatabase(harness.DB), selector)

	replayRequest := admissiontest.IncidentCreate(t, firstInput)
	replayResult, err := store.CreateIncident(
		context.Background(),
		actor,
		replayRequest,
		"req-txn-u-2-04-replay",
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

	divergentInput := firstInput
	divergentInput.Title = "Different title"
	divergentRequest := admissiontest.IncidentCreate(t, divergentInput)
	if _, err := store.CreateIncident(
		context.Background(),
		actor,
		divergentRequest,
		"req-txn-u-2-04-divergent",
	); !errors.Is(err, authn.ErrClientTxnConflict) {
		t.Fatalf("divergent replay must return client transaction conflict: %v", err)
	}
	if stableAfterConflict := storetest.SnapshotIncidentCreateReplaySideEffects(t, storetest.PostgresReplayDatabase(harness.DB), selector); stableAfterConflict != stableBefore {
		t.Fatalf("divergent replay must not change durable side effects: before=%+v after=%+v", stableBefore, stableAfterConflict)
	}

	secondActorInput := admissiontest.IncidentCreateInput{
		ClientTxnID: firstInput.ClientTxnID,
		IncidentKey: "IR-U204-ACTOR2",
		Title:       "Second Actor Incident",
	}
	secondActorRequest := admissiontest.IncidentCreate(t, secondActorInput)
	secondActorResult, err := store.CreateIncident(
		context.Background(),
		secondActor,
		secondActorRequest,
		"req-txn-u-2-04-actor-two",
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
		ClientTxnID: secondActorInput.ClientTxnID,
		IncidentID:  secondActorResult.Incident.ID,
	}); got != wantSecondActorSideEffects {
		t.Fatalf("unexpected second-actor durable side effects: got=%+v want=%+v", got, wantSecondActorSideEffects)
	}
}

func TestStoreIncidentPatchReturnsTypedVersionConflictDetails_Unit(t *testing.T) {
	harness := storetest.StartStore(t, "incident_membership-u-2-14")
	store, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            harness.DB,
		PreferenceBootstrap: workbookstartuppostgres.NewWriter(),
		Now:                 time.Now,
	})
	if err != nil {
		t.Fatalf("construct Incidents application: %v", err)
	}
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

	incidentResult := storetest.CreateIncidentInStore(
		t, harness.Incidents, admin,
		"txn-incident_membership-u-2-14-incident", "IR-U214", "Incident administration incident-storage",
	)

	ctx := context.Background()
	tlp := "TLP:AMBER"
	updated, changed, err := store.UpdateIncident(
		ctx,
		admin,
		incidentResult.Incident.ID,
		admissiontest.IncidentPatch(t, admissiontest.IncidentPatchInput{
			BaseIncidentVersion: incidentResult.Incident.IncidentVersion,
			TLP:                 admissiontest.OptionalStringInput{Present: true, Value: &tlp},
		}),
		"req-incident_membership-u-2-14-update",
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
		admissiontest.IncidentPatch(t, admissiontest.IncidentPatchInput{
			BaseIncidentVersion: incidentResult.Incident.IncidentVersion,
			TLP:                 admissiontest.OptionalStringInput{Present: true, Value: &staleTLP},
		}),
		"req-incident_membership-u-2-14-stale",
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
	store, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            harness.DB,
		PreferenceBootstrap: workbookstartuppostgres.NewWriter(),
		Now:                 time.Now,
	})
	if err != nil {
		t.Fatalf("construct Incidents application: %v", err)
	}
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

	incidentResult := storetest.CreateIncidentInStore(
		t, harness.Incidents, admin,
		"txn-incident_membership-u-2-07-incident", "IR-U207", "Incident administration incident-storage",
	)

	membershipResult := storetest.CreateMembershipInStore(
		t,
		harness.DB,
		admin,
		incidentResult.Incident.ID,
		target,
		"txn-incident_membership-u-2-07-membership",
		"viewer",
	)
	staleVersion := membershipResult.Membership.MembershipVersion + 1

	if _, _, err := store.UpdateMembership(
		context.Background(),
		admin,
		incidentResult.Incident.ID,
		target.ID,
		admissiontest.MembershipPatch(t, admissiontest.MembershipPatchInput{
			BaseMembershipVersion: staleVersion,
			Role:                  "reviewer",
		}),
		"req-incident_membership-u-2-07-patch",
	); !errors.Is(err, incidents.ErrMembershipVersionConflict) {
		t.Fatalf("stale membership patch must reject with version conflict: %v", err)
	}

	current := storetest.GetMembership(t, harness.DB, incidentResult.Incident.ID, target.ID)
	if current.Role != membershipResult.Membership.Role || current.MembershipVersion != membershipResult.Membership.MembershipVersion {
		t.Fatalf("stale membership patch must not mutate membership state: before=%#v after=%#v", membershipResult.Membership, current)
	}

	noOp, changed, err := store.UpdateMembership(
		context.Background(),
		admin,
		incidentResult.Incident.ID,
		target.ID,
		admissiontest.MembershipPatch(t, admissiontest.MembershipPatchInput{
			BaseMembershipVersion: membershipResult.Membership.MembershipVersion,
			Role:                  membershipResult.Membership.Role,
		}),
		"req-incident_membership-u-2-07-no-op",
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

	current = storetest.GetMembership(t, harness.DB, incidentResult.Incident.ID, target.ID)
	if current.Role != membershipResult.Membership.Role || current.MembershipVersion != membershipResult.Membership.MembershipVersion {
		t.Fatalf("same-role membership patch must not mutate durable state: before=%#v after=%#v", membershipResult.Membership, current)
	}

	if _, err := store.DeleteMembership(
		context.Background(),
		admin,
		incidentResult.Incident.ID,
		target.ID,
		admissiontest.MembershipDelete(t, staleVersion),
		"req-incident_membership-u-2-07-delete",
	); !errors.Is(err, incidents.ErrMembershipVersionConflict) {
		t.Fatalf("stale membership delete must reject with version conflict: %v", err)
	}

	current = storetest.GetMembership(t, harness.DB, incidentResult.Incident.ID, target.ID)
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
