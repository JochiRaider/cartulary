package savedviews

import (
	"net/http"
	"net/url"
	"strconv"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func savedViewReadError(code, reason string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusBadRequest, Code: code, Details: map[string]any{"reason_code": reason}}
}

func parseSavedViewReadQuery(raw string) (url.Values, *httpapi.APIError) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return nil, savedViewReadError("invalid_saved_view_read_request", "malformed_query")
	}
	for key, entries := range values {
		if !utf8.ValidString(key) {
			return nil, savedViewReadError("invalid_saved_view_read_request", "malformed_query")
		}
		for _, value := range entries {
			if !utf8.ValidString(value) {
				return nil, savedViewReadError("invalid_saved_view_read_request", "malformed_query")
			}
		}
	}
	return values, nil
}

func validateSavedViewDetailQuery(raw string) *httpapi.APIError {
	values, err := parseSavedViewReadQuery(raw)
	if err != nil {
		return err
	}
	if err := httpapi.ValidateSingletonReadQuery(values); err != nil {
		return err
	}
	if len(values) != 0 {
		return savedViewReadError("invalid_saved_view_read_request", "unknown_query_member")
	}
	return nil
}

func resolveSavedViewListRead(codec *pagination.Codec, raw, actorID, incidentID string) (pagination.Binding, *pagination.Cursor, *httpapi.APIError) {
	empty := pagination.Binding{}
	values, apiErr := parseSavedViewReadQuery(raw)
	if apiErr != nil {
		return empty, nil, apiErr
	}
	for _, entries := range values {
		if len(entries) != 1 {
			return empty, nil, savedViewReadError("invalid_list_query", "duplicate_query_member")
		}
	}
	for _, alias := range []string{"page", "offset", "page_size", "block_size"} {
		if values.Has(alias) {
			return empty, nil, invalidPaginationRequest(pagination.ReasonInvalidLimit)
		}
	}
	for key := range values {
		if key != "limit" && key != "cursor_token" && key != "view_schema_id" {
			return empty, nil, savedViewReadError("invalid_list_query", "unknown_query_member")
		}
	}
	schema := values.Get("view_schema_id")
	if values.Has("view_schema_id") {
		if _, ok := viewschema.Lookup(schema); !ok {
			return empty, nil, savedViewReadError("invalid_list_query", "invalid_filter_value")
		}
	}
	var query pagination.Query
	if values.Has("limit") {
		limit, err := strconv.Atoi(values.Get("limit"))
		if err != nil || limit < 1 || limit > pagination.MaxLimit {
			return empty, nil, invalidPaginationRequest(pagination.ReasonInvalidLimit)
		}
		query.Limit = &limit
	}
	var cursor *pagination.Cursor
	if values.Has("cursor_token") {
		decoded, err := codec.Decode(values.Get("cursor_token"))
		if err != nil {
			return empty, nil, invalidPaginationRequest(pagination.ReasonInvalidCursorToken)
		}
		cursor = &decoded
	}
	binding := pagination.Binding{Route: "incident.saved-views.list", ActorUserID: actorID, Limit: query.EffectiveLimit(cursor), Scope: map[string]string{"incident_id": incidentID, "view_schema_id": schema}}
	if cursor != nil {
		// Validate all non-filter bindings first. The legacy missing schema binding
		// denotes unfiltered discovery only; the encrypted token is decoded first.
		priorSchema := cursor.Scope["view_schema_id"]
		normalized := *cursor
		normalized.Scope = make(map[string]string, len(cursor.Scope)+1)
		for key, value := range cursor.Scope {
			normalized.Scope[key] = value
		}
		normalized.Scope["view_schema_id"] = schema
		if err := normalized.Validate(binding); err != nil {
			return empty, nil, invalidPaginationRequest(pagination.ReasonInvalidCursorToken)
		}
		if priorSchema != schema {
			return empty, nil, invalidPaginationRequest(pagination.ReasonCursorQueryMismatch)
		}
	}
	return binding, cursor, nil
}
