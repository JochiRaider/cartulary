package revisions

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
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

// dependentRefresh holds the served rows of a refresh scope's live dependents,
// captured before a history operation mutates any source. Delete, restore,
// rollback and row restore refresh dependents in storage; publishing their
// changed chips, counts and summaries live as well keeps every mutation path
// equal to patch (REQ-01-081, REQ-01-351) without owner-specific code.
type dependentRefresh struct {
	rows []dependentRow
}

type dependentRow struct {
	recordID     uuid.UUID
	viewSchemaID string
	source       deleterestorecontract.DeleteRestoreSource
	before       map[string]any
}

// captureDependentsTx reads the served row of every live scope member that is
// not a primary record of the operation. It must run before the operation's
// first source mutation: some served cells are computed live from links, so a
// later read would already show the new value.
func (s *commandStore) captureDependentsTx(ctx context.Context, tx pgx.Tx, scope []uuid.UUID, primary ...uuid.UUID) (dependentRefresh, error) {
	var refresh dependentRefresh
	for _, id := range canonicalRecordIDs(scope) {
		if slices.Contains(primary, id) {
			continue
		}
		envelope, err := s.envelopes.LoadEnvelopeTx(ctx, tx, id, false)
		if err != nil {
			return dependentRefresh{}, adaptEnvelopeError(err)
		}
		if envelope.DeletedAt != nil {
			continue
		}
		source, ok := s.deleteRestoreSources.Source(envelope.RecordType)
		if !ok {
			return dependentRefresh{}, ErrUnsupportedRecordType
		}
		view, err := source.ViewSchemaID(ctx, tx, id)
		if err != nil {
			return dependentRefresh{}, err
		}
		before, err := s.loadLiveRecordTx(ctx, tx, view, id, source)
		if err != nil {
			return dependentRefresh{}, err
		}
		refresh.rows = append(refresh.rows, dependentRow{recordID: id, viewSchemaID: view, source: source, before: before})
	}
	return refresh, nil
}

// publishDependentChangesTx publishes one invalidate per captured dependent
// whose served fields changed, at its unchanged row version. Ordinals follow
// every record mutation of the change set, so dependents publish after the
// operation's own records, as patch publishes its dependents.
func (s *commandStore) publishDependentChangesTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, refresh dependentRefresh) error {
	if len(refresh.rows) == 0 {
		return nil
	}
	var ordinal int
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(max(sequence_no), 0)
  FROM change_set_mutations
 WHERE change_set_id = $1
`, changeSetID).Scan(&ordinal); err != nil {
		return fmt.Errorf("load dependent publication ordinal: %w", err)
	}
	for _, dependent := range refresh.rows {
		after, err := s.loadLiveRecordTx(ctx, tx, dependent.viewSchemaID, dependent.recordID, dependent.source)
		if err != nil {
			return err
		}
		changedFieldKeys := rollbackChangedFieldKeys(dependent.before, after)
		if len(changedFieldKeys) == 0 {
			continue
		}
		envelope, err := s.envelopes.LoadEnvelopeTx(ctx, tx, dependent.recordID, false)
		if err != nil {
			return adaptEnvelopeError(err)
		}
		effect, err := s.recordPublicationEffectTx(ctx, tx, changeSetID, dependent.recordID, envelope.RowVersion, dependent.viewSchemaID, "invalidate", changedFieldKeys)
		if err != nil {
			return err
		}
		effect.MutationOrdinal = ordinal
		ordinal++
		if err := s.recordPublications.AppendRecordChangedTx(ctx, tx, effect); err != nil {
			return err
		}
	}
	return nil
}
