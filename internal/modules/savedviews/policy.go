package savedviews

import (
	"bytes"
	"encoding/json"
	"reflect"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type createRequest struct {
	ViewSchemaID string
	DisplayName  string
	Scope        scope
	QueryJSON    []byte
	LayoutJSON   []byte
}

type optionalString struct {
	Present bool
	Value   string
}

type optionalScope struct {
	Present bool
	Value   scope
}

type optionalJSON struct {
	Present bool
	Value   []byte
}

type patchRequest struct {
	BaseSavedViewVersion int64
	DisplayName          optionalString
	Scope                optionalScope
	QueryJSON            optionalJSON
	LayoutJSON           optionalJSON
}

type mutationPolicyError struct {
	Field          string
	FieldKey       string
	FilterIndex    *int
	ReasonCode     string
	RequestedCount *int
	MaxCount       *int
}

func validateViewSchemaID(viewSchemaID string) *mutationPolicyError {
	if _, ok := viewschema.Lookup(viewSchemaID); !ok {
		return &mutationPolicyError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
	return nil
}

func normalizeOrdinaryScope(value string) (scope, *mutationPolicyError) {
	parsed, ok := parseScope(value)
	if !ok {
		return "", &mutationPolicyError{Field: "scope", ReasonCode: "invalid_scope"}
	}
	if !isOrdinaryCreateScope(parsed) {
		return "", &mutationPolicyError{Field: "scope", ReasonCode: "system_scope_forbidden"}
	}
	return parsed, nil
}

func normalizeMutationDisplayName(value string) (string, *mutationPolicyError) {
	normalized, ok := normalizeDisplayName(value)
	if !ok {
		return "", &mutationPolicyError{Field: "display_name", ReasonCode: "invalid_value"}
	}
	return normalized, nil
}

func normalizeMutationQuery(raw json.RawMessage, viewSchemaID string) ([]byte, *mutationPolicyError) {
	normalized, validationErr := viewquery.NormalizePersisted(raw, viewSchemaID)
	if validationErr == nil {
		return normalized, nil
	}
	return nil, &mutationPolicyError{
		Field:          validationErr.Field,
		FieldKey:       validationErr.FieldKey,
		FilterIndex:    validationErr.FilterIndex,
		ReasonCode:     validationErr.ReasonCode,
		RequestedCount: validationErr.RequestedCount,
		MaxCount:       validationErr.MaxCount,
	}
}

func normalizeMutationLayout(raw json.RawMessage, viewSchemaID string) ([]byte, *mutationPolicyError) {
	normalized, layoutErr := viewschema.NormalizeLayout(raw, viewSchemaID)
	if layoutErr == nil {
		return normalized, nil
	}
	return nil, &mutationPolicyError{Field: layoutErr.Field, ReasonCode: layoutErr.ReasonCode}
}

func applyPatch(current savedViewRecord, request patchRequest, updatedAt time.Time) (savedViewRecord, bool, error) {
	next := current
	if request.DisplayName.Present {
		next.DisplayName = request.DisplayName.Value
	}
	if request.Scope.Present {
		next.Scope = request.Scope.Value
	}
	if request.QueryJSON.Present {
		next.QueryJSON = append([]byte(nil), request.QueryJSON.Value...)
	}
	if request.LayoutJSON.Present {
		next.LayoutJSON = append([]byte(nil), request.LayoutJSON.Value...)
	}

	sameQuery, err := jsonStructurallyEqual(current.QueryJSON, next.QueryJSON)
	if err != nil {
		return savedViewRecord{}, false, err
	}
	sameLayout, err := jsonStructurallyEqual(current.LayoutJSON, next.LayoutJSON)
	if err != nil {
		return savedViewRecord{}, false, err
	}
	if current.DisplayName == next.DisplayName &&
		current.Scope == next.Scope &&
		sameQuery &&
		sameLayout {
		return current, false, nil
	}

	next.UpdatedAt = updatedAt.UTC()
	next.SavedViewVersion = current.SavedViewVersion + 1
	return next, true, nil
}

func jsonStructurallyEqual(left []byte, right []byte) (bool, error) {
	var leftValue any
	var rightValue any
	leftDecoder := json.NewDecoder(bytes.NewReader(left))
	leftDecoder.UseNumber()
	if err := leftDecoder.Decode(&leftValue); err != nil {
		return false, err
	}
	rightDecoder := json.NewDecoder(bytes.NewReader(right))
	rightDecoder.UseNumber()
	if err := rightDecoder.Decode(&rightValue); err != nil {
		return false, err
	}
	return reflect.DeepEqual(leftValue, rightValue), nil
}
