package graphprojection

import "errors"

func wrapQueryContractError(err error, graphViewID, projectionRunID, vertexID, edgeID string) error {
	if err == nil {
		return nil
	}
	var queryErr *QueryError
	if errors.As(err, &queryErr) {
		return err
	}
	switch {
	case errors.Is(err, ErrGraphViewNotFound):
		return NewQueryError("graph_view_not_found", "", map[string]any{"graph_view_id": graphViewID}, ErrGraphViewNotFound)
	case errors.Is(err, ErrProjectionRunNotFound):
		return NewQueryError("projection_run_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID}, ErrProjectionRunNotFound)
	case errors.Is(err, ErrVertexNotFound):
		return NewQueryError("vertex_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID, "vertex_id": vertexID}, ErrVertexNotFound)
	case errors.Is(err, ErrEdgeNotFound):
		return NewQueryError("edge_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID, "edge_id": edgeID}, ErrEdgeNotFound)
	case errors.Is(err, ErrGraphViewUnavailable):
		return NewQueryError("projection_not_available", "", map[string]any{"graph_view_id": graphViewID, "state": ""}, ErrGraphViewUnavailable)
	default:
		return err
	}
}

func wrapLifecycleContractError(err error, operation, graphViewID, projectionRunID string) error {
	if err == nil {
		return nil
	}
	var lifecycleErr *LifecycleError
	if errors.As(err, &lifecycleErr) {
		return err
	}
	switch {
	case errors.Is(err, ErrGraphViewNotFound):
		return lifecycleGraphViewNotFound(operation, graphViewID)
	case errors.Is(err, ErrProjectionRunNotFound):
		return &LifecycleError{Code: "projection_run_not_found", Details: map[string]any{"operation": operation, "graph_view_id": graphViewID, "projection_run_id": projectionRunID}}
	default:
		return err
	}
}

func lifecycleGraphViewNotFound(operation, graphViewID string) *LifecycleError {
	return &LifecycleError{Code: "graph_view_not_found", Details: map[string]any{"operation": operation, "graph_view_id": graphViewID}}
}

func invalidOperationRequest(operation, field, reason string) *LifecycleError {
	return &LifecycleError{Code: "invalid_operation_request", ReasonCode: reason, Field: field, Details: map[string]any{"operation": operation, "field": field, "reason_code": reason}}
}
