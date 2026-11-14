package tasksdecisionassembly

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type envelopeReader interface {
	LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID) (enveloperead.Envelope, error)
}

type envelopeAdapter struct{ reader envelopeReader }

func NewEnvelopeFactReader() tasksdecisions.EnvelopeFactReader {
	return envelopeAdapter{reader: enveloperead.Reader{}}
}

func (a envelopeAdapter) LoadEnvelopeFactTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (tasksdecisions.EnvelopeFact, bool, error) {
	envelope, err := a.reader.LoadEnvelopeTx(ctx, tx, recordID)
	if errors.Is(err, enveloperead.ErrEnvelopeNotFound) {
		return tasksdecisions.EnvelopeFact{}, false, nil
	}
	if err != nil {
		return tasksdecisions.EnvelopeFact{}, false, err
	}
	return tasksdecisions.EnvelopeFact{IncidentID: envelope.IncidentID, RecordType: envelope.RecordType, Deleted: envelope.DeletedAt != nil}, true, nil
}
