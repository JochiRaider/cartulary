package tasksdecisions

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type MutationFacade struct {
	pool             postgres.DB
	idempotency      IdempotencyCapability
	keepSaved        conflicttokens.IdempotencyPort
	incidentAccess   IncidentStateCapability
	memberReferences MemberReferenceCapability
	recordStore      RecordEnvelopeCapability
	linkStore        LinkCapability
	projectionRows   taskdecisionprojection.Rows
	revisions        RevisionCapability
	conflictTokens   conflicttokens.ConflictTokenCodec
	conflictFields   conflicttokens.FieldResolver
	publications     collaboration.RecordChangedAppender
	catalog          *sourcecatalog.Catalog
}

func NewMutationContribution(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	dependencies MutationDependencies,
) (*MutationFacade, error) {
	if pool == nil {
		return nil, fmt.Errorf("tasks/decisions mutation composition: Postgres is required")
	}
	if err := dependencies.validate(); err != nil {
		return nil, err
	}
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return nil, fmt.Errorf("compose Tasks/Decisions source catalog: %w", err)
	}
	return &MutationFacade{
		pool:             pool,
		idempotency:      dependencies.Idempotency,
		keepSaved:        dependencies.KeepSavedIdempotency,
		incidentAccess:   dependencies.IncidentState,
		memberReferences: dependencies.MemberReferences,
		recordStore:      dependencies.RecordEnvelopes,
		linkStore:        dependencies.Links,
		projectionRows:   dependencies.Projections,
		revisions:        dependencies.Revisions,
		conflictTokens:   conflictTokens,
		conflictFields:   dependencies.ConflictFields,
		publications:     dependencies.Collaboration,
		catalog:          catalog,
	}, nil
}
