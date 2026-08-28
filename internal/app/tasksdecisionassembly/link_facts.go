package tasksdecisionassembly

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
)

type linkFactsAdapter struct {
	reader linkRecordFactReader
}

type linkRecordFactReader interface {
	LoadRecordTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (links.ActiveFacts, error)
}

func NewLinkFactsCapability() tasksdecisions.LinkFactsCapability {
	return newLinkFactsCapability(links.FactReader{})
}

func newLinkFactsCapability(reader linkRecordFactReader) tasksdecisions.LinkFactsCapability {
	return linkFactsAdapter{reader: reader}
}

func (adapter linkFactsAdapter) LoadRecordLinkFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) ([]tasksdecisions.LinkFact, error) {
	facts, err := adapter.reader.LoadRecordTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return nil, err
	}
	result := make([]tasksdecisions.LinkFact, 0, len(facts.RecordLinks))
	for _, fact := range facts.RecordLinks {
		converted := tasksdecisions.LinkFact{
			SourceRecordID:      fact.SrcRecordID,
			DestinationRecordID: fact.DstRecordID,
			LinkType:            fact.LinkType.String(),
		}
		if fact.FieldKey != nil {
			converted.FieldKey = *fact.FieldKey
			converted.HasFieldKey = true
		}
		result = append(result, converted)
	}
	return result, nil
}
