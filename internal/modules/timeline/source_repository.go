package timeline

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
)

func (s *store) loadSourceRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (sourceRecord, error) {
	record, err := s.sourceRepository.LoadTx(ctx, tx, recordID)
	if errors.Is(err, sourcerepository.ErrNotFound) {
		return sourceRecord{}, ErrRecordNotFound
	}
	return record, err
}

func (s *store) loadSourceRecordForIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (sourceRecord, error) {
	record, err := s.sourceRepository.LoadForIncidentTx(ctx, tx, incidentID, recordID)
	if errors.Is(err, sourcerepository.ErrNotFound) {
		return sourceRecord{}, ErrRecordNotFound
	}
	return record, err
}
