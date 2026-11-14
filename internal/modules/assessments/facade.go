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
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
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
}

type CreateCommand struct {
	ActorUserID uuid.UUID
	IncidentID  uuid.UUID
	Input       CreateInput
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
	// Nil identifies a legacy receipt. New receipts retain the original default
	// independently of an explicitly supplied assessment time and returned row.
	DefaultAssessedAt *time.Time
	Result            CreateResult
}

type RecordEnvelopeCreate struct {
	IncidentID uuid.UUID
	ActorID    uuid.UUID
	Now        time.Time
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
	CanonicalRow  map[string]any
	LinkMutations []SupportLinkMutation
}

type SupportLinkMutation struct {
	RecordLinkID uuid.UUID
	Operation    string
	BeforeValue  map[string]any
	AfterValue   map[string]any
}

type CreateIdempotencyPort interface {
	// AdmitCreateTx serializes the scoped identity before reading its receipt.
	AdmitCreateTx(context.Context, pgx.Tx, CreateIdempotencyKey) (CreateIdempotencyRecord, bool, error)
	StoreCreateTx(context.Context, pgx.Tx, CreateIdempotencyKey, CreateIdempotencyRecord) error
}

type CreateReferences struct {
	SubjectRef  uuid.UUID
	SubjectType string
	SupportRefs []uuid.UUID
}

// ReferenceAdmissionResult describes only facts protected through the caller's
// transaction. An admission error must not return usable partial facts.
type ReferenceAdmissionResult struct {
	SubjectValid bool
	SupportValid bool
}

type ReferenceAdmission interface {
	AdmitAssessmentReferencesTx(context.Context, pgx.Tx, uuid.UUID, CreateReferences) (ReferenceAdmissionResult, error)
}

type AssessorValidator interface {
	ValidateAssessmentAssessorTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (bool, error)
}

type RecordEnvelopeCreator interface {
	CreateAssessmentEnvelopeTx(context.Context, pgx.Tx, RecordEnvelopeCreate) (uuid.UUID, error)
}

type InitialSupportLinkApplier interface {
	ApplyInitialAssessmentSupportLinksTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, []uuid.UUID, time.Time) ([]SupportLinkMutation, error)
}

type CreateRevisionAppender interface {
	AppendAssessmentCreateRevisionTx(context.Context, pgx.Tx, CreateRevision) (uuid.UUID, error)
}

type AssessmentProjectionPort interface {
	RefreshAndLoadAssessmentRowTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
}

type FacadeDependencies struct {
	Idempotency  CreateIdempotencyPort
	References   ReferenceAdmission
	Assessors    AssessorValidator
	Records      RecordEnvelopeCreator
	SupportLinks InitialSupportLinkApplier
	Revisions    CreateRevisionAppender
	Projections  AssessmentProjectionPort
}

type Facade struct {
	pool         dbport.DB
	idempotency  CreateIdempotencyPort
	supportLinks InitialSupportLinkApplier
	revisions    CreateRevisionAppender
	creator      assessmentCreateService
}

func NewFacade(pool dbport.DB, dependencies FacadeDependencies) (*Facade, error) {
	switch {
	case isNilDependency(pool):
		return nil, errors.New("construct assessment facade: database is required")
	case isNilDependency(dependencies.Idempotency):
		return nil, errors.New("construct assessment facade: idempotency port is required")
	case isNilDependency(dependencies.References):
		return nil, errors.New("construct assessment facade: reference admission is required")
	case isNilDependency(dependencies.Assessors):
		return nil, errors.New("construct assessment facade: assessor validator is required")
	case isNilDependency(dependencies.Records):
		return nil, errors.New("construct assessment facade: record-envelope creator is required")
	case isNilDependency(dependencies.SupportLinks):
		return nil, errors.New("construct assessment facade: support-link applier is required")
	case isNilDependency(dependencies.Revisions):
		return nil, errors.New("construct assessment facade: revision appender is required")
	case isNilDependency(dependencies.Projections):
		return nil, errors.New("construct assessment facade: projection port is required")
	}
	return &Facade{
		pool:         pool,
		idempotency:  dependencies.Idempotency,
		supportLinks: dependencies.SupportLinks,
		revisions:    dependencies.Revisions,
		creator: newAssessmentCreateService(
			dependencies.References,
			dependencies.Assessors,
			dependencies.Records,
			dependencies.Projections,
		),
	}, nil
}

func (f *Facade) Create(ctx context.Context, command CreateCommand) (CreateResult, error) {
	if f == nil {
		return CreateResult{}, errors.New("create assessment: facade is required")
	}
	if err := validateFacadeCreateCommand(command); err != nil {
		return CreateResult{}, err
	}
	if err := f.creator.validateInput(command.Input); err != nil {
		return CreateResult{}, err
	}
	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return CreateResult{}, fmt.Errorf("begin assessment create transaction: %w", err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()

	key := createIdempotencyKey(command)
	existing, found, err := f.idempotency.AdmitCreateTx(ctx, tx, key)
	if err != nil {
		return CreateResult{}, fmt.Errorf("query assessment create idempotency: %w", err)
	}
	if found {
		requestHash := legacyCreateRequestHash(command.Input)
		if existing.DefaultAssessedAt != nil {
			requestHash = createRequestHash(command.Input, command.ActorUserID, *existing.DefaultAssessedAt)
		}
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return CreateResult{}, ErrClientTxnConflict
		}
		result := existing.Result
		result.Outcome = CreateOutcomeReplayed
		return result, nil
	}

	input := command.Input
	now := command.Now.UTC().Truncate(time.Microsecond)
	create := assessmentCreateContext{
		IncidentID:  command.IncidentID,
		ActorUserID: command.ActorUserID,
		Input:       input,
		Now:         now,
	}
	if err := f.creator.admitReferencesTx(ctx, tx, create); err != nil {
		return CreateResult{}, err
	}
	supportRefs := uniqueUUIDs(input.SupportRefs)

	assessorID, err := f.creator.resolveAssessorTx(ctx, tx, create)
	if err != nil {
		return CreateResult{}, err
	}

	recordID, err := f.creator.insertTx(ctx, tx, create, assessorID)
	if err != nil {
		return CreateResult{}, err
	}
	linkMutations, err := f.supportLinks.ApplyInitialAssessmentSupportLinksTx(
		ctx,
		tx,
		command.IncidentID,
		recordID,
		command.ActorUserID,
		supportRefs,
		now,
	)
	if err != nil {
		return CreateResult{}, fmt.Errorf("apply assessment support links: %w", err)
	}
	row, err := f.creator.refreshProjectionTx(ctx, tx, recordID)
	if err != nil {
		return CreateResult{}, err
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
		CanonicalRow:  row,
		LinkMutations: linkMutations,
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
	receipt := CreateIdempotencyRecord{
		RequestHash:       createRequestHash(input, command.ActorUserID, now),
		DefaultAssessedAt: &now,
		Result:            result,
	}
	if err := f.idempotency.StoreCreateTx(ctx, tx, key, receipt); err != nil {
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
	switch {
	case command.ActorUserID == uuid.Nil:
		return errors.New("create assessment: actor user id is required")
	case command.IncidentID == uuid.Nil:
		return errors.New("create assessment: incident id is required")
	default:
		return nil
	}
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
	case len(input.SupportRefs) > assessmentpolicy.MaxInitialSupportReferences:
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
