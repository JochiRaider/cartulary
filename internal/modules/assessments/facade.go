package assessments

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	assessmentpolicy "github.com/JochiRaider/cartulary/internal/modules/assessments/internal/policy"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type CreateOutcome string

const (
	CreateOutcomeCommitted CreateOutcome = "committed"
	CreateOutcomeReplayed  CreateOutcome = "replayed"
)

var ErrClientTxnConflict = errors.New("assessments: client transaction conflict")

type CreateInput struct {
	ClientTxnID     string
	SubjectRef      uuid.UUID
	SubjectType     string
	AssessmentState string
	ConfidenceScore *int
	Rationale       string
	Assessor        *uuid.UUID
	AssessedAt      *time.Time
	SupportRefs     []uuid.UUID
}

type CreateIdempotencyKey struct {
	RouteKey    string
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
	RequestHash []byte
}

type CreateCommand struct {
	ActorUserID uuid.UUID
	IncidentID  uuid.UUID
	Input       CreateInput
	Idempotency CreateIdempotencyKey
	RequestID   string
	Now         time.Time
}

type CreateResult struct {
	Outcome      CreateOutcome
	CanonicalRow map[string]any
	RecordID     uuid.UUID
	ChangeSetID  uuid.UUID
	RowVersion   int64
}

type CreateIdempotencyRecord struct {
	RequestHash []byte
	Result      CreateResult
}

type RecordEnvelopeCreate struct {
	IncidentID uuid.UUID
	RecordType string
	ActorID    uuid.UUID
	Now        time.Time
	RowVersion int64
}

type CreateRevision struct {
	IncidentID    uuid.UUID
	ActorUserID   uuid.UUID
	RouteKey      string
	ClientTxnID   string
	RequestID     string
	CreatedAt     time.Time
	RecordID      uuid.UUID
	RowVersion    int64
	AfterVersion  string
	CanonicalRow  map[string]any
	TargetKind    string
	OperationKind string
}

type CreateIdempotencyPort interface {
	LookupCreate(context.Context, CreateIdempotencyKey) (CreateIdempotencyRecord, bool, error)
	StoreCreateTx(context.Context, pgx.Tx, CreateIdempotencyKey, CreateResult) error
}

type SubjectValidator interface {
	ValidateAssessmentSubjectTx(context.Context, pgx.Tx, uuid.UUID, string, uuid.UUID) (bool, error)
}

type AssessorValidator interface {
	ValidateAssessmentAssessorTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (bool, error)
}

type SupportTargetValidator interface {
	ValidateAssessmentSupportTargetsTx(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) (bool, error)
}

type RecordEnvelopeCreator interface {
	CreateAssessmentEnvelopeTx(context.Context, pgx.Tx, RecordEnvelopeCreate) (uuid.UUID, error)
}

type InitialSupportLinkApplier interface {
	ApplyInitialAssessmentSupportLinksTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, []uuid.UUID, time.Time) error
}

type CreateRevisionAppender interface {
	AppendAssessmentCreateRevisionTx(context.Context, pgx.Tx, CreateRevision) (uuid.UUID, error)
}

type AssessmentProjectionPort interface {
	RefreshAndLoadAssessmentRowTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
}

type FacadeDependencies struct {
	Idempotency    CreateIdempotencyPort
	Subjects       SubjectValidator
	Assessors      AssessorValidator
	SupportTargets SupportTargetValidator
	Records        RecordEnvelopeCreator
	SupportLinks   InitialSupportLinkApplier
	Revisions      CreateRevisionAppender
	Projections    AssessmentProjectionPort
}

type Facade struct {
	pool         postgres.DB
	source       assessmentSourceRepository
	idempotency  CreateIdempotencyPort
	subjects     SubjectValidator
	assessors    AssessorValidator
	support      SupportTargetValidator
	records      RecordEnvelopeCreator
	supportLinks InitialSupportLinkApplier
	revisions    CreateRevisionAppender
	projections  AssessmentProjectionPort
}

func NewFacade(pool postgres.DB, dependencies FacadeDependencies) (*Facade, error) {
	switch {
	case pool == nil:
		return nil, errors.New("construct assessment facade: database is required")
	case dependencies.Idempotency == nil:
		return nil, errors.New("construct assessment facade: idempotency port is required")
	case dependencies.Subjects == nil:
		return nil, errors.New("construct assessment facade: subject validator is required")
	case dependencies.Assessors == nil:
		return nil, errors.New("construct assessment facade: assessor validator is required")
	case dependencies.SupportTargets == nil:
		return nil, errors.New("construct assessment facade: support-target validator is required")
	case dependencies.Records == nil:
		return nil, errors.New("construct assessment facade: record-envelope creator is required")
	case dependencies.SupportLinks == nil:
		return nil, errors.New("construct assessment facade: support-link applier is required")
	case dependencies.Revisions == nil:
		return nil, errors.New("construct assessment facade: revision appender is required")
	case dependencies.Projections == nil:
		return nil, errors.New("construct assessment facade: projection port is required")
	}
	return &Facade{
		pool:         pool,
		source:       assessmentSourceRepository{},
		idempotency:  dependencies.Idempotency,
		subjects:     dependencies.Subjects,
		assessors:    dependencies.Assessors,
		support:      dependencies.SupportTargets,
		records:      dependencies.Records,
		supportLinks: dependencies.SupportLinks,
		revisions:    dependencies.Revisions,
		projections:  dependencies.Projections,
	}, nil
}

func (f *Facade) Create(ctx context.Context, command CreateCommand) (CreateResult, error) {
	if f == nil {
		return CreateResult{}, errors.New("create assessment: facade is required")
	}
	if err := validateFacadeCreateCommand(command); err != nil {
		return CreateResult{}, err
	}
	key := command.Idempotency
	existing, found, err := f.idempotency.LookupCreate(ctx, key)
	if err != nil {
		return CreateResult{}, fmt.Errorf("query assessment create idempotency: %w", err)
	}
	if found {
		if !bytes.Equal(existing.RequestHash, key.RequestHash) {
			return CreateResult{}, ErrClientTxnConflict
		}
		result := existing.Result
		result.Outcome = CreateOutcomeReplayed
		return result, nil
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CreateResult{}, fmt.Errorf("begin assessment create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	input := command.Input
	valid, err := f.subjects.ValidateAssessmentSubjectTx(ctx, tx, command.IncidentID, input.SubjectType, input.SubjectRef)
	if err != nil {
		return CreateResult{}, fmt.Errorf("validate assessment subject: %w", err)
	}
	if !valid {
		return CreateResult{}, &CreateValidationError{Field: "assessment.subject_ref", ReasonCode: "invalid_value"}
	}

	supportRefs := uniqueUUIDs(input.SupportRefs)
	valid, err = f.support.ValidateAssessmentSupportTargetsTx(ctx, tx, command.IncidentID, supportRefs)
	if err != nil {
		return CreateResult{}, fmt.Errorf("validate assessment support targets: %w", err)
	}
	if !valid {
		return CreateResult{}, &CreateValidationError{Field: "assessment.support_refs", ReasonCode: "invalid_value"}
	}

	assessorID := command.ActorUserID
	if input.Assessor != nil {
		assessorID = *input.Assessor
	}
	valid, err = f.assessors.ValidateAssessmentAssessorTx(ctx, tx, command.IncidentID, assessorID)
	if err != nil {
		return CreateResult{}, fmt.Errorf("validate assessment assessor: %w", err)
	}
	if !valid {
		return CreateResult{}, &CreateValidationError{Field: "assessment.assessor", ReasonCode: "invalid_value"}
	}

	now := command.Now.UTC()
	assessedAt := now
	if input.AssessedAt != nil {
		assessedAt = input.AssessedAt.UTC()
	}
	recordID, err := f.records.CreateAssessmentEnvelopeTx(ctx, tx, RecordEnvelopeCreate{
		IncidentID: command.IncidentID,
		RecordType: "assessment",
		ActorID:    command.ActorUserID,
		Now:        now,
		RowVersion: 1,
	})
	if err != nil {
		return CreateResult{}, fmt.Errorf("create assessment record envelope: %w", err)
	}
	if err := f.source.InsertTx(ctx, tx, assessmentSourceCreate{
		RecordID:        recordID,
		IncidentID:      command.IncidentID,
		SubjectRef:      input.SubjectRef,
		SubjectType:     input.SubjectType,
		AssessmentState: input.AssessmentState,
		ConfidenceScore: input.ConfidenceScore,
		Rationale:       input.Rationale,
		Assessor:        assessorID,
		AssessedAt:      assessedAt,
		Now:             now,
	}); err != nil {
		return CreateResult{}, err
	}
	if err := f.supportLinks.ApplyInitialAssessmentSupportLinksTx(
		ctx,
		tx,
		command.IncidentID,
		recordID,
		command.ActorUserID,
		supportRefs,
		now,
	); err != nil {
		return CreateResult{}, fmt.Errorf("apply assessment support links: %w", err)
	}
	row, err := f.projections.RefreshAndLoadAssessmentRowTx(ctx, tx, recordID)
	if err != nil {
		return CreateResult{}, fmt.Errorf("refresh assessment projection: %w", err)
	}
	rowVersion, err := canonicalRowVersion(row)
	if err != nil {
		return CreateResult{}, err
	}
	changeSetID, err := f.revisions.AppendAssessmentCreateRevisionTx(ctx, tx, CreateRevision{
		IncidentID:    command.IncidentID,
		ActorUserID:   command.ActorUserID,
		RouteKey:      key.RouteKey,
		ClientTxnID:   input.ClientTxnID,
		RequestID:     command.RequestID,
		CreatedAt:     now,
		RecordID:      recordID,
		RowVersion:    rowVersion,
		AfterVersion:  assessmentVersionID(recordID, rowVersion),
		CanonicalRow:  row,
		TargetKind:    "assessment",
		OperationKind: "create",
	})
	if err != nil {
		return CreateResult{}, fmt.Errorf("append assessment create revision: %w", err)
	}

	result := CreateResult{
		Outcome:      CreateOutcomeCommitted,
		CanonicalRow: row,
		RecordID:     recordID,
		ChangeSetID:  changeSetID,
		RowVersion:   rowVersion,
	}
	if err := f.idempotency.StoreCreateTx(ctx, tx, key, result); err != nil {
		if errors.Is(err, ErrClientTxnConflict) {
			return CreateResult{}, ErrClientTxnConflict
		}
		return CreateResult{}, fmt.Errorf("store assessment create idempotency: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateResult{}, fmt.Errorf("commit assessment create transaction: %w", err)
	}
	return result, nil
}

func validateFacadeCreateCommand(command CreateCommand) error {
	input := command.Input
	switch {
	case command.ActorUserID == uuid.Nil:
		return errors.New("create assessment: actor user id is required")
	case command.IncidentID == uuid.Nil:
		return errors.New("create assessment: incident id is required")
	case strings.TrimSpace(command.Idempotency.RouteKey) == "":
		return errors.New("create assessment: idempotency route key is required")
	case strings.TrimSpace(command.Idempotency.ScopeKey) == "":
		return errors.New("create assessment: idempotency scope key is required")
	case command.Idempotency.ActorUserID != command.ActorUserID:
		return errors.New("create assessment: idempotency actor does not match command actor")
	case command.Idempotency.ClientTxnID != input.ClientTxnID:
		return errors.New("create assessment: idempotency client transaction does not match input")
	case len(command.Idempotency.RequestHash) == 0:
		return errors.New("create assessment: request hash is required")
	}
	return validateCreateInputShape(input)
}

func validateCreateInputShape(input CreateInput) error {
	switch {
	case strings.TrimSpace(input.ClientTxnID) == "":
		return &CreateValidationError{Field: "client_txn_id", ReasonCode: "missing_required_field"}
	case input.SubjectRef == uuid.Nil:
		return &CreateValidationError{Field: "assessment.subject_ref", ReasonCode: "missing_required_field"}
	case !assessmentpolicy.ValidSubjectType(input.SubjectType):
		return &CreateValidationError{Field: "assessment.subject_type", ReasonCode: "invalid_value"}
	case !assessmentpolicy.ValidState(input.AssessmentState):
		return &CreateValidationError{Field: "assessment.assessment_state", ReasonCode: "invalid_value"}
	case input.Rationale == "":
		return &CreateValidationError{Field: "assessment.rationale", ReasonCode: "missing_required_field"}
	case !validConfidenceScore(input.ConfidenceScore):
		return &CreateValidationError{Field: "assessment.confidence_score", ReasonCode: "invalid_value"}
	case len(input.SupportRefs) > maxSupportActions:
		return &CreateValidationError{Field: "assessment.support_refs", ReasonCode: "invalid_value"}
	default:
		return nil
	}
}

func canonicalRowVersion(row map[string]any) (int64, error) {
	value, ok := row["row_version"]
	if !ok {
		return 0, errors.New("load assessment canonical row: row_version is required")
	}
	switch typed := value.(type) {
	case int64:
		if typed > 0 {
			return typed, nil
		}
	}
	return 0, fmt.Errorf("load assessment canonical row: invalid row_version %#v", value)
}

func validConfidenceScore(value *int) bool {
	_, valid := assessmentpolicy.ConfidenceBand(value)
	return valid
}

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	unique := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}

func assessmentVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("assessment:%s:%d", recordID.String(), rowVersion)
}
