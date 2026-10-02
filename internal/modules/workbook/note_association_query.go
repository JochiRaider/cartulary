package workbook

import (
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

const noteAssociationRoute = "workbook.note-associations"
const noteAssociationOrder = "created_at_desc_link_id_desc"

type noteAssociationListRequest struct {
	query   NoteAssociationQuery
	binding pagination.Binding
}

// decodeNoteAssociationListRequest admits only the public read contract. The
// caller must establish current visibility first; source eligibility and link
// semantics remain behind NoteAssociationProvider.
func decodeNoteAssociationListRequest(raw string, actorID uuid.UUID, target RecordTarget, codec *pagination.Codec) (noteAssociationListRequest, *httpapi.APIError) {
	listError := func(reason string) (noteAssociationListRequest, *httpapi.APIError) {
		return noteAssociationListRequest{}, &httpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_list_query", Message: "invalid list query", Details: map[string]any{"reason_code": reason}}
	}
	pagingError := func(reason string) (noteAssociationListRequest, *httpapi.APIError) {
		return noteAssociationListRequest{}, &httpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_pagination_request", Message: "invalid pagination request", Details: map[string]any{"reason_code": reason}}
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return listError("malformed_query")
	}
	for key, entries := range values {
		if !utf8.ValidString(key) {
			return listError("malformed_query")
		}
		for _, value := range entries {
			if !utf8.ValidString(value) {
				return listError("malformed_query")
			}
		}
	}
	for _, entries := range values {
		if len(entries) != 1 {
			return listError("duplicate_query_member")
		}
	}
	for _, alias := range []string{"page", "offset", "page_size", "block_size"} {
		if values.Has(alias) {
			return pagingError(pagination.ReasonInvalidLimit)
		}
	}
	for key := range values {
		if key != "kind" && key != "limit" && key != "cursor_token" {
			return listError("unknown_query_member")
		}
	}
	kind := values.Get("kind")
	switch kind {
	case "source", "evidence", "related_note": // Exact public wire tokens, not relationship policy.
	default:
		return listError("invalid_filter_value")
	}
	limit := pagination.DefaultLimit
	if values.Has("limit") {
		limit, err = strconv.Atoi(values.Get("limit"))
		if err != nil || limit < 1 || limit > pagination.MaxLimit {
			return pagingError(pagination.ReasonInvalidLimit)
		}
	}
	var cursor *pagination.Cursor
	if values.Has("cursor_token") {
		decoded, err := codec.Decode(values.Get("cursor_token"))
		if err != nil || decoded.Route != noteAssociationRoute || decoded.ActorUserID != actorID.String() {
			return pagingError(pagination.ReasonInvalidCursorToken)
		}
		cursor = &decoded
		if !values.Has("limit") {
			limit = cursor.Limit
		}
	}
	request := noteAssociationListRequest{
		query: NoteAssociationQuery{ActorID: actorID, Target: target, Kind: kind, Limit: limit},
		binding: pagination.Binding{Route: noteAssociationRoute, ActorUserID: actorID.String(), Limit: limit, Scope: map[string]string{
			"incident": target.IncidentID.String(), "note": target.RecordID.String(), "kind": kind, "order": noteAssociationOrder,
		}},
	}
	if cursor == nil {
		return request, nil
	}
	if err := cursor.Validate(request.binding); err != nil {
		return pagingError(pagination.ReasonCursorQueryMismatch)
	}
	stamp, timeErr := time.Parse(time.RFC3339Nano, cursor.Position["created_at"])
	id, idErr := uuid.Parse(cursor.Position["link_id"])
	if cursor.Mode != pagination.ModeKeyset || len(cursor.Position) != 2 || timeErr != nil || idErr != nil || id == uuid.Nil {
		return pagingError(pagination.ReasonInvalidCursorToken)
	}
	request.query.BeforeTime = &stamp
	request.query.BeforeID = id
	return request, nil
}
