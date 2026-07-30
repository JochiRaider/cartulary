package projectionprovider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProjectionInput struct {
	RecordID            uuid.UUID
	IncidentID          uuid.UUID
	RowVersion          int64
	SubjectRef          uuid.UUID
	SubjectType         string
	AssessmentState     string
	ConfidenceScore     *int
	ConfidenceBand      string
	Rationale           string
	Assessor            uuid.UUID
	AssessedAt          time.Time
	SupportingLinkCount int
}

type ProjectionMutationKind string

const (
	ProjectionMutationUpsert ProjectionMutationKind = "upsert"
	ProjectionMutationDelete ProjectionMutationKind = "delete"
)

type ProjectionMutation struct {
	Kind     ProjectionMutationKind
	RecordID uuid.UUID
	Input    ProjectionInput
}

func (mutation ProjectionMutation) Validate() error {
	if mutation.RecordID == uuid.Nil {
		return errors.New("assessment projection mutation record_id is required")
	}
	switch mutation.Kind {
	case ProjectionMutationUpsert:
		if mutation.Input.RecordID != mutation.RecordID {
			return errors.New("assessment projection upsert record_id does not match input")
		}
		return validateProjectionInput(mutation.Input)
	case ProjectionMutationDelete:
		if !isZeroProjectionInput(mutation.Input) {
			return errors.New("assessment projection delete must not carry an input")
		}
		return nil
	default:
		return fmt.Errorf("unsupported assessment projection mutation kind %q", mutation.Kind)
	}
}

type Envelope struct {
	RecordID   uuid.UUID
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
	DeletedAt  *time.Time
}

type SupportFacts struct {
	ActiveTargetCount int
}

type EnvelopeReader interface {
	LoadAssessmentProjectionEnvelopeTx(context.Context, pgx.Tx, uuid.UUID) (Envelope, bool, error)
}

type SupportFactReader interface {
	LoadAssessmentProjectionSupportFactsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (SupportFacts, error)
}

type Source struct {
	envelopes EnvelopeReader
	support   SupportFactReader
}

func NewSource(envelopes EnvelopeReader, support SupportFactReader) *Source {
	return &Source{envelopes: envelopes, support: support}
}

func (s *Source) BuildProjectionMutationTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (ProjectionMutation, error) {
	if recordID == uuid.Nil {
		return ProjectionMutation{}, errors.New("assessment projection mutation record_id is required")
	}
	if s == nil || s.envelopes == nil || s.support == nil {
		return ProjectionMutation{}, errors.New("assessment projection source dependencies are required")
	}
	snapshot, found, err := loadSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return ProjectionMutation{}, err
	}
	if !found || snapshot.DeletedAt != nil {
		return ProjectionMutation{Kind: ProjectionMutationDelete, RecordID: recordID}, nil
	}
	input, found, err := s.projectionInputTx(ctx, tx, snapshot)
	if err != nil {
		return ProjectionMutation{}, err
	}
	if !found {
		return ProjectionMutation{Kind: ProjectionMutationDelete, RecordID: recordID}, nil
	}
	return ProjectionMutation{
		Kind:     ProjectionMutationUpsert,
		RecordID: recordID,
		Input:    input,
	}, nil
}

type ProjectionInputPage struct {
	Inputs       []ProjectionInput
	NextRecordID *uuid.UUID
}

func (s *Source) ListProjectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) (ProjectionInputPage, error) {
	if incidentID == uuid.Nil {
		return ProjectionInputPage{}, errors.New("assessment projection enumeration incident_id is required")
	}
	if limit <= 0 || limit > 1000 {
		return ProjectionInputPage{}, fmt.Errorf(
			"assessment projection enumeration limit %d is outside 1..1000",
			limit,
		)
	}
	if s == nil || s.envelopes == nil || s.support == nil {
		return ProjectionInputPage{}, errors.New("assessment projection source dependencies are required")
	}
	snapshots, err := listSnapshotPageTx(ctx, tx, incidentID, afterRecordID, limit+1)
	if err != nil {
		return ProjectionInputPage{}, err
	}
	hasMore := len(snapshots) > limit
	if hasMore {
		snapshots = snapshots[:limit]
	}
	inputs := make([]ProjectionInput, 0, len(snapshots))
	for _, snapshot := range snapshots {
		input, found, err := s.projectionInputTx(ctx, tx, snapshot)
		if err != nil {
			return ProjectionInputPage{}, err
		}
		if found {
			inputs = append(inputs, input)
		}
	}
	var nextRecordID *uuid.UUID
	if hasMore {
		next := snapshots[len(snapshots)-1].RecordID
		nextRecordID = &next
	}
	return ProjectionInputPage{Inputs: inputs, NextRecordID: nextRecordID}, nil
}

type sourceSnapshot struct {
	RecordID        uuid.UUID
	IncidentID      uuid.UUID
	SubjectRef      uuid.UUID
	SubjectType     string
	AssessmentState string
	ConfidenceScore *int
	Rationale       string
	Assessor        uuid.UUID
	AssessedAt      time.Time
	DeletedAt       *time.Time
}

func (s *Source) projectionInputTx(
	ctx context.Context,
	tx pgx.Tx,
	snapshot sourceSnapshot,
) (ProjectionInput, bool, error) {
	envelope, found, err := s.envelopes.LoadAssessmentProjectionEnvelopeTx(
		ctx,
		tx,
		snapshot.RecordID,
	)
	if err != nil {
		return ProjectionInput{}, false, err
	}
	if !found ||
		envelope.DeletedAt != nil ||
		envelope.RecordType != "assessment" ||
		envelope.IncidentID != snapshot.IncidentID {
		return ProjectionInput{}, false, nil
	}
	facts, err := s.support.LoadAssessmentProjectionSupportFactsTx(
		ctx,
		tx,
		snapshot.IncidentID,
		snapshot.RecordID,
	)
	if err != nil {
		return ProjectionInput{}, false, err
	}
	input := ProjectionInput{
		RecordID:            snapshot.RecordID,
		IncidentID:          snapshot.IncidentID,
		RowVersion:          envelope.RowVersion,
		SubjectRef:          snapshot.SubjectRef,
		SubjectType:         snapshot.SubjectType,
		AssessmentState:     snapshot.AssessmentState,
		ConfidenceScore:     cloneInt(snapshot.ConfidenceScore),
		ConfidenceBand:      deriveConfidenceBand(snapshot.ConfidenceScore),
		Rationale:           snapshot.Rationale,
		Assessor:            snapshot.Assessor,
		AssessedAt:          snapshot.AssessedAt.UTC(),
		SupportingLinkCount: facts.ActiveTargetCount,
	}
	if err := validateProjectionInput(input); err != nil {
		return ProjectionInput{}, false, err
	}
	return input, true, nil
}

func loadSnapshotTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (sourceSnapshot, bool, error) {
	row := tx.QueryRow(ctx, `
SELECT
    record_id,
    incident_id,
    subject_record_id,
    subject_type,
    assessment_state,
    confidence_score,
    rationale,
    assessor_user_id,
    assessed_at,
    deleted_at
  FROM assessments
 WHERE record_id = $1
`, recordID)
	snapshot, err := scanSnapshot(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return sourceSnapshot{}, false, nil
	}
	if err != nil {
		return sourceSnapshot{}, false, fmt.Errorf("load assessment projection source: %w", err)
	}
	return snapshot, true, nil
}

func listSnapshotPageTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) ([]sourceSnapshot, error) {
	rows, err := tx.Query(ctx, `
SELECT
    record_id,
    incident_id,
    subject_record_id,
    subject_type,
    assessment_state,
    confidence_score,
    rationale,
    assessor_user_id,
    assessed_at,
    deleted_at
  FROM assessments
 WHERE incident_id = $1
   AND deleted_at IS NULL
   AND ($2::uuid IS NULL OR record_id > $2)
 ORDER BY record_id
 LIMIT $3
`, incidentID, afterRecordID, limit)
	if err != nil {
		return nil, fmt.Errorf("list assessment projection sources: %w", err)
	}
	defer rows.Close()
	snapshots := make([]sourceSnapshot, 0, limit)
	for rows.Next() {
		snapshot, err := scanSnapshot(rows)
		if err != nil {
			return nil, fmt.Errorf("scan assessment projection source: %w", err)
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assessment projection sources: %w", err)
	}
	return snapshots, nil
}

func scanSnapshot(row pgx.Row) (sourceSnapshot, error) {
	var (
		snapshot        sourceSnapshot
		confidenceScore pgtype.Int4
	)
	if err := row.Scan(
		&snapshot.RecordID,
		&snapshot.IncidentID,
		&snapshot.SubjectRef,
		&snapshot.SubjectType,
		&snapshot.AssessmentState,
		&confidenceScore,
		&snapshot.Rationale,
		&snapshot.Assessor,
		&snapshot.AssessedAt,
		&snapshot.DeletedAt,
	); err != nil {
		return sourceSnapshot{}, err
	}
	if confidenceScore.Valid {
		value := int(confidenceScore.Int32)
		snapshot.ConfidenceScore = &value
	}
	snapshot.AssessedAt = snapshot.AssessedAt.UTC()
	if snapshot.DeletedAt != nil {
		deletedAt := snapshot.DeletedAt.UTC()
		snapshot.DeletedAt = &deletedAt
	}
	return snapshot, nil
}

func validateProjectionInput(input ProjectionInput) error {
	switch {
	case input.RecordID == uuid.Nil:
		return errors.New("assessment projection input record_id is required")
	case input.IncidentID == uuid.Nil:
		return errors.New("assessment projection input incident_id is required")
	case input.RowVersion <= 0:
		return errors.New("assessment projection input row_version must be positive")
	case input.SubjectRef == uuid.Nil:
		return errors.New("assessment projection input subject_ref is required")
	case input.SubjectType != "host" && input.SubjectType != "identity":
		return fmt.Errorf("assessment projection input subject_type %q is invalid", input.SubjectType)
	case input.Assessor == uuid.Nil:
		return errors.New("assessment projection input assessor is required")
	case input.AssessedAt.IsZero():
		return errors.New("assessment projection input assessed_at is required")
	case input.SupportingLinkCount < 0:
		return errors.New("assessment projection input supporting_link_count must be non-negative")
	}
	switch input.AssessmentState {
	case "unknown", "suspected", "confirmed", "disproven", "cleared":
	default:
		return fmt.Errorf(
			"assessment projection input assessment_state %q is invalid",
			input.AssessmentState,
		)
	}
	if input.ConfidenceBand != deriveConfidenceBand(input.ConfidenceScore) {
		return errors.New("assessment projection input confidence_band is not canonical")
	}
	return nil
}

func isZeroProjectionInput(input ProjectionInput) bool {
	return input.RecordID == uuid.Nil &&
		input.IncidentID == uuid.Nil &&
		input.RowVersion == 0 &&
		input.SubjectRef == uuid.Nil &&
		input.SubjectType == "" &&
		input.AssessmentState == "" &&
		input.ConfidenceScore == nil &&
		input.ConfidenceBand == "" &&
		input.Rationale == "" &&
		input.Assessor == uuid.Nil &&
		input.AssessedAt.IsZero() &&
		input.SupportingLinkCount == 0
}

func deriveConfidenceBand(score *int) string {
	switch {
	case score == nil:
		return "unset"
	case *score >= 0 && *score <= 39:
		return "low"
	case *score >= 40 && *score <= 69:
		return "medium"
	case *score >= 70 && *score <= 100:
		return "high"
	default:
		return ""
	}
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
