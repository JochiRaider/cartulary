package importassembly

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type assessmentImportRevisionAdapter struct {
	appender *revisions.Appender
}

func (a assessmentImportRevisionAdapter) AppendAssessmentImportRevisionTx(
	ctx context.Context,
	tx pgx.Tx,
	revision assessments.ImportRevision,
) error {
	afterVersion := revision.AfterVersion
	if err := a.appender.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:    revision.ChangeSetID,
		SequenceNo:     revision.SequenceNo,
		TargetKind:     "record",
		TargetID:       revision.RecordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersion,
		AfterValue:     revision.CanonicalRow,
	}); err != nil {
		return err
	}
	return a.appender.AppendRecordRevisionTx(
		ctx,
		tx,
		revisions.AppendRecordRevisionParams{
			ChangeSetID: revision.ChangeSetID,
			RecordID:    revision.RecordID,
			RowVersion:  revision.RowVersion,
			AfterValue:  revision.CanonicalRow,
		},
	)
}

func newAssessmentImportCreateFacade(
	targetViewSchemaID string,
	facadeID string,
	pool postgres.DB,
	projectionCatalog *projections.Catalog,
	appender *revisions.Appender,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	entityStore := hostidentity.NewStore(pool, appender, nil)
	return assessments.NewImportCreateFacade(
		targetViewSchemaID,
		facadeID,
		assessments.ImportCreateDependencies{
			Subjects:    assessmentassembly.NewSubjectValidator(pool, entityStore),
			Assessors:   assessmentassembly.NewAssessorValidator(pool),
			Records:     assessmentassembly.NewRecordEnvelopeCreator(pool),
			Revisions:   assessmentImportRevisionAdapter{appender: appender},
			Projections: assessmentassembly.NewProjectionPort(pool, projectionCatalog),
		},
	)
}
