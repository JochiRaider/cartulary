package revisions

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"
)

var ErrInvalidCommandServiceDependency = errors.New("revisions: invalid command service dependency")

type CommandServiceDependencies struct {
	Transactions                TransactionRunner
	Authorization               CommandAuthorizer
	Idempotency                 IdempotencyPort
	ImportedAttributionResolver ImportedAttributionResolver
	Projections                 ProjectionServices
	ProviderContributions       []ProviderContribution
	Appender                    *Appender
	RecordEnvelopes             RecordEnvelopePort
	Clock                       func() time.Time
}

type CommandService struct {
	commands *commandStore
	history  *historyStore
	now      func() time.Time
}

func NewCommandService(dependencies CommandServiceDependencies) (*CommandService, error) {
	checks := []struct {
		name  string
		value any
	}{
		{name: "transactions", value: dependencies.Transactions},
		{name: "authorization", value: dependencies.Authorization},
		{name: "idempotency", value: dependencies.Idempotency},
		{name: "imported attribution resolver", value: dependencies.ImportedAttributionResolver},
		{name: "projection services", value: dependencies.Projections},
		{name: "appender", value: dependencies.Appender},
		{name: "record envelopes", value: dependencies.RecordEnvelopes},
		{name: "clock", value: dependencies.Clock},
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
		transactions:            dependencies.Transactions,
		appender:                dependencies.Appender,
		envelopes:               dependencies.RecordEnvelopes,
		authorization:           dependencies.Authorization,
		idempotency:             dependencies.Idempotency,
		projections:             dependencies.Projections,
		deleteRestoreSources:    deleteRestoreSources,
		rowRollbackProviders:    rowRollbackProviders,
		nonRowRollbackProviders: nonRowRollbackProviders,
	}
	return &CommandService{
		commands: store,
		history:  newHistoryStore(dependencies.Transactions, dependencies.RecordEnvelopes, dependencies.ImportedAttributionResolver, store),
		now:      dependencies.Clock,
	}, nil
}

func (s *CommandService) GetHistory(ctx context.Context, query HistoryQuery) (HistoryResult, error) {
	record, err := s.history.GetHistoryRecord(ctx, query.RecordID)
	if err != nil {
		return HistoryResult{}, err
	}
	resources, err := s.history.ListRecordHistory(ctx, record)
	if err != nil {
		return HistoryResult{}, err
	}
	return HistoryResult{Record: record, Resources: resources}, nil
}

func (s *CommandService) RollbackRecord(ctx context.Context, command RollbackCommand) (RollbackResult, error) {
	command.effectiveAt = s.now().UTC()
	return s.commands.RollbackRecord(ctx, command)
}

func (s *CommandService) SoftDeleteRecord(ctx context.Context, command DeleteRestoreCommand) (DeleteRestoreResult, error) {
	command.effectiveAt = s.now().UTC()
	return s.commands.SoftDeleteRecord(ctx, command)
}

func (s *CommandService) RestoreRecord(ctx context.Context, command DeleteRestoreCommand) (DeleteRestoreResult, error) {
	command.effectiveAt = s.now().UTC()
	return s.commands.RestoreRecord(ctx, command)
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
