package artifacts

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func requireNoteTagLabelAdmission(t *testing.T) {
	for _, tc := range []struct {
		name, raw, label, folded string
		valid                    bool
	}{
		{"format and NFC", " \u2003Cafe\u0301\u200d  TAG ", "Café\u200d  TAG", "café\u200d  tag", true},
		{"scalar maximum", strings.Repeat("界", 63) + "\u200d", strings.Repeat("界", 63) + "\u200d", strings.Repeat("界", 63) + "\u200d", true},
		{"too long", strings.Repeat("界", 64) + "\u200d", "", "", false},
		{"empty", " \u2003 ", "", "", false},
		{"leading control", "\u000bTag", "", "", false},
		{"interior control", "Ta\u0085g", "", "", false},
		{"trailing control", "Tag\u0085", "", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(tc.raw)
			if err != nil {
				t.Fatal(err)
			}
			payload := fmt.Sprintf(`{"kind":"collection_actions_v1","actions":[{"op":"add_tag","tag_name":%s}]}`, encoded)
			createBody := `{"client_txn_id":"tag-create","note.title":"Note","note.tags":` + payload + `}`
			for _, route := range []string{"create", "contextual", "patch", "conflict"} {
				t.Run(route, func(t *testing.T) {
					var actions []collectionAction
					var failure *AdmissionError
					switch route {
					case "create":
						admitted, err := AdmitCreate(NotesViewSchemaID, strings.NewReader(createBody))
						failure = err
						actions = admitted.requestValue().Collections["note.tags"].Actions
					case "contextual":
						admitted, err := AdmitContextualNote(strings.NewReader(createBody))
						failure = err
						actions = admitted.requestValue().Collections["note.tags"].Actions
					case "patch":
						admitted, err := AdmitPatch(strings.NewReader(`{"view_schema_id":"cartulary.view.notes.v1","base_row_version":1,"client_txn_id":"tag-patch","changes":[{"field_key":"note.tags","action_payload":` + payload + `}]}`))
						failure = err
						if failure == nil {
							changes := admitted.requestValue().Changes
							if len(changes) != 1 || changes[0].Collection == nil {
								t.Fatalf("tag patch shape: %#v", changes)
							}
							actions = changes[0].Collection.Actions
						}
					case "conflict":
						claims := ConflictAdmissionContext{
							Version: 3, RecordID: uuid.MustParse("11111111-2222-4333-8444-555555555555"), ViewSchemaID: NotesViewSchemaID,
							RouteKey: "workbook.records.conflicts.resolve", FieldKey: "note.tags", ConflictResolutionClass: "collection_review",
							BaseRowVersion: 1, CurrentRowVersion: 2, OriginalRequestHash: "original-hash",
							IssuedAt: time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC), ExpiresAt: time.Date(2026, 7, 30, 12, 30, 0, 0, time.UTC),
						}
						admitted, err := AdmitConflictResolution(strings.NewReader(`{"conflict_token":"token","resolution_kind":"merged_value","client_txn_id":"tag-conflict","resolved_value":`+payload+`}`), "token", claims)
						failure = err
						if failure == nil {
							patch := admitted.requestValue().Patch
							if patch == nil || len(patch.Changes) != 1 || patch.Changes[0].Collection == nil {
								t.Fatalf("tag conflict patch shape: %#v", patch)
							}
							actions = patch.Changes[0].Collection.Actions
						}
					}
					if !tc.valid {
						if failure == nil || failure.field != "note.tags" || failure.ReasonCode() != admissionInvalidValue {
							t.Fatalf("tag rejection=%#v; want note.tags/invalid_value", failure)
						}
						return
					}
					if failure != nil || len(actions) != 1 || actions[0].Op != "add_tag" || actions[0].RawText != tc.label || actions[0].NormalizedText != tc.folded {
						t.Fatalf("tag admission=%#v failure=%v; want %q/%q", actions, failure, tc.label, tc.folded)
					}
				})
			}
		})
	}
}
