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
	catalog        *projectionruntime.Catalog
	store          *projectionruntime.Store
	restore        *projectionruntime.RestoreRebuilder
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

	catalog, err := projectionruntime.NewCatalog(
		contracts,
		projectionruntime.ProviderSources{
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
		return Ports{}, fmt.Errorf("compose Projections catalog: %w", err)
	}
	physical, err := projectionstorage.New(dependencies.Postgres)
	if err != nil {
		return Ports{}, fmt.Errorf("compose Projections storage: %w", err)
	}
	store, err := projectionruntime.NewStore(dependencies.Postgres, catalog, physical)
	if err != nil {
		return Ports{}, fmt.Errorf("compose Projections store: %w", err)
	}
	recoveryState := providercontract.RecoveryStateContribution()
	if err := validateRecoveryState(catalog.DescriptorSet(), recoveryState); err != nil {
		return Ports{}, fmt.Errorf("compose Projections recovery state: %w", err)
	}
	restoreRebuilder := projectionruntime.NewRestoreRebuilderFromStore(store)
	entityRows := projectionruntime.NewEntityRowsFromStore(store, dependencies.Entities.Source())
	artifactRows := projectionruntime.NewArtifactRowsFromStore(store, dependencies.Artifacts.Source())
	taskDecisionRows := projectionruntime.NewTaskDecisionRowsFromStore(
		store,
		dependencies.TasksDecisions.TaskRequestSource(),
		dependencies.TasksDecisions.DecisionSource(),
	)
	evidenceRows := &evidencePort{
		rows:    projectionruntime.NewEvidenceRowsFromStore(store, dependencies.Evidence.Source()),
		rebuild: store,
	}
	ports := Ports{
		catalog: catalog,
		store:   store,
		restore: restoreRebuilder,
		timeline: timelineprojection.Ports{
			Writer:    projectionruntime.NewTimelineRowsFromStore(store),
			Rebuilder: store,
		},
		entities: entityprojection.Ports{
			Writer:    entityRows,
			Rebuilder: store,
			Reader:    entityRows,
		},
		indicators: indicatorprojection.Ports{
			Rows: projectionruntime.NewIndicatorRowsFromStore(
				store,
				dependencies.Indicators.Source(),
			),
			Rebuilder: store,
		},
		assessments: assessmentprojection.Ports{
			Rows: projectionruntime.NewAssessmentRowsFromStore(
				store,
				dependencies.Assessments.Source(),
			),
			Rebuilder: store,
		},
		artifacts: artifactprojection.Ports{
			Rows:      artifactRows,
			Rebuilder: store,
			Reader:    artifactRows,
		},
		evidence: evidenceprojection.Ports{
			Rows:      evidenceRows,
			Rebuilder: store,
		},
		parties: partyprojection.Ports{
			Rows:      projectionruntime.NewPartyRowsFromStore(store, dependencies.Parties.Source()),
			Rebuilder: store,
		},
		tasksDecisions: taskdecisionprojection.Ports{
			Rows:      taskDecisionRows,
			Rebuilder: store,
			Reader:    taskDecisionRows,
		},
	}
	if err := ports.validate(); err != nil {
		return Ports{}, fmt.Errorf("compose Projections ports: %w", err)
	}
	return ports, nil
}

func (ports Ports) DescriptorSet() providercontract.DescriptorSet {
	if ports.catalog == nil {
		return providercontract.DescriptorSet{}
	}
	return ports.catalog.DescriptorSet()
}

func (ports Ports) RecoveryStateContribution() recoverystate.Contribution {
	return providercontract.RecoveryStateContribution()
}

func (ports Ports) WorkbookQueryProvider(viewSchemaID string) (workbook.QueryProvider, bool) {
	if ports.store == nil || !ports.store.Supports(viewSchemaID) {
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
		return ports.store.QueryRowsPage(
			ctx,
			command.IncidentID,
			viewSchemaID,
			command.Query,
			command.Window,
		)
	}), true
}

func (ports Ports) RecoveryPorts() restorecontract.ProjectionPorts {
	if ports.restore == nil {
		return restorecontract.ProjectionPorts{}
	}
	return restorecontract.ProjectionPorts{
		Rebuilder:         ports.restore,
		StateContribution: ports.RecoveryStateContribution(),
	}
}

func (ports Ports) RestoreProbeQuery() workbookrestoreprobe.ProjectionQuery {
	if ports.store == nil {
		return nil
	}
	return ports.store
}

func (ports Ports) RevisionServices() revisions.ProjectionServices {
	if ports.store == nil {
		return nil
	}
	return ports.store
}

func (ports Ports) RevisionLiveRecords() revisions.LiveRecordReader {
	if ports.store == nil {
		return nil
	}
	return ports.store
}

type SourceTextRows interface {
	RefreshRowTx(context.Context, pgx.Tx, string, uuid.UUID) error
	LoadRowTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error)
}

type ImportRebuilder interface {
	RebuildImportedIncidentTx(context.Context, pgx.Tx, uuid.UUID) error
}

func (ports Ports) SourceTextRows() SourceTextRows {
	if ports.store == nil {
		return nil
	}
	return ports.store
}

func (ports Ports) ImportRebuilder() ImportRebuilder {
	if ports.store == nil {
		return nil
	}
	return ports.store
}

func (ports Ports) Timeline() timelineprojection.Ports { return ports.timeline }

func (ports Ports) Entities() entityprojection.Ports { return ports.entities }

func (ports Ports) Indicators() indicatorprojection.Ports { return ports.indicators }

func (ports Ports) Assessments() assessmentprojection.Ports { return ports.assessments }

func (ports Ports) Artifacts() artifactprojection.Ports { return ports.artifacts }

func (ports Ports) Evidence() evidenceprojection.Ports { return ports.evidence }

func (ports Ports) Parties() partyprojection.Ports { return ports.parties }

func (ports Ports) TasksDecisions() taskdecisionprojection.Ports { return ports.tasksDecisions }

func (ports Ports) validate() error {
	if ports.catalog == nil || ports.DescriptorSet().Len() == 0 {
		return fmt.Errorf("descriptor catalog is required")
	}
	if ports.store == nil {
		return fmt.Errorf("runtime store is required")
	}
	if ports.restore == nil {
		return fmt.Errorf("restore rebuilder is required")
	}
	ownerPorts := []struct {
		name  string
		ready bool
	}{
		{name: "Timeline", ready: ports.timeline.Writer != nil && ports.timeline.Rebuilder != nil},
		{name: "Entities", ready: ports.entities.Writer != nil && ports.entities.Rebuilder != nil && ports.entities.Reader != nil},
		{name: "Indicators", ready: ports.indicators.Rows != nil && ports.indicators.Rebuilder != nil},
		{name: "Assessments", ready: ports.assessments.Rows != nil && ports.assessments.Rebuilder != nil},
		{name: "Artifacts", ready: ports.artifacts.Rows != nil && ports.artifacts.Rebuilder != nil && ports.artifacts.Reader != nil},
		{name: "Evidence", ready: ports.evidence.Rows != nil && ports.evidence.Rebuilder != nil},
		{name: "Parties", ready: ports.parties.Rows != nil && ports.parties.Rebuilder != nil},
		{name: "Tasks/Decisions", ready: ports.tasksDecisions.Rows != nil && ports.tasksDecisions.Rebuilder != nil && ports.tasksDecisions.Reader != nil},
	}
	for _, owner := range ownerPorts {
		if !owner.ready {
			return fmt.Errorf("%s ports are incomplete", owner.name)
		}
	}
	return nil
}

type evidencePort struct {
	rows    *projectionruntime.EvidenceRows
	rebuild interface {
		RebuildIncidentViewsTx(context.Context, pgx.Tx, uuid.UUID, []string) error
	}
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
