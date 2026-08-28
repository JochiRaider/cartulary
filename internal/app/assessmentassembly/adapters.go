package assessmentassembly

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type subjectValidator struct {
	entities *hostidentity.SourceFacts
	records  *records.Store
}

func NewSubjectValidator(
	pool postgres.DB,
	entities *hostidentity.SourceFacts,
) (assessments.SubjectValidator, error) {
	if isNilDependency(pool) {
		return nil, errors.New("compose assessment subject validator: database is required")
	}
	if isNilDependency(entities) {
		return nil, errors.New("compose assessment subject validator: entity source facts are required")
	}
	return subjectValidator{
		entities: entities,
		records:  records.NewStore(pool),
	}, nil
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
	auth      *authn.Store
	incidents *admission.Checker
}

func NewAssessorValidator(pool postgres.DB) (assessments.AssessorValidator, error) {
	if isNilDependency(pool) {
		return nil, errors.New("compose assessment assessor validator: database is required")
	}
	return assessorValidator{
		auth:      authn.NewStore(pool),
		incidents: admission.NewChecker(pool),
	}, nil
}

func (a assessorValidator) ValidateAssessmentAssessorTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	userID uuid.UUID,
) (bool, error) {
	if _, err := a.incidents.CheckTx(ctx, tx, incidentID, userID, admission.Requirement{
		AllowedRoles: admission.RolesMember,
		Lifecycle:    admission.LifecycleAny,
	}); admission.IsDenied(err, admission.DenialNotVisible) {
		return false, nil
	} else if err != nil {
		return false, err
	}
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
) (assessments.SupportTargetValidator, error) {
	if isNilDependency(pool) {
		return nil, errors.New("compose assessment support-target validator: database is required")
	}
	return supportTargetValidator{records: records.NewStore(pool)}, nil
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
) (assessments.RecordEnvelopeCreator, error) {
	if isNilDependency(pool) {
		return nil, errors.New("compose assessment record-envelope creator: database is required")
	}
	return recordEnvelopeCreator{records: records.NewStore(pool)}, nil
}

func (a recordEnvelopeCreator) CreateAssessmentEnvelopeTx(
	ctx context.Context,
	tx pgx.Tx,
	create assessments.RecordEnvelopeCreate,
) (uuid.UUID, error) {
	return a.records.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      create.IncidentID,
		RecordType:      "assessment",
		CreatedByUserID: create.ActorID,
		CreatedAt:       create.Now,
		UpdatedByUserID: create.ActorID,
		UpdatedAt:       create.Now,
		RowVersion:      1,
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
) ([]assessments.SupportLinkMutation, error) {
	mutations := make([]assessments.SupportLinkMutation, 0, len(supportRecordIDs))
	for _, supportRecordID := range supportRecordIDs {
		result, err := a.links.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
			IncidentID:  incidentID,
			SrcRecordID: assessmentID,
			DstRecordID: supportRecordID,
			LinkType:    links.LinkTypeSupportedBy,
			Provenance:  links.LinkProvenanceManual,
			OwnerUserID: actorUserID,
			Now:         now,
		})
		if err != nil {
			return nil, err
		}
		if mutation, ok := result.Mutation(); ok {
			beforeValue, _ := mutation.BeforeValue().(map[string]any)
			afterValue, _ := mutation.AfterValue().(map[string]any)
			recordLinkID, err := uuid.Parse(mutation.TargetID())
			if err != nil {
				return nil, err
			}
			mutations = append(mutations, assessments.SupportLinkMutation{
				RecordLinkID: recordLinkID,
				Operation:    mutation.OperationKind(),
				BeforeValue:  cloneMutationMap(beforeValue),
				AfterValue:   cloneMutationMap(afterValue),
			})
		}
	}
	return mutations, nil
}

func cloneMutationMap(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	cloned := make(map[string]any, len(value))
	for key, item := range value {
		cloned[key] = item
	}
	return cloned
}

type projectionPort struct {
	rows assessmentprojection.Rows
}

func NewProjectionPort(
	rows assessmentprojection.Rows,
) (assessments.AssessmentProjectionPort, error) {
	if isNilDependency(rows) {
		return nil, errors.New("compose assessment projection port: rows are required")
	}
	return projectionPort{rows: rows}, nil
}

func (a projectionPort) RefreshAndLoadAssessmentRowTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (map[string]any, error) {
	if err := a.rows.RefreshAssessmentTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return a.rows.LoadAssessmentTx(ctx, tx, recordID)
}

type mergeProjectionPort struct {
	rows assessmentprojection.Rows
}

func NewMergeEffects(rows assessmentprojection.Rows, snapshots assessments.MergeSnapshotCapturePort) (*assessments.MergeEffects, error) {
	if isNilDependency(rows) {
		return nil, errors.New("compose assessment merge effects: projection rows are required")
	}
	if isNilDependency(snapshots) {
		return nil, errors.New("compose assessment merge effects: snapshot capture port is required")
	}
	return assessments.NewMergeEffects(mergeProjectionPort{rows: rows}, snapshots)
}

func (a mergeProjectionPort) RefreshAssessmentProjectionTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) error {
	return a.rows.RefreshAssessmentTx(ctx, tx, recordID)
}
