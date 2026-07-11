package projections_test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	recordstoretest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
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
}

func TestRebuildRestoreProjectionsReportsProviderResultsAndReplacesStaleRows(t *testing.T) {
	ctx := context.Background()
	harness := recordstoretest.StartStore(t, "projection-restore-rebuild-result")
	rebuilder := projections.NewRestoreRebuilder(harness.DB)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "projection-restore@example.test", "Projection Restore", "ProjectionRestore1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-projection-restore-incident", "IR-PROJECTION-RESTORE", "Projection restore")
	timelineRecordID := uuid.New()
	recordstoretest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelineRecordID)

	insertStaleTimelineProjectionRow(t, harness.DB, incident.ID, timelineRecordID)

	operationID := uuid.New()
	result, err := rebuilder.RebuildRestoreProjections(ctx, restorecontract.ProjectionRebuildRequest{
		RestoreOperationID:     operationID,
		RestoredSourceStateRef: "backup_set:" + uuid.NewString(),
		RebuildScope:           restorecontract.ProjectionRebuildScopeAllActiveProviders,
		ProviderRegistryRef:    restorecontract.ProviderRegistryRefCodeBacked,
	})
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
		if len(providerResult.RowCountsByTable) == 0 {
			t.Fatalf("provider %s did not report row counts", providerResult.ProviderKey)
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
VALUES ($1, $2, 1, $3, $3, 'reviewed', '2026-07-01T14:30:00Z', $3, $3::date, 'disabled', 'stale projection row')
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
