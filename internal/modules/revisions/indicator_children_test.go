package revisions_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/records"
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
		Postgres:        harness.Pool,
		Revisions:       harness.Revisions.Appender(),
		RecordEnvelopes: records.NewStore(harness.Pool),
		Projections:     harness.Projections.IndicatorProjectionPort(),
		SourceText:      harness.IndicatorSourceText,
		Clock:           func() time.Time { return time.Date(2026, 7, 9, 15, 0, 0, 0, time.UTC) },
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
		changes, unsubscribe := harness.Collaboration.SubscribeIncident(incidentID, 8)
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
			ClientTxnID: "txn-i-7-06-observation-resolve", RequestID: "req-i-7-06-observation-resolve",
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
			ClientTxnID: "txn-i-7-06-observation-reresolve", RequestID: "req-i-7-06-observation-reresolve",
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

	t.Run("row rollback rejects malformed retained history without durable effects", func(t *testing.T) {
		malformed := []struct {
			name   string
			mutate func(map[string]any, uuid.UUID, uuid.UUID)
		}{
			{name: "record identity mismatch", mutate: func(source map[string]any, _ uuid.UUID, _ uuid.UUID) {
				source["record_id"] = uuid.New().String()
			}},
			{name: "incident identity mismatch", mutate: func(source map[string]any, _ uuid.UUID, _ uuid.UUID) {
				source["incident_id"] = uuid.New().String()
			}},
			{name: "dedupe mismatch", mutate: func(source map[string]any, _ uuid.UUID, _ uuid.UUID) {
				source["dedupe_key"] = strings.Repeat("f", 64)
			}},
			{name: "malformed hash pair", mutate: func(source map[string]any, _ uuid.UUID, _ uuid.UUID) {
				source["hash_algorithm"] = "sha256"
			}},
			{name: "noncanonical presentation", mutate: func(source map[string]any, _ uuid.UUID, _ uuid.UUID) {
				source["display_value"] = " retained-malformed.example.test "
			}},
			{name: "unknown source member", mutate: func(source map[string]any, _ uuid.UUID, _ uuid.UUID) {
				source["row_version"] = float64(1)
			}},
		}
		for index, testCase := range malformed {
			t.Run(testCase.name, func(t *testing.T) {
				suffix := "malformed-" + string(rune('a'+index))
				current := indicatorRollbackSource(uuid.Nil, incidentID, "current-"+suffix+".example.test")
				retained := indicatorRollbackSource(uuid.Nil, incidentID, "retained-"+suffix+".example.test")
				recordID, historyRef := seedIndicatorRowRollback(t, harness.DB, incidentID, actorID, suffix, current, retained)
				testCase.mutate(retained, recordID, incidentID)
				updateIndicatorRollbackBeforeValue(t, harness.DB, historyRef, indicatorRollbackSnapshot(recordID, incidentID, 1, retained))
				before := indicatorRollbackDurableState(t, harness.DB, incidentID, recordID)
				requireRollbackReasonCode(t, rollbackRecord(t, harness, login, recordID, map[string]any{
					"base_row_version": 2,
					"client_txn_id":    "txn-i-7-06-indicator-row-" + suffix,
					"target":           map[string]any{"kind": "history_entry", "history_entry_ref": historyRef},
				}), "target_not_reversible")
				after := indicatorRollbackDurableState(t, harness.DB, incidentID, recordID)
				if after != before {
					t.Fatalf("malformed Indicator rollback changed durable state:\nbefore=%s\nafter=%s", before, after)
				}
			})
		}
	})

	t.Run("row rollback restores full partial and explicitly cleared source patches", func(t *testing.T) {
		const hashValue = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		cases := []struct {
			name     string
			current  func(uuid.UUID) map[string]any
			retained func(uuid.UUID) map[string]any
			expected func(uuid.UUID) map[string]any
		}{
			{
				name: "full snapshot rekeys and restores representations",
				current: func(recordID uuid.UUID) map[string]any {
					return indicatorRollbackSource(recordID, incidentID, "current-full.example.test")
				},
				retained: func(recordID uuid.UUID) map[string]any {
					source := indicatorRollbackSource(recordID, incidentID, "retained-full.example.test")
					source["defanged_value"] = "retained-full[.]example[.]test"
					source["hash_algorithm"] = "sha256"
					source["hash_value"] = hashValue
					source["stix_pattern"] = "[domain-name:value = 'retained-full.example.test']"
					refreshIndicatorRollbackDedupe(source)
					return source
				},
				expected: func(recordID uuid.UUID) map[string]any {
					source := indicatorRollbackSource(recordID, incidentID, "retained-full.example.test")
					source["defanged_value"] = "retained-full[.]example[.]test"
					source["hash_algorithm"] = "sha256"
					source["hash_value"] = hashValue
					source["stix_pattern"] = "[domain-name:value = 'retained-full.example.test']"
					refreshIndicatorRollbackDedupe(source)
					return source
				},
			},
			{
				name: "partial patch preserves omitted representations",
				current: func(recordID uuid.UUID) map[string]any {
					source := indicatorRollbackSource(recordID, incidentID, "current-partial.example.test")
					source["defanged_value"] = "current-partial[.]example[.]test"
					source["hash_algorithm"] = "sha256"
					source["hash_value"] = hashValue
					source["stix_pattern"] = "[domain-name:value = 'current-partial.example.test']"
					refreshIndicatorRollbackDedupe(source)
					return source
				},
				retained: func(recordID uuid.UUID) map[string]any {
					source := map[string]any{
						"record_id": recordID.String(), "incident_id": incidentID.String(),
						"display_value": "retained-partial.example.test", "normalized_value": "retained-partial.example.test",
					}
					source["dedupe_key"] = indicatorRollbackDedupe("domain_name", "atomic", "retained-partial.example.test", "retained-partial.example.test", "sha256", hashValue)
					return source
				},
				expected: func(recordID uuid.UUID) map[string]any {
					source := indicatorRollbackSource(recordID, incidentID, "retained-partial.example.test")
					source["defanged_value"] = "current-partial[.]example[.]test"
					source["hash_algorithm"] = "sha256"
					source["hash_value"] = hashValue
					source["stix_pattern"] = "[domain-name:value = 'current-partial.example.test']"
					refreshIndicatorRollbackDedupe(source)
					return source
				},
			},
			{
				name: "explicit nulls clear nullable representations",
				current: func(recordID uuid.UUID) map[string]any {
					source := indicatorRollbackSource(recordID, incidentID, "current-clear.example.test")
					source["defanged_value"] = "current-clear[.]example[.]test"
					source["hash_algorithm"] = "sha256"
					source["hash_value"] = hashValue
					source["stix_pattern"] = "[domain-name:value = 'current-clear.example.test']"
					refreshIndicatorRollbackDedupe(source)
					return source
				},
				retained: func(recordID uuid.UUID) map[string]any {
					return map[string]any{
						"record_id": recordID.String(), "incident_id": incidentID.String(),
						"dedupe_key":     indicatorRollbackDedupe("domain_name", "atomic", "current-clear.example.test", "current-clear.example.test", "", ""),
						"defanged_value": nil, "hash_algorithm": nil, "hash_value": nil, "stix_pattern": nil,
					}
				},
				expected: func(recordID uuid.UUID) map[string]any {
					return indicatorRollbackSource(recordID, incidentID, "current-clear.example.test")
				},
			},
		}
		for index, testCase := range cases {
			t.Run(testCase.name, func(t *testing.T) {
				suffix := "valid-" + string(rune('a'+index))
				recordID := uuid.New()
				current := testCase.current(recordID)
				retained := testCase.retained(recordID)
				recordID, historyRef := seedIndicatorRowRollbackWithID(t, harness.DB, incidentID, actorID, recordID, suffix, current, retained)
				httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, recordID, map[string]any{
					"base_row_version": 2,
					"client_txn_id":    "txn-i-7-06-indicator-row-" + suffix,
					"target":           map[string]any{"kind": "history_entry", "history_entry_ref": historyRef},
				}), 200)
				got := loadIndicatorRollbackSource(t, harness.DB, recordID)
				if want := testCase.expected(recordID); !reflect.DeepEqual(got, want) {
					t.Fatalf("restored Indicator source = %#v, want %#v", got, want)
				}
				if countRows(t, harness.DB, `SELECT COUNT(*) FROM records WHERE record_id = $1 AND row_version = 3`, recordID) != 1 ||
					countRows(t, harness.DB, `SELECT COUNT(*) FROM indicator_active_identities WHERE indicator_record_id = $1 AND dedupe_key = $2`, recordID, got["dedupe_key"]) != 1 ||
					countRows(t, harness.DB, `SELECT COUNT(*) FROM indicator_grid_projection WHERE record_id = $1 AND dedupe_key = $2`, recordID, got["dedupe_key"]) != 1 {
					t.Fatal("Indicator rollback did not preserve envelope, active identity, and projection consequences")
				}
			})
		}
	})
}

func indicatorChildObservationParams(incidentID uuid.UUID, sourceID uuid.UUID, fieldKey string, resolvedID *uuid.UUID, clientTxnID string) indicators.IndicatorObservationCreateParams {
	return indicators.IndicatorObservationCreateParams{
		IncidentID: incidentID, SourceRecordID: sourceID, BaseRowVersion: 1,
		SourceFieldKey: fieldKey, SpanStartByte: 0, SpanEndByte: len("record-support-source-row"),
		ResolvedIndicatorRecordID: resolvedID, ClientTxnID: clientTxnID,
		RequestID: "req-" + clientTxnID,
	}
}

func indicatorChildLifecycleParams(incidentID uuid.UUID, indicatorID uuid.UUID, baseRowVersion int64, validFrom time.Time, clientTxnID string) indicators.IndicatorLifecycleAppendParams {
	return indicators.IndicatorLifecycleAppendParams{
		IncidentID: incidentID, IndicatorRecordID: indicatorID, BaseRowVersion: baseRowVersion,
		LifecycleState: "active", ValidFrom: validFrom, SupportRefs: []uuid.UUID{},
		ClientTxnID: clientTxnID, RequestID: "req-" + clientTxnID,
	}
}

func seedIndicatorChildRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, suffix string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	value := "history_revision-" + suffix + ".example.test"
	indicatortest.SeedRecord(t, db, incidentID, actorID, recordID, "domain_name", "atomic", value)
	return recordID
}

func indicatorRollbackSource(recordID uuid.UUID, incidentID uuid.UUID, displayValue string) map[string]any {
	source := map[string]any{
		"record_id": recordID.String(), "incident_id": incidentID.String(),
		"indicator_type": "domain_name", "value_kind": "atomic",
		"display_value": displayValue, "normalized_value": displayValue,
		"defanged_value": nil, "hash_algorithm": nil, "hash_value": nil, "stix_pattern": nil,
	}
	refreshIndicatorRollbackDedupe(source)
	return source
}

func refreshIndicatorRollbackDedupe(source map[string]any) {
	source["dedupe_key"] = indicatorRollbackDedupe(
		source["indicator_type"].(string), source["value_kind"].(string), source["display_value"].(string),
		nullableIndicatorRollbackText(source["normalized_value"]), nullableIndicatorRollbackText(source["hash_algorithm"]), nullableIndicatorRollbackText(source["hash_value"]),
	)
}

func indicatorRollbackDedupe(indicatorType string, valueKind string, displayValue string, normalizedValue string, hashAlgorithm string, hashValue string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{indicatorType, valueKind, displayValue, normalizedValue, hashAlgorithm, hashValue}, "\x1f")))
	return hex.EncodeToString(sum[:])
}

func nullableIndicatorRollbackText(value any) string {
	text, _ := value.(string)
	return text
}

func seedIndicatorRowRollback(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, suffix string, current map[string]any, retained map[string]any) (uuid.UUID, string) {
	t.Helper()
	recordID := uuid.New()
	current["record_id"] = recordID.String()
	current["incident_id"] = incidentID.String()
	retained["record_id"] = recordID.String()
	retained["incident_id"] = incidentID.String()
	return seedIndicatorRowRollbackWithID(t, db, incidentID, actorID, recordID, suffix, current, retained)
}

func seedIndicatorRowRollbackWithID(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, suffix string, current map[string]any, retained map[string]any) (uuid.UUID, string) {
	t.Helper()
	indicatortest.SeedRecord(t, db, incidentID, actorID, recordID, current["indicator_type"].(string), current["value_kind"].(string), current["display_value"].(string))
	mustExec(t, db, `
UPDATE indicators
   SET indicator_type = $2, value_kind = $3, display_value = $4, normalized_value = $5,
       dedupe_key = $6, defanged_value = $7, hash_algorithm = $8, hash_value = $9, stix_pattern = $10
 WHERE record_id = $1
`, recordID, current["indicator_type"], current["value_kind"], current["display_value"], current["normalized_value"], current["dedupe_key"], current["defanged_value"], current["hash_algorithm"], current["hash_value"], current["stix_pattern"])
	advancedAt := time.Date(2026, 7, 9, 18, 0, 0, 0, time.UTC)
	advanceRecordFixtureWithAudit(t, db, recordID, 2, &advancedAt, &actorID)
	changeSetID := uuid.New()
	historyRef := "href-indicator-row-" + suffix + "-rollback"
	seedRollbackMutationWithRef(
		t, db, incidentID, actorID, recordID, changeSetID, 1, "indicator", recordID.String(), "field_update",
		indicatorRollbackSnapshot(recordID, incidentID, 1, retained), indicatorRollbackSnapshot(recordID, incidentID, 2, current), historyRef,
	)
	return recordID, historyRef
}

func indicatorRollbackSnapshot(recordID uuid.UUID, incidentID uuid.UUID, rowVersion int64, source map[string]any) map[string]any {
	return map[string]any{
		"snapshot_schema_id": "cartulary.revisions.snapshot.indicator.v1",
		"record": map[string]any{
			"record_id": recordID.String(), "incident_id": incidentID.String(), "record_type": "indicator", "row_version": rowVersion,
		},
		"source": source,
	}
}

func updateIndicatorRollbackBeforeValue(t testing.TB, db *sql.DB, historyRef string, before map[string]any) {
	t.Helper()
	mustExec(t, db, `
UPDATE change_set_mutations AS mutation
   SET before_value = $2
  FROM record_history_entry_refs AS history_ref
 WHERE history_ref.history_entry_ref = $1
   AND mutation.change_set_id = history_ref.change_set_id
   AND mutation.sequence_no = history_ref.mutation_sequence_no
`, historyRef, jsonOrNil(t, before))
}

func indicatorRollbackDurableState(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID) string {
	t.Helper()
	return stringScalar(t, db, `
SELECT jsonb_build_object(
    'record', to_jsonb(record_row),
    'source', to_jsonb(indicator_row),
    'identity', (SELECT to_jsonb(identity_row) FROM indicator_active_identities AS identity_row WHERE identity_row.indicator_record_id = $1),
    'projection', (SELECT to_jsonb(projection_row) FROM indicator_grid_projection AS projection_row WHERE projection_row.record_id = $1),
    'change_sets', (SELECT COUNT(*) FROM change_sets WHERE incident_id = $2),
    'mutations', (SELECT COUNT(*) FROM change_set_mutations AS mutation JOIN change_sets AS change_set USING (change_set_id) WHERE change_set.incident_id = $2),
    'revisions', (SELECT COUNT(*) FROM record_revisions WHERE record_id = $1),
    'idempotency', (SELECT COUNT(*) FROM route_idempotency WHERE scope_key = $1::text)
)::text
  FROM records AS record_row
  JOIN indicators AS indicator_row ON indicator_row.record_id = record_row.record_id
 WHERE record_row.record_id = $1
`, recordID, incidentID)
}

func loadIndicatorRollbackSource(t testing.TB, db *sql.DB, recordID uuid.UUID) map[string]any {
	t.Helper()
	var encoded []byte
	if err := db.QueryRowContext(context.Background(), `SELECT to_jsonb(indicator_row) FROM indicators AS indicator_row WHERE record_id = $1`, recordID).Scan(&encoded); err != nil {
		t.Fatalf("load restored Indicator source: %v", err)
	}
	var source map[string]any
	if err := json.Unmarshal(encoded, &source); err != nil {
		t.Fatalf("decode restored Indicator source: %v", err)
	}
	return source
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
