package revisions_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	referencefixture "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func expectedIntervalHistory(id, incidentID, recordID, actorID uuid.UUID, createdAt, validFrom time.Time) map[string]any {
	return map[string]any{"indicator_state_interval_id": id.String(), "incident_id": incidentID.String(), "indicator_record_id": recordID.String(), "lifecycle_state": "active",
		"valid_from": validFrom.UTC().Format(time.RFC3339Nano), "valid_to": nil, "confidence": nil, "rationale": nil, "support_refs": []string{}, "assessor": nil,
		"assessed_at": createdAt.UTC().Format(time.RFC3339Nano), "row_version": 1, "created_by_user_id": actorID.String(), "created_at": createdAt.UTC().Format(time.RFC3339Nano), "deleted_at": nil, "deleted_by_user_id": nil}
}

func requireIntervalHistory(t testing.TB, db *sql.DB, changeSetID, operation string, id uuid.UUID, before, after map[string]any) {
	t.Helper()
	encode := func(value map[string]any) any {
		if value == nil {
			return nil
		}
		payload, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return payload
	}
	var count int
	var exact bool
	if err := db.QueryRow(`SELECT count(*),coalesce(bool_and(operation_kind=$3 AND coalesce(before_value,'null'::jsonb)=coalesce($4::jsonb,'null'::jsonb) AND after_value=$5::jsonb),false)
FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='indicator_state_interval' AND target_id=$2`, changeSetID, id.String(), operation, encode(before), encode(after)).Scan(&count, &exact); err != nil || count != 1 || !exact {
		t.Fatalf("complete interval history count=%d exact=%v error=%v", count, exact, err)
	}
	if err := db.QueryRow(`SELECT (to_jsonb(i)-'valid_from'-'valid_to'-'assessed_at'-'created_at'-'deleted_at')=($2::jsonb-'valid_from'-'valid_to'-'assessed_at'-'created_at'-'deleted_at')
AND valid_from=($2::jsonb->>'valid_from')::timestamptz AND valid_to IS NOT DISTINCT FROM ($2::jsonb->>'valid_to')::timestamptz
AND assessed_at=($2::jsonb->>'assessed_at')::timestamptz AND created_at=($2::jsonb->>'created_at')::timestamptz AND deleted_at IS NOT DISTINCT FROM ($2::jsonb->>'deleted_at')::timestamptz
FROM indicator_state_intervals i WHERE indicator_state_interval_id=$1`, id, encode(after)).Scan(&exact); err != nil || !exact {
		t.Fatalf("complete interval source exact=%v error=%v", exact, err)
	}
}

func TestCompleteIndicatorIntervalHistory_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "complete-interval-history", TestRouteMode: httptestx.TestRouteModeDisabled})
	var clockBase time.Time
	if err := harness.DB.QueryRow(`SELECT clock_timestamp() + interval '5 minutes'`).Scan(&clockBase); err != nil {
		t.Fatal(err)
	}
	login, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incidentID, _ := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-INTERVAL-HISTORY")
	application, err := indicators.NewApplication(indicators.ApplicationDependencies{
		ReferencePacks:  referencefixture.IndicatorRegistry{},
		Postgres:        harness.Pool,
		Idempotency:     indicatorassembly.NewIdempotencyPort(authn.NewStore(harness.Pool)),
		IncidentState:   admission.NewChecker(harness.Pool),
		Revisions:       harness.Revisions.Appender(),
		Collaboration:   harness.Collaboration.RecordChanges(),
		RecordEnvelopes: records.NewStore(harness.Pool),
		Projections:     harness.Projections.IndicatorProjectionPort(),
		SourceText:      harness.IndicatorSourceText,
		Clock:           func() time.Time { return clockBase },
	})
	if err != nil {
		t.Fatalf("compose Indicator test owner: %v", err)
	}
	httptestx.SetClockFixed(t, harness.Server, clockBase)
	actor := authn.UserRecord{ID: actorID}
	indicatorID := seedIndicatorChildRecord(t, harness.DB, incidentID, actorID, "interval")
	createdAt := clockBase.Add(3 * time.Minute)
	created, err := application.AppendIndicatorLifecycleInterval(context.Background(), actor.ID, indicatorChildLifecycleParams(
		incidentID, indicatorID, 1, createdAt, "txn-i-7-06-interval-create",
	))
	if err != nil {
		t.Fatalf("create lifecycle interval: %v", err)
	}
	interval, changeSetID := created.Interval, created.ChangeSetID
	active := expectedIntervalHistory(interval.IntervalID, incidentID, indicatorID, actorID, clockBase, createdAt)
	requireIntervalHistory(t, harness.DB, changeSetID.String(), "create", interval.IntervalID, nil, active)

	item := historyItemForChangeSetTarget(t, harness, login, indicatorID, changeSetID, "indicator_state_interval", interval.IntervalID.String())
	requireHistoryActionContains(t, item, "history_entry")
	ref := historyEntryRefForTarget(t, harness, login, indicatorID, "indicator_state_interval", interval.IntervalID.String())
	requireIndicatorChildCurrentness(t, harness.DB, incidentID, indicatorID, interval.IntervalID,
		"indicator_state_intervals", "indicator_state_interval_id", map[string]any{
			"lifecycle_state": "retired", "valid_from": createdAt.Add(-time.Second).UTC().Format(time.RFC3339Nano),
			"valid_to": createdAt.Add(time.Hour).UTC().Format(time.RFC3339Nano), "confidence": 75,
			"rationale": "different rationale", "support_refs": []string{indicatorID.String()}, "assessor": "reviewer",
			"assessed_at": clockBase.Add(time.Second).UTC().Format(time.RFC3339Nano),
			"created_at":  clockBase.Add(-time.Second).UTC().Format(time.RFC3339Nano),
		}, func() *http.Response {
			return rollbackRecord(t, harness, login, indicatorID, map[string]any{
				"base_row_version": 2, "client_txn_id": "txn-i-7-06-interval-create-rollback",
				"target": map[string]any{"kind": "history_entry", "history_entry_ref": ref},
			})
		})
	httptestx.SetClockFixed(t, harness.Server, createdAt.Add(time.Minute))
	rollbackData := httptestx.RequireSuccessData(t, rollbackRecord(t, harness, login, indicatorID, map[string]any{
		"base_row_version": 2,
		"client_txn_id":    "txn-i-7-06-interval-create-rollback",
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": ref},
	}), 200)
	deleted := expectedIntervalHistory(interval.IntervalID, incidentID, indicatorID, actorID, clockBase, createdAt)
	deleted["row_version"] = 2
	deleted["deleted_at"] = createdAt.Add(time.Minute).UTC().Format(time.RFC3339Nano)
	deleted["deleted_by_user_id"] = actorID.String()
	requireIntervalHistory(t, harness.DB, rollbackData["rollback_change_set_id"].(string), "rollback", interval.IntervalID, active, deleted)

	if countRows(t, harness.DB, `SELECT COUNT(*) FROM indicator_state_intervals WHERE indicator_state_interval_id = $1 AND deleted_at IS NOT NULL AND deleted_by_user_id = $2 AND row_version = 2`, interval.IntervalID, actorID) != 1 {
		t.Fatal("interval create reversal did not retain a versioned tombstone")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM indicator_grid_projection WHERE record_id = $1 AND lifecycle_summary IS NULL`, indicatorID) != 1 {
		t.Fatal("tombstoned interval remained in lifecycle selection")
	}
}
