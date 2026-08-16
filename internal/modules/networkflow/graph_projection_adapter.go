package networkflow

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	graphProjectionSchemaID = graphprojection.ProjectionSchemaIDV2
	graphSourceOwnerID      = "network_flow_activity"
)

type graphProjectionPort interface {
	GraphViewID(string) (string, error)
	ProjectEphemeral(context.Context, string, json.RawMessage) (map[string]any, error)
	ProjectSaved(context.Context, string, json.RawMessage, func(context.Context) error) (graphprojection.ProjectionResultV2, error)
}

func (a *graphProjectionAdapter) ProjectSaved(ctx context.Context, graphViewID string, input json.RawMessage, cancellationCheck func(context.Context) error) (graphprojection.ProjectionResultV2, error) {
	result, err := a.engine.Project(ctx, graphprojection.InvocationContextV2{GraphViewID: graphViewID, SourceOwnerID: graphSourceOwnerID, CancellationCheck: cancellationCheck}, input)
	if err == nil {
		return result, nil
	}
	var projectionErr *graphprojection.ProjectionErrorV2
	if errors.As(err, &projectionErr) {
		reason := "projection_unavailable"
		if projectionErr.Code == "invalid_projection_request" || projectionErr.Code == "projection_validation_failed" || projectionErr.Code == "projection_resource_limit_exceeded" {
			reason = "adapter_contract_rejected"
		}
		return graphprojection.ProjectionResultV2{}, &graphProjectionAdapterError{cause: err, reason: reason}
	}
	return graphprojection.ProjectionResultV2{}, &graphProjectionAdapterError{cause: err, reason: "projection_unavailable"}
}

type graphProjectionAdapter struct {
	engine *graphprojection.EngineV2
}

type graphProjectionAdapterError struct {
	cause  error
	reason string
}

func (e *graphProjectionAdapterError) Error() string {
	return "Network Flow graph projection adapter: " + e.reason
}

func (e *graphProjectionAdapterError) Unwrap() error {
	return e.cause
}

func newGraphProjectionAdapter() graphProjectionPort {
	return &graphProjectionAdapter{engine: graphprojection.NewEngineV2()}
}

func (a *graphProjectionAdapter) GraphViewID(key string) (string, error) {
	return graphprojection.DeriveGraphViewIDV2(graphSourceOwnerID, key)
}

func (a *graphProjectionAdapter) ProjectEphemeral(ctx context.Context, graphViewID string, input json.RawMessage) (map[string]any, error) {
	result, err := a.engine.Project(ctx, graphprojection.InvocationContextV2{GraphViewID: graphViewID, SourceOwnerID: graphSourceOwnerID}, input)
	if err == nil {
		return result.Resource(), nil
	}
	var projectionErr *graphprojection.ProjectionErrorV2
	if errors.As(err, &projectionErr) {
		if projectionErr.Code == "invalid_projection_request" || projectionErr.Code == "projection_validation_failed" || projectionErr.Code == "projection_resource_limit_exceeded" {
			return nil, &graphProjectionAdapterError{cause: err, reason: "adapter_contract_rejected"}
		}
		return nil, &graphProjectionAdapterError{cause: err, reason: "projection_unavailable"}
	}
	return nil, &graphProjectionAdapterError{cause: err, reason: "projection_unavailable"}
}

// graphProjectionFailedForProjectionError is retained beside the sole provider
// adapter so provider-specific errors cannot leak into the composer or routes.
func graphProjectionFailedForProjectionError(err error) *httpapi.APIError {
	var projectionErr *graphprojection.ProjectionErrorV2
	if errors.As(err, &projectionErr) {
		if projectionErr.Code == "invalid_projection_request" || projectionErr.Code == "projection_validation_failed" || projectionErr.Code == "projection_resource_limit_exceeded" {
			return graphProjectionFailed("adapter_contract_rejected")
		}
	}
	return graphProjectionFailed("projection_unavailable")
}
