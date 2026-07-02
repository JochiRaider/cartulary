package incidents

import (
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type IncidentCreateBootstrap struct {
	CreatorRole                        string
	CreatesIncidentWorkbookPreferences bool
	CreatesUserWorkbookPreferences     bool
}

func DefaultIncidentCreateBootstrap() IncidentCreateBootstrap {
	return IncidentCreateBootstrap{
		CreatorRole:                        "admin",
		CreatesIncidentWorkbookPreferences: true,
		CreatesUserWorkbookPreferences:     true,
	}
}

func IncidentCreateIdempotencyScope() string {
	return "actor"
}

func IncidentCreateRequestHash(request CreateIncidentRequest) []byte {
	return hashRequestPayload(map[string]any{
		"client_txn_id":             request.ClientTxnID,
		"incident_key":              request.IncidentKey,
		"title":                     request.Title,
		"description":               request.Description,
		"severity":                  request.Severity,
		"tlp":                       request.TLP,
		"current_phase":             request.CurrentPhase,
		"primary_external_case_ref": request.PrimaryExternalCaseRef,
	})
}

func IncidentLifecycleRequestHash(request IncidentLifecycleRequest) []byte {
	return hashRequestPayload(map[string]any{
		"base_incident_version": request.BaseIncidentVersion,
		"client_txn_id":         request.ClientTxnID,
		"reason":                request.Reason,
	})
}

func ApplyIncidentPatch(current IncidentRecord, request IncidentPatchRequest, actorUserID uuid.UUID, updatedAt time.Time) (IncidentRecord, bool) {
	next := current
	if request.Description.Present {
		next.Description = request.Description.Value
	}
	if request.Severity.Present {
		next.Severity = request.Severity.Value
	}
	if request.TLP.Present {
		next.TLP = request.TLP.Value
	}
	if request.CurrentPhase.Present {
		next.CurrentPhase = request.CurrentPhase.Value
	}
	if request.PrimaryExternalCaseRef.Present {
		next.PrimaryExternalCaseRef = request.PrimaryExternalCaseRef.Value
	}

	if stringPointersEqual(current.Description, next.Description) &&
		stringPointersEqual(current.Severity, next.Severity) &&
		stringPointersEqual(current.TLP, next.TLP) &&
		stringPointersEqual(current.CurrentPhase, next.CurrentPhase) &&
		stringPointersEqual(current.PrimaryExternalCaseRef, next.PrimaryExternalCaseRef) {
		return current, false
	}

	next.UpdatedAt = updatedAt.UTC()
	next.UpdatedByUserID = &actorUserID
	next.IncidentVersion = current.IncidentVersion + 1
	return next, true
}

func IncidentAccessError(membership *MembershipRecord, isDeploymentAdmin bool, roles ...string) *httpapi.APIError {
	_ = isDeploymentAdmin
	if membership == nil {
		return incidentNotFoundError()
	}
	if len(roles) == 0 {
		return nil
	}
	if !slices.Contains(roles, membership.Role) {
		return authorizationDeniedError(requiredRoleDescription(roles...))
	}
	return nil
}
