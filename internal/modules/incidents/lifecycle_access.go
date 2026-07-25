package incidents

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type IncidentCreateBootstrap struct {
	CreatorRole string
}

func DefaultIncidentCreateBootstrap() IncidentCreateBootstrap {
	return IncidentCreateBootstrap{
		CreatorRole: "admin",
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

func IncidentLifecycleRequestHash(action string, request IncidentLifecycleRequest) []byte {
	return hashRequestPayload(map[string]any{
		"action_route":          action,
		"base_incident_version": request.BaseIncidentVersion,
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

func RequireIncidentMembership(ctx context.Context, access Access, incidentID uuid.UUID, userID uuid.UUID) (MembershipRecord, *httpapi.APIError) {
	if access == nil {
		return MembershipRecord{}, httpapi.InternalAPIError(errors.New("incidents access is required"))
	}
	record, err := access.GetIncidentMembershipForUser(ctx, incidentID, userID)
	if access.IsMembershipNotFound(err) || errors.Is(err, ErrMembershipNotFound) {
		return MembershipRecord{}, IncidentAccessError(nil, false)
	}
	if err != nil {
		return MembershipRecord{}, httpapi.InternalAPIError(err)
	}
	return record, nil
}

func RequireIncidentRole(ctx context.Context, access Access, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (MembershipRecord, *httpapi.APIError) {
	record, apiErr := RequireIncidentMembership(ctx, access, incidentID, userID)
	if apiErr != nil {
		return MembershipRecord{}, apiErr
	}
	if apiErr := IncidentAccessError(&record, false, roles...); apiErr != nil {
		return MembershipRecord{}, apiErr
	}
	return record, nil
}
