package adapters

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	projectionruntime "github.com/JochiRaider/cartulary/internal/modules/projections/internal/runtime"
	projectionstorage "github.com/JochiRaider/cartulary/internal/modules/projections/internal/storage"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	workbookrestoreprobe "github.com/JochiRaider/cartulary/internal/modules/workbook/restoreprobe"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

type Dependencies struct {
	Postgres       postgres.DB
	Timeline       timelineprojection.Contribution
	Entities       entityprojection.Contribution
	Indicators     indicatorprojection.Contribution
	Assessments    assessmentprojection.Contribution
	Artifacts      artifactprojection.Contribution
	Evidence       evidenceprojection.Contribution
	Parties        partyprojection.Contribution
	TasksDecisions taskdecisionprojection.Contribution
}

// Ports is the fail-closed projection composition result. It exposes only
// immutable descriptor facts, consumer-owned interfaces, and typed owner
// facades; concrete runtime, storage, and query implementations remain private.
type Ports struct {
	descriptors    providercontract.DescriptorSet
	recoveryState  recoverystate.Contribution
	declarative    *projectionruntime.DeclarativeCatalog
	storage        *projectionstorage.Store
	semanticQuery  *queryengine.Engine
	operational    *projectionruntime.Catalog
	query          *projectionruntime.QueryService
	rebuild        *projectionruntime.RebuildService
	coordinator    *projectionruntime.Coordinator
	timeline       timelineprojection.Ports
	entities       entityprojection.Ports
	indicators     indicatorprojection.Ports
	assessments    assessmentprojection.Ports
	artifacts      artifactprojection.Ports
	evidence       evidenceprojection.Ports
	parties        partyprojection.Ports
	tasksDecisions taskdecisionprojection.Ports
}

func New(dependencies Dependencies) (Ports, error) {
	if dependencies.Postgres == nil {
		return Ports{}, fmt.Errorf("compose Projections: Postgres is required")
	}
	contributions := []struct {
		name     string
		contract providercontract.Contribution
	}{
		{name: "Timeline", contract: dependencies.Timeline.ProjectionContribution()},
		{name: "Entities", contract: dependencies.Entities.ProjectionContribution()},
		{name: "Indicators", contract: dependencies.Indicators.ProjectionContribution()},
		{name: "Assessments", contract: dependencies.Assessments.ProjectionContribution()},
		{name: "Artifacts", contract: dependencies.Artifacts.ProjectionContribution()},
		{name: "Evidence", contract: dependencies.Evidence.ProjectionContribution()},
		{name: "Parties", contract: dependencies.Parties.ProjectionContribution()},
		{name: "TasksDecisions", contract: dependencies.TasksDecisions.ProjectionContribution()},
	}
	contracts := make([]providercontract.Contribution, 0, len(contributions))
	for _, contribution := range contributions {
		if contribution.contract.IsZero() {
			return Ports{}, fmt.Errorf("compose Projections: %s contribution is required", contribution.name)
		}
		contracts = append(contracts, contribution.contract)
	}

	catalog, err := projectionruntime.NewDeclarativeCatalog(contracts)
	if err != nil {
		return Ports{}, fmt.Errorf("compose Projections catalog: %w", err)
	}
	store, err := projectionstorage.New(dependencies.Postgres)
	if err != nil {
		return Ports{}, fmt.Errorf("compose Projections storage: %w", err)
	}
	query, err := queryengine.New(catalog.SurfaceIntents())
	if err != nil {
		return Ports{}, fmt.Errorf("compose Projections query engine: %w", err)
	}
	recoveryState := providercontract.RecoveryStateContribution()
	if err := validateRecoveryState(catalog.DescriptorSet(), recoveryState); err != nil {
		return Ports{}, fmt.Errorf("compose Projections recovery state: %w", err)
	}
	operational, err := projectionruntime.NewOperationalCatalog(
		catalog.DescriptorSet(),
		projectionruntime.OperationalDependencies{
			Timeline:     dependencies.Timeline.Source(),
			Entities:     dependencies.Entities.Source(),
			Indicators:   dependencies.Indicators.Source(),
			Assessments:  dependencies.Assessments.Source(),
			Artifacts:    dependencies.Artifacts.Source(),
			Evidence:     dependencies.Evidence.Source(),
			Parties:      dependencies.Parties.Source(),
			TaskRequests: dependencies.TasksDecisions.TaskRequestSource(),
			Decisions:    dependencies.TasksDecisions.DecisionSource(),
		},
	)
	if err != nil {
		return Ports{}, fmt.Errorf("compose Projections operational catalog: %w", err)
	}
	queryService := projectionruntime.NewQueryService(dependencies.Postgres, operational)
	rebuildService := projectionruntime.NewRebuildService(dependencies.Postgres, operational)
	coordinator := projectionruntime.NewCoordinator(dependencies.Postgres, operational)
	entityRows := projectionruntime.NewEntityRows(dependencies.Postgres, dependencies.Entities.Source())
	artifactRows := projectionruntime.NewArtifactRows(dependencies.Postgres, dependencies.Artifacts.Source())
	taskDecisionRows := projectionruntime.NewTaskDecisionRows(
		dependencies.Postgres,
		dependencies.TasksDecisions.TaskRequestSource(),
		dependencies.TasksDecisions.DecisionSource(),
	)
	evidenceRows := &evidencePort{
		rows:    projectionruntime.NewEvidenceRows(dependencies.Postgres, dependencies.Evidence.Source()),
		rebuild: rebuildService,
	}
	ports := Ports{
		descriptors:   catalog.DescriptorSet(),
		recoveryState: recoveryState,
		declarative:   catalog,
		storage:       store,
		semanticQuery: query,
		operational:   operational,
		query:         queryService,
		rebuild:       rebuildService,
		coordinator:   coordinator,
		timeline: timelineprojection.Ports{
			Writer:    projectionruntime.NewTimelineRows(dependencies.Postgres),
			Rebuilder: rebuildService,
		},
		entities: entityprojection.Ports{
			Writer:    entityRows,
			Rebuilder: rebuildService,
			Reader:    entityRows,
		},
		indicators: indicatorprojection.Ports{
			Rows: projectionruntime.NewIndicatorRows(
				dependencies.Postgres,
				dependencies.Indicators.Source(),
			),
			Rebuilder: rebuildService,
		},
		assessments: assessmentprojection.Ports{
			Rows: projectionruntime.NewAssessmentRows(
				dependencies.Postgres,
				dependencies.Assessments.Source(),
			),
			Rebuilder: rebuildService,
		},
		artifacts: artifactprojection.Ports{
			Rows:      artifactRows,
			Rebuilder: rebuildService,
			Reader:    artifactRows,
		},
		evidence: evidenceprojection.Ports{
			Rows:      evidenceRows,
			Rebuilder: rebuildService,
		},
		parties: partyprojection.Ports{
			Rows:      projectionruntime.NewPartyRows(dependencies.Postgres, dependencies.Parties.Source()),
			Rebuilder: rebuildService,
		},
		tasksDecisions: taskdecisionprojection.Ports{
			Rows:      taskDecisionRows,
			Rebuilder: rebuildService,
			Reader:    taskDecisionRows,
		},
	}
	if !ports.Ready() {
		return Ports{}, fmt.Errorf("compose Projections: incomplete ports")
	}
	return ports, nil
}

func (ports Ports) DescriptorSet() providercontract.DescriptorSet {
	return ports.descriptors
}

func (ports Ports) RecoveryStateContribution() recoverystate.Contribution {
	return providercontract.RecoveryStateContribution()
}

func (ports Ports) WorkbookQueryProvider(viewSchemaID string) (workbook.QueryProvider, bool) {
	if ports.query == nil || !ports.query.Supports(viewSchemaID) {
		return nil, false
	}
	return workbook.QueryProviderFunc(func(ctx context.Context, command workbook.QueryCommand) (querypage.Result, error) {
		if command.ViewSchemaID != viewSchemaID {
			return querypage.Result{}, fmt.Errorf(
				"projection query contribution %q received view schema %q",
				viewSchemaID,
				command.ViewSchemaID,
			)
		}
		return ports.query.QueryRowsPage(
			ctx,
			command.IncidentID,
			viewSchemaID,
			command.Query,
			command.Window,
		)
	}), true
}

func (ports Ports) RecoveryPorts() restorecontract.ProjectionPorts {
	if ports.rebuild == nil {
		return restorecontract.ProjectionPorts{}
	}
	return restorecontract.ProjectionPorts{
		Rebuilder:         ports.rebuild.RestoreRebuilder(),
		StateContribution: ports.RecoveryStateContribution(),
	}
}

func (ports Ports) RestoreProbeQuery() workbookrestoreprobe.ProjectionQuery {
	return ports.query
}

func (ports Ports) RevisionServices() revisions.ProjectionServices {
	return ports.coordinator
}

type SourceTextRows interface {
	RefreshRowTx(context.Context, pgx.Tx, string, uuid.UUID) error
	LoadRowTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error)
}

type ImportRebuilder interface {
	RebuildImportedIncidentTx(context.Context, pgx.Tx, uuid.UUID) error
}

func (ports Ports) SourceTextRows() SourceTextRows {
	return ports.coordinator
}

func (ports Ports) ImportRebuilder() ImportRebuilder {
	return ports.rebuild
}

func (ports Ports) Timeline() timelineprojection.Ports { return ports.timeline }

func (ports Ports) Entities() entityprojection.Ports { return ports.entities }

func (ports Ports) Indicators() indicatorprojection.Ports { return ports.indicators }

func (ports Ports) Assessments() assessmentprojection.Ports { return ports.assessments }

func (ports Ports) Artifacts() artifactprojection.Ports { return ports.artifacts }

func (ports Ports) Evidence() evidenceprojection.Ports { return ports.evidence }

func (ports Ports) Parties() partyprojection.Ports { return ports.parties }

func (ports Ports) TasksDecisions() taskdecisionprojection.Ports { return ports.tasksDecisions }

func (ports Ports) Ready() bool {
	return ports.descriptors.Len() > 0 &&
		ports.recoveryState.OwnerID == "module.projections" &&
		ports.declarative != nil && ports.declarative.Ready() &&
		ports.storage != nil && ports.storage.Ready() &&
		ports.semanticQuery != nil && ports.semanticQuery.Ready() &&
		ports.operational != nil && ports.query != nil && ports.rebuild != nil && ports.coordinator != nil &&
		ports.timeline.Ready() && ports.entities.Ready() && ports.indicators.Ready() &&
		ports.assessments.Ready() && ports.artifacts.Ready() && ports.evidence.Ready() &&
		ports.parties.Ready() && ports.tasksDecisions.Ready()
}

type evidencePort struct {
	rows    *projectionruntime.EvidenceRows
	rebuild *projectionruntime.RebuildService
}

func (port *evidencePort) RefreshEvidenceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return port.rows.RefreshEvidenceTx(ctx, tx, recordID)
}

func (port *evidencePort) LoadEvidenceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return port.rows.LoadEvidenceTx(ctx, tx, recordID)
}

func (port *evidencePort) RefreshEvidenceSupportTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return port.rebuild.RebuildIncidentViewsTx(ctx, tx, incidentID, []string{
		timeline.TimelineViewSchemaID,
		hostidentity.HostsViewSchemaID,
		hostidentity.IdentitiesViewSchemaID,
	})
}

func (port *evidencePort) RebuildEvidenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return port.rows.RebuildEvidenceTx(ctx, tx, incidentID)
}

func validateRecoveryState(
	descriptors providercontract.DescriptorSet,
	contribution recoverystate.Contribution,
) error {
	descriptorTables := make([]string, 0, len(contribution.Tables))
	for _, descriptor := range descriptors.All() {
		if descriptor.Status == providercontract.ProviderStatusActive {
			descriptorTables = append(descriptorTables, descriptor.ProjectionTableIDs...)
		}
	}
	slices.Sort(descriptorTables)
	stateTables := make([]string, 0, len(contribution.Tables))
	for _, table := range contribution.Tables {
		stateTables = append(stateTables, table.TableName)
	}
	slices.Sort(stateTables)
	if !slices.Equal(descriptorTables, stateTables) {
		return fmt.Errorf(
			"active descriptor projection tables %v do not match recovery state tables %v",
			descriptorTables,
			stateTables,
		)
	}
	return nil
}
