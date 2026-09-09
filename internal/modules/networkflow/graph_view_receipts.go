package networkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

var errGraphViewReceiptInvalid = errors.New("saved graph admission receipt is incompatible")

func graphViewMutationBytes(route, pathIdentity string, body map[string]any) []byte {
	var input bytes.Buffer
	for _, value := range []string{"cartulary.network_flow.mutation_request_digest.v1", route, pathIdentity} {
		input.WriteString(value)
		input.WriteByte(0)
	}
	input.Write(canonicalJSON(body))
	input.WriteByte(0)
	return input.Bytes()
}

func (s *Service) requireSavedGraphRole(ctx context.Context, incidentID, actorID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, actorID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleOpen})
	return grant, savedGraphAdmissionError(err, requiredRole)
}

func savedGraphAdmissionError(err error, requiredRole string) *httpapi.APIError {
	switch {
	case err == nil:
		return nil
	case admission.IsDenied(err, admission.DenialNotVisible):
		return &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return &httpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Message: "authorization denied", Details: map[string]any{"required_role": requiredRole}}
	case admission.IsDenied(err, admission.DenialIncidentClosed):
		return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
	default:
		return httpapi.InternalAPIError(err)
	}
}

// GraphViewReceiptReconciler implements the finalizer's transaction-bound
// receipt port. Auth owns receipt storage; Network Flow owns receipt meaning.
// It deliberately performs no UPDATE and needs no current declaration or job
// handler payload, so an acknowledgement survives later lifecycle transitions.
type GraphViewReceiptReconciler struct{}

func (GraphViewReceiptReconciler) ReconcileFinalIdempotencyOutcomeTx(ctx context.Context, tx pgx.Tx, key jobs.RouteIdempotencyKey, requestHash []byte, job jobs.Resource) (bool, error) {
	authKey := authn.RouteIdempotencyKey{RouteKey: key.RouteKey, ActorUserID: key.ActorUserID, ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID}
	record, err := authn.GetRouteIdempotencyTx(ctx, tx, authKey)
	if errors.Is(err, authn.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !bytes.Equal(record.RequestHash, requestHash) {
		return false, errGraphViewReceiptInvalid
	}
	payload, err := decodeStoredNetworkFlowResponse(record.ResponseJSON)
	if err != nil {
		return false, errGraphViewReceiptInvalid
	}
	if err := validateGraphViewReceipt(authKey, record.StatusCode, payload, &job); err != nil {
		return false, err
	}
	return true, nil
}

func validateGraphViewReceipt(key authn.RouteIdempotencyKey, status int, payload map[string]any, job *jobs.Resource) error {
	incidentText, pathIdentity, ok := strings.Cut(key.ScopeKey, ":")
	incidentID, err := uuid.Parse(incidentText)
	if !ok || err != nil || incidentID == uuid.Nil || key.ActorUserID == uuid.Nil || key.ClientTxnID == "" {
		return errGraphViewReceiptInvalid
	}
	if key.RouteKey == routeKeyGraphViewsDelete {
		if status != http.StatusNoContent || len(payload) != 0 || job != nil || !strings.HasPrefix(pathIdentity, "graph_view_id:") || !graphViewIDPattern.MatchString(strings.TrimPrefix(pathIdentity, "graph_view_id:")) {
			return errGraphViewReceiptInvalid
		}
		return nil
	}
	graph, ok := payload["graph_view"].(map[string]any)
	if !ok {
		return errGraphViewReceiptInvalid
	}
	declaration, err := graphViewDeclarationFromPublicResource(graph)
	if err != nil || declaration.IncidentID != incidentID || declaration.DeclarationState != GraphViewDeclarationStateActive {
		return errGraphViewReceiptInvalid
	}
	if key.RouteKey == routeKeyGraphViewsCreate {
		if pathIdentity != "graph-views" || declaration.CreatedByUserID != key.ActorUserID || declaration.GraphViewVersion != 1 || declaration.MaterializationGeneration != 1 {
			return errGraphViewReceiptInvalid
		}
	} else if pathIdentity != "graph_view_id:"+declaration.GraphViewID {
		return errGraphViewReceiptInvalid
	}
	if key.RouteKey == routeKeyGraphViewsPatch {
		if status != http.StatusOK || len(payload) != 2 || payload["schema_id"] != "cartulary.network_flow.graph_view_mutation_result.v4" || job != nil {
			return errGraphViewReceiptInvalid
		}
		return nil
	}
	if key.RouteKey != routeKeyGraphViewsCreate && key.RouteKey != routeKeyGraphViewsRefresh {
		return errGraphViewReceiptInvalid
	}
	reference, ok := payload["job"].(map[string]any)
	if !ok || len(reference) != 2 || len(payload) != 3 || status != http.StatusAccepted || payload["schema_id"] != "cartulary.network_flow.graph_view_accepted.v4" || declaration.LatestJobID == nil {
		return errGraphViewReceiptInvalid
	}
	jobID := declaration.LatestJobID.String()
	if reference["job_id"] != jobID || reference["status_route"] != "/api/v1/jobs/"+jobID {
		return errGraphViewReceiptInvalid
	}
	if job != nil {
		if job.JobID != jobID || job.StatusRoute != reference["status_route"] || job.Scope.Kind != jobs.ScopeKindIncident || job.Scope.IncidentID == nil || *job.Scope.IncidentID != incidentID {
			return errGraphViewReceiptInvalid
		}
		if job.Status == jobs.StatusSucceeded {
			if job.ResultSummary == nil || job.ResultSummary.Code != "network_flow_graph_view_materialized" || len(job.ResultSummary.ResourceRefs) != 1 {
				return errGraphViewReceiptInvalid
			}
			ref := job.ResultSummary.ResourceRefs[0]
			if ref.Kind != graphViewResultResourceKind || ref.ID != declaration.GraphViewID || ref.Route != graphViewRoute(incidentID, declaration.GraphViewID) {
				return errGraphViewReceiptInvalid
			}
		}
	}
	return nil
}

// This decoder validates the public projection, never reconstructs a current
// declaration. Retained acknowledgements are historical admission evidence.
func graphViewDeclarationFromPublicResource(value map[string]any) (GraphViewDeclaration, error) {
	keys := []string{"schema_id", "graph_view_id", "incident_id", "display_name", "state", "semantic_query", "semantic_query_sha256", "desired_source_snapshot_id", "selected_result_binding", "graph_view_version", "materialization_generation", "created_by", "created_at", "updated_at", "latest_job_id", "last_failure_code", "last_failed_at"}
	if len(value) != len(keys) {
		return GraphViewDeclaration{}, errGraphViewReceiptInvalid
	}
	for _, key := range keys {
		if _, present := value[key]; !present {
			return GraphViewDeclaration{}, errGraphViewReceiptInvalid
		}
	}
	var resource struct {
		SchemaID                  string            `json:"schema_id"`
		GraphViewID               string            `json:"graph_view_id"`
		IncidentID                uuid.UUID         `json:"incident_id"`
		DisplayName               string            `json:"display_name"`
		State                     string            `json:"state"`
		SemanticQuery             json.RawMessage   `json:"semantic_query"`
		SemanticQuerySHA256       string            `json:"semantic_query_sha256"`
		DesiredSourceSnapshotID   string            `json:"desired_source_snapshot_id"`
		SelectedResultBinding     map[string]string `json:"selected_result_binding"`
		GraphViewVersion          int64             `json:"graph_view_version"`
		MaterializationGeneration int64             `json:"materialization_generation"`
		CreatedBy                 uuid.UUID         `json:"created_by"`
		CreatedAt                 time.Time         `json:"created_at"`
		UpdatedAt                 time.Time         `json:"updated_at"`
		LatestJobID               *uuid.UUID        `json:"latest_job_id"`
		LastFailureCode           *string           `json:"last_failure_code"`
		LastFailedAt              *time.Time        `json:"last_failed_at"`
	}
	if err := json.Unmarshal(canonicalJSON(value), &resource); err != nil || resource.SchemaID != "cartulary.network_flow.graph_view.v4" {
		return GraphViewDeclaration{}, errGraphViewReceiptInvalid
	}
	declaration := GraphViewDeclaration{
		GraphViewID: resource.GraphViewID, IncidentID: resource.IncidentID,
		DisplayName: resource.DisplayName, NormalizedDisplayName: strings.ToLower(resource.DisplayName), DeclarationState: resource.State,
		SemanticQueryJSON: resource.SemanticQuery, SemanticQuerySHA256: resource.SemanticQuerySHA256,
		DesiredSourceSnapshotID: resource.DesiredSourceSnapshotID, GraphViewVersion: resource.GraphViewVersion,
		MaterializationGeneration: resource.MaterializationGeneration, CreatedByUserID: resource.CreatedBy,
		CreatedAt: resource.CreatedAt, UpdatedAt: resource.UpdatedAt, LatestJobID: resource.LatestJobID,
		LastFailureCode: resource.LastFailureCode, LastFailedAt: resource.LastFailedAt,
	}
	if selected := resource.SelectedResultBinding; selected != nil {
		if len(selected) != 7 {
			return GraphViewDeclaration{}, errGraphViewReceiptInvalid
		}
		declaration.SelectedResult = &GraphViewSelectedResultBinding{
			ProjectionResultID: selected["projection_result_id"], SourceSnapshotID: selected["source_snapshot_id"],
			ProjectionSchemaID: selected["projection_schema_id"], ProjectionVersion: selected["projection_version"],
			NormalizedConfigurationSHA256: selected["normalized_configuration_sha256"], NormalizedSourceSHA256: selected["normalized_source_sha256"], CanonicalOutputSHA256: selected["canonical_output_sha256"],
		}
	}
	if !validGraphViewDeclaration(declaration) || GraphViewSemanticQuerySHA256(resource.SemanticQuery) != resource.SemanticQuerySHA256 {
		return GraphViewDeclaration{}, errGraphViewReceiptInvalid
	}
	return declaration, nil
}
