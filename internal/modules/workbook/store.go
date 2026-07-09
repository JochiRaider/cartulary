package workbook

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/linkednotes"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/workbook/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	AssessmentsViewSchemaID          = "cartulary.view.assessments.v1"
	CommLogViewSchemaID              = artifacts.CommLogViewSchemaID
	DecisionsViewSchemaID            = "cartulary.view.decisions.v1"
	EvidenceViewSchemaID             = "cartulary.view.evidence.v1"
	FindingsViewSchemaID             = artifacts.FindingsViewSchemaID
	ForensicKeywordsViewSchemaID     = artifacts.ForensicKeywordsViewSchemaID
	HandoffViewSchemaID              = artifacts.HandoffViewSchemaID
	InvestigativeQueriesViewSchemaID = artifacts.InvestigativeQueriesViewSchemaID
	LessonViewSchemaID               = artifacts.LessonViewSchemaID
	NotesViewSchemaID                = artifacts.NotesViewSchemaID
	PartiesViewSchemaID              = "cartulary.view.parties.v1"
	StatusReviewViewSchemaID         = artifacts.StatusReviewViewSchemaID
	TaskRequestsViewSchemaID         = "cartulary.view.task_requests.v1"
)

type Store struct {
	pool            postgres.DB
	authStore       *authn.Store
	incidentAccess  incidents.Access
	recordStore     *records.Store
	revisionStore   *revisions.Store
	timelineStore   *timeline.Facade
	artifactStore   *artifacts.Store
	linkedNoteStore *linkednotes.Facade
	evidenceStore   *evidence.Store
	entityStore     *hostidentity.Store
	indicatorStore  *indicators.Store
	partyStore      *parties.Store
	linkStore       workbookLinkPort
	taskStore       *tasksdecisions.Store
	supersedeStore  *tasksdecisions.SupersedeFacade
	projectionRows  *projectionadapters.WorkbookRows
	rowProjector    *projectionadapters.RowProjector
	conflictTokens  conflicts.ConflictTokenCodec
}

type workbookLinkPort interface {
	SyncFieldReferenceCommandTx(context.Context, pgx.Tx, links.SyncFieldReferenceCommand) (bool, error)
	ValidatePartyRefCollectionTx(context.Context, pgx.Tx, links.PartyRefCollectionValidation) error
	ValidateRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionValidation) error
	ValidateTagCollectionTx(context.Context, pgx.Tx, links.TagCollectionValidation) error
	ApplyPartyRefCollectionTx(context.Context, pgx.Tx, links.PartyRefCollectionCommand) (bool, error)
	ApplyRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionCommand) (bool, error)
	ApplyTagCollectionTx(context.Context, pgx.Tx, links.TagCollectionCommand) (bool, error)
}

func NewStore(pool postgres.DB) *Store {
	return newStoreWithTimelineFacade(pool, nil)
}

func newStoreWithTimelineFacade(pool postgres.DB, timelineStore *timeline.Facade) *Store {
	if timelineStore == nil {
		timelineStore = timeline.NewFacade(pool)
	}
	return &Store{
		pool:            pool,
		authStore:       authn.NewStore(pool),
		incidentAccess:  incidents.NewAccess(pool),
		recordStore:     records.NewStore(),
		revisionStore:   revisions.NewStore(),
		timelineStore:   timelineStore,
		artifactStore:   artifacts.NewStore(),
		linkedNoteStore: linkednotes.NewFacade(pool),
		evidenceStore:   evidence.NewStore(pool),
		entityStore:     hostidentity.NewStore(pool),
		indicatorStore:  indicators.NewStore(pool),
		partyStore:      parties.NewStore(pool),
		linkStore:       links.NewStore(),
		taskStore:       tasksdecisions.NewStore(),
		supersedeStore:  tasksdecisions.NewSupersedeFacade(pool),
		projectionRows:  projectionadapters.NewWorkbookRows(pool),
		rowProjector:    projectionadapters.NewRowProjector(pool),
		conflictTokens:  conflicts.NewConflictTokenCodecForTesting("workbook"),
	}
}

func (s *Store) SetConflictTokenCodec(codec conflicts.ConflictTokenCodec) {
	s.conflictTokens = codec
	if s.timelineStore != nil {
		s.timelineStore.SetConflictTokenCodec(codec)
	}
}

func (s *Store) QueryRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta) ([]map[string]any, error) {
	switch viewSchemaID {
	case timeline.TimelineViewSchemaID:
		return s.timelineStore.QueryTimelineRows(ctx, incidentID, query)
	case hostidentity.HostsViewSchemaID:
		return s.entityStore.QueryHostRows(ctx, incidentID, query)
	case hostidentity.IdentitiesViewSchemaID:
		return s.entityStore.QueryIdentityRows(ctx, incidentID, query)
	case indicators.ViewSchemaID:
		return s.indicatorStore.QueryRows(ctx, incidentID, query)
	default:
		if !s.projectionRows.Supports(viewSchemaID) {
			return nil, fmt.Errorf("workbook query surface %q not mapped", viewSchemaID)
		}
		return s.projectionRows.QueryRows(ctx, incidentID, viewSchemaID, query)
	}
}
