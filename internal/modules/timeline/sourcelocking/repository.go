package sourcelocking

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
)

type EnvelopeReader interface {
	LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID, bool) (sourcerepository.Envelope, error)
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
	envelope, err := r.envelopes.LoadEnvelopeTx(ctx, tx, recordID, true)
	if err != nil {
		if errors.Is(err, sourcerepository.ErrEnvelopeNotFound) {
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
