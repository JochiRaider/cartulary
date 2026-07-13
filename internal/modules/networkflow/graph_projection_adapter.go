package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const graphProjectionSchemaID = "graph_projection.v1"

type graphProjectionPort interface {
	GraphViewID(string) (string, error)
	ProjectEphemeral(context.Context, json.RawMessage) (map[string]any, error)
}

type graphProjectionAdapter struct {
	service *graphprojection.Service
}

type graphProjectionAdapterError struct {
	reason string
}

func (e *graphProjectionAdapterError) Error() string {
	return "Network Flow graph projection adapter: " + e.reason
}

func newGraphProjectionAdapter(now func() time.Time) graphProjectionPort {
	return &graphProjectionAdapter{service: graphprojection.NewService(graphprojection.ServiceOptions{Now: now})}
}

func (a *graphProjectionAdapter) GraphViewID(key string) (string, error) {
	return a.service.GraphViewID(key)
}

func (a *graphProjectionAdapter) ProjectEphemeral(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	result, err := a.service.ProjectEphemeral(ctx, graphprojection.EphemeralProjectionRequest{ProjectionInput: input})
	if err == nil {
		return result.Resource(), nil
	}
	var lifecycleErr *graphprojection.LifecycleError
	if errors.As(err, &lifecycleErr) {
		if lifecycleErr.Code == "invalid_projection_request" || lifecycleErr.Code == "ephemeral_projection_failed" && lifecycleErr.ReasonCode == "fatal_validation" {
			return nil, &graphProjectionAdapterError{reason: "adapter_contract_rejected"}
		}
		return nil, &graphProjectionAdapterError{reason: "projection_unavailable"}
	}
	return nil, &graphProjectionAdapterError{reason: "projection_unavailable"}
}

// graphProjectionFailedForProjectionError is retained beside the sole provider
// adapter so provider-specific errors cannot leak into the composer or routes.
func graphProjectionFailedForProjectionError(err error) *httpapi.APIError {
	var lifecycleErr *graphprojection.LifecycleError
	if errors.As(err, &lifecycleErr) {
		if lifecycleErr.Code == "invalid_projection_request" || lifecycleErr.Code == "ephemeral_projection_failed" && lifecycleErr.ReasonCode == "fatal_validation" {
			return graphProjectionFailed("adapter_contract_rejected")
		}
	}
	return graphProjectionFailed("projection_unavailable")
}
