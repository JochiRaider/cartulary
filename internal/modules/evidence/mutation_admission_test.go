package evidence

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEvidenceMutationAdmissionAndReplayHashing_Unit(t *testing.T) {
	requireEvidenceBoundTextAdmission(t)
	t.Run("create normalization and default are canonical", func(t *testing.T) {
		omitted, failure := AdmitCreateJSON(strings.NewReader(`{
			"client_txn_id":"txn-evidence-create",
			"evidence.title":"  Disk image  "
		}`))
		if failure != nil {
			t.Fatalf("decode omitted-default create: %#v", failure)
		}
		explicit, failure := AdmitCreateJSON(strings.NewReader(`{
			"client_txn_id":"txn-evidence-create-other",
			"evidence.lifecycle_state":"requested",
			"evidence.title":"Disk image"
		}`))
		if failure != nil {
			t.Fatalf("decode explicit-default create: %#v", failure)
		}
		if got := *omitted.requestValue().Values["evidence.title"].Text; got != "Disk image" {
			t.Fatalf("normalized title got %q", got)
		}
		omittedHash := hex.EncodeToString(omitted.requestHash())
		explicitHash := hex.EncodeToString(explicit.requestHash())
		const want = "29e3f59295df8149b5ada9c51cdde43dcfb30469a7a5c546af274a1e1e9a4df4"
		if omittedHash != want || explicitHash != want {
			t.Fatalf("create hashes got omitted=%s explicit=%s want=%s", omittedHash, explicitHash, want)
		}
	})

	t.Run("patch changes sort before hashing", func(t *testing.T) {
		admission, failure := AdmitPatchJSON(strings.NewReader(`{
			"view_schema_id":"cartulary.view.evidence.v1",
			"base_row_version":3,
			"client_txn_id":"txn-evidence-patch",
			"changes":[
				{"field_key":"evidence.title","value":"Disk image"},
				{"field_key":"evidence.lifecycle_state","value":"received"}
			]
		}`))
		if failure != nil {
			t.Fatalf("decode patch: %#v", failure)
		}
		request := admission.requestValue()
		if request.Changes[0].FieldKey != "evidence.lifecycle_state" || request.Changes[1].FieldKey != "evidence.title" {
			t.Fatalf("patch changes are not canonical: %#v", request.Changes)
		}
		const want = "2687e07b9b9fd93ee54cfb6051cd6ff1224f523230c0e7dec6dafa8fbec61153"
		if got := hex.EncodeToString(admission.requestHash()); got != want {
			t.Fatalf("patch hash got %s want %s", got, want)
		}
	})

	t.Run("conflict claims and keep-saved hash stay owner-owned", func(t *testing.T) {
		now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
		claims := ConflictAdmissionContext{
			Version:                 2,
			RecordID:                uuid.MustParse("00000000-0000-4000-8000-000000210101"),
			ViewSchemaID:            ViewSchemaID,
			RouteKey:                string(OperationConflictResolve),
			FieldKey:                "evidence.title",
			ConflictResolutionClass: "text_compare_merge",
			BaseRowVersion:          3,
			CurrentRowVersion:       4,
			OriginalRequestHash:     "opaque-original-hash",
			IssuedAt:                now,
			ExpiresAt:               now.Add(time.Hour),
		}
		admission, failure := AdmitConflictResolveJSON(
			strings.NewReader(`{
				"conflict_token":"opaque-token",
				"resolution_kind":"keep_saved",
				"client_txn_id":"txn-evidence-conflict"
			}`),
			"opaque-token",
			claims,
		)
		if failure != nil {
			t.Fatalf("decode conflict resolution: %#v", failure)
		}
		const want = "5bf28191f2c841cb695adbf505fff2ea22c50a0b3f7e143c35276688c430ad5e"
		if got := hex.EncodeToString(admission.requestHash()); got != want {
			t.Fatalf("conflict hash got %s want %s", got, want)
		}
	})

	t.Run("initial blob input is closed and typed", func(t *testing.T) {
		for name, body := range map[string]string{
			"null":      `{"client_txn_id":"txn","evidence.title":"Disk image","evidence.initial_object_blob_id":null}`,
			"malformed": `{"client_txn_id":"txn","evidence.title":"Disk image","evidence.initial_object_blob_id":"not-a-uuid"}`,
			"unknown":   `{"client_txn_id":"txn","evidence.title":"Disk image","evidence.unowned":true}`,
		} {
			t.Run(name, func(t *testing.T) {
				if _, failure := AdmitCreateJSON(strings.NewReader(body)); failure == nil {
					t.Fatal("admission unexpectedly succeeded")
				}
			})
		}
	})
}

func requireEvidenceBoundTextAdmission(t *testing.T) {
	t.Helper()
	for _, field := range []struct {
		key      string
		limit    int
		conflict string
	}{
		{"evidence.title", 512, "text_compare_merge"},
		{"evidence.storage_ref", 1024, "atomic_replace"},
		{"evidence.collector_party_text", 256, "text_compare_merge"},
		{"evidence.source_party_text", 256, "text_compare_merge"},
	} {
		for _, mode := range []string{"create", "patch", "conflict"} {
			t.Run(field.key+" "+mode, func(t *testing.T) {
				admit := func(value any) (FieldValue, []byte, *AdmissionFailure) {
					t.Helper()
					body := map[string]any{"client_txn_id": "bound-text", field.key: value}
					if mode == "patch" {
						body = map[string]any{"client_txn_id": "bound-text", "view_schema_id": ViewSchemaID, "base_row_version": 1, "changes": []any{map[string]any{"field_key": field.key, "value": value}}}
					}
					if mode == "conflict" {
						body = map[string]any{"client_txn_id": "bound-text", "conflict_token": "opaque-token", "resolution_kind": "use_unsaved", "resolved_value": value}
					}
					encoded, err := json.Marshal(body)
					if err != nil {
						t.Fatal(err)
					}
					switch mode {
					case "create":
						result, failure := AdmitCreateJSON(bytes.NewReader(encoded))
						if failure != nil {
							return FieldValue{}, nil, failure
						}
						return result.requestValue().Values[field.key], result.requestHash(), nil
					case "patch":
						result, failure := AdmitPatchJSON(bytes.NewReader(encoded))
						if failure != nil {
							return FieldValue{}, nil, failure
						}
						return *result.requestValue().Changes[0].Value, result.requestHash(), nil
					default:
						now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
						claims := ConflictAdmissionContext{Version: 2, RecordID: uuid.MustParse("00000000-0000-4000-8000-000000000001"), ViewSchemaID: ViewSchemaID, RouteKey: string(OperationConflictResolve), FieldKey: field.key, ConflictResolutionClass: field.conflict, BaseRowVersion: 1, CurrentRowVersion: 2, OriginalRequestHash: "original", IssuedAt: now, ExpiresAt: now.Add(time.Hour)}
						result, failure := AdmitConflictResolveJSON(bytes.NewReader(encoded), "opaque-token", claims)
						if failure != nil {
							return FieldValue{}, nil, failure
						}
						return *result.requestValue().Patch.Changes[0].Value, result.requestHash(), nil
					}
				}
				for _, test := range []struct{ input, want string }{
					{" \u00a0Cafe\u0301  MiXeD\u200dText\u00a0 ", "Café  MiXeD\u200dText"},
					{strings.Repeat("😀", field.limit), strings.Repeat("😀", field.limit)},
					{strings.Repeat("e\u0301", field.limit), strings.Repeat("é", field.limit)},
				} {
					got, hash, failure := admit(test.input)
					if failure != nil || got.Text == nil || *got.Text != test.want {
						t.Fatalf("normalized value=%#v failure=%v want=%q", got, failure, test.want)
					}
					_, normalizedHash, failure := admit(test.want)
					if failure != nil || !bytes.Equal(hash, normalizedHash) {
						t.Fatal("normalized request hashes differ")
					}
					_, differentHash, failure := admit("Different text")
					if failure != nil || bytes.Equal(hash, differentHash) {
						t.Fatal("distinct source text has equal hash")
					}
				}
				_, nullHash, failure := admit(nil)
				if failure != nil {
					t.Fatal(failure)
				}
				for _, value := range []any{nil, "", " \u00a0\u2003 "} {
					got, hash, failure := admit(value)
					if failure != nil || !reflect.DeepEqual(got, FieldValue{}) || !bytes.Equal(hash, nullHash) {
						t.Fatalf("clear value=%#v failure=%v", got, failure)
					}
				}
				invalid := []string{strings.Repeat("😀", field.limit+1)}
				for control := rune(0); control <= 159; control++ {
					if control > 31 && control < 127 {
						continue
					}
					invalid = append(invalid, string(control)+"Text", "Te"+string(control)+"xt", "Text"+string(control))
				}
				for _, value := range invalid {
					_, _, failure := admit(value)
					if failure == nil || failure.ReasonCode() != admissionInvalidValue {
						t.Fatalf("invalid text admitted or wrong failure: %v", failure)
					}
					key, present := failure.Field()
					if !present || key != field.key {
						t.Fatalf("failure field=%q", key)
					}
				}
			})
		}
	}
}
