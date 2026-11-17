package workbook_test

import (
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

type collectionReviewInput struct {
	create, serverActions, clientActions map[string]any
	baseValue, serverValue, clientValue  string
}

func TestCollectionReviewRouteResolve_Integration(t *testing.T) {
	exerciseHostIdentityCollectionConflicts(t)
	harness, login, actorID, incidentID := ConflictFixture(t, "workbook-collection-review", "IR-WORKBOOK-COLLECTION-REVIEW")
	// Each closure owns its subject's complete payload and reference arrangement.
	// No case consumes a record version or subject mutated by another case.
	cases := []struct {
		fieldKey, viewSchemaID, itemKind, valueKey string
		arrange                                    func(t *testing.T) collectionReviewInput
	}{
		{fieldKey: "note.tags", viewSchemaID: "cartulary.view.notes.v1", itemKind: "tag", valueKey: "display_text", arrange: func(t *testing.T) collectionReviewInput {
			base := "base-tag"
			server := "server-tag"
			client := "client-tag"
			return collectionReviewInput{
				create:        map[string]any{"note.title": "Collection note", "note.tags": collectionActions(addToken(base))},
				serverActions: collectionActions(addToken(server)), clientActions: collectionActions(addToken(client)),
				baseValue: base, serverValue: server, clientValue: client,
			}
		}},
		{fieldKey: "comm_log.decision_ids", viewSchemaID: "cartulary.view.comm_log.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedDecisionRecord(t, harness, incidentID, actorID, "base reference")
			server := seedDecisionRecord(t, harness, incidentID, actorID, "server reference")
			client := seedDecisionRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"comm_log.comm_type": "briefing", "comm_log.audience": "leadership", "comm_log.channel_or_meeting": "Bridge", "comm_log.summary": "Coordination update", "comm_log.decision_ids": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "comm_log.action_task_ids", viewSchemaID: "cartulary.view.comm_log.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedTaskRecord(t, harness, incidentID, actorID, "base reference")
			server := seedTaskRecord(t, harness, incidentID, actorID, "server reference")
			client := seedTaskRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"comm_log.comm_type": "briefing", "comm_log.audience": "leadership", "comm_log.channel_or_meeting": "Bridge", "comm_log.summary": "Coordination update", "comm_log.action_task_ids": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "comm_log.audience_party_ids", viewSchemaID: "cartulary.view.comm_log.v1", itemKind: "party_ref", valueKey: "party_id", arrange: func(t *testing.T) collectionReviewInput {
			baseRow := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{"client_txn_id": "txn-" + strings.ReplaceAll("comm_log.audience_party_ids", ".", "-") + "-base-party", "party.display_name": "base party", "party.party_kind": "person"})["row"].(map[string]any)
			base := appsupport.MustUUID(t, baseRow["record_id"].(string))
			serverRow := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{"client_txn_id": "txn-" + strings.ReplaceAll("comm_log.audience_party_ids", ".", "-") + "-server-party", "party.display_name": "server party", "party.party_kind": "person"})["row"].(map[string]any)
			server := appsupport.MustUUID(t, serverRow["record_id"].(string))
			clientRow := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{"client_txn_id": "txn-" + strings.ReplaceAll("comm_log.audience_party_ids", ".", "-") + "-client-party", "party.display_name": "client party", "party.party_kind": "person"})["row"].(map[string]any)
			client := appsupport.MustUUID(t, clientRow["record_id"].(string))
			return collectionReviewInput{
				create:        map[string]any{"comm_log.comm_type": "briefing", "comm_log.audience": "leadership", "comm_log.channel_or_meeting": "Bridge", "comm_log.summary": "Coordination update", "comm_log.audience_party_ids": collectionActions(addPartyRef(base))},
				serverActions: collectionActions(addPartyRef(server)), clientActions: collectionActions(addPartyRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "comm_log.attendee_party_ids", viewSchemaID: "cartulary.view.comm_log.v1", itemKind: "party_ref", valueKey: "party_id", arrange: func(t *testing.T) collectionReviewInput {
			baseRow := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{"client_txn_id": "txn-" + strings.ReplaceAll("comm_log.attendee_party_ids", ".", "-") + "-base-party", "party.display_name": "base party", "party.party_kind": "person"})["row"].(map[string]any)
			base := appsupport.MustUUID(t, baseRow["record_id"].(string))
			serverRow := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{"client_txn_id": "txn-" + strings.ReplaceAll("comm_log.attendee_party_ids", ".", "-") + "-server-party", "party.display_name": "server party", "party.party_kind": "person"})["row"].(map[string]any)
			server := appsupport.MustUUID(t, serverRow["record_id"].(string))
			clientRow := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{"client_txn_id": "txn-" + strings.ReplaceAll("comm_log.attendee_party_ids", ".", "-") + "-client-party", "party.display_name": "client party", "party.party_kind": "person"})["row"].(map[string]any)
			client := appsupport.MustUUID(t, clientRow["record_id"].(string))
			return collectionReviewInput{
				create:        map[string]any{"comm_log.comm_type": "briefing", "comm_log.audience": "leadership", "comm_log.channel_or_meeting": "Bridge", "comm_log.summary": "Coordination update", "comm_log.attendee_party_ids": collectionActions(addPartyRef(base))},
				serverActions: collectionActions(addPartyRef(server)), clientActions: collectionActions(addPartyRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "decision.support_refs", viewSchemaID: "cartulary.view.decisions.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedEvidenceRecord(t, harness, incidentID, actorID, "base reference")
			server := seedEvidenceRecord(t, harness, incidentID, actorID, "server reference")
			client := seedEvidenceRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"decision.summary": "Contain endpoint", "decision.decision_type": "containment", "decision.rationale": "Containment required", "decision.support_refs": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "decision.affected_record_ids", viewSchemaID: "cartulary.view.decisions.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedEvidenceRecord(t, harness, incidentID, actorID, "base reference")
			server := seedEvidenceRecord(t, harness, incidentID, actorID, "server reference")
			client := seedEvidenceRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"decision.summary": "Contain endpoint", "decision.decision_type": "containment", "decision.rationale": "Containment required", "decision.affected_record_ids": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "handoff.open_task_ids", viewSchemaID: "cartulary.view.handoff.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedTaskRecord(t, harness, incidentID, actorID, "base reference")
			server := seedTaskRecord(t, harness, incidentID, actorID, "server reference")
			client := seedTaskRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"handoff.current_state_summary": "Night shift owns containment", "handoff.incoming_owner_user_id": actorID.String(), "handoff.open_task_ids": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "handoff.open_decision_ids", viewSchemaID: "cartulary.view.handoff.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedDecisionRecord(t, harness, incidentID, actorID, "base reference")
			server := seedDecisionRecord(t, harness, incidentID, actorID, "server reference")
			client := seedDecisionRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"handoff.current_state_summary": "Night shift owns containment", "handoff.incoming_owner_user_id": actorID.String(), "handoff.open_decision_ids": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "handoff.open_risk_refs", viewSchemaID: "cartulary.view.handoff.v1", itemKind: "risk_ref", valueKey: "risk_ref_text", arrange: func(t *testing.T) collectionReviewInput {
			base := "base-risk"
			server := "server-risk"
			client := "client-risk"
			return collectionReviewInput{
				create:        map[string]any{"handoff.current_state_summary": "Night shift owns containment", "handoff.incoming_owner_user_id": actorID.String(), "handoff.open_risk_refs": collectionActions(addRiskRef(base))},
				serverActions: collectionActions(addRiskRef(server)), clientActions: collectionActions(addRiskRef(client)),
				baseValue: base, serverValue: server, clientValue: client,
			}
		}},
		{fieldKey: "lesson.follow_up_task_ids", viewSchemaID: "cartulary.view.lesson.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedTaskRecord(t, harness, incidentID, actorID, "base reference")
			server := seedTaskRecord(t, harness, incidentID, actorID, "server reference")
			client := seedTaskRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"lesson.summary": "Preserve logs earlier", "lesson.follow_up_task_ids": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "lesson.evidence_refs", viewSchemaID: "cartulary.view.lesson.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedEvidenceRecord(t, harness, incidentID, actorID, "base reference")
			server := seedEvidenceRecord(t, harness, incidentID, actorID, "server reference")
			client := seedEvidenceRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"lesson.summary": "Preserve logs earlier", "lesson.evidence_refs": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "status_review.blocked_task_ids", viewSchemaID: "cartulary.view.status_review.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedTaskRecord(t, harness, incidentID, actorID, "base reference")
			server := seedTaskRecord(t, harness, incidentID, actorID, "server reference")
			client := seedTaskRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"status_review.current_state_summary": "Containment stable", "status_review.blocked_task_ids": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "status_review.pending_evidence_ids", viewSchemaID: "cartulary.view.status_review.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedEvidenceRecord(t, harness, incidentID, actorID, "base reference")
			server := seedEvidenceRecord(t, harness, incidentID, actorID, "server reference")
			client := seedEvidenceRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"status_review.current_state_summary": "Containment stable", "status_review.pending_evidence_ids": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "status_review.open_decision_ids", viewSchemaID: "cartulary.view.status_review.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedDecisionRecord(t, harness, incidentID, actorID, "base reference")
			server := seedDecisionRecord(t, harness, incidentID, actorID, "server reference")
			client := seedDecisionRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"status_review.current_state_summary": "Containment stable", "status_review.open_decision_ids": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
		{fieldKey: "task.linked_record_ids", viewSchemaID: "cartulary.view.task_requests.v1", itemKind: "record_ref", valueKey: "linked_record_id", arrange: func(t *testing.T) collectionReviewInput {
			base := seedEvidenceRecord(t, harness, incidentID, actorID, "base reference")
			server := seedEvidenceRecord(t, harness, incidentID, actorID, "server reference")
			client := seedEvidenceRecord(t, harness, incidentID, actorID, "client reference")
			return collectionReviewInput{
				create:        map[string]any{"task.title": "Collect logs", "task.task_kind": "collection", "task.linked_record_ids": collectionActions(addRecordRef(base))},
				serverActions: collectionActions(addRecordRef(server)), clientActions: collectionActions(addRecordRef(client)),
				baseValue: base.String(), serverValue: server.String(), clientValue: client.String(),
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.fieldKey, func(t *testing.T) {
			input := tc.arrange(t)
			// REQ-01-333a labels per linked record, across rows and previews.
			labels := map[string]string{}
			txn := "txn-review-" + strings.NewReplacer(".", "-", "_", "-").Replace(tc.fieldKey)
			input.create["client_txn_id"] = txn + "-create"
			baseRow := requireWorkbookCreate(t, harness, login, incidentID, tc.viewSchemaID, input.create)["row"].(map[string]any)
			recordID := appsupport.MustUUID(t, baseRow["record_id"].(string))
			baseVersion := int64(baseRow["row_version"].(float64))
			requireCollectionReviewValue(t, cellMapValue(t, baseRow, tc.fieldKey), tc.itemKind, tc.valueKey, []string{input.baseValue})
			requireLinkedRecordLabels(t, labels, cellMapValue(t, baseRow, tc.fieldKey))
			requireCollectionReviewValue(t, cellMapValue(t, RequireQueriedRow(t, harness, login, incidentID, tc.viewSchemaID, recordID), tc.fieldKey), tc.itemKind, tc.valueKey, []string{input.baseValue})
			serverRow := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
				"view_schema_id": tc.viewSchemaID, "base_row_version": baseVersion, "client_txn_id": txn + "-server",
				"changes": []map[string]any{{"field_key": tc.fieldKey, "action_payload": input.serverActions}},
			})["row"].(map[string]any)
			serverVersion := int64(serverRow["row_version"].(float64))
			if serverVersion != baseVersion+1 {
				t.Fatal("server write must advance exactly one row version")
			}
			requireCollectionReviewValue(t, cellMapValue(t, serverRow, tc.fieldKey), tc.itemKind, tc.valueKey, []string{input.baseValue, input.serverValue})
			requireLinkedRecordLabels(t, labels, cellMapValue(t, serverRow, tc.fieldKey))
			before := snapshotWorkbookState(t, harness, login, incidentID, tc.viewSchemaID)
			response := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
				"view_schema_id": tc.viewSchemaID, "base_row_version": baseVersion, "client_txn_id": txn + "-client",
				"changes": []map[string]any{{"field_key": tc.fieldKey, "action_payload": input.clientActions}},
			})
			envelope := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "same_field_conflict")
			requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, harness, login, incidentID, tc.viewSchemaID))
			conflict := envelope["error"].(map[string]any)["conflict"].(map[string]any)
			RequireNoLegacyConflictAliases(t, conflict)
			token, ok := conflict["conflict_token"].(string)
			if !ok || token == "" {
				t.Fatal("conflict must carry a nonempty opaque string token")
			}
			if conflict["record_id"] != recordID.String() || conflict["field_key"] != tc.fieldKey || conflict["server_updated_by"] != actorID.String() || conflict["conflict_resolution_class"] != "collection_review" {
				t.Fatal("collection conflict identity mismatch")
			}
			stamp, ok := conflict["server_updated_at"].(string)
			if !ok {
				t.Fatal("server_updated_at must be a string")
			}
			if _, err := time.Parse(time.RFC3339Nano, stamp); err != nil {
				t.Fatal(err)
			}
			RequireConflictVersion(t, conflict, "base_row_version", baseVersion)
			RequireConflictVersion(t, conflict, "current_row_version", serverVersion)
			requireCollectionReviewValue(t, conflict["base_value"].(map[string]any), tc.itemKind, tc.valueKey, []string{input.baseValue})
			requireCollectionReviewValue(t, conflict["server_value"].(map[string]any), tc.itemKind, tc.valueKey, []string{input.baseValue, input.serverValue})
			requireCollectionReviewValue(t, conflict["client_value"].(map[string]any), tc.itemKind, tc.valueKey, []string{input.baseValue, input.clientValue})
			for _, member := range []string{"base_value", "server_value", "client_value"} {
				requireLinkedRecordLabels(t, labels, conflict[member].(map[string]any))
			}
			beforeResolve := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
			body := map[string]any{"conflict_token": token, "resolution_kind": "merged_value", "client_txn_id": txn + "-resolve", "resolved_value": input.clientActions}
			resolved := ResolveConflict(t, harness, login, recordID, token, body)
			row := resolved["row"].(map[string]any)
			if int64(row["row_version"].(float64)) != serverVersion+1 {
				t.Fatal("resolution must advance exactly one row version")
			}
			want := []string{input.baseValue, input.serverValue, input.clientValue}
			requireCollectionReviewValue(t, cellMapValue(t, row, tc.fieldKey), tc.itemKind, tc.valueKey, want)
			// The resolved row reads every label, including the client item's,
			// so the preview labels above must equal the read-path labels.
			requireLinkedRecordLabels(t, labels, cellMapValue(t, row, tc.fieldKey))
			requireCollectionReviewValue(t, cellMapValue(t, RequireQueriedRow(t, harness, login, incidentID, tc.viewSchemaID, recordID), tc.fieldKey), tc.itemKind, tc.valueKey, want)
			afterResolve := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
			// Links and tags have their own reversible mutation in addition to the
			// parent row mutation. Handoff risk references belong to its row snapshot.
			mutationRows := 2
			if tc.itemKind == "risk_ref" {
				mutationRows = 1
			}
			if afterResolve.ChangeSets != beforeResolve.ChangeSets+1 || afterResolve.MutationRows != beforeResolve.MutationRows+mutationRows || afterResolve.RecordRevisions != beforeResolve.RecordRevisions+1 || afterResolve.RouteIdempotency != beforeResolve.RouteIdempotency+1 {
				t.Fatalf("resolution must persist one change, revision and receipt, and %d reversible mutations: before=%+v after=%+v", mutationRows, beforeResolve, afterResolve)
			}
			workbookscenariotest.RequireChangeSetAttribution(t, harness.DB, resolved["change_set_id"].(string), actorID.String(), "workbook.records.conflicts.resolve", txn+"-resolve")
			revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, resolved["change_set_id"].(string))
			beforeReplay := snapshotWorkbookState(t, harness, login, incidentID, tc.viewSchemaID)
			replay := ResolveConflict(t, harness, login, recordID, token, body)
			if !reflect.DeepEqual(resolved, replay) {
				t.Fatal("collection resolution replay changed original result")
			}
			requireWorkbookStateUnchanged(t, beforeReplay, snapshotWorkbookState(t, harness, login, incidentID, tc.viewSchemaID))
		})
	}
}

// requireLinkedRecordLabels checks the REQ-01-333a labels of record_ref and
// party_ref items: each is free of identifiers, live targets never read
// "Unavailable record", and a linked record reads the same label wherever it
// appears. labels accumulates the observations of one case.
func requireLinkedRecordLabels(t testing.TB, labels map[string]string, value map[string]any) {
	t.Helper()
	for _, raw := range value["items"].([]any) {
		item := raw.(map[string]any)
		key := ""
		switch item["item_kind"] {
		case "record_ref":
			key = "linked_record_id"
		case "party_ref":
			key = "party_id"
		default:
			continue
		}
		id, _ := item[key].(string)
		label, _ := item["display_text"].(string)
		if label == "" || label == "Unavailable record" || recordLabelUUIDPattern.MatchString(label) {
			t.Fatalf("linked record %s label %q is not its REQ-01-333a summary", id, label)
		}
		if previous, seen := labels[id]; seen && previous != label {
			t.Fatalf("linked record %s label %q differs from %q served elsewhere", id, label, previous)
		}
		labels[id] = label
	}
}

// TestCollectionReviewUnavailableLinkedRecordLabel_Integration covers the
// REQ-01-333a preview rule: conflict detection precedes reference validation,
// so a stale client addition can name a record that is not live in the
// incident, and its preview item reads "Unavailable record".
func TestCollectionReviewUnavailableLinkedRecordLabel_Integration(t *testing.T) {
	harness, login, actorID, incidentID := ConflictFixture(t, "workbook-collection-unavailable-label", "IR-WORKBOOK-COLLECTION-UNAVAILABLE")
	for _, tc := range []struct {
		fieldKey, viewSchemaID string
		create                 func(base uuid.UUID) map[string]any
		seed                   func(label string) uuid.UUID
	}{
		{fieldKey: "task.linked_record_ids", viewSchemaID: "cartulary.view.task_requests.v1",
			create: func(base uuid.UUID) map[string]any {
				return map[string]any{"task.title": "Unavailable label task", "task.task_kind": "collection", "task.linked_record_ids": collectionActions(addRecordRef(base))}
			},
			seed: func(label string) uuid.UUID { return seedTaskRecord(t, harness, incidentID, actorID, label) }},
		{fieldKey: "comm_log.decision_ids", viewSchemaID: "cartulary.view.comm_log.v1",
			create: func(base uuid.UUID) map[string]any {
				return map[string]any{"comm_log.comm_type": "briefing", "comm_log.audience": "leadership", "comm_log.channel_or_meeting": "Bridge", "comm_log.summary": "Unavailable label update", "comm_log.decision_ids": collectionActions(addRecordRef(base))}
			},
			seed: func(label string) uuid.UUID { return seedDecisionRecord(t, harness, incidentID, actorID, label) }},
	} {
		t.Run(tc.fieldKey, func(t *testing.T) {
			base, server, absent := tc.seed("Unavailable base"), tc.seed("Unavailable server"), uuid.New()
			txn := "txn-unavailable-" + strings.NewReplacer(".", "-", "_", "-").Replace(tc.fieldKey)
			create := tc.create(base)
			create["client_txn_id"] = txn + "-create"
			baseRow := requireWorkbookCreate(t, harness, login, incidentID, tc.viewSchemaID, create)["row"].(map[string]any)
			recordID := appsupport.MustUUID(t, baseRow["record_id"].(string))
			baseVersion := int64(baseRow["row_version"].(float64))
			requireWorkbookPatch(t, harness, login, recordID, map[string]any{
				"view_schema_id": tc.viewSchemaID, "base_row_version": baseVersion, "client_txn_id": txn + "-server",
				"changes": []map[string]any{{"field_key": tc.fieldKey, "action_payload": collectionActions(addRecordRef(server))}},
			})
			response := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
				"view_schema_id": tc.viewSchemaID, "base_row_version": baseVersion, "client_txn_id": txn + "-client",
				"changes": []map[string]any{{"field_key": tc.fieldKey, "action_payload": collectionActions(addRecordRef(absent))}},
			})
			envelope := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "same_field_conflict")
			client := envelope["error"].(map[string]any)["conflict"].(map[string]any)["client_value"].(map[string]any)
			got := map[string]string{}
			for _, raw := range client["items"].([]any) {
				item := raw.(map[string]any)
				got[item["linked_record_id"].(string)] = item["display_text"].(string)
			}
			want := map[string]string{base.String(): "Unavailable base", absent.String(): "Unavailable record"}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("client preview labels = %v, want %v", got, want)
			}
		})
	}
}

func requireCollectionReviewValue(t testing.TB, value map[string]any, itemKind, valueKey string, want []string) {
	t.Helper()
	if value["kind"] != "collection_value_v1" {
		t.Fatal("expected collection_value_v1")
	}
	items, ok := value["items"].([]any)
	if !ok || len(items) != len(want) {
		t.Fatalf("collection contains %d items, want %d", len(items), len(want))
	}
	actual := make([]string, 0, len(items))
	refs := map[string]bool{}
	for _, raw := range items {
		item := raw.(map[string]any)
		if item["item_kind"] != itemKind {
			t.Fatal("collection changed item kind")
		}
		ref, ok := item["item_ref"].(string)
		if !ok || ref == "" || refs[ref] {
			t.Fatal("collection item identity missing or duplicated")
		}
		refs[ref] = true
		got, ok := item[valueKey].(string)
		if !ok || got == "" {
			t.Fatal("collection semantic value missing")
		}
		actual = append(actual, got)
	}
	expected := append([]string(nil), want...)
	sort.Strings(actual)
	sort.Strings(expected)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("collection contents = %v, want %v", actual, expected)
	}
}

func TestCollectionMalformedActions_Integration(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "workbook-collection-malformed", "IR-WORKBOOK-COLLECTION-MALFORMED")
	row := CreateNote(t, harness, login, incidentID, "txn-malformed-note", "Malformed collection", "Body")
	recordID := appsupport.MustUUID(t, row["record_id"].(string))
	for _, tc := range []struct {
		name    string
		payload any
	}{
		{"raw-array", []any{}}, {"raw-string", "client-tag"}, {"replace-all", map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "replace_all", "raw_text": "client-tag"}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID)
			response := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
				"view_schema_id": artifacts.NotesViewSchemaID, "base_row_version": row["row_version"], "client_txn_id": "txn-malformed-" + tc.name,
				"changes": []map[string]any{{"field_key": "note.tags", "action_payload": tc.payload}},
			})
			httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_mutation_payload")
			requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, harness, login, incidentID, artifacts.NotesViewSchemaID))
		})
	}
}
