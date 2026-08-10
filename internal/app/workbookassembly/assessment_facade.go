package workbookassembly

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

const assessmentCreateResultSchemaID = "cartulary.assessments.create_result.v1"

type assessmentIdempotencyAdapter struct {
	store *authn.Store
}

type storedAssessmentCreateResult struct {
	SchemaID     string         `json:"schema_id"`
	RecordID     string         `json:"record_id"`
	ChangeSetID  string         `json:"change_set_id"`
	RowVersion   int64          `json:"row_version"`
	CanonicalRow map[string]any `json:"row"`
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
	result, err := decodeAssessmentCreateResult(record.ResponseJSON)
	if err != nil {
		return assessments.CreateIdempotencyRecord{}, false, err
	}
	return assessments.CreateIdempotencyRecord{
		RequestHash: append([]byte(nil), record.RequestHash...),
		Result:      result,
	}, true, nil
}

func (assessmentIdempotencyAdapter) StoreCreateTx(ctx context.Context, tx pgx.Tx, key assessments.CreateIdempotencyKey, result assessments.CreateResult) error {
	payload := storedAssessmentCreateResult{
		SchemaID:     assessmentCreateResultSchemaID,
		RecordID:     result.RecordID.String(),
		ChangeSetID:  result.ChangeSetID.String(),
		RowVersion:   result.RowVersion,
		CanonicalRow: result.CanonicalRow,
	}
	err := authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey{
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

func decodeAssessmentCreateResult(data []byte) (assessments.CreateResult, error) {
	var stored storedAssessmentCreateResult
	if err := json.Unmarshal(data, &stored); err != nil {
		return assessments.CreateResult{}, fmt.Errorf("decode assessment create result: %w", err)
	}
	if stored.SchemaID == assessmentCreateResultSchemaID {
		recordID, err := uuid.Parse(stored.RecordID)
		if err != nil {
			return assessments.CreateResult{}, fmt.Errorf("decode assessment create record id: %w", err)
		}
		changeSetID, err := uuid.Parse(stored.ChangeSetID)
		if err != nil {
			return assessments.CreateResult{}, fmt.Errorf("decode assessment create change set id: %w", err)
		}
		return assessments.CreateResult{
			Outcome:      assessments.CreateOutcomeCommitted,
			CanonicalRow: stored.CanonicalRow,
			RecordID:     recordID,
			ChangeSetID:  changeSetID,
			RowVersion:   stored.RowVersion,
		}, nil
	}

	var legacy struct {
		ChangeSetID string         `json:"change_set_id"`
		Row         map[string]any `json:"row"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return assessments.CreateResult{}, fmt.Errorf("decode legacy assessment create result: %w", err)
	}
	recordIDText, ok := legacy.Row["record_id"].(string)
	if !ok {
		return assessments.CreateResult{}, errors.New("decode legacy assessment create result: record_id is required")
	}
	recordID, err := uuid.Parse(recordIDText)
	if err != nil {
		return assessments.CreateResult{}, fmt.Errorf("decode legacy assessment create record id: %w", err)
	}
	changeSetID, err := uuid.Parse(legacy.ChangeSetID)
	if err != nil {
		return assessments.CreateResult{}, fmt.Errorf("decode legacy assessment create change set id: %w", err)
	}
	rowVersion, err := assessmentRowVersion(legacy.Row)
	if err != nil {
		return assessments.CreateResult{}, err
	}
	return assessments.CreateResult{
		Outcome:      assessments.CreateOutcomeCommitted,
		CanonicalRow: legacy.Row,
		RecordID:     recordID,
		ChangeSetID:  changeSetID,
		RowVersion:   rowVersion,
	}, nil
}

func assessmentRowVersion(row map[string]any) (int64, error) {
	switch value := row["row_version"].(type) {
	case float64:
		if value > 0 && value == float64(int64(value)) {
			return int64(value), nil
		}
	case int64:
		if value > 0 {
			return value, nil
		}
	}
	return 0, fmt.Errorf("decode assessment create result: invalid row_version %#v", row["row_version"])
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
	if err := a.appender.AppendCapturedRecordMutationTx(ctx, tx, revisions.AppendCapturedRecordMutationParams{
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
	if err := a.appender.AppendCapturedRecordRevisionTx(ctx, tx, revisions.AppendCapturedRecordRevisionParams{
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

func newAssessmentFacade(
	pool postgres.DB,
	projectionRows assessmentprojection.Rows,
	entityStore *hostidentity.Store,
	appender *revisions.Appender,
) (*assessments.Facade, error) {
	if appender == nil {
		return nil, errors.New("compose assessment facade: revision appender is required")
	}
	authStore := authn.NewStore(pool)
	return assessments.NewFacade(pool, assessments.FacadeDependencies{
		Idempotency:    assessmentIdempotencyAdapter{store: authStore},
		Subjects:       assessmentassembly.NewSubjectValidator(pool, entityStore),
		Assessors:      assessmentassembly.NewAssessorValidator(pool),
		SupportTargets: assessmentassembly.NewSupportTargetValidator(pool),
		Records:        assessmentassembly.NewRecordEnvelopeCreator(pool),
		SupportLinks:   assessmentassembly.NewSupportLinkApplier(),
		Revisions:      assessmentRevisionAdapter{appender: appender},
		Projections:    assessmentassembly.NewProjectionPort(projectionRows),
	})
}
