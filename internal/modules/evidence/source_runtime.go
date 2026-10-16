package evidence

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// SourceRuntime composes transaction-bound Evidence capabilities independently
// of HTTP, object-store and background-worker lifetimes.
type SourceRuntime struct {
	workbook     MutationContribution
	attachments  TimelineAttachmentContribution
	importCreate ownerfacade.ImportOwnerCreateFacade
}

type SourceRuntimeDependencies struct {
	Postgres       postgres.DB
	ConflictTokens *conflicttokens.ConflictTokenCodec
	Revisions      *revisions.Appender
	Collaboration  collaboration.RecordChangedAppender
	Mutation       MutationDependencies
}

func NewSourceRuntime(dependencies SourceRuntimeDependencies) (*SourceRuntime, error) {
	if isNilMutationCapability(dependencies.Postgres) || dependencies.ConflictTokens == nil || dependencies.Revisions == nil || isNilMutationCapability(dependencies.Collaboration) {
		return nil, fmt.Errorf("compose Evidence source runtime: source dependencies are required")
	}
	if err := dependencies.Mutation.validate(); err != nil {
		return nil, err
	}
	mutations, err := newSourceMutationService(dependencies.Postgres, dependencies.Mutation.ProjectionRows, dependencies.Revisions, dependencies.Collaboration, dependencies.Mutation.IncidentState, dependencies.Mutation.RecordEnvelopes)
	if err != nil {
		return nil, err
	}
	workbook, err := newMutationFacade(dependencies.Postgres, *dependencies.ConflictTokens, mutations, dependencies.Mutation)
	if err != nil {
		return nil, err
	}
	imports, err := newImportCreateFacade("evidence.import_create", mutations)
	if err != nil {
		return nil, err
	}
	return &SourceRuntime{workbook: workbook, attachments: timelineAttachmentReader{projections: dependencies.Mutation.ProjectionRows}, importCreate: imports}, nil
}

func (runtime *SourceRuntime) MutationContribution() MutationContribution { return runtime.workbook }
func (runtime *SourceRuntime) TimelineAttachmentContribution() TimelineAttachmentContribution {
	return runtime.attachments
}
func (runtime *SourceRuntime) ImportCreateFacade() ownerfacade.ImportOwnerCreateFacade {
	return runtime.importCreate
}
