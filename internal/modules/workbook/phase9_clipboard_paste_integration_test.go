package workbook_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestPhase9_I_9_01_TimelineClipboardPastePersistsOrderedMutationsAndConflicts(t *testing.T) {
	harness := phase4test.StartServer(t, "phase9-i-9-01-clipboard-paste")
	adminLogin, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-i-9-01-incident",
		"incident_key":  "IR-PHASE9-I901",
		"title":         "Phase 9 clipboard paste",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	existing := requireWorkbookCreate(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":    "txn-phase9-i-9-01-existing",
		"timeline.summary": "Existing base",
	})
	existingRow := existing["row"].(map[string]any)
	existingID := phase4test.MustUUID(t, existingRow["record_id"].(string))

	pasteData := requireClipboardPaste(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"view_schema_id":  timeline.TimelineViewSchemaID,
		"client_txn_id":   "txn-phase9-i-9-01-paste",
		"clipboard_text":  "Updated existing\tDetails from paste\tgateway-one\tpreserve-me\nCreated from paste\tCreated details\tgateway-two\tpreserve-new",
		"format":          "tsv",
		"start_field_key": "timeline.summary",
		"columns": []string{
			"timeline.summary",
			"timeline.details",
			"timeline.host_refs",
		},
		"targets": []map[string]any{
			{"kind": "record", "record_id": existingID.String(), "base_row_version": 1},
			{"kind": "create"},
		},
	}, http.StatusOK)
	changeSetID := pasteData["change_set_id"].(string)
	rows := pasteData["rows"].([]any)
	if len(rows) != 2 {
		t.Fatalf("expected two pasted rows, got %#v", rows)
	}
	patchedRow := rows[0].(map[string]any)
	createdRow := rows[1].(map[string]any)
	createdID := phase4test.MustUUID(t, createdRow["record_id"].(string))
	requireCellValue(t, patchedRow, "timeline.summary", "Updated existing")
	requireCellValue(t, createdRow, "timeline.summary", "Created from paste")

	requireChangeSetSource(t, harness, changeSetID, "timeline.clipboard_paste", "txn-phase9-i-9-01-paste")
	requireMutationTargets(t, harness, changeSetID, []string{existingID.String(), createdID.String()})
	requireRevisionCount(t, harness, existingID, 2)
	requireRevisionCount(t, harness, createdID, 1)
	requireRawCaptureValue(t, harness, existingID, "preserve-me")
	requireRawCaptureValue(t, harness, createdID, "preserve-new")
	requireMentionOriginKind(t, harness, existingID, "timeline.host_refs", "clipboard_paste")
	requireMentionOriginKind(t, harness, createdID, "timeline.host_refs", "clipboard_paste")

	firstConflictBase := requireWorkbookCreate(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":    "txn-phase9-i-9-01-conflict-base-first",
		"timeline.summary": "Conflict base first",
	})
	firstConflictRow := firstConflictBase["row"].(map[string]any)
	firstConflictID := phase4test.MustUUID(t, firstConflictRow["record_id"].(string))
	secondConflictBase := requireWorkbookCreate(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":    "txn-phase9-i-9-01-conflict-base-second",
		"timeline.summary": "Conflict base second",
	})
	secondConflictRow := secondConflictBase["row"].(map[string]any)
	secondConflictID := phase4test.MustUUID(t, secondConflictRow["record_id"].(string))
	requireWorkbookPatch(t, harness, adminLogin, firstConflictID, map[string]any{
		"view_schema_id":   timeline.TimelineViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-i-9-01-server-update-first",
		"changes": []map[string]any{{
			"field_key": "timeline.summary",
			"value":     "Server value first",
		}},
	})
	requireWorkbookPatch(t, harness, adminLogin, secondConflictID, map[string]any{
		"view_schema_id":   timeline.TimelineViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-i-9-01-server-update-second",
		"changes": []map[string]any{{
			"field_key": "timeline.summary",
			"value":     "Server value second",
		}},
	})

	partial := requireClipboardPaste(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"view_schema_id":  timeline.TimelineViewSchemaID,
		"client_txn_id":   "txn-phase9-i-9-01-partial-conflict",
		"clipboard_text":  "Client conflicting summary first\nClient conflicting summary second\nNon-conflicting created",
		"format":          "tsv",
		"start_field_key": "timeline.summary",
		"columns":         []string{"timeline.summary"},
		"targets": []map[string]any{
			{"kind": "record", "record_id": firstConflictID.String(), "base_row_version": 1},
			{"kind": "record", "record_id": secondConflictID.String(), "base_row_version": 1},
			{"kind": "create"},
		},
	}, http.StatusOK)
	if partial["change_set_id"] == changeSetID {
		t.Fatalf("partial paste must use a distinct change_set: %#v", partial)
	}
	if len(partial["rows"].([]any)) != 1 || len(partial["conflicts"].([]any)) != 2 {
		t.Fatalf("partial paste must commit only non-conflicting writes and batch conflicts: %#v", partial)
	}
	partialRows := partial["rows"].([]any)
	partialCreatedID := phase4test.MustUUID(t, partialRows[0].(map[string]any)["record_id"].(string))
	requireMutationTargets(t, harness, partial["change_set_id"].(string), []string{partialCreatedID.String()})
	conflicts := partial["conflicts"].([]any)
	firstConflict := conflicts[0].(map[string]any)
	secondConflict := conflicts[1].(map[string]any)
	requirePasteConflict(t, firstConflict, firstConflictID, "Conflict base first", "Server value first", "Client conflicting summary first")
	requirePasteConflict(t, secondConflict, secondConflictID, "Conflict base second", "Server value second", "Client conflicting summary second")
	if firstConflict["record_id"] == secondConflict["record_id"] || firstConflict["conflict_token"] == secondConflict["conflict_token"] {
		t.Fatalf("paste conflicts must remain per-cell entries with distinct identity: first=%#v second=%#v", firstConflict, secondConflict)
	}

	firstResolveData := resolveTimelineConflict(t, harness, adminLogin, firstConflictID, firstConflict["conflict_token"].(string), map[string]any{
		"conflict_token":  firstConflict["conflict_token"].(string),
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-phase9-i-9-01-resolve-conflict-first",
		"resolved_value":  "Client conflicting summary first",
	})
	firstResolveChangeSetID := firstResolveData["change_set_id"].(string)
	if firstResolveChangeSetID == partial["change_set_id"].(string) {
		t.Fatalf("conflict resolution must create a separate change_set: paste=%s resolve=%s", partial["change_set_id"], firstResolveChangeSetID)
	}
	requireChangeSetSource(t, harness, firstResolveChangeSetID, "timeline.records.conflicts.resolve", "txn-phase9-i-9-01-resolve-conflict-first")
	secondResolveData := resolveTimelineConflict(t, harness, adminLogin, secondConflictID, secondConflict["conflict_token"].(string), map[string]any{
		"conflict_token":  secondConflict["conflict_token"].(string),
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-phase9-i-9-01-resolve-conflict-second",
		"resolved_value":  "Client conflicting summary second",
	})
	secondResolveChangeSetID := secondResolveData["change_set_id"].(string)
	if secondResolveChangeSetID == partial["change_set_id"].(string) || secondResolveChangeSetID == firstResolveChangeSetID {
		t.Fatalf("each conflict resolution must create its own change_set: paste=%s first=%s second=%s", partial["change_set_id"], firstResolveChangeSetID, secondResolveChangeSetID)
	}
	requireChangeSetSource(t, harness, secondResolveChangeSetID, "timeline.records.conflicts.resolve", "txn-phase9-i-9-01-resolve-conflict-second")
}

func TestPhase9_I_9_01_EntityOriginClipboardPasteUsesSharedIngest(t *testing.T) {
	harness := phase4test.StartServer(t, "phase9-i-9-01-entity-origin-paste")
	adminLogin, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-i-9-01-entity-incident",
		"incident_key":  "IR-PHASE9-I901-ENTITY",
		"title":         "Phase 9 entity-origin paste",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	existingHost := requireWorkbookCreate(t, harness, adminLogin, incidentID, entities.HostsViewSchemaID, map[string]any{
		"client_txn_id":     "txn-phase9-i-9-01-existing-host",
		"host.display_name": "Existing Gateway",
		"host.hostname":     "shared-gateway",
	})
	existingHostID := existingHost["row"].(map[string]any)["record_id"].(string)
	hostPaste := requireClipboardPaste(t, harness, adminLogin, incidentID, entities.HostsViewSchemaID, map[string]any{
		"view_schema_id":  entities.HostsViewSchemaID,
		"client_txn_id":   "txn-phase9-i-9-01-host-paste",
		"clipboard_text":  "Renamed Gateway\tshared-gateway\nNew Host\tnew-host",
		"format":          "tsv",
		"start_field_key": "host.display_name",
		"columns":         []string{"host.display_name", "host.hostname"},
		"targets":         []map[string]any{{"kind": "create"}, {"kind": "create"}},
	}, http.StatusOK)
	hostRows := hostPaste["rows"].([]any)
	if len(hostRows) != 2 {
		t.Fatalf("expected two entity-origin host rows, got %#v", hostRows)
	}
	if got := hostRows[0].(map[string]any)["record_id"]; got != existingHostID {
		t.Fatalf("host paste must reuse exact-match record_id: got %#v want %s rows=%#v", got, existingHostID, hostRows)
	}
	requireChangeSetSource(t, harness, hostPaste["change_set_id"].(string), "entities.hosts.clipboard_paste", "txn-phase9-i-9-01-host-paste")
	requireEntityOriginAndNoMentions(t, harness, hostRows[0].(map[string]any)["record_id"].(string), "hosts", "entity_sheet")

	identityPaste := requireClipboardPaste(t, harness, adminLogin, incidentID, entities.IdentitiesViewSchemaID, map[string]any{
		"view_schema_id":  entities.IdentitiesViewSchemaID,
		"client_txn_id":   "txn-phase9-i-9-01-identity-paste",
		"clipboard_text":  "Analyst One\tanalyst.one@example.test",
		"format":          "tsv",
		"start_field_key": "identity.display_name",
		"columns":         []string{"identity.display_name", "identity.email"},
		"targets":         []map[string]any{{"kind": "create"}},
	}, http.StatusOK)
	identityRows := identityPaste["rows"].([]any)
	if len(identityRows) != 1 {
		t.Fatalf("expected one identity-origin pasted row, got %#v", identityRows)
	}
	requireChangeSetSource(t, harness, identityPaste["change_set_id"].(string), "entities.identities.clipboard_paste", "txn-phase9-i-9-01-identity-paste")
	requireEntityOriginAndNoMentions(t, harness, identityRows[0].(map[string]any)["record_id"].(string), "identities", "entity_sheet")
}

func TestPhase9_I_9_01_BulkMutationsPersistOneVisibleBatch(t *testing.T) {
	harness := phase4test.StartServer(t, "phase9-i-9-01-bulk-mutations")
	adminLogin, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-i-9-01-bulk-incident",
		"incident_key":  "IR-PHASE9-I901-BULK",
		"title":         "Phase 9 bulk mutations",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	first := requireWorkbookCreate(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":    "txn-phase9-i-9-01-bulk-first",
		"timeline.summary": "Bulk first",
	})
	second := requireWorkbookCreate(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":    "txn-phase9-i-9-01-bulk-second",
		"timeline.summary": "Bulk second",
	})
	firstID := first["row"].(map[string]any)["record_id"].(string)
	secondID := second["row"].(map[string]any)["record_id"].(string)

	fill := requireBulkMutation(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID,
		"client_txn_id":  "txn-phase9-i-9-01-fill-down",
		"kind":           "fill_down_v1",
		"field_key":      "timeline.source_text",
		"value":          "Filled source text",
		"targets": []map[string]any{
			{"record_id": firstID, "base_row_version": 1},
			{"record_id": secondID, "base_row_version": 1},
		},
	})
	requireChangeSetSource(t, harness, fill["change_set_id"].(string), "workbook.bulk_mutations", "txn-phase9-i-9-01-fill-down")
	requireMutationTargets(t, harness, fill["change_set_id"].(string), []string{firstID, secondID})

	tag := requireBulkMutation(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID,
		"client_txn_id":  "txn-phase9-i-9-01-tag-bulk",
		"kind":           "multi_row_tag_assignment_v1",
		"tag_name":       "bulk-tag",
		"targets": []map[string]any{
			{"record_id": firstID, "base_row_version": 2},
			{"record_id": secondID, "base_row_version": 2},
		},
	})
	requireChangeSetSource(t, harness, tag["change_set_id"].(string), "workbook.bulk_mutations", "txn-phase9-i-9-01-tag-bulk")
	requireRecordTag(t, harness, firstID, "bulk-tag")
	requireRecordTag(t, harness, secondID, "bulk-tag")
}

func TestPhase9_I_9_01_ClipboardPasteAndBulkRejectCrossIncidentTargets(t *testing.T) {
	harness := phase4test.StartServer(t, "phase9-i-9-01-cross-incident-batch-targets")
	adminLogin, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incidentA := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-i-9-01-cross-incident-a",
		"incident_key":  "IR-PHASE9-I901-XA",
		"title":         "Phase 9 cross incident A",
	})
	incidentAID := phase4test.MustUUID(t, incidentA["incident_id"].(string))
	incidentB := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-i-9-01-cross-incident-b",
		"incident_key":  "IR-PHASE9-I901-XB",
		"title":         "Phase 9 cross incident B",
	})
	incidentBID := phase4test.MustUUID(t, incidentB["incident_id"].(string))

	local := requireWorkbookCreate(t, harness, adminLogin, incidentAID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":    "txn-phase9-i-9-01-cross-local",
		"timeline.summary": "Local batch target",
	})
	localID := phase4test.MustUUID(t, local["row"].(map[string]any)["record_id"].(string))
	foreign := requireWorkbookCreate(t, harness, adminLogin, incidentBID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":    "txn-phase9-i-9-01-cross-foreign",
		"timeline.summary": "Foreign batch target",
	})
	foreignID := phase4test.MustUUID(t, foreign["row"].(map[string]any)["record_id"].(string))

	for _, baseRowVersion := range []int{1, 99} {
		txnID := "txn-phase9-i-9-01-cross-paste"
		if baseRowVersion > 1 {
			txnID = "txn-phase9-i-9-01-cross-paste-future-version"
		}
		body := requireClipboardPaste(t, harness, adminLogin, incidentAID, timeline.TimelineViewSchemaID, map[string]any{
			"view_schema_id":  timeline.TimelineViewSchemaID,
			"client_txn_id":   txnID,
			"clipboard_text":  "Should not create\nShould not disclose",
			"format":          "tsv",
			"start_field_key": "timeline.summary",
			"columns":         []string{"timeline.summary"},
			"targets": []map[string]any{
				{"kind": "create"},
				{"kind": "record", "record_id": foreignID.String(), "base_row_version": baseRowVersion},
			},
		}, http.StatusNotFound)
		requireNoVersionOracle(t, body)
		requireNoTimelineSummary(t, harness, incidentAID, "Should not create")
		requireTimelineSummaryAndVersion(t, harness, foreignID, "Foreign batch target", 1)
		requireNoChangeSetForClientTxn(t, harness, txnID)
	}

	fillBody := requireBulkMutationStatus(t, harness, adminLogin, incidentAID, timeline.TimelineViewSchemaID, map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID,
		"client_txn_id":  "txn-phase9-i-9-01-cross-fill-down",
		"kind":           "fill_down_v1",
		"field_key":      "timeline.source_text",
		"value":          "Should not fill",
		"targets": []map[string]any{
			{"record_id": localID.String(), "base_row_version": 1},
			{"record_id": foreignID.String(), "base_row_version": 1},
		},
	}, http.StatusNotFound)
	requireNoVersionOracle(t, fillBody)
	requireTimelineSummaryAndVersion(t, harness, localID, "Local batch target", 1)
	requireTimelineSummaryAndVersion(t, harness, foreignID, "Foreign batch target", 1)
	requireNoTimelineSourceText(t, harness, localID)
	requireNoTimelineSourceText(t, harness, foreignID)
	requireNoChangeSetForClientTxn(t, harness, "txn-phase9-i-9-01-cross-fill-down")

	tagBody := requireBulkMutationStatus(t, harness, adminLogin, incidentAID, timeline.TimelineViewSchemaID, map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID,
		"client_txn_id":  "txn-phase9-i-9-01-cross-tag",
		"kind":           "multi_row_tag_assignment_v1",
		"tag_name":       "should-not-tag",
		"targets": []map[string]any{
			{"record_id": localID.String(), "base_row_version": 1},
			{"record_id": foreignID.String(), "base_row_version": 1},
		},
	}, http.StatusNotFound)
	requireNoVersionOracle(t, tagBody)
	requireNoRecordTag(t, harness, localID.String(), "should-not-tag")
	requireNoRecordTag(t, harness, foreignID.String(), "should-not-tag")
	requireNoChangeSetForClientTxn(t, harness, "txn-phase9-i-9-01-cross-tag")
}

func requireClipboardPaste(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any, wantStatus int) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/clipboard-paste",
		body,
		phase4test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	if wantStatus == http.StatusOK {
		return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	}
	if resp.StatusCode != wantStatus {
		t.Fatalf("clipboard paste status: got %d want %d body=%#v", resp.StatusCode, wantStatus, httptestx.ReadJSONBody(t, resp))
	}
	return httptestx.ReadJSONBody(t, resp)
}

func requireBulkMutation(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any) map[string]any {
	t.Helper()
	return requireBulkMutationStatus(t, harness, login, incidentID, viewSchemaID, body, http.StatusOK)
}

func requireBulkMutationStatus(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any, wantStatus int) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/bulk-mutations",
		body,
		phase4test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	if wantStatus == http.StatusOK {
		return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	}
	if resp.StatusCode != wantStatus {
		t.Fatalf("bulk mutation status: got %d want %d body=%#v", resp.StatusCode, wantStatus, httptestx.ReadJSONBody(t, resp))
	}
	return httptestx.ReadJSONBody(t, resp)
}

func requireChangeSetSource(t testing.TB, harness *phase4test.ServerHarness, changeSetID string, wantSource string, wantTxnID string) {
	t.Helper()
	var source, clientTxnID string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT source, client_txn_id
  FROM change_sets
 WHERE change_set_id = $1
`, changeSetID).Scan(&source, &clientTxnID); err != nil {
		t.Fatalf("query change_set source: %v", err)
	}
	if source != wantSource || clientTxnID != wantTxnID {
		t.Fatalf("unexpected change_set attribution: got source=%q txn=%q", source, clientTxnID)
	}
}

func requireMutationTargets(t testing.TB, harness *phase4test.ServerHarness, changeSetID string, want []string) {
	t.Helper()
	rows, err := harness.DB.QueryContext(context.Background(), `
SELECT target_id
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND target_kind = 'timeline_record'
 ORDER BY sequence_no ASC
`, changeSetID)
	if err != nil {
		t.Fatalf("query mutation targets: %v", err)
	}
	defer rows.Close()
	got := make([]string, 0)
	for rows.Next() {
		var target string
		if err := rows.Scan(&target); err != nil {
			t.Fatalf("scan mutation target: %v", err)
		}
		got = append(got, target)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate mutation targets: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("mutation target count: got %#v want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("mutation target order: got %#v want %#v", got, want)
		}
	}
}

func requirePasteConflict(t testing.TB, conflict map[string]any, recordID uuid.UUID, wantBase string, wantServer string, wantClient string) {
	t.Helper()
	if _, ok := timeline.ParseConflictToken(conflict["conflict_token"].(string)); !ok {
		t.Fatalf("paste conflict token is not parseable: %#v", conflict)
	}
	if conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != "timeline.summary" ||
		conflict["base_value"] != wantBase ||
		conflict["server_value"] != wantServer ||
		conflict["client_value"] != wantClient {
		t.Fatalf("unexpected paste conflict payload: got %#v want record=%s base=%q server=%q client=%q", conflict, recordID, wantBase, wantServer, wantClient)
	}
}

func requireRevisionCount(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID, want int) {
	t.Helper()
	var got int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, recordID).Scan(&got); err != nil {
		t.Fatalf("query revision count: %v", err)
	}
	if got != want {
		t.Fatalf("revision count for %s: got %d want %d", recordID, got, want)
	}
}

func requireRawCaptureValue(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID, want string) {
	t.Helper()
	var raw []byte
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT raw_capture FROM timeline_events WHERE record_id = $1`, recordID).Scan(&raw); err != nil {
		t.Fatalf("query raw_capture: %v", err)
	}
	var capture map[string]any
	if err := json.Unmarshal(raw, &capture); err != nil {
		t.Fatalf("decode raw_capture: %v", err)
	}
	columns, ok := capture["import_columns"].([]any)
	if !ok {
		t.Fatalf("raw_capture missing import_columns: %#v", capture)
	}
	for _, entry := range columns {
		object := entry.(map[string]any)
		if object["source_kind"] == "clipboard_paste" && object["raw_value"] == want {
			return
		}
	}
	t.Fatalf("raw_capture did not preserve %q: %#v", want, capture)
}

func requireMentionOriginKind(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID, fieldKey string, want string) {
	t.Helper()
	var got string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT origin_kind
  FROM entity_mentions
 WHERE source_record_id = $1
   AND source_field_key = $2
 ORDER BY ordinal ASC
 LIMIT 1
`, recordID, fieldKey).Scan(&got); err != nil {
		t.Fatalf("query mention origin kind: %v", err)
	}
	if got != want {
		t.Fatalf("mention origin kind: got %q want %q", got, want)
	}
}

func requireEntityOriginAndNoMentions(t testing.TB, harness *phase4test.ServerHarness, recordID string, tableName string, wantOrigin string) {
	t.Helper()
	query := ""
	switch tableName {
	case "hosts":
		query = `SELECT entity_origin FROM hosts WHERE record_id::text = $1`
	case "identities":
		query = `SELECT entity_origin FROM identities WHERE record_id::text = $1`
	default:
		t.Fatalf("unsupported entity table %q", tableName)
	}
	var got string
	if err := harness.DB.QueryRowContext(context.Background(), query, recordID).Scan(&got); err != nil {
		t.Fatalf("query %s entity_origin: %v", tableName, err)
	}
	if got != wantOrigin {
		t.Fatalf("entity-origin paste used wrong origin for %s: got %q want %q", recordID, got, wantOrigin)
	}
	if count := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id::text = $1`, recordID); count != 0 {
		t.Fatalf("entity-origin paste must not create mentions for %s, got %d", recordID, count)
	}
}

func requireRecordTag(t testing.TB, harness *phase4test.ServerHarness, recordID string, tagName string) {
	t.Helper()
	if count := phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE record_id::text = $1
   AND tag_name = $2
   AND deleted_at IS NULL
`, recordID, tagName); count != 1 {
		t.Fatalf("expected one active tag %q for %s, got %d", tagName, recordID, count)
	}
}

func requireNoRecordTag(t testing.TB, harness *phase4test.ServerHarness, recordID string, tagName string) {
	t.Helper()
	if count := phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE record_id::text = $1
   AND tag_name = $2
   AND deleted_at IS NULL
`, recordID, tagName); count != 0 {
		t.Fatalf("expected no active tag %q for %s, got %d", tagName, recordID, count)
	}
}

func requireNoVersionOracle(t testing.TB, body map[string]any) {
	t.Helper()
	errorValue, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error envelope, got %#v", body)
	}
	if errorValue["code"] != "incident_not_found" {
		t.Fatalf("unexpected cross-incident error code: got %v want incident_not_found in %#v", errorValue["code"], body)
	}
	details := httptestx.RequireErrorDetails(t, body)
	for _, key := range []string{"record_id", "base_row_version", "current_row_version", "conflicts", "rows"} {
		if _, ok := details[key]; ok {
			t.Fatalf("cross-incident target error leaked %s in details: %#v", key, details)
		}
	}
}

func requireTimelineSummaryAndVersion(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID, wantSummary string, wantVersion int64) {
	t.Helper()
	var summary string
	var rowVersion int64
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT e.summary, r.row_version
  FROM timeline_events e
  JOIN records r
    ON r.record_id = e.record_id
 WHERE e.record_id = $1
`, recordID).Scan(&summary, &rowVersion); err != nil {
		t.Fatalf("query timeline summary/version: %v", err)
	}
	if summary != wantSummary || rowVersion != wantVersion {
		t.Fatalf("timeline record %s changed: got summary=%q version=%d want summary=%q version=%d", recordID, summary, rowVersion, wantSummary, wantVersion)
	}
}

func requireNoTimelineSummary(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, summary string) {
	t.Helper()
	if count := phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM timeline_events
 WHERE incident_id = $1
   AND summary = $2
`, incidentID, summary); count != 0 {
		t.Fatalf("expected no timeline row summary %q in incident %s, got %d", summary, incidentID, count)
	}
}

func requireNoTimelineSourceText(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID) {
	t.Helper()
	var sourceText sql.NullString
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT source_text FROM timeline_events WHERE record_id = $1`, recordID).Scan(&sourceText); err != nil {
		t.Fatalf("query timeline source_text: %v", err)
	}
	if sourceText.Valid {
		t.Fatalf("expected no source_text for %s, got %q", recordID, sourceText.String)
	}
}

func requireNoChangeSetForClientTxn(t testing.TB, harness *phase4test.ServerHarness, clientTxnID string) {
	t.Helper()
	if count := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE client_txn_id = $1`, clientTxnID); count != 0 {
		t.Fatalf("expected no change_set for %s, got %d", clientTxnID, count)
	}
}

func resolveTimelineConflict(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, recordID uuid.UUID, token string, body map[string]any) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/conflicts/"+token+"/resolve",
		body,
		phase4test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}
