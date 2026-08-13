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
	return r.store.ApplyTimelineMutationTx(ctx, tx, mutation)
}

type EntityRows struct {
	store          *Store
	source         entityprojection.SourceReader
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

func NewEntityRowsFromStore(store *Store, source entityprojection.SourceReader) *EntityRows {
	return &EntityRows{
		store:          store,
		source:         source,
		hostReader:     queryengine.NewHostReader(store.pool),
		identityReader: queryengine.NewIdentityReader(store.pool),
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
var _ artifactprojection.Rebuilder = (*Store)(nil)
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

func (r *EvidenceRows) RebuildEvidenceTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) error {
	return r.store.rebuildIncidentEvidenceTxCore(ctx, tx, incidentID, r.source)
}

var _ evidenceprojection.Rebuilder = (*Store)(nil)

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

func (r *PartyRows) RebuildPartiesTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) error {
	return r.store.rebuildIncidentPartiesTxCore(ctx, tx, incidentID, r.source)
}

var _ partyprojection.Rows = (*PartyRows)(nil)
var _ partyprojection.Rebuilder = (*Store)(nil)

type TaskDecisionRows struct {
	store             *Store
	taskRequestSource TaskRequestSource
	decisionSource    DecisionSource
	taskReader        taskdecisionprojection.TaskReader
	decisionReader    interface {
		CollectDecisionDerivedFactsTx(context.Context, pgx.Tx, uuid.UUID) ([]taskdecisionprojection.DecisionDerivedFact, error)
	}
}

func NewTaskDecisionRowsFromStore(
	store *Store,
	taskRequestSource TaskRequestSource,
	decisionSource DecisionSource,
) *TaskDecisionRows {
	return &TaskDecisionRows{
		store:             store,
		taskRequestSource: taskRequestSource,
		decisionSource:    decisionSource,
		taskReader:        queryengine.NewTaskReader(),
		decisionReader:    queryengine.NewDecisionReader(),
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
var _ taskdecisionprojection.Rebuilder = (*Store)(nil)
var _ taskdecisionprojection.Reader = (*TaskDecisionRows)(nil)

func (r *AssessmentRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.store, viewSchemaID, recordID)
}

func (r *ArtifactRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.store, viewSchemaID, recordID)
}

func (r *PartyRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return loadProviderRowTx(ctx, tx, r.store, viewSchemaID, recordID)
}

func (r *TaskDecisionRows) loadTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
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
