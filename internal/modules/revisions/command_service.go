package revisions

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var ErrInvalidCommandServiceDependency = errors.New("revisions: invalid command service dependency")

type CommandServiceDependencies struct {
	Database                    postgres.DB
	ImportedAttributionResolver ImportedAttributionResolver
	ProjectionRebuilder         ProjectionRebuilder
	DeleteRestoreProviders      *DeleteRestoreProviderCatalog
	RowRollbackProviders        *RowProviderCatalog
	NonRowRollbackProviders     *NonRowProviderCatalog
}

type CommandService struct {
	commands *commandStore
	history  *historyStore
}

type historyStore struct {
	store *commandStore
}

func NewCommandService(dependencies CommandServiceDependencies) (*CommandService, error) {
	checks := []struct {
		name  string
		value any
	}{
		{name: "database", value: dependencies.Database},
		{name: "imported attribution resolver", value: dependencies.ImportedAttributionResolver},
		{name: "projection rebuilder", value: dependencies.ProjectionRebuilder},
		{name: "delete/restore provider catalog", value: dependencies.DeleteRestoreProviders},
		{name: "row rollback provider catalog", value: dependencies.RowRollbackProviders},
		{name: "non-row rollback provider catalog", value: dependencies.NonRowRollbackProviders},
	}
	for _, check := range checks {
		if nilDependency(check.value) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidCommandServiceDependency, check.name)
		}
	}
	store := &commandStore{
		db:                          dependencies.Database,
		appender:                    NewAppender(),
		incidentAccess:              incidents.NewAccess(dependencies.Database),
		importedAttributionResolver: dependencies.ImportedAttributionResolver,
		projectionRebuilder:         dependencies.ProjectionRebuilder,
		deleteRestoreProviders:      dependencies.DeleteRestoreProviders,
		rowRollbackProviders:        dependencies.RowRollbackProviders,
		nonRowRollbackProviders:     dependencies.NonRowRollbackProviders,
	}
	return &CommandService{commands: store, history: &historyStore{store: store}}, nil
}

func (s *CommandService) GetHistoryRecord(ctx context.Context, recordID uuid.UUID) (RecordHistoryRecord, error) {
	return s.history.GetHistoryRecord(ctx, recordID)
}

func (s *CommandService) ListRecordHistory(ctx context.Context, record RecordHistoryRecord) ([]map[string]any, error) {
	return s.history.ListRecordHistory(ctx, record)
}

func (s *CommandService) RollbackRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request RollbackRequest, requestHash []byte, requestID string, now time.Time) (RollbackResult, error) {
	return s.commands.RollbackRecord(ctx, actor, recordID, request, requestHash, requestID, now)
}

func (s *CommandService) SoftDeleteRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request DeleteRestoreRequest, requestHash []byte, requestID string, now time.Time) (DeleteRestoreResult, error) {
	return s.commands.SoftDeleteRecord(ctx, actor, recordID, request, requestHash, requestID, now)
}

func (s *CommandService) RestoreRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request DeleteRestoreRequest, requestHash []byte, requestID string, now time.Time) (DeleteRestoreResult, error) {
	return s.commands.RestoreRecord(ctx, actor, recordID, request, requestHash, requestID, now)
}

func (s *historyStore) GetHistoryRecord(ctx context.Context, recordID uuid.UUID) (RecordHistoryRecord, error) {
	return s.store.GetHistoryRecord(ctx, recordID)
}

func (s *historyStore) ListRecordHistory(ctx context.Context, record RecordHistoryRecord) ([]map[string]any, error) {
	return s.store.ListRecordHistory(ctx, record)
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
