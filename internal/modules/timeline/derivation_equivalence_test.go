package timeline

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/rowpresenter"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func TestCanonicalRowDerivesCompleteTypedCollections_Unit(t *testing.T) {
	recordID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	incidentID := uuid.MustParse("20000000-0000-4000-8000-000000000002")
	replacementID := uuid.MustParse("30000000-0000-4000-8000-000000000003")
	hostMentionID := uuid.MustParse("40000000-0000-4000-8000-000000000004")
	identityMentionID := uuid.MustParse("50000000-0000-4000-8000-000000000005")
	evidenceID := uuid.MustParse("60000000-0000-4000-8000-000000000006")
	tagID := uuid.MustParse("70000000-0000-4000-8000-000000000007")
	hostID := uuid.MustParse("81000000-0000-4000-8000-000000000008")
	dateEntered := "2026-07-24T23:30:00-04:00"
	activityUTC := "2026-07-25T03:30:00Z"
	activityLocal := "2026-07-24T23:30:00-04:00"
	synopsis := "canonical derivation"
	recordedAt := time.Date(2026, 7, 25, 3, 31, 0, 0, time.UTC)

	source := sourcerepository.Snapshot{
		RecordID:              recordID,
		IncidentID:            incidentID,
		DateEnteredText:       &dateEntered,
		ActivityUTCText:       &activityUTC,
		ActivityLocalText:     &activityLocal,
		ActivitySynopsisText:  &synopsis,
		ActivityTimePairState: "paired_user_preserved",
		CaptureState:          captureStateReviewed,
		RowVersion:            9,
		RecordedAt:            recordedAt,
		EditedAt:              recordedAt.Add(time.Minute),
	}
	canonicalPath := workbookprojection.Derive(source, &replacementID)
	method := "auto_match"
	alias := "Host A"
	workbookprojection.ApplyCollectionFacts(&canonicalPath, workbookprojection.CollectionFacts{
		Mentions: []workbookprojection.MentionFact{
			{
				MentionID:        hostMentionID,
				EntityType:       "host",
				SourceFieldKey:   "timeline.host_refs",
				RawText:          "host-a",
				ResolutionStatus: "resolved",
				RowVersion:       2,
				ResolvedRecordID: &hostID,
				ResolutionMethod: &method,
				MatchedAliasText: &alias,
			},
			{
				MentionID:        identityMentionID,
				EntityType:       "identity",
				SourceFieldKey:   "timeline.identity_refs",
				RawText:          "analyst@example.test",
				ResolutionStatus: "unresolved",
				RowVersion:       1,
			},
		},
		ResolvedLinks: []workbookprojection.LinkFact{{
			TargetRecordID: hostID,
			LinkType:       "observed_on_host",
			Provenance:     "auto_match",
		}},
		Tags: []workbookprojection.TagFact{{
			RecordTagID: tagID,
			TagName:     "critical",
		}},
		AttachedEvidence: []workbookprojection.EvidenceFact{{
			RecordID:       evidenceID,
			Title:          "packet.pcap",
			LifecycleState: "available",
			UploadState:    "available",
		}},
		ReplacementRecordID: &replacementID,
	})

	row := buildRow(canonicalPath)
	cells := row["cells"].(map[string]any)
	groupValues := row["group_values"].(map[string]any)
	if cells["timeline.activity_sort_ts"].(map[string]any)["value"] != "2026-07-25T03:30:00Z" {
		t.Fatalf("UTC activity value must win canonical time derivation: %#v", cells["timeline.activity_sort_ts"])
	}
	if cells["timeline.date_entered_sort_day"].(map[string]any)["value"] != "2026-07-25" {
		t.Fatalf("offset date must normalize through the time contract: %#v", cells["timeline.date_entered_sort_day"])
	}
	if groupValues["timeline.has_evidence"] != true || groupValues["timeline.has_unresolved_mentions"] != true {
		t.Fatalf("canonical group values lost collection state: %#v", groupValues)
	}
	input := canonicalPath.ProjectionInput()
	if len(input.HostRefs) != 1 || input.HostRefs[0].MatchedAliasText == nil || *input.HostRefs[0].MatchedAliasText != alias {
		t.Fatalf("typed host references lost owner facts: %#v", input.HostRefs)
	}
	if len(input.AttachedEvidence) != 1 || input.AttachedEvidence[0].LinkedRecordID != evidenceID {
		t.Fatalf("typed evidence references lost owner facts: %#v", input.AttachedEvidence)
	}
}

func TestCanonicalRowDerivationPreservesNullableFields_Unit(t *testing.T) {
	source := sourcerepository.Snapshot{
		RecordID:              uuid.MustParse("80000000-0000-4000-8000-000000000008"),
		IncidentID:            uuid.MustParse("90000000-0000-4000-8000-000000000009"),
		ActivityTimePairState: "unpaired",
		CaptureState:          captureStateRough,
		RowVersion:            1,
		RecordedAt:            time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
		EditedAt:              time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
	}
	canonicalPath := workbookprojection.Derive(source, nil)
	assertCanonicalJSONEqual(t, buildRow(canonicalPath), rowpresenter.BuildRow(canonicalPath.PresenterRecord()))

	cells := buildRow(canonicalPath)["cells"].(map[string]any)
	for _, fieldKey := range []string{
		"timeline.date_entered_text",
		"timeline.activity_sort_ts",
		"timeline.date_entered_sort_day",
		"timeline.replacement_record_id",
	} {
		if cells[fieldKey].(map[string]any)["value"] != nil {
			t.Fatalf("%s must remain null, got %#v", fieldKey, cells[fieldKey])
		}
	}
	for _, fieldKey := range []string{
		"timeline.host_refs",
		"timeline.identity_refs",
		"timeline.attached_evidence_ids",
		"timeline.tags",
	} {
		value := cells[fieldKey].(map[string]any)["value"].(map[string]any)
		if items := value["items"].([]map[string]any); len(items) != 0 {
			t.Fatalf("%s must derive an empty collection, got %#v", fieldKey, items)
		}
	}
}

func TestProjectionMutationContractRejectsMalformed_Unit(t *testing.T) {
	recordID := uuid.MustParse("a0000000-0000-4000-8000-00000000000a")
	input := workbookprojection.Derive(sourcerepository.Snapshot{
		RecordID:              recordID,
		IncidentID:            uuid.MustParse("b0000000-0000-4000-8000-00000000000b"),
		ActivityTimePairState: "empty",
		CaptureState:          captureStateRough,
		RowVersion:            1,
		RecordedAt:            time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
		EditedAt:              time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
	}, nil).ProjectionInput()

	valid := workbookprojection.ProjectionMutation{
		Kind:     workbookprojection.ProjectionMutationUpsert,
		RecordID: recordID,
		Input:    input,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid projection mutation rejected: %v", err)
	}

	for name, mutation := range map[string]workbookprojection.ProjectionMutation{
		"unknown kind": {
			Kind:     "replace",
			RecordID: recordID,
		},
		"upsert identity mismatch": {
			Kind:     workbookprojection.ProjectionMutationUpsert,
			RecordID: uuid.New(),
			Input:    input,
		},
		"delete with input": {
			Kind:     workbookprojection.ProjectionMutationDelete,
			RecordID: recordID,
			Input:    input,
		},
		"missing record": {
			Kind: workbookprojection.ProjectionMutationDelete,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := mutation.Validate(); err == nil {
				t.Fatalf("malformed projection mutation accepted: %#v", mutation)
			}
		})
	}
}

func assertCanonicalJSONEqual(t *testing.T, left any, right any) {
	t.Helper()
	leftJSON, err := json.Marshal(left)
	if err != nil {
		t.Fatalf("marshal left canonical JSON: %v", err)
	}
	rightJSON, err := json.Marshal(right)
	if err != nil {
		t.Fatalf("marshal right canonical JSON: %v", err)
	}
	if string(leftJSON) != string(rightJSON) {
		t.Fatalf("canonical JSON mismatch:\nleft:  %s\nright: %s", leftJSON, rightJSON)
	}
}
