package assessmentassembly

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type subjectValidator struct {
	entities *hostidentity.Store
	records  *records.Store
}

func NewSubjectValidator(
	pool postgres.DB,
	entities *hostidentity.Store,
) assessments.SubjectValidator {
	return subjectValidator{
		entities: entities,
		records:  records.NewStore(pool),
	}
}

func (a subjectValidator) ValidateAssessmentSubjectTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	subjectType string,
	recordID uuid.UUID,
) (bool, error) {
	envelope, err := a.records.LoadEnvelopeTx(ctx, tx, recordID, false)
	if errors.Is(err, records.ErrEnvelopeNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if envelope.IncidentID != incidentID ||
		envelope.RecordType != subjectType ||
		envelope.DeletedAt != nil {
		return false, nil
	}
	err = a.entities.ValidateResolvedTargetTx(
		ctx,
		tx,
		incidentID,
		subjectType,
		recordID,
	)
	if errors.Is(err, hostidentity.ErrHostIdentityRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

type assessorValidator struct {
	auth *authn.Store
}

func NewAssessorValidator(pool postgres.DB) assessments.AssessorValidator {
	return assessorValidator{auth: authn.NewStore(pool)}
}

func (a assessorValidator) ValidateAssessmentAssessorTx(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
) (bool, error) {
	user, err := a.auth.GetUserByIDForUpdateTx(ctx, tx, userID)
	if errors.Is(err, authn.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return user.IsActive, nil
}

type supportTargetValidator struct {
	records *records.Store
}

func NewSupportTargetValidator(
	pool postgres.DB,
) assessments.SupportTargetValidator {
	return supportTargetValidator{records: records.NewStore(pool)}
}

func (a supportTargetValidator) ValidateAssessmentSupportTargetsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordIDs []uuid.UUID,
) (bool, error) {
	envelopes, err := a.records.LoadEnvelopesTx(ctx, tx, recordIDs, false)
	if err != nil {
		return false, err
	}
	if len(envelopes) != len(recordIDs) {
		return false, nil
	}
	for _, recordID := range recordIDs {
		envelope := envelopes[recordID]
		if envelope.IncidentID != incidentID || envelope.DeletedAt != nil {
			return false, nil
		}
	}
	return true, nil
}

type recordEnvelopeCreator struct {
	records *records.Store
}

func NewRecordEnvelopeCreator(
	pool postgres.DB,
) assessments.RecordEnvelopeCreator {
	return recordEnvelopeCreator{records: records.NewStore(pool)}
}

func (a recordEnvelopeCreator) CreateAssessmentEnvelopeTx(
	ctx context.Context,
	tx pgx.Tx,
	create assessments.RecordEnvelopeCreate,
) (uuid.UUID, error) {
	return a.records.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      create.IncidentID,
		RecordType:      create.RecordType,
		CreatedByUserID: create.ActorID,
		CreatedAt:       create.Now,
		UpdatedByUserID: create.ActorID,
		UpdatedAt:       create.Now,
		RowVersion:      create.RowVersion,
	})
}

type supportLinkApplier struct {
	links *links.Store
}

func NewSupportLinkApplier() assessments.InitialSupportLinkApplier {
	return supportLinkApplier{links: links.NewStore()}
}

func (a supportLinkApplier) ApplyInitialAssessmentSupportLinksTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	assessmentID uuid.UUID,
	actorUserID uuid.UUID,
	supportRecordIDs []uuid.UUID,
	now time.Time,
) error {
	for _, supportRecordID := range supportRecordIDs {
		if _, _, err := a.links.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
			IncidentID:  incidentID,
			SrcRecordID: assessmentID,
			DstRecordID: supportRecordID,
			LinkType:    links.LinkType(links.LinkTypeSupportedBy),
			Provenance:  links.LinkProvenance(links.LinkProvenanceManual),
			OwnerUserID: actorUserID,
			Now:         now,
		}); err != nil {
			return err
		}
	}
	return nil
}

type projectionPort struct {
	coordinator *projections.Coordinator
}

func NewProjectionPort(
	pool postgres.DB,
	catalog *projections.Catalog,
) assessments.AssessmentProjectionPort {
	return projectionPort{
		coordinator: projections.NewCoordinator(pool, catalog),
	}
}

func (a projectionPort) RefreshAndLoadAssessmentRowTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (map[string]any, error) {
	if err := a.coordinator.RefreshRowTx(
		ctx,
		tx,
		assessments.AssessmentsViewSchemaID,
		recordID,
	); err != nil {
		return nil, err
	}
	return a.coordinator.LoadRowTx(
		ctx,
		tx,
		assessments.AssessmentsViewSchemaID,
		recordID,
	)
}

type mergeProjectionPort struct {
	coordinator *projections.Coordinator
}

func NewMergeEffects(
	pool postgres.DB,
	catalog *projections.Catalog,
) *assessments.MergeEffects {
	return assessments.NewMergeEffects(mergeProjectionPort{
		coordinator: projections.NewCoordinator(pool, catalog),
	})
}

func (a mergeProjectionPort) RefreshAssessmentProjectionTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) error {
	return a.coordinator.RefreshRowTx(
		ctx,
		tx,
		assessments.AssessmentsViewSchemaID,
		recordID,
	)
}
