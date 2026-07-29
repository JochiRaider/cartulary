package incidents

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrIncidentNotFound          = errors.New("incidents: incident not found")
	ErrIncidentKeyConflict       = errors.New("incidents: incident key conflict")
	ErrIncidentVersionConflict   = errors.New("incidents: incident version conflict")
	ErrMembershipNotFound        = errors.New("incidents: membership not found")
	ErrMembershipExistsUsePatch  = errors.New("incidents: membership exists use patch")
	ErrMembershipVersionConflict = errors.New("incidents: membership version conflict")
	ErrLastIncidentAdmin         = errors.New("incidents: last incident admin")
	ErrIncidentClosed            = errors.New("incidents: incident closed")
	ErrIncidentIllegalTransition = errors.New("incidents: illegal incident transition")
	ErrInitialAdminUnavailable   = errors.New("incidents: initial admin unavailable")
	ErrIncidentRoleDenied        = errors.New("incidents: incident role denied")
)

// IncidentVersionConflictError carries the optimistic-concurrency values needed
// for clients to reconcile a stale incident metadata patch.
type IncidentVersionConflictError struct {
	IncidentID             uuid.UUID
	BaseIncidentVersion    int64
	CurrentIncidentVersion int64
}

func (e *IncidentVersionConflictError) Error() string {
	return ErrIncidentVersionConflict.Error()
}

func (e *IncidentVersionConflictError) Unwrap() error {
	return ErrIncidentVersionConflict
}

func (e *IncidentVersionConflictError) Details() map[string]any {
	return map[string]any{
		"incident_id":              e.IncidentID.String(),
		"base_incident_version":    e.BaseIncidentVersion,
		"current_incident_version": e.CurrentIncidentVersion,
	}
}

type IncidentListPosition struct {
	UpdatedAt time.Time
	ID        uuid.UUID
}

type IncidentListPageRequest struct {
	AnchorUpdatedAt *time.Time
	After           *IncidentListPosition
	Limit           int
	SearchTokens    []string
	Status          *string
}

type IncidentRecord struct {
	ID                     uuid.UUID
	IncidentKey            string
	Title                  string
	Description            *string
	Status                 string
	Severity               *string
	TLP                    *string
	CurrentPhase           *string
	PrimaryExternalCaseRef *string
	CreatedByUserID        uuid.UUID
	CreatedAt              time.Time
	UpdatedAt              time.Time
	UpdatedByUserID        *uuid.UUID
	IncidentVersion        int64
	ClosedAt               *time.Time
}

type MembershipRecord struct {
	IncidentID        uuid.UUID
	UserID            uuid.UUID
	DisplayName       string
	Role              string
	JoinedAt          time.Time
	AddedByUserID     uuid.UUID
	UpdatedAt         time.Time
	UpdatedByUserID   *uuid.UUID
	MembershipVersion int64
}

type CreateIncidentResult struct {
	Incident   IncidentRecord
	Payload    map[string]any
	StatusCode int
	Location   string
}

type MembershipCreateResult struct {
	Membership MembershipRecord
	Payload    map[string]any
	StatusCode int
}

type IncidentLifecycleResult struct {
	Incident   IncidentRecord
	Payload    map[string]any
	StatusCode int
	Replayed   bool
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func incidentLocation(incidentID uuid.UUID) string {
	return "/api/v1/incidents/" + incidentID.String()
}

func extractUUID(value any) (uuid.UUID, error) {
	text, ok := value.(string)
	if !ok || text == "" {
		return uuid.UUID{}, errors.New("missing uuid string")
	}
	return uuid.Parse(text)
}

func stringPointersEqual(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
