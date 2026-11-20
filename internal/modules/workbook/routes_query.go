package workbook

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/google/uuid"
)

func (s *service) handleQuery(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, apiErr := admittedPathUUID(r, "incident_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	viewSchemaID, apiErr := admittedPathString(r, "view_schema_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	query, apiErr := decodeViewQueryRequest(r, viewSchemaID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	ctx, finishTelemetry := s.startWorkbookQuery(r.Context(), viewSchemaID)
	r = r.WithContext(ctx)
	telemetryResult := "failed"
	telemetryErrorCode := ""
	telemetryRowCount := -1
	defer func() {
		finishTelemetry(telemetryResult, telemetryErrorCode, telemetryRowCount)
	}()
	scope, scopeErr := workbookQueryScope(incidentID, viewSchemaID, query.Meta)
	if scopeErr != nil {
		apiErr := internalAPIError(scopeErr)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	binding, cursor, reasonCode := s.cursorCodec.ResolveViewQuery(
		query.Pagination,
		"workbook.view-query",
		principal.User.ID.String(),
		scope,
	)
	if reasonCode != "" {
		apiErr := invalidViewQuery("", reasonCode)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}

	window := querypage.Window{Limit: binding.Limit}
	if cursor != nil {
		window.Position = cursor.Position
	}
	queryProvider, ok := s.contributions.QueryFor(viewSchemaID)
	if !ok {
		apiErr := internalAPIError(fmt.Errorf("workbook query surface %q is not registered", viewSchemaID))
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	page, err := queryProvider.QueryRowsPage(r.Context(), QueryCommand{
		IncidentID:   incidentID,
		ViewSchemaID: viewSchemaID,
		Query:        query.Meta,
		Window:       window,
	})
	if err != nil {
		if errors.Is(err, querypage.ErrInvalidPosition) {
			apiErr := invalidViewQuery("", pagination.ReasonInvalidCursorToken)
			telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
			writeAPIError(w, r, apiErr)
			return
		}
		apiErr := internalAPIError(err)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	rows, nextCursor, err := pageBoundedWorkbookResources(binding, query.Meta, page)
	switch {
	case errors.Is(err, pagination.ErrInvalidCursorToken), errors.Is(err, querypage.ErrInvalidPosition):
		apiErr := invalidViewQuery("", pagination.ReasonInvalidCursorToken)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	case err != nil:
		apiErr := internalAPIError(err)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		apiErr := internalAPIError(err)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	var nextToken *string
	if nextCursor != nil {
		token, err := s.cursorCodec.Encode(*nextCursor)
		if err != nil {
			apiErr := internalAPIError(err)
			telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
			writeAPIError(w, r, apiErr)
			return
		}
		nextToken = &token
	}
	telemetryResult = "success"
	telemetryRowCount = len(rows)
	_ = httpapi.WriteSuccessWithMeta(w, r, http.StatusOK, map[string]any{
		"incident_id":    incidentID.String(),
		"view_schema_id": viewSchemaID,
		"rows":           rows,
	}, httpapi.EnvelopeMeta{
		RequestID: httpapi.RequestIDFromContext(r.Context()),
		Paging: &httpapi.PagingMeta{
			Limit:      binding.Limit,
			HasMore:    nextToken != nil,
			NextCursor: nextToken,
		},
		Query: query.Meta,
	})
}

func decodeViewQueryRequest(r *http.Request, viewSchemaID string) (viewquery.Query, *httpapi.APIError) {
	query, err := viewquery.Decode(r.Body, viewSchemaID)
	if err != nil {
		return viewquery.Query{}, invalidViewQueryValidation(err)
	}
	return query, nil
}

func workbookQueryScope(incidentID uuid.UUID, viewSchemaID string, queryMeta viewschema.QueryMeta) (map[string]string, error) {
	payload, err := json.Marshal(queryMeta)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"incident_id":    incidentID.String(),
		"view_schema_id": viewSchemaID,
		"query_contract": string(payload),
	}, nil
}

func invalidViewQuery(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_view_query",
		Message: "invalid view query",
		Details: details,
	}
}

func invalidViewQueryValidation(err *viewquery.ValidationError) *httpapi.APIError {
	if err == nil {
		return invalidViewQuery("", "")
	}
	details := map[string]any{}
	if err.Field != "" {
		details["field"] = err.Field
	}
	if err.FieldKey != "" {
		details["field_key"] = err.FieldKey
	}
	if err.FilterIndex != nil {
		details["filter_index"] = *err.FilterIndex
	}
	if err.ReasonCode != "" {
		details["reason_code"] = err.ReasonCode
	}
	if err.RequestedCount != nil {
		details["requested_count"] = *err.RequestedCount
	}
	if err.MaxCount != nil {
		details["max_count"] = *err.MaxCount
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_view_query",
		Message: "invalid view query",
		Details: details,
	}
}
