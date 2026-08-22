package workbookassembly

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type assessmentIdempotencyAdapter struct {
	store *authn.Store
}

func (a assessmentIdempotencyAdapter) LookupCreate(ctx context.Context, key assessments.CreateIdempotencyKey) (assessments.CreateIdempotencyRecord, bool, error) {
	record, err := a.store.GetRouteIdempotency(ctx, authn.RouteIdempotencyKey{
		RouteKey:    key.RouteKey,
		ActorUserID: key.ActorUserID,
		ScopeKey:    key.ScopeKey,
		ClientTxnID: key.ClientTxnID,
	})
	if errors.Is(err, authn.ErrNotFound) {
		return assessments.CreateIdempotencyRecord{}, false, nil
	}
	if err != nil {
		return assessments.CreateIdempotencyRecord{}, false, err
	}
	result, err := decodeAssessmentCreateResult(record.ResponseJSON, record.ScopeKey, record.StatusCode, record.RequestHash)
	if err != nil {
		return assessments.CreateIdempotencyRecord{}, false, err
	}
	return assessments.CreateIdempotencyRecord{
		RequestHash: append([]byte(nil), record.RequestHash...),
		Result:      result,
	}, true, nil
}

func (assessmentIdempotencyAdapter) StoreCreateTx(ctx context.Context, tx pgx.Tx, key assessments.CreateIdempotencyKey, result assessments.CreateResult) error {
	payload, err := encodeAssessmentCreateResult(key.ScopeKey, key.RequestHash, result)
	if err != nil {
		return err
	}
	err = authn.InsertRouteIdempotency(ctx, tx, authn.RouteIdempotencyKey{
		RouteKey:    key.RouteKey,
		ActorUserID: key.ActorUserID,
		ScopeKey:    key.ScopeKey,
		ClientTxnID: key.ClientTxnID,
	}, nil, key.RequestHash, http.StatusCreated, payload)
	if authn.IsUniqueViolation(err) {
		return assessments.ErrClientTxnConflict
	}
	return err
}

type assessmentRevisionAdapter struct {
	appender *revisions.Appender
}

func (a assessmentRevisionAdapter) AppendAssessmentCreateRevisionTx(ctx context.Context, tx pgx.Tx, create assessments.CreateRevision) (uuid.UUID, error) {
	clientTxnID := create.ClientTxnID
	requestID := create.RequestID
	changeSetID, err := a.appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  create.IncidentID,
		ActorUserID: create.ActorUserID,
		Source:      create.RouteKey,
		ClientTxnID: &clientTxnID,
		RequestID:   &requestID,
		CreatedAt:   create.CreatedAt,
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	afterSnapshot, err := a.appender.CaptureRecordSnapshotTx(ctx, tx, create.RecordID)
	if err != nil {
		return uuid.UUID{}, err
	}
	afterVersion := create.AfterVersion
	if err := a.appender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     create.TargetKind,
		RecordID:       create.RecordID,
		OperationKind:  create.OperationKind,
		AfterVersionID: &afterVersion,
		AfterSnapshot:  &afterSnapshot,
	}); err != nil {
		return uuid.UUID{}, err
	}
	for index, mutation := range create.LinkMutations {
		if err := a.appender.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    index + 2,
			TargetKind:    "record_link",
			TargetID:      mutation.RecordLinkID.String(),
			OperationKind: mutation.Operation,
			BeforeValue:   mutation.BeforeValue,
			AfterValue:    mutation.AfterValue,
		}); err != nil {
			return uuid.UUID{}, err
		}
	}
	if err := a.appender.AppendRecordRevisionAndIntentTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID:   changeSetID,
		RecordID:      create.RecordID,
		RowVersion:    create.RowVersion,
		AfterSnapshot: &afterSnapshot,
		LiveChange: revisions.LiveRecordChange{
			AfterValue: create.CanonicalRow,
		},
	}); err != nil {
		return uuid.UUID{}, err
	}
	return changeSetID, nil
}

func NewAssessmentMutationContribution(
	pool postgres.DB,
	projectionRows assessmentprojection.Rows,
	entitySourceFacts *hostidentity.SourceFacts,
	appender *revisions.Appender,
) (*assessments.Facade, error) {
	if appender == nil {
		return nil, errors.New("compose assessment facade: revision appender is required")
	}
	authStore := authn.NewStore(pool)
	return assessments.NewFacade(pool, assessments.FacadeDependencies{
		Idempotency:    assessmentIdempotencyAdapter{store: authStore},
		Subjects:       assessmentassembly.NewSubjectValidator(pool, entitySourceFacts),
		Assessors:      assessmentassembly.NewAssessorValidator(pool),
		SupportTargets: assessmentassembly.NewSupportTargetValidator(pool),
		Records:        assessmentassembly.NewRecordEnvelopeCreator(pool),
		SupportLinks:   assessmentassembly.NewSupportLinkApplier(),
		Revisions:      assessmentRevisionAdapter{appender: appender},
		Projections:    assessmentassembly.NewProjectionPort(projectionRows),
	})
}
