package httpapi

import (
	"encoding/json"
	"net/url"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

func parseMembershipListQuery(raw string) (url.Values, *platformhttpapi.APIError) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return nil, invalidListQuery("malformed_query")
	}
	keys := make([]string, 0, len(values))
	for key, entries := range values {
		if !utf8.ValidString(key) {
			return nil, invalidListQuery("malformed_query")
		}
		for _, value := range entries {
			if !utf8.ValidString(value) {
				return nil, invalidListQuery("malformed_query")
			}
		}
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if len(values[key]) != 1 {
			return nil, invalidListQuery("duplicate_query_member")
		}
	}
	for _, key := range keys {
		switch key {
		case "limit":
			if values.Get(key) == "" {
				return nil, invalidPaginationRequest(pagination.ReasonInvalidLimit)
			}
		case "cursor_token":
			if values.Get(key) == "" {
				return nil, invalidPaginationRequest(pagination.ReasonInvalidCursorToken)
			}
		case "offset", "page", "page_size", "block_size":
			return nil, invalidPaginationRequest(pagination.ReasonInvalidLimit)
		default:
			return nil, invalidListQuery("unknown_query_member")
		}
	}
	return values, nil
}

func membershipListPageRequest(binding pagination.Binding, cursor *pagination.Cursor) (incidents.MembershipListPageRequest, string) {
	page := incidents.MembershipListPageRequest{Limit: binding.Limit + 1}
	if cursor == nil {
		return page, ""
	}
	if cursor.Mode != pagination.ModeKeyset || len(cursor.Position) != 2 {
		return incidents.MembershipListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	joinedAt, err := time.Parse(time.RFC3339Nano, cursor.Position["last_joined_at"])
	if err != nil || joinedAt.IsZero() || !joinedAt.Equal(joinedAt.Truncate(time.Microsecond)) {
		return incidents.MembershipListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	userID, err := uuid.Parse(cursor.Position["last_user_id"])
	if err != nil || userID == uuid.Nil {
		return incidents.MembershipListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	page.After = &incidents.MembershipListPosition{JoinedAt: joinedAt.UTC(), UserID: userID}
	return page, ""
}

func buildMembershipListPage(binding pagination.Binding, records []incidents.MembershipRecord) ([]json.RawMessage, *pagination.Cursor, error) {
	hasMore := len(records) > binding.Limit
	if hasMore {
		records = records[:binding.Limit]
	}
	resources := make([]map[string]any, 0, len(records))
	for _, record := range records {
		resources = append(resources, incidents.BuildMembershipResource(record))
	}
	rows, err := pagination.MarshalResources(resources)
	if err != nil || !hasMore {
		return rows, nil, err
	}
	last := records[len(records)-1]
	return rows, &pagination.Cursor{Version: pagination.CursorVersion, Mode: pagination.ModeKeyset, Route: binding.Route, ActorUserID: binding.ActorUserID, Limit: binding.Limit, Scope: binding.Scope, Position: map[string]string{"last_joined_at": last.JoinedAt.UTC().Format(time.RFC3339Nano), "last_user_id": last.UserID.String()}}, nil
}
