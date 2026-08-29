package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entitycontract "github.com/JochiRaider/cartulary/internal/modules/entities/projectioncontract"
	entityports "github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectioncontract"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionports"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type TimelineRows struct {
	store *Store
}

func NewTimelineRowsFromStore(store *Store) *TimelineRows {
	return &TimelineRows{store: store}
}

func (r *TimelineRows) ApplyTimelineMutationTx(ctx context.Context, tx pgx.Tx, mutation timelineprojection.ProjectionMutation) error {
	return r.store.applyTimelineMutationTx(ctx, tx, mutation)
}

func (r *TimelineRows) ApplyTimelineFixtureBatchTx(ctx context.Context, tx pgx.Tx, inputs []timelineprojection.ProjectionInput) error {
	return r.store.applyTimelineFixtureBatchTx(ctx, tx, inputs)
}

func (r *TimelineRows) CountTimelineFixtureRows(ctx context.Context, incidentID uuid.UUID) (int, error) {
	return r.store.countTimelineFixtureRows(ctx, incidentID)
}

func (r *TimelineRows) CountTimelineFixtureRowsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (int, error) {
	return r.store.countTimelineFixtureRowsTx(ctx, tx, incidentID)
}

type entityRows struct {
	store          *Store
	source         entitycontract.SourceReader
	hostReader     *queryengine.HostReader
	identityReader *queryengine.IdentityReader
}

type IndicatorRows struct {
	store  *Store
	source indicatorprojection.SourceReader
}

func NewIndicatorRowsFromStore(
	store *Store,
	source indicatorprojection.SourceReader,
) *IndicatorRows {
	return &IndicatorRows{
		store:  store,
		source: source,
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
	return loadProviderRowTx(ctx, tx, r.store, indicatorsViewSchemaID, recordID)
}

func NewEntityPortViewsFromStore(
	store *Store,
	source entitycontract.SourceReader,
) (entityports.MutationRows, entityports.QueryReader, entityports.ReportingReader) {
	rows := &entityRows{
		store:          store,
		source:         source,
		hostReader:     queryengine.NewHostReader(store.pool),
		identityReader: queryengine.NewIdentityReader(store.pool),
	}
	return rows, rows, rows
}

func (r *entityRows) RefreshHostTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshHostTxCore(ctx, tx, recordID, r.source)
}

func (r *entityRows) RefreshIdentityTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshIdentityTxCore(ctx, tx, recordID, r.source)
}

func (r *entityRows) DeleteHostTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if r == nil || r.store == nil || r.store.physical == nil {
		return fmt.Errorf("projection storage is required")
	}
	return r.store.physical.DeleteHostRowTx(ctx, tx, recordID)
}

func (r *entityRows) DeleteIdentityTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if r == nil || r.store == nil || r.store.physical == nil {
		return fmt.Errorf("projection storage is required")
	}
	return r.store.physical.DeleteIdentityRowTx(ctx, tx, recordID)
}

func (r *entityRows) SelectHostQueryProjections(
	ctx context.Context,
	incidentID uuid.UUID,
	query viewschema.QueryMeta,
	window querypage.Window,
) ([]entityports.HostQueryProjection, error) {
	if r == nil || r.hostReader == nil {
		return nil, fmt.Errorf("host projection reader is required")
	}
	return r.hostReader.SelectHostQueryProjections(ctx, incidentID, query, window)
}

func (r *entityRows) CollectHostDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]entityports.DerivedFact, error) {
	if r == nil || r.hostReader == nil {
		return nil, fmt.Errorf("host projection reader is required")
	}
	return r.hostReader.CollectHostDerivedFactsTx(ctx, tx, incidentID)
}

func (r *entityRows) SelectIdentityQueryProjections(
	ctx context.Context,
	incidentID uuid.UUID,
	query viewschema.QueryMeta,
	window querypage.Window,
) ([]entityports.IdentityQueryProjection, error) {
	if r == nil || r.identityReader == nil {
		return nil, fmt.Errorf("identity projection reader is required")
	}
	return r.identityReader.SelectIdentityQueryProjections(ctx, incidentID, query, window)
}

func (r *entityRows) CollectIdentityDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]entityports.DerivedFact, error) {
	if r == nil || r.identityReader == nil {
		return nil, fmt.Errorf("identity projection reader is required")
	}
	return r.identityReader.CollectIdentityDerivedFactsTx(ctx, tx, incidentID)
}

type AssessmentRows struct {
	store  *Store
	source assessmentprojection.SourceReader
}

func NewAssessmentRowsFromStore(
	store *Store,
	source assessmentprojection.SourceReader,
) *AssessmentRows {
	return &AssessmentRows{
		store:  store,
		source: source,
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
	return r.store.applyAssessmentMutationTx(ctx, tx, mutation)
}

func (r *AssessmentRows) LoadAssessmentTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, assessmentsViewSchemaID, recordID)
}

type ArtifactRows struct {
	store  *Store
	source artifactprojection.SourceReader
	reader *queryengine.ArtifactReader
}

func NewArtifactRowsFromStore(store *Store, source artifactprojection.SourceReader) *ArtifactRows {
	return &ArtifactRows{
		store:  store,
		source: source,
		reader: queryengine.NewArtifactReader(),
	}
}

func (r *ArtifactRows) RefreshArtifactTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshArtifactTxCore(ctx, tx, recordID, r.source)
}

func (r *ArtifactRows) LoadArtifactTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, viewSchemaID, recordID)
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
var _ artifactprojection.Reader = (*ArtifactRows)(nil)

type EvidenceRows struct {
	store  *Store
	source evidenceprojection.SourceReader
}

func NewEvidenceRowsFromStore(store *Store, source evidenceprojection.SourceReader) *EvidenceRows {
	return &EvidenceRows{
		store:  store,
		source: source,
	}
}

func (r *EvidenceRows) RefreshEvidenceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshEvidenceTxCore(ctx, tx, recordID, r.source)
}

func (r *EvidenceRows) LoadEvidenceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	if r == nil || r.source == nil {
		return nil, errors.New("evidence projection source is required")
	}
	input, found, err := r.source.LoadProjectionInputTx(ctx, tx, recordID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, pgx.ErrNoRows
	}
	return evidenceprojection.ViewRow(input), nil
}

type PartyRows struct {
	store  *Store
	source partyprojection.SourceReader
}

func NewPartyRowsFromStore(store *Store, source partyprojection.SourceReader) *PartyRows {
	return &PartyRows{
		store:  store,
		source: source,
	}
}

func (r *PartyRows) RefreshPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshPartyTxCore(ctx, tx, recordID, r.source)
}

func (r *PartyRows) LoadPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, partiesViewSchemaID, recordID)
}

var _ partyprojection.Rows = (*PartyRows)(nil)

type TaskDecisionMutationRows struct {
	store             *Store
	taskRequestSource TaskRequestSource
	decisionSource    DecisionSource
}

func NewTaskDecisionMutationRowsFromStore(
	store *Store,
	taskRequestSource TaskRequestSource,
	decisionSource DecisionSource,
) *TaskDecisionMutationRows {
	return &TaskDecisionMutationRows{
		store:             store,
		taskRequestSource: taskRequestSource,
		decisionSource:    decisionSource,
	}
}

func (r *TaskDecisionMutationRows) RefreshTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshTaskRequestTxCore(ctx, tx, recordID, r.taskRequestSource)
}

func (r *TaskDecisionMutationRows) RefreshDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return r.store.refreshDecisionTxCore(ctx, tx, recordID, r.decisionSource)
}

func (r *TaskDecisionMutationRows) LoadTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, taskRequestsViewSchemaID, recordID)
}

func (r *TaskDecisionMutationRows) LoadDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return r.loadTx(ctx, tx, decisionsViewSchemaID, recordID)
}

var _ taskdecisionprojection.MutationRows = (*TaskDecisionMutationRows)(nil)

type taskDerivedFactReader interface {
	CollectTaskDerivedFactsTx(
		context.Context,
		pgx.Tx,
		uuid.UUID,
	) ([]taskdecisionprojection.TaskDerivedFact, error)
}

type decisionDerivedFactReader interface {
	CollectDecisionDerivedFactsTx(
		context.Context,
		pgx.Tx,
		uuid.UUID,
	) ([]taskdecisionprojection.DecisionDerivedFact, error)
}

type TaskDecisionReportingReader struct {
	taskReader     taskDerivedFactReader
	decisionReader decisionDerivedFactReader
}

func NewTaskDecisionReportingReader() *TaskDecisionReportingReader {
	return &TaskDecisionReportingReader{
		taskReader:     queryengine.NewTaskReader(),
		decisionReader: queryengine.NewDecisionReader(),
	}
}

func (r *TaskDecisionReportingReader) CollectTaskDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]taskdecisionprojection.TaskDerivedFact, error) {
	if r == nil || r.taskReader == nil {
		return nil, errors.New("task-request projection reader is required")
	}
	return r.taskReader.CollectTaskDerivedFactsTx(ctx, tx, incidentID)
}

func (r *TaskDecisionReportingReader) CollectDecisionDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]taskdecisionprojection.DecisionDerivedFact, error) {
	if r == nil || r.decisionReader == nil {
		return nil, errors.New("decision projection reader is required")
	}
	return r.decisionReader.CollectDecisionDerivedFactsTx(ctx, tx, incidentID)
}

var _ taskdecisionprojection.ReportingReader = (*TaskDecisionReportingReader)(nil)

func (r *AssessmentRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.store, viewSchemaID, recordID)
}

func (r *ArtifactRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.store, viewSchemaID, recordID)
}

func (r *PartyRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.store, viewSchemaID, recordID)
}

func (r *TaskDecisionMutationRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.store, viewSchemaID, recordID)
}

func loadProviderRowTx(
	ctx context.Context,
	tx pgx.Tx,
	store *Store,
	viewSchemaID string,
	recordID uuid.UUID,
) (map[string]any, error) {
	if store == nil {
		return nil, errors.New("projection store is required")
	}
	return store.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}
