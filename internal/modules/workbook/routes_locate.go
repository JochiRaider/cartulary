package workbook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/google/uuid"
)

func decodeLocatorRequest(r *http.Request, schema string) (uuid.UUID, viewquery.Query, *httpapi.APIError) {
	raw, err := strictjson.DecodeObject(r.Body)
	if err != nil {
		return uuid.Nil, viewquery.Query{}, invalidViewQuery("", "malformed_locator_request")
	}
	if r.URL.RawQuery != "" || r.URL.ForceQuery {
		return uuid.Nil, viewquery.Query{}, invalidViewQuery("", "unknown_locator_member")
	}
	for key := range raw {
		switch key {
		case "record_id", "sort", "filters", "group_by":
		default:
			return uuid.Nil, viewquery.Query{}, invalidViewQuery("", "unknown_locator_member")
		}
	}
	var value string
	if err := json.Unmarshal(raw["record_id"], &value); err != nil {
		return uuid.Nil, viewquery.Query{}, invalidViewQuery("record_id", "invalid_record_id")
	}
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, viewquery.Query{}, invalidViewQuery("record_id", "invalid_record_id")
	}
	delete(raw, "record_id")
	// Ordinary query normalization admits omitted arrays. Locator explicitly
	// distinguishes omission from null, in the ordinary sort/filter order.
	for _, field := range []string{"sort", "filters"} {
		if bytes.Equal(bytes.TrimSpace(raw[field]), []byte("null")) {
			raw[field] = json.RawMessage(`{}`)
		}
	}

	payload, _ := json.Marshal(raw)
	query, validation := viewquery.Decode(bytes.NewReader(payload), schema)
	if validation != nil {
		return uuid.Nil, viewquery.Query{}, invalidViewQueryValidation(validation)
	}
	return id, query, nil
}

func (s *service) handleLocate(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	schema := r.PathValue("view_schema_id")
	provider, ok := s.contributions.QueryFor(schema)
	if !ok {
		writeAPIError(w, r, invalidViewQuery("view_schema_id", "unknown_view_schema"))
		return
	}
	id, query, apiErr := decodeLocatorRequest(r, schema)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	ctx, finish := s.startWorkbookQuery(r.Context(), schema)
	r = r.WithContext(ctx)
	result, code, count := "failed", "", -1
	defer func() { finish(result, code, count) }()
	fail := func(err error) {
		apiErr := internalAPIError(err)
		result, code = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
	}
	location, err := provider.LocateRows(ctx, LocateCommand{IncidentID: incidentID, ViewSchemaID: schema, RecordID: id, Query: query.Meta})
	if err != nil {
		fail(err)
		return
	}
	meta := httpapi.EnvelopeMeta{RequestID: httpapi.RequestIDFromContext(ctx)}
	data := map[string]any{"outcome": location.Outcome}
	switch location.Outcome {
	case "unavailable":
	case "outside_query":
		data["target_record_id"] = id.String()
	case "located":
		if len(location.Page.Rows) < 1 || len(location.Page.Rows) > pagination.DefaultLimit || fmt.Sprint(location.Page.Rows[0]["record_id"]) != id.String() {
			fail(fmt.Errorf("invalid locator provider result"))
			return
		}
		scope, err := workbookQueryScope(incidentID, schema, query.Meta)
		if err != nil {
			fail(err)
			return
		}
		binding, _, reason := s.cursorCodec.ResolveViewQuery(pagination.Query{}, "workbook.view-query", principal.User.ID.String(), scope)
		if reason != "" {
			fail(fmt.Errorf("locator cursor binding failed"))
			return
		}
		rows, next, err := pageBoundedWorkbookResources(binding, query.Meta, location.Page)
		if err != nil {
			fail(err)
			return
		}
		encode := func(cursor *pagination.Cursor) (*string, error) {
			if cursor == nil {
				return nil, nil
			}
			value, err := s.cursorCodec.Encode(*cursor)
			return &value, err
		}
		nextToken, err := encode(next)
		if err != nil {
			fail(err)
			return
		}
		var startToken *string
		if len(location.Predecessor) > 0 {
			startToken, err = encode(&pagination.Cursor{Mode: pagination.ModeKeyset, Route: binding.Route, ActorUserID: binding.ActorUserID, Limit: binding.Limit, Scope: binding.Scope, Position: location.Predecessor})
			if err != nil {
				fail(err)
				return
			}
		}
		data["target_record_id"], data["incident_id"], data["view_schema_id"] = id.String(), incidentID.String(), schema
		data["rows"], data["window_start_cursor"] = rows, startToken
		meta.Query = query.Meta
		meta.Paging = &httpapi.PagingMeta{Limit: pagination.DefaultLimit, HasMore: nextToken != nil, NextCursor: nextToken}
		count = len(rows)
	default:
		fail(fmt.Errorf("unknown locator provider outcome"))
		return
	}
	if err := s.slideSessionIfNeeded(ctx, &principal, r.Method, r.URL.Path); err != nil {
		fail(err)
		return
	}
	result = "success"
	_ = httpapi.WriteSuccessWithMeta(w, r, http.StatusOK, data, meta)
}
