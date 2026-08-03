package revisions

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrRecordNotFound = errors.New("revisions: record not found")

// historyStore coordinates history-specific ports and services. SQL is isolated
// in historyQueryRepository; row interpretation is isolated in
// historyRowMaterializer; transport pagination remains outside this service.
type historyStore struct {
	transactions    TransactionRunner
	envelopes       RecordEnvelopePort
	repository      historyQueryRepository
	materializer    historyRowMaterializer
	attribution     importedHistoryAttributionDecorator
	pages           historyPageAssembler
	rollbackActions historyRollbackActionEvaluator
}

func newHistoryStore(transactions TransactionRunner, envelopes RecordEnvelopePort, attribution ImportedAttributionResolver, commands *commandStore) *historyStore {
	return &historyStore{
		transactions:    transactions,
		envelopes:       envelopes,
		repository:      historyQueryRepository{},
		materializer:    historyRowMaterializer{},
		attribution:     importedHistoryAttributionDecorator{resolver: attribution},
		pages:           historyPageAssembler{},
		rollbackActions: newHistoryRollbackActionEvaluator(commands),
	}
}

func (s *historyStore) GetHistoryRecord(ctx context.Context, recordID uuid.UUID) (RecordHistoryRecord, error) {
	if s.envelopes == nil {
		return RecordHistoryRecord{}, errors.New("revisions history store: envelope dependency is nil")
	}
	envelope, err := s.envelopes.LoadEnvelope(ctx, recordID)
	if err != nil {
		if errors.Is(err, ErrEnvelopeNotFound) {
			return RecordHistoryRecord{}, ErrRecordNotFound
		}
		return RecordHistoryRecord{}, err
	}
	return RecordHistoryRecord{
		IncidentID:  envelope.IncidentID,
		RecordID:    envelope.RecordID,
		RecordType:  envelope.RecordType,
		RowVersion:  envelope.RowVersion,
		Deleted:     envelope.DeletedAt != nil,
		DeletedAt:   envelope.DeletedAt,
		DeletedByID: envelope.DeletedByUserID,
	}, nil
}

func (s *historyStore) ListRecordHistory(ctx context.Context, record RecordHistoryRecord) ([]map[string]any, error) {
	if s.transactions == nil {
		return nil, errors.New("revisions history store: postgres dependency is nil")
	}
	tx, err := s.transactions.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin record history transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	mutationRows, err := s.repository.LoadMutationRowsTx(ctx, tx, record)
	if err != nil {
		return nil, err
	}
	items := make([]RecordHistoryItem, 0, len(mutationRows))
	for _, row := range mutationRows {
		item := s.materializer.Mutation(record, row)
		if item.hasTargetEntry && item.HistoryEntryRef == nil {
			generated, err := s.repository.EnsureHistoryEntryRefTx(ctx, tx, record.RecordID, item.ChangeSetID, item.sequenceNo)
			if err != nil {
				return nil, err
			}
			item.HistoryEntryRef = &generated
		}
		items = append(items, item)
	}

	revisionRows, err := s.repository.LoadRevisionRowsTx(ctx, tx, record)
	if err != nil {
		return nil, err
	}
	items = append(items, s.materializer.Revisions(record, revisionRows, items)...)
	if err := s.attribution.DecorateTx(ctx, tx, record.IncidentID, items); err != nil {
		return nil, err
	}
	if err := s.rollbackActions.DecorateTx(ctx, tx, record, items); err != nil {
		return nil, err
	}
	resources := s.pages.Resources(items)
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit record history transaction: %w", err)
	}
	return resources, nil
}
