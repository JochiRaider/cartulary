package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/gen/contractnetworkflow"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
)

// routeOwnerID owns the Network Flow operations of the public route registry.
const routeOwnerID = "module.networkflow"

// operationAdmission lives only for this request. Each continuation authenticates
// and checks current authority again; applications also admit direct callers.
type operationAdmission struct {
	principal            httpauth.Principal
	incidentID           uuid.UUID
	tableID, graphViewID string
}

type admittedIncidentContextKey struct{}

// operationRoles is the role each operation requires (NF NLSpec Table 5-A).
func operationRoles(operationID string) (admission.RoleSet, string) {
	switch operationID {
	case "nf.tables.patch", "nf.graph_views.create", "nf.graph_views.patch", "nf.graph_views.refresh", "nf.indicator_links.create":
		return admission.RolesEditorAdmin, "editor|admin"
	case "nf.tables.delete", "nf.graph_views.delete":
		return admission.RolesReviewerAdmin, "reviewer|admin"
	default:
		return admission.RolesMember, ""
	}
}

// admitIncident is the Network Flow owner gate (NF NLSpec Table 21-C level 2;
// Core 01 REQ-01-707 family clause). Central admission runs it after
// authentication and cookie CSRF (level 1) and after the incident identifier,
// and before the table or graph-view identifier (level 3): it decides the
// incident's visibility, the operation's role and the open lifecycle.
func (s *routeService) admitIncident(r *http.Request, operationID string) (*http.Request, *httpapi.APIError) {
	principal, authenticated := httpauth.AdmittedPrincipal(r)
	incidentID, admitted := httpapi.PathUUID(r, "incident_id")
	if !authenticated || !admitted {
		return r, httpapi.InternalAPIError(errors.New("network flow operation was not admitted centrally"))
	}
	roles, role := operationRoles(operationID)
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, roles, role); apiErr != nil {
		return r, apiErr
	}
	return r.WithContext(context.WithValue(r.Context(), admittedIncidentContextKey{}, incidentID)), nil
}

// bodylessOperations are the operations whose route contract declares no
// request schema (NF NLSpec Table 21-C level 4), from the packaged route
// contract, never from the request's method.
var bodylessOperations = func() map[string]bool {
	var contract struct {
		Routes []struct {
			RouteID         string  `json:"route_id"`
			RequestSchemaID *string `json:"request_schema_id"`
		} `json:"routes"`
	}
	if err := json.Unmarshal([]byte(contractnetworkflow.Index["contracts/network-flow/routes.v1.json"].JSON), &contract); err != nil {
		panic(err)
	}
	bodyless := make(map[string]bool, len(contract.Routes))
	for _, route := range contract.Routes {
		if route.RequestSchemaID == nil {
			bodyless[route.RouteID] = true
		}
	}
	return bodyless
}()

// admittedOperation returns the centrally admitted request facts, then applies
// Table 21-C level 4 framing: no URL query parameter, and no body for an
// operation whose contract declares none.
func admittedOperation(r *http.Request) (operationAdmission, *httpapi.APIError) {
	var result operationAdmission
	principal, authenticated := httpauth.AdmittedPrincipal(r)
	incidentID, gated := r.Context().Value(admittedIncidentContextKey{}).(uuid.UUID)
	operation, resolved := httpapi.OperationMetadataFromContext(r.Context())
	if !authenticated || !gated || !resolved {
		return result, httpapi.InternalAPIError(errors.New("network flow operation was not admitted through its owner gate"))
	}
	result.principal, result.incidentID = principal, incidentID
	result.tableID, _ = httpapi.PathString(r, "network_flow_table_id")
	result.graphViewID, _ = httpapi.PathString(r, "graph_view_id")
	if r.URL.RawQuery != "" {
		return result, semanticHTTPError(invalidNetworkFlowRequest("query", "unknown_member"))
	}
	if bodylessOperations[operation.OperationID] && r.Body != nil {
		first, err := io.ReadAll(io.LimitReader(r.Body, 1))
		if len(first) != 0 || err != nil {
			return result, semanticHTTPError(invalidNetworkFlowRequest("body", "variant_member_conflict"))
		}
	}
	return result, nil
}

func (a operationAdmission) identity() readIdentity {
	return readIdentity{ActorID: a.principal.User.ID, SessionID: a.principal.Session.ID, IncidentID: a.incidentID}
}
