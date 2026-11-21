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
	inputs, err := s.projectionInputsTx(ctx, tx, snapshot.IncidentID, []sourceSnapshot{snapshot})
	if err != nil {
		return workbookprojection.ProjectionMutation{}, err
	}
	if len(inputs) == 0 {
		return workbookprojection.ProjectionMutation{Kind: workbookprojection.ProjectionMutationDelete, RecordID: recordID}, nil
	}
	return workbookprojection.ProjectionMutation{
		Kind:     workbookprojection.ProjectionMutationUpsert,
		RecordID: recordID,
		Input:    inputs[0],
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
	inputs, err := s.projectionInputsTx(ctx, tx, incidentID, snapshots)
	if err != nil {
		return workbookprojection.ProjectionInputPage{}, err
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

// projectionInputsTx derives the rows of one incident's snapshots whose
// envelope is a live Assessment in that incident, with one envelope read and
// one support-fact read for the whole set.
func (s *Source) projectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	snapshots []sourceSnapshot,
) ([]workbookprojection.ProjectionInput, error) {
	recordIDs := make([]uuid.UUID, 0, len(snapshots))
	for _, snapshot := range snapshots {
		recordIDs = append(recordIDs, snapshot.RecordID)
	}
	envelopes, err := s.envelopes.LoadAssessmentProjectionEnvelopesTx(ctx, tx, recordIDs)
	if err != nil {
		return nil, err
	}
	live := make([]sourceSnapshot, 0, len(snapshots))
	liveIDs := make([]uuid.UUID, 0, len(snapshots))
	for _, snapshot := range snapshots {
		envelope, found := envelopes[snapshot.RecordID]
		if !found ||
			envelope.DeletedAt != nil ||
			envelope.RecordType != "assessment" ||
			envelope.IncidentID != snapshot.IncidentID {
			continue
		}
		live = append(live, snapshot)
		liveIDs = append(liveIDs, snapshot.RecordID)
	}
	if len(live) == 0 {
		return []workbookprojection.ProjectionInput{}, nil
	}
	facts, err := s.support.LoadAssessmentProjectionSupportFactsTx(ctx, tx, incidentID, liveIDs)
	if err != nil {
		return nil, err
	}
	inputs := make([]workbookprojection.ProjectionInput, 0, len(live))
	for _, snapshot := range live {
		confidenceBand, _ := assessmentpolicy.ConfidenceBand(snapshot.ConfidenceScore)
		input := workbookprojection.ProjectionInput{
			RecordID:            snapshot.RecordID,
			IncidentID:          snapshot.IncidentID,
			RowVersion:          envelopes[snapshot.RecordID].RowVersion,
			SubjectRef:          snapshot.SubjectRef,
			SubjectType:         snapshot.SubjectType,
			AssessmentState:     snapshot.AssessmentState,
			ConfidenceScore:     cloneInt(snapshot.ConfidenceScore),
			ConfidenceBand:      confidenceBand,
			Rationale:           snapshot.Rationale,
			Assessor:            snapshot.Assessor,
			AssessedAt:          snapshot.AssessedAt.UTC(),
			SupportingLinkCount: facts[snapshot.RecordID].ActiveTargetCount,
		}
		if err := input.Validate(); err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}
	return inputs, nil
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
