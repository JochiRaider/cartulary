package versionid

import (
	"testing"

	"github.com/google/uuid"
)

func TestFormatUsesTimelineRecordIdentityAndExactRowVersion(t *testing.T) {
	recordID := uuid.MustParse("11111111-2222-4333-8444-555555555555")
	if got, want := Format(recordID, 42), "timeline_record:11111111-2222-4333-8444-555555555555:42"; got != want {
		t.Fatalf("version ID = %q, want %q", got, want)
	}
}
