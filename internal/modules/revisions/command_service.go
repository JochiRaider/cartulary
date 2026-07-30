package revisions

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var ErrInvalidCommandServiceDependency = errors.New("revisions: invalid command service dependency")

type CommandServiceDependencies struct {
	Database                    postgres.DB
	ImportedAttributionResolver ImportedAttributionResolver
	Projections                 ProjectionServices
	ProviderContributions       []ProviderContribution
	Appender                    *Appender
	EnvelopeStore               *records.Store
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
		{name: "projection services", value: dependencies.Projections},
		{name: "appender", value: dependencies.Appender},
		{name: "envelope store", value: dependencies.EnvelopeStore},
	}
	for _, check := range checks {
		if nilDependency(check.value) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidCommandServiceDependency, check.name)
		}
	}
	deleteRestoreSources, rowRollbackProviders, nonRowRollbackProviders, err := buildProviderCatalogs(dependencies.ProviderContributions)
	if err != nil {
		return nil, fmt.Errorf("%w: provider contributions: %w", ErrInvalidCommandServiceDependency, err)
	}
	store := &commandStore{
		db:                          dependencies.Database,
		appender:                    dependencies.Appender,
		envelopes:                   dependencies.EnvelopeStore,
		incidentAccess:              incidents.NewAccess(dependencies.Database),
		importedAttributionResolver: dependencies.ImportedAttributionResolver,
		projections:                 dependencies.Projections,
		deleteRestoreSources:        deleteRestoreSources,
		rowRollbackProviders:        rowRollbackProviders,
		nonRowRollbackProviders:     nonRowRollbackProviders,
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
