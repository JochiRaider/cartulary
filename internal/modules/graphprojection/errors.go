package graphprojection

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

type LifecycleError struct {
	Code       string
	ReasonCode string
	Field      string
	Details    map[string]any
	cause      error
}

func NewLifecycleError(code, reasonCode string, details map[string]any, cause error) *LifecycleError {
	return &LifecycleError{Code: code, ReasonCode: reasonCode, Details: cloneErrorDetails(details), cause: cause}
}

func (err *LifecycleError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.cause
}

func (err *LifecycleError) Error() string { return contractErrorText(err.Code, err.ReasonCode) }

func (err *LifecycleError) Retryable() bool {
	return err != nil && ((err.Code == "operation_conflict" && err.ReasonCode == "run_already_active") || (err.Code == "ephemeral_projection_failed" && err.ReasonCode == "projection_computation_failed"))
}

func (err *LifecycleError) EnvelopeJSON() ([]byte, error) {
	orders := map[string][]string{
		"invalid_operation_request":   {"operation", "field", "reason_code"},
		"invalid_projection_request":  {"operation", "reason_code", "field", "validation_code"},
		"invalid_operation":           {"operation", "reason_code"},
		"operation_conflict":          {"operation", "reason_code", "active_projection_run_id"},
		"graph_view_not_found":        {"operation", "graph_view_id"},
		"projection_run_not_found":    {"operation", "graph_view_id", "projection_run_id"},
		"ephemeral_projection_failed": {"operation", "reason_code", "graph_view_id", "ephemeral_projection_id", "validation_summary"},
	}
	order, ok := orders[err.Code]
	if !ok {
		return nil, fmt.Errorf("unregistered lifecycle error code %q", err.Code)
	}
	return marshalErrorEnvelope(err.Code, err.Retryable(), err.Details, order)
}

type QueryError struct {
	Code       string
	ReasonCode string
	Field      string
	Details    map[string]any
	cause      error
}

func NewQueryError(code, reasonCode string, details map[string]any, cause error) *QueryError {
	return &QueryError{Code: code, ReasonCode: reasonCode, Details: cloneErrorDetails(details), cause: cause}
}

func (err *QueryError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.cause
}

func (err *QueryError) Error() string { return contractErrorText(err.Code, err.ReasonCode) }

func (err *QueryError) Retryable() bool {
	return err != nil && err.Code == "projection_not_available" && (err.ReasonCode == "creating" || err.ReasonCode == "refreshing" || err.ReasonCode == "accepted" || err.ReasonCode == "computing")
}

func (err *QueryError) EnvelopeJSON() ([]byte, error) {
	orders := map[string][]string{
		"invalid_argument":           {"field", "reason_code"},
		"graph_view_not_found":       {"graph_view_id"},
		"projection_not_available":   {"graph_view_id", "state"},
		"projection_run_not_found":   {"graph_view_id", "projection_run_id"},
		"projection_run_failed":      {"graph_view_id", "projection_run_id"},
		"projection_run_invalidated": {"graph_view_id", "projection_run_id", "invalidation"},
		"vertex_not_found":           {"graph_view_id", "projection_run_id", "vertex_id"},
		"edge_not_found":             {"graph_view_id", "projection_run_id", "edge_id"},
		"cursor_invalid":             {"reason_code"},
	}
	order, ok := orders[err.Code]
	if !ok {
		return nil, fmt.Errorf("unregistered query error code %q", err.Code)
	}
	return marshalErrorEnvelope(err.Code, err.Retryable(), err.Details, order)
}

func IsLifecycleError(err error, code, reason string) bool {
	var contractErr *LifecycleError
	return errors.As(err, &contractErr) && contractErr.Code == code && contractErr.ReasonCode == reason
}

func IsQueryError(err error, code, reason string) bool {
	var contractErr *QueryError
	return errors.As(err, &contractErr) && contractErr.Code == code && contractErr.ReasonCode == reason
}

func contractErrorText(code, reason string) string {
	if reason == "" {
		return code
	}
	return code + ": " + reason
}

func cloneErrorDetails(details map[string]any) map[string]any {
	if details == nil {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(details))
	for key, value := range details {
		cloned[key] = value
	}
	return cloned
}

func marshalErrorEnvelope(code string, retryable bool, details map[string]any, order []string) ([]byte, error) {
	allowed := make(map[string]bool, len(order))
	for _, key := range order {
		allowed[key] = true
	}
	unknown := make([]string, 0)
	for key := range details {
		if !allowed[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return nil, fmt.Errorf("error %s has unregistered details %v", code, unknown)
	}
	var output bytes.Buffer
	output.WriteString(`{"status":"error","error":{"code":`)
	codeJSON, _ := json.Marshal(code)
	output.Write(codeJSON)
	output.WriteString(`,"retryable":`)
	retryableJSON, _ := json.Marshal(retryable)
	output.Write(retryableJSON)
	output.WriteString(`,"details":{`)
	written := 0
	for _, key := range order {
		value, present := details[key]
		if !present {
			continue
		}
		if written > 0 {
			output.WriteByte(',')
		}
		keyJSON, _ := json.Marshal(key)
		valueJSON, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("marshal error detail %s: %w", key, err)
		}
		output.Write(keyJSON)
		output.WriteByte(':')
		output.Write(valueJSON)
		written++
	}
	output.WriteString(`}}}`)
	return output.Bytes(), nil
}
