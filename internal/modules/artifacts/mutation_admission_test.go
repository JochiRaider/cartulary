package artifacts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestArtifactMutationAdmissionAndReplayHashing(t *testing.T) {
	t.Run("create admission is fixed to Artifact views and normalizes owner fields", func(t *testing.T) {
		admission, admissionErr := AdmitCreate(NotesViewSchemaID, strings.NewReader(`{
			"client_txn_id":"txn-artifact-admission-create",
			"note.title":"  Owner title  ",
			"note.body":"Owner body"
		}`))
		if admissionErr != nil {
			t.Fatalf("decode Artifact create: %#v", admissionErr)
		}
		request := admission.requestValue()
		if got := *request.Values["note.title"].Text; got != "Owner title" {
			t.Fatalf("normalized title = %q, want Owner title", got)
		}
		requireHashHex(t, admission.requestHash(), "0bc8483018d9cca7bb393c4075534cee23567a77d34a4034e2f282c53da9cdf6")
		request.Values["note.title"] = fieldValue{}
		if got := *admission.requestValue().Values["note.title"].Text; got != "Owner title" {
			t.Fatalf("mutation escaped admission defensive copy: %q", got)
		}
		if _, admissionErr := AdmitCreate("cartulary.view.task_requests.v1", strings.NewReader(`{"client_txn_id":"txn"}`)); admissionErr == nil {
			t.Fatal("non-Artifact view reached Artifact create admission")
		}
		if (CreateAdmission{}).valid() {
			t.Fatal("zero create admission is valid")
		}
	})

	t.Run("patch hashing ignores transaction and change order after normalization", func(t *testing.T) {
		left, admissionErr := AdmitPatch(strings.NewReader(`{
			"view_schema_id":"cartulary.view.comm_log.v1",
			"base_row_version":1,
			"client_txn_id":"txn-left",
			"changes":[
				{"field_key":"comm_log.summary","value":" Updated "},
				{"field_key":"comm_log.privilege_tag","value":"legal"}
			]
		}`))
		if admissionErr != nil {
			t.Fatalf("decode left Artifact patch: %#v", admissionErr)
		}
		right, admissionErr := AdmitPatch(strings.NewReader(`{
			"view_schema_id":"cartulary.view.comm_log.v1",
			"base_row_version":1,
			"client_txn_id":"txn-right",
			"changes":[
				{"field_key":"comm_log.privilege_tag","value":"legal"},
				{"field_key":"comm_log.summary","value":"Updated"}
			]
		}`))
		if admissionErr != nil {
			t.Fatalf("decode right Artifact patch: %#v", admissionErr)
		}
		if !bytes.Equal(left.requestHash(), right.requestHash()) {
			t.Fatal("canonical Artifact patch hash depends on transaction ID or change order")
		}
		requireHashHex(t, left.requestHash(), "97b5e8baf397af1358e6d4973613f943cf5f76c0faba813655e690d95e634c36")
	})

	t.Run("collection admission preserves display text and validates stable item references", func(t *testing.T) {
		admission, admissionErr := AdmitPatch(strings.NewReader(`{
			"view_schema_id":"cartulary.view.handoff.v1",
			"base_row_version":1,
			"client_txn_id":"txn-risk",
			"changes":[{"field_key":"handoff.open_risk_refs","action_payload":{
				"kind":"collection_actions_v1",
				"actions":[{"op":"add_risk_ref","risk_ref_text":"  Preserve Display Text  "}]
			}}]
		}`))
		if admissionErr != nil {
			t.Fatalf("decode Artifact collection patch: %#v", admissionErr)
		}
		request := admission.requestValue()
		action := request.Changes[0].Collection.Actions[0]
		if action.RiskRefText != "Preserve Display Text" || action.NormalizedText != "preserve display text" {
			t.Fatalf("risk reference admission = %#v", action)
		}
		if _, admissionErr := AdmitPatch(strings.NewReader(`{
			"view_schema_id":"cartulary.view.handoff.v1",
			"base_row_version":1,
			"client_txn_id":"txn-risk-invalid",
			"changes":[{"field_key":"handoff.open_risk_refs","action_payload":{
				"kind":"collection_actions_v1",
				"actions":[{"op":"remove_risk_ref","item_ref":"risk_ref:11111111-2222-3333-4444-555555555555 "}]
			}}]
		}`)); admissionErr == nil {
			t.Fatal("padded Artifact risk item reference was admitted")
		}
	})

	t.Run("conflict admission binds token claims and contextual hashes", func(t *testing.T) {
		recordID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
		claims := ConflictAdmissionContext{
			Version: 3, RecordID: recordID, ViewSchemaID: NotesViewSchemaID,
			RouteKey: "workbook.records.conflicts.resolve", FieldKey: "note.body",
			ConflictResolutionClass: "text_compare_merge",
			BaseRowVersion:          1, CurrentRowVersion: 2,
			OriginalRequestHash: "original-hash",
			IssuedAt:            time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC),
			ExpiresAt:           time.Date(2026, 7, 30, 12, 30, 0, 0, time.UTC),
		}
		admission, admissionErr := AdmitConflictResolution(strings.NewReader(`{
			"conflict_token":"opaque-token",
			"resolution_kind":"merged_value",
			"client_txn_id":"txn-conflict",
			"resolved_value":"  merged body  "
		}`), "opaque-token", claims)
		if admissionErr != nil {
			t.Fatalf("decode Artifact conflict resolution: %#v", admissionErr)
		}
		request := admission.requestValue()
		if request.Patch == nil || request.Patch.ViewSchemaID != NotesViewSchemaID ||
			request.Patch.BaseRowVersion != 2 || request.CanonicalValue != "merged body" {
			t.Fatalf("Artifact conflict admission = %#v", request)
		}
		requireHashHex(t, admission.requestHash(), "c470b5515c71d59624d638012559060091e9dfd4084729f1b8a683c468879fee")

		note, admissionErr := AdmitContextualNote(strings.NewReader(`{
			"client_txn_id":"txn-contextual-note",
			"note.title":"Contextual owner note"
		}`))
		if admissionErr != nil {
			t.Fatalf("decode contextual Artifact note: %#v", admissionErr)
		}
		requireHashHex(t, note.requestHash(recordID), "05941318af2aa2e4014cfef3a3b56c43ff81ef6969fdf9b49532ea19d51f2271")
	})

	t.Run("strict JSON failures preserve the public admission error", func(t *testing.T) {
		for name, payload := range map[string]string{
			"non_object":          `[]`,
			"trailing_value":      `{"client_txn_id":"txn"}{}`,
			"duplicate_top_level": `{"client_txn_id":"txn","client_txn_id":"other"}`,
			"duplicate_nested": `{
				"client_txn_id":"txn",
				"note.tags":{"kind":"collection_actions_v1","actions":[
					{"op":"add_tag","op":"add_tag","tag_name":"owner"}
				]}
			}`,
		} {
			t.Run(name, func(t *testing.T) {
				_, admissionErr := AdmitCreate(NotesViewSchemaID, strings.NewReader(payload))
				requireAdmissionError(t, admissionErr, "", "request_not_object", nil)
			})
		}
	})

	t.Run("patch and collection limits preserve exact detail facts", func(t *testing.T) {
		changes := make([]string, 33)
		for index := range changes {
			changes[index] = `{"field_key":"note.title","value":"owner"}`
		}
		_, admissionErr := AdmitPatch(strings.NewReader(fmt.Sprintf(`{
			"view_schema_id":"cartulary.view.notes.v1",
			"base_row_version":1,
			"client_txn_id":"txn-too-many-changes",
			"changes":[%s]
		}`, strings.Join(changes, ","))))
		requireAdmissionError(t, admissionErr, "changes", "change_count_exceeded", map[string]any{
			"requested_count": 33,
			"max_count":       32,
		})

		actions := make([]string, 65)
		for index := range actions {
			actions[index] = `{"op":"add_tag","tag_name":"owner"}`
		}
		_, admissionErr = AdmitPatch(strings.NewReader(fmt.Sprintf(`{
			"view_schema_id":"cartulary.view.notes.v1",
			"base_row_version":1,
			"client_txn_id":"txn-too-many-actions",
			"changes":[{"field_key":"note.tags","action_payload":{
				"kind":"collection_actions_v1","actions":[%s]
			}}]
		}`, strings.Join(actions, ","))))
		requireAdmissionError(t, admissionErr, "note.tags.actions", "collection_action_count_exceeded", map[string]any{
			"field_key":       "note.tags",
			"requested_count": 65,
			"max_count":       64,
		})
	})

	t.Run("zero and incomplete admissions fail closed", func(t *testing.T) {
		if (CreateAdmission{}).valid() || (PatchAdmission{}).valid() ||
			(ContextualNoteAdmission{}).valid() || (ConflictResolveAdmission{}).valid() {
			t.Fatal("a zero admission was valid")
		}
		if _, admissionErr := AdmitConflictResolution(
			strings.NewReader(`{"conflict_token":"token","resolution_kind":"keep_saved","client_txn_id":"txn"}`),
			"token",
			ConflictAdmissionContext{RecordID: uuid.New(), ViewSchemaID: NotesViewSchemaID},
		); admissionErr == nil {
			t.Fatal("incomplete conflict context was admitted")
		}
	})
}

func requireHashHex(t testing.TB, hash []byte, want string) {
	t.Helper()
	if got := hex.EncodeToString(hash); got != want {
		t.Fatalf("request hash = %s, want %s", got, want)
	}
}

func requireAdmissionError(t testing.TB, admissionErr *AdmissionError, field string, reason string, extra map[string]any) {
	t.Helper()
	if admissionErr == nil {
		t.Fatal("admission unexpectedly succeeded")
	}
	if got := string(admissionErr.ReasonCode()); got != reason {
		t.Fatalf("admission reason = %q, want %q", got, reason)
	}
	if got, present := admissionErr.Field(); got != field || present != (field != "") {
		t.Fatalf("admission field = %q, %t; want %q", got, present, field)
	}
	if requested, present := extra["requested_count"]; present {
		if got, ok := admissionErr.RequestedCount(); !ok || got != requested {
			t.Fatalf("requested count = %d, %t; want %v", got, ok, requested)
		}
	}
	if maximum, present := extra["max_count"]; present {
		if got, ok := admissionErr.MaximumCount(); !ok || got != maximum {
			t.Fatalf("maximum count = %d, %t; want %v", got, ok, maximum)
		}
	}
	if collectionField, present := extra["field_key"]; present {
		if got, ok := admissionErr.CollectionField(); !ok || got != collectionField {
			t.Fatalf("collection field = %q, %t; want %v", got, ok, collectionField)
		}
	}
}
