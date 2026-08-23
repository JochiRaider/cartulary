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
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
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

type failAfterMentionInsertPort struct {
	timeline.MentionPort
	err error
}

func (port failAfterMentionInsertPort) InsertTx(ctx context.Context, tx pgx.Tx, params timeline.MentionCreateParams) error {
	if err := port.MentionPort.InsertTx(ctx, tx, params); err != nil {
		return err
	}
	return port.err
}

type failAfterLinkUpsertPort struct {
	timeline.LinkPort
	err error
}

func (port failAfterLinkUpsertPort) UpsertLinkCommandTx(ctx context.Context, tx pgx.Tx, command timeline.UpsertLinkCommand) (timeline.RecordLinkCommandResult, error) {
	result, err := port.LinkPort.UpsertLinkCommandTx(ctx, tx, command)
	if err != nil {
		return timeline.RecordLinkCommandResult{}, err
	}
	return result, port.err
}

type failAfterRevisionAppendPort struct {
	timeline.RevisionPort
	err error
}

func (port failAfterRevisionAppendPort) AppendRecordRevisionTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordRevisionParams) error {
	if err := port.RevisionPort.AppendRecordRevisionTx(ctx, tx, params); err != nil {
		return err
	}
	return port.err
}

type failAfterEntityProjectionPort struct {
	timeline.EntityProjectionPort
	err error
}

func (port failAfterEntityProjectionPort) RefreshHostTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if err := port.EntityProjectionPort.RefreshHostTx(ctx, tx, recordID); err != nil {
		return err
	}
	return port.err
}

func (port failAfterEntityProjectionPort) RefreshIdentityTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if err := port.EntityProjectionPort.RefreshIdentityTx(ctx, tx, recordID); err != nil {
		return err
	}
	return port.err
}

type failAfterTimelineProjectionPort struct {
	workbookprojection.Writer
	err error
}

func (port failAfterTimelineProjectionPort) ApplyTimelineMutationTx(ctx context.Context, tx pgx.Tx, mutation workbookprojection.ProjectionMutation) error {
	if err := port.Writer.ApplyTimelineMutationTx(ctx, tx, mutation); err != nil {
		return err
	}
	return port.err
}

type failAfterCollaborationPort struct {
	timeline.CollaborationPort
	err error
}

func (port failAfterCollaborationPort) AppendRecordChangeIntentTx(ctx context.Context, tx pgx.Tx, params timeline.RecordChangeIntentParams) error {
	if err := port.CollaborationPort.AppendRecordChangeIntentTx(ctx, tx, params); err != nil {
		return err
	}
	return port.err
}

type failAfterIdempotencyPort struct {
	timeline.IdempotencyPort
	err error
}

func (port failAfterIdempotencyPort) InsertRouteIdempotencyPayload(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey, targetUserID *uuid.UUID, requestHash []byte, statusCode int, payload any) error {
	if err := port.IdempotencyPort.InsertRouteIdempotencyPayload(ctx, tx, key, targetUserID, requestHash, statusCode, payload); err != nil {
		return err
	}
	return port.err
}

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

		patchPayload := timelinetest.TimelineCollectionPatchPayload(
			timelinetest.FieldHostRefs,
			1,
			"txn-entity_linking-u-4-08-host-patch",
			timelinetest.CollectionActions(
				timelinetest.AddTokenAction(" vpn   gateway "),
			),
		)
		resp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			patchPayload,
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

		beforeReplay := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
		replayResp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			patchPayload,
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		replayData := requireSuccessEnvelopeWithBody(t, replayResp, http.StatusOK)["data"].(map[string]any)
		if replayData["change_set_id"] != data["change_set_id"] {
			t.Fatalf("exact auto-resolution replay must return the original change set: got %#v want %#v", replayData["change_set_id"], data["change_set_id"])
		}
		if afterReplay := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID); afterReplay != beforeReplay {
			t.Fatalf("exact auto-resolution replay must be zero-write: before=%+v after=%+v", beforeReplay, afterReplay)
		}

		divergentResp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				1,
				"txn-entity_linking-u-4-08-host-patch",
				timelinetest.CollectionActions(timelinetest.AddTokenAction("different alias")),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")
		if afterConflict := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID); afterConflict != beforeReplay {
			t.Fatalf("divergent auto-resolution replay must be zero-write: before=%+v after=%+v", beforeReplay, afterConflict)
		}

		undoResp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				2,
				"txn-entity_linking-u-4-08-host-undo",
				timelinetest.CollectionActions(map[string]any{
					"op":       "revert_to_unresolved",
					"item_ref": item["item_ref"],
				}),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		undoData := requireSuccessEnvelopeWithBody(t, undoResp, http.StatusOK)["data"].(map[string]any)
		if got := int64(undoData["row"].(map[string]any)["row_version"].(float64)); got != 3 {
			t.Fatalf("auto-resolution Undo must advance row_version to 3, got %d", got)
		}
		undoItem := requireSingleCollectionItem(t, undoData["row"].(map[string]any), timelinetest.FieldHostRefs)
		if undoItem["item_kind"] != "unresolved_mention" || undoItem["raw_text"] != " vpn   gateway " {
			t.Fatalf("Undo must restore the authoritative unresolved token, got %#v", undoItem)
		}
		undoneMention := lookupMention(t, harness.DB, mentionID)
		entitytest.RequireMentionStatus(t, undoneMention, entitytest.MentionStatusUnresolved)
		if undoneMention.ResolvedRecordID != nil || undoneMention.ResolvedByUserID != nil || undoneMention.ResolvedAt != nil || undoneMention.ResolutionMethod != nil {
			t.Fatalf("Undo must clear active resolution metadata, got %#v", undoneMention)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND deleted_at IS NULL
`, incidentID, recordID); got != 0 {
			t.Fatalf("Undo must remove the active auto-match link, got %d", got)
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

	t.Run("no-match suppressor and forbidden rewrite tokens remain unresolved", func(t *testing.T) {
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

		tokenCases := append([]string{"WS-999"}, entitytest.AutoResolutionSuppressedTokens...)
		for _, rawText := range tokenCases {
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
		if items[0]["raw_text"] != " vpn   gateway " || items[1]["raw_text"] != "WS-023?" {
			t.Fatalf("mixed auto-resolution batch must preserve submitted action order, got %#v", items)
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

	t.Run("authorization precedes lifecycle evaluation and both rejection paths are zero-write", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-u-4-08-security-order")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-08-security-incident",
			"incident_key":  "IR-ENTITY-LINKING-U408-SECURITY",
			"title":         "Record relationships Timeline authorization ordering",
		})
		incidentID := incident["incident_id"].(string)
		seedHostRecord(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "Security host", "security-host")
		seedEntityAlias(t, harness.DB, mustUUID(t, incidentID), mustUUID(t, adminID), entitytest.CanonicalHostRecordID, "host", "Security Host Alias")
		created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-entity_linking-u-4-08-security-row",
			"timeline.activity_synopsis_text": "Security ordering row",
		})
		recordID := created["row"].(map[string]any)["record_id"].(string)
		replacement := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-entity_linking-u-4-08-security-replacement",
			"timeline.activity_synopsis_text": "Security ordering replacement",
		})
		replacementID := replacement["row"].(map[string]any)["record_id"].(string)
		supersedeResp := doJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID+"/supersede",
			map[string]any{
				"base_row_version":      1,
				"client_txn_id":         "txn-entity_linking-u-4-08-security-supersede",
				"reason":                "exercise authorization and lifecycle ordering",
				"replacement_record_id": replacementID,
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireSuccessEnvelope(t, supersedeResp, http.StatusOK)

		viewerEmail := "auto-resolution-viewer@example.test"
		viewerPassword := "AutoResolutionViewer1!"
		viewerID := seedLocalUserFlags(t, harness.DB, viewerEmail, "Auto Resolution Viewer", viewerPassword, false, false, true)
		createMembership(t, harness.Server, incidentID, viewerID, viewerEmail, "viewer", adminLogin)
		viewerSession, viewerCSRF := loginLocalUser(t, harness.Server, viewerEmail, viewerPassword)
		before := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
		viewerTxnID := "txn-entity_linking-u-4-08-security-viewer-patch"
		viewerResp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				2,
				viewerTxnID,
				timelinetest.CollectionActions(timelinetest.AddTokenAction(" Security   Host Alias ")),
			),
			withCookies(viewerSession, viewerCSRF),
			withHeader(authn.CSRFHeaderName, viewerCSRF.Value),
		)
		httptestx.RequireErrorEnvelope(t, viewerResp, http.StatusForbidden, "authorization_denied")
		if after := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID); after != before {
			t.Fatalf("authorization rejection must be zero-write: before=%+v after=%+v", before, after)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = 'timeline.records.patch'
   AND actor_user_id::text = $1
   AND scope_key = $2
   AND client_txn_id = $3
`, viewerID, recordID, viewerTxnID); got != 0 {
			t.Fatalf("authorization rejection must not persist idempotency, got %d", got)
		}

		adminTxnID := "txn-entity_linking-u-4-08-security-admin-patch"
		adminResp := doJSON(
			t,
			http.MethodPatch,
			harness.Server.HTTP.URL+"/api/v1/records/"+recordID,
			timelinetest.TimelineCollectionPatchPayload(
				timelinetest.FieldHostRefs,
				2,
				adminTxnID,
				timelinetest.CollectionActions(timelinetest.AddTokenAction(" Security   Host Alias ")),
			),
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, adminResp, http.StatusConflict, "illegal_transition")
		if after := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID); after != before {
			t.Fatalf("lifecycle rejection must be zero-write: before=%+v after=%+v", before, after)
		}
		if got := queryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id::text = $1`, recordID); got != 0 {
			t.Fatalf("rejected auto-resolution must not persist mentions, got %d", got)
		}
		if got := queryCount(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE src_record_id::text = $1 AND link_type = 'observed_on_host' AND deleted_at IS NULL`, recordID); got != 0 {
			t.Fatalf("rejected auto-resolution must not persist Host links, got %d", got)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = 'timeline.records.patch'
   AND actor_user_id::text = $1
   AND scope_key = $2
   AND client_txn_id = $3
`, adminID, recordID, adminTxnID); got != 0 {
			t.Fatalf("lifecycle rejection must not persist idempotency, got %d", got)
		}
	})

	t.Run("import create preserves matching tokens as unresolved observations", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-u-4-08-import-ineligible")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-08-import-incident",
			"incident_key":  "IR-ENTITY-LINKING-U408-IMPORT",
			"title":         "Record relationships Timeline import eligibility",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		actorID := mustUUID(t, adminID)
		seedHostRecord(t, harness.DB, incidentID, actorID, entitytest.CanonicalHostRecordID, "Import host", "import-host")
		seedEntityAlias(t, harness.DB, incidentID, actorID, entitytest.CanonicalHostRecordID, "host", "Import Host Alias")

		ctx := context.Background()
		tx, err := harness.Pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin Timeline import transaction: %v", err)
		}
		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
		now := time.Now().UTC()
		changeSetID, err := harness.Revisions.Appender().AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
			IncidentID:  incidentID,
			ActorUserID: actorID,
			Source:      "imports.unit.apply",
			CreatedAt:   now,
		})
		if err != nil {
			t.Fatalf("append Timeline import change set: %v", err)
		}
		synopsis := "Imported matching Host token"
		facade := timelineFacadeWithCollaboratorMutation(t, harness, nil)
		response, err := facade.CreateImportRowTx(ctx, tx, ownerfacade.ImportOwnerCreateCommand{
			Request: ownerfacade.ImportOwnerCreateRequest{
				IncidentID:         incidentID,
				ActorUserID:        actorID,
				TargetViewSchemaID: timeline.TimelineViewSchemaID,
				ImportSessionID:    uuid.New(),
				ImportUnitID:       uuid.New(),
				MappingFingerprint: "timeline-auto-resolution-import-ineligible",
				SourceFileKind:     "csv",
				ParserProfileID:    "synthetic",
				ParserVersion:      "1",
				LocatorKind:        "row",
				Locator:            "1",
				ClientTxnID:        "txn-entity_linking-u-4-08-import-row",
				FieldValues: []ownerfacade.ImportFieldValue{
					{
						FieldKey:        "timeline.activity_synopsis_text",
						NormalizedValue: ownerfacade.ImportScalarValue{Kind: "text", Text: &synopsis},
					},
					{
						FieldKey: timelinetest.FieldHostRefs,
						NormalizedValue: ownerfacade.ImportScalarValue{
							Kind: "collection_token",
							CollectionToken: &ownerfacade.ImportCollectionToken{
								RawText:        " Import   Host Alias ",
								NormalizedText: "Import Host Alias",
							},
						},
					},
				},
			},
			ChangeSetID: changeSetID,
			SequenceNo:  1,
			Now:         now,
		})
		if err != nil {
			t.Fatalf("create imported Timeline row: %v", err)
		}
		item := requireSingleCollectionItem(t, response.RowRefresh, timelinetest.FieldHostRefs)
		if item["item_kind"] != "unresolved_mention" || item["raw_text"] != " Import   Host Alias " {
			t.Fatalf("import create must preserve an exact-match token as unresolved, got %#v", item)
		}
		var mentionCount int
		if err := tx.QueryRow(ctx, `
SELECT COUNT(*)
  FROM entity_mentions
 WHERE source_record_id = $1
   AND resolution_status = 'unresolved'
   AND resolved_record_id IS NULL
`, response.RecordID).Scan(&mentionCount); err != nil || mentionCount != 1 {
			t.Fatalf("import create must persist one unresolved mention in the caller transaction: count=%d err=%v", mentionCount, err)
		}
		var linkCount int
		if err := tx.QueryRow(ctx, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND deleted_at IS NULL
`, incidentID, response.RecordID).Scan(&linkCount); err != nil || linkCount != 0 {
			t.Fatalf("import create must not auto-create a resolved link: count=%d err=%v", linkCount, err)
		}
	})

	t.Run("Entities fact port filters aliases and conceals invalid targets on the borrowed transaction", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-u-4-08-entities-port")
		adminLogin, adminID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-08-port-incident",
			"incident_key":  "IR-ENTITY-LINKING-U408-PORT-A",
			"title":         "Record relationships Timeline Entities port",
		})
		otherIncident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-u-4-08-port-other-incident",
			"incident_key":  "IR-ENTITY-LINKING-U408-PORT-B",
			"title":         "Record relationships Timeline Entities port isolation",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		otherIncidentID := mustUUID(t, otherIncident["incident_id"].(string))
		actorID := mustUUID(t, adminID)
		activeHostID := uuid.New()
		mergedHostID := uuid.New()
		crossIncidentHostID := uuid.New()
		identityID := uuid.New()
		seedHostRecord(t, harness.DB, incidentID, actorID, activeHostID, "Active port host", "active-port-host")
		seedHostRecord(t, harness.DB, incidentID, actorID, mergedHostID, "Merged port host", "merged-port-host")
		seedHostRecord(t, harness.DB, otherIncidentID, actorID, crossIncidentHostID, "Cross-incident port host", "cross-port-host")
		seedIdentityRecord(t, harness.DB, incidentID, actorID, identityID, "Port identity", "port@example.test", "port@example.test", "PORT")
		seedEntityAlias(t, harness.DB, incidentID, actorID, activeHostID, "host", "Active Alias")
		seedEntityAlias(t, harness.DB, incidentID, actorID, mergedHostID, "host", "Merged Alias")
		seedEntityAlias(t, harness.DB, otherIncidentID, actorID, crossIncidentHostID, "host", "Cross Alias")
		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE hosts
   SET host_state = 'merged', merged_into_record_id = $1
 WHERE record_id = $2
`, activeHostID, mergedHostID); err != nil {
			t.Fatalf("mark port host merged: %v", err)
		}

		ctx := context.Background()
		tx, err := harness.Pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin borrowed Entities fact transaction: %v", err)
		}
		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
		store := hostidentity.NewSourceFacts()

		aliases, err := store.ListEligibleAliasesTx(ctx, tx, incidentID, "host")
		if err != nil {
			t.Fatalf("list eligible host aliases: %v", err)
		}
		if len(aliases) != 1 || aliases[0].RecordID != activeHostID || aliases[0].RawText != "Active Alias" {
			t.Fatalf("eligible aliases must contain only same-incident active Host candidates, got %#v", aliases)
		}
		unsupportedAliases, err := store.ListEligibleAliasesTx(ctx, tx, incidentID, "indicator")
		if err != nil || len(unsupportedAliases) != 0 {
			t.Fatalf("unsupported entity types must have no eligible aliases: aliases=%#v err=%v", unsupportedAliases, err)
		}
		if err := store.ValidateResolvedTargetTx(ctx, tx, incidentID, "host", activeHostID); err != nil {
			t.Fatalf("validate active same-incident Host target: %v", err)
		}
		invalidTargets := []struct {
			name       string
			entityType string
			recordID   uuid.UUID
		}{
			{name: "cross incident", entityType: "host", recordID: crossIncidentHostID},
			{name: "wrong entity type", entityType: "host", recordID: identityID},
			{name: "inactive lifecycle", entityType: "host", recordID: mergedHostID},
			{name: "missing record", entityType: "host", recordID: uuid.New()},
			{name: "unsupported entity type", entityType: "indicator", recordID: activeHostID},
		}
		for _, target := range invalidTargets {
			t.Run(target.name, func(t *testing.T) {
				err := store.ValidateResolvedTargetTx(ctx, tx, incidentID, target.entityType, target.recordID)
				if !errors.Is(err, hostidentity.ErrHostIdentityRecordNotFound) {
					t.Fatalf("invalid targets must use the concealed not-found result, got %v", err)
				}
			})
		}
		var one int
		if err := tx.QueryRow(ctx, `SELECT 1`).Scan(&one); err != nil || one != 1 {
			t.Fatalf("Entities fact port must leave the borrowed transaction usable: one=%d err=%v", one, err)
		}
	})

	t.Run("every transactional participant rolls back auto-resolution as one unit", func(t *testing.T) {
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

		faults := []struct {
			name   string
			mutate func(*timeline.Collaborators, error)
		}{
			{
				name: "mention insertion",
				mutate: func(collaborators *timeline.Collaborators, forced error) {
					collaborators.Collections.Mentions = failAfterMentionInsertPort{MentionPort: collaborators.Collections.Mentions, err: forced}
				},
			},
			{
				name: "link upsert",
				mutate: func(collaborators *timeline.Collaborators, forced error) {
					collaborators.Collections.Links = failAfterLinkUpsertPort{LinkPort: collaborators.Collections.Links, err: forced}
				},
			},
			{
				name: "revision append",
				mutate: func(collaborators *timeline.Collaborators, forced error) {
					collaborators.Core.Revisions = failAfterRevisionAppendPort{RevisionPort: collaborators.Core.Revisions, err: forced}
				},
			},
			{
				name: "entity projection refresh",
				mutate: func(collaborators *timeline.Collaborators, forced error) {
					collaborators.Commit.EntityProjection = failAfterEntityProjectionPort{EntityProjectionPort: collaborators.Commit.EntityProjection, err: forced}
				},
			},
			{
				name: "Timeline projection refresh",
				mutate: func(collaborators *timeline.Collaborators, forced error) {
					collaborators.Commit.Projection = failAfterTimelineProjectionPort{Writer: collaborators.Commit.Projection, err: forced}
				},
			},
			{
				name: "collaboration intent publication",
				mutate: func(collaborators *timeline.Collaborators, forced error) {
					collaborators.Commit.Collaboration = failAfterCollaborationPort{CollaborationPort: collaborators.Commit.Collaboration, err: forced}
				},
			},
			{
				name: "idempotency persistence",
				mutate: func(collaborators *timeline.Collaborators, forced error) {
					collaborators.Core.Idempotency = failAfterIdempotencyPort{IdempotencyPort: collaborators.Core.Idempotency, err: forced}
				},
			},
		}

		for index, fault := range faults {
			t.Run(fault.name, func(t *testing.T) {
				caseKey := strings.NewReplacer(" ", "-", "_", "-").Replace(strings.ToLower(fault.name))
				created := createTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
					"client_txn_id":                   "txn-entity_linking-u-4-08-rollback-row-" + caseKey,
					"timeline.activity_synopsis_text": "Rollback row " + fault.name,
				})
				recordID := created["row"].(map[string]any)["record_id"].(string)
				recordUUID := mustUUID(t, recordID)
				clientTxnID := "txn-entity_linking-u-4-08-rollback-patch-" + caseKey
				forced := errors.New("forced auto-match rollback after " + fault.name)

				beforeCounters := asserttest.SnapshotCounters(t, asserttest.SQLDatabase(harness.DB), incidentID, recordID)
				beforeCollaboration := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM collaboration_event_intents
 WHERE source_record_id::text = $1
`, recordID)

				facade := timelineFacadeWithCollaboratorMutation(t, harness, func(collaborators *timeline.Collaborators) {
					fault.mutate(collaborators, forced)
				})
				normalized, ok := fieldnorm.NormalizeMentionToken(" vpn   gateway ")
				if !ok {
					t.Fatal("normalize rollback mention token")
				}
				request := timeline.PatchRequest{
					ViewSchemaID:   timeline.TimelineViewSchemaID,
					BaseRowVersion: 1,
					ClientTxnID:    clientTxnID,
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
					RecordID:    recordUUID,
					Request:     request,
					RequestHash: timelineadmission.PatchRequestHash(request),
					RequestID:   "req-entity_linking-u-4-08-rollback-patch-" + caseKey,
					Now:         time.Now().UTC().Add(time.Duration(index) * time.Millisecond),
				})
				if err == nil || !strings.Contains(err.Error(), forced.Error()) {
					t.Fatalf("expected %q, got %v", forced, err)
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
`, "timeline.records.patch", adminID, recordID, clientTxnID); got != 0 {
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
				if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM host_grid_projection
 WHERE record_id = $1
`, entitytest.CanonicalHostRecordID); got != 0 {
					t.Fatalf("rollback must not persist entity projection refresh, got %d", got)
				}
				if got := queryCount(t, harness.DB, `
SELECT row_version
  FROM timeline_grid_projection
 WHERE record_id::text = $1
`, recordID); got != 1 {
					t.Fatalf("rollback must preserve Timeline projection row_version=1, got %d", got)
				}
				if got := queryCount(t, harness.DB, `
SELECT row_version
  FROM records
 WHERE record_id::text = $1
`, recordID); got != 1 {
					t.Fatalf("rollback must preserve record envelope row_version=1, got %d", got)
				}
				if got := queryCount(t, harness.DB, `
SELECT row_version
  FROM timeline_events
 WHERE record_id::text = $1
`, recordID); got != 1 {
					t.Fatalf("rollback must preserve Timeline source row_version=1, got %d", got)
				}
				if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM collaboration_event_intents
 WHERE source_record_id::text = $1
`, recordID); got != beforeCollaboration {
					t.Fatalf("rollback must preserve collaboration intents: got %d want %d", got, beforeCollaboration)
				}
			})
		}
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
	var items []map[string]any
	switch rawItems := value["items"].(type) {
	case []any:
		items = make([]map[string]any, 0, len(rawItems))
		for _, rawItem := range rawItems {
			items = append(items, rawItem.(map[string]any))
		}
	case []map[string]any:
		items = append([]map[string]any(nil), rawItems...)
	default:
		t.Fatalf("expected %s collection items, got %#v", fieldKey, value["items"])
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
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, host_state,
    row_version, created_at, updated_at, created_by_user_id, updated_by_user_id
)
SELECT r.record_id, r.incident_id, $3, $4, 'canonical',
       r.row_version, r.created_at, r.updated_at,
       r.created_by_user_id, r.updated_by_user_id
  FROM records r
 WHERE r.record_id = $1
   AND r.incident_id = $2
`, recordID, incidentID, displayName, hostname); err != nil {
		t.Fatalf("seed host record: %v", err)
	}
}

func seedIdentityRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, displayName string, upn string, email string, samAccountName string) {
	t.Helper()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "identity")

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO identities (
    record_id, incident_id, display_name, upn, email, sam_account_name,
    identity_state, row_version, created_at, updated_at,
    created_by_user_id, updated_by_user_id
)
SELECT r.record_id, r.incident_id, $3, $4, $5, $6,
       'canonical', r.row_version, r.created_at, r.updated_at,
       r.created_by_user_id, r.updated_by_user_id
  FROM records r
 WHERE r.record_id = $1
   AND r.incident_id = $2
`, recordID, incidentID, displayName, upn, email, samAccountName); err != nil {
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
