package tasksdecisions

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestTaskDecisionMutationAdmissionAndReplayHashing(t *testing.T) {
	t.Run("bound source text across entry paths", requireTaskDecisionBoundTextAdmission)
	t.Run("record-reference confidence is rejected by each create and patch field", func(t *testing.T) {
		for _, tc := range []struct{ view, field, values string }{
			{TaskRequestsViewSchemaID, "task.linked_record_ids", `"task.title":"Task","task.task_kind":"collection"`},
			{DecisionsViewSchemaID, "decision.support_refs", `"decision.summary":"Decision","decision.decision_type":"containment","decision.rationale":"Rationale"`},
			{DecisionsViewSchemaID, "decision.affected_record_ids", `"decision.summary":"Decision","decision.decision_type":"containment","decision.rationale":"Rationale"`},
		} {
			t.Run(tc.field, func(t *testing.T) {
				for _, forbidden := range []bool{false, true} {
					extra := ""
					if forbidden {
						extra = `,"confidence":80`
					}
					payload := `{"kind":"collection_actions_v1","actions":[{"op":"add_record_ref","linked_record_id":"10000000-0000-4000-8000-000000000001"` + extra + `}]}`
					create := fmt.Sprintf(`{"client_txn_id":"confidence-create",%s,%q:%s}`, tc.values, tc.field, payload)
					patch := fmt.Sprintf(`{"view_schema_id":%q,"client_txn_id":"confidence-patch","base_row_version":1,"changes":[{"field_key":%q,"action_payload":%s}]}`, tc.view, tc.field, payload)
					_, createErr := AdmitCreateJSON(tc.view, strings.NewReader(create))
					_, patchErr := AdmitPatchJSON(strings.NewReader(patch))
					for route, failure := range map[string]*AdmissionFailure{"create": createErr, "patch": patchErr} {
						if !forbidden {
							if failure != nil {
								t.Fatalf("%s valid control: %v", route, failure)
							}
						} else {
							requireAdmissionFailure(t, failure, tc.field, "invalid_value", "", 0, 0, false)
						}
					}
				}
			})
		}
	})
	t.Run("create admission owns fixed surfaces normalization and hash", func(t *testing.T) {
		request, admissionFailure := AdmitCreateJSON(TaskRequestsViewSchemaID, strings.NewReader(`{
			"client_txn_id":"txn-task-admission",
			"task.title":"  Collect   endpoint logs  ",
			"task.task_kind":"collection"
		}`))
		if admissionFailure != nil {
			t.Fatalf("admit task create: %v", admissionFailure)
		}
		if got := *request.Values["task.title"].Text; got != "Collect   endpoint logs" {
			t.Fatalf("normalized title = %q", got)
		}
		reordered := CreateRequest{
			ViewSchemaID: TaskRequestsViewSchemaID, ClientTxnID: "ignored-by-hash",
			Values: map[string]FieldValue{
				"task.task_kind": request.Values["task.task_kind"],
				"task.title":     request.Values["task.title"],
			},
			Collections: map[string]CollectionActionPayload{},
		}
		if !slices.Equal(CreateRequestHash(request), CreateRequestHash(reordered)) {
			t.Fatal("create hash depends on map order or client transaction ID")
		}
		requireHashGolden(t, CreateRequestHash(request), "d0a52c09050580f048fc083dc38fbd1a721636fa4b8f99c1acaab332599e1c51")
		changed := reordered
		changed.Values = map[string]FieldValue{
			"task.task_kind": request.Values["task.task_kind"],
			"task.title":     {Text: stringPointer("Collect different endpoint logs")},
		}
		if slices.Equal(CreateRequestHash(request), CreateRequestHash(changed)) {
			t.Fatal("meaningfully different normalized create values hash equally")
		}
		if _, admissionFailure := AdmitCreateJSON("cartulary.view.notes.v1", strings.NewReader(`{}`)); admissionFailure == nil {
			t.Fatal("Task/Decision admission accepted a foreign view")
		}
	})

	t.Run("patch and conflict admission own record-reference collections", func(t *testing.T) {
		recordID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
		body := `{
			"view_schema_id":"cartulary.view.decisions.v1",
			"base_row_version":2,
			"client_txn_id":"txn-decision-patch",
			"changes":[{"field_key":"decision.support_refs","action_payload":{
				"kind":"collection_actions_v1",
				"actions":[{"op":"add_record_ref","linked_record_id":"` + recordID.String() + `"}]
			}}]
		}`
		request, admissionFailure := AdmitPatchJSON(strings.NewReader(body))
		if admissionFailure != nil {
			t.Fatalf("admit decision patch: %v", admissionFailure)
		}
		change := request.Changes[0]
		if change.Collection == nil || *change.Collection.Actions[0].LinkedRecordID != recordID {
			t.Fatalf("unexpected collection admission: %#v", change)
		}
		claims := ConflictClaims{
			RecordID: uuid.MustParse("20000000-0000-4000-8000-000000000002"), ViewSchemaID: DecisionsViewSchemaID,
			FieldKey: "decision.rationale", CurrentRowVersion: 3,
		}
		conflict, admissionFailure := AdmitConflictResolveJSON(
			strings.NewReader(`{
				"conflict_token":"opaque-token",
				"resolution_kind":"merged_value",
				"client_txn_id":"txn-decision-resolve",
				"resolved_value":"  Combined rationale.  "
			}`),
			"opaque-token",
			claims,
		)
		if admissionFailure != nil {
			t.Fatalf("admit decision conflict: %v", admissionFailure)
		}
		if conflict.Patch == nil || *conflict.Patch.Changes[0].Value.Text != "Combined rationale." {
			t.Fatalf("unexpected conflict admission: %#v", conflict)
		}
		if len(ConflictResolveRequestHash(claims, conflict)) != 32 || len(PatchRequestHash(request)) != 32 {
			t.Fatal("owner hashes are not SHA-256 values")
		}
		requireHashGolden(t, PatchRequestHash(request), "f15969c36cac7c6e7f6af111dbdcced3ca22e5b191705ed8aa68f550a910d9b3")
		requireHashGolden(t, ConflictResolveRequestHash(claims, conflict), "1b3c4a5b84e02adef381b72af19c46d4876eec6aac7c0cfd03c43acc529c0763")
	})

	t.Run("decision supersede admission and hash are owner local", func(t *testing.T) {
		replacement := uuid.MustParse("30000000-0000-4000-8000-000000000003")
		body := []byte(`{
			"base_row_version":4,
			"client_txn_id":"txn-decision-supersede",
			"reason":"  Use the better-supported decision.  ",
			"replacement_record_id":"` + replacement.String() + `"
		}`)
		request, admissionFailure := AdmitSupersedeJSON(bytes.NewReader(body))
		if admissionFailure != nil {
			t.Fatalf("admit decision supersede: %v", admissionFailure)
		}
		if request.Reason != "Use the better-supported decision." || *request.ReplacementRecordID != replacement {
			t.Fatalf("unexpected supersede admission: %#v", request)
		}
		if len(SupersedeRequestHash(request)) != 32 {
			t.Fatal("supersede hash is not SHA-256")
		}
		requireHashGolden(t, SupersedeRequestHash(request), "6f287bdccd47088f99cda59cc956ce80890b34694f81d213e915b6442c89877d")
	})

	t.Run("Decision supersession reason follows reason_note_v1", func(t *testing.T) {
		t.Run("rejected controls cannot disappear at boundaries", func(t *testing.T) {
			for control := rune(0); control <= '\u009f'; control++ {
				if (control > '\u001f' && control < '\u007f') || control == '\t' || control == '\n' || control == '\r' {
					continue
				}
				for position, raw := range []string{string(control) + "Reason", "Rea" + string(control) + "son", "Reason" + string(control)} {
					t.Run(fmt.Sprintf("U%04X-position%d", control, position), func(t *testing.T) {
						encoded, err := json.Marshal(raw)
						if err != nil {
							t.Fatal(err)
						}
						_, failure := AdmitSupersedeJSON(strings.NewReader(`{"base_row_version":4,"client_txn_id":"control-reason","reason":` + string(encoded) + `}`))
						requireAdmissionFailure(t, failure, "reason", "invalid_value", "", 0, 0, false)
					})
				}
			}
		})
		for _, test := range []struct {
			reason     string
			accepted   bool
			normalized string
		}{
			{strings.Repeat("界", 4096), true, strings.Repeat("界", 4096)},
			{strings.Repeat("界", 4097), false, ""},
			{" \u2003Cafe\u0301\r\nnext\tline\u200d ", true, "Café\nnext\tline\u200d"},
			{"\tReason\rnext\t \n", true, "Reason\nnext"},
			{"hidden\u0081control", false, ""},
			{" \n ", false, ""},
		} {
			reason, err := json.Marshal(test.reason)
			if err != nil {
				t.Fatal(err)
			}
			request, failure := AdmitSupersedeJSON(strings.NewReader(`{"base_row_version":4,"client_txn_id":"reason-check","reason":` + string(reason) + `}`))
			if (failure == nil) != test.accepted {
				t.Fatalf("reason admission: expected accepted=%v, failure=%v, scalar count=%d", test.accepted, failure, len([]rune(test.reason)))
			}
			if failure == nil && request.Reason != test.normalized {
				t.Fatalf("reason normalization mismatch: %q", request.Reason)
			}
		}
	})

	t.Run("admission failures expose closed semantic detail variants", func(t *testing.T) {
		_, plain := AdmitCreateJSON("cartulary.view.notes.v1", strings.NewReader(`{}`))
		requireAdmissionFailure(t, plain, "view_schema_id", "unknown_view_schema", "", 0, 0, false)
		if got := plain.Error(); got != "tasksdecisions: invalid mutation admission" {
			t.Fatalf("safe admission error = %q", got)
		}

		emptyCollection := `{
			"view_schema_id":"cartulary.view.decisions.v1",
			"base_row_version":1,
			"client_txn_id":"txn-empty-collection",
			"changes":[{"field_key":"decision.support_refs","action_payload":{
				"kind":"collection_actions_v1","actions":[]
			}}]
		}`
		_, empty := AdmitPatchJSON(strings.NewReader(emptyCollection))
		requireAdmissionFailure(
			t, empty, "decision.support_refs.actions", "empty_collection_actions",
			"decision.support_refs", 0, 0, false,
		)

		patchLimit := `{
			"view_schema_id":"cartulary.view.task_requests.v1",
			"base_row_version":1,
			"client_txn_id":"txn-patch-limit",
			"changes":[` + repeatedJSONObjectMembers(33) + `]
		}`
		_, patchCount := AdmitPatchJSON(strings.NewReader(patchLimit))
		requireAdmissionFailure(t, patchCount, "changes", "change_count_exceeded", "", 33, 32, true)

		collectionLimit := `{
			"view_schema_id":"cartulary.view.decisions.v1",
			"base_row_version":1,
			"client_txn_id":"txn-collection-limit",
			"changes":[{"field_key":"decision.support_refs","action_payload":{
				"kind":"collection_actions_v1","actions":[` + repeatedJSONObjectMembers(65) + `]
			}}]
		}`
		_, collectionCount := AdmitPatchJSON(strings.NewReader(collectionLimit))
		requireAdmissionFailure(
			t, collectionCount, "decision.support_refs.actions", "collection_action_count_exceeded",
			"decision.support_refs", 65, 64, true,
		)
	})
}

func requireTaskDecisionBoundTextAdmission(t *testing.T) {
	t.Helper()
	// Field bindings and limits are independent expectations, not registry output.
	for _, binding := range []struct {
		view, field         string
		limit               int
		multiline, optional bool
	}{
		{TaskRequestsViewSchemaID, "task.title", 512, false, false},
		{TaskRequestsViewSchemaID, "task.workstream", 512, false, true},
		{TaskRequestsViewSchemaID, "task.requester_party_text", 256, false, true},
		{TaskRequestsViewSchemaID, "task.blocked_reason", 4096, true, true},
		{TaskRequestsViewSchemaID, "task.external_ticket_ref", 1024, false, true},
		{TaskRequestsViewSchemaID, "task.closure_summary", 16384, true, true},
		{DecisionsViewSchemaID, "decision.summary", 512, false, false},
		{DecisionsViewSchemaID, "decision.rationale", 16384, true, false},
	} {
		t.Run(binding.field, func(t *testing.T) {
			encode := func(value any) *bytes.Reader {
				t.Helper()
				data, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				return bytes.NewReader(data)
			}
			admit := func(route string, raw any) (*string, []byte, *AdmissionFailure) {
				switch route {
				case "create":
					values := map[string]any{"client_txn_id": "bound-text", "task.title": "Task", "task.task_kind": "request"}
					if binding.view == DecisionsViewSchemaID {
						values = map[string]any{"client_txn_id": "bound-text", "decision.summary": "Decision", "decision.decision_type": "scope", "decision.rationale": "Rationale"}
					}
					values[binding.field] = raw
					request, failure := AdmitCreateJSON(binding.view, encode(values))
					if failure != nil {
						return nil, nil, failure
					}
					return request.Values[binding.field].Text, CreateRequestHash(request), nil
				case "patch":
					request, failure := AdmitPatchJSON(encode(map[string]any{"view_schema_id": binding.view, "base_row_version": 1, "client_txn_id": "bound-text", "changes": []map[string]any{{"field_key": binding.field, "value": raw}}}))
					if failure != nil {
						return nil, nil, failure
					}
					return request.Changes[0].Value.Text, PatchRequestHash(request), nil
				case "conflict":
					claims := ConflictClaims{RecordID: uuid.MustParse("10000000-0000-4000-8000-000000000001"), ViewSchemaID: binding.view, FieldKey: binding.field, CurrentRowVersion: 2}
					request, failure := AdmitConflictResolveJSON(encode(map[string]any{"conflict_token": "text-token", "resolution_kind": "use_unsaved", "client_txn_id": "bound-text", "resolved_value": raw}), "text-token", claims)
					if failure != nil {
						return nil, nil, failure
					}
					return request.Patch.Changes[0].Value.Text, ConflictResolveRequestHash(claims, request), nil
				default:
					t.Fatalf("unknown route %s", route)
					return nil, nil, nil
				}
			}
			for _, route := range []string{"create", "patch", "conflict"} {
				t.Run(route, func(t *testing.T) {
					cases := []struct{ raw, want string }{{" \u2003Cafe\u0301  source\u200d ", "Café  source\u200d"}, {strings.Repeat("😀", binding.limit), strings.Repeat("😀", binding.limit)}, {strings.Repeat("e\u0301", binding.limit), strings.Repeat("é", binding.limit)}}
					if binding.multiline {
						cases = append(cases, struct{ raw, want string }{" \r\nBefore\r\n\tAfter\rLast\n ", "Before\n\tAfter\nLast"})
					}
					for _, tc := range cases {
						got, hash, failure := admit(route, tc.raw)
						if failure != nil || got == nil || *got != tc.want {
							t.Fatalf("normalized %s: got=%v failure=%v", binding.field, got, failure)
						}
						canonical, canonicalHash, failure := admit(route, tc.want)
						if failure != nil || canonical == nil || *canonical != tc.want || !bytes.Equal(hash, canonicalHash) {
							t.Fatal("canonical and normalized request differ")
						}
					}
					_, firstHash, failure := admit(route, "One source")
					if failure != nil {
						t.Fatal(failure)
					}
					_, secondHash, failure := admit(route, "Different source")
					if failure != nil || bytes.Equal(firstHash, secondHash) {
						t.Fatal("different source hashes compare equally")
					}
					invalid := []string{strings.Repeat("😀", binding.limit+1)}
					for r := rune(0); r <= 159; r++ {
						if r > 31 && r < 127 {
							continue
						}
						if binding.multiline && (r == '\n' || r == '\t' || r == '\r') {
							continue
						}
						invalid = append(invalid, "before"+string(r)+"after", string(r)+"value", "value"+string(r))
					}
					for _, raw := range invalid {
						_, _, failure := admit(route, raw)
						requireAdmissionFailure(t, failure, binding.field, "invalid_value", "", 0, 0, false)
					}
					var clearHash []byte
					for _, raw := range []any{nil, "", " \u2003 "} {
						got, hash, failure := admit(route, raw)
						if binding.optional {
							if failure != nil || got != nil {
								t.Fatalf("optional clear=%v failure=%v", got, failure)
							}
							if clearHash != nil && !bytes.Equal(clearHash, hash) {
								t.Fatal("null and normalized-empty hashes differ")
							}
							clearHash = hash
						} else {
							reason := "invalid_value"
							if raw == nil {
								reason = "field_not_nullable"
							}
							requireAdmissionFailure(t, failure, binding.field, reason, "", 0, 0, false)
						}
					}
				})
			}
		})
	}
}

func requireHashGolden(t testing.TB, hash []byte, want string) {
	t.Helper()
	if got := hex.EncodeToString(hash); got != want {
		t.Fatalf("request hash = %s, want %s", got, want)
	}
}

func stringPointer(value string) *string { return &value }

func requireAdmissionFailure(
	t testing.TB,
	failure *AdmissionFailure,
	wantField string,
	wantReasonCode string,
	wantCollectionFieldKey string,
	wantRequestedCount int,
	wantMaxCount int,
	wantCountLimit bool,
) {
	t.Helper()
	if failure == nil {
		t.Fatal("expected admission failure")
	}
	if failure.Field() != wantField || failure.ReasonCode() != wantReasonCode {
		t.Fatalf(
			"admission identity = (%q, %q), want (%q, %q)",
			failure.Field(), failure.ReasonCode(), wantField, wantReasonCode,
		)
	}
	fieldKey, hasFieldKey := failure.CollectionFieldKey()
	if hasFieldKey != (wantCollectionFieldKey != "") || fieldKey != wantCollectionFieldKey {
		t.Fatalf(
			"collection field key = (%q, %t), want (%q, %t)",
			fieldKey, hasFieldKey, wantCollectionFieldKey, wantCollectionFieldKey != "",
		)
	}
	requestedCount, maxCount, hasCountLimit := failure.CountLimit()
	if hasCountLimit != wantCountLimit || requestedCount != wantRequestedCount || maxCount != wantMaxCount {
		t.Fatalf(
			"count limit = (%d, %d, %t), want (%d, %d, %t)",
			requestedCount, maxCount, hasCountLimit, wantRequestedCount, wantMaxCount, wantCountLimit,
		)
	}
}

func repeatedJSONObjectMembers(count int) string {
	return strings.TrimSuffix(strings.Repeat(`{},`, count), ",")
}
