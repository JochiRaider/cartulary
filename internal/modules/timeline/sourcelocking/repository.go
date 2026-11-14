package sourcelocking

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
)

type EnvelopeReader interface {
	LoadEnvelopesForNoKeyUpdateTx(context.Context, pgx.Tx, []uuid.UUID) (map[uuid.UUID]sourcerepository.Envelope, error)
}
type Repository struct{ envelopes EnvelopeReader }

func New(envelopes EnvelopeReader) *Repository { return &Repository{envelopes: envelopes} }
func (r *Repository) LoadTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (sourcerepository.Snapshot, error) {
	return r.loadTx(ctx, tx, nil, recordID)
}

func (r *Repository) LoadForIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (sourcerepository.Snapshot, error) {
	return r.loadTx(ctx, tx, &incidentID, recordID)
}

func (r *Repository) loadTx(ctx context.Context, tx pgx.Tx, incidentID *uuid.UUID, recordID uuid.UUID) (sourcerepository.Snapshot, error) {
	// Retain non-key protection when a coordinator already owns the complete
	// reference set; an envelope lock upgrade would conflict with foreign keys.
	envelopes, err := r.envelopes.LoadEnvelopesForNoKeyUpdateTx(ctx, tx, []uuid.UUID{recordID})
	if err != nil {
		return sourcerepository.Snapshot{}, err
	}
	envelope, ok := envelopes[recordID]
	if !ok {
		return sourcerepository.Snapshot{}, sourcerepository.ErrNotFound
	}
	query := sourcerepository.SelectSQL + "\n WHERE record_id = $1"
	args := []any{recordID}
	if incidentID != nil {
		query += "\n   AND incident_id = $2"
		args = append(args, *incidentID)
	}
	query += "\n FOR UPDATE"
	snapshot, err := sourcerepository.ScanSource(tx.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sourcerepository.Snapshot{}, sourcerepository.ErrNotFound
		}
		return sourcerepository.Snapshot{}, err
	}
	if !sourcerepository.EligibleEnvelope(snapshot, envelope, incidentID) {
		return sourcerepository.Snapshot{}, sourcerepository.ErrNotFound
	}
	sourcerepository.ApplyEnvelope(&snapshot, envelope)
	return snapshot, nil
}
