package projections

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
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

func (c *Coordinator) DeleteRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	return c.store.DeleteRowTx(ctx, tx, viewSchemaID, recordID)
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
	store *Store
}

func NewEntityRows(pool postgres.DB) *EntityRows {
	return &EntityRows{store: NewStore(pool, nil)}
}

func (r *EntityRows) RefreshHostTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshHostTxCore(ctx, tx, recordID)
}

func (r *EntityRows) RefreshIdentityTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshIdentityTxCore(ctx, tx, recordID)
}

func (r *EntityRows) DeleteHostTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.DeleteRowTx(ctx, tx, hostsViewSchemaID, recordID)
}

func (r *EntityRows) DeleteIdentityTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.DeleteRowTx(ctx, tx, identitiesViewSchemaID, recordID)
}

func (r *EntityRows) RebuildHostsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return r.store.rebuildIncidentHostsTxCore(ctx, tx, incidentID)
}

func (r *EntityRows) RebuildIdentitiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return r.store.rebuildIncidentIdentitiesTxCore(ctx, tx, incidentID)
}

type AssessmentRows struct {
	store    *Store
	source   AssessmentSource
	surfaces map[string]genericSurface
}

func NewAssessmentRows(
	pool postgres.DB,
	source AssessmentSource,
	querySurfaces ...providercontract.QuerySurface,
) *AssessmentRows {
	return &AssessmentRows{
		store:    NewStore(pool, nil),
		source:   source,
		surfaces: rowQuerySurfaces(querySurfaces),
	}
}

func (r *AssessmentRows) RefreshTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshAssessmentTxCore(ctx, tx, recordID, r.source)
}

func (r *AssessmentRows) ApplyMutationTx(
	ctx context.Context,
	tx pgx.Tx,
	mutation assessmentprojection.ProjectionMutation,
) error {
	return r.store.ApplyAssessmentMutationTx(ctx, tx, mutation)
}

func (r *AssessmentRows) LoadTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, assessmentsViewSchemaID, recordID)
}

type ArtifactRows struct {
	store    *Store
	surfaces map[string]genericSurface
}

func NewArtifactRows(pool postgres.DB, querySurfaces ...providercontract.QuerySurface) *ArtifactRows {
	return &ArtifactRows{store: NewStore(pool, nil), surfaces: rowQuerySurfaces(querySurfaces)}
}

func (r *ArtifactRows) RefreshTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshArtifactTxCore(ctx, tx, recordID)
}

func (r *ArtifactRows) LoadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, viewSchemaID, recordID)
}

type EvidenceRows struct {
	store    *Store
	surfaces map[string]genericSurface
}

func NewEvidenceRows(pool postgres.DB, querySurfaces ...providercontract.QuerySurface) *EvidenceRows {
	return &EvidenceRows{store: NewStore(pool, nil), surfaces: rowQuerySurfaces(querySurfaces)}
}

func (r *EvidenceRows) RefreshTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshEvidenceTxCore(ctx, tx, recordID)
}

func (r *EvidenceRows) LoadTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, evidenceViewSchemaID, recordID)
}

type PartyRows struct {
	store    *Store
	surfaces map[string]genericSurface
}

func NewPartyRows(pool postgres.DB, querySurfaces ...providercontract.QuerySurface) *PartyRows {
	return &PartyRows{store: NewStore(pool, nil), surfaces: rowQuerySurfaces(querySurfaces)}
}

func (r *PartyRows) RefreshTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshPartyTxCore(ctx, tx, recordID)
}

func (r *PartyRows) LoadTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, partiesViewSchemaID, recordID)
}

type TaskDecisionRows struct {
	store    *Store
	source   TaskDecisionSource
	surfaces map[string]genericSurface
}

func NewTaskDecisionRows(
	pool postgres.DB,
	source TaskDecisionSource,
	querySurfaces ...providercontract.QuerySurface,
) *TaskDecisionRows {
	return &TaskDecisionRows{
		store:    NewStore(pool, nil),
		source:   source,
		surfaces: rowQuerySurfaces(querySurfaces),
	}
}

func (r *TaskDecisionRows) RefreshTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshTaskRequestTxCore(ctx, tx, recordID, r.source)
}

func (r *TaskDecisionRows) RefreshDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshDecisionTxCore(ctx, tx, recordID, r.source)
}

func (r *TaskDecisionRows) LoadTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, taskRequestsViewSchemaID, recordID)
}

func (r *TaskDecisionRows) LoadDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, decisionsViewSchemaID, recordID)
}

func (r *TaskDecisionRows) LoadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	switch viewSchemaID {
	case taskRequestsViewSchemaID:
		return r.LoadTaskRequestTx(ctx, tx, recordID)
	case decisionsViewSchemaID:
		return r.LoadDecisionTx(ctx, tx, recordID)
	default:
		return nil, fmt.Errorf("task/decision projection view %q is not supported", viewSchemaID)
	}
}

func rowQuerySurfaces(contracts []providercontract.QuerySurface) map[string]genericSurface {
	surfaces := make(map[string]genericSurface, len(contracts))
	for _, contract := range contracts {
		surface, err := genericSurfaceFromContract(contract)
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
