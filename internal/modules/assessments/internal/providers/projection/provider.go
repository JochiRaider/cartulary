package projection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	assessmentpolicy "github.com/JochiRaider/cartulary/internal/modules/assessments/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
)

type Source struct {
	envelopes workbookprojection.EnvelopeReader
	support   workbookprojection.SupportFactReader
}

func NewSource(
	envelopes workbookprojection.EnvelopeReader,
	support workbookprojection.SupportFactReader,
) *Source {
	return &Source{envelopes: envelopes, support: support}
}

func (s *Source) BuildProjectionMutationTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (workbookprojection.ProjectionMutation, error) {
	if recordID == uuid.Nil {
		return workbookprojection.ProjectionMutation{}, errors.New("assessment projection mutation record_id is required")
	}
	if s == nil || s.envelopes == nil || s.support == nil {
		return workbookprojection.ProjectionMutation{}, errors.New("assessment projection source dependencies are required")
	}
	snapshot, found, err := loadSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return workbookprojection.ProjectionMutation{}, err
	}
	if !found || snapshot.DeletedAt != nil {
		return workbookprojection.ProjectionMutation{Kind: workbookprojection.ProjectionMutationDelete, RecordID: recordID}, nil
	}
	input, found, err := s.projectionInputTx(ctx, tx, snapshot)
	if err != nil {
		return workbookprojection.ProjectionMutation{}, err
	}
	if !found {
		return workbookprojection.ProjectionMutation{Kind: workbookprojection.ProjectionMutationDelete, RecordID: recordID}, nil
	}
	return workbookprojection.ProjectionMutation{
		Kind:     workbookprojection.ProjectionMutationUpsert,
		RecordID: recordID,
		Input:    input,
	}, nil
}

func (s *Source) ListProjectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) (workbookprojection.ProjectionInputPage, error) {
	if incidentID == uuid.Nil {
		return workbookprojection.ProjectionInputPage{}, errors.New("assessment projection enumeration incident_id is required")
	}
	if limit <= 0 || limit > 1000 {
		return workbookprojection.ProjectionInputPage{}, fmt.Errorf(
			"assessment projection enumeration limit %d is outside 1..1000",
			limit,
		)
	}
	if s == nil || s.envelopes == nil || s.support == nil {
		return workbookprojection.ProjectionInputPage{}, errors.New("assessment projection source dependencies are required")
	}
	snapshots, err := listSnapshotPageTx(ctx, tx, incidentID, afterRecordID, limit+1)
	if err != nil {
		return workbookprojection.ProjectionInputPage{}, err
	}
	hasMore := len(snapshots) > limit
	if hasMore {
		snapshots = snapshots[:limit]
	}
	inputs := make([]workbookprojection.ProjectionInput, 0, len(snapshots))
	for _, snapshot := range snapshots {
		input, found, err := s.projectionInputTx(ctx, tx, snapshot)
		if err != nil {
			return workbookprojection.ProjectionInputPage{}, err
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
	return workbookprojection.ProjectionInputPage{Inputs: inputs, NextRecordID: nextRecordID}, nil
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
) (workbookprojection.ProjectionInput, bool, error) {
	envelope, found, err := s.envelopes.LoadAssessmentProjectionEnvelopeTx(
		ctx,
		tx,
		snapshot.RecordID,
	)
	if err != nil {
		return workbookprojection.ProjectionInput{}, false, err
	}
	if !found ||
		envelope.DeletedAt != nil ||
		envelope.RecordType != "assessment" ||
		envelope.IncidentID != snapshot.IncidentID {
		return workbookprojection.ProjectionInput{}, false, nil
	}
	facts, err := s.support.LoadAssessmentProjectionSupportFactsTx(
		ctx,
		tx,
		snapshot.IncidentID,
		snapshot.RecordID,
	)
	if err != nil {
		return workbookprojection.ProjectionInput{}, false, err
	}
	confidenceBand, _ := assessmentpolicy.ConfidenceBand(snapshot.ConfidenceScore)
	input := workbookprojection.ProjectionInput{
		RecordID:            snapshot.RecordID,
		IncidentID:          snapshot.IncidentID,
		RowVersion:          envelope.RowVersion,
		SubjectRef:          snapshot.SubjectRef,
		SubjectType:         snapshot.SubjectType,
		AssessmentState:     snapshot.AssessmentState,
		ConfidenceScore:     cloneInt(snapshot.ConfidenceScore),
		ConfidenceBand:      confidenceBand,
		Rationale:           snapshot.Rationale,
		Assessor:            snapshot.Assessor,
		AssessedAt:          snapshot.AssessedAt.UTC(),
		SupportingLinkCount: facts.ActiveTargetCount,
	}
	if err := input.Validate(); err != nil {
		return workbookprojection.ProjectionInput{}, false, err
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

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
