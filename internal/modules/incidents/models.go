package incidents

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	errInvalidMutationAdmission  = errors.New("incidents: invalid mutation admission")
	errInvalidMutationTime       = errors.New("incidents: invalid mutation time")
	ErrIncidentNotFound          = errors.New("incidents: incident not found")
	ErrIncidentKeyConflict       = errors.New("incidents: incident key conflict")
	ErrIncidentVersionConflict   = errors.New("incidents: incident version conflict")
	ErrMembershipNotFound        = errors.New("incidents: membership not found")
	ErrMembershipExistsUsePatch  = errors.New("incidents: membership exists use patch")
	ErrMembershipVersionConflict = errors.New("incidents: membership version conflict")
	ErrLastIncidentAdmin         = errors.New("incidents: last incident admin")
	ErrIncidentIllegalTransition = errors.New("incidents: illegal incident transition")
)

// IncidentVersionConflictError carries the optimistic-concurrency values needed
// for clients to reconcile a stale incident metadata patch.
type IncidentVersionConflictError struct {
	IncidentID             uuid.UUID
	BaseIncidentVersion    int64
	CurrentIncidentVersion int64
}

type IncidentKeyConflictError struct {
	incidentKeyCanonical string
}

func (e *IncidentKeyConflictError) Error() string { return ErrIncidentKeyConflict.Error() }
func (e *IncidentKeyConflictError) Unwrap() error { return ErrIncidentKeyConflict }

func (e *IncidentKeyConflictError) IncidentKeyCanonical() string {
	if e == nil {
		return ""
	}
	return e.incidentKeyCanonical
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
	Incident IncidentRecord
	Payload  map[string]any
	Created  bool
}

type MembershipCreateResult struct {
	Membership MembershipRecord
	Payload    map[string]any
	Created    bool
}

type IncidentLifecycleResult struct {
	Incident IncidentRecord
	Payload  map[string]any
	Commit   TerminalMutationCommit
}

type MembershipDeleteResult struct {
	Commit TerminalMutationCommit
}

type terminalMutationDisposition uint8

const (
	terminalMutationNewCommit terminalMutationDisposition = iota + 1
	terminalMutationReplay
)

// TerminalMutationCommit exposes only validated queries about whether an
// application mutation committed new state and, for a new commit, the
// administrative-audit event UUID committed in the same transaction.
type TerminalMutationCommit struct {
	disposition terminalMutationDisposition
	effectKey   uuid.UUID
}

func NewTerminalMutationCommit(effectKey uuid.UUID) (TerminalMutationCommit, error) {
	if effectKey == uuid.Nil {
		return TerminalMutationCommit{}, errors.New("incidents: new terminal mutation commit requires an effect key")
	}
	return TerminalMutationCommit{disposition: terminalMutationNewCommit, effectKey: effectKey}, nil
}

func ReplayTerminalMutationCommit() TerminalMutationCommit {
	return TerminalMutationCommit{disposition: terminalMutationReplay}
}

func (commit TerminalMutationCommit) IsNewCommit() bool {
	return commit.disposition == terminalMutationNewCommit && commit.effectKey != uuid.Nil
}

func (commit TerminalMutationCommit) IsReplay() bool {
	return commit.disposition == terminalMutationReplay && commit.effectKey == uuid.Nil
}

func (commit TerminalMutationCommit) EffectKey() (uuid.UUID, bool) {
	if !commit.IsNewCommit() {
		return uuid.Nil, false
	}
	return commit.effectKey, true
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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
