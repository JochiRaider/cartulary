package workbook_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestWorkbookClipboardFidelityAndLegacyReceipt_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook-clipboard-fidelity")
	login, actor := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{"client_txn_id": "fidelity-incident", "incident_key": "IR-FIDELITY", "title": "Clipboard fidelity"})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	for _, surface := range []struct{ view, field, route string }{
		{timeline.TimelineViewSchemaID, "timeline.raw_activity_text", "timeline.clipboard_paste"},
		{entitycontract.HostsViewSchemaID, "host.display_name", "entities.hosts.clipboard_paste"},
		{entitycontract.IdentitiesViewSchemaID, "identity.display_name", "entities.identities.clipboard_paste"},
	} {
		t.Run(surface.view, func(t *testing.T) {
			body := map[string]any{"view_schema_id": surface.view, "client_txn_id": "fidelity-valid-" + surface.view,
				"clipboard_text": "\"alpha, \"\"quote\"\"\"\r\n00123", "format": "csv", "header_mode": "none",
				"start_field_key": surface.field, "columns": []string{surface.field}, "targets": []map[string]any{{"kind": "create"}, {"kind": "create"}}}
			accepted := requireClipboardPaste(t, harness, login, incidentID, surface.view, body, http.StatusOK)
			rows := accepted["rows"].([]any)
			if len(rows) != 2 {
				t.Fatalf("lost geometry: %#v", rows)
			}
			requireCellValue(t, rows[0].(map[string]any), surface.field, "alpha, \"quote\"")
			requireCellValue(t, rows[1].(map[string]any), surface.field, "00123")
			if surface.view == timeline.TimelineViewSchemaID {
				clear := map[string]any{"view_schema_id": surface.view, "client_txn_id": "fidelity-clear",
					"clipboard_text": "\"\"\r\n\"\"", "format": "csv", "header_mode": "none",
					"start_field_key": surface.field, "columns": []string{surface.field},
					"targets": []map[string]any{
						{"kind": "record", "record_id": rows[0].(map[string]any)["record_id"], "base_row_version": 1},
						{"kind": "record", "record_id": rows[1].(map[string]any)["record_id"], "base_row_version": 1}}}
				cleared := requireClipboardPaste(t, harness, login, incidentID, surface.view, clear, http.StatusOK)
				for _, row := range cleared["rows"].([]any) {
					requireCellValue(t, row.(map[string]any), surface.field, "")
				}
			}
			var beforeRecords, beforeRevisions int
			if err := harness.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM records WHERE incident_id=$1`, incidentID).Scan(&beforeRecords); err != nil {
				t.Fatal(err)
			}
			if err := harness.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM record_revisions`).Scan(&beforeRevisions); err != nil {
				t.Fatal(err)
			}
			invalidInputs := []string{"\"unterminated", "a\tb\nshort", "one", "first\n\"bad\u0001control\"", strings.Repeat("x", 8_388_609)}
			if surface.view != timeline.TimelineViewSchemaID {
				invalidInputs = append(invalidInputs, "\"\"\n\"\"")
			}
			for index, invalid := range invalidInputs {
				txn := fmt.Sprintf("fidelity-invalid-%s-%d", surface.view, index)
				body["client_txn_id"], body["clipboard_text"], body["format"] = txn, invalid, "tsv"
				requireClipboardPaste(t, harness, login, incidentID, surface.view, body, http.StatusBadRequest)
				requireNoChangeSetForClientTxn(t, harness, txn)
			}
			var afterRecords, afterRevisions int
			if err := harness.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM records WHERE incident_id=$1`, incidentID).Scan(&afterRecords); err != nil {
				t.Fatal(err)
			}
			if err := harness.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM record_revisions`).Scan(&afterRevisions); err != nil {
				t.Fatal(err)
			}
			if beforeRecords != afterRecords || beforeRevisions != afterRevisions {
				t.Fatal("rejected input wrote records or history")
			}
			// An authored legacy receipt fixture: this text could be admitted by the
			// previous permissive/auto parser. New parsing must not gate retrieval.
			legacyTxn := "fidelity-legacy-" + surface.view
			body["client_txn_id"], body["clipboard_text"], body["format"] = legacyTxn, "legacy\n\nrow", "auto"
			delete(body, "header_mode")
			identity := map[string]any{"view_schema_id": surface.view, "clipboard_text": body["clipboard_text"], "format": "auto", "start_field_key": surface.field, "columns": body["columns"]}
			if surface.view == timeline.TimelineViewSchemaID {
				identity["targets"] = body["targets"]
			}
			identityBytes, err := json.Marshal(identity)
			if err != nil {
				t.Fatal(err)
			}
			hash := sha256.Sum256(identityBytes)
			receipt, err := json.Marshal(accepted)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := harness.DB.ExecContext(t.Context(), `INSERT INTO route_idempotency(route_key,scope_key,client_txn_id,actor_user_id,request_hash,status_code,response_json) VALUES($1,$2,$3,$4,$5,200,$6)`, surface.route, incidentID.String()+":"+surface.view, legacyTxn, actor, hash[:], receipt); err != nil {
				t.Fatal(err)
			}
			replayed := requireClipboardPaste(t, harness, login, incidentID, surface.view, body, http.StatusOK)
			if !reflect.DeepEqual(accepted, replayed) {
				t.Fatalf("legacy replay changed receipt: %#v", replayed)
			}
			requireNoChangeSetForClientTxn(t, harness, legacyTxn)
			if surface.view != timeline.TimelineViewSchemaID {
				body["targets"] = []map[string]any{{"kind": "create"}}
				requireClipboardPaste(t, harness, login, incidentID, surface.view, body, http.StatusBadRequest)
				body["targets"] = []map[string]any{{"kind": "create"}, {"kind": "create"}}
			}
			body["header_mode"] = "none"
			requireClipboardPaste(t, harness, login, incidentID, surface.view, body, http.StatusConflict)
		})
	}
}

func TestTimelineClipboardPastePersistsOrderedMutationsAndConflicts_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook_interaction-i-9-01-clipboard-paste")
	adminLogin, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook_interaction-i-9-01-incident",
		"incident_key":  "IR-WORKBOOK-INTERACTION-I901",
		"title":         "Workbook inspector clipboard paste",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	existing := requireWorkbookCreate(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":                   "txn-workbook_interaction-i-9-01-existing",
		"timeline.activity_synopsis_text": "Existing base",
	})
	existingRow := existing["row"].(map[string]any)
	existingID := appsupport.MustUUID(t, existingRow["record_id"].(string))

	pasteData := requireClipboardPaste(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"view_schema_id":  timeline.TimelineViewSchemaID,
		"client_txn_id":   "txn-workbook_interaction-i-9-01-paste",
		"clipboard_text":  "Updated existing\tDetails from paste\tgateway-one\tpreserve-me\nCreated from paste\tCreated details\tgateway-two\tpreserve-new",
		"format":          "tsv",
		"start_field_key": "timeline.activity_synopsis_text",
		"columns": []string{
			"timeline.activity_synopsis_text",
			"timeline.raw_activity_text",
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
	createdID := appsupport.MustUUID(t, createdRow["record_id"].(string))
	requireCellValue(t, patchedRow, "timeline.activity_synopsis_text", "Updated existing")
	requireCellValue(t, createdRow, "timeline.activity_synopsis_text", "Created from paste")

	requireChangeSetSource(t, harness, changeSetID, "timeline.clipboard_paste", "txn-workbook_interaction-i-9-01-paste")
	requireMutationTargets(t, harness, changeSetID, []string{existingID.String(), createdID.String()})
	requireRevisionCount(t, harness, existingID, 2)
	requireRevisionCount(t, harness, createdID, 1)
	requireProvenanceValue(t, harness, existingID, "preserve-me")
	requireProvenanceValue(t, harness, createdID, "preserve-new")
	requireMentionOriginKind(t, harness, existingID, "timeline.host_refs", "clipboard_paste")
	requireMentionOriginKind(t, harness, createdID, "timeline.host_refs", "clipboard_paste")

	firstConflictBase := requireWorkbookCreate(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":                   "txn-workbook_interaction-i-9-01-conflict-base-first",
		"timeline.activity_synopsis_text": "Conflict base first",
	})
	firstConflictRow := firstConflictBase["row"].(map[string]any)
	firstConflictID := appsupport.MustUUID(t, firstConflictRow["record_id"].(string))
	secondConflictBase := requireWorkbookCreate(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":                   "txn-workbook_interaction-i-9-01-conflict-base-second",
		"timeline.activity_synopsis_text": "Conflict base second",
	})
	secondConflictRow := secondConflictBase["row"].(map[string]any)
	secondConflictID := appsupport.MustUUID(t, secondConflictRow["record_id"].(string))
	requireWorkbookPatch(t, harness, adminLogin, firstConflictID, map[string]any{
		"view_schema_id":   timeline.TimelineViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook_interaction-i-9-01-server-update-first",
		"changes": []map[string]any{{
			"field_key": "timeline.activity_synopsis_text",
			"value":     "Server value first",
		}},
	})
	requireWorkbookPatch(t, harness, adminLogin, secondConflictID, map[string]any{
		"view_schema_id":   timeline.TimelineViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook_interaction-i-9-01-server-update-second",
		"changes": []map[string]any{{
			"field_key": "timeline.activity_synopsis_text",
			"value":     "Server value second",
		}},
	})

	partial := requireClipboardPaste(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"view_schema_id":  timeline.TimelineViewSchemaID,
		"client_txn_id":   "txn-workbook_interaction-i-9-01-partial-conflict",
		"clipboard_text":  "Client conflicting summary first\nClient conflicting summary second\nNon-conflicting created",
		"format":          "tsv",
		"start_field_key": "timeline.activity_synopsis_text",
		"columns":         []string{"timeline.activity_synopsis_text"},
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
	partialCreatedID := appsupport.MustUUID(t, partialRows[0].(map[string]any)["record_id"].(string))
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
		"client_txn_id":   "txn-workbook_interaction-i-9-01-resolve-conflict-first",
		"resolved_value":  "Client conflicting summary first",
	})
	firstResolveChangeSetID := firstResolveData["change_set_id"].(string)
	if firstResolveChangeSetID == partial["change_set_id"].(string) {
		t.Fatalf("conflict resolution must create a separate change_set: paste=%s resolve=%s", partial["change_set_id"], firstResolveChangeSetID)
	}
	requireChangeSetSource(t, harness, firstResolveChangeSetID, "timeline.records.conflicts.resolve", "txn-workbook_interaction-i-9-01-resolve-conflict-first")
	secondResolveData := resolveTimelineConflict(t, harness, adminLogin, secondConflictID, secondConflict["conflict_token"].(string), map[string]any{
		"conflict_token":  secondConflict["conflict_token"].(string),
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-workbook_interaction-i-9-01-resolve-conflict-second",
		"resolved_value":  "Client conflicting summary second",
	})
	secondResolveChangeSetID := secondResolveData["change_set_id"].(string)
	if secondResolveChangeSetID == partial["change_set_id"].(string) || secondResolveChangeSetID == firstResolveChangeSetID {
		t.Fatalf("each conflict resolution must create its own change_set: paste=%s first=%s second=%s", partial["change_set_id"], firstResolveChangeSetID, secondResolveChangeSetID)
	}
	requireChangeSetSource(t, harness, secondResolveChangeSetID, "timeline.records.conflicts.resolve", "txn-workbook_interaction-i-9-01-resolve-conflict-second")
}

func TestEntityOriginClipboardPasteUsesSharedIngest_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook_interaction-i-9-01-entity-origin-paste")
	adminLogin, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook_interaction-i-9-01-entity-incident",
		"incident_key":  "IR-WORKBOOK-INTERACTION-I901-ENTITY",
		"title":         "Workbook inspector entity-origin paste",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	existingHost := requireWorkbookCreate(t, harness, adminLogin, incidentID, entitycontract.HostsViewSchemaID, map[string]any{
		"client_txn_id":     "txn-workbook_interaction-i-9-01-existing-host",
		"host.display_name": "Existing Gateway",
		"host.hostname":     "shared-gateway",
	})
	existingHostID := existingHost["row"].(map[string]any)["record_id"].(string)
	hostPaste := requireClipboardPaste(t, harness, adminLogin, incidentID, entitycontract.HostsViewSchemaID, map[string]any{
		"view_schema_id":  entitycontract.HostsViewSchemaID,
		"client_txn_id":   "txn-workbook_interaction-i-9-01-host-paste",
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
	requireChangeSetSource(t, harness, hostPaste["change_set_id"].(string), "entities.hosts.clipboard_paste", "txn-workbook_interaction-i-9-01-host-paste")
	requireEntityOriginAndNoMentions(t, harness, hostRows[0].(map[string]any)["record_id"].(string), "hosts", "entity_sheet")

	identityPaste := requireClipboardPaste(t, harness, adminLogin, incidentID, entitycontract.IdentitiesViewSchemaID, map[string]any{
		"view_schema_id":  entitycontract.IdentitiesViewSchemaID,
		"client_txn_id":   "txn-workbook_interaction-i-9-01-identity-paste",
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
	requireChangeSetSource(t, harness, identityPaste["change_set_id"].(string), "entities.identities.clipboard_paste", "txn-workbook_interaction-i-9-01-identity-paste")
	requireEntityOriginAndNoMentions(t, harness, identityRows[0].(map[string]any)["record_id"].(string), "identities", "entity_sheet")

	// Two source rows may reuse one entity. One paste still owns only one
	// row revision for that record, and exact replay returns its original receipt.
	repeatedRequest := map[string]any{
		"view_schema_id":  entitycontract.HostsViewSchemaID,
		"client_txn_id":   "txn-repeated-host-paste",
		"clipboard_text":  "First name\tshared-gateway\nFinal name\tshared-gateway",
		"format":          "tsv",
		"start_field_key": "host.display_name",
		"columns":         []string{"host.display_name", "host.hostname"},
		"targets":         []map[string]any{{"kind": "create"}, {"kind": "create"}},
	}
	repeated := requireClipboardPaste(t, harness, adminLogin, incidentID, entitycontract.HostsViewSchemaID, repeatedRequest, http.StatusOK)
	var revisionCount int
	if err := harness.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM record_revisions WHERE change_set_id = $1 AND record_id = $2`, repeated["change_set_id"], existingHostID).Scan(&revisionCount); err != nil {
		t.Fatal(err)
	}
	if revisionCount != 1 {
		t.Fatalf("one paste must create one revision per reused entity, got %d", revisionCount)
	}
	replayed := requireClipboardPaste(t, harness, adminLogin, incidentID, entitycontract.HostsViewSchemaID, repeatedRequest, http.StatusOK)
	if replayed["change_set_id"] != repeated["change_set_id"] {
		t.Fatalf("replay created another change set")
	}
	repeatedRows := repeated["rows"].([]any)
	for _, value := range repeatedRows {
		row := value.(map[string]any)
		if row["record_id"] != existingHostID || row["row_version"] != hostRows[0].(map[string]any)["row_version"].(float64)+1 {
			t.Fatalf("repeated reuse must return the final version once advanced: %#v", repeatedRows)
		}
	}
	requireChangeSetSource(t, harness, repeated["change_set_id"].(string), "entities.hosts.clipboard_paste", "txn-repeated-host-paste")
	for _, entityCase := range []struct{ view, display, identifier, text string }{
		{entitycontract.IdentitiesViewSchemaID, "identity.display_name", "identity.email", "First identity\tanalyst.one@example.test\nFinal identity\tanalyst.one@example.test"},
		{entitycontract.HostsViewSchemaID, "host.display_name", "host.hostname", "First new host\tnew-repeated-host\nFinal new host\tnew-repeated-host"},
	} {
		request := map[string]any{"view_schema_id": entityCase.view, "client_txn_id": "txn-repeated-" + entityCase.identifier, "clipboard_text": entityCase.text, "format": "tsv", "start_field_key": entityCase.display, "columns": []string{entityCase.display, entityCase.identifier}, "targets": []map[string]any{{"kind": "create"}, {"kind": "create"}}}
		result := requireClipboardPaste(t, harness, adminLogin, incidentID, entityCase.view, request, http.StatusOK)
		rows := result["rows"].([]any)
		first, last := rows[0].(map[string]any), rows[1].(map[string]any)
		if first["record_id"] != last["record_id"] || first["row_version"] != last["row_version"] {
			t.Fatalf("repeated identity must return one committed record/version: %#v", rows)
		}
		if err := harness.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM record_revisions WHERE change_set_id = $1 AND record_id = $2`, result["change_set_id"], first["record_id"]).Scan(&revisionCount); err != nil {
			t.Fatal(err)
		}
		if revisionCount != 1 {
			t.Fatalf("expected one final revision, got %d", revisionCount)
		}
		var mutationCount int
		if err := harness.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM change_set_mutations WHERE change_set_id = $1 AND target_id = $2`, result["change_set_id"], first["record_id"]).Scan(&mutationCount); err != nil {
			t.Fatal(err)
		}
		if mutationCount != 2 {
			t.Fatalf("ordered source mutations lost: %d", mutationCount)
		}
		replayed := requireClipboardPaste(t, harness, adminLogin, incidentID, entityCase.view, request, http.StatusOK)
		if replayed["change_set_id"] != result["change_set_id"] {
			t.Fatal("replay changed attribution")
		}

		recordID := appsupport.MustUUID(t, first["record_id"].(string))
		requireWorkbookPatch(t, harness, adminLogin, recordID, map[string]any{
			"view_schema_id": entityCase.view, "client_txn_id": "txn-after-" + entityCase.identifier,
			"base_row_version": first["row_version"], "changes": []map[string]any{{"field_key": entityCase.display, "value": "Intervening entity name"}},
		})
		replayed = requireClipboardPaste(t, harness, adminLogin, incidentID, entityCase.view, request, http.StatusOK)
		if !reflect.DeepEqual(replayed, result) {
			t.Fatalf("replay reconstructed a historical receipt: %#v", replayed)
		}
		var currentVersion int64
		if err := harness.DB.QueryRowContext(t.Context(), `SELECT row_version FROM records WHERE record_id = $1`, recordID).Scan(&currentVersion); err != nil {
			t.Fatal(err)
		}
		if currentVersion != int64(first["row_version"].(float64))+1 {
			t.Fatalf("replay advanced entity revision: %d", currentVersion)
		}
	}

}

func TestBulkMutationsPersistOneVisibleBatch_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook_interaction-i-9-01-bulk-mutations")
	adminLogin, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook_interaction-i-9-01-bulk-incident",
		"incident_key":  "IR-WORKBOOK-INTERACTION-I901-BULK",
		"title":         "Workbook inspector bulk mutations",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	first := requireWorkbookCreate(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":                   "txn-workbook_interaction-i-9-01-bulk-first",
		"timeline.activity_synopsis_text": "Bulk first",
	})
	second := requireWorkbookCreate(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":                   "txn-workbook_interaction-i-9-01-bulk-second",
		"timeline.activity_synopsis_text": "Bulk second",
	})
	firstID := first["row"].(map[string]any)["record_id"].(string)
	secondID := second["row"].(map[string]any)["record_id"].(string)

	fill := requireBulkMutation(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID,
		"client_txn_id":  "txn-workbook_interaction-i-9-01-fill-down",
		"kind":           "fill_down_v1",
		"field_key":      "timeline.raw_activity_text",
		"value":          "Filled source text",
		"targets": []map[string]any{
			{"record_id": firstID, "base_row_version": 1},
			{"record_id": secondID, "base_row_version": 1},
		},
	})
	requireChangeSetSource(t, harness, fill["change_set_id"].(string), "workbook.bulk_mutations", "txn-workbook_interaction-i-9-01-fill-down")
	requireMutationTargets(t, harness, fill["change_set_id"].(string), []string{firstID, secondID})

	tag := requireBulkMutation(t, harness, adminLogin, incidentID, timeline.TimelineViewSchemaID, map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID,
		"client_txn_id":  "txn-workbook_interaction-i-9-01-tag-bulk",
		"kind":           "multi_row_tag_assignment_v1",
		"tag_name":       "bulk-tag",
		"targets": []map[string]any{
			{"record_id": firstID, "base_row_version": 2},
			{"record_id": secondID, "base_row_version": 2},
		},
	})
	requireChangeSetSource(t, harness, tag["change_set_id"].(string), "workbook.bulk_mutations", "txn-workbook_interaction-i-9-01-tag-bulk")
	requireRecordTag(t, harness, firstID, "bulk-tag")
	requireRecordTag(t, harness, secondID, "bulk-tag")
}

func TestClipboardPasteAndBulkRejectCrossIncidentTargets_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook_interaction-i-9-01-cross-incident-batch-targets")
	adminLogin, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incidentA := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook_interaction-i-9-01-cross-incident-a",
		"incident_key":  "IR-WORKBOOK-INTERACTION-I901-XA",
		"title":         "Workbook inspector cross incident A",
	})
	incidentAID := appsupport.MustUUID(t, incidentA["incident_id"].(string))
	incidentB := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook_interaction-i-9-01-cross-incident-b",
		"incident_key":  "IR-WORKBOOK-INTERACTION-I901-XB",
		"title":         "Workbook inspector cross incident B",
	})
	incidentBID := appsupport.MustUUID(t, incidentB["incident_id"].(string))

	local := requireWorkbookCreate(t, harness, adminLogin, incidentAID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":                   "txn-workbook_interaction-i-9-01-cross-local",
		"timeline.activity_synopsis_text": "Local batch target",
	})
	localID := appsupport.MustUUID(t, local["row"].(map[string]any)["record_id"].(string))
	foreign := requireWorkbookCreate(t, harness, adminLogin, incidentBID, timeline.TimelineViewSchemaID, map[string]any{
		"client_txn_id":                   "txn-workbook_interaction-i-9-01-cross-foreign",
		"timeline.activity_synopsis_text": "Foreign batch target",
	})
	foreignID := appsupport.MustUUID(t, foreign["row"].(map[string]any)["record_id"].(string))

	for _, baseRowVersion := range []int{1, 99} {
		txnID := "txn-workbook_interaction-i-9-01-cross-paste"
		if baseRowVersion > 1 {
			txnID = "txn-workbook_interaction-i-9-01-cross-paste-future-version"
		}
		body := requireClipboardPaste(t, harness, adminLogin, incidentAID, timeline.TimelineViewSchemaID, map[string]any{
			"view_schema_id":  timeline.TimelineViewSchemaID,
			"client_txn_id":   txnID,
			"clipboard_text":  "Should not create\nShould not disclose",
			"format":          "tsv",
			"start_field_key": "timeline.activity_synopsis_text",
			"columns":         []string{"timeline.activity_synopsis_text"},
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
		"client_txn_id":  "txn-workbook_interaction-i-9-01-cross-fill-down",
		"kind":           "fill_down_v1",
		"field_key":      "timeline.raw_activity_text",
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
	requireNoChangeSetForClientTxn(t, harness, "txn-workbook_interaction-i-9-01-cross-fill-down")

	tagBody := requireBulkMutationStatus(t, harness, adminLogin, incidentAID, timeline.TimelineViewSchemaID, map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID,
		"client_txn_id":  "txn-workbook_interaction-i-9-01-cross-tag",
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
	requireNoChangeSetForClientTxn(t, harness, "txn-workbook_interaction-i-9-01-cross-tag")
}

func TestWorkbookBatchAdmissionHasNoPartialEffects_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook-batch-admission-recovery")
	login, actor := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{"client_txn_id": "batch-admission-incident", "incident_key": "IR-BATCH-ADMISSION", "title": "Batch admission"})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	viewer := authflowtest.SeedLocalUserRecord(t, harness.DB, "batch-viewer@example.test", "Batch Viewer", "BatchViewerPass1!", false, false, true)
	incidentstoretest.SeedMembership(t, harness.DB, incidentID, viewer.ID, viewer.DisplayName, "viewer", actor)
	viewerLogin := LoginLocalUserNoMFA(t, harness, viewer.Email, "BatchViewerPass1!")
	local := requireWorkbookCreate(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"client_txn_id": "batch-admission-local", "timeline.activity_synopsis_text": "Unchanged local"})
	localID := appsupport.MustUUID(t, local["row"].(map[string]any)["record_id"].(string))
	host := requireWorkbookCreate(t, harness, login, incidentID, entitycontract.HostsViewSchemaID, map[string]any{"client_txn_id": "batch-admission-host", "host.display_name": "Wrong target type"})
	hostID := appsupport.MustUUID(t, host["row"].(map[string]any)["record_id"].(string))
	deleted := requireWorkbookCreate(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"client_txn_id": "batch-admission-deleted", "timeline.activity_synopsis_text": "Deleted target"})
	deletedID := appsupport.MustUUID(t, deleted["row"].(map[string]any)["record_id"].(string))
	httptestx.RequireSuccessEnvelope(t, deleteRecordViaWorkbookRoute(t, harness, login, deletedID, map[string]any{"client_txn_id": "batch-admission-delete", "base_row_version": 1}), http.StatusOK)
	for _, invalid := range []struct {
		name string
		id   uuid.UUID
	}{{"missing", uuid.New()}, {"deleted", deletedID}, {"wrong-type", hostID}} {
		for _, kind := range []string{"paste", "fill_down_v1", "multi_row_tag_assignment_v1"} {
			t.Run(invalid.name+"/"+kind, func(t *testing.T) {
				txn := "batch-invalid-" + invalid.name + "-" + kind
				body := map[string]any{"view_schema_id": timeline.TimelineViewSchemaID, "client_txn_id": txn}
				targets := []map[string]any{{"record_id": localID.String(), "base_row_version": 1}, {"record_id": invalid.id.String(), "base_row_version": 99}}
				var result map[string]any
				if kind == "paste" {
					targets[0]["kind"], targets[1]["kind"] = "record", "record"
					body["targets"], body["clipboard_text"], body["format"] = targets, "Forbidden first\nForbidden second", "tsv"
					body["start_field_key"], body["columns"] = "timeline.activity_synopsis_text", []string{"timeline.activity_synopsis_text"}
					result = requireClipboardPaste(t, harness, login, incidentID, timeline.TimelineViewSchemaID, body, http.StatusNotFound)
				} else {
					body["targets"], body["kind"] = targets, kind
					if kind == "fill_down_v1" {
						body["field_key"], body["value"] = "timeline.raw_activity_text", "Forbidden fill"
					} else {
						body["tag_name"] = "forbidden-tag"
					}
					result = requireBulkMutationStatus(t, harness, login, incidentID, timeline.TimelineViewSchemaID, body, http.StatusNotFound)
				}
				requireNoVersionOracle(t, result)
				requireNoChangeSetForClientTxn(t, harness, txn)
				requireTimelineSummaryAndVersion(t, harness, localID, "Unchanged local", 1)
				requireNoRecordTag(t, harness, localID.String(), "forbidden-tag")
			})
		}
	}
	// Authorization precedes parsing, target lookup, limits, and receipt lookup.
	// Entity-origin targets are create/upsert intents, never client record IDs.
	for _, action := range []struct{ view, field, kind string }{
		{timeline.TimelineViewSchemaID, "timeline.activity_synopsis_text", "paste"},
		{entitycontract.HostsViewSchemaID, "host.display_name", "paste"},
		{entitycontract.IdentitiesViewSchemaID, "identity.display_name", "paste"},
		{timeline.TimelineViewSchemaID, "timeline.raw_activity_text", "fill_down_v1"},
		{timeline.TimelineViewSchemaID, "timeline.tags", "multi_row_tag_assignment_v1"},
	} {
		for _, mode := range []string{"excessive", "wrong-view"} {
			txn := fmt.Sprintf("batch-%s-%s-%s", mode, action.view, action.kind)
			targets := make([]map[string]any, 501)
			for i := range targets {
				targets[i] = map[string]any{"record_id": uuid.NewString(), "base_row_version": 1}
			}
			body := map[string]any{"view_schema_id": action.view, "client_txn_id": txn, "targets": targets}
			if action.kind == "paste" {
				for i := range targets {
					targets[i] = map[string]any{"kind": "create"}
				}
				body["clipboard_text"], body["format"], body["start_field_key"], body["columns"] = strings.TrimSuffix(strings.Repeat("Forbidden row\n", 501), "\n"), "tsv", action.field, []string{action.field}
			} else {
				body["kind"] = action.kind
				if action.kind == "fill_down_v1" {
					body["field_key"], body["value"] = action.field, "Forbidden fill"
				} else {
					body["tag_name"] = "forbidden-tag"
				}
			}
			if mode == "wrong-view" {
				body["view_schema_id"] = "cartulary.view.notes.v1"
			}
			for _, role := range []struct {
				login  appsupport.LoginResult
				status int
			}{{viewerLogin, http.StatusForbidden}, {login, http.StatusBadRequest}} {
				var result map[string]any
				if action.kind == "paste" {
					result = requireClipboardPaste(t, harness, role.login, incidentID, action.view, body, role.status)
				} else {
					result = requireBulkMutationStatus(t, harness, role.login, incidentID, action.view, body, role.status)
				}
				errorBody := result["error"].(map[string]any)
				expectedCode := "invalid_mutation_payload"
				if role.status == http.StatusForbidden {
					expectedCode = "authorization_denied"
				}
				if errorBody["code"] != expectedCode {
					t.Fatalf("unexpected admission error: %#v", result)
				}
				for _, key := range []string{"rows", "conflicts", "current_row_version", "server_value"} {
					if _, present := result[key]; present {
						t.Fatalf("protected batch member %s disclosed", key)
					}
					if details, ok := errorBody["details"].(map[string]any); ok {
						if _, present := details[key]; present {
							t.Fatalf("protected error member %s disclosed", key)
						}
					}
				}
				requireNoChangeSetForClientTxn(t, harness, txn)
			}
		}
	}
	requireNoTimelineSummary(t, harness, incidentID, "Forbidden row")
	requireTimelineSummaryAndVersion(t, harness, localID, "Unchanged local", 1)
}

func requireClipboardPaste(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any, wantStatus int) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/clipboard-paste",
		body,
		appsupport.WithCookies(login.SessionCookie, login.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	if wantStatus == http.StatusOK {
		return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	}
	if resp.StatusCode != wantStatus {
		t.Fatalf("clipboard paste status: got %d want %d body=%#v", resp.StatusCode, wantStatus, httptestx.ReadJSONBody(t, resp))
	}
	return httptestx.ReadJSONBody(t, resp)
}

func requireBulkMutation(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any) map[string]any {
	t.Helper()
	return requireBulkMutationStatus(t, harness, login, incidentID, viewSchemaID, body, http.StatusOK)
}

func requireBulkMutationStatus(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any, wantStatus int) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/bulk-mutations",
		body,
		appsupport.WithCookies(login.SessionCookie, login.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	if wantStatus == http.StatusOK {
		return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	}
	if resp.StatusCode != wantStatus {
		t.Fatalf("bulk mutation status: got %d want %d body=%#v", resp.StatusCode, wantStatus, httptestx.ReadJSONBody(t, resp))
	}
	return httptestx.ReadJSONBody(t, resp)
}

func requireChangeSetSource(t testing.TB, harness *appsupport.ServerHarness, changeSetID string, wantSource string, wantTxnID string) {
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

func requireMutationTargets(t testing.TB, harness *appsupport.ServerHarness, changeSetID string, want []string) {
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
	if token, ok := conflict["conflict_token"].(string); !ok || token == "" {
		t.Fatalf("paste conflict token is empty: %#v", conflict)
	}
	if conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != "timeline.activity_synopsis_text" ||
		conflict["base_value"] != wantBase ||
		conflict["server_value"] != wantServer ||
		conflict["client_value"] != wantClient {
		t.Fatalf("unexpected paste conflict payload: got %#v want record=%s base=%q server=%q client=%q", conflict, recordID, wantBase, wantServer, wantClient)
	}
}

func requireRevisionCount(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID, want int) {
	t.Helper()
	var got int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, recordID).Scan(&got); err != nil {
		t.Fatalf("query revision count: %v", err)
	}
	if got != want {
		t.Fatalf("revision count for %s: got %d want %d", recordID, got, want)
	}
}

func requireProvenanceValue(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID, want string) {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM timeline_source_provenance
 WHERE record_id = $1
   AND source_kind = 'clipboard_paste'
   AND raw_value = $2
`, recordID, want).Scan(&count); err != nil {
		t.Fatalf("query timeline source provenance: %v", err)
	}
	if count != 1 {
		t.Fatalf("timeline source provenance count for %q: got %d want 1", want, count)
	}
}

func requireMentionOriginKind(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID, fieldKey string, want string) {
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

func requireEntityOriginAndNoMentions(t testing.TB, harness *appsupport.ServerHarness, recordID string, tableName string, wantOrigin string) {
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
	if count := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id::text = $1`, recordID); count != 0 {
		t.Fatalf("entity-origin paste must not create mentions for %s, got %d", recordID, count)
	}
}

func requireRecordTag(t testing.TB, harness *appsupport.ServerHarness, recordID string, tagName string) {
	t.Helper()
	if count := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE record_id::text = $1
   AND tag_name = $2
   AND deleted_at IS NULL
`, recordID, tagName); count != 1 {
		t.Fatalf("expected one active tag %q for %s, got %d", tagName, recordID, count)
	}
}

func requireNoRecordTag(t testing.TB, harness *appsupport.ServerHarness, recordID string, tagName string) {
	t.Helper()
	if count := appsupport.QueryCount(t, harness.DB, `
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

func requireTimelineSummaryAndVersion(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID, wantSummary string, wantVersion int64) {
	t.Helper()
	var summary string
	var rowVersion int64
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT e.activity_synopsis_text, r.row_version
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

func requireNoTimelineSummary(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID, summary string) {
	t.Helper()
	if count := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM timeline_events
 WHERE incident_id = $1
   AND activity_synopsis_text = $2
`, incidentID, summary); count != 0 {
		t.Fatalf("expected no timeline row summary %q in incident %s, got %d", summary, incidentID, count)
	}
}

func requireNoTimelineSourceText(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID) {
	t.Helper()
	var sourceText sql.NullString
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT raw_activity_text FROM timeline_events WHERE record_id = $1`, recordID).Scan(&sourceText); err != nil {
		t.Fatalf("query timeline raw_activity_text: %v", err)
	}
	if sourceText.Valid {
		t.Fatalf("expected no raw_activity_text for %s, got %q", recordID, sourceText.String)
	}
}

func requireNoChangeSetForClientTxn(t testing.TB, harness *appsupport.ServerHarness, clientTxnID string) {
	t.Helper()
	if count := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE client_txn_id = $1`, clientTxnID); count != 0 {
		t.Fatalf("expected no change_set for %s, got %d", clientTxnID, count)
	}
}

func resolveTimelineConflict(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, token string, body map[string]any) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/conflicts/"+token+"/resolve",
		body,
		appsupport.WithCookies(login.SessionCookie, login.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}
