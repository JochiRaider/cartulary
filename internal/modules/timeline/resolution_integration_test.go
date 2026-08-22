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

	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineadmission "github.com/JochiRaider/cartulary/internal/modules/timeline/admission"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	revisiontest "github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// timeline-resolution / REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568, REQ-02-163..REQ-02-185, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 / AC-205, AC-388..AC-392.
func TestAutoResolutionEligibility_Integration(t *testing.T) {
	t.Run("host alias exact equality auto resolves in the same patch change set", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-u-4-08-host-auto-match")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-08-host-incident",
			"incident_key":  "IR-ENTITY-LINKING-U408-A",
			"title":         "Record relationships timeline-resolution host auto match",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "Gateway record", "gateway-record-01")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "host", "VPN Gateway")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-entity_linking-u-4-08-host-row",
			"timeline.activity_synopsis_text": "Auto-match host row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)
		beforeCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)

		resp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				1,
				"txn-entity_linking-u-4-08-host-patch",
				timelinetest.CollectionActions(
					timelinetest.AddTokenAction(" vpn   gateway "),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
		afterCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
		revisiontest.RequireExactlyOneChangeSet(t, beforeCounters.ChangeSets, afterCounters.ChangeSets)
		if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("unexpected auto-match row_version: got %d want 2", got)
		}

		changeSet := asserttest.LookupChangeSet(t, asserttest.SQLDatabase(harness.DB), data["change_set_id"].(string))
		revisiontest.RequireActorAttribution(t, changeSet.ActorUserID, adminID, changeSet.Source, "timeline.records.patch")
		if changeSet.ClientTxnID != "txn-entity_linking-u-4-08-host-patch" {
			t.Fatalf("unexpected auto-match client_txn_id: %#v", changeSet)
		}

		link := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), entitytest.CanonicalHostRecordID, "observed_on_host")
		linktest.RequireActiveLink(
			t,
			link,
			mustUUID(t, recordID),
			entitytest.CanonicalHostRecordID,
			"observed_on_host",
			linktest.LinkProvenanceAutoMatch,
			intPointer(100),
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
		requireViewRowFieldSurface(t, "timeline-resolution", row, timeline.TimelineViewSchemaID)
		item := requireSingleCollectionItem(t, row, timelinetest.FieldHostRefs)
		if item["item_kind"] != "resolved_ref" {
			t.Fatalf("expected resolved_ref item after auto-match, got %#v", item)
		}
		if item["raw_text"] != " vpn   gateway " {
			t.Fatalf("expected raw_text to preserve analyst token, got %#v", item)
		}
		if item["resolved_record_id"] != entitytest.CanonicalHostRecordID.String() {
			t.Fatalf("unexpected resolved_record_id after auto-match: %#v", item)
		}
		if item["resolution_method"] != linktest.LinkProvenanceAutoMatch {
			t.Fatalf("expected auto-match resolution method marker, got %#v", item)
		}
		if item["auto_resolved"] != true {
			t.Fatalf("expected auto_resolved marker on resolved item, got %#v", item)
		}
		if item["matched_alias_text"] != "VPN Gateway" {
			t.Fatalf("expected matched_alias_text to round-trip in refreshed row, got %#v", item)
		}
		if item["provenance"] != linktest.LinkProvenanceAutoMatch {
			t.Fatalf("expected auto_match provenance in refreshed row, got %#v", item)
		}
		if confidence, ok := item["confidence"].(float64); !ok || int(confidence) != 100 {
			t.Fatalf("expected confidence=100 in refreshed row, got %#v", item)
		}

		mentionID := mentionIDFromItemRef(t, item["item_ref"].(string))
		mention := lookupMention(t, harness.DB, mentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected auto-match mention to resolve to host %s, got %#v", entitytest.CanonicalHostRecordID, mention)
		}
		if mention.ResolvedByUserID == nil || *mention.ResolvedByUserID != mustUUID(t, adminID) {
			t.Fatalf("expected auto-match attribution to current actor, got %#v", mention)
		}
		if mention.ResolvedAt == nil {
			t.Fatalf("expected auto-match resolved_at attribution, got %#v", mention)
		}
		if mention.ResolutionMethod == nil || *mention.ResolutionMethod != linktest.LinkProvenanceAutoMatch {
			t.Fatalf("expected auto-match resolution_method, got %#v", mention)
		}
	})

	t.Run("identity alias exact equality derives the identity link type", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-u-4-08-identity-auto-match")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-08-identity-incident",
			"incident_key":  "IR-ENTITY-LINKING-U408-B",
			"title":         "Record relationships timeline-resolution identity auto match",
		})
		incidentID := incident["incident_id"].(string)
		seedIdentityRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalIdentityRecordID, "Identity record", "identity-record@example.test", "identity-record@example.test", "IDENTITYREC")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalIdentityRecordID, "identity", "Analyst Alex")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-entity_linking-u-4-08-identity-row",
			"timeline.activity_synopsis_text": "Auto-match identity row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)

		resp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldIdentityRefs,
				1,
				"txn-entity_linking-u-4-08-identity-patch",
				timelinetest.CollectionActions(
					timelinetest.AddTokenAction(" analyst   alex "),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
		if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("unexpected identity auto-match row_version: got %d want 2", got)
		}

		link := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), entitytest.CanonicalIdentityRecordID, "observed_as_identity")
		linktest.RequireActiveLink(
			t,
			link,
			mustUUID(t, recordID),
			entitytest.CanonicalIdentityRecordID,
			"observed_as_identity",
			linktest.LinkProvenanceAutoMatch,
			intPointer(100),
		)
		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requireViewRowFieldSurface(t, "timeline-resolution", row, timeline.TimelineViewSchemaID)
		item := requireSingleCollectionItem(t, row, timelinetest.FieldIdentityRefs)
		if item["matched_alias_text"] != "Analyst Alex" || item["provenance"] != linktest.LinkProvenanceAutoMatch {
			t.Fatalf("expected identity auto-match metadata in refreshed row, got %#v", item)
		}
	})

	t.Run("create route auto resolves eligible host and identity tokens", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-u-4-08-create-auto-match")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-08-create-incident",
			"incident_key":  "IR-ENTITY-LINKING-U408-CREATE",
			"title":         "Record relationships timeline-resolution create auto match",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "Create Host", "create-host")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "host", "Create Host Alias")
		seedIdentityRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalIdentityRecordID, "Create Identity", "create.identity@example.test", "create.identity@example.test", "CREATEID")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalIdentityRecordID, "identity", "Create Identity Alias")

		resp := doJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":                   "txn-entity_linking-u-4-08-create-row",
				"timeline.activity_synopsis_text": "Create auto-match row",
				timelinetest.FieldHostRefs: timelinetest.CollectionActions(
					timelinetest.AddTokenAction(" create   host alias "),
				),
				timelinetest.FieldIdentityRefs: timelinetest.CollectionActions(
					timelinetest.AddTokenAction(" create   identity alias "),
				),
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusCreated)["data"].(map[string]any)
		row := data["row"].(map[string]any)
		recordID := row["record_id"].(string)
		requireMutationRecorded(t, harness.DB, data["change_set_id"].(string), recordID, adminID, "timeline.rows.create", "txn-entity_linking-u-4-08-create-row", 3, 1)

		hostLink := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), entitytest.CanonicalHostRecordID, "observed_on_host")
		linktest.RequireActiveLink(
			t,
			hostLink,
			mustUUID(t, recordID),
			entitytest.CanonicalHostRecordID,
			"observed_on_host",
			linktest.LinkProvenanceAutoMatch,
			intPointer(100),
		)
		identityLink := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), entitytest.CanonicalIdentityRecordID, "observed_as_identity")
		linktest.RequireActiveLink(
			t,
			identityLink,
			mustUUID(t, recordID),
			entitytest.CanonicalIdentityRecordID,
			"observed_as_identity",
			linktest.LinkProvenanceAutoMatch,
			intPointer(100),
		)

		hostItem := requireSingleCollectionItem(t, row, timelinetest.FieldHostRefs)
		if hostItem["item_kind"] != "resolved_ref" || hostItem["resolved_record_id"] != entitytest.CanonicalHostRecordID.String() || hostItem["resolution_method"] != linktest.LinkProvenanceAutoMatch || hostItem["auto_resolved"] != true {
			t.Fatalf("expected create host token to auto-resolve, got %#v", hostItem)
		}
		if hostItem["raw_text"] != " create   host alias " || hostItem["matched_alias_text"] != "Create Host Alias" {
			t.Fatalf("expected create host token to preserve raw and matched alias text, got %#v", hostItem)
		}
		identityItem := requireSingleCollectionItem(t, row, timelinetest.FieldIdentityRefs)
		if identityItem["item_kind"] != "resolved_ref" || identityItem["resolved_record_id"] != entitytest.CanonicalIdentityRecordID.String() || identityItem["resolution_method"] != linktest.LinkProvenanceAutoMatch || identityItem["auto_resolved"] != true {
			t.Fatalf("expected create identity token to auto-resolve, got %#v", identityItem)
		}
		if identityItem["raw_text"] != " create   identity alias " || identityItem["matched_alias_text"] != "Create Identity Alias" {
			t.Fatalf("expected create identity token to preserve raw and matched alias text, got %#v", identityItem)
		}

		hostMention := lookupMention(t, harness.DB, mentionIDFromItemRef(t, hostItem["item_ref"].(string)))
		entitytest.RequireMentionStatus(t, hostMention, entitytest.MentionStatusResolved)
		if hostMention.ResolvedRecordID == nil || *hostMention.ResolvedRecordID != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected create host mention to resolve to host, got %#v", hostMention)
		}
		identityMention := lookupMention(t, harness.DB, mentionIDFromItemRef(t, identityItem["item_ref"].(string)))
		entitytest.RequireMentionStatus(t, identityMention, entitytest.MentionStatusResolved)
		if identityMention.ResolvedRecordID == nil || *identityMention.ResolvedRecordID != entitytest.CanonicalIdentityRecordID {
			t.Fatalf("expected create identity mention to resolve to identity, got %#v", identityMention)
		}
	})

	t.Run("suppressor and forbidden rewrite tokens remain unresolved", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-u-4-08-unresolved")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-08-unresolved-incident",
			"incident_key":  "IR-ENTITY-LINKING-U408-C",
			"title":         "Record relationships timeline-resolution unresolved eligibility",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "Host record", "host-record-23")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "host", "WS-023")

		for _, rawText := range append([]string{}, entitytest.AutoResolutionSuppressedTokens...) {
			t.Run(rawText, func(t *testing.T) {
				created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
					"client_txn_id":                   "txn-entity_linking-u-4-08-unresolved-row-" + strings.ReplaceAll(rawText, " ", "_"),
					"timeline.activity_synopsis_text": "Unresolved eligibility row",
				})
				recordID := created["row"].(map[string]any)["record_id"].(string)
				beforeCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)

				resp := doJSON(
					t,
					http.MethodPatch,
					harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
					timelinetest.TimelineCollectionPatchPayload(
						timelinetest.FieldHostRefs,
						1,
						"txn-entity_linking-u-4-08-unresolved-patch-"+strings.ReplaceAll(rawText, " ", "_"),
						timelinetest.CollectionActions(
							timelinetest.AddTokenAction(rawText),
						),
					),
					withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
					withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
				)
				data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
				afterCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
				revisiontest.RequireExactlyOneChangeSet(t, beforeCounters.ChangeSets, afterCounters.ChangeSets)
				if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
					t.Fatalf("unexpected unresolved-path row_version for %q: got %d want 2", rawText, got)
				}

				row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
				requireViewRowFieldSurface(t, "timeline-resolution", row, timeline.TimelineViewSchemaID)
				item := requireSingleCollectionItem(t, row, timelinetest.FieldHostRefs)
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
				entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusUnresolved)
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
		harness := appsupport.StartServer(t, "entity_linking-u-4-08-competing")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-08-competing-incident",
			"incident_key":  "IR-ENTITY-LINKING-U408-D",
			"title":         "Record relationships timeline-resolution competing aliases",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "Host A", "host-a")
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.DuplicateHostRecordID, "Host B", "host-b")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "host", "WS-023")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.DuplicateHostRecordID, "host", "WS-023")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-entity_linking-u-4-08-competing-row",
			"timeline.activity_synopsis_text": "Competing auto-match row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)

		resp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				1,
				"txn-entity_linking-u-4-08-competing-patch",
				timelinetest.CollectionActions(
					timelinetest.AddTokenAction("WS-023"),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)

		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requireViewRowFieldSurface(t, "timeline-resolution", row, timeline.TimelineViewSchemaID)
		item := requireSingleCollectionItem(t, row, timelinetest.FieldHostRefs)
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
		harness := appsupport.StartServer(t, "entity_linking-u-4-08-mixed")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-08-mixed-incident",
			"incident_key":  "IR-ENTITY-LINKING-U408-E",
			"title":         "Record relationships timeline-resolution mixed eligibility",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "Gateway record", "gateway-record-02")
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.StubHostRecordID, "Host record", "host-record-23")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "host", "VPN Gateway")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.StubHostRecordID, "host", "WS-023")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-entity_linking-u-4-08-mixed-row",
			"timeline.activity_synopsis_text": "Mixed auto-match row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)
		beforeCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)

		resp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				1,
				"txn-entity_linking-u-4-08-mixed-patch",
				timelinetest.CollectionActions(
					timelinetest.AddTokenAction(" vpn   gateway "),
					timelinetest.AddTokenAction("WS-023?"),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
		afterCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
		revisiontest.RequireExactlyOneChangeSet(t, beforeCounters.ChangeSets, afterCounters.ChangeSets)
		if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("unexpected mixed auto-match row_version: got %d want 2", got)
		}

		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requireViewRowFieldSurface(t, "timeline-resolution", row, timeline.TimelineViewSchemaID)
		items := collectionItems(t, row, timelinetest.FieldHostRefs)
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
		harness := appsupport.StartServer(t, "entity_linking-u-4-08-rollback")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-08-rollback-incident",
			"incident_key":  "IR-ENTITY-LINKING-U408-F",
			"title":         "Record relationships timeline-resolution rollback",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "Gateway record", "gateway-record-03")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "host", "VPN Gateway")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-entity_linking-u-4-08-rollback-row",
			"timeline.activity_synopsis_text": "Rollback row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)
		rollbackRecordID = mustUUID(t, recordID)
		rollbackEnabled = true

		beforeCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
		asserttest.AwaitIncidentStreamIdle(t, asserttest.SQLDatabase(harness.DB), incidentID)
		socket := connectTimelineSocket(t, harness.Server, incidentID, adminLogin.sessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hubChanges, unsubscribe := harness.Collaboration.SubscribeIncident(mustUUID(t, incidentID), 4)
		defer unsubscribe()

		facade := timelineFacadeWithProjectionFailure(t, harness, func(mutation workbookprojection.ProjectionMutation) error {
			if rollbackEnabled && mutation.RecordID == rollbackRecordID {
				return errors.New("forced auto-match rollback")
			}
			return nil
		})
		normalized, ok := fieldnorm.NormalizeMentionToken(" vpn   gateway ")
		if !ok {
			t.Fatal("normalize rollback mention token")
		}
		request := timeline.PatchRequest{
			ViewSchemaID:   timeline.TimelineViewSchemaID,
			BaseRowVersion: 1,
			ClientTxnID:    "txn-entity_linking-u-4-08-rollback-patch",
			CanonicalChange: []timeline.PatchChange{{
				FieldKey: timelinetest.FieldHostRefs,
				ActionPayload: &timeline.CollectionActionPayload{Actions: []timeline.CollectionAction{{
					Op:             "add_token",
					RawText:        " vpn   gateway ",
					NormalizedText: normalized,
				}}},
			}},
		}
		_, err := facade.PatchRow(context.Background(), timeline.PatchRowCommand{
			Actor:       loadTimelineTestUser(t, harness, adminID),
			RecordID:    rollbackRecordID,
			Request:     request,
			RequestHash: timelineadmission.PatchRequestHash(request),
			RequestID:   "req-entity_linking-u-4-08-rollback-patch",
			Now:         time.Now().UTC(),
		})
		if err == nil || !strings.Contains(err.Error(), "forced auto-match rollback") {
			t.Fatalf("expected forced projection rollback, got %v", err)
		}

		afterCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
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
`, "timeline.records.patch", adminID, recordID, "txn-entity_linking-u-4-08-rollback-patch"); got != 0 {
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
		requireViewRowFieldSurface(t, "timeline-resolution", row, timeline.TimelineViewSchemaID)
		if got := int64(row["row_version"].(float64)); got != 1 {
			t.Fatalf("rollback must preserve source row_version, got %d", got)
		}
		asserttest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
		expectNoTimelineSocketMessage(t, socket)
	})

	t.Run("projection rebuild never backfills auto-resolution for previously unresolved tokens", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-08-rebuild")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-08-rebuild-incident",
			"incident_key":  "IR-ENTITY-LINKING-I408-G",
			"title":         "Record relationships timeline-resolution projection rebuild",
		})
		incidentID := incident["incident_id"].(string)
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-08-rebuild-row",
			timelinetest.FieldHostRefs: timelinetest.CollectionActions(
				timelinetest.AddTokenAction("VPN Gateway"),
			),
		})
		record := created["row"].(map[string]any)
		recordID := record["record_id"].(string)
		rowBefore := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requireViewRowFieldSurface(t, "timeline-resolution", rowBefore, timeline.TimelineViewSchemaID)
		itemBefore := requireSingleCollectionItem(t, rowBefore, timelinetest.FieldHostRefs)
		if itemBefore["item_kind"] != "unresolved_mention" {
			t.Fatalf("expected unresolved token before later alias creation, got %#v", itemBefore)
		}

		entitytest.SeedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "Gateway node", "gateway-node.example.test", "", "")
		entitytest.SeedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "host", "VPN Gateway")

		beforeCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
		if err := harness.Projections.RebuildTimeline(context.Background(), mustUUID(t, incidentID)); err != nil {
			t.Fatalf("rebuild incident timeline: %v", err)
		}
		afterCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
		if afterCounters != beforeCounters {
			t.Fatalf("timeline projection rebuild must not create late auto-resolution side effects, before=%+v after=%+v", beforeCounters, afterCounters)
		}

		envelope := queryTimelineEnvelope(t, harness.Server, incidentID, adminLogin)
		contractassert.RequireDefaultQueryMeta(t, envelope, timeline.TimelineViewSchemaID)
		rawRows := envelope["data"].(map[string]any)["rows"].([]any)
		rows := make([]map[string]any, 0, len(rawRows))
		for _, rawRow := range rawRows {
			rows = append(rows, rawRow.(map[string]any))
		}
		rowAfter := findRow(t, rows, recordID)
		requireViewRowFieldSurface(t, "timeline-resolution", rowAfter, timeline.TimelineViewSchemaID)
		itemAfter := requireSingleCollectionItem(t, rowAfter, timelinetest.FieldHostRefs)
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
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusUnresolved)
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

// timeline-resolution / REQ-01-311, REQ-01-314..REQ-01-320, REQ-02-248, REQ-03-280 / AC-394, AC-396, AC-397.
func TestManualTimelineConfidenceNull_Integration(t *testing.T) {
	t.Run("create route add_resolved_ref persists manual host link and replays cleanly", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-09-create-add-resolved-ref")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-09-create-incident",
			"incident_key":  "IR-ENTITY-LINKING-I409-D",
			"title":         "Record relationships timeline-resolution create add_resolved_ref",
		})
		incidentID := incident["incident_id"].(string)
		entitytest.SeedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "WS-023.corp.example", "WS-023.corp.example", "", "")

		createPayload := map[string]any{
			"client_txn_id":                   "txn-entity_linking-i-4-09-create-row",
			"timeline.activity_synopsis_text": "Create-boundary manual host relationship row",
			timelinetest.FieldHostRefs: timelinetest.CollectionActions(
				timelinetest.AddResolvedRefAction("WS-023", entitytest.CanonicalHostRecordID),
			),
		}
		createResp := doJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			createPayload,
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createData := requireSuccessEnvelopeWithBody(t, createResp, http.StatusCreated)["data"].(map[string]any)
		recordID := createData["row"].(map[string]any)["record_id"].(string)
		requireMutationRecorded(t, harness.DB, createData["change_set_id"].(string), recordID, adminID, "timeline.rows.create", "txn-entity_linking-i-4-09-create-row", 2, 1)

		link := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), entitytest.CanonicalHostRecordID, "observed_on_host")
		linktest.RequireActiveLink(
			t,
			link,
			mustUUID(t, recordID),
			entitytest.CanonicalHostRecordID,
			"observed_on_host",
			"manual",
			nil,
		)

		envelope := queryTimelineEnvelope(t, harness.Server, incidentID, adminLogin)
		contractassert.RequireDefaultQueryMeta(t, envelope, timeline.TimelineViewSchemaID)
		rawRows := envelope["data"].(map[string]any)["rows"].([]any)
		rows := make([]map[string]any, 0, len(rawRows))
		for _, rawRow := range rawRows {
			rows = append(rows, rawRow.(map[string]any))
		}
		row := findRow(t, rows, recordID)
		requireViewRowFieldSurface(t, "timeline-resolution", row, timeline.TimelineViewSchemaID)
		item := requireSingleCollectionItem(t, row, timelinetest.FieldHostRefs)
		if item["item_kind"] != "resolved_ref" || item["resolved_record_id"] != entitytest.CanonicalHostRecordID.String() {
			t.Fatalf("unexpected create-route current-state item: %#v", item)
		}
		if item["raw_text"] != "WS-023" {
			t.Fatalf("expected create-route raw_text preservation, got %#v", item)
		}
		if confidence, ok := item["confidence"]; !ok || confidence != nil {
			t.Fatalf("expected create-route current-state confidence:null, got %#v", item)
		}
		if item["provenance"] != "manual" {
			t.Fatalf("unexpected create-route provenance: %#v", item)
		}

		countersAfterCreate := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
		replayResp := doJSON(
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
		contractassert.RequireReplayScaffold(t, contractassert.ReplayExpectation{
			FirstStatus:     http.StatusCreated,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: contractassert.ReplayCounts{
				ChangeSets:   countersAfterCreate.ChangeSets,
				MutationRows: countersAfterCreate.MutationRows,
				Revisions:    countersAfterCreate.Revisions,
			},
			StableAfter: contractassert.ReplayCounts{
				ChangeSets:   asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID).ChangeSets,
				MutationRows: asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID).MutationRows,
				Revisions:    asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID).Revisions,
			},
		})

		divergentResp := doJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-entity_linking-i-4-09-create-row",
				timelinetest.FieldHostRefs: timelinetest.CollectionActions(
					timelinetest.AddResolvedRefAction("WS-024", entitytest.CanonicalHostRecordID),
				),
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		contractassert.RequireDivergentReplayRejected(
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
		deniedResp := doJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-entity_linking-i-4-09-create-denied",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		deniedBody := httptestx.RequireErrorEnvelope(t, deniedResp, http.StatusForbidden, "authorization_denied")
		contractassert.RequireAuthorizationReDerived(
			t,
			contractassert.AuthorizationOutcome{Status: http.StatusCreated},
			contractassert.AuthorizationOutcome{Status: deniedResp.StatusCode, Code: deniedBody["error"].(map[string]any)["code"].(string)},
		)
	})

	t.Run("add_resolved_ref persists manual host link with null confidence", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-u-4-09-add-resolved-ref")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-09-add-incident",
			"incident_key":  "IR-ENTITY-LINKING-U409-A",
			"title":         "Record relationships timeline-resolution add_resolved_ref",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "WS-023.corp.example", "WS-023.corp.example")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-entity_linking-u-4-09-add-row",
			"timeline.activity_synopsis_text": "Manual host relationship row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)
		beforeCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)

		resp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				1,
				"txn-entity_linking-i-4-09-add-patch",
				timelinetest.CollectionActions(
					timelinetest.AddResolvedRefAction("WS-023", entitytest.CanonicalHostRecordID),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
		afterCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
		revisiontest.RequireExactlyOneChangeSet(t, beforeCounters.ChangeSets, afterCounters.ChangeSets)
		if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("unexpected add_resolved_ref row_version: got %d want 2", got)
		}
		requireMutationRecorded(t, harness.DB, data["change_set_id"].(string), recordID, adminID, "timeline.records.patch", "txn-entity_linking-i-4-09-add-patch", 2, 2)

		link := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), entitytest.CanonicalHostRecordID, "observed_on_host")
		linktest.RequireActiveLink(
			t,
			link,
			mustUUID(t, recordID),
			entitytest.CanonicalHostRecordID,
			"observed_on_host",
			"manual",
			nil,
		)
		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requireViewRowFieldSurface(t, "timeline-resolution", row, timeline.TimelineViewSchemaID)
		item := requireSingleCollectionItem(t, row, timelinetest.FieldHostRefs)
		if item["item_kind"] != "resolved_ref" {
			t.Fatalf("expected resolved_ref item after add_resolved_ref, got %#v", item)
		}
		if item["resolved_record_id"] != entitytest.CanonicalHostRecordID.String() {
			t.Fatalf("unexpected resolved_record_id after add_resolved_ref: %#v", item)
		}
		if item["raw_text"] != "WS-023" {
			t.Fatalf("expected raw_text to remain authoritative in current-state read, got %#v", item)
		}
		if confidence, ok := item["confidence"]; !ok || confidence != nil {
			t.Fatalf("expected current-state read to preserve confidence:null, got %#v", item)
		}
		if item["provenance"] != "manual" {
			t.Fatalf("unexpected current-state link provenance: %#v", item)
		}

		replayResp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				1,
				"txn-entity_linking-i-4-09-add-patch",
				timelinetest.CollectionActions(
					timelinetest.AddResolvedRefAction("WS-023", entitytest.CanonicalHostRecordID),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		replayData := requireSuccessEnvelopeWithBody(t, replayResp, http.StatusOK)["data"].(map[string]any)
		if replayData["change_set_id"] != data["change_set_id"] {
			t.Fatalf("expected add_resolved_ref replay to reuse original payload, got %#v %#v", data, replayData)
		}
		contractassert.RequireReplayScaffold(t, contractassert.ReplayExpectation{
			FirstStatus:     http.StatusOK,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore: contractassert.ReplayCounts{
				ChangeSets:   afterCounters.ChangeSets,
				MutationRows: afterCounters.MutationRows,
				Revisions:    afterCounters.Revisions,
			},
			StableAfter: contractassert.ReplayCounts{
				ChangeSets:   asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID).ChangeSets,
				MutationRows: asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID).MutationRows,
				Revisions:    asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID).Revisions,
			},
		})

		divergentResp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				1,
				"txn-entity_linking-i-4-09-add-patch",
				timelinetest.CollectionActions(
					timelinetest.AddResolvedRefAction("WS-024", entitytest.CanonicalHostRecordID),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		contractassert.RequireDivergentReplayRejected(
			t,
			divergentResp.StatusCode,
			httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")["error"].(map[string]any)["code"].(string),
			"client_txn_conflict",
		)

		entitytest.SeedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.DuplicateHostRecordID, "WS-024.corp.example", "WS-024.corp.example", "", "")
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
		deniedResp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				2,
				"txn-entity_linking-i-4-09-add-denied",
				timelinetest.CollectionActions(
					timelinetest.AddResolvedRefAction("WS-024", entitytest.DuplicateHostRecordID),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		deniedBody := httptestx.RequireErrorEnvelope(t, deniedResp, http.StatusForbidden, "authorization_denied")
		contractassert.RequireAuthorizationReDerived(
			t,
			contractassert.AuthorizationOutcome{Status: http.StatusOK},
			contractassert.AuthorizationOutcome{Status: deniedResp.StatusCode, Code: deniedBody["error"].(map[string]any)["code"].(string)},
		)
	})

	t.Run("resolve_item persists manual identity link with null confidence", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-u-4-09-resolve-item")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-09-resolve-incident",
			"incident_key":  "IR-ENTITY-LINKING-U409-B",
			"title":         "Record relationships timeline-resolution resolve_item",
		})
		incidentID := incident["incident_id"].(string)
		seedIdentityRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalIdentityRecordID, "Alex Analyst", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-09-resolve-row",
			timelinetest.FieldIdentityRefs: timelinetest.CollectionActions(
				timelinetest.AddTokenAction("alex.analyst@example.test"),
			),
		})
		record := created["row"].(map[string]any)
		recordID := record["record_id"].(string)
		unresolvedItem := requireSingleCollectionItem(t, record, timelinetest.FieldIdentityRefs)
		mentionID := mentionIDFromItemRef(t, unresolvedItem["item_ref"].(string))
		beforeMention := lookupMention(t, harness.DB, mentionID)

		resp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldIdentityRefs,
				1,
				"txn-entity_linking-u-4-09-resolve-patch",
				timelinetest.CollectionActions(
					entitytest.ResolveItemAction(unresolvedItem["item_ref"].(string), entitytest.CanonicalIdentityRecordID),
				),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		data := requireSuccessEnvelopeWithBody(t, resp, http.StatusOK)["data"].(map[string]any)
		if got := int64(data["row"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("unexpected resolve_item row_version: got %d want 2", got)
		}

		link := lookupActiveLink(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, recordID), entitytest.CanonicalIdentityRecordID, "observed_as_identity")
		linktest.RequireActiveLink(
			t,
			link,
			mustUUID(t, recordID),
			entitytest.CanonicalIdentityRecordID,
			"observed_as_identity",
			"manual",
			nil,
		)
		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requireViewRowFieldSurface(t, "timeline-resolution", row, timeline.TimelineViewSchemaID)
		item := requireSingleCollectionItem(t, row, timelinetest.FieldIdentityRefs)
		if item["item_kind"] != "resolved_ref" {
			t.Fatalf("expected resolved_ref item after resolve_item, got %#v", item)
		}
		if item["resolved_record_id"] != entitytest.CanonicalIdentityRecordID.String() {
			t.Fatalf("unexpected resolved_record_id after resolve_item: %#v", item)
		}
		if item["raw_text"] != beforeMention.RawText {
			t.Fatalf("expected resolve_item to preserve raw_text, before=%q after=%#v", beforeMention.RawText, item)
		}
		if confidence, ok := item["confidence"]; !ok || confidence != nil {
			t.Fatalf("expected current-state read to preserve confidence:null, got %#v", item)
		}
		if item["provenance"] != "manual" {
			t.Fatalf("unexpected current-state link provenance: %#v", item)
		}
	})

	t.Run("resolve_item without a target is rejected without creating an identity", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-u-4-09-resolve-create")
		adminLogin, _ := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-09-resolve-create-incident",
			"incident_key":  "IR-ENTITY-LINKING-U409-D",
			"title":         "Record relationships timeline-resolution resolve_item create",
		})
		incidentID := incident["incident_id"].(string)
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-entity_linking-u-4-09-resolve-create-row",
			"timeline.activity_synopsis_text": "Manual identity create-from-mention row",
			timelinetest.FieldIdentityRefs: timelinetest.CollectionActions(
				timelinetest.AddTokenAction("vpn.user@example.test"),
			),
		})
		record := created["row"].(map[string]any)
		recordID := record["record_id"].(string)
		unresolvedItem := requireSingleCollectionItem(t, record, timelinetest.FieldIdentityRefs)
		mentionID := mentionIDFromItemRef(t, unresolvedItem["item_ref"].(string))
		beforeCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)

		resp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldIdentityRefs,
				1,
				"txn-entity_linking-u-4-09-resolve-create-patch",
				timelinetest.CollectionActions(
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

		afterCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
		if afterCounters != beforeCounters {
			t.Fatalf("expected rejected resolve_item to leave counters unchanged, before=%+v after=%+v", beforeCounters, afterCounters)
		}
		mention := lookupMention(t, harness.DB, mentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusUnresolved)
		if mention.ResolvedRecordID != nil {
			t.Fatalf("expected mention to remain unresolved, got %#v", mention)
		}
		if got := queryCount(t, harness.DB, `SELECT COUNT(*) FROM identities WHERE incident_id::text = $1`, incidentID); got != 0 {
			t.Fatalf("expected missing-target resolve to create no identity rows, got %d", got)
		}
	})

	t.Run("client supplied confidence is rejected without side effects", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-u-4-09-reject")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-09-reject-incident",
			"incident_key":  "IR-ENTITY-LINKING-U409-C",
			"title":         "Record relationships timeline-resolution rejection",
		})
		incidentID := incident["incident_id"].(string)
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-09-reject-row",
			timelinetest.FieldHostRefs: timelinetest.CollectionActions(
				timelinetest.AddTokenAction("WS-023"),
			),
		})
		record := created["row"].(map[string]any)
		recordID := record["record_id"].(string)
		unresolvedItem := requireSingleCollectionItem(t, record, timelinetest.FieldHostRefs)
		mentionID := mentionIDFromItemRef(t, unresolvedItem["item_ref"].(string))
		beforeMention := lookupMention(t, harness.DB, mentionID)
		beforeCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
		asserttest.AwaitIncidentStreamIdle(t, asserttest.SQLDatabase(harness.DB), incidentID)
		socket := connectTimelineSocket(t, harness.Server, incidentID, adminLogin.sessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hubChanges, unsubscribe := harness.Collaboration.SubscribeIncident(mustUUID(t, incidentID), 4)
		defer unsubscribe()

		resp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				1,
				"txn-entity_linking-u-4-09-reject-patch",
				timelinetest.CollectionActions(
					map[string]any{
						"op":                 "resolve_item",
						"item_ref":           unresolvedItem["item_ref"].(string),
						"resolved_record_id": entitytest.CanonicalHostRecordID.String(),
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

		afterCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
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
`, "timeline.records.patch", adminID, recordID, "txn-entity_linking-u-4-09-reject-patch"); got != 0 {
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
`, incidentID, recordID, entitytest.CanonicalHostRecordID.String()); got != 0 {
			t.Fatalf("rejected payload must not create a record_link, got %d", got)
		}

		afterMention := lookupMention(t, harness.DB, mentionID)
		envelopetest.RequireRowVersionStable(t, beforeMention.RowVersion, afterMention.RowVersion)
		entitytest.RequireMentionStatus(t, afterMention, entitytest.MentionStatusUnresolved)
		entitytest.RequireRawTextPreserved(t, beforeMention.RawText, afterMention.RawText)

		row := findRow(t, queryTimelineRows(t, harness.Server, incidentID, adminLogin), recordID)
		requireViewRowFieldSurface(t, "timeline-resolution", row, timeline.TimelineViewSchemaID)
		if got := int64(row["row_version"].(float64)); got != 1 {
			t.Fatalf("rejected payload must not advance row_version, got %d", got)
		}
		item := requireSingleCollectionItem(t, row, timelinetest.FieldHostRefs)
		if item["item_kind"] != "unresolved_mention" {
			t.Fatalf("rejected payload must not refresh misleading resolution state, got %#v", item)
		}

		asserttest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
		expectNoTimelineSocketMessage(t, socket)
	})
}

func lookupActiveLink(t testing.TB, db *sql.DB, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, linkType string) linktest.LinkFixture {
	t.Helper()

	var (
		link        linktest.LinkFixture
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

func requireViewRowFieldSurface(t testing.TB, testID string, row map[string]any, viewSchemaID string) {
	t.Helper()

	contractassert.RequireFieldKeyConformance(
		t,
		workbookscenariotest.SortedRowFieldKeys(t, row),
		workbookscenariotest.AllowedFieldKeys(t, testID, viewSchemaID),
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

func lookupMention(t testing.TB, db *sql.DB, mentionID uuid.UUID) entitytest.MentionFixture {
	t.Helper()

	var mention entitytest.MentionFixture
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
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "host")

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO hosts (record_id, incident_id, display_name, hostname, host_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, 'canonical', $5, $5)
`, recordID, incidentID, displayName, hostname, actorUserID); err != nil {
		t.Fatalf("seed host record: %v", err)
	}
}

func seedIdentityRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, displayName string, upn string, email string, samAccountName string) {
	t.Helper()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "identity")

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

func intPointer(value int) *int {
	return &value
}
