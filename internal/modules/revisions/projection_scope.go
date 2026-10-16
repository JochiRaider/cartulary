package revisions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ProjectionDependencyReader belongs to an authoritative relationship owner.
// It returns direct read-model dependents, including inactive endpoint records
// needed for restore. It must not traverse the graph transitively or mutate it.
type ProjectionDependencyReader interface {
	ProjectionDependentsTx(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) ([]uuid.UUID, error)
}

func (s *commandStore) projectionScopeTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, changed []uuid.UUID) ([]uuid.UUID, error) {
	changed = canonicalRecordIDs(changed)
	if len(changed) == 0 {
		return nil, fmt.Errorf("projection refresh requires an explicit mutation scope")
	}
	scope := append([]uuid.UUID(nil), changed...)
	for _, reader := range s.projectionDependencies {
		dependents, err := reader.ProjectionDependentsTx(ctx, tx, incidentID, changed)
		if err != nil {
			return nil, err
		}
		for _, id := range dependents {
			if id == uuid.Nil {
				return nil, fmt.Errorf("projection dependency has no record identity")
			}
		}
		scope = append(scope, dependents...)
	}
	scope = canonicalRecordIDs(scope)
	for _, id := range scope {
		envelope, err := s.envelopes.LoadEnvelopeTx(ctx, tx, id, false)
		if err != nil {
			return nil, adaptEnvelopeError(err)
		}
		if envelope.IncidentID != incidentID {
			return nil, fmt.Errorf("projection dependency crosses incident boundary")
		}
		if _, ok := s.deleteRestoreSources.Source(envelope.RecordType); !ok {
			return nil, fmt.Errorf("projection dependency has no source owner: %s", envelope.RecordType)
		}
	}
	return scope, nil
}

func (s *commandStore) refreshProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, scope []uuid.UUID) error {
	if len(scope) == 0 {
		return fmt.Errorf("projection refresh requires a protected scope")
	}
	for _, id := range canonicalRecordIDs(scope) {
		envelope, err := s.envelopes.LoadEnvelopeTx(ctx, tx, id, false)
		if err != nil {
			return adaptEnvelopeError(err)
		}
		if envelope.IncidentID != incidentID {
			return fmt.Errorf("projection refresh crosses incident boundary")
		}
		source, ok := s.deleteRestoreSources.Source(envelope.RecordType)
		if !ok {
			return ErrUnsupportedRecordType
		}
		view, err := source.ViewSchemaID(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := s.projections.RefreshRowTx(ctx, tx, view, id); err != nil {
			return err
		}
	}
	return nil
}
