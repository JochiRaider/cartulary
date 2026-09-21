package httpapi

import (
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
	"reflect"
	"testing"
	"time"
)

func TestHistoryKeyset_Unit(t *testing.T) {
	binding := pagination.Binding{Route: "records.history", ActorUserID: "actor", Limit: 1, Scope: map[string]string{"record_id": "record"}}
	for _, position := range []revisions.HistoryPosition{
		{CommittedAt: time.Date(2026, 9, 21, 12, 0, 0, 123456000, time.UTC), ChangeSetID: uuid.New(), Kind: revisions.HistoryMutation, SequenceNo: 7},
		{CommittedAt: time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC), ChangeSetID: uuid.New(), Kind: revisions.HistoryRevision, RevisionNo: 2},
	} {
		cursor := historyCursor(binding, &position)
		got, err := historyPosition(cursor)
		if err != nil || !reflect.DeepEqual(got, &position) {
			t.Fatalf("roundtrip: %+v %v", got, err)
		}
		for key, invalid := range map[string]string{"version": "old", "kind": "unknown", "sequence_no": "-1", "revision_no": "-1", "committed_at": "invalid", "change_set_id": "invalid"} {
			bad := historyCursor(binding, &position)
			bad.Position[key] = invalid
			if _, err := historyPosition(bad); !errors.Is(err, pagination.ErrInvalidCursorToken) {
				t.Fatalf("accepted invalid %s: %v", key, err)
			}
		}
	}
	for _, invalid := range []*pagination.Cursor{
		pagination.NewOffsetCursor(binding, 1), {Mode: pagination.ModeKeyset},
		{Mode: pagination.ModeKeyset, Position: map[string]string{"after_history_item_ref": "legacy"}},
	} {
		if _, err := historyPosition(invalid); !errors.Is(err, pagination.ErrInvalidCursorToken) {
			t.Fatalf("invalid cursor accepted: %+v %v", invalid, err)
		}
	}
	if got, err := historyPosition(nil); got != nil || err != nil {
		t.Fatalf("initial: %+v %v", got, err)
	}
	if historyCursor(binding, nil) != nil {
		t.Fatal("terminal page has continuation")
	}
	items := historyResources([]revisions.RecordHistoryItem{{CommittedAt: time.Date(2026, 9, 21, 12, 0, 0, 0, time.FixedZone("offset", 3600))}})
	if !reflect.DeepEqual(items[0]["available_rollback_actions"], []string{}) || items[0]["committed_at"] != "2026-09-21T11:00:00Z" {
		t.Fatalf("wire normalization: %#v", items)
	}
	if len(historyResources(nil)) != 0 {
		t.Fatal("nonempty empty page")
	}
}
