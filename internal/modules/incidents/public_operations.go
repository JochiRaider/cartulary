package incidents

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		incidentPublicOperation(http.MethodGet, "/api/v1/incidents", "listVisibleIncidents", false, http.StatusOK),
		incidentPublicOperation(http.MethodPost, "/api/v1/incidents", "createIncident", true, http.StatusCreated),
		incidentPublicOperation(http.MethodGet, "/api/v1/incidents/{incident_id}", "getIncident", false, http.StatusOK),
		incidentPublicOperation(http.MethodPatch, "/api/v1/incidents/{incident_id}", "patchIncident", true, http.StatusOK),
		incidentPublicOperation(http.MethodPost, "/api/v1/incidents/{incident_id}/close", "closeIncident", true, http.StatusOK),
		incidentPublicOperation(http.MethodGet, "/api/v1/incidents/{incident_id}/memberships", "listIncidentMemberships", false, http.StatusOK),
		incidentPublicOperation(http.MethodGet, "/api/v1/incidents/{incident_id}/membership-audit-events", "listIncidentMembershipAuditEvents", false, http.StatusOK),
		incidentPublicOperation(http.MethodPost, "/api/v1/incidents/{incident_id}/memberships", "createIncidentMembership", true, http.StatusCreated),
		incidentPublicOperation(http.MethodDelete, "/api/v1/incidents/{incident_id}/memberships/{user_id}", "deleteIncidentMembership", true, http.StatusNoContent),
		incidentPublicOperation(http.MethodPatch, "/api/v1/incidents/{incident_id}/memberships/{user_id}", "patchIncidentMembership", true, http.StatusOK),
		incidentPublicOperation(http.MethodPost, "/api/v1/incidents/{incident_id}/reopen", "reopenIncident", true, http.StatusOK),
	}
}

func incidentPublicOperation(
	method string,
	pathTemplate string,
	operationID string,
	stateChanging bool,
	successStatus int,
) httpapi.PublicOperation {
	return httpapi.PublicOperation{
		OwnerID:        "module.incidents",
		Method:         method,
		PathTemplate:   pathTemplate,
		OperationID:    operationID,
		Authentication: httpapi.PublicAuthenticationSession,
		StateChanging:  stateChanging,
		SuccessStatus:  successStatus,
	}
}
