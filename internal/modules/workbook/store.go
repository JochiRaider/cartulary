package workbook

import (
	"context"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/linkednotes"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
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
	recordTargets   workbookRecordTargetPort
	linkedNoteStore workbookLinkedNotePort
	supersedeStore  workbookSupersedePort
	contributions   *WorkbookContributionCatalog
}

type StoreDependencies struct {
	RecordTargets       workbookRecordTargetPort
	LinkedNoteOwner     workbookLinkedNotePort
	SupersedeOwner      workbookSupersedePort
	ContributionCatalog *WorkbookContributionCatalog
}

func NewStore(dependencies StoreDependencies) *Store {
	if dependencies.RecordTargets == nil {
		panic("workbook record target owner is required")
	}
	if dependencies.LinkedNoteOwner == nil {
		panic("workbook linked-note owner is required")
	}
	if dependencies.SupersedeOwner == nil {
		panic("workbook supersede owner is required")
	}
	if dependencies.ContributionCatalog == nil {
		panic("workbook contribution catalog is required")
	}
	return &Store{
		recordTargets:   dependencies.RecordTargets,
		linkedNoteStore: dependencies.LinkedNoteOwner,
		supersedeStore:  dependencies.SupersedeOwner,
		contributions:   dependencies.ContributionCatalog,
	}
}

type workbookLinkedNotePort interface {
	Create(
		ctx context.Context,
		command linkednotes.CreateCommand,
	) (linkednotes.MutationResult, error)
	SourceIncident(ctx context.Context, sourceRecordID uuid.UUID) (uuid.UUID, error)
}

type workbookSupersedePort interface {
	SupersedeDecision(
		ctx context.Context,
		command tasksdecisions.SupersedeCommand,
	) (tasksdecisions.SupersedeMutationResult, error)
}
