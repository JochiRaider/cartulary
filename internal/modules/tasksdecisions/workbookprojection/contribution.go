package workbookprojection

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	taskRequestsViewSchemaID = "cartulary.view.task_requests.v1"
	decisionsViewSchemaID    = "cartulary.view.decisions.v1"
)

type TaskRequestProjectionInput struct {
	RecordID           uuid.UUID
	IncidentID         uuid.UUID
	RowVersion         int64
	Title              *string
	Status             string
	OwnerUserID        *uuid.UUID
	Priority           *string
	TaskKind           *string
	Workstream         *string
	DueAt              *time.Time
	RequesterPartyText *string
	RequesterPartyID   *uuid.UUID
	BlockedReason      *string
	CompletedAt        *time.Time
	ExternalTicketRef  *string
	ClosureSummary     *string
	DecisionRecordID   *uuid.UUID
	LinkedRecordCount  int
	UpdatedAt          time.Time
	NoOwner            bool
}

type TaskRequestProjectionInputPage struct {
	Inputs       []TaskRequestProjectionInput
	NextRecordID *uuid.UUID
}

type TaskRequestSourceReader interface {
	LoadTaskRequestProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (TaskRequestProjectionInput, bool, error)
	ListTaskRequestProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (TaskRequestProjectionInputPage, error)
}

type DecisionProjectionInput struct {
	RecordID            uuid.UUID
	IncidentID          uuid.UUID
	RowVersion          int64
	Summary             *string
	Status              string
	OwnerUserID         *uuid.UUID
	DecisionType        *string
	DecidedAt           *time.Time
	Rationale           *string
	AffectedRecordCount int
	SupersedesRecordID  *uuid.UUID
	UpdatedAt           time.Time
	IsSuperseded        bool
}

type DecisionProjectionInputPage struct {
	Inputs       []DecisionProjectionInput
	NextRecordID *uuid.UUID
}

type DecisionSourceReader interface {
	LoadDecisionProjectionInputTx(context.Context, pgx.Tx, uuid.UUID) (DecisionProjectionInput, bool, error)
	ListDecisionProjectionInputsTx(context.Context, pgx.Tx, uuid.UUID, *uuid.UUID, int) (DecisionProjectionInputPage, error)
}

type TaskDerivedFact struct {
	RecordID uuid.UUID
	Value    map[string]any
}

type TaskReader interface {
	CollectTaskDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]TaskDerivedFact, error)
}

type DecisionDerivedFact struct {
	RecordID uuid.UUID
	Value    map[string]any
}

type Reader interface {
	TaskReader
	CollectDecisionDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]DecisionDerivedFact, error)
}

type Rows interface {
	RefreshTaskRequestTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadTaskRequestTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	RebuildTaskRequestsTx(context.Context, pgx.Tx, uuid.UUID) error
	RefreshDecisionTx(context.Context, pgx.Tx, uuid.UUID) error
	LoadDecisionTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	RebuildDecisionsTx(context.Context, pgx.Tx, uuid.UUID) error
}

type Rebuilder interface {
	RebuildTaskRequests(context.Context, uuid.UUID) error
	RebuildDecisions(context.Context, uuid.UUID) error
}

type Ports struct {
	Rows      Rows
	Rebuilder Rebuilder
	Reader    Reader
}

type Contribution struct {
	contract          providercontract.Contribution
	taskRequestSource TaskRequestSourceReader
	decisionSource    DecisionSourceReader
}

func NewContribution(
	taskRequestSource TaskRequestSourceReader,
	decisionSource DecisionSourceReader,
) (Contribution, error) {
	if taskRequestSource == nil {
		return Contribution{}, fmt.Errorf("task-request projection source is required")
	}
	if decisionSource == nil {
		return Contribution{}, fmt.Errorf("decision projection source is required")
	}
	intents, err := SurfaceIntents()
	if err != nil {
		return Contribution{}, err
	}
	contract, err := providercontract.NewContribution("tasksdecisions", Descriptors(), intents)
	if err != nil {
		return Contribution{}, err
	}
	return Contribution{
		contract:          contract,
		taskRequestSource: taskRequestSource,
		decisionSource:    decisionSource,
	}, nil
}

func (contribution Contribution) ProjectionContribution() providercontract.Contribution {
	return contribution.contract
}

func (contribution Contribution) TaskRequestSource() TaskRequestSourceReader {
	return contribution.taskRequestSource
}

func (contribution Contribution) DecisionSource() DecisionSourceReader {
	return contribution.decisionSource
}

func Descriptors() []providercontract.ProviderDescriptor {
	return []providercontract.ProviderDescriptor{
		{
			SchemaVersion:                providercontract.DescriptorSchemaVersion,
			Status:                       providercontract.ProviderStatusActive,
			ProviderID:                   "task_request",
			SourceOwnerModule:            "tasksdecisions",
			ViewSchemaIDs:                []string{taskRequestsViewSchemaID},
			SourceRecordTypes:            []string{"task_request"},
			SourceAuthorityModules:       []string{"links", "records", "tasksdecisions"},
			ProjectionTableIDs:           []string{"task_request_grid_projection"},
			ProjectionStorageOwnerModule: "projections",
			Capabilities: providercontract.ProviderCapabilities{
				Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true,
			},
			RestoreRebuild: providercontract.RestoreRebuildRequired,
			FacadePackages: []string{"internal/modules/tasksdecisions/workbookprojection"},
			RebuildAfter:   []string{"party"},
			CharacterizationRefs: []string{
				"internal/modules/tasksdecisions/task_decisions_store_test.go",
				"internal/modules/projections/internal/runtime/query_test.go",
			},
		},
		{
			SchemaVersion:                providercontract.DescriptorSchemaVersion,
			Status:                       providercontract.ProviderStatusActive,
			ProviderID:                   "decision",
			SourceOwnerModule:            "tasksdecisions",
			ViewSchemaIDs:                []string{decisionsViewSchemaID},
			SourceRecordTypes:            []string{"decision"},
			SourceAuthorityModules:       []string{"links", "records", "tasksdecisions"},
			ProjectionTableIDs:           []string{"decision_grid_projection"},
			ProjectionStorageOwnerModule: "projections",
			Capabilities: providercontract.ProviderCapabilities{
				Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true,
			},
			RestoreRebuild: providercontract.RestoreRebuildRequired,
			FacadePackages: []string{"internal/modules/tasksdecisions/workbookprojection"},
			RebuildAfter:   []string{"task_request"},
			CharacterizationRefs: []string{
				"internal/modules/tasksdecisions/task_decisions_store_test.go",
				"internal/modules/projections/internal/runtime/query_test.go",
			},
		},
	}
}

func SurfaceIntents() ([]providercontract.SurfaceIntent, error) {
	intents := make([]providercontract.SurfaceIntent, 0, 2)
	for _, viewSchemaID := range []string{taskRequestsViewSchemaID, decisionsViewSchemaID} {
		intent, err := surfaceIntent(viewSchemaID)
		if err != nil {
			return nil, err
		}
		intents = append(intents, intent)
	}
	return intents, nil
}

func surfaceIntent(viewSchemaID string) (providercontract.SurfaceIntent, error) {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		return providercontract.SurfaceIntent{}, fmt.Errorf("Tasks/Decisions projection surface %q has no view-schema contract", viewSchemaID)
	}
	fields := schema.Fields()
	fieldKeys := make([]string, 0, len(fields))
	for fieldKey := range fields {
		fieldKeys = append(fieldKeys, fieldKey)
	}
	slices.Sort(fieldKeys)
	return providercontract.SurfaceIntent{ViewSchemaID: viewSchemaID, FieldKeys: fieldKeys}, nil
}
