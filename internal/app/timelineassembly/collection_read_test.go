package timelineassembly

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func TestEvidenceTimelineFactMappingIsLosslessAndOrdered(t *testing.T) {
	firstID := uuid.MustParse("3670034a-9fa7-45dd-b6b2-b11b904944c4")
	secondID := uuid.MustParse("974a02a1-52d1-4dd1-a896-cf551e15b131")
	input := []evidence.TimelineFact{
		{
			RecordID:       firstID,
			Title:          "Disk image",
			LifecycleState: "available",
			UploadState:    "available",
		},
		{
			RecordID:       secondID,
			Title:          "Memory capture",
			LifecycleState: "quarantined",
			UploadState:    "quarantined",
		},
	}
	want := []workbookprojection.EvidenceFact{
		{
			RecordID:       firstID,
			Title:          "Disk image",
			LifecycleState: "available",
			UploadState:    "available",
		},
		{
			RecordID:       secondID,
			Title:          "Memory capture",
			LifecycleState: "quarantined",
			UploadState:    "quarantined",
		},
	}

	if got := mapEvidenceTimelineFacts(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("mapped facts = %#v, want %#v", got, want)
	}
	if got := mapEvidenceTimelineFacts(nil); got == nil || len(got) != 0 {
		t.Fatalf("nil input mapped to %#v, want non-nil empty slice", got)
	}
}
