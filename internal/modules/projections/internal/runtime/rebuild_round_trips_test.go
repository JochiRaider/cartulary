package runtime_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	projectiontestsupport "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
)

const (
	roundTripTimelineView = "cartulary.view.timeline.v2"
	roundTripPartyView    = "cartulary.view.parties.v1"
)

// TestIncidentRebuildRoundTripsAreBoundedByPages: an incident rebuild runs
// inside the import transaction (Core 01 REQ-01-641), so its round trips are
// set by provider pages, never by records. Every rebuilt row equals the row
// the single-record refresh writes for it.
func TestIncidentRebuildRoundTripsAreBoundedByPages(t *testing.T) {
	ctx := t.Context()
	harness := appsupport.StartStore(t, "projection-rebuild-round-trips")
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "rebuild-round-trips@example.test", DisplayName: "Rebuild Round Trips"})
	runtime := projectiontestsupport.MustBuild(t, harness.DB)
	rebuild := func(records int) int {
		t.Helper()
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, fmt.Sprintf("txn-rebuild-round-trips-%d", records), fmt.Sprintf("IR-ROUND-TRIPS-%d", records), "Rebuild round trips")
		seeded := seedRoundTripIncident(t, harness.DB, incident.ID, actor.ID, records)
		tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		counter := dbassert.CountRoundTrips(tx)
		if err := runtime.ImportRebuilder().RebuildImportedIncidentTx(ctx, counter, incident.ID); err != nil {
			t.Fatalf("rebuild %d-record incident: %v", records, err)
		}
		trips := counter.RoundTrips()
		for view, recordIDs := range seeded {
			for _, recordID := range recordIDs {
				rebuilt, err := runtime.RevisionLiveRecords().LoadRowTx(ctx, tx, view, recordID)
				if err != nil {
					t.Fatalf("load rebuilt %s row %s: %v", view, recordID, err)
				}
				if err := runtime.RevisionRows().RefreshRowTx(ctx, tx, view, recordID); err != nil {
					t.Fatalf("refresh %s row %s: %v", view, recordID, err)
				}
				refreshed, err := runtime.RevisionLiveRecords().LoadRowTx(ctx, tx, view, recordID)
				if err != nil {
					t.Fatalf("load refreshed %s row %s: %v", view, recordID, err)
				}
				if !reflect.DeepEqual(rebuilt, refreshed) {
					t.Fatalf("%s row %s: rebuilt %#v, refreshed %#v", view, recordID, rebuilt, refreshed)
				}
			}
		}
		return trips
	}
	few, many := rebuild(2), rebuild(12)
	if few != many {
		t.Fatalf("incident rebuild round trips grow with records: %d for 2 records of each kind, %d for 12", few, many)
	}
}

// seedRoundTripIncident seeds records of each kind with every Timeline
// collection fact: an unresolved mention, a tag and attached Evidence.
func seedRoundTripIncident(t testing.TB, db dbport.DB, incidentID, actorID uuid.UUID, records int) map[string][]uuid.UUID {
	t.Helper()
	ctx := t.Context()
	seeded := map[string][]uuid.UUID{}
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	for index := range records {
		timelineID, evidenceID, partyID := uuid.New(), uuid.New(), uuid.New()
		timelinetest.SeedTimelineRecordWithSourceText(t, db, incidentID, actorID, timelineID, fmt.Sprintf("round trip %d", index))
		envelopetest.SeedRecordEnvelope(t, db, incidentID, actorID, evidenceID, "evidence")
		exec(`INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state) VALUES ($1, $2, $3, 'requested', 'pending')`,
			evidenceID, incidentID, fmt.Sprintf("Evidence %d", index))
		exec(`
INSERT INTO entity_mentions (
    source_record_id, entity_type, source_field_key, origin_kind, origin_locator,
    raw_text, normalized_text, resolution_status, ordinal, created_by_user_id
) VALUES ($1, 'host', 'timeline.host_refs', 'manual_entry', 'fixture', $2, $2, 'unresolved', 1, $3)
`, timelineID, fmt.Sprintf("host-%d", index), actorID)
		exec(`
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, provenance,
    owner_user_id, created_by_user_id, decided_at
) VALUES ($1, $2, $3, 'attached_evidence', 'manual', $4, $4, now())
`, incidentID, timelineID, evidenceID, actorID)
		exec(`
INSERT INTO record_tags (incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, 'Priority', 'priority', $3)
`, incidentID, timelineID, actorID)
		envelopetest.SeedRecordEnvelope(t, db, incidentID, actorID, partyID, "party")
		exec(`INSERT INTO parties (record_id, incident_id, display_name, party_kind) VALUES ($1, $2, $3, 'team')`,
			partyID, incidentID, fmt.Sprintf("Party %d", index))
		seeded[roundTripTimelineView] = append(seeded[roundTripTimelineView], timelineID)
		seeded[roundTripPartyView] = append(seeded[roundTripPartyView], partyID)
	}
	return seeded
}
