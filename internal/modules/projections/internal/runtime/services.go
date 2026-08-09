package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type QueryService struct {
	store *Store
}

type RebuildService struct {
	store *Store
}

type Coordinator struct {
	store *Store
}

func NewCoordinator(pool postgres.DB, catalog *Catalog) *Coordinator {
	return &Coordinator{store: NewStore(pool, catalog)}
}

func (c *Coordinator) RefreshRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	return c.store.RefreshRowTx(ctx, tx, viewSchemaID, recordID)
}

func (c *Coordinator) LoadRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return c.store.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}

func (c *Coordinator) Supports(viewSchemaID string) bool {
	return c != nil && c.store != nil && c.store.SupportsQuerySurface(viewSchemaID)
}

func (c *Coordinator) RebuildIncidentViewTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, incidentID uuid.UUID) error {
	return c.store.RebuildIncidentViewTx(ctx, tx, viewSchemaID, incidentID)
}

func (c *Coordinator) RebuildIncidentViewsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, viewSchemaIDs []string) error {
	return c.store.RebuildIncidentViewsTx(ctx, tx, incidentID, viewSchemaIDs)
}

func (c *Coordinator) RebuildIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return c.store.RebuildIncidentTx(ctx, tx, incidentID)
}

func NewRebuildService(pool postgres.DB, catalog *Catalog) *RebuildService {
	return &RebuildService{store: NewStore(pool, catalog)}
}

func (r *RebuildService) RebuildIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return r.store.RebuildIncidentTx(ctx, tx, incidentID)
}

func (r *RebuildService) RebuildTimeline(ctx context.Context, incidentID uuid.UUID) error {
	return r.store.RebuildIncidentTimeline(ctx, incidentID)
}

func (r *RebuildService) RebuildHosts(ctx context.Context, incidentID uuid.UUID) error {
	return r.store.RebuildIncidentHosts(ctx, incidentID)
}

func (r *RebuildService) RebuildIdentities(ctx context.Context, incidentID uuid.UUID) error {
	return r.store.RebuildIncidentIdentities(ctx, incidentID)
}

func (r *RebuildService) RebuildIndicators(ctx context.Context, incidentID uuid.UUID) error {
	return r.store.RebuildIncidentIndicators(ctx, incidentID)
}

func (r *RebuildService) RebuildAssessments(ctx context.Context, incidentID uuid.UUID) error {
	tx, err := r.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin assessment projection rebuild: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := r.store.RebuildIncidentViewTx(ctx, tx, assessmentsViewSchemaID, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit assessment projection rebuild: %w", err)
	}
	return nil
}

func (r *RebuildService) RebuildArtifacts(ctx context.Context, incidentID uuid.UUID) error {
	tx, err := r.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin artifact projection rebuild: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := r.store.RebuildIncidentViewTx(ctx, tx, notesViewSchemaID, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit artifact projection rebuild: %w", err)
	}
	return nil
}

func (r *RebuildService) RebuildEvidence(ctx context.Context, incidentID uuid.UUID) error {
	tx, err := r.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin Evidence projection rebuild: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := r.store.RebuildIncidentViewTx(ctx, tx, evidenceViewSchemaID, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit Evidence projection rebuild: %w", err)
	}
	return nil
}

func (r *RebuildService) RebuildParties(ctx context.Context, incidentID uuid.UUID) error {
	tx, err := r.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin Party projection rebuild: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := r.store.RebuildIncidentViewTx(ctx, tx, partiesViewSchemaID, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit Party projection rebuild: %w", err)
	}
	return nil
}

func (r *RebuildService) RebuildTaskRequests(ctx context.Context, incidentID uuid.UUID) error {
	return r.rebuildTaskDecisionProvider(ctx, incidentID, taskRequestsViewSchemaID, "task-request")
}

func (r *RebuildService) RebuildDecisions(ctx context.Context, incidentID uuid.UUID) error {
	return r.rebuildTaskDecisionProvider(ctx, incidentID, decisionsViewSchemaID, "decision")
}

func (r *RebuildService) rebuildTaskDecisionProvider(
	ctx context.Context,
	incidentID uuid.UUID,
	viewSchemaID string,
	providerName string,
) error {
	tx, err := r.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin %s projection rebuild: %w", providerName, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := r.store.RebuildIncidentViewTx(ctx, tx, viewSchemaID, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit %s projection rebuild: %w", providerName, err)
	}
	return nil
}

func (r *RebuildService) RebuildImportedIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return r.RebuildIncidentTx(ctx, tx, incidentID)
}

func (r *RebuildService) RebuildIncidentViewsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, viewSchemaIDs []string) error {
	return r.store.RebuildIncidentViewsTx(ctx, tx, incidentID, viewSchemaIDs)
}

func (r *RebuildService) RestoreRebuilder() *RestoreRebuilder {
	return NewRestoreRebuilderFromStore(r.store)
}

func NewQueryService(pool postgres.DB, catalog *Catalog) *QueryService {
	return &QueryService{store: NewStore(pool, catalog)}
}

func (q *QueryService) Supports(viewSchemaID string) bool {
	if q == nil || q.store == nil || q.store.registry == nil {
		return false
	}
	_, ok := q.store.registry.querySurfaceForView(viewSchemaID)
	return ok
}

func (q *QueryService) QueryRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta) ([]map[string]any, error) {
	return q.store.QueryRows(ctx, incidentID, viewSchemaID, query)
}

func (q *QueryService) QueryRowsPage(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
	return q.store.QueryRowsPage(ctx, incidentID, viewSchemaID, query, window)
}

func (q *QueryService) LoadRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return q.store.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}

type TimelineRows struct {
	store *Store
}

func NewTimelineRows(pool postgres.DB) *TimelineRows {
	return &TimelineRows{store: NewStore(pool, nil)}
}

func (r *TimelineRows) ApplyTimelineMutationTx(ctx context.Context, tx pgx.Tx, mutation timelineprojection.ProjectionMutation) error {
	return r.store.ApplyTimelineMutationTx(ctx, tx, mutation)
}

type EntityRows struct {
	store          *Store
	source         entityprojection.SourceReader
	hostReader     *queryengine.HostReader
	identityReader *queryengine.IdentityReader
}

type IndicatorRows struct {
	store    *Store
	source   indicatorprojection.SourceReader
	surfaces map[string]genericSurface
}

func NewIndicatorRows(
	pool postgres.DB,
	source indicatorprojection.SourceReader,
) *IndicatorRows {
	return &IndicatorRows{
		store:    NewStore(pool, nil),
		source:   source,
		surfaces: rowPlans(queryengine.IndicatorPlans()),
	}
}

func (r *IndicatorRows) RefreshIndicatorTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) error {
	return r.store.refreshIndicatorTxCore(ctx, tx, recordID, r.source)
}

func (r *IndicatorRows) LoadIndicatorTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.surfaces, indicatorsViewSchemaID, recordID)
}

func (r *IndicatorRows) DeleteIndicatorTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) error {
	if r == nil || r.store == nil || r.store.physical == nil {
		return errors.New("projection storage is required")
	}
	return r.store.physical.DeleteIndicatorRowTx(ctx, tx, recordID)
}

func (r *IndicatorRows) RebuildIndicatorsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) error {
	return r.store.rebuildIncidentIndicatorsTxCore(ctx, tx, incidentID, r.source)
}

func NewEntityRows(pool postgres.DB, sources ...entityprojection.SourceReader) *EntityRows {
	var source entityprojection.SourceReader
	if len(sources) == 1 {
		source = sources[0]
	}
	return &EntityRows{
		store:          NewStore(pool, nil),
		source:         source,
		hostReader:     queryengine.NewHostReader(pool),
		identityReader: queryengine.NewIdentityReader(pool),
	}
}

func (r *EntityRows) RefreshHostTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshHostTxCore(ctx, tx, recordID, r.source)
}

func (r *EntityRows) RefreshIdentityTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshIdentityTxCore(ctx, tx, recordID, r.source)
}

func (r *EntityRows) DeleteHostTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if r == nil || r.store == nil || r.store.physical == nil {
		return fmt.Errorf("projection storage is required")
	}
	return r.store.physical.DeleteHostRowTx(ctx, tx, recordID)
}

func (r *EntityRows) DeleteIdentityTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if r == nil || r.store == nil || r.store.physical == nil {
		return fmt.Errorf("projection storage is required")
	}
	return r.store.physical.DeleteIdentityRowTx(ctx, tx, recordID)
}

func (r *EntityRows) RebuildHostsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return r.store.rebuildIncidentHostsTxCore(ctx, tx, incidentID, r.source)
}

func (r *EntityRows) SelectHostQueryProjections(
	ctx context.Context,
	incidentID uuid.UUID,
	query viewschema.QueryMeta,
	window querypage.Window,
) ([]entityprojection.HostQueryProjection, error) {
	if r == nil || r.hostReader == nil {
		return nil, fmt.Errorf("host projection reader is required")
	}
	return r.hostReader.SelectHostQueryProjections(ctx, incidentID, query, window)
}

func (r *EntityRows) CollectHostDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]entityprojection.DerivedFact, error) {
	if r == nil || r.hostReader == nil {
		return nil, fmt.Errorf("host projection reader is required")
	}
	return r.hostReader.CollectHostDerivedFactsTx(ctx, tx, incidentID)
}

func (r *EntityRows) RebuildIdentitiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return r.store.rebuildIncidentIdentitiesTxCore(ctx, tx, incidentID, r.source)
}

func (r *EntityRows) SelectIdentityQueryProjections(
	ctx context.Context,
	incidentID uuid.UUID,
	query viewschema.QueryMeta,
	window querypage.Window,
) ([]entityprojection.IdentityQueryProjection, error) {
	if r == nil || r.identityReader == nil {
		return nil, fmt.Errorf("identity projection reader is required")
	}
	return r.identityReader.SelectIdentityQueryProjections(ctx, incidentID, query, window)
}

func (r *EntityRows) CollectIdentityDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]entityprojection.DerivedFact, error) {
	if r == nil || r.identityReader == nil {
		return nil, fmt.Errorf("identity projection reader is required")
	}
	return r.identityReader.CollectIdentityDerivedFactsTx(ctx, tx, incidentID)
}

type AssessmentRows struct {
	store    *Store
	source   assessmentprojection.SourceReader
	surfaces map[string]genericSurface
}

func NewAssessmentRows(
	pool postgres.DB,
	source assessmentprojection.SourceReader,
) *AssessmentRows {
	return &AssessmentRows{
		store:    NewStore(pool, nil),
		source:   source,
		surfaces: rowPlans(queryengine.AssessmentPlans()),
	}
}

func (r *AssessmentRows) RefreshAssessmentTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshAssessmentTxCore(ctx, tx, recordID, r.source)
}

func (r *AssessmentRows) ApplyAssessmentMutationTx(
	ctx context.Context,
	tx pgx.Tx,
	mutation assessmentprojection.ProjectionMutation,
) error {
	return r.store.ApplyAssessmentMutationTx(ctx, tx, mutation)
}

func (r *AssessmentRows) LoadAssessmentTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, assessmentsViewSchemaID, recordID)
}

func (r *AssessmentRows) RebuildAssessmentsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) error {
	return r.store.rebuildIncidentAssessmentsTxCore(ctx, tx, incidentID, r.source)
}

type ArtifactRows struct {
	store    *Store
	source   artifactprojection.SourceReader
	surfaces map[string]genericSurface
	reader   *queryengine.ArtifactReader
}

func NewArtifactRows(pool postgres.DB, source artifactprojection.SourceReader) *ArtifactRows {
	return &ArtifactRows{
		store:    NewStore(pool, nil),
		source:   source,
		surfaces: rowPlans(queryengine.ArtifactPlans()),
		reader:   queryengine.NewArtifactReader(),
	}
}

func (r *ArtifactRows) RefreshArtifactTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshArtifactTxCore(ctx, tx, recordID, r.source)
}

func (r *ArtifactRows) LoadArtifactTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, viewSchemaID, recordID)
}

func (r *ArtifactRows) RebuildArtifactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) error {
	return r.store.rebuildIncidentArtifactsTxCore(ctx, tx, incidentID, r.source)
}

func (r *ArtifactRows) CollectDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]artifactprojection.DerivedFact, error) {
	if r == nil || r.reader == nil {
		return nil, errors.New("artifact projection reader is required")
	}
	return r.reader.CollectDerivedFactsTx(ctx, tx, incidentID)
}

var _ artifactprojection.Rows = (*ArtifactRows)(nil)
var _ artifactprojection.Rebuilder = (*RebuildService)(nil)
var _ artifactprojection.Reader = (*ArtifactRows)(nil)

type EvidenceRows struct {
	store    *Store
	source   evidenceprojection.SourceReader
	surfaces map[string]genericSurface
}

func NewEvidenceRows(pool postgres.DB, source evidenceprojection.SourceReader) *EvidenceRows {
	return &EvidenceRows{
		store:    NewStore(pool, nil),
		source:   source,
		surfaces: rowPlans(queryengine.EvidencePlans()),
	}
}

func (r *EvidenceRows) RefreshEvidenceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshEvidenceTxCore(ctx, tx, recordID, r.source)
}

func (r *EvidenceRows) LoadEvidenceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, evidenceViewSchemaID, recordID)
}

func (r *EvidenceRows) RebuildEvidenceTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) error {
	return r.store.rebuildIncidentEvidenceTxCore(ctx, tx, incidentID, r.source)
}

var _ evidenceprojection.Rebuilder = (*RebuildService)(nil)

type PartyRows struct {
	store    *Store
	source   partyprojection.SourceReader
	surfaces map[string]genericSurface
}

func NewPartyRows(pool postgres.DB, source partyprojection.SourceReader) *PartyRows {
	return &PartyRows{
		store:    NewStore(pool, nil),
		source:   source,
		surfaces: rowPlans(queryengine.PartyPlans()),
	}
}

func (r *PartyRows) RefreshPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshPartyTxCore(ctx, tx, recordID, r.source)
}

func (r *PartyRows) LoadPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, partiesViewSchemaID, recordID)
}

func (r *PartyRows) RebuildPartiesTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) error {
	return r.store.rebuildIncidentPartiesTxCore(ctx, tx, incidentID, r.source)
}

var _ partyprojection.Rows = (*PartyRows)(nil)
var _ partyprojection.Rebuilder = (*RebuildService)(nil)

type TaskDecisionRows struct {
	store             *Store
	taskRequestSource TaskRequestSource
	decisionSource    DecisionSource
	taskReader        taskdecisionprojection.TaskReader
	decisionReader    interface {
		CollectDecisionDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]taskdecisionprojection.DecisionDerivedFact, error)
	}
	surfaces map[string]genericSurface
}

func NewTaskDecisionRows(
	pool postgres.DB,
	taskRequestSource TaskRequestSource,
	decisionSource DecisionSource,
) *TaskDecisionRows {
	surfaces := queryengine.TaskRequestPlans()
	surfaces = append(surfaces, queryengine.DecisionPlans()...)
	return &TaskDecisionRows{
		store:             NewStore(pool, nil),
		taskRequestSource: taskRequestSource,
		decisionSource:    decisionSource,
		taskReader:        queryengine.NewTaskReader(),
		decisionReader:    queryengine.NewDecisionReader(),
		surfaces:          rowPlans(surfaces),
	}
}

func (r *TaskDecisionRows) CollectDecisionDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]taskdecisionprojection.DecisionDerivedFact, error) {
	if r == nil || r.decisionReader == nil {
		return nil, errors.New("decision projection reader is required")
	}
	return r.decisionReader.CollectDecisionDerivedFactsTx(ctx, tx, incidentID)
}

func (r *TaskDecisionRows) RefreshTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshTaskRequestTxCore(ctx, tx, recordID, r.taskRequestSource)
}

func (r *TaskDecisionRows) RefreshDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshDecisionTxCore(ctx, tx, recordID, r.decisionSource)
}

func (r *TaskDecisionRows) LoadTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, taskRequestsViewSchemaID, recordID)
}

func (r *TaskDecisionRows) LoadDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, decisionsViewSchemaID, recordID)
}

func (r *TaskDecisionRows) RebuildTaskRequestsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) error {
	return r.store.rebuildIncidentTaskRequestsTxCore(ctx, tx, incidentID, r.taskRequestSource)
}

func (r *TaskDecisionRows) RebuildDecisionsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) error {
	return r.store.rebuildIncidentDecisionsTxCore(ctx, tx, incidentID, r.decisionSource)
}

func (r *TaskDecisionRows) CollectTaskDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]taskdecisionprojection.TaskDerivedFact, error) {
	if r == nil || r.taskReader == nil {
		return nil, errors.New("task-request projection reader is required")
	}
	return r.taskReader.CollectTaskDerivedFactsTx(ctx, tx, incidentID)
}

var _ taskdecisionprojection.Rows = (*TaskDecisionRows)(nil)
var _ taskdecisionprojection.Rebuilder = (*RebuildService)(nil)
var _ taskdecisionprojection.Reader = (*TaskDecisionRows)(nil)

func rowPlans(contracts []queryengine.Surface) map[string]genericSurface {
	surfaces := make(map[string]genericSurface, len(contracts))
	for _, contract := range contracts {
		surface, err := genericSurfaceFromPlan(contract)
		if err != nil {
			panic(fmt.Sprintf("construct provider row query surface %q: %v", contract.ViewSchemaID, err))
		}
		if _, exists := surfaces[surface.viewSchemaID]; exists {
			panic(fmt.Sprintf("construct provider row query surface %q: duplicate", surface.viewSchemaID))
		}
		surfaces[surface.viewSchemaID] = surface
	}
	return surfaces
}

func (r *AssessmentRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.surfaces, viewSchemaID, recordID)
}

func (r *ArtifactRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.surfaces, viewSchemaID, recordID)
}

func (r *EvidenceRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.surfaces, viewSchemaID, recordID)
}

func (r *PartyRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.surfaces, viewSchemaID, recordID)
}

func (r *TaskDecisionRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.surfaces, viewSchemaID, recordID)
}

func loadProviderRowTx(ctx context.Context, tx pgx.Tx, surfaces map[string]genericSurface, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	surface, ok := surfaces[viewSchemaID]
	if !ok {
		return nil, fmt.Errorf("provider row query surface %q not configured", viewSchemaID)
	}
	return loadRowTx(ctx, tx, surface, recordID)
}
