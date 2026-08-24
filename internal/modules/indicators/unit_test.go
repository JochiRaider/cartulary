package indicators_test

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

// indicator-storage / REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082 / AC-017, AC-077..AC-079.
func TestIndicatorObservationSeparation_Integration(t *testing.T) {
	harness := appsupport.StartStore(t, "entity_linking-u-4-07-indicators")
	application := newIndicatorTestApplication(t, harness.DB, revisionsupport.MustAppender(t))
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u407@example.test", "U407", "U407EntityLinkingPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-07-incident", "IR-U407", "Record relationships indicators")
	legacyRecordID := uuid.MustParse("00000000-0000-4000-8000-000000000407")
	legacyChangeSetID := uuid.MustParse("00000000-0000-4000-8000-000000000408")
	legacyHash, err := hex.DecodeString("49dd4b43356f985be78b671d6b57cfe912dcfc2782573acc9fb6c2cda8b5e6a6")
	if err != nil {
		t.Fatalf("decode deployed Indicator create hash: %v", err)
	}
	if _, err := harness.DB.Exec(context.Background(), `
INSERT INTO route_idempotency (
    route_key, scope_key, client_txn_id, actor_user_id, request_hash, status_code, response_json
) VALUES ($1, $2, $3, $4, $5, 201, $6::jsonb)
`, "indicators.rows.create", incident.ID.String()+":"+indicators.ViewSchemaID, "txn-indicator", actor.ID, legacyHash,
		`{"view_schema_id":"cartulary.view.indicators.v1","change_set_id":"00000000-0000-4000-8000-000000000408","row":{"record_id":"00000000-0000-4000-8000-000000000407","row_version":17}}`); err != nil {
		t.Fatalf("seed deployed Indicator idempotency row: %v", err)
	}
	legacyReplay, err := application.CreateIndicatorRow(context.Background(), actor.ID, incident.ID, indicators.CreateCommand{
		ClientTxnID: "txn-indicator", IndicatorType: "ipv4_addr", ValueKind: "atomic", DisplayValue: "203[.]0[.]113[.]7",
	}, "req-deployed-indicator-replay")
	if err != nil {
		t.Fatalf("replay deployed Indicator idempotency row: %v", err)
	}
	if !legacyReplay.Replayed || legacyReplay.RecordID != legacyRecordID || legacyReplay.ChangeSetID != legacyChangeSetID || legacyReplay.RowVersion != 17 {
		t.Fatalf("deployed Indicator replay = %#v", legacyReplay)
	}
	var legacyDurableRows int
	if err := harness.DB.QueryRow(context.Background(), `SELECT COUNT(*) FROM records WHERE record_id = $1`, legacyRecordID).Scan(&legacyDurableRows); err != nil || legacyDurableRows != 0 {
		t.Fatalf("deployed replay durable record count = %d, %v", legacyDurableRows, err)
	}

	create := func(clientTxnID string) indicators.CreateResult {
		t.Helper()
		example := indicatortest.PrimaryExample()
		result, err := application.CreateIndicatorRow(context.Background(), actor.ID, incident.ID, indicators.CreateCommand{
			ClientTxnID:   clientTxnID,
			IndicatorType: example.IndicatorType,
			ValueKind:     example.ValueKind,
			DisplayValue:  example.DisplayValue,
		}, "req-"+clientTxnID)
		if err != nil {
			t.Fatalf("create indicator: %v", err)
		}
		return result
	}
	first := create("txn-entity_linking-u-4-07-first")
	second := create("txn-entity_linking-u-4-07-second")
	replayed := create("txn-entity_linking-u-4-07-second")
	if first.RecordID != second.RecordID || !first.Created || first.Replayed || second.Created || second.Replayed ||
		!replayed.Replayed || replayed.Created || replayed.RecordID != second.RecordID || replayed.ChangeSetID != second.ChangeSetID {
		t.Fatalf("canonical indicator dedupe/replay failed: first=%#v second=%#v replay=%#v", first, second, replayed)
	}

	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.RecordID)
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.SiblingRecordID)
	for index, sourceRecordID := range []struct {
		id    uuid.UUID
		field string
	}{
		{id: timelinetest.RecordID, field: timelinetest.FieldSourceText},
		{id: timelinetest.SiblingRecordID, field: timelinetest.FieldSummary},
	} {
		result, err := application.CreateIndicatorObservation(context.Background(), actor.ID, manualObservationParams(
			incident.ID, sourceRecordID.id, sourceRecordID.field, &first.RecordID,
			"txn-entity-linking-observation-"+string(rune('1'+index)),
		))
		if err != nil || result.Observation.ObservationID == first.RecordID {
			t.Fatalf("create source-bound observation %d: %#v %v", index, result, err)
		}
	}
	if _, err := application.AppendIndicatorLifecycleInterval(context.Background(), actor.ID, lifecycleAppendParams(
		incident.ID, first.RecordID, 3, indicatortest.PastTime(), "txn-entity-linking-lifecycle",
	)); err != nil {
		t.Fatalf("append lifecycle interval: %v", err)
	}
	projection := lookupIndicatorProjection(t, harness.DB, first.RecordID)
	if projection.ObservationCount != 2 || projection.FirstObservedAt == nil || projection.LastObservedAt == nil || projection.LifecycleSummary == nil || *projection.LifecycleSummary != "active" {
		t.Fatalf("observations did not remain distinct in projection: %#v", projection)
	}
}
