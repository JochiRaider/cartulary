package revisions

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type CommandService struct {
	store *Store
}

func NewCommandService(store *Store) *CommandService {
	return &CommandService{store: store}
}

func (s *CommandService) GetHistoryRecord(ctx context.Context, recordID uuid.UUID) (RecordHistoryRecord, error) {
	return s.store.GetHistoryRecord(ctx, recordID)
}

func (s *CommandService) ListRecordHistory(ctx context.Context, record RecordHistoryRecord) ([]map[string]any, error) {
	return s.store.ListRecordHistory(ctx, record)
}

func (s *CommandService) RollbackRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request RollbackRequest, requestHash []byte, requestID string, now time.Time) (RollbackResult, error) {
	return s.store.RollbackRecord(ctx, actor, recordID, request, requestHash, requestID, now)
}

func (s *CommandService) SoftDeleteRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request DeleteRestoreRequest, requestHash []byte, requestID string, now time.Time) (DeleteRestoreResult, error) {
	return s.store.SoftDeleteRecord(ctx, actor, recordID, request, requestHash, requestID, now)
}

func (s *CommandService) RestoreRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request DeleteRestoreRequest, requestHash []byte, requestID string, now time.Time) (DeleteRestoreResult, error) {
	return s.store.RestoreRecord(ctx, actor, recordID, request, requestHash, requestID, now)
}
