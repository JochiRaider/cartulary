package assessments

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrMergeProtectedSetChanged = errors.New("assessments: merge protected set changed")

type MergeProtectedSetChangedError struct {
	RecordID uuid.UUID
}

func (e *MergeProtectedSetChangedError) Error() string {
	return ErrMergeProtectedSetChanged.Error()
}

func (e *MergeProtectedSetChangedError) Unwrap() error {
	return ErrMergeProtectedSetChanged
}

type MergeMutation struct {
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
}

type RepointMergedAssessmentsCommand struct {
	IncidentID         uuid.UUID
	SubjectType        string
	SurvivorRecordID   uuid.UUID
	LoserRecordID      uuid.UUID
	ProtectedRecordIDs map[uuid.UUID]struct{}
	Now                time.Time
}

type RepointMergedAssessmentsResult struct {
	Mutations      []MergeMutation
	RepointedCount int
}

type mergeAssessmentRecord struct {
	RecordID        uuid.UUID
	IncidentID      uuid.UUID
	SubjectRecordID uuid.UUID
	SubjectType     string
	AssessmentState string
	ConfidenceScore *int
	Rationale       string
	AssessorUserID  uuid.UUID
	AssessedAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
}

func (s *Store) LoadMergeProtectedRecordIDsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, subjectType string, loserRecordID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := tx.Query(ctx, `
SELECT record_id
  FROM assessments
 WHERE incident_id = $1
   AND subject_type = $2
   AND subject_record_id = $3
   AND deleted_at IS NULL
 ORDER BY record_id ASC
`, incidentID, subjectType, loserRecordID)
	if err != nil {
		return nil, fmt.Errorf("load merge assessment protected set: %w", err)
	}
	defer rows.Close()

	recordIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var recordID uuid.UUID
		if err := rows.Scan(&recordID); err != nil {
			return nil, fmt.Errorf("scan merge assessment protected set: %w", err)
		}
		recordIDs = append(recordIDs, recordID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate merge assessment protected set: %w", err)
	}
	return recordIDs, nil
}

func (s *Store) RepointMergedAssessmentsTx(ctx context.Context, tx pgx.Tx, command RepointMergedAssessmentsCommand) (RepointMergedAssessmentsResult, error) {
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
    created_at,
    updated_at,
    deleted_at,
    deleted_by_user_id
  FROM assessments
 WHERE incident_id = $1
   AND subject_type = $2
   AND subject_record_id = $3
   AND deleted_at IS NULL
 ORDER BY assessed_at ASC, record_id ASC
 FOR UPDATE
`, command.IncidentID, command.SubjectType, command.LoserRecordID)
	if err != nil {
		return RepointMergedAssessmentsResult{}, fmt.Errorf("load merged assessments: %w", err)
	}
	defer rows.Close()

	records := make([]mergeAssessmentRecord, 0)
	for rows.Next() {
		record, err := scanMergeAssessmentRecord(rows)
		if err != nil {
			return RepointMergedAssessmentsResult{}, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return RepointMergedAssessmentsResult{}, fmt.Errorf("iterate merged assessments: %w", err)
	}
	rows.Close()

	result := RepointMergedAssessmentsResult{Mutations: make([]MergeMutation, 0)}
	for _, record := range records {
		if _, ok := command.ProtectedRecordIDs[record.RecordID]; !ok {
			return RepointMergedAssessmentsResult{}, &MergeProtectedSetChangedError{RecordID: record.RecordID}
		}
		before := buildMergeAssessmentValue(record)
		if _, err := tx.Exec(ctx, `
UPDATE assessments
   SET subject_record_id = $2,
       updated_at = $3
 WHERE record_id = $1
`, record.RecordID, command.SurvivorRecordID, command.Now.UTC()); err != nil {
			return RepointMergedAssessmentsResult{}, fmt.Errorf("repoint merged assessment: %w", err)
		}
		record.SubjectRecordID = command.SurvivorRecordID
		record.UpdatedAt = command.Now.UTC()
		if err := s.projectionRows.RefreshTx(ctx, tx, record.RecordID); err != nil {
			return RepointMergedAssessmentsResult{}, err
		}
		result.Mutations = append(result.Mutations, MergeMutation{
			TargetKind:    "assessment",
			TargetID:      record.RecordID.String(),
			OperationKind: "patch",
			BeforeValue:   before,
			AfterValue:    buildMergeAssessmentValue(record),
		})
		result.RepointedCount++
	}
	return result, nil
}

func scanMergeAssessmentRecord(row pgx.Row) (mergeAssessmentRecord, error) {
	var (
		record          mergeAssessmentRecord
		confidence      pgtype.Int4
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := row.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.SubjectRecordID,
		&record.SubjectType,
		&record.AssessmentState,
		&confidence,
		&record.Rationale,
		&record.AssessorUserID,
		&record.AssessedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
		&deletedAt,
		&deletedByUserID,
	); err != nil {
		return mergeAssessmentRecord{}, fmt.Errorf("scan merge assessment record: %w", err)
	}
	if confidence.Valid {
		value := int(confidence.Int32)
		record.ConfidenceScore = &value
	}
	record.AssessedAt = record.AssessedAt.UTC()
	record.CreatedAt = record.CreatedAt.UTC()
	record.UpdatedAt = record.UpdatedAt.UTC()
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	record.DeletedByUserID = uuidPointerFromPG(deletedByUserID)
	return record, nil
}

func buildMergeAssessmentValue(record mergeAssessmentRecord) map[string]any {
	return map[string]any{
		"record_id":          record.RecordID.String(),
		"incident_id":        record.IncidentID.String(),
		"subject_record_id":  record.SubjectRecordID.String(),
		"subject_type":       record.SubjectType,
		"assessment_state":   record.AssessmentState,
		"confidence_score":   record.ConfidenceScore,
		"rationale":          record.Rationale,
		"assessor_user_id":   record.AssessorUserID.String(),
		"assessed_at":        formatTimestamp(record.AssessedAt),
		"deleted_at":         formatTimestampPointer(record.DeletedAt),
		"deleted_by_user_id": formatUUIDPointer(record.DeletedByUserID),
	}
}

func formatTimestampPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func uuidPointerFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.UUID(value.Bytes)
	return &parsed
}
