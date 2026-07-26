package adapters

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	TimelineViewSchemaID             = "cartulary.view.timeline.v2"
	HostsViewSchemaID                = "cartulary.view.hosts.v1"
	IdentitiesViewSchemaID           = "cartulary.view.identities.v1"
	IndicatorsViewSchemaID           = "cartulary.view.indicators.v1"
	AssessmentsViewSchemaID          = "cartulary.view.assessments.v1"
	EvidenceViewSchemaID             = "cartulary.view.evidence.v1"
	NotesViewSchemaID                = "cartulary.view.notes.v1"
	PartiesViewSchemaID              = "cartulary.view.parties.v1"
	TaskRequestsViewSchemaID         = "cartulary.view.task_requests.v1"
	DecisionsViewSchemaID            = "cartulary.view.decisions.v1"
	CommLogViewSchemaID              = "cartulary.view.comm_log.v1"
	HandoffViewSchemaID              = "cartulary.view.handoff.v1"
	StatusReviewViewSchemaID         = "cartulary.view.status_review.v1"
	LessonViewSchemaID               = "cartulary.view.lesson.v1"
	FindingsViewSchemaID             = "cartulary.view.findings.v1"
	InvestigativeQueriesViewSchemaID = "cartulary.view.investigative_queries.v1"
	ForensicKeywordsViewSchemaID     = "cartulary.view.forensic_keywords.v1"
)

type WorkbookRows struct {
	store *projections.Store
}

func NewWorkbookRows(pool postgres.DB) *WorkbookRows {
	return &WorkbookRows{store: projections.NewStore(pool)}
}

func (w *WorkbookRows) Supports(viewSchemaID string) bool {
	return projections.SupportsQuerySurface(viewSchemaID)
}

func (w *WorkbookRows) QueryRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta) ([]map[string]any, error) {
	return w.store.QueryRows(ctx, incidentID, viewSchemaID, query)
}

func (w *WorkbookRows) QueryRowsPage(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
	return w.store.QueryRowsPage(ctx, incidentID, viewSchemaID, query, window)
}

func (w *WorkbookRows) LoadRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return w.store.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}

type RowProjector struct {
	store *projections.Store
}

func NewRowProjector(pool postgres.DB, timelineSources ...projections.TimelineSource) *RowProjector {
	options := make([]projections.StoreOption, 0, 1)
	if len(timelineSources) > 0 && timelineSources[0] != nil {
		options = append(options, projections.WithTimelineSource(timelineSources[0]))
	}
	return &RowProjector{store: projections.NewStore(pool, options...)}
}

func (p *RowProjector) RefreshRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	return p.store.RefreshRowTx(ctx, tx, viewSchemaID, recordID)
}

func (p *RowProjector) DeleteRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	return p.store.DeleteRowTx(ctx, tx, viewSchemaID, recordID)
}

func (p *RowProjector) LoadRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return p.store.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}

func (p *RowProjector) QueryRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta) ([]map[string]any, error) {
	return p.store.QueryRows(ctx, incidentID, viewSchemaID, query)
}

func (p *RowProjector) RebuildIncidentTimeline(ctx context.Context, incidentID uuid.UUID) error {
	return p.store.RebuildIncidentTimeline(ctx, incidentID)
}

func (p *RowProjector) RebuildIncidentHosts(ctx context.Context, incidentID uuid.UUID) error {
	return p.store.RebuildIncidentHosts(ctx, incidentID)
}

func (p *RowProjector) RebuildIncidentIdentities(ctx context.Context, incidentID uuid.UUID) error {
	return p.store.RebuildIncidentIdentities(ctx, incidentID)
}

func (p *RowProjector) RebuildIncidentIndicators(ctx context.Context, incidentID uuid.UUID) error {
	return p.store.RebuildIncidentIndicators(ctx, incidentID)
}

func (p *RowProjector) RebuildIncidentViewTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, incidentID uuid.UUID) error {
	return p.store.RebuildIncidentViewTx(ctx, tx, viewSchemaID, incidentID)
}

func (p *RowProjector) RebuildIncidentViewsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, viewSchemaIDs []string) error {
	return p.store.RebuildIncidentViewsTx(ctx, tx, incidentID, viewSchemaIDs)
}

func (p *RowProjector) RebuildIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return p.store.RebuildIncidentTx(ctx, tx, incidentID)
}

type TimelineWriter struct {
	store *projections.Store
}

func NewTimelineWriter(pool postgres.DB) *TimelineWriter {
	return &TimelineWriter{store: projections.NewStore(pool)}
}

func (w *TimelineWriter) ApplyTimelineMutationTx(ctx context.Context, tx pgx.Tx, mutation timelineprojection.ProjectionMutation) error {
	return w.store.ApplyTimelineMutationTx(ctx, tx, mutation)
}

func (w *TimelineWriter) RebuildIncidentHostsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return w.store.RebuildIncidentViewTx(ctx, tx, HostsViewSchemaID, incidentID)
}

func (w *TimelineWriter) RebuildIncidentIdentitiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return w.store.RebuildIncidentViewTx(ctx, tx, IdentitiesViewSchemaID, incidentID)
}

func NewRestoreRebuilder(pool postgres.DB, timelineSources ...projections.TimelineSource) restorecontract.ProjectionRebuilder {
	return projections.NewRestoreRebuilder(pool, timelineSources...)
}
