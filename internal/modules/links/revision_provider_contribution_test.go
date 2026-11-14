package links_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func TestRevisionProviderContributionOwnsLinkAndTagHistoryFacet_Unit(t *testing.T) {
	exerciseLinksHistoryReferenceProjection(t)
	contribution := links.RevisionProviderContribution()
	if contribution.SourceOwnerModule != revisions.SourceOwnerLinks || len(contribution.NonRowTargets) != 2 {
		t.Fatalf("unexpected Links revision contribution: %#v", contribution)
	}

	sourceID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	destinationID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	recordID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	for _, target := range contribution.NonRowTargets {
		if target.RollbackProvider == nil || target.HistoryValidator == nil || target.HistoryReferences == nil {
			t.Fatalf("target %q has incomplete semantics", target.TargetKind)
		}
		var mutation revisions.StoredMutation
		var want []uuid.UUID
		switch target.TargetKind {
		case "record_link":
			mutation = revisions.StoredMutation{
				TargetKind: "record_link",
				TargetID:   uuid.NewString(),
				BeforeValue: map[string]any{
					"src_record_id": sourceID.String(),
					"dst_record_id": destinationID.String(),
				},
				AfterValue: map[string]any{
					"src_record_id": destinationID.String(),
					"dst_record_id": sourceID.String(),
				},
			}
			want = []uuid.UUID{sourceID, destinationID}
		case "record_tag":
			mutation = revisions.StoredMutation{
				TargetKind: "record_tag",
				TargetID:   "record_tag:" + recordID.String() + ":" + uuid.NewString(),
				AfterValue: map[string]any{"record_id": recordID.String()},
			}
			want = []uuid.UUID{recordID}
		default:
			t.Fatalf("unexpected Links target %q", target.TargetKind)
		}
		description, err := target.HistoryFacet.DescribeMutation(mutation)
		if err != nil {
			t.Fatalf("describe %s: %v", target.TargetKind, err)
		}
		if !slices.Equal(description.HistoryRecordIDs, want) || !slices.Equal(description.HistoryEntryRecordIDs, want) {
			t.Fatalf("unexpected %s associations: records=%v entries=%v want=%v", target.TargetKind, description.HistoryRecordIDs, description.HistoryEntryRecordIDs, want)
		}
	}
}

func exerciseLinksHistoryReferenceProjection(t *testing.T) {
	t.Helper()
	firstIncident := uuid.MustParse("aaaaaaaa-1111-4111-8111-111111111111")
	secondIncident := uuid.MustParse("bbbbbbbb-2222-4222-8222-222222222222")
	firstRecord := uuid.MustParse("cccccccc-3333-4333-8333-333333333333")
	secondRecord := uuid.MustParse("dddddddd-4444-4444-8444-444444444444")
	for _, target := range links.RevisionProviderContribution().NonRowTargets {
		t.Run("reference_projection/"+target.TargetKind, func(t *testing.T) {
			before := map[string]any{"incident_id": firstIncident.String()}
			after := map[string]any{"incident_id": secondIncident.String()}
			fields := []string{"incident_id", "record_id"}
			want := []revisions.HistoryRecordReference{{IncidentID: firstIncident, RecordID: firstRecord}, {IncidentID: secondIncident, RecordID: secondRecord}}
			if target.TargetKind == "record_link" {
				before["src_record_id"], before["dst_record_id"] = firstRecord.String(), secondRecord.String()
				after["src_record_id"], after["dst_record_id"] = secondRecord.String(), firstRecord.String()
				fields = []string{"incident_id", "src_record_id", "dst_record_id"}
				want = []revisions.HistoryRecordReference{{IncidentID: firstIncident, RecordID: firstRecord}, {IncidentID: firstIncident, RecordID: secondRecord}, {IncidentID: secondIncident, RecordID: secondRecord}, {IncidentID: secondIncident, RecordID: firstRecord}}
			} else {
				before["record_id"], after["record_id"] = firstRecord.String(), secondRecord.String()
			}
			mutation := revisions.StoredMutation{TargetKind: target.TargetKind, BeforeValue: before, AfterValue: after}
			got, err := target.HistoryReferences(mutation, uuid.Nil)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("reference sides=%v err=%v want=%v", got, err, want)
			}
			if err := target.HistoryValidator.ValidateHistoryMutation(mutation); err == nil {
				t.Fatal("reference-only fixture must not pass complete history admission")
			}
			for _, side := range []string{"before", "after"} {
				for _, field := range fields {
					for _, defect := range []string{"missing", "null", "zero", "noncanonical", "type"} {
						t.Run(side+"/"+field+"/"+defect, func(t *testing.T) {
							copyMap := func(value map[string]any) map[string]any {
								copy := make(map[string]any, len(value))
								for k, v := range value {
									copy[k] = v
								}
								return copy
							}
							candidate := mutation
							candidate.BeforeValue, candidate.AfterValue = copyMap(before), copyMap(after)
							value := candidate.BeforeValue
							if side == "after" {
								value = candidate.AfterValue
							}
							switch defect {
							case "missing":
								delete(value, field)
							case "null":
								value[field] = nil
							case "zero":
								value[field] = uuid.Nil.String()
							case "noncanonical":
								value[field] = strings.ToUpper(value[field].(string))
							case "type":
								value[field] = 1
							}
							original := copyMap(value)
							if _, err := target.HistoryReferences(candidate, uuid.Nil); err == nil {
								t.Fatal("invalid reference scalar admitted")
							}
							if !reflect.DeepEqual(value, original) {
								t.Fatal("reference projection changed retained input")
							}
						})
					}
				}
			}
			if got, err := target.HistoryReferences(revisions.StoredMutation{TargetKind: target.TargetKind}, uuid.Nil); err != nil || len(got) != 0 {
				t.Fatalf("absent sides produced references=%v err=%v", got, err)
			}
		})
	}
}
