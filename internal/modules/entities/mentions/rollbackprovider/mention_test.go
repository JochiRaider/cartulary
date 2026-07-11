package rollbackprovider

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestParseMentionTargetRejectsMalformedIdentity(t *testing.T) {
	t.Parallel()
	mentionID := uuid.New()
	target := rollbackcontract.NonRowTarget{
		TargetKind: "entity_mention", TargetID: mentionID.String(), OperationKind: "patch",
		BeforeValue: map[string]any{"entity_mention_id": uuid.New().String()},
	}
	if _, err := parseMentionTarget(target); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("parseMentionTarget error = %v", err)
	}
}

func TestParseMentionTargetAcceptsRetainedMention(t *testing.T) {
	t.Parallel()
	mentionID, sourceID := uuid.New(), uuid.New()
	target := rollbackcontract.NonRowTarget{
		TargetKind: "entity_mention", TargetID: mentionID.String(), OperationKind: "patch",
		BeforeValue: map[string]any{
			"entity_mention_id": mentionID.String(), "source_record_id": sourceID.String(),
			"source_field_key": "timeline.host_refs", "entity_type": "host", "origin_kind": "workbook",
			"origin_locator": "cell", "raw_text": "HOST-1", "normalized_text": "host-1",
			"resolution_status": "unresolved", "row_version": float64(1),
		},
	}
	identity, err := parseMentionTarget(target)
	if err != nil || identity.sourceRecordID != sourceID {
		t.Fatalf("parseMentionTarget = %#v, %v", identity, err)
	}
}

func TestMentionChangedKeysAreOwnerDefined(t *testing.T) {
	t.Parallel()

	keys := mentionChangedFieldKeys("timeline_event", "timeline.host_refs")
	want := []string{"timeline.has_unresolved_mentions", "timeline.host_refs"}
	if len(keys) != len(want) || keys[0] != want[0] || keys[1] != want[1] {
		t.Fatalf("mention changed keys = %v, want %v", keys, want)
	}

	keys = mentionChangedFieldKeys("host", "host.aliases")
	if len(keys) != 1 || keys[0] != "host.aliases" {
		t.Fatalf("host mention changed keys = %v", keys)
	}
}
