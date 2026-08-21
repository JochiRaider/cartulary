package artifacts_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
)

func TestArtifactMutationAdmissionAndReplayHashing(t *testing.T) {
	t.Run("create admission is fixed to Artifact views and normalizes owner fields", func(t *testing.T) {
		request, apiErr := artifacts.DecodeCreateRequest(artifacts.NotesViewSchemaID, strings.NewReader(`{
			"client_txn_id":"txn-artifact-admission-create",
			"note.title":"  Owner title  ",
			"note.body":"Owner body"
		}`))
		if apiErr != nil {
			t.Fatalf("decode Artifact create: %#v", apiErr)
		}
		if got := *request.Values["note.title"].Text; got != "Owner title" {
			t.Fatalf("normalized title = %q, want Owner title", got)
		}
		if _, apiErr := artifacts.DecodeCreateRequest("cartulary.view.task_requests.v1", strings.NewReader(`{"client_txn_id":"txn"}`)); apiErr == nil {
			t.Fatal("non-Artifact view reached Artifact create admission")
		}
	})

	t.Run("patch hashing ignores transaction and change order after normalization", func(t *testing.T) {
		left, apiErr := artifacts.DecodePatchRequest(strings.NewReader(`{
			"view_schema_id":"cartulary.view.comm_log.v1",
			"base_row_version":1,
			"client_txn_id":"txn-left",
			"changes":[
				{"field_key":"comm_log.summary","value":" Updated "},
				{"field_key":"comm_log.privilege_tag","value":"legal"}
			]
		}`))
		if apiErr != nil {
			t.Fatalf("decode left Artifact patch: %#v", apiErr)
		}
		right, apiErr := artifacts.DecodePatchRequest(strings.NewReader(`{
			"view_schema_id":"cartulary.view.comm_log.v1",
			"base_row_version":1,
			"client_txn_id":"txn-right",
			"changes":[
				{"field_key":"comm_log.privilege_tag","value":"legal"},
				{"field_key":"comm_log.summary","value":"Updated"}
			]
		}`))
		if apiErr != nil {
			t.Fatalf("decode right Artifact patch: %#v", apiErr)
		}
		if !bytes.Equal(artifacts.PatchRequestHash(left), artifacts.PatchRequestHash(right)) {
			t.Fatal("canonical Artifact patch hash depends on transaction ID or change order")
		}
	})

	t.Run("collection admission preserves display text and validates stable item references", func(t *testing.T) {
		request, apiErr := artifacts.DecodePatchRequest(strings.NewReader(`{
			"view_schema_id":"cartulary.view.handoff.v1",
			"base_row_version":1,
			"client_txn_id":"txn-risk",
			"changes":[{"field_key":"handoff.open_risk_refs","action_payload":{
				"kind":"collection_actions_v1",
				"actions":[{"op":"add_risk_ref","risk_ref_text":"  Preserve Display Text  "}]
			}}]
		}`))
		if apiErr != nil {
			t.Fatalf("decode Artifact collection patch: %#v", apiErr)
		}
		action := request.Changes[0].Collection.Actions[0]
		if action.RiskRefText != "Preserve Display Text" || action.NormalizedText != "preserve display text" {
			t.Fatalf("risk reference admission = %#v", action)
		}
		if _, apiErr := artifacts.DecodePatchRequest(strings.NewReader(`{
			"view_schema_id":"cartulary.view.handoff.v1",
			"base_row_version":1,
			"client_txn_id":"txn-risk-invalid",
			"changes":[{"field_key":"handoff.open_risk_refs","action_payload":{
				"kind":"collection_actions_v1",
				"actions":[{"op":"remove_risk_ref","item_ref":"risk_ref:11111111-2222-3333-4444-555555555555 "}]
			}}]
		}`)); apiErr == nil {
			t.Fatal("padded Artifact risk item reference was admitted")
		}
	})

	t.Run("conflict admission binds token claims and contextual hashes", func(t *testing.T) {
		recordID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
		claims := artifacts.ConflictClaims{
			RecordID: recordID, ViewSchemaID: artifacts.NotesViewSchemaID,
			FieldKey: "note.body", CurrentRowVersion: 2,
		}
		request, apiErr := artifacts.DecodeConflictResolveRequest(strings.NewReader(`{
			"conflict_token":"opaque-token",
			"resolution_kind":"merged_value",
			"client_txn_id":"txn-conflict",
			"resolved_value":"  merged body  "
		}`), "opaque-token", claims)
		if apiErr != nil {
			t.Fatalf("decode Artifact conflict resolution: %#v", apiErr)
		}
		if request.Patch == nil || request.Patch.ViewSchemaID != artifacts.NotesViewSchemaID ||
			request.Patch.BaseRowVersion != 2 || request.CanonicalValue != "merged body" {
			t.Fatalf("Artifact conflict admission = %#v", request)
		}
		if len(artifacts.ConflictResolveRequestHash(claims, request)) == 0 {
			t.Fatal("Artifact conflict hash is empty")
		}

		note, apiErr := artifacts.DecodeContextualNoteCreateRequest(strings.NewReader(`{
			"client_txn_id":"txn-contextual-note",
			"note.title":"Contextual owner note"
		}`))
		if apiErr != nil {
			t.Fatalf("decode contextual Artifact note: %#v", apiErr)
		}
		if len(artifacts.ContextualNoteCreateRequestHash(recordID, note)) == 0 {
			t.Fatal("contextual Artifact note hash is empty")
		}
	})
}
