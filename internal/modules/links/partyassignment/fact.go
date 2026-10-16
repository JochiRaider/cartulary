// Package partyassignment owns the immutable Core assignment source projection.
// It carries source facts and equality-only version identity, never disclosure
// policy, renderer state or a synthetic revision history.
package partyassignment

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
)

// State is the closed source projection. It carries actual
// assignment provenance; it contains neither disclosure decisions nor labels.
type State struct {
	AssignmentID             uuid.UUID  `json:"assignment_id"`
	IncidentID               uuid.UUID  `json:"incident_id"`
	SubjectRecordID          uuid.UUID  `json:"subject_record_id"`
	SubjectRecordType        string     `json:"subject_record_type"`
	PartyID                  uuid.UUID  `json:"party_id"`
	AssignmentKind           string     `json:"assignment_kind"`
	AssignmentState          string     `json:"assignment_state"`
	SourceRef                *struct{}  `json:"source_ref"`
	Provenance               string     `json:"provenance"`
	CreatedByUserID          uuid.UUID  `json:"created_by_user_id"`
	CreatedAt                string     `json:"created_at"`
	SupersededByAssignmentID *uuid.UUID `json:"superseded_by_assignment_id"`
	DeletedAt                *string    `json:"deleted_at"`
}

type Fact struct {
	State
	RowVersion string `json:"row_version"`
}

// NewFact validates an active source projection and binds its exact state.
func NewFact(state State) (Fact, error) {
	invalid := errors.New("invalid party assignment state")
	if state.AssignmentID == uuid.Nil || state.IncidentID == uuid.Nil || state.SubjectRecordID == uuid.Nil || state.PartyID == uuid.Nil || state.SubjectRecordID == state.PartyID || state.CreatedByUserID == uuid.Nil ||
		(state.SubjectRecordType != "host" && state.SubjectRecordType != "identity") || state.AssignmentState != "active" ||
		state.SourceRef != nil || state.SupersededByAssignmentID != nil || state.DeletedAt != nil {
		return Fact{}, invalid
	}
	switch state.AssignmentKind {
	case "subject_party", "responsible_party", "observer_party":
	default:
		return Fact{}, invalid
	}
	switch state.Provenance {
	case "manual", "import", "system", "rollback":
	default:
		return Fact{}, invalid
	}
	stamp, err := time.Parse("2006-01-02T15:04:05.000000Z", state.CreatedAt)
	if err != nil || stamp.Year() < 1 || stamp.Year() > 9999 || stamp.Format("2006-01-02T15:04:05.000000Z") != state.CreatedAt {
		return Fact{}, invalid
	}
	raw, err := canonicaljson.Marshal(state)
	if err != nil {
		return Fact{}, err
	}
	digest := sha256.Sum256(append([]byte("cartulary.record_party_assignment_state.v1\n"), raw...))
	return Fact{State: state, RowVersion: "rpa_" + hex.EncodeToString(digest[:])}, nil
}

// Validate checks a retained fact without consulting current mutable source.
func (f Fact) Validate() error {
	expected, err := NewFact(f.State)
	if err != nil {
		return err
	}
	if expected.RowVersion != f.RowVersion {
		return errors.New("invalid assignment state version")
	}
	return nil
}
