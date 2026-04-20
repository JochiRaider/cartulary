package pagination

import (
	"encoding/json"
	"errors"
	"net/url"
	"testing"
	"time"
)

func TestResolveRequestReusesCursorBoundLimitAndValidatesBindings(t *testing.T) {
	cursorToken, err := EncodeCursor(Cursor{
		Route:       "incidents.list",
		ActorUserID: "user-1",
		Limit:       25,
		Scope:       map[string]string{"incident_id": "incident-1"},
		SnapshotID:  "snapshot-1",
		Offset:      25,
	})
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}

	binding, cursor, reason := ResolveRequest(url.Values{
		"cursor_token": []string{cursorToken},
	}, "incidents.list", "user-1", map[string]string{"incident_id": "incident-1"})
	if reason != "" {
		t.Fatalf("unexpected resolve reason: %s", reason)
	}
	if cursor == nil {
		t.Fatal("expected decoded cursor")
	}
	if binding.Limit != 25 || binding.Route != "incidents.list" || binding.ActorUserID != "user-1" {
		t.Fatalf("unexpected resolved binding: %#v", binding)
	}

	_, _, reason = ResolveRequest(url.Values{
		"cursor_token": []string{cursorToken},
		"limit":        []string{"50"},
	}, "incidents.list", "user-1", map[string]string{"incident_id": "incident-1"})
	if reason != ReasonInvalidCursorToken {
		t.Fatalf("expected invalid cursor for mismatched explicit limit, got %q", reason)
	}

	_, _, reason = ResolveRequest(url.Values{
		"cursor_token": []string{cursorToken},
	}, "incidents.list", "user-2", map[string]string{"incident_id": "incident-1"})
	if reason != ReasonInvalidCursorToken {
		t.Fatalf("expected invalid cursor for mismatched actor, got %q", reason)
	}

	_, _, reason = ResolveRequest(url.Values{
		"page": []string{"2"},
	}, "incidents.list", "user-1", nil)
	if reason != ReasonPaginationNotSupported {
		t.Fatalf("expected pagination_not_supported, got %q", reason)
	}

	_, _, reason = ResolveRequest(url.Values{
		"limit": []string{"0"},
	}, "incidents.list", "user-1", nil)
	if reason != ReasonInvalidLimit {
		t.Fatalf("expected invalid_limit, got %q", reason)
	}
}

func TestRegistryProvidesSnapshotStablePagesAndExpiresInactiveChains(t *testing.T) {
	now := time.Date(2026, time.April, 20, 20, 0, 0, 0, time.UTC)
	registry := NewRegistry(WithNow(func() time.Time { return now }))
	binding := Binding{
		Route:       "users.list",
		ActorUserID: "admin-1",
		Limit:       2,
	}
	rows := []json.RawMessage{
		json.RawMessage(`{"user_id":"1","display_name":"one"}`),
		json.RawMessage(`{"user_id":"2","display_name":"two"}`),
		json.RawMessage(`{"user_id":"3","display_name":"three"}`),
		json.RawMessage(`{"user_id":"4","display_name":"four"}`),
		json.RawMessage(`{"user_id":"5","display_name":"five"}`),
	}

	pageOne, cursor := registry.Start(binding, rows)
	if len(pageOne) != 2 || cursor == nil || cursor.Offset != 2 {
		t.Fatalf("unexpected first page: rows=%d cursor=%#v", len(pageOne), cursor)
	}

	now = now.Add(9 * time.Minute)
	pageTwo, cursorTwo, err := registry.Continue(binding, *cursor)
	if err != nil {
		t.Fatalf("continue page two: %v", err)
	}
	if len(pageTwo) != 2 || cursorTwo == nil || cursorTwo.Offset != 4 {
		t.Fatalf("unexpected second page: rows=%d cursor=%#v", len(pageTwo), cursorTwo)
	}

	now = now.Add(9 * time.Minute)
	pageThree, cursorThree, err := registry.Continue(binding, *cursorTwo)
	if err != nil {
		t.Fatalf("continue page three: %v", err)
	}
	if len(pageThree) != 1 || cursorThree != nil {
		t.Fatalf("unexpected third page: rows=%d cursor=%#v", len(pageThree), cursorThree)
	}

	expiringRows := []json.RawMessage{
		json.RawMessage(`{"incident_id":"1"}`),
		json.RawMessage(`{"incident_id":"2"}`),
		json.RawMessage(`{"incident_id":"3"}`),
	}
	now = time.Date(2026, time.April, 20, 21, 0, 0, 0, time.UTC)
	_, expiringCursor := registry.Start(Binding{
		Route:       "incidents.list",
		ActorUserID: "analyst-1",
		Limit:       1,
	}, expiringRows)
	if expiringCursor == nil {
		t.Fatal("expected expiring cursor")
	}

	now = now.Add(MinimumRetention + time.Second)
	_, _, err = registry.Continue(Binding{
		Route:       "incidents.list",
		ActorUserID: "analyst-1",
		Limit:       1,
	}, *expiringCursor)
	if !errors.Is(err, ErrCursorSnapshotExpired) {
		t.Fatalf("expected expired snapshot error, got %v", err)
	}
}
