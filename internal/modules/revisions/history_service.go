package revisions

import (
	"context"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"sort"
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
	rollbackActions historyRollbackActionEvaluator
}

func newHistoryStore(transactions TransactionRunner, envelopes RecordEnvelopePort, attribution ImportedAttributionResolver, commands *commandStore) *historyStore {
	return &historyStore{
		transactions:    transactions,
		envelopes:       envelopes,
		repository:      historyQueryRepository{},
		materializer:    historyRowMaterializer{catalog: commands.targetSemantics},
		attribution:     importedHistoryAttributionDecorator{resolver: attribution},
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

func (s *historyStore) ListRecordHistory(ctx context.Context, record RecordHistoryRecord, query HistoryQuery) (HistoryResult, error) {
	if err := query.validate(); err != nil {
		return HistoryResult{}, err
	}
	if s.transactions == nil {
		return HistoryResult{}, errors.New("revisions history store: postgres dependency is nil")
	}
	// Selected change-set facts are immutable. Eligibility uses current source state;
	// entry-reference allocation remains transactional and stable across readers.
	tx, err := s.transactions.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return HistoryResult{}, fmt.Errorf("begin record history transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	descriptors, err := s.repository.SelectDescriptorsTx(ctx, tx, record, query)
	if err != nil {
		return HistoryResult{}, err
	}
	result := HistoryResult{Record: record, Items: make([]RecordHistoryItem, 0, query.Limit)}
	if len(descriptors) > query.Limit {
		descriptors = descriptors[:query.Limit]
		next := descriptors[len(descriptors)-1]
		result.Next = &next
	}
	rowKinds := make([]string, 0)
	for kind, entry := range s.materializer.catalog.byTargetKind {
		if entry.dispatchClass == rollbackcontract.DispatchRow {
			rowKinds = append(rowKinds, kind)
		}
	}
	sort.Strings(rowKinds)
	mutations, err := s.repository.LoadMutationRowsTx(ctx, tx, record, descriptors, rowKinds)
	if err != nil {
		return HistoryResult{}, err
	}
	type mutationKey struct {
		changeSet uuid.UUID
		sequence  int
	}
	mutationRows := make(map[mutationKey]mutationHistoryRow, len(mutations))
	revisionPositions := make([]HistoryPosition, 0, len(descriptors))
	for _, row := range mutations {
		if row.RevisionCount > 1 {
			return HistoryResult{}, errors.New("multiple record revisions for one mutation event")
		}
		key := mutationKey{row.ChangeSetID, row.SequenceNo}
		if _, duplicate := mutationRows[key]; duplicate {
			return HistoryResult{}, errors.New("duplicate history mutation facts")
		}
		mutationRows[key] = row
		if row.CoalesceRevision && row.RevisionNo != nil {
			revisionPositions = append(revisionPositions, HistoryPosition{ChangeSetID: row.ChangeSetID, RevisionNo: *row.RevisionNo})
		}
	}
	for _, descriptor := range descriptors {
		if descriptor.Kind == HistoryRevision {
			revisionPositions = append(revisionPositions, descriptor)
		}
	}
	revisions, err := s.repository.LoadRevisionRowsTx(ctx, tx, record, revisionPositions)
	if err != nil {
		return HistoryResult{}, err
	}
	type revisionKey struct {
		changeSet uuid.UUID
		revision  int64
	}
	revisionRows := make(map[revisionKey]revisionHistoryRow, len(revisions))
	for _, row := range revisions {
		revisionRows[revisionKey{row.ChangeSetID, row.RevisionNo}] = row
	}
	for _, descriptor := range descriptors {
		var item RecordHistoryItem
		if descriptor.Kind == HistoryMutation {
			row, ok := mutationRows[mutationKey{descriptor.ChangeSetID, descriptor.SequenceNo}]
			if !ok {
				return HistoryResult{}, errors.New("selected history mutation is missing")
			}
			item, err = s.materializer.Mutation(record, row)
			if err != nil {
				return HistoryResult{}, fmt.Errorf("project record history mutation: %w", err)
			}
			if row.CoalesceRevision && row.RevisionNo != nil {
				revision, ok := revisionRows[revisionKey{row.ChangeSetID, *row.RevisionNo}]
				if !ok {
					return HistoryResult{}, errors.New("selected coalesced revision is missing")
				}
				if err := s.materializer.CoalesceRevision(record, &item, revision); err != nil {
					return HistoryResult{}, err
				}
			}
			if item.hasTargetEntry && item.HistoryEntryRef == nil {
				ref, err := s.repository.EnsureHistoryEntryRefTx(ctx, tx, record.RecordID, item.ChangeSetID, item.sequenceNo)
				if err != nil {
					return HistoryResult{}, err
				}
				item.HistoryEntryRef = &ref
			}
		} else {
			row, ok := revisionRows[revisionKey{descriptor.ChangeSetID, descriptor.RevisionNo}]
			if !ok {
				return HistoryResult{}, errors.New("selected history revision is missing")
			}
			item, err = s.materializer.Revision(record, row)
			if err != nil {
				return HistoryResult{}, fmt.Errorf("project record history revision: %w", err)
			}
		}
		result.Items = append(result.Items, item)
	}
	if err := s.attribution.DecorateTx(ctx, tx, record.IncidentID, result.Items); err != nil {
		return HistoryResult{}, err
	}
	if err := s.rollbackActions.DecorateTx(ctx, tx, record, result.Items); err != nil {
		return HistoryResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return HistoryResult{}, fmt.Errorf("commit record history transaction: %w", err)
	}
	return result, nil
}
