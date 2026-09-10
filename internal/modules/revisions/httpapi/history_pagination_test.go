package httpapi

import (
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

func TestHistoryKeyset_Unit(t *testing.T) {
	binding := pagination.Binding{Route: "records.history", ActorUserID: "actor", Limit: 1, Scope: map[string]string{"record_id": "record"}}
	resources := []map[string]any{{"history_item_ref": "a"}, {"history_item_ref": "b"}}
	page, cursor, err := pageRecordHistory(binding, nil, resources)
	if err != nil || len(page) != 1 || cursor == nil || cursor.Mode != pagination.ModeKeyset {
		t.Fatalf("initial page: %s %+v %v", page, cursor, err)
	}
	live := append([]map[string]any{{"history_item_ref": "new"}}, resources...)
	page, terminal, err := pageRecordHistory(binding, cursor, live)
	if err != nil || terminal != nil || len(page) != 1 || string(page[0]) != `{"history_item_ref":"b"}` {
		t.Fatalf("live continuation: %s %+v %v", page, terminal, err)
	}
	for _, invalid := range []*pagination.Cursor{
		pagination.NewOffsetCursor(binding, 1),
		{Mode: pagination.ModeKeyset},
		{Mode: pagination.ModeKeyset, Position: map[string]string{"after_history_item_ref": "missing"}},
		{Mode: pagination.ModeKeyset, Position: map[string]string{"after_history_item_ref": "a", "offset": "1"}},
	} {
		if _, _, err := pageRecordHistory(binding, invalid, resources); !errors.Is(err, pagination.ErrInvalidCursorToken) {
			t.Fatalf("invalid cursor accepted: %+v %v", invalid, err)
		}
	}
	page, terminal, err = pageRecordHistory(binding, nil, nil)
	if err != nil || terminal != nil || len(page) != 0 {
		t.Fatalf("empty page: %s %+v %v", page, terminal, err)
	}
}
