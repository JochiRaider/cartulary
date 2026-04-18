package timeline_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/assertx"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
	"github.com/JochiRaider/cartulary/internal/testutil/timelinetest"
)

// U-4-09 / REQ-01-311, REQ-01-314..REQ-01-320, REQ-02-248, REQ-03-280 / AC-394, AC-396, AC-397.
func TestPhase4_ManualTimelineConfidenceNull_U_4_09_Authoritative(t *testing.T) {
	t.Run("add_resolved_ref persists manual host link with null confidence", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-09-add-resolved-ref")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-09-add-incident",
			"incident_key":  "IR-PHASE4-U409-A",
			"title":         "Phase 4 U-4-09 add_resolved_ref",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "WS-023.corp.example", "WS-023.corp.example")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-phase4-u-4-09-add-row",
			"timeline.summary": "Manual host relationship row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)

		resp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineHostRefs,
				1,
				"txn-phase4-u-4-09-add-patch",
				fixtures.CollectionActions(
					fixtures.AddResolvedRefAction("WS-023", golden.Phase4CanonicalHostRecordID),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("unexpected add_resolved_ref row_version: got %d want 2", got)
		}

		link := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), golden.Phase4CanonicalHostRecordID, "observed_on_host")
		assertx.RequireActiveLink(
			t,
			link,
			mustUUID(t, recordID),
			golden.Phase4CanonicalHostRecordID,
			"observed_on_host",
			golden.Phase4ManualLinkExpectation.Provenance,
			golden.Phase4ManualLinkExpectation.Confidence,
		)
		readLink := serializeActiveLinkRead(t, link)
		if confidence, ok := readLink["confidence"]; !ok || confidence != nil {
			t.Fatalf("expected serialized helper read to preserve confidence:null, got %#v", readLink)
		}
		if readLink["provenance"] != golden.Phase4ManualLinkExpectation.Provenance {
			t.Fatalf("unexpected serialized link provenance: %#v", readLink)
		}

		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		item := requireSingleCollectionItem(t, row, golden.Phase4FieldTimelineHostRefs)
		if item["item_kind"] != "resolved_ref" {
			t.Fatalf("expected resolved_ref item after add_resolved_ref, got %#v", item)
		}
		if item["resolved_record_id"] != golden.Phase4CanonicalHostRecordID.String() {
			t.Fatalf("unexpected resolved_record_id after add_resolved_ref: %#v", item)
		}
		if item["raw_text"] != "WS-023" {
			t.Fatalf("expected raw_text to remain authoritative in current-state read, got %#v", item)
		}
	})

	t.Run("resolve_item persists manual identity link with null confidence", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-09-resolve-item")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-09-resolve-incident",
			"incident_key":  "IR-PHASE4-U409-B",
			"title":         "Phase 4 U-4-09 resolve_item",
		})
		incidentID := incident["incident_id"].(string)
		seedIdentityRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalIdentityID, "Alex Analyst", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-09-resolve-row",
			golden.Phase4FieldTimelineIdentityRefs: fixtures.CollectionActions(
				fixtures.AddTokenAction("alex.analyst@example.test"),
			),
		})
		record := created["row"].(map[string]any)
		recordID := record["record_id"].(string)
		unresolvedItem := requireSingleCollectionItem(t, record, golden.Phase4FieldTimelineIdentityRefs)
		mentionID := mentionIDFromItemRef(t, unresolvedItem["item_ref"].(string))
		beforeMention := lookupMention(t, harness.DB, mentionID)

		resp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineIdentityRefs,
				1,
				"txn-phase4-u-4-09-resolve-patch",
				fixtures.CollectionActions(
					fixtures.ResolveItemAction(unresolvedItem["item_ref"].(string), golden.Phase4CanonicalIdentityID),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("unexpected resolve_item row_version: got %d want 2", got)
		}

		link := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), golden.Phase4CanonicalIdentityID, "observed_as_identity")
		assertx.RequireActiveLink(
			t,
			link,
			mustUUID(t, recordID),
			golden.Phase4CanonicalIdentityID,
			"observed_as_identity",
			golden.Phase4ManualLinkExpectation.Provenance,
			golden.Phase4ManualLinkExpectation.Confidence,
		)
		readLink := serializeActiveLinkRead(t, link)
		if confidence, ok := readLink["confidence"]; !ok || confidence != nil {
			t.Fatalf("expected serialized helper read to preserve confidence:null, got %#v", readLink)
		}

		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		item := requireSingleCollectionItem(t, row, golden.Phase4FieldTimelineIdentityRefs)
		if item["item_kind"] != "resolved_ref" {
			t.Fatalf("expected resolved_ref item after resolve_item, got %#v", item)
		}
		if item["resolved_record_id"] != golden.Phase4CanonicalIdentityID.String() {
			t.Fatalf("unexpected resolved_record_id after resolve_item: %#v", item)
		}
		if item["raw_text"] != beforeMention.RawText {
			t.Fatalf("expected resolve_item to preserve raw_text, before=%q after=%#v", beforeMention.RawText, item)
		}
	})

	t.Run("client supplied confidence is rejected without side effects", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-09-reject")
		adminLogin, _ := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-09-reject-incident",
			"incident_key":  "IR-PHASE4-U409-C",
			"title":         "Phase 4 U-4-09 rejection",
		})
		incidentID := incident["incident_id"].(string)
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-09-reject-row",
			golden.Phase4FieldTimelineHostRefs: fixtures.CollectionActions(
				fixtures.AddTokenAction("WS-023"),
			),
		})
		record := created["row"].(map[string]any)
		recordID := record["record_id"].(string)
		unresolvedItem := requireSingleCollectionItem(t, record, golden.Phase4FieldTimelineHostRefs)
		mentionID := mentionIDFromItemRef(t, unresolvedItem["item_ref"].(string))
		beforeMention := lookupMention(t, harness.DB, mentionID)
		beforeCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)
		socket := connectTimelineSocket(t, harness.Server, incidentID, adminLogin.sessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hubChanges, unsubscribe := harness.Server.Runtime.WSHub.SubscribeRecordChanges(4)
		defer unsubscribe()

		resp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineHostRefs,
				1,
				"txn-phase4-u-4-09-reject-patch",
				fixtures.CollectionActions(
					map[string]any{
						"op":                 "resolve_item",
						"item_ref":           unresolvedItem["item_ref"].(string),
						"resolved_record_id": golden.Phase4CanonicalHostRecordID.String(),
						"confidence":         80,
					},
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
		if body["error"].(map[string]any)["code"] != "invalid_mutation_payload" {
			t.Fatalf("unexpected rejection envelope: %#v", body)
		}

		afterCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)
		if afterCounters != beforeCounters {
			t.Fatalf("expected rejected payload to leave counters unchanged, before=%+v after=%+v", beforeCounters, afterCounters)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = $1
   AND scope_key = $2
   AND client_txn_id = $3
`, "timeline.records.patch", recordID, "txn-phase4-u-4-09-reject-patch"); got != 0 {
			t.Fatalf("rejected payload must not persist idempotency success state, got %d rows", got)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND dst_record_id::text = $3
   AND link_type = 'observed_on_host'
   AND deleted_at IS NULL
`, incidentID, recordID, golden.Phase4CanonicalHostRecordID.String()); got != 0 {
			t.Fatalf("rejected payload must not create a record_link, got %d", got)
		}

		afterMention := lookupMention(t, harness.DB, mentionID)
		assertx.RequireRowVersionStable(t, beforeMention.RowVersion, afterMention.RowVersion)
		assertx.RequireMentionStatus(t, afterMention, golden.Phase4MentionStatusUnresolved)
		assertx.RequireRawTextPreserved(t, beforeMention.RawText, afterMention.RawText)

		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		if got := int64(row["row_version"].(float64)); got != 1 {
			t.Fatalf("rejected payload must not advance row_version, got %d", got)
		}
		item := requireSingleCollectionItem(t, row, golden.Phase4FieldTimelineHostRefs)
		if item["item_kind"] != "unresolved_mention" {
			t.Fatalf("rejected payload must not refresh misleading resolution state, got %#v", item)
		}

		timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
		expectNoTimelineSocketMessage(t, socket)
	})
}

func lookupActiveLink(t testing.TB, db *sql.DB, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, linkType string) fixtures.LinkFixture {
	t.Helper()

	var (
		link        fixtures.LinkFixture
		confidence  sql.NullInt64
		deletedAt   sql.NullTime
		recordLink  string
		incidentRaw string
		sourceRaw   string
		targetRaw   string
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT record_link_id::text, incident_id::text, src_record_id::text, dst_record_id::text, link_type, provenance, confidence, deleted_at
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $4
   AND deleted_at IS NULL
`, incidentID, sourceID, targetID, linkType).Scan(&recordLink, &incidentRaw, &sourceRaw, &targetRaw, &link.LinkType, &link.Provenance, &confidence, &deletedAt); err != nil {
		t.Fatalf("lookup active link: %v", err)
	}

	link.RecordLinkID = mustUUID(t, recordLink)
	link.IncidentID = mustUUID(t, incidentRaw)
	link.SourceID = mustUUID(t, sourceRaw)
	link.TargetID = mustUUID(t, targetRaw)
	if confidence.Valid {
		value := int(confidence.Int64)
		link.Confidence = &value
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		link.DeletedAt = &value
	}
	return link
}

func serializeActiveLinkRead(t testing.TB, link fixtures.LinkFixture) map[string]any {
	t.Helper()

	data, err := json.Marshal(map[string]any{
		"record_link_id": link.RecordLinkID.String(),
		"provenance":     link.Provenance,
		"confidence":     link.Confidence,
	})
	if err != nil {
		t.Fatalf("marshal active link helper read: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal active link helper read: %v", err)
	}
	return payload
}

func requireSingleCollectionItem(t testing.TB, row map[string]any, fieldKey string) map[string]any {
	t.Helper()

	cells := row["cells"].(map[string]any)
	cell := cells[fieldKey].(map[string]any)
	value := cell["value"].(map[string]any)
	items := value["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected exactly one %s item, got %#v", fieldKey, value)
	}
	return items[0].(map[string]any)
}

func mentionIDFromItemRef(t testing.TB, itemRef string) uuid.UUID {
	t.Helper()

	const prefix = "entity_mention:"
	if !strings.HasPrefix(itemRef, prefix) {
		t.Fatalf("unexpected mention item_ref: %s", itemRef)
	}
	return mustUUID(t, strings.TrimPrefix(itemRef, prefix))
}

func lookupMention(t testing.TB, db *sql.DB, mentionID uuid.UUID) fixtures.EntityMentionFixture {
	t.Helper()

	var mention fixtures.EntityMentionFixture
	var (
		mentionIDRaw     string
		sourceRecordID   string
		resolvedRecordID sql.NullString
		resolvedByUserID sql.NullString
		resolvedAt       sql.NullTime
		resolutionMethod sql.NullString
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT entity_mention_id::text, source_record_id::text, raw_text, resolution_status, row_version, resolved_record_id::text, resolved_by_user_id::text, resolved_at, resolution_method
  FROM entity_mentions
 WHERE entity_mention_id = $1
`, mentionID).Scan(
		&mentionIDRaw,
		&sourceRecordID,
		&mention.RawText,
		&mention.ResolutionStatus,
		&mention.RowVersion,
		&resolvedRecordID,
		&resolvedByUserID,
		&resolvedAt,
		&resolutionMethod,
	); err != nil {
		t.Fatalf("lookup mention: %v", err)
	}

	mention.EntityMentionID = mustUUID(t, mentionIDRaw)
	mention.SourceRecordID = mustUUID(t, sourceRecordID)
	if resolvedRecordID.Valid {
		value := mustUUID(t, resolvedRecordID.String)
		mention.ResolvedRecordID = &value
	}
	if resolvedByUserID.Valid {
		value := mustUUID(t, resolvedByUserID.String)
		mention.ResolvedByUserID = &value
	}
	if resolvedAt.Valid {
		value := resolvedAt.Time.UTC()
		mention.ResolvedAt = &value
	}
	if resolutionMethod.Valid {
		value := resolutionMethod.String
		mention.ResolutionMethod = &value
	}
	return mention
}

func seedHostRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, displayName string, hostname string) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO hosts (record_id, incident_id, display_name, hostname, host_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, 'canonical', $5, $5)
`, recordID, incidentID, displayName, hostname, actorUserID); err != nil {
		t.Fatalf("seed host record: %v", err)
	}
}

func seedIdentityRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, displayName string, upn string, email string, samAccountName string) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO identities (record_id, incident_id, display_name, upn, email, sam_account_name, identity_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, 'canonical', $7, $7)
`, recordID, incidentID, displayName, upn, email, samAccountName, actorUserID); err != nil {
		t.Fatalf("seed identity record: %v", err)
	}
}
