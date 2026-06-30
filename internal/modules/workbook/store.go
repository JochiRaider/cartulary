package workbook

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
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
	recordStore     *records.Store
	revisionStore   *revisions.Store
	timelineStore   *timeline.Facade
	artifactStore   *artifacts.Store
	evidenceStore   *evidence.Store
	entityStore     *entities.Store
	linkStore       *links.Store
	taskStore       *tasksdecisions.Store
	projectionStore *projections.Store
	conflictTokens  revisions.ConflictTokenCodec
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
		recordStore:     records.NewStore(),
		revisionStore:   revisions.NewStore(),
		timelineStore:   timelineStore,
		artifactStore:   artifacts.NewStore(),
		evidenceStore:   evidence.NewStore(pool),
		entityStore:     entities.NewStore(pool),
		linkStore:       links.NewStore(),
		taskStore:       tasksdecisions.NewStore(),
		projectionStore: projections.NewStore(pool),
		conflictTokens:  revisions.NewConflictTokenCodecForTesting("workbook"),
	}
}

func (s *Store) SetConflictTokenCodec(codec revisions.ConflictTokenCodec) {
	s.conflictTokens = codec
	if s.timelineStore != nil {
		s.timelineStore.SetConflictTokenCodec(codec)
	}
}

func (s *Store) QueryRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta) ([]map[string]any, error) {
	switch viewSchemaID {
	case timeline.TimelineViewSchemaID:
		return s.timelineStore.QueryTimelineRows(ctx, incidentID, query)
	case entities.HostsViewSchemaID:
		return s.entityStore.QueryHostRows(ctx, incidentID, query)
	case entities.IdentitiesViewSchemaID:
		return s.entityStore.QueryIdentityRows(ctx, incidentID, query)
	case entities.IndicatorsViewSchemaID:
		return s.entityStore.QueryIndicatorRows(ctx, incidentID, query)
	default:
		if !projections.SupportsQuerySurface(viewSchemaID) {
			return nil, fmt.Errorf("workbook query surface %q not mapped", viewSchemaID)
		}
		return s.projectionStore.QueryRows(ctx, incidentID, viewSchemaID, query)
	}
}
