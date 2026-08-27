package workbookassembly

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
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
	appender     *revisions.Appender
	publications collaboration.RecordChangedAppender
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
	afterVersion := fmt.Sprintf("assessment:%s:%d", create.RecordID.String(), create.RowVersion)
	if err := a.appender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "assessment",
		RecordID:       create.RecordID,
		OperationKind:  "create",
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
	fieldKeys := assessmentRowFieldKeys(create.CanonicalRow)
	if err := a.appender.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
		ChangeSetID:   changeSetID,
		RecordID:      create.RecordID,
		RowVersion:    create.RowVersion,
		AfterSnapshot: &afterSnapshot,
		ConflictFacts: assessmentRevisionFacts(create.CanonicalRow, fieldKeys),
	}); err != nil {
		return uuid.UUID{}, err
	}
	patch := collabprotocol.BuildViewRowPatch(create.CanonicalRow, fieldKeys)
	changeKind := "invalidate"
	if patch != nil {
		changeKind = "patch"
	}
	if err := a.publications.AppendRecordChangedTx(ctx, tx, collaboration.RecordChangeIntentInput{
		IncidentID: create.IncidentID, RecordID: create.RecordID, ChangeSetID: changeSetID,
		ActorUserID: create.ActorUserID, RowVersion: create.RowVersion, ClientTxnID: create.ClientTxnID,
		CreatedAt: create.CreatedAt, PublicFieldKeys: fieldKeys,
		AffectedViews: []collaboration.AffectedViewChange{{ViewSchemaID: assessments.AssessmentsViewSchemaID, RecordID: create.RecordID, RowVersion: create.RowVersion, ChangeKind: changeKind, PatchCells: patch}},
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
	publications collaboration.RecordChangedAppender,
) (*assessments.Facade, error) {
	if isNilDependency(appender) {
		return nil, errors.New("compose assessment facade: revision appender is required")
	}
	if isNilDependency(publications) {
		return nil, errors.New("compose assessment facade: collaboration publication appender is required")
	}
	subjects, err := assessmentassembly.NewSubjectValidator(pool, entitySourceFacts)
	if err != nil {
		return nil, fmt.Errorf("compose assessment facade: %w", err)
	}
	assessors, err := assessmentassembly.NewAssessorValidator(pool)
	if err != nil {
		return nil, fmt.Errorf("compose assessment facade: %w", err)
	}
	supportTargets, err := assessmentassembly.NewSupportTargetValidator(pool)
	if err != nil {
		return nil, fmt.Errorf("compose assessment facade: %w", err)
	}
	recordEnvelopes, err := assessmentassembly.NewRecordEnvelopeCreator(pool)
	if err != nil {
		return nil, fmt.Errorf("compose assessment facade: %w", err)
	}
	projections, err := assessmentassembly.NewProjectionPort(projectionRows)
	if err != nil {
		return nil, fmt.Errorf("compose assessment facade: %w", err)
	}
	authStore := authn.NewStore(pool)
	return assessments.NewFacade(pool, assessments.FacadeDependencies{
		Idempotency:    assessmentIdempotencyAdapter{store: authStore},
		Subjects:       subjects,
		Assessors:      assessors,
		SupportTargets: supportTargets,
		Records:        recordEnvelopes,
		SupportLinks:   assessmentassembly.NewSupportLinkApplier(),
		Revisions:      assessmentRevisionAdapter{appender: appender, publications: publications},
		Projections:    projections,
	})
}

func assessmentRowFieldKeys(row map[string]any) []string {
	cells, _ := row["cells"].(map[string]any)
	keys := make([]string, 0, len(cells))
	for key := range cells {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func assessmentRevisionFacts(row map[string]any, fieldKeys []string) []revisions.RevisionConflictFact {
	cells, _ := row["cells"].(map[string]any)
	facts := make([]revisions.RevisionConflictFact, 0, len(fieldKeys))
	for _, key := range fieldKeys {
		value, present := cells[key]
		facts = append(facts, revisions.RevisionConflictFact{FieldKey: key, AfterPresent: present, AfterValue: value})
	}
	return facts
}
