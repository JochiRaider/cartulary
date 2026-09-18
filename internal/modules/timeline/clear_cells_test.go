package timeline

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
)

func TestClearCellsSourcePlanAndDateIntent_Unit(t *testing.T) {
	fields := mutationpolicy.DirectWritableFieldKeys()
	rows, err := buildClearCellsOwnerRows(fields, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if len(row.Cells) != 10 {
			t.Fatalf("wrong field count: %#v", row)
		}
		for index, cell := range row.Cells {
			if cell.FieldKey != fields[index] || cell.Change.TextValue != nil || cell.Change.CanonicalValue() != nil {
				t.Fatalf("clear is not explicit null: %#v", cell)
			}
		}
	}
	for _, count := range []int{0, 501} {
		if _, err := buildClearCellsOwnerRows(fields, count); err == nil {
			t.Fatalf("admitted %d targets", count)
		}
	}
	if _, err := buildClearCellsOwnerRows([]string{"timeline.tags"}, 1); err == nil {
		t.Fatal("collection admitted")
	}
	utc, local, offset := "2026-09-18T12:00:00Z", "2026-09-18T07:00:00-05:00", -300
	profile := TimeConversionProfile{Enabled: true, LocalOffsetMinutes: &offset}
	before := sourcerepository.Snapshot{ActivityUTCText: &utc, ActivityLocalText: &local, ActivityLocalGenerated: true, ActivityTimePairState: "paired_generated"}
	next := before
	clear := PatchChange{FieldKey: "timeline.activity_utc_text"}
	applyPatchChangeToSource(&next, clear)
	applyTimelineDatePatch(&next, before, profile, []PatchChange{clear})
	if next.ActivityUTCText != nil || *next.ActivityLocalText != local || next.ActivityTimePairState != "conversion_unavailable" {
		t.Fatalf("clear regenerated or altered counterpart: %#v", next)
	}
	before = next
	other := "Unrelated"
	applyPatchChangeToSource(&next, PatchChange{FieldKey: "timeline.analyst_text", TextValue: &other})
	applyTimelineDatePatch(&next, before, profile, []PatchChange{{FieldKey: "timeline.analyst_text", TextValue: &other}})
	if next.ActivityUTCText != nil {
		t.Fatal("unrelated edit regenerated clear")
	}
	before = next
	newLocal := "2026-09-18T08:00:00-05:00"
	date := PatchChange{FieldKey: "timeline.activity_local_text", TextValue: &newLocal}
	applyPatchChangeToSource(&next, date)
	applyTimelineDatePatch(&next, before, profile, []PatchChange{date})
	if next.ActivityUTCText == nil || *next.ActivityUTCText != "2026-09-18T13:00:00Z" {
		t.Fatalf("date edit did not generate counterpart: %#v", next)
	}
	before = next
	empty := ""
	changes := []PatchChange{{FieldKey: "timeline.activity_utc_text", TextValue: &empty}, {FieldKey: "timeline.activity_local_text", TextValue: &newLocal}}
	for _, change := range changes {
		applyPatchChangeToSource(&next, change)
	}
	applyTimelineDatePatch(&next, before, profile, changes)
	if next.ActivityUTCText == nil || *next.ActivityUTCText != "" {
		t.Fatal("explicit empty text was overwritten or coerced")
	}
}
