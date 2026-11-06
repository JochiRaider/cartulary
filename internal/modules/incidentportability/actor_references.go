package incidentportability

import (
	"errors"
	"github.com/google/uuid"
)

// ActorReferences collects explicit source attribution before the bundle owner
// closes its historical roster. It cannot create accounts or grant permissions.
type ActorReferences interface {
	RecordActorReference(uuid.UUID) error
}

// ExportActorID selects the historical attribution, when present, and records
// the exact portable actor. Owners call it only for declared actor fields.
func ExportActorID(references ActorReferences, historical string, local uuid.UUID) (string, error) {
	id := local
	if historical != "" {
		var err error
		id, err = uuid.Parse(historical)
		if err != nil || id.String() != historical {
			return "", errors.New("portable actor identity is invalid")
		}
	}
	if id == uuid.Nil {
		return "", errors.New("portable actor identity is missing")
	}
	if references != nil {
		if err := references.RecordActorReference(id); err != nil {
			return "", err
		}
	}
	return id.String(), nil
}
