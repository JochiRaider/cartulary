package workbook

import (
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/linkednotes"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
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
	pool              postgres.DB
	authStore         *authn.Store
	recordTargets     *records.RouteTargetResolver
	timelineStore     *timeline.Facade
	artifactMutations *artifacts.WorkbookFacade
	linkedNoteStore   *linkednotes.Facade
	evidenceMutations *evidence.WorkbookFacade
	entityStore       *hostidentity.Store
	indicatorStore    *indicators.Store
	partyMutations    *parties.WorkbookFacade
	taskMutations     *tasksdecisions.WorkbookFacade
	supersedeStore    *tasksdecisions.SupersedeFacade
	projectionRows    workbookProjectionQueryPort
	conflictTokens    conflicttokens.ConflictTokenCodec
}

func NewStore(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec, projectionQuery *projections.QueryService) *Store {
	return newStoreWithTimelineFacade(pool, nil, conflictTokens, projectionQuery)
}

func newStoreWithTimelineFacade(pool postgres.DB, timelineStore *timeline.Facade, conflictTokens conflicttokens.ConflictTokenCodec, projectionQuery *projections.QueryService) *Store {
	if projectionQuery == nil {
		panic("workbook projection query is required")
	}
	return &Store{
		pool:              pool,
		authStore:         authn.NewStore(pool),
		recordTargets:     records.NewRouteTargetResolver(pool),
		timelineStore:     timelineStore,
		artifactMutations: artifacts.NewWorkbookFacade(pool, conflictTokens),
		linkedNoteStore:   linkednotes.NewFacade(pool),
		evidenceMutations: evidence.NewWorkbookFacade(pool, conflictTokens),
		entityStore:       hostidentity.NewStore(pool),
		indicatorStore:    indicators.NewStore(pool),
		partyMutations:    parties.NewWorkbookFacade(pool, conflictTokens),
		taskMutations:     tasksdecisions.NewWorkbookFacade(pool, conflictTokens),
		supersedeStore:    tasksdecisions.NewSupersedeFacade(pool),
		projectionRows:    projectionQuery,
		conflictTokens:    conflictTokens,
	}
}
