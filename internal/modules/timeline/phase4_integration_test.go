package timeline_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/testutil/assertx"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
	"github.com/JochiRaider/cartulary/internal/testutil/timelinetest"
)

// I-4-08 / REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568, REQ-02-163..REQ-02-185, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 / AC-205, AC-388..AC-392.
func TestPhase4_AutoResolutionEligibility_I_4_08(t *testing.T) {
	t.Run("host alias exact equality auto resolves in the same patch change set", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-08-host-auto-match")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-08-host-incident",
			"incident_key":  "IR-PHASE4-U408-A",
			"title":         "Phase 4 I-4-08 host auto match",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "Gateway record", "gateway-record-01")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "host", "VPN Gateway")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-phase4-u-4-08-host-row",
			"timeline.activity_synopsis_text": "Auto-match host row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)
		beforeCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)

		resp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineHostRefs,
				1,
				"txn-phase4-u-4-08-host-patch",
				fixtures.CollectionActions(
					fixtures.AddTokenAction(" vpn   gateway "),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
		afterCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)
		assertx.RequireExactlyOneChangeSet(t, beforeCounters.ChangeSets, afterCounters.ChangeSets)
		if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("unexpected auto-match row_version: got %d want 2", got)
		}

		changeSet := timelinetest.LookupChangeSet(t, harness.DB, data["change_set_id"].(string))
		assertx.RequireActorAttribution(t, changeSet.ActorUserID, adminID, changeSet.Source, "timeline.records.patch")
		if changeSet.ClientTxnID != "txn-phase4-u-4-08-host-patch" {
			t.Fatalf("unexpected auto-match client_txn_id: %#v", changeSet)
		}

		link := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), golden.Phase4CanonicalHostRecordID, "observed_on_host")
		assertx.RequireActiveLink(
			t,
			link,
			mustUUID(t, recordID),
			golden.Phase4CanonicalHostRecordID,
			"observed_on_host",
			golden.Phase4AutoMatchLinkExpectation.Provenance,
			golden.Phase4AutoMatchLinkExpectation.Confidence,
		)
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND deleted_at IS NULL
`, incidentID, recordID); got != 1 {
			t.Fatalf("expected exactly one active relationship link after auto-match, got %d", got)
		}

		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requirePhase4ViewRowFieldSurface(t, "I-4-08", row, timeline.TimelineViewSchemaID)
		item := requireSingleCollectionItem(t, row, golden.Phase4FieldTimelineHostRefs)
		if item["item_kind"] != "resolved_ref" {
			t.Fatalf("expected resolved_ref item after auto-match, got %#v", item)
		}
		if item["raw_text"] != " vpn   gateway " {
			t.Fatalf("expected raw_text to preserve analyst token, got %#v", item)
		}
		if item["resolved_record_id"] != golden.Phase4CanonicalHostRecordID.String() {
			t.Fatalf("unexpected resolved_record_id after auto-match: %#v", item)
		}
		if item["resolution_method"] != golden.Phase4LinkProvenanceAutoMatch {
			t.Fatalf("expected auto-match resolution method marker, got %#v", item)
		}
		if item["auto_resolved"] != true {
			t.Fatalf("expected auto_resolved marker on resolved item, got %#v", item)
		}
		if item["matched_alias_text"] != "VPN Gateway" {
			t.Fatalf("expected matched_alias_text to round-trip in refreshed row, got %#v", item)
		}
		if item["provenance"] != golden.Phase4LinkProvenanceAutoMatch {
			t.Fatalf("expected auto_match provenance in refreshed row, got %#v", item)
		}
		if confidence, ok := item["confidence"].(float64); !ok || int(confidence) != 100 {
			t.Fatalf("expected confidence=100 in refreshed row, got %#v", item)
		}

		mentionID := mentionIDFromItemRef(t, item["item_ref"].(string))
		mention := lookupMention(t, harness.DB, mentionID)
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected auto-match mention to resolve to host %s, got %#v", golden.Phase4CanonicalHostRecordID, mention)
		}
		if mention.ResolvedByUserID == nil || *mention.ResolvedByUserID != mustUUID(t, adminID) {
			t.Fatalf("expected auto-match attribution to current actor, got %#v", mention)
		}
		if mention.ResolvedAt == nil {
			t.Fatalf("expected auto-match resolved_at attribution, got %#v", mention)
		}
		if mention.ResolutionMethod == nil || *mention.ResolutionMethod != golden.Phase4LinkProvenanceAutoMatch {
			t.Fatalf("expected auto-match resolution_method, got %#v", mention)
		}
	})

	t.Run("identity alias exact equality derives the identity link type", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-08-identity-auto-match")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-08-identity-incident",
			"incident_key":  "IR-PHASE4-U408-B",
			"title":         "Phase 4 I-4-08 identity auto match",
		})
		incidentID := incident["incident_id"].(string)
		seedIdentityRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalIdentityID, "Identity record", "identity-record@example.test", "identity-record@example.test", "IDENTITYREC")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalIdentityID, "identity", "Analyst Alex")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-phase4-u-4-08-identity-row",
			"timeline.activity_synopsis_text": "Auto-match identity row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)

		resp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineIdentityRefs,
				1,
				"txn-phase4-u-4-08-identity-patch",
				fixtures.CollectionActions(
					fixtures.AddTokenAction(" analyst   alex "),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
		if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("unexpected identity auto-match row_version: got %d want 2", got)
		}

		link := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), golden.Phase4CanonicalIdentityID, "observed_as_identity")
		assertx.RequireActiveLink(
			t,
			link,
			mustUUID(t, recordID),
			golden.Phase4CanonicalIdentityID,
			"observed_as_identity",
			golden.Phase4AutoMatchLinkExpectation.Provenance,
			golden.Phase4AutoMatchLinkExpectation.Confidence,
		)
		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requirePhase4ViewRowFieldSurface(t, "I-4-08", row, timeline.TimelineViewSchemaID)
		item := requireSingleCollectionItem(t, row, golden.Phase4FieldTimelineIdentityRefs)
		if item["matched_alias_text"] != "Analyst Alex" || item["provenance"] != golden.Phase4LinkProvenanceAutoMatch {
			t.Fatalf("expected identity auto-match metadata in refreshed row, got %#v", item)
		}
	})

	t.Run("create route auto resolves eligible host and identity tokens", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-08-create-auto-match")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-08-create-incident",
			"incident_key":  "IR-PHASE4-U408-CREATE",
			"title":         "Phase 4 I-4-08 create auto match",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "Create Host", "create-host")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "host", "Create Host Alias")
		seedIdentityRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalIdentityID, "Create Identity", "create.identity@example.test", "create.identity@example.test", "CREATEID")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalIdentityID, "identity", "Create Identity Alias")

		resp := doPhase3JSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":                   "txn-phase4-u-4-08-create-row",
				"timeline.activity_synopsis_text": "Create auto-match row",
				golden.Phase4FieldTimelineHostRefs: fixtures.CollectionActions(
					fixtures.AddTokenAction(" create   host alias "),
				),
				golden.Phase4FieldTimelineIdentityRefs: fixtures.CollectionActions(
					fixtures.AddTokenAction(" create   identity alias "),
				),
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusCreated)["data"].(map[string]any)
		row := data["row"].(map[string]any)
		recordID := row["record_id"].(string)
		requireMutationRecorded(t, harness.DB, data["change_set_id"].(string), recordID, adminID, "timeline.rows.create", "txn-phase4-u-4-08-create-row", 1, 1)

		hostLink := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), golden.Phase4CanonicalHostRecordID, "observed_on_host")
		assertx.RequireActiveLink(
			t,
			hostLink,
			mustUUID(t, recordID),
			golden.Phase4CanonicalHostRecordID,
			"observed_on_host",
			golden.Phase4AutoMatchLinkExpectation.Provenance,
			golden.Phase4AutoMatchLinkExpectation.Confidence,
		)
		identityLink := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), golden.Phase4CanonicalIdentityID, "observed_as_identity")
		assertx.RequireActiveLink(
			t,
			identityLink,
			mustUUID(t, recordID),
			golden.Phase4CanonicalIdentityID,
			"observed_as_identity",
			golden.Phase4AutoMatchLinkExpectation.Provenance,
			golden.Phase4AutoMatchLinkExpectation.Confidence,
		)

		hostItem := requireSingleCollectionItem(t, row, golden.Phase4FieldTimelineHostRefs)
		if hostItem["item_kind"] != "resolved_ref" || hostItem["resolved_record_id"] != golden.Phase4CanonicalHostRecordID.String() || hostItem["resolution_method"] != golden.Phase4LinkProvenanceAutoMatch || hostItem["auto_resolved"] != true {
			t.Fatalf("expected create host token to auto-resolve, got %#v", hostItem)
		}
		if hostItem["raw_text"] != " create   host alias " || hostItem["matched_alias_text"] != "Create Host Alias" {
			t.Fatalf("expected create host token to preserve raw and matched alias text, got %#v", hostItem)
		}
		identityItem := requireSingleCollectionItem(t, row, golden.Phase4FieldTimelineIdentityRefs)
		if identityItem["item_kind"] != "resolved_ref" || identityItem["resolved_record_id"] != golden.Phase4CanonicalIdentityID.String() || identityItem["resolution_method"] != golden.Phase4LinkProvenanceAutoMatch || identityItem["auto_resolved"] != true {
			t.Fatalf("expected create identity token to auto-resolve, got %#v", identityItem)
		}
		if identityItem["raw_text"] != " create   identity alias " || identityItem["matched_alias_text"] != "Create Identity Alias" {
			t.Fatalf("expected create identity token to preserve raw and matched alias text, got %#v", identityItem)
		}

		hostMention := lookupMention(t, harness.DB, mentionIDFromItemRef(t, hostItem["item_ref"].(string)))
		assertx.RequireMentionStatus(t, hostMention, golden.Phase4MentionStatusResolved)
		if hostMention.ResolvedRecordID == nil || *hostMention.ResolvedRecordID != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected create host mention to resolve to host, got %#v", hostMention)
		}
		identityMention := lookupMention(t, harness.DB, mentionIDFromItemRef(t, identityItem["item_ref"].(string)))
		assertx.RequireMentionStatus(t, identityMention, golden.Phase4MentionStatusResolved)
		if identityMention.ResolvedRecordID == nil || *identityMention.ResolvedRecordID != golden.Phase4CanonicalIdentityID {
			t.Fatalf("expected create identity mention to resolve to identity, got %#v", identityMention)
		}
	})

	t.Run("suppressor and forbidden rewrite tokens remain unresolved", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-08-unresolved")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-08-unresolved-incident",
			"incident_key":  "IR-PHASE4-U408-C",
			"title":         "Phase 4 I-4-08 unresolved eligibility",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "Host record", "host-record-23")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "host", "WS-023")

		for _, rawText := range append([]string{}, golden.Phase4AutoResolutionSuppressedTokens...) {
			t.Run(rawText, func(t *testing.T) {
				created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
					"client_txn_id":                   "txn-phase4-u-4-08-unresolved-row-" + strings.ReplaceAll(rawText, " ", "_"),
					"timeline.activity_synopsis_text": "Unresolved eligibility row",
				})
				recordID := created["row"].(map[string]any)["record_id"].(string)
				beforeCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)

				resp := doPhase3JSON(
					t,
					http.MethodPatch,
					harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
					fixtures.TimelineCollectionPatchPayload(
						golden.Phase4FieldTimelineHostRefs,
						1,
						"txn-phase4-u-4-08-unresolved-patch-"+strings.ReplaceAll(rawText, " ", "_"),
						fixtures.CollectionActions(
							fixtures.AddTokenAction(rawText),
						),
					),
					withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
					withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
				)
				data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
				afterCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)
				assertx.RequireExactlyOneChangeSet(t, beforeCounters.ChangeSets, afterCounters.ChangeSets)
				if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
					t.Fatalf("unexpected unresolved-path row_version for %q: got %d want 2", rawText, got)
				}

				row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
				requirePhase4ViewRowFieldSurface(t, "I-4-08", row, timeline.TimelineViewSchemaID)
				item := requireSingleCollectionItem(t, row, golden.Phase4FieldTimelineHostRefs)
				if item["item_kind"] != "unresolved_mention" {
					t.Fatalf("expected unresolved_mention item for %q, got %#v", rawText, item)
				}
				if item["raw_text"] != rawText {
					t.Fatalf("expected unresolved raw_text %q to remain authoritative, got %#v", rawText, item)
				}
				if _, ok := item["matched_alias_text"]; ok {
					t.Fatalf("unresolved token %q must not surface matched_alias_text, got %#v", rawText, item)
				}
				if _, ok := item["provenance"]; ok {
					t.Fatalf("unresolved token %q must not surface provenance, got %#v", rawText, item)
				}
				if _, ok := item["confidence"]; ok {
					t.Fatalf("unresolved token %q must not surface confidence, got %#v", rawText, item)
				}

				mentionID := mentionIDFromItemRef(t, item["item_ref"].(string))
				mention := lookupMention(t, harness.DB, mentionID)
				assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusUnresolved)
				if mention.ResolvedRecordID != nil || mention.ResolutionMethod != nil {
					t.Fatalf("expected unresolved mention for %q, got %#v", rawText, mention)
				}
				if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND deleted_at IS NULL
`, incidentID, recordID); got != 0 {
					t.Fatalf("expected no active relationship link for unresolved token %q, got %d", rawText, got)
				}
			})
		}
	})

	t.Run("competing alias candidates stay unresolved", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-08-competing")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-08-competing-incident",
			"incident_key":  "IR-PHASE4-U408-D",
			"title":         "Phase 4 I-4-08 competing aliases",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "Host A", "host-a")
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4DuplicateHostRecordID, "Host B", "host-b")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "host", "WS-023")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4DuplicateHostRecordID, "host", "WS-023")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-phase4-u-4-08-competing-row",
			"timeline.activity_synopsis_text": "Competing auto-match row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)

		resp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineHostRefs,
				1,
				"txn-phase4-u-4-08-competing-patch",
				fixtures.CollectionActions(
					fixtures.AddTokenAction("WS-023"),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)

		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requirePhase4ViewRowFieldSurface(t, "I-4-08", row, timeline.TimelineViewSchemaID)
		item := requireSingleCollectionItem(t, row, golden.Phase4FieldTimelineHostRefs)
		if item["item_kind"] != "unresolved_mention" {
			t.Fatalf("competing alias candidates must remain unresolved, got %#v", item)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND deleted_at IS NULL
`, incidentID, recordID); got != 0 {
			t.Fatalf("competing alias candidates must not create active links, got %d", got)
		}
	})

	t.Run("mixed eligible and ineligible tokens stay coupled to one accepted patch", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-08-mixed")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-08-mixed-incident",
			"incident_key":  "IR-PHASE4-U408-E",
			"title":         "Phase 4 I-4-08 mixed eligibility",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "Gateway record", "gateway-record-02")
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4StubHostRecordID, "Host record", "host-record-23")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "host", "VPN Gateway")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4StubHostRecordID, "host", "WS-023")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-phase4-u-4-08-mixed-row",
			"timeline.activity_synopsis_text": "Mixed auto-match row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)
		beforeCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)

		resp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineHostRefs,
				1,
				"txn-phase4-u-4-08-mixed-patch",
				fixtures.CollectionActions(
					fixtures.AddTokenAction(" vpn   gateway "),
					fixtures.AddTokenAction("WS-023?"),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
		afterCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)
		assertx.RequireExactlyOneChangeSet(t, beforeCounters.ChangeSets, afterCounters.ChangeSets)
		if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("unexpected mixed auto-match row_version: got %d want 2", got)
		}

		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requirePhase4ViewRowFieldSurface(t, "I-4-08", row, timeline.TimelineViewSchemaID)
		items := collectionItems(t, row, golden.Phase4FieldTimelineHostRefs)
		if len(items) != 2 {
			t.Fatalf("expected two host ref items after mixed patch, got %#v", items)
		}
		resolvedItem := requireCollectionItemByRawText(t, items, " vpn   gateway ")
		unresolvedItem := requireCollectionItemByRawText(t, items, "WS-023?")
		if resolvedItem["item_kind"] != "resolved_ref" || unresolvedItem["item_kind"] != "unresolved_mention" {
			t.Fatalf("expected one resolved and one unresolved item, got resolved=%#v unresolved=%#v", resolvedItem, unresolvedItem)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND deleted_at IS NULL
`, incidentID, recordID); got != 1 {
			t.Fatalf("expected mixed patch to create exactly one active link, got %d", got)
		}
	})

	t.Run("late patch rollback leaves no auto-resolution side effects", func(t *testing.T) {
		rollbackEnabled := false
		var rollbackRecordID uuid.UUID
		harness := phase4test.StartServerWithDependencies(t, "phase4-u-4-08-rollback", timeline.DependencySetForTesting(
			timeline.WithBeforeCommitHookForTesting(func(routeKey string, hookedRecordID uuid.UUID) error {
				if rollbackEnabled && routeKey == "timeline.records.patch" && hookedRecordID == rollbackRecordID {
					return errors.New("forced auto-match rollback")
				}
				return nil
			}),
		))
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-08-rollback-incident",
			"incident_key":  "IR-PHASE4-U408-F",
			"title":         "Phase 4 I-4-08 rollback",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "Gateway record", "gateway-record-03")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "host", "VPN Gateway")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-phase4-u-4-08-rollback-row",
			"timeline.activity_synopsis_text": "Rollback row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)
		rollbackRecordID = mustUUID(t, recordID)
		rollbackEnabled = true

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
				"txn-phase4-u-4-08-rollback-patch",
				fixtures.CollectionActions(
					fixtures.AddTokenAction(" vpn   gateway "),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusInternalServerError, "internal_error")

		afterCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)
		if afterCounters != beforeCounters {
			t.Fatalf("rollback must leave counters unchanged, before=%+v after=%+v", beforeCounters, afterCounters)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id::text = $2
   AND scope_key = $3
   AND client_txn_id = $4
`, "timeline.records.patch", adminID, recordID, "txn-phase4-u-4-08-rollback-patch"); got != 0 {
			t.Fatalf("rollback must not persist patch idempotency, got %d rows", got)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM entity_mentions
 WHERE source_record_id::text = $1
`, recordID); got != 0 {
			t.Fatalf("rollback must not persist auto-match mentions, got %d", got)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND deleted_at IS NULL
`, incidentID, recordID); got != 0 {
			t.Fatalf("rollback must not persist active links, got %d", got)
		}

		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requirePhase4ViewRowFieldSurface(t, "I-4-08", row, timeline.TimelineViewSchemaID)
		if got := int64(row["row_version"].(float64)); got != 1 {
			t.Fatalf("rollback must preserve source row_version, got %d", got)
		}
		timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
		expectNoTimelineSocketMessage(t, socket)
	})

	t.Run("projection rebuild never backfills auto-resolution for previously unresolved tokens", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-08-rebuild")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-08-rebuild-incident",
			"incident_key":  "IR-PHASE4-I408-G",
			"title":         "Phase 4 I-4-08 projection rebuild",
		})
		incidentID := incident["incident_id"].(string)
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-08-rebuild-row",
			golden.Phase4FieldTimelineHostRefs: fixtures.CollectionActions(
				fixtures.AddTokenAction("VPN Gateway"),
			),
		})
		record := created["row"].(map[string]any)
		recordID := record["record_id"].(string)
		rowBefore := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requirePhase4ViewRowFieldSurface(t, "I-4-08", rowBefore, timeline.TimelineViewSchemaID)
		itemBefore := requireSingleCollectionItem(t, rowBefore, golden.Phase4FieldTimelineHostRefs)
		if itemBefore["item_kind"] != "unresolved_mention" {
			t.Fatalf("expected unresolved token before later alias creation, got %#v", itemBefore)
		}

		phase4test.SeedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "Gateway node", "gateway-node.example.test", "", "")
		phase4test.SeedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "host", "VPN Gateway")

		beforeCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)
		projectionStore := projections.NewStore(harness.Server.Runtime.Postgres)
		if err := projectionStore.RebuildIncidentTimeline(context.Background(), mustUUID(t, incidentID)); err != nil {
			t.Fatalf("rebuild incident timeline: %v", err)
		}
		afterCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)
		if afterCounters != beforeCounters {
			t.Fatalf("timeline projection rebuild must not create late auto-resolution side effects, before=%+v after=%+v", beforeCounters, afterCounters)
		}

		envelope := queryTimelineEnvelope(t, harness.Server, incidentID, adminLogin, map[string]any{})
		httptestx.RequireDefaultQueryMeta(t, envelope, timeline.TimelineViewSchemaID)
		rowAfter := findRow(t, envelope["data"].(map[string]any)["rows"].([]any), recordID)
		requirePhase4ViewRowFieldSurface(t, "I-4-08", rowAfter, timeline.TimelineViewSchemaID)
		itemAfter := requireSingleCollectionItem(t, rowAfter, golden.Phase4FieldTimelineHostRefs)
		if itemAfter["item_kind"] != "unresolved_mention" {
			t.Fatalf("projection rebuild must not late-auto-resolve unresolved tokens, got %#v", itemAfter)
		}
		if itemAfter["raw_text"] != "VPN Gateway" {
			t.Fatalf("projection rebuild must preserve authoritative raw_text, got %#v", itemAfter)
		}
		if _, ok := itemAfter["matched_alias_text"]; ok {
			t.Fatalf("projection rebuild must not synthesize matched_alias_text, got %#v", itemAfter)
		}
		if _, ok := itemAfter["provenance"]; ok {
			t.Fatalf("projection rebuild must not synthesize provenance, got %#v", itemAfter)
		}
		if _, ok := itemAfter["confidence"]; ok {
			t.Fatalf("projection rebuild must not synthesize confidence, got %#v", itemAfter)
		}

		mention := lookupMention(t, harness.DB, mentionIDFromItemRef(t, itemAfter["item_ref"].(string)))
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusUnresolved)
		if mention.ResolvedRecordID != nil || mention.ResolutionMethod != nil {
			t.Fatalf("projection rebuild must leave mention unresolved, got %#v", mention)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND deleted_at IS NULL
`, incidentID, recordID); got != 0 {
			t.Fatalf("projection rebuild must not create active links, got %d", got)
		}
	})
}

// I-4-09 / REQ-01-311, REQ-01-314..REQ-01-320, REQ-02-248, REQ-03-280 / AC-394, AC-396, AC-397.
func TestPhase4_ManualTimelineConfidenceNull_I_4_09(t *testing.T) {
	t.Run("create route add_resolved_ref persists manual host link and replays cleanly", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-09-create-add-resolved-ref")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-09-create-incident",
			"incident_key":  "IR-PHASE4-I409-D",
			"title":         "Phase 4 I-4-09 create add_resolved_ref",
		})
		incidentID := incident["incident_id"].(string)
		phase4test.SeedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "WS-023.corp.example", "WS-023.corp.example", "", "")

		createPayload := map[string]any{
			"client_txn_id":                   "txn-phase4-i-4-09-create-row",
			"timeline.activity_synopsis_text": "Create-boundary manual host relationship row",
			golden.Phase4FieldTimelineHostRefs: fixtures.CollectionActions(
				fixtures.AddResolvedRefAction("WS-023", golden.Phase4CanonicalHostRecordID),
			),
		}
		createResp := doPhase3JSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			createPayload,
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createData := requireSuccessEnvelopeWithBody(t, createResp, http.StatusCreated)["data"].(map[string]any)
		recordID := createData["row"].(map[string]any)["record_id"].(string)
		requireMutationRecorded(t, harness.DB, createData["change_set_id"].(string), recordID, adminID, "timeline.rows.create", "txn-phase4-i-4-09-create-row", 1, 1)

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

		envelope := queryTimelineEnvelope(t, harness.Server, incidentID, adminLogin, map[string]any{})
		httptestx.RequireDefaultQueryMeta(t, envelope, timeline.TimelineViewSchemaID)
		row := findRow(t, envelope["data"].(map[string]any)["rows"].([]any), recordID)
		requirePhase4ViewRowFieldSurface(t, "I-4-09", row, timeline.TimelineViewSchemaID)
		item := requireSingleCollectionItem(t, row, golden.Phase4FieldTimelineHostRefs)
		if item["item_kind"] != "resolved_ref" || item["resolved_record_id"] != golden.Phase4CanonicalHostRecordID.String() {
			t.Fatalf("unexpected create-route current-state item: %#v", item)
		}
		if item["raw_text"] != "WS-023" {
			t.Fatalf("expected create-route raw_text preservation, got %#v", item)
		}
		if confidence, ok := item["confidence"]; !ok || confidence != nil {
			t.Fatalf("expected create-route current-state confidence:null, got %#v", item)
		}
		if item["provenance"] != golden.Phase4ManualLinkExpectation.Provenance {
			t.Fatalf("unexpected create-route provenance: %#v", item)
		}

		countersAfterCreate := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)
		replayResp := doPhase3JSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			createPayload,
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		replayData := requireSuccessEnvelopeWithBody(t, replayResp, http.StatusOK)["data"].(map[string]any)
		if replayData["change_set_id"] != createData["change_set_id"] {
			t.Fatalf("expected create replay to reuse original payload, got %#v %#v", createData, replayData)
		}
		httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
			FirstStatus:     http.StatusCreated,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: httptestx.ReplayCounts{
				ChangeSets:   countersAfterCreate.ChangeSets,
				MutationRows: countersAfterCreate.MutationRows,
				Revisions:    countersAfterCreate.Revisions,
			},
			StableAfter: httptestx.ReplayCounts{
				ChangeSets:   timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID).ChangeSets,
				MutationRows: timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID).MutationRows,
				Revisions:    timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID).Revisions,
			},
		})

		divergentResp := doPhase3JSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-phase4-i-4-09-create-row",
				golden.Phase4FieldTimelineHostRefs: fixtures.CollectionActions(
					fixtures.AddResolvedRefAction("WS-024", golden.Phase4CanonicalHostRecordID),
				),
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireDivergentReplayRejected(
			t,
			divergentResp.StatusCode,
			httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")["error"].(map[string]any)["code"].(string),
			"client_txn_conflict",
		)

		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'viewer',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, mustUUID(t, incidentID), mustUUID(t, adminID), mustUUID(t, adminID)); err != nil {
			t.Fatalf("demote create actor membership: %v", err)
		}
		deniedResp := doPhase3JSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-phase4-i-4-09-create-denied",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		deniedBody := httptestx.RequireErrorEnvelope(t, deniedResp, http.StatusForbidden, "authorization_denied")
		httptestx.RequireAuthorizationReDerived(
			t,
			httptestx.AuthorizationOutcome{Status: http.StatusCreated},
			httptestx.AuthorizationOutcome{Status: deniedResp.StatusCode, Code: deniedBody["error"].(map[string]any)["code"].(string)},
		)
	})

	t.Run("add_resolved_ref persists manual host link with null confidence", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-09-add-resolved-ref")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-09-add-incident",
			"incident_key":  "IR-PHASE4-U409-A",
			"title":         "Phase 4 I-4-09 add_resolved_ref",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4CanonicalHostRecordID, "WS-023.corp.example", "WS-023.corp.example")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-phase4-u-4-09-add-row",
			"timeline.activity_synopsis_text": "Manual host relationship row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)
		beforeCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)

		resp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineHostRefs,
				1,
				"txn-phase4-i-4-09-add-patch",
				fixtures.CollectionActions(
					fixtures.AddResolvedRefAction("WS-023", golden.Phase4CanonicalHostRecordID),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
		afterCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)
		assertx.RequireExactlyOneChangeSet(t, beforeCounters.ChangeSets, afterCounters.ChangeSets)
		if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("unexpected add_resolved_ref row_version: got %d want 2", got)
		}
		requireMutationRecorded(t, harness.DB, data["change_set_id"].(string), recordID, adminID, "timeline.records.patch", "txn-phase4-i-4-09-add-patch", 1, 2)

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
		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requirePhase4ViewRowFieldSurface(t, "I-4-09", row, timeline.TimelineViewSchemaID)
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
		if confidence, ok := item["confidence"]; !ok || confidence != nil {
			t.Fatalf("expected current-state read to preserve confidence:null, got %#v", item)
		}
		if item["provenance"] != golden.Phase4ManualLinkExpectation.Provenance {
			t.Fatalf("unexpected current-state link provenance: %#v", item)
		}

		replayResp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineHostRefs,
				1,
				"txn-phase4-i-4-09-add-patch",
				fixtures.CollectionActions(
					fixtures.AddResolvedRefAction("WS-023", golden.Phase4CanonicalHostRecordID),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		replayData := requireSuccessEnvelopeWithBody(t, replayResp, http.StatusOK)["data"].(map[string]any)
		if replayData["change_set_id"] != data["change_set_id"] {
			t.Fatalf("expected add_resolved_ref replay to reuse original payload, got %#v %#v", data, replayData)
		}
		httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: httptestx.ReplayCounts{
				ChangeSets:   afterCounters.ChangeSets,
				MutationRows: afterCounters.MutationRows,
				Revisions:    afterCounters.Revisions,
			},
			StableAfter: httptestx.ReplayCounts{
				ChangeSets:   timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID).ChangeSets,
				MutationRows: timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID).MutationRows,
				Revisions:    timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID).Revisions,
			},
		})

		divergentResp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineHostRefs,
				1,
				"txn-phase4-i-4-09-add-patch",
				fixtures.CollectionActions(
					fixtures.AddResolvedRefAction("WS-024", golden.Phase4CanonicalHostRecordID),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireDivergentReplayRejected(
			t,
			divergentResp.StatusCode,
			httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")["error"].(map[string]any)["code"].(string),
			"client_txn_conflict",
		)

		phase4test.SeedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), golden.Phase4DuplicateHostRecordID, "WS-024.corp.example", "WS-024.corp.example", "", "")
		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'viewer',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, mustUUID(t, incidentID), mustUUID(t, adminID), mustUUID(t, adminID)); err != nil {
			t.Fatalf("demote patch actor membership: %v", err)
		}
		deniedResp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineHostRefs,
				2,
				"txn-phase4-i-4-09-add-denied",
				fixtures.CollectionActions(
					fixtures.AddResolvedRefAction("WS-024", golden.Phase4DuplicateHostRecordID),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		deniedBody := httptestx.RequireErrorEnvelope(t, deniedResp, http.StatusForbidden, "authorization_denied")
		httptestx.RequireAuthorizationReDerived(
			t,
			httptestx.AuthorizationOutcome{Status: http.StatusOK},
			httptestx.AuthorizationOutcome{Status: deniedResp.StatusCode, Code: deniedBody["error"].(map[string]any)["code"].(string)},
		)
	})

	t.Run("resolve_item persists manual identity link with null confidence", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-09-resolve-item")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-09-resolve-incident",
			"incident_key":  "IR-PHASE4-U409-B",
			"title":         "Phase 4 I-4-09 resolve_item",
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
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
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
		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requirePhase4ViewRowFieldSurface(t, "I-4-09", row, timeline.TimelineViewSchemaID)
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
		if confidence, ok := item["confidence"]; !ok || confidence != nil {
			t.Fatalf("expected current-state read to preserve confidence:null, got %#v", item)
		}
		if item["provenance"] != golden.Phase4ManualLinkExpectation.Provenance {
			t.Fatalf("unexpected current-state link provenance: %#v", item)
		}
	})

	t.Run("resolve_item without a target is rejected without creating an identity", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-09-resolve-create")
		adminLogin, _ := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-09-resolve-create-incident",
			"incident_key":  "IR-PHASE4-U409-D",
			"title":         "Phase 4 I-4-09 resolve_item create",
		})
		incidentID := incident["incident_id"].(string)
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-phase4-u-4-09-resolve-create-row",
			"timeline.activity_synopsis_text": "Manual identity create-from-mention row",
			golden.Phase4FieldTimelineIdentityRefs: fixtures.CollectionActions(
				fixtures.AddTokenAction("vpn.user@example.test"),
			),
		})
		record := created["row"].(map[string]any)
		recordID := record["record_id"].(string)
		unresolvedItem := requireSingleCollectionItem(t, record, golden.Phase4FieldTimelineIdentityRefs)
		mentionID := mentionIDFromItemRef(t, unresolvedItem["item_ref"].(string))
		beforeCounters := timelinetest.SnapshotCounters(t, harness.DB, incidentID, recordID)

		resp := doPhase3JSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			fixtures.TimelineCollectionPatchPayload(
				golden.Phase4FieldTimelineIdentityRefs,
				1,
				"txn-phase4-u-4-09-resolve-create-patch",
				fixtures.CollectionActions(
					map[string]any{
						"op":       "resolve_item",
						"item_ref": unresolvedItem["item_ref"].(string),
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
			t.Fatalf("expected rejected resolve_item to leave counters unchanged, before=%+v after=%+v", beforeCounters, afterCounters)
		}
		mention := lookupMention(t, harness.DB, mentionID)
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusUnresolved)
		if mention.ResolvedRecordID != nil {
			t.Fatalf("expected mention to remain unresolved, got %#v", mention)
		}
		if got := queryCount(t, harness.DB, `SELECT COUNT(*) FROM identities WHERE incident_id::text = $1`, incidentID); got != 0 {
			t.Fatalf("expected missing-target resolve to create no identity rows, got %d", got)
		}
	})

	t.Run("client supplied confidence is rejected without side effects", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-u-4-09-reject")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-u-4-09-reject-incident",
			"incident_key":  "IR-PHASE4-U409-C",
			"title":         "Phase 4 I-4-09 rejection",
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
   AND actor_user_id::text = $2
   AND scope_key = $3
   AND client_txn_id = $4
`, "timeline.records.patch", adminID, recordID, "txn-phase4-u-4-09-reject-patch"); got != 0 {
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
		requirePhase4ViewRowFieldSurface(t, "I-4-09", row, timeline.TimelineViewSchemaID)
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

func requireSuccessEnvelopeWithBody(t testing.TB, resp *http.Response, wantStatus int) map[string]any {
	t.Helper()

	if resp.StatusCode != wantStatus {
		body := httptestx.ReadJSONBody(t, resp)
		t.Fatalf("unexpected status: got %d want %d body=%#v", resp.StatusCode, wantStatus, body)
	}
	return httptestx.RequireSuccessEnvelope(t, resp, wantStatus)
}

func requirePhase4ViewRowFieldSurface(t testing.TB, testID string, row map[string]any, viewSchemaID string) {
	t.Helper()

	httptestx.RequireFieldKeyConformance(
		t,
		phase4test.SortedRowFieldKeys(t, row),
		phase4test.AllowedFieldKeys(t, testID, viewSchemaID),
	)
}

func requireSingleCollectionItem(t testing.TB, row map[string]any, fieldKey string) map[string]any {
	t.Helper()

	items := collectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected exactly one %s item, got %#v", fieldKey, items)
	}
	return items[0]
}

func collectionItems(t testing.TB, row map[string]any, fieldKey string) []map[string]any {
	t.Helper()

	cells := row["cells"].(map[string]any)
	cell := cells[fieldKey].(map[string]any)
	value := cell["value"].(map[string]any)
	rawItems := value["items"].([]any)
	items := make([]map[string]any, 0, len(rawItems))
	for _, rawItem := range rawItems {
		items = append(items, rawItem.(map[string]any))
	}
	return items
}

func requireCollectionItemByRawText(t testing.TB, items []map[string]any, rawText string) map[string]any {
	t.Helper()

	for _, item := range items {
		if item["raw_text"] == rawText {
			return item
		}
	}
	t.Fatalf("expected collection item with raw_text=%q, got %#v", rawText, items)
	return nil
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
	phase4test.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "host")

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO hosts (record_id, incident_id, display_name, hostname, host_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, 'canonical', $5, $5)
`, recordID, incidentID, displayName, hostname, actorUserID); err != nil {
		t.Fatalf("seed host record: %v", err)
	}
}

func seedIdentityRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, displayName string, upn string, email string, samAccountName string) {
	t.Helper()
	phase4test.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "identity")

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO identities (record_id, incident_id, display_name, upn, email, sam_account_name, identity_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, 'canonical', $7, $7)
`, recordID, incidentID, displayName, upn, email, samAccountName, actorUserID); err != nil {
		t.Fatalf("seed identity record: %v", err)
	}
}

func seedEntityAlias(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, entityType string, rawText string) {
	t.Helper()

	normalized, ok := fieldnorm.NormalizeLine(rawText)
	if !ok {
		t.Fatalf("normalize entity alias %q", rawText)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO entity_aliases (incident_id, record_id, entity_type, raw_text, normalized_text, classification, created_by_user_id, created_at)
VALUES ($1, $2, $3, $4, $5, 'suggestion_only', $6, now())
`, incidentID, recordID, entityType, rawText, normalized, actorUserID); err != nil {
		t.Fatalf("seed entity alias: %v", err)
	}
}
