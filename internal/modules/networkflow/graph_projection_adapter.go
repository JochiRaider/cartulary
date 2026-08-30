package networkflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	graphProjectionSchemaID = graphprojection.ProjectionSchemaIDV2
	graphSourceOwnerID      = "network_flow_activity"
)

type graphProjectionPort interface {
	ProjectEphemeral(context.Context, string, json.RawMessage) (map[string]any, error)
	ProjectSaved(context.Context, string, json.RawMessage, func(context.Context) error) (graphprojection.ProjectionResultV2, error)
}

func (a *graphProjectionAdapter) ProjectSaved(ctx context.Context, graphViewID string, input json.RawMessage, cancellationCheck func(context.Context) error) (graphprojection.ProjectionResultV2, error) {
	result, err := graphprojection.ProjectV2(ctx, graphprojection.InvocationContextV2{GraphViewID: graphViewID, SourceOwnerID: graphSourceOwnerID, CancellationCheck: cancellationCheck}, input)
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

type graphProjectionAdapter struct{}

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
	return &graphProjectionAdapter{}
}

func (a *graphProjectionAdapter) ProjectEphemeral(ctx context.Context, graphViewID string, input json.RawMessage) (map[string]any, error) {
	result, err := graphprojection.ProjectV2(ctx, graphprojection.InvocationContextV2{GraphViewID: graphViewID, SourceOwnerID: graphSourceOwnerID}, input)
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

func deriveNetworkFlowGraphViewID(key string) (string, error) {
	if !validGraphProjectionIdentifier(graphSourceOwnerID) || !validGraphProjectionIdentifier(key) {
		return "", fmt.Errorf("network flow graph projection identity input is invalid")
	}
	var transcript bytes.Buffer
	for _, field := range []string{"cartulary.graph_projection_graph_view_identity.v2", graphSourceOwnerID, key} {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len([]byte(field))))
		transcript.Write(length[:])
		transcript.WriteString(field)
	}
	digest := sha256.Sum256(transcript.Bytes())
	return "gv_" + hex.EncodeToString(digest[:]), nil
}

func validGraphProjectionIdentifier(value string) bool {
	if !utf8.ValidString(value) || len([]byte(value)) == 0 || len([]byte(value)) > 255 || strings.ContainsAny(value, "/\\") {
		return false
	}
	var first, last rune
	for index, character := range value {
		if character == 0 || unicode.IsControl(character) || character >= 0x7f && character <= 0x9f {
			return false
		}
		if index == 0 {
			first = character
		}
		last = character
	}
	return !unicode.IsSpace(first) && !unicode.IsSpace(last)
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
