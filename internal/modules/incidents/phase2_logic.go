package incidents

import (
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
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

func ApplyIncidentPatch(current IncidentRecord, request IncidentPatchRequest, actorUserID uuid.UUID, updatedAt time.Time) (IncidentRecord, bool) {
	next := current
	if request.TLP.Present {
		next.TLP = request.TLP.Value
	}
	if request.CurrentPhase.Present {
		next.CurrentPhase = request.CurrentPhase.Value
	}
	if request.PrimaryExternalCaseRef.Present {
		next.PrimaryExternalCaseRef = request.PrimaryExternalCaseRef.Value
	}

	if stringPointersEqual(current.TLP, next.TLP) &&
		stringPointersEqual(current.CurrentPhase, next.CurrentPhase) &&
		stringPointersEqual(current.PrimaryExternalCaseRef, next.PrimaryExternalCaseRef) {
		return current, false
	}

	next.UpdatedAt = updatedAt.UTC()
	next.UpdatedByUserID = &actorUserID
	next.IncidentVersion = current.IncidentVersion + 1
	return next, true
}

func BuildExtensionsResponseData(profiles []httpapi.ExtensionProfile) map[string]any {
	extensions := make([]map[string]any, 0, len(profiles))
	for _, profile := range profiles {
		extensions = append(extensions, BuildExtensionResource(profile))
	}
	return map[string]any{"extensions": extensions}
}

func IncidentAccessError(membership *MembershipRecord, isDeploymentAdmin bool, roles ...string) *auth.APIError {
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
