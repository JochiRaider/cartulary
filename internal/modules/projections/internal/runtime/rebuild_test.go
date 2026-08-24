package runtime_test

import (
	"context"
	"errors"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/timelinefactassembly"
	assessmenttest "github.com/JochiRaider/cartulary/internal/modules/assessments/testsupport"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	projections "github.com/JochiRaider/cartulary/internal/modules/projections/internal/runtime"
	projectiontestsupport "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	timelineprovider "github.com/JochiRaider/cartulary/internal/modules/timeline/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestRebuildRestoreProjectionsRejectsInvalidRequestBeforeStoreAccess(t *testing.T) {
	rebuilder := projections.NewRestoreRebuilderFromStore(nil)
	result, err := rebuilder.RebuildRestoreProjections(context.Background(), restorecontract.ProjectionRebuildRequest{})
	if err == nil || !strings.Contains(err.Error(), "restore_operation_id is required") {
		t.Fatalf("invalid restore projection request error got %v", err)
	}
	if result.Status != restorecontract.ProjectionRebuildStatusFailed ||
		result.ReadinessOutcome != restorecontract.ProjectionReadinessIncomplete ||
		len(result.Errors) != 1 ||
		result.Errors[0].Code != "invalid_restore_projection_rebuild_request" {
		t.Fatalf("invalid request result = %#v", result)
	}
	if result.ProviderResults == nil || result.Warnings == nil || result.Errors == nil {
		t.Fatalf("restore result collections must be non-nil: %#v", result)
	}
}

func TestTimelineProjectionSourceEnumerationIsDeterministicAndKeysetPaged(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "projection-timeline-source-paging")
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "projection-page@example.test", "Projection Page", "ProjectionPage1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-projection-page-incident", "IR-PROJECTION-PAGE", "Projection paging")
	recordIDs := []uuid.UUID{
		uuid.MustParse("10000000-0000-4000-8000-000000000003"),
		uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		uuid.MustParse("10000000-0000-4000-8000-000000000002"),
	}
	for _, recordID := range recordIDs {
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, recordID)
	}

	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin projection source paging: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	source := timelineprovider.NewSource(timelinefactassembly.NewLinkReader())
	first, err := source.ListProjectionInputsTx(ctx, tx, incident.ID, nil, 2)
	if err != nil {
		t.Fatalf("list first projection source page: %v", err)
	}
	if len(first.Inputs) != 2 || first.NextRecordID == nil ||
		first.Inputs[0].RecordID.String() != "10000000-0000-4000-8000-000000000001" ||
		first.Inputs[1].RecordID.String() != "10000000-0000-4000-8000-000000000002" {
		t.Fatalf("unexpected first projection source page: %#v", first)
	}
	second, err := source.ListProjectionInputsTx(ctx, tx, incident.ID, first.NextRecordID, 2)
	if err != nil {
		t.Fatalf("list second projection source page: %v", err)
	}
	if len(second.Inputs) != 1 || second.NextRecordID != nil ||
		second.Inputs[0].RecordID.String() != "10000000-0000-4000-8000-000000000003" {
		t.Fatalf("unexpected second projection source page: %#v", second)
	}
}

func TestPartyProjectionSourceEnumerationIsDeterministicCompleteAndActiveOnly(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "projection-party-source-paging")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"party-projection-page@example.test",
		"Party Projection Page",
		"PartyProjectionPage1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-party-projection-page-incident",
		"IR-PARTY-PROJECTION-PAGE",
		"Party projection paging",
	)
	foreignIncident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-party-projection-page-foreign-incident",
		"IR-PARTY-PROJECTION-FOREIGN",
		"Foreign Party projection paging",
	)
	recordIDs := []uuid.UUID{
		uuid.MustParse("20000000-0000-4000-8000-000000000003"),
		uuid.MustParse("20000000-0000-4000-8000-000000000001"),
		uuid.MustParse("20000000-0000-4000-8000-000000000002"),
	}
	updatedAt := time.Date(2026, 8, 24, 1, 0, 0, 0, time.FixedZone("fixture", 2*60*60))
	for _, recordID := range recordIDs {
		if _, err := harness.DB.Exec(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id, updated_by_user_id,
    row_version
) VALUES ($1, $2, 'party', $3, $3, 7)
`, recordID, incident.ID, actor.ID); err != nil {
			t.Fatalf("seed Party projection envelope %s: %v", recordID, err)
		}
		if _, err := harness.DB.Exec(ctx, `
INSERT INTO parties (
    record_id, incident_id, display_name, party_kind, organization_name,
    role_title, primary_email, timezone_name, external_ref, notes, updated_at
) VALUES ($1, $2, $3, 'organization', $4, $5, $6, $7, $8, $9, $10)
`, recordID, incident.ID, "Party "+recordID.String(), nil, nil, nil, nil, nil, nil, updatedAt); err != nil {
			t.Fatalf("seed Party projection source %s: %v", recordID, err)
		}
	}
	fullRecordID := recordIDs[1]
	if _, err := harness.DB.Exec(ctx, `
UPDATE parties
   SET organization_name = 'Coordination',
       role_title = 'Lead',
       primary_email = 'Owner@Example.Test',
       timezone_name = 'US/Eastern',
       external_ref = 'EXT-Party',
       notes = E'Line one\nLine two'
 WHERE record_id = $1
`, fullRecordID); err != nil {
		t.Fatalf("populate Party projection optionals: %v", err)
	}
	deletedRecordID := recordIDs[2]
	if _, err := harness.DB.Exec(ctx, `UPDATE records SET deleted_at = now(), deleted_by_user_id = $2 WHERE record_id = $1`, deletedRecordID, actor.ID); err != nil {
		t.Fatalf("delete Party projection source: %v", err)
	}
	foreignRecordID := uuid.MustParse("20000000-0000-4000-8000-000000000000")
	if _, err := harness.DB.Exec(ctx, `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'party', $3, $3)
`, foreignRecordID, foreignIncident.ID, actor.ID); err != nil {
		t.Fatalf("seed foreign Party projection envelope: %v", err)
	}
	if _, err := harness.DB.Exec(ctx, `
INSERT INTO parties (record_id, incident_id, display_name, party_kind)
VALUES ($1, $2, 'Foreign Party', 'team')
`, foreignRecordID, foreignIncident.ID); err != nil {
		t.Fatalf("seed foreign Party projection source: %v", err)
	}

	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin Party projection source transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	contribution, err := parties.NewProjectionContribution()
	if err != nil {
		t.Fatalf("construct Party projection contribution: %v", err)
	}
	source := contribution.Source()
	if _, err := source.ListProjectionInputsTx(ctx, tx, incident.ID, nil, 0); err == nil {
		t.Fatal("zero Party projection page limit unexpectedly succeeded")
	}
	if _, err := source.ListProjectionInputsTx(ctx, tx, incident.ID, nil, 1001); err == nil {
		t.Fatal("oversized Party projection page limit unexpectedly succeeded")
	}
	first, err := source.ListProjectionInputsTx(ctx, tx, incident.ID, nil, 1)
	if err != nil {
		t.Fatalf("list first Party projection source page: %v", err)
	}
	if len(first.Inputs) != 1 || first.NextRecordID == nil || first.Inputs[0].RecordID != fullRecordID {
		t.Fatalf("unexpected first Party projection source page: %#v", first)
	}
	input := first.Inputs[0]
	if input.RowVersion != 7 || input.OrganizationName == nil || *input.OrganizationName != "Coordination" ||
		input.RoleTitle == nil || *input.RoleTitle != "Lead" ||
		input.PrimaryEmail == nil || *input.PrimaryEmail != "Owner@Example.Test" ||
		input.TimezoneName == nil || *input.TimezoneName != "US/Eastern" ||
		input.ExternalRef == nil || *input.ExternalRef != "EXT-Party" ||
		input.Notes == nil || *input.Notes != "Line one\nLine two" ||
		input.UpdatedAt.Location() != time.UTC {
		t.Fatalf("Party projection source lost typed values: %#v", input)
	}
	second, err := source.ListProjectionInputsTx(ctx, tx, incident.ID, first.NextRecordID, 1)
	if err != nil {
		t.Fatalf("list second Party projection source page: %v", err)
	}
	if len(second.Inputs) != 1 || second.NextRecordID != nil || second.Inputs[0].RecordID != recordIDs[0] {
		t.Fatalf("unexpected second Party projection source page: %#v", second)
	}
	if second.Inputs[0].OrganizationName != nil || second.Inputs[0].RoleTitle != nil ||
		second.Inputs[0].PrimaryEmail != nil || second.Inputs[0].TimezoneName != nil ||
		second.Inputs[0].ExternalRef != nil || second.Inputs[0].Notes != nil {
		t.Fatalf("Party projection nulls were not preserved: %#v", second.Inputs[0])
	}
	if _, found, err := source.LoadProjectionInputTx(ctx, tx, fullRecordID); err != nil || !found {
		t.Fatalf("load active Party projection source: found=%t err=%v", found, err)
	}
	if _, found, err := source.LoadProjectionInputTx(ctx, tx, deletedRecordID); err != nil || found {
		t.Fatalf("load deleted Party projection source: found=%t err=%v", found, err)
	}
	if _, found, err := source.LoadProjectionInputTx(ctx, tx, uuid.New()); err != nil || found {
		t.Fatalf("load absent Party projection source: found=%t err=%v", found, err)
	}
}

func TestAssessmentProjectionSourceEnumerationIsDeterministicAndKeysetPaged(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "projection-assessment-source-paging")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"assessment-projection-page@example.test",
		"Assessment Projection Page",
		"AssessmentProjectionPage1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-assessment-projection-page-incident",
		"IR-ASSESSMENT-PROJECTION-PAGE",
		"Assessment projection paging",
	)
	subjectID := uuid.New()
	entitytest.SeedHostRecord(
		t,
		harness.DB,
		incident.ID,
		actor.ID,
		subjectID,
		"Assessment projection subject",
		"assessment-projection-subject",
		"",
		"",
	)
	recordIDs := []uuid.UUID{
		uuid.MustParse("30000000-0000-4000-8000-000000000003"),
		uuid.MustParse("30000000-0000-4000-8000-000000000001"),
		uuid.MustParse("30000000-0000-4000-8000-000000000002"),
	}
	for _, recordID := range recordIDs {
		assessmenttest.SeedAssessment(
			t,
			harness.DB,
			incident.ID,
			actor.ID,
			recordID,
			subjectID,
			"host",
			"confirmed",
		)
	}

	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin assessment projection source paging: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	source := projectiontestsupport.MustAssessmentSource(t)
	first, err := source.ListProjectionInputsTx(ctx, tx, incident.ID, nil, 2)
	if err != nil {
		t.Fatalf("list first assessment projection source page: %v", err)
	}
	if len(first.Inputs) != 2 || first.NextRecordID == nil ||
		first.Inputs[0].RecordID.String() != "30000000-0000-4000-8000-000000000001" ||
		first.Inputs[1].RecordID.String() != "30000000-0000-4000-8000-000000000002" {
		t.Fatalf("unexpected first assessment projection page: %#v", first)
	}
	second, err := source.ListProjectionInputsTx(
		ctx,
		tx,
		incident.ID,
		first.NextRecordID,
		2,
	)
	if err != nil {
		t.Fatalf("list second assessment projection source page: %v", err)
	}
	if len(second.Inputs) != 1 || second.NextRecordID != nil ||
		second.Inputs[0].RecordID.String() != "30000000-0000-4000-8000-000000000003" {
		t.Fatalf("unexpected second assessment projection page: %#v", second)
	}
}

type commitFailDB struct {
	postgres.DB
}

func (db commitFailDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return commitFailTx{Tx: tx}, nil
}

type commitFailTx struct {
	pgx.Tx
}

func (commitFailTx) Commit(context.Context) error {
	return errors.New("injected commit failure")
}

func TestRebuildRestoreProjectionsClearsClaimsWhenCommitFails(t *testing.T) {
	harness := appsupport.StartStore(t, "projection-restore-commit-failure")
	failingDB := commitFailDB{DB: harness.DB}
	rebuilder := projectiontestsupport.MustBuild(t, failingDB).RecoveryPorts().Rebuilder
	result, err := rebuilder.RebuildRestoreProjections(context.Background(), validProjectionRebuildRequest())
	if err == nil || !strings.Contains(err.Error(), "injected commit failure") {
		t.Fatalf("commit failure error = %v", err)
	}
	if result.Status != restorecontract.ProjectionRebuildStatusFailed ||
		result.ReadinessOutcome != restorecontract.ProjectionReadinessIncomplete ||
		result.ReadinessSatisfied() {
		t.Fatalf("commit failure readiness = %#v", result)
	}
	if len(result.ProviderResults) == 0 {
		t.Fatal("commit failure omitted the ordered provider diagnostic prefix")
	}
	for _, providerResult := range result.ProviderResults {
		if providerResult.Status != restorecontract.ProjectionProviderResultFailed ||
			len(providerResult.RebuiltViewSchemaIDs) != 0 ||
			len(providerResult.RebuiltProjectionTables) != 0 {
			t.Fatalf("commit failure retained rebuilt claims: %#v", providerResult)
		}
	}
}

func TestRebuildRestoreProjectionsReportsProviderResultsAndReplacesStaleRows(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "projection-restore-rebuild-result")
	rebuilder := projectiontestsupport.MustBuild(t, harness.DB).RecoveryPorts().Rebuilder
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "projection-restore@example.test", "Projection Restore", "ProjectionRestore1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-projection-restore-incident", "IR-PROJECTION-RESTORE", "Projection restore")
	timelineRecordID := uuid.New()
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelineRecordID)

	insertStaleTimelineProjectionRow(t, harness.DB, incident.ID, timelineRecordID)

	operationID := uuid.New()
	request := restorecontract.ProjectionRebuildRequest{
		RestoreOperationID:     operationID,
		RestoredSourceStateRef: "backup_set:" + uuid.NewString(),
		RebuildScope:           restorecontract.ProjectionRebuildScopeAllActiveProviders,
		ProviderRegistryRef:    restorecontract.ProviderRegistryRefCodeBacked,
	}
	result, err := rebuilder.RebuildRestoreProjections(ctx, request)
	if err != nil {
		t.Fatalf("rebuild restore projections: %v", err)
	}
	if result.RestoreOperationID != operationID ||
		result.Status != restorecontract.ProjectionRebuildStatusSucceeded ||
		result.ReadinessOutcome != restorecontract.ProjectionReadinessReady ||
		!result.ReadinessSatisfied() {
		t.Fatalf("restore projection rebuild result not ready: %#v", result)
	}
	wantProviders := []string{"timeline", "host", "identity", "indicator", "assessment", "artifact", "evidence", "party", "task_request", "decision"}
	wantViews := map[string][]string{
		"timeline":     {"cartulary.view.timeline.v2"},
		"host":         {"cartulary.view.hosts.v1"},
		"identity":     {"cartulary.view.identities.v1"},
		"indicator":    {"cartulary.view.indicators.v1"},
		"assessment":   {"cartulary.view.assessments.v1"},
		"artifact":     {"cartulary.view.notes.v1", "cartulary.view.comm_log.v1", "cartulary.view.handoff.v1", "cartulary.view.status_review.v1", "cartulary.view.lesson.v1", "cartulary.view.findings.v1", "cartulary.view.investigative_queries.v1", "cartulary.view.forensic_keywords.v1"},
		"evidence":     {"cartulary.view.evidence.v1"},
		"party":        {"cartulary.view.parties.v1"},
		"task_request": {"cartulary.view.task_requests.v1"},
		"decision":     {"cartulary.view.decisions.v1"},
	}
	if got := projectionProviderResultKeys(result.ProviderResults); !reflect.DeepEqual(got, wantProviders) {
		t.Fatalf("provider result order got %#v want %#v", got, wantProviders)
	}
	for _, providerResult := range result.ProviderResults {
		if providerResult.Status != restorecontract.ProjectionProviderResultSucceeded {
			t.Fatalf("provider %s status got %q", providerResult.ProviderKey, providerResult.Status)
		}
		if providerResult.IncidentCount != 1 {
			t.Fatalf("provider %s incident count got %d want 1", providerResult.ProviderKey, providerResult.IncidentCount)
		}
		if !reflect.DeepEqual(providerResult.RebuiltViewSchemaIDs, wantViews[providerResult.ProviderKey]) {
			t.Fatalf("provider %s rebuilt views got %#v want %#v", providerResult.ProviderKey, providerResult.RebuiltViewSchemaIDs, wantViews[providerResult.ProviderKey])
		}
		if len(providerResult.RebuiltProjectionTables) == 0 {
			t.Fatalf("provider %s did not report rebuilt projection tables", providerResult.ProviderKey)
		}
		for _, tableResult := range providerResult.RebuiltProjectionTables {
			if tableResult.ProjectionTableID == "" || tableResult.RowCount < 0 {
				t.Fatalf("provider %s invalid table result %#v", providerResult.ProviderKey, tableResult)
			}
		}
	}

	var synopsis string
	if err := harness.DB.QueryRow(ctx, `
SELECT activity_synopsis_text
  FROM timeline_grid_projection
 WHERE record_id = $1
`, timelineRecordID).Scan(&synopsis); err != nil {
		t.Fatalf("load rebuilt timeline projection row: %v", err)
	}
	if synopsis != "record-support-source-row" {
		t.Fatalf("stale timeline projection row was not replaced, got %q", synopsis)
	}

	retry, err := rebuilder.RebuildRestoreProjections(ctx, request)
	if err != nil {
		t.Fatalf("retry restore projection rebuild: %v", err)
	}
	if !reflect.DeepEqual(retry.ProviderResults, result.ProviderResults) {
		t.Fatalf("retry provider results changed:\nfirst %#v\nretry %#v", result.ProviderResults, retry.ProviderResults)
	}
}

func validProjectionRebuildRequest() restorecontract.ProjectionRebuildRequest {
	return restorecontract.ProjectionRebuildRequest{
		RestoreOperationID:     uuid.New(),
		RestoredSourceStateRef: "backup_set:" + uuid.NewString(),
		RebuildScope:           restorecontract.ProjectionRebuildScopeAllActiveProviders,
		ProviderRegistryRef:    restorecontract.ProviderRegistryRefCodeBacked,
	}
}

func insertStaleTimelineProjectionRow(t *testing.T, db interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}, incidentID uuid.UUID, recordID uuid.UUID) {
	t.Helper()

	staleAt := time.Date(2026, 7, 1, 14, 30, 0, 0, time.UTC)
	if _, err := db.Exec(context.Background(), `
INSERT INTO timeline_grid_projection (
    record_id,
    incident_id,
    row_version,
    recorded_at,
    edited_at,
    capture_state,
    date_entered_text,
    activity_sort_ts,
    date_entered_sort_day,
    activity_time_pair_state,
    activity_synopsis_text
)
VALUES ($1, $2, 1, $3::timestamptz, $3::timestamptz, 'reviewed', '2026-07-01T14:30:00Z', $3::timestamptz, $3::date, 'disabled', 'stale projection row')
`, recordID, incidentID, staleAt); err != nil {
		t.Fatalf("insert stale timeline projection row: %v", err)
	}
}

func projectionProviderResultKeys(results []restorecontract.ProjectionProviderResult) []string {
	keys := make([]string, 0, len(results))
	for _, result := range results {
		keys = append(keys, result.ProviderKey)
	}
	return keys
}
