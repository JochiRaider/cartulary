package revisions_test

import (
	"context"
	"database/sql"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIndicatorChildHistoryRollback_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "history_revision-i-7-06-indicator-child-rollback")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, _ := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-I706")
	store, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:    harness.Server.Runtime.Postgres,
		Revisions:   harness.Server.Runtime.Revisions.Appender(),
		Projections: harness.Server.Runtime.Timeline.ProjectionCoordinator,
		SourceText:  indicatorassembly.NewSourceTextPort(harness.Server.Runtime.Timeline.ProjectionCoordinator),
	})
	if err != nil {
		t.Fatalf("compose Indicator test owner: %v", err)
	}
	actor := authn.UserRecord{ID: actorID}

	t.Run("resolved observation create reversal tombstones and invalidates every affected record", func(t *testing.T) {
		sourceID := uuid.New()
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, actorID, sourceID)
		indicatorID := seedIndicatorChildRecord(t, harness.DB, incidentID, actorID, "create")
		createdAt := time.Date(2026, 7, 9, 15, 0, 0, 0, time.UTC)
		created, err := store.CreateIndicatorObservation(context.Background(), actor, indicatorChildObservationParams(
			incidentID, sourceID, "timeline.raw_activity_text", &indicatorID, "txn-i-7-06-observation-create",
		))
		if err != nil {
			t.Fatalf("create resolved observation: %v", err)
		}
		observation, changeSetID := created.Observation, created.ChangeSetID
		var sourceItem map[string]any
		for _, recordID := range []uuid.UUID{sourceID, indicatorID} {
			item := historyItemForChangeSetTarget(t, harness, login, recordID, changeSetID, "indicator_observation", observation.ObservationID.String())
			requireHistoryActionContains(t, item, "history_entry")
			if recordID == sourceID {
				sourceItem = item
			}
		}
		ref := historyEntryRefForTarget(t, harness, login, sourceID, "indicator_observation", observation.ObservationID.String())
		if ref == "" || ref != stringField(t, sourceItem, "history_entry_ref") {
			t.Fatalf("observation selector was not stable across history reads: item=%#v ref=%q", sourceItem, ref)
		}
		lockTx, err := harness.DB.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatalf("begin Indicator lock holder: %v", err)
		}
		if _, err := lockTx.ExecContext(context.Background(), `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE`, indicatorID); err != nil {
			_ = lockTx.Rollback()
			t.Fatalf("hold affected Indicator lock: %v", err)
		}
		locked := rollbackRecord(t, harness, login, sourceID, map[string]any{
			"base_row_version": 2,
			"client_txn_id":    "txn-i-7-06-observation-locked",
			"target":           map[string]any{"kind": "history_entry", "history_entry_ref": ref},
		})
		httptestx.RequireErrorEnvelope(t, locked, 409, "record_locked")
		if err := lockTx.Rollback(); err != nil {
			t.Fatalf("release affected Indicator lock: %v", err)
		}
		asserttest.AwaitIncidentStreamIdle(t, asserttest.SQLDatabase(harness.DB), incidentID.String())
		changes, unsubscribe := harness.Server.Runtime.CollaborationHub.SubscribeIncident(incidentID, 8)
		defer unsubscribe()
		httptestx.SetClockFixed(t, harness.Server, createdAt.Add(time.Minute))
		body := map[string]any{
			"base_row_version": 2,
			"client_txn_id":    "txn-i-7-06-observation-create-rollback",
			"target":           map[string]any{"kind": "history_entry", "history_entry_ref": ref},
		}
		payload := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, sourceID, body), 200)["data"].(map[string]any)
		requireAffectedRecords(t, payload, sourceID, indicatorID)
		firstChange := asserttest.AwaitRecordChange(t, changes, 5*time.Second)
		secondChange := asserttest.AwaitRecordChange(t, changes, 5*time.Second)
		received := []uuid.UUID{firstChange.RecordID, secondChange.RecordID}
		if !sameUUIDSet(received, []uuid.UUID{sourceID, indicatorID}) {
			t.Fatalf("ordinary rollback events = %v", received)
		}
		for _, change := range []struct {
			recordID uuid.UUID
			keys     []string
		}{{firstChange.RecordID, firstChange.ChangedFieldKeys}, {secondChange.RecordID, secondChange.ChangedFieldKeys}} {
			if change.recordID == sourceID && len(change.keys) != 0 {
				t.Fatalf("source event reported unchanged text fields = %v", change.keys)
			}
			if change.recordID == indicatorID && (!slices.Contains(change.keys, "indicator.observation_count") || !slices.Contains(change.keys, "indicator.first_observed_at") || !slices.Contains(change.keys, "indicator.last_observed_at")) {
				t.Fatalf("Indicator event changed keys = %v", change.keys)
			}
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM indicator_observations WHERE indicator_observation_id = $1 AND deleted_at IS NOT NULL AND deleted_by_user_id = $2 AND row_version = 2`, observation.ObservationID, actorID) != 1 {
			t.Fatal("observation create reversal did not retain a versioned tombstone")
		}
		if countRows(t, harness.DB, `SELECT observation_count FROM indicator_grid_projection WHERE record_id = $1`, indicatorID) != 0 {
			t.Fatal("tombstoned observation remained in indicator aggregate")
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM records WHERE record_id IN ($1, $2) AND row_version = 3`, sourceID, indicatorID) != 2 {
			t.Fatal("observation rollback did not advance every affected record envelope")
		}
		rollbackChangeSetID := payload["rollback_change_set_id"].(string)
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1 AND record_id IN ($2, $3)`, rollbackChangeSetID, sourceID, indicatorID) != 2 {
			t.Fatal("observation rollback did not append every affected record revision")
		}
		replay := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, sourceID, body), 200)["data"].(map[string]any)
		if replay["rollback_change_set_id"] != rollbackChangeSetID || countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE source = 'rollback' AND client_txn_id = 'txn-i-7-06-observation-create-rollback'`) != 1 {
			t.Fatalf("observation rollback replay was not idempotent: first=%#v replay=%#v", payload, replay)
		}
	})

	t.Run("observation resolution reversal restores exact unresolved state", func(t *testing.T) {
		sourceID := uuid.New()
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, actorID, sourceID)
		indicatorID := seedIndicatorChildRecord(t, harness.DB, incidentID, actorID, "resolve")
		createdAt := time.Date(2026, 7, 9, 16, 0, 0, 0, time.UTC)
		created, err := store.CreateIndicatorObservation(context.Background(), actor, indicatorChildObservationParams(
			incidentID, sourceID, "timeline.activity_synopsis_text", nil, "txn-i-7-06-observation-unresolved",
		))
		if err != nil {
			t.Fatalf("create unresolved observation: %v", err)
		}
		observation := created.Observation
		resolvedResult, err := store.ResolveIndicatorObservation(context.Background(), actor, indicators.IndicatorObservationResolveParams{
			ObservationID: observation.ObservationID, ResolvedIndicatorRecordID: indicatorID, BaseRowVersion: 1,
			ClientTxnID: "txn-i-7-06-observation-resolve", RequestID: "req-i-7-06-observation-resolve", RequestHash: []byte("hash-i-7-06-observation-resolve"),
		})
		if err != nil || resolvedResult.Observation.RowVersion != 2 {
			t.Fatalf("resolve observation = %#v, %v", resolvedResult, err)
		}
		changeSetID := resolvedResult.ChangeSetID
		for _, recordID := range []uuid.UUID{sourceID, indicatorID} {
			_ = historyItemForChangeSetTarget(t, harness, login, recordID, changeSetID, "indicator_observation", observation.ObservationID.String())
		}
		ref := historyEntryRefForTarget(t, harness, login, indicatorID, "indicator_observation", observation.ObservationID.String())
		httptestx.SetClockFixed(t, harness.Server, createdAt.Add(2*time.Minute))
		payload := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, indicatorID, map[string]any{
			"base_row_version": 2,
			"client_txn_id":    "txn-i-7-06-observation-resolve-rollback",
			"target":           map[string]any{"kind": "history_entry", "history_entry_ref": ref},
		}), 200)["data"].(map[string]any)
		requireAffectedRecords(t, payload, sourceID, indicatorID)
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM indicator_observations WHERE indicator_observation_id = $1 AND resolution_status = 'unresolved' AND resolved_indicator_record_id IS NULL AND resolved_by_user_id IS NULL AND resolved_at IS NULL AND deleted_at IS NULL AND row_version = 3`, observation.ObservationID) != 1 {
			t.Fatal("resolution rollback did not restore the exact unresolved state")
		}
		if countRows(t, harness.DB, `SELECT observation_count FROM indicator_grid_projection WHERE record_id = $1`, indicatorID) != 0 {
			t.Fatal("resolution rollback did not remove the observation from canonical aggregates")
		}
	})

	t.Run("re-resolution reversal protects and restores old and new canonical indicators", func(t *testing.T) {
		sourceID := uuid.New()
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, actorID, sourceID)
		oldIndicatorID := seedIndicatorChildRecord(t, harness.DB, incidentID, actorID, "reresolve-old")
		newIndicatorID := seedIndicatorChildRecord(t, harness.DB, incidentID, actorID, "reresolve-new")
		createdAt := time.Date(2026, 7, 9, 16, 30, 0, 0, time.UTC)
		created, err := store.CreateIndicatorObservation(context.Background(), actor, indicatorChildObservationParams(
			incidentID, sourceID, "timeline.activity_synopsis_text", &oldIndicatorID, "txn-i-7-06-observation-reresolve-create",
		))
		if err != nil {
			t.Fatalf("create initially resolved observation: %v", err)
		}
		observation := created.Observation
		resolved, err := store.ResolveIndicatorObservation(context.Background(), actor, indicators.IndicatorObservationResolveParams{
			ObservationID: observation.ObservationID, ResolvedIndicatorRecordID: newIndicatorID, BaseRowVersion: 1,
			ClientTxnID: "txn-i-7-06-observation-reresolve", RequestID: "req-i-7-06-observation-reresolve", RequestHash: []byte("hash-i-7-06-observation-reresolve"),
		})
		if err != nil {
			t.Fatalf("re-resolve observation: %v", err)
		}
		changeSetID := resolved.ChangeSetID
		for _, recordID := range []uuid.UUID{sourceID, oldIndicatorID, newIndicatorID} {
			_ = historyItemForChangeSetTarget(t, harness, login, recordID, changeSetID, "indicator_observation", observation.ObservationID.String())
		}
		if countRows(t, harness.DB, `SELECT observation_count FROM indicator_grid_projection WHERE record_id = $1`, oldIndicatorID) != 0 || countRows(t, harness.DB, `SELECT observation_count FROM indicator_grid_projection WHERE record_id = $1`, newIndicatorID) != 1 {
			t.Fatal("re-resolution did not move active aggregate effects to the new Indicator")
		}
		ref := historyEntryRefForTarget(t, harness, login, newIndicatorID, "indicator_observation", observation.ObservationID.String())
		httptestx.SetClockFixed(t, harness.Server, createdAt.Add(2*time.Minute))
		payload := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, newIndicatorID, map[string]any{
			"base_row_version": 2,
			"client_txn_id":    "txn-i-7-06-observation-reresolve-rollback",
			"target":           map[string]any{"kind": "history_entry", "history_entry_ref": ref},
		}), 200)["data"].(map[string]any)
		requireAffectedRecords(t, payload, sourceID, oldIndicatorID, newIndicatorID)
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM indicator_observations WHERE indicator_observation_id = $1 AND resolved_indicator_record_id = $2 AND resolution_status = 'resolved' AND row_version = 3 AND deleted_at IS NULL`, observation.ObservationID, oldIndicatorID) != 1 {
			t.Fatal("re-resolution rollback did not restore the old canonical Indicator")
		}
		if countRows(t, harness.DB, `SELECT observation_count FROM indicator_grid_projection WHERE record_id = $1`, oldIndicatorID) != 1 || countRows(t, harness.DB, `SELECT observation_count FROM indicator_grid_projection WHERE record_id = $1`, newIndicatorID) != 0 {
			t.Fatal("re-resolution rollback did not restore old/new aggregate effects")
		}
	})

	t.Run("lifecycle create reversal tombstones and clears lifecycle projection", func(t *testing.T) {
		indicatorID := seedIndicatorChildRecord(t, harness.DB, incidentID, actorID, "interval")
		createdAt := time.Date(2026, 7, 9, 17, 0, 0, 0, time.UTC)
		created, err := store.AppendIndicatorLifecycleInterval(context.Background(), actor, indicatorChildLifecycleParams(
			incidentID, indicatorID, 1, createdAt, "txn-i-7-06-interval-create",
		))
		if err != nil {
			t.Fatalf("create lifecycle interval: %v", err)
		}
		interval, changeSetID := created.Interval, created.ChangeSetID
		item := historyItemForChangeSetTarget(t, harness, login, indicatorID, changeSetID, "indicator_state_interval", interval.IntervalID.String())
		requireHistoryActionContains(t, item, "history_entry")
		ref := historyEntryRefForTarget(t, harness, login, indicatorID, "indicator_state_interval", interval.IntervalID.String())
		httptestx.SetClockFixed(t, harness.Server, createdAt.Add(time.Minute))
		httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, indicatorID, map[string]any{
			"base_row_version": 2,
			"client_txn_id":    "txn-i-7-06-interval-create-rollback",
			"target":           map[string]any{"kind": "history_entry", "history_entry_ref": ref},
		}), 200)
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM indicator_state_intervals WHERE indicator_state_interval_id = $1 AND deleted_at IS NOT NULL AND deleted_by_user_id = $2 AND row_version = 2`, interval.IntervalID, actorID) != 1 {
			t.Fatal("interval create reversal did not retain a versioned tombstone")
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM indicator_grid_projection WHERE record_id = $1 AND lifecycle_summary IS NULL`, indicatorID) != 1 {
			t.Fatal("tombstoned interval remained in lifecycle selection")
		}
	})
}

func indicatorChildObservationParams(incidentID uuid.UUID, sourceID uuid.UUID, fieldKey string, resolvedID *uuid.UUID, clientTxnID string) indicators.IndicatorObservationCreateParams {
	return indicators.IndicatorObservationCreateParams{
		IncidentID: incidentID, SourceRecordID: sourceID, BaseRowVersion: 1,
		SourceFieldKey: fieldKey, SpanStartByte: 0, SpanEndByte: len("record-support-source-row"),
		ResolvedIndicatorRecordID: resolvedID, ClientTxnID: clientTxnID,
		RequestID: "req-" + clientTxnID, RequestHash: []byte("hash-" + clientTxnID),
	}
}

func indicatorChildLifecycleParams(incidentID uuid.UUID, indicatorID uuid.UUID, baseRowVersion int64, validFrom time.Time, clientTxnID string) indicators.IndicatorLifecycleAppendParams {
	return indicators.IndicatorLifecycleAppendParams{
		IncidentID: incidentID, IndicatorRecordID: indicatorID, BaseRowVersion: baseRowVersion,
		LifecycleState: "active", ValidFrom: validFrom, SupportRefs: []uuid.UUID{},
		ClientTxnID: clientTxnID, RequestID: "req-" + clientTxnID, RequestHash: []byte("hash-" + clientTxnID),
	}
}

func seedIndicatorChildRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, suffix string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	value := "history_revision-" + suffix + ".example.test"
	indicatortest.SeedRecord(t, db, incidentID, actorID, recordID, "domain_name", "atomic", value)
	return recordID
}

func sameUUIDSet(left []uuid.UUID, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	set := make(map[uuid.UUID]int, len(left))
	for _, value := range left {
		set[value]++
	}
	for _, value := range right {
		set[value]--
	}
	for _, count := range set {
		if count != 0 {
			return false
		}
	}
	return true
}
