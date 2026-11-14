package artifacts

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestArtifactMutationAdmissionAndReplayHashing(t *testing.T) {
	t.Run("shared tag label contract", requireNoteTagLabelAdmission)
	t.Run("risk text contracts across create patch and conflict", requireArtifactRiskTextAdmission)
	t.Run("bound text contracts across create patch and conflict", requireArtifactBoundTextAdmission)
	t.Run("collection confidence belongs to neither create nor patch admission", func(t *testing.T) {
		const target = "10000000-0000-4000-8000-000000000001"
		const comm = `"comm_log.comm_type":"briefing","comm_log.audience":"leadership","comm_log.channel_or_meeting":"Bridge","comm_log.summary":"Coordination"`
		const handoff = `"handoff.incoming_owner_user_id":"10000000-0000-4000-8000-000000000002","handoff.current_state_summary":"Coordination"`
		for _, tc := range []struct {
			view, field, values, action string
		}{
			{CommLogViewSchemaID, "comm_log.decision_ids", comm, "record"},
			{CommLogViewSchemaID, "comm_log.action_task_ids", comm, "record"},
			{CommLogViewSchemaID, "comm_log.audience_party_ids", comm, "party"},
			{CommLogViewSchemaID, "comm_log.attendee_party_ids", comm, "party"},
			{HandoffViewSchemaID, "handoff.open_task_ids", handoff, "record"},
			{HandoffViewSchemaID, "handoff.open_decision_ids", handoff, "record"},
			{HandoffViewSchemaID, "handoff.open_risk_refs", handoff, "risk"},
			{StatusReviewViewSchemaID, "status_review.blocked_task_ids", `"status_review.current_state_summary":"Coordination"`, "record"},
			{StatusReviewViewSchemaID, "status_review.pending_evidence_ids", `"status_review.current_state_summary":"Coordination"`, "record"},
			{StatusReviewViewSchemaID, "status_review.open_decision_ids", `"status_review.current_state_summary":"Coordination"`, "record"},
			{LessonViewSchemaID, "lesson.follow_up_task_ids", `"lesson.summary":"Coordination"`, "record"},
			{LessonViewSchemaID, "lesson.evidence_refs", `"lesson.summary":"Coordination"`, "record"},
		} {
			t.Run(tc.field, func(t *testing.T) {
				action := `"op":"add_record_ref","linked_record_id":"` + target + `"`
				if tc.action == "party" {
					action = `"op":"add_party_ref","party_id":"` + target + `"`
				} else if tc.action == "risk" {
					action = `"op":"add_risk_ref","risk_ref_text":"A risk"`
				}
				for _, forbidden := range []bool{false, true} {
					members := action
					if forbidden {
						members += `,"confidence":80`
					}
					payload := `{"kind":"collection_actions_v1","actions":[{` + members + `}]}`
					create := fmt.Sprintf(`{"client_txn_id":"confidence-create",%s,%q:%s}`, tc.values, tc.field, payload)
					patch := fmt.Sprintf(`{"view_schema_id":%q,"client_txn_id":"confidence-patch","base_row_version":1,"changes":[{"field_key":%q,"action_payload":%s}]}`, tc.view, tc.field, payload)
					_, createErr := AdmitCreate(tc.view, strings.NewReader(create))
					_, patchErr := AdmitPatch(strings.NewReader(patch))
					for route, err := range map[string]*AdmissionError{"create": createErr, "patch": patchErr} {
						if !forbidden {
							if err != nil {
								t.Fatalf("%s valid control: %v", route, err)
							}
						} else if err == nil || err.field != tc.field || err.ReasonCode() != admissionInvalidValue {
							t.Fatalf("%s confidence rejection: %#v, want %s/invalid_value", route, err, tc.field)
						}
					}
				}
			})
		}
	})
	t.Run("Note creation rejects missing signal malformed text and collection operations", func(t *testing.T) {
		tests := []struct {
			name   string
			fields map[string]any
			field  string
			reason AdmissionReasonCode
		}{
			{"empty", map[string]any{}, "payload", admissionMissingMinimumCreateSignal},
			{"whitespace", map[string]any{"note.title": "\u2003 ", "note.body": " \n "}, "payload", admissionMissingMinimumCreateSignal},
			{"tags only", map[string]any{"note.tags": map[string]any{"kind": "collection_actions_v1", "actions": []any{map[string]any{"op": "add_tag", "tag_name": "tag"}}}}, "payload", admissionMissingMinimumCreateSignal},
			{"title control", map[string]any{"note.title": "bad\u0001"}, "note.title", admissionInvalidValue},
			{"body control", map[string]any{"note.body": "bad\u0085"}, "note.body", admissionInvalidValue},
			{"title length", map[string]any{"note.title": strings.Repeat("😀", 513)}, "note.title", admissionInvalidValue},
			{"body length", map[string]any{"note.body": strings.Repeat("x", 16385)}, "note.body", admissionInvalidValue},
			{"tag length", map[string]any{"note.title": "Valid", "note.tags": map[string]any{"kind": "collection_actions_v1", "actions": []any{map[string]any{"op": "add_tag", "tag_name": strings.Repeat("x", 65)}}}}, "note.tags", admissionInvalidValue},
			{"collection operation", map[string]any{"note.title": "Rejected collection", "note.tags": map[string]any{"kind": "collection_actions_v1", "actions": []any{map[string]any{"op": "replace"}}}}, "note.tags", admissionInvalidValue},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				tc.fields["client_txn_id"] = "txn-note-invalid"
				data, err := json.Marshal(tc.fields)
				if err != nil {
					t.Fatal(err)
				}
				_, createErr := AdmitCreate(NotesViewSchemaID, bytes.NewReader(data))
				_, contextualErr := AdmitContextualNote(bytes.NewReader(data))
				for route, admissionErr := range map[string]*AdmissionError{"create": createErr, "contextual": contextualErr} {
					if admissionErr == nil || admissionErr.field != tc.field || admissionErr.ReasonCode() != tc.reason {
						t.Fatalf("%s admission error = %#v, want %s/%s", route, admissionErr, tc.field, tc.reason)
					}
				}
			})
		}
	})
	t.Run("coordination source input preserves omission hashes and distinguishes clear and identities", func(t *testing.T) {
		source := uuid.MustParse("11111111-2222-4333-8444-555555555555")
		base := `{"client_txn_id":"txn","lesson.summary":"Lesson"`
		omitted, err := AdmitCreate(LessonViewSchemaID, strings.NewReader(base+`}`))
		if err != nil {
			t.Fatal(err)
		}
		want := hashArtifactMutationPayload(map[string]any{"view_schema_id": LessonViewSchemaID, "values": map[string]any{"lesson.summary": "Lesson"}, "collection_ops": map[string]any{}, "create_inputs": map[string]any{}})
		if !bytes.Equal(omitted.requestHash(), want[:]) {
			t.Fatal("omitted-input historical hash changed")
		}
		hashes := map[string]bool{hex.EncodeToString(omitted.requestHash()): true}
		for _, input := range []string{`null`, `"` + source.String() + `"`, `"11111111-2222-4333-8444-555555555556"`} {
			got, err := AdmitCreate(LessonViewSchemaID, strings.NewReader(base+`,"coordination.source_record_id":`+input+`}`))
			if err != nil {
				t.Fatal(err)
			}
			if !got.requestValue().CoordinationSourcePresent {
				t.Fatal("source presence lost")
			}
			hash := hex.EncodeToString(got.requestHash())
			if hashes[hash] {
				t.Fatal("distinct source inputs compare equal")
			}
			hashes[hash] = true
			request := got.requestValue()
			if request.CoordinationSourceRecordID != nil {
				*request.CoordinationSourceRecordID = uuid.Nil
				if *got.requestValue().CoordinationSourceRecordID == uuid.Nil {
					t.Fatal("mutable source escaped admission")
				}
			}
		}
		for _, input := range []string{`""`, `123`, `[]`, `{}`, `" ` + source.String() + `"`, `"urn:uuid:` + source.String() + `"`} {
			if _, err := AdmitCreate(LessonViewSchemaID, strings.NewReader(base+`,"coordination.source_record_id":`+input+`}`)); err == nil || err.field != coordinationSourceInput {
				t.Fatalf("invalid source accepted: %s, %v", input, err)
			}
		}
		if _, err := AdmitCreate(NotesViewSchemaID, strings.NewReader(`{"client_txn_id":"txn","note.title":"Note","coordination.source_record_id":null}`)); err == nil {
			t.Fatal("source input widened Notes admission")
		}
	})

	t.Run("Note create optional clearing and Unicode stay within declared contracts", func(t *testing.T) {
		for _, title := range []string{``, `,"note.title":""`, `,"note.title":null`} {
			admitted, err := AdmitContextualNote(strings.NewReader(`{"client_txn_id":"note-clear","note.body":"Body"` + title + `}`))
			if err != nil {
				t.Fatal(err)
			}
			if _, present := admitted.requestValue().Values["note.title"]; present {
				t.Fatal("empty create title was retained as a stored scalar")
			}
		}
		admitted, err := AdmitContextualNote(strings.NewReader(`{"client_txn_id":"note-unicode","note.title":"Format\u200dtext","note.body":"Body\u200dtext","note.tags":{"kind":"collection_actions_v1","actions":[{"op":"add_tag","tag_name":"Tag\u200dtext"}]}}`))
		if err != nil {
			t.Fatal(err)
		}
		if got := *admitted.requestValue().Values["note.title"].Text; got != "Format\u200dtext" {
			t.Fatalf("format character changed: %q", got)
		}
	})
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
		if action.RiskRefText != "Preserve Display Text" || action.NormalizedText != "" {
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

func requireArtifactBoundTextAdmission(t *testing.T) {
	const actor = "10000000-0000-4000-8000-000000000001"
	bases := map[string]map[string]any{
		NotesViewSchemaID:                {"note.title": "Title", "note.body": "Body"},
		CommLogViewSchemaID:              {"comm_log.comm_type": "briefing", "comm_log.audience": "Audience", "comm_log.channel_or_meeting": "Bridge", "comm_log.summary": "Summary"},
		HandoffViewSchemaID:              {"handoff.incoming_owner_user_id": actor, "handoff.current_state_summary": "Summary"},
		StatusReviewViewSchemaID:         {"status_review.current_state_summary": "Summary"},
		LessonViewSchemaID:               {"lesson.summary": "Summary"},
		FindingsViewSchemaID:             {"finding.statement": "Statement"},
		InvestigativeQueriesViewSchemaID: {"investigative_query.platform": "Platform", "investigative_query.purpose": "Purpose", "investigative_query.query_text": "Query"},
		ForensicKeywordsViewSchemaID:     {"forensic_keyword.pattern": "Pattern", "forensic_keyword.reason": "Reason"},
	}
	// Independent field bindings and scalar limits, not derived from the registry
	// under test. Notes keep optional bindings on existing-row writes as well.
	for _, binding := range []struct {
		view, field          string
		limit                int
		multiline, clearable bool
	}{
		{NotesViewSchemaID, "note.title", 512, false, true},
		{NotesViewSchemaID, "note.body", 16384, true, true},
		{CommLogViewSchemaID, "comm_log.audience", 256, false, false},
		{CommLogViewSchemaID, "comm_log.channel_or_meeting", 512, false, false},
		{CommLogViewSchemaID, "comm_log.summary", 512, false, false},
		{CommLogViewSchemaID, "comm_log.privilege_tag", 512, false, true},
		{HandoffViewSchemaID, "handoff.current_state_summary", 16384, true, false},
		{HandoffViewSchemaID, "handoff.next_checks", 16384, true, true},
		{StatusReviewViewSchemaID, "status_review.current_state_summary", 16384, true, false},
		{StatusReviewViewSchemaID, "status_review.active_risks_summary", 16384, true, true},
		{LessonViewSchemaID, "lesson.summary", 512, false, false},
		{FindingsViewSchemaID, "finding.statement", 16384, true, false},
		{InvestigativeQueriesViewSchemaID, "investigative_query.platform", 512, false, false},
		{InvestigativeQueriesViewSchemaID, "investigative_query.purpose", 512, false, false},
		{InvestigativeQueriesViewSchemaID, "investigative_query.query_text", 16384, true, false},
		{ForensicKeywordsViewSchemaID, "forensic_keyword.pattern", 512, false, false},
		{ForensicKeywordsViewSchemaID, "forensic_keyword.reason", 4096, true, false},
	} {
		t.Run(binding.field, func(t *testing.T) {
			class := "text_compare_merge"
			if binding.field == "comm_log.privilege_tag" {
				class = "atomic_replace"
			}
			claims := ConflictAdmissionContext{
				Version: 3, RecordID: uuid.MustParse(actor), ViewSchemaID: binding.view,
				RouteKey: "workbook.records.conflicts.resolve", FieldKey: binding.field, ConflictResolutionClass: class,
				BaseRowVersion: 1, CurrentRowVersion: 2, OriginalRequestHash: "original-hash",
				IssuedAt: time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC), ExpiresAt: time.Date(2026, 7, 30, 12, 30, 0, 0, time.UTC),
			}
			admit := func(route string, raw any) (*string, []byte, *AdmissionError) {
				encode := func(value any) *bytes.Reader {
					data, err := json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
					return bytes.NewReader(data)
				}
				switch route {
				case "create":
					fields := maps.Clone(bases[binding.view])
					fields["client_txn_id"] = "text-create"
					fields[binding.field] = raw
					a, err := AdmitCreate(binding.view, encode(fields))
					if err != nil {
						return nil, nil, err
					}
					return a.requestValue().Values[binding.field].Text, a.requestHash(), nil
				case "patch":
					a, err := AdmitPatch(encode(map[string]any{"view_schema_id": binding.view, "base_row_version": 1, "client_txn_id": "text-patch", "changes": []any{map[string]any{"field_key": binding.field, "value": raw}}}))
					if err != nil {
						return nil, nil, err
					}
					return a.requestValue().Changes[0].Value.Text, a.requestHash(), nil
				default:
					a, err := AdmitConflictResolution(encode(map[string]any{"conflict_token": "token", "resolution_kind": "merged_value", "client_txn_id": "text-conflict", "resolved_value": raw}), "token", claims)
					if err != nil {
						return nil, nil, err
					}
					return a.requestValue().Patch.Changes[0].Value.Text, a.requestHash(), nil
				}
			}
			controls := []struct{ input, want string }{
				{"\u2003 e\u0301  A\u200dB \u2003", "é  A\u200dB"},
				{strings.Repeat("😀", binding.limit), strings.Repeat("😀", binding.limit)},
				{" " + strings.Repeat("e\u0301", binding.limit) + " ", strings.Repeat("é", binding.limit)},
			}
			if binding.multiline {
				controls = append(controls, struct{ input, want string }{" \r\ne\u0301\r\n\tsecond\rthird\n ", "é\n\tsecond\nthird"})
			}
			for _, route := range []string{"create", "patch", "conflict"} {
				t.Run(route, func(t *testing.T) {
					for i, tc := range controls {
						got, hash, err := admit(route, tc.input)
						if err != nil || got == nil || *got != tc.want {
							t.Fatalf("control %d: value=%v error=%v", i, got, err)
						}
						canonical, canonicalHash, err := admit(route, tc.want)
						if err != nil || canonical == nil || *canonical != tc.want || !bytes.Equal(hash, canonicalHash) {
							t.Fatalf("control %d canonical equality: %v", i, err)
						}
					}
					rejects := []any{strings.Repeat("😀", binding.limit+1), strings.Repeat("e\u0301", binding.limit+1), true, 42, []any{}, map[string]any{}}
					for r := rune(0); r <= 159; r++ {
						if r > 31 && r < 127 {
							continue
						}
						if binding.multiline && (r == '\n' || r == '\r' || r == '\t') {
							continue
						}
						rejects = append(rejects, "before"+string(r)+"after", string(r)+"before")
					}
					for i, raw := range rejects {
						_, _, err := admit(route, raw)
						if err == nil || err.field != binding.field {
							t.Fatalf("rejected input %d admitted or wrong field: %v", i, err)
						}
					}
					var clearHash []byte
					for _, raw := range []any{nil, "", " \u2003 "} {
						got, hash, err := admit(route, raw)
						optional := binding.clearable || (route == "create" && binding.view == NotesViewSchemaID)
						if !optional {
							if err == nil || err.field != binding.field {
								t.Fatalf("required empty input accepted: %v", err)
							}
							continue
						}
						if err != nil || got != nil {
							t.Fatalf("clear value=%v error=%v", got, err)
						}
						if clearHash != nil && !bytes.Equal(clearHash, hash) {
							t.Fatal("null and normalized-empty hash differently")
						}
						clearHash = hash
					}
				})
			}
		})
	}
}

func requireArtifactRiskTextAdmission(t *testing.T) {
	const actor = "10000000-0000-4000-8000-000000000001"
	const field = "handoff.open_risk_refs"
	claims := ConflictAdmissionContext{
		Version: 3, RecordID: uuid.MustParse(actor), ViewSchemaID: HandoffViewSchemaID,
		RouteKey: "workbook.records.conflicts.resolve", FieldKey: field, ConflictResolutionClass: "collection_review",
		BaseRowVersion: 1, CurrentRowVersion: 2, OriginalRequestHash: "original-hash",
		IssuedAt: time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC), ExpiresAt: time.Date(2026, 7, 30, 12, 30, 0, 0, time.UTC),
	}
	for _, route := range []string{"create", "patch", "conflict"} {
		t.Run(route, func(t *testing.T) {
			admit := func(raw any) (collectionAction, []byte, *AdmissionError) {
				encode := func(v any) *bytes.Reader {
					data, err := json.Marshal(v)
					if err != nil {
						t.Fatal(err)
					}
					return bytes.NewReader(data)
				}
				payload := map[string]any{"kind": "collection_actions_v1", "actions": []any{map[string]any{"op": "add_risk_ref", "risk_ref_text": raw}}}
				switch route {
				case "create":
					a, err := AdmitCreate(HandoffViewSchemaID, encode(map[string]any{"client_txn_id": "risk-create", "handoff.incoming_owner_user_id": actor, "handoff.current_state_summary": "Summary", field: payload}))
					if err != nil {
						return collectionAction{}, nil, err
					}
					return a.requestValue().Collections[field].Actions[0], a.requestHash(), nil
				case "patch":
					a, err := AdmitPatch(encode(map[string]any{"view_schema_id": HandoffViewSchemaID, "base_row_version": 1, "client_txn_id": "risk-patch", "changes": []any{map[string]any{"field_key": field, "action_payload": payload}}}))
					if err != nil {
						return collectionAction{}, nil, err
					}
					return a.requestValue().Changes[0].Collection.Actions[0], a.requestHash(), nil
				default:
					a, err := AdmitConflictResolution(encode(map[string]any{"conflict_token": "token", "resolution_kind": "merged_value", "client_txn_id": "risk-conflict", "resolved_value": payload}), "token", claims)
					if err != nil {
						return collectionAction{}, nil, err
					}
					return a.requestValue().Patch.Changes[0].Collection.Actions[0], a.requestHash(), nil
				}
			}
			upper, upperHash, upperErr := admit("Canonical Risk")
			lower, lowerHash, lowerErr := admit("canonical risk")
			if upperErr != nil || lowerErr != nil || upper.RiskRefText != "Canonical Risk" || lower.RiskRefText != "canonical risk" || bytes.Equal(upperHash, lowerHash) {
				t.Fatalf("case-distinct risk replay collapsed: %v / %v", upperErr, lowerErr)
			}
			// Keep all other seed members identical so distinct request hashes cannot
			// mask a missing or folded canonical risk value in the conflict preview.
			upperID := conflictLocalUUID(claims.RecordID, field, upper, []byte("same-request"), 0)
			lowerID := conflictLocalUUID(claims.RecordID, field, lower, []byte("same-request"), 0)
			if upperID == lowerID || upperID != conflictLocalUUID(claims.RecordID, field, upper, []byte("same-request"), 0) {
				t.Fatal("conflict preview risk identity is collapsed or unstable")
			}
			for i, tc := range []struct{ input, want string }{
				{"\u2003 e\u0301  Risk\u200d \u2003", "é  Risk\u200d"},
				{strings.Repeat("😀", 512), strings.Repeat("😀", 512)},
				{" " + strings.Repeat("e\u0301", 512) + " ", strings.Repeat("é", 512)},
			} {
				got, hash, err := admit(tc.input)
				if err != nil || got.RiskRefText != tc.want {
					t.Fatalf("control %d: value=%q error=%v", i, got.RiskRefText, err)
				}
				canonical, expectedHash, err := admit(tc.want)
				if err != nil || canonical.RiskRefText != tc.want || !bytes.Equal(hash, expectedHash) {
					t.Fatalf("control %d canonical equality: %v", i, err)
				}
			}
			rejects := []any{nil, "", " \u2003 ", strings.Repeat("😀", 513), strings.Repeat("e\u0301", 513), true, 42, []any{}, map[string]any{}}
			for r := rune(0); r <= 159; r++ {
				if r > 31 && r < 127 {
					continue
				}
				rejects = append(rejects, "before"+string(r)+"after", string(r)+"before")
			}
			for i, raw := range rejects {
				_, _, err := admit(raw)
				if err == nil || err.field != field || err.ReasonCode() != admissionInvalidValue {
					t.Fatalf("rejected input %d admitted or wrong error: %#v", i, err)
				}
			}
		})
	}
}
