package tasksdecisionassembly

import (
	"context"
	linkfacts "github.com/JochiRaider/cartulary/internal/modules/links/facts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type retainedSupersessionReader interface {
	LoadRetainedSupersessionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) ([]linkfacts.SupersessionFact, error)
}

type supersessionAdapter struct{ reader retainedSupersessionReader }

func NewSupersessionReader() tasksdecisions.SupersessionReader {
	return supersessionAdapter{reader: linkfacts.FactReader{}}
}

func (a supersessionAdapter) LoadRetainedSupersessionTx(ctx context.Context, tx pgx.Tx, incidentID, recordID uuid.UUID) ([]tasksdecisions.SupersessionFact, error) {
	facts, err := a.reader.LoadRetainedSupersessionTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return nil, err
	}
	result := make([]tasksdecisions.SupersessionFact, 0, len(facts))
	for _, fact := range facts {
		result = append(result, tasksdecisions.SupersessionFact{SourceRecordID: fact.SourceRecordID, DestinationRecordID: fact.DestinationRecordID})
	}
	return result, nil
}
