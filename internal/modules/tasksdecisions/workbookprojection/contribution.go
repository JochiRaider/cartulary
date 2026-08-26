package workbookprojection

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
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
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return Contribution{}, fmt.Errorf("compose Tasks/Decisions projection catalog: %w", err)
	}
	intents, err := surfaceIntents(catalog)
	if err != nil {
		return Contribution{}, err
	}
	contract, err := providercontract.NewContribution("tasksdecisions", descriptors(catalog), intents)
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
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return nil
	}
	return descriptors(catalog)
}

func descriptors(catalog *sourcecatalog.Catalog) []providercontract.ProviderDescriptor {
	result := make([]providercontract.ProviderDescriptor, 0, 2)
	for _, recordType := range []string{"task_request", "decision"} {
		surface, ok := catalog.SurfaceByRecordType(recordType)
		if !ok {
			return nil
		}
		descriptor := providercontract.ProviderDescriptor{
			SchemaVersion:                providercontract.DescriptorSchemaVersion,
			Status:                       providercontract.ProviderStatusActive,
			ProviderID:                   surface.RecordType,
			SourceOwnerModule:            "tasksdecisions",
			ViewSchemaIDs:                []string{surface.ViewSchemaID},
			SourceRecordTypes:            []string{surface.RecordType},
			SourceAuthorityModules:       []string{"links", "records", "tasksdecisions"},
			ProjectionTableIDs:           []string{surface.BaseProjection},
			ProjectionStorageOwnerModule: "projections",
			Capabilities: providercontract.ProviderCapabilities{
				Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true,
			},
			RestoreRebuild:       providercontract.RestoreRebuildRequired,
			FacadePackages:       []string{"internal/modules/tasksdecisions/workbookprojection"},
			RebuildAfter:         projectionRebuildAfter(surface.RecordType),
			CharacterizationRefs: projectionCharacterizationRefs(surface.RecordType),
		}
		result = append(result, descriptor)
	}
	return result
}

func projectionCharacterizationRefs(recordType string) []string {
	switch recordType {
	case "task_request":
		return []string{
			"internal/modules/tasksdecisions/task_mutation_store_test.go",
			"internal/modules/projections/internal/runtime/query_test.go",
		}
	case "decision":
		return []string{
			"internal/modules/tasksdecisions/decision_mutation_store_test.go",
			"internal/modules/projections/internal/runtime/query_test.go",
		}
	default:
		return nil
	}
}

func projectionRebuildAfter(recordType string) []string {
	switch recordType {
	case "task_request":
		return []string{"party"}
	case "decision":
		return []string{"task_request"}
	default:
		return nil
	}
}

func SurfaceIntents() ([]providercontract.SurfaceIntent, error) {
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return nil, fmt.Errorf("compose Tasks/Decisions projection catalog: %w", err)
	}
	return surfaceIntents(catalog)
}

func surfaceIntents(catalog *sourcecatalog.Catalog) ([]providercontract.SurfaceIntent, error) {
	intents := make([]providercontract.SurfaceIntent, 0, 2)
	for _, recordType := range []string{"task_request", "decision"} {
		surface, ok := catalog.SurfaceByRecordType(recordType)
		if !ok {
			return nil, fmt.Errorf("Tasks/Decisions projection record type %q has no source catalog entry", recordType)
		}
		intent, err := surfaceIntent(surface.ViewSchemaID)
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
