package incidents

import "github.com/google/uuid"

type CreateIncidentRequest struct {
	ClientTxnID            string
	IncidentKey            string
	Title                  string
	Description            *string
	Severity               *string
	TLP                    *string
	CurrentPhase           *string
	PrimaryExternalCaseRef *string
}

type MembershipCreateRequest struct {
	ClientTxnID string
	UserID      *uuid.UUID
	Email       *string
	Role        string
}

type MembershipPatchRequest struct {
	BaseMembershipVersion int64
	Role                  string
}

type MembershipDeleteRequest struct {
	BaseMembershipVersion int64
}

type OptionalNullableString struct {
	Present bool
	Value   *string
}

type IncidentPatchRequest struct {
	BaseIncidentVersion    int64
	Description            OptionalNullableString
	Severity               OptionalNullableString
	TLP                    OptionalNullableString
	CurrentPhase           OptionalNullableString
	PrimaryExternalCaseRef OptionalNullableString
}

type IncidentLifecycleRequest struct {
	BaseIncidentVersion int64
	ClientTxnID         string
	Reason              string
}

func BuildIncidentResource(record IncidentRecord) map[string]any {
	return map[string]any{
		"incident_id":               record.ID,
		"incident_key":              record.IncidentKey,
		"title":                     record.Title,
		"description":               record.Description,
		"status":                    record.Status,
		"severity":                  record.Severity,
		"tlp":                       record.TLP,
		"current_phase":             record.CurrentPhase,
		"primary_external_case_ref": record.PrimaryExternalCaseRef,
		"created_by_user_id":        record.CreatedByUserID,
		"created_at":                record.CreatedAt,
		"updated_at":                record.UpdatedAt,
		"updated_by_user_id":        record.UpdatedByUserID,
		"incident_version":          record.IncidentVersion,
		"closed_at":                 record.ClosedAt,
	}
}

func BuildMembershipResource(record MembershipRecord) map[string]any {
	return map[string]any{
		"incident_id":        record.IncidentID,
		"user_id":            record.UserID,
		"display_name":       record.DisplayName,
		"role":               record.Role,
		"joined_at":          record.JoinedAt,
		"added_by_user_id":   record.AddedByUserID,
		"updated_at":         record.UpdatedAt,
		"updated_by_user_id": record.UpdatedByUserID,
		"membership_version": record.MembershipVersion,
	}
}
