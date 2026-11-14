package tasksdecisionassembly

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	linkfacts "github.com/JochiRaider/cartulary/internal/modules/links/facts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
)

type linkFieldReferenceReader interface {
	LoadFieldReferencesTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, string) ([]linkfacts.FieldReferenceFact, error)
}

type fieldReferenceAdapter struct{ reader linkFieldReferenceReader }

func NewFieldReferenceReader() tasksdecisions.FieldReferenceReader {
	return fieldReferenceAdapter{reader: linkfacts.FactReader{}}
}

func (a fieldReferenceAdapter) LoadFieldReferencesTx(ctx context.Context, tx pgx.Tx, incidentID, sourceID uuid.UUID, fieldKey string) ([]tasksdecisions.FieldReferenceFact, error) {
	facts, err := a.reader.LoadFieldReferencesTx(ctx, tx, incidentID, sourceID, fieldKey)
	if err != nil {
		return nil, err
	}
	result := make([]tasksdecisions.FieldReferenceFact, 0, len(facts))
	for _, fact := range facts {
		result = append(result, tasksdecisions.FieldReferenceFact{DestinationRecordID: fact.DestinationRecordID, LinkType: fact.LinkType.String()})
	}
	return result, nil
}
