package networkflow

import (
	"io"
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
)

// operationAdmission lives only for this request. Each continuation authenticates
// and checks current authority again; applications also admit direct callers.
type operationAdmission struct {
	principal            httpauth.Principal
	incidentID           uuid.UUID
	tableID, graphViewID string
}

func (s *routeService) admitOperation(r *http.Request, route string) (operationAdmission, *httpapi.APIError) {
	var result operationAdmission
	roles, role := admission.RolesMember, ""
	switch route {
	case "nf.tables.patch", "nf.graph_views.create", "nf.graph_views.patch", "nf.graph_views.refresh", "nf.indicator_links.create":
		roles, role = admission.RolesEditorAdmin, "editor|admin"
	case "nf.tables.delete", "nf.graph_views.delete":
		roles, role = admission.RolesReviewerAdmin, "reviewer|admin"
	}
	principal, apiErr := s.authenticate(r, role != "")
	if apiErr != nil {
		return result, apiErr
	}
	result.principal = principal
	incidentID, err := uuid.Parse(r.PathValue("incident_id"))
	if err != nil {
		return result, &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, roles, role); apiErr != nil {
		return result, apiErr
	}
	result.incidentID = incidentID
	if strings.HasPrefix(route, "nf.tables.") && route != "nf.tables.list" || route == "nf.rejected_rows.query" {
		result.tableID = r.PathValue("network_flow_table_id")
		if !linkTableIDPattern.MatchString(result.tableID) {
			return result, semanticHTTPError(tableReadFailure(errTableNotFound))
		}
	}
	if strings.HasPrefix(route, "nf.graph_views.") && route != "nf.graph_views.list" && route != "nf.graph_views.create" {
		result.graphViewID = r.PathValue("graph_view_id")
		if !graphViewIDPattern.MatchString(result.graphViewID) {
			return result, networkFlowAPIError(http.StatusNotFound, "network_flow_graph_view_not_found", "graph_view_id", "not_found")
		}
	}
	if r.URL.RawQuery != "" {
		return result, semanticHTTPError(invalidNetworkFlowRequest("query", "unknown_member"))
	}
	if r.Method == http.MethodGet {
		if r.Body != nil {
			first, err := io.ReadAll(io.LimitReader(r.Body, 1))
			if len(first) != 0 || err != nil {
				return result, semanticHTTPError(invalidNetworkFlowRequest("body", "variant_member_conflict"))
			}
		}
	}
	return result, nil
}

func (a operationAdmission) identity() readIdentity {
	return readIdentity{ActorID: a.principal.User.ID, SessionID: a.principal.Session.ID, IncidentID: a.incidentID}
}
