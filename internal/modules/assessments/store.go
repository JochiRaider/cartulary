package assessments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool           postgres.DB
	authStore      *authn.Store
	recordStore    *records.Store
	revisionsStore *revisions.Store
	rowProjector   *projectionadapters.RowProjector
	linkStore      assessmentLinkPort
}

type assessmentLinkPort interface {
	UpsertLinkCommandTx(context.Context, pgx.Tx, links.UpsertLinkCommand) (links.RecordLink, bool, error)
}

func NewStore(pool postgres.DB) *Store {
	return &Store{
		pool:           pool,
		authStore:      authn.NewStore(pool),
		recordStore:    records.NewStore(),
		revisionsStore: revisions.NewStore(),
		rowProjector:   projectionadapters.NewRowProjector(pool),
		linkStore:      links.NewStore(),
	}
}

func (s *Store) CreateAssessmentRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + AssessmentsViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    assessmentCreateRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed assessment create payload: %w", err)
		}
		recordID, err := extractUUIDFromPayload(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query assessment create idempotency: %w", err)
	}

	if err := validateCreateRequestShape(request); err != nil {
		return MutationResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin assessment create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := validateSubjectTx(ctx, tx, incidentID, *request.SubjectRef, request.SubjectType); err != nil {
		return MutationResult{}, err
	}
	if err := validateSupportRefsTx(ctx, tx, incidentID, request.SupportRefs); err != nil {
		return MutationResult{}, err
	}
	if request.Assessor != nil {
		if err := validateAssessorTx(ctx, tx, *request.Assessor); err != nil {
			return MutationResult{}, err
		}
	}

	assessedAt := now.UTC()
	if request.AssessedAt != nil {
		assessedAt = request.AssessedAt.UTC()
	}
	assessor := actor.ID
	if request.Assessor != nil {
		assessor = *request.Assessor
	}

	recordID, err := s.recordStore.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      incidentID,
		RecordType:      "assessment",
		CreatedByUserID: actor.ID,
		CreatedAt:       now.UTC(),
		UpdatedByUserID: actor.ID,
		UpdatedAt:       now.UTC(),
		RowVersion:      1,
	})
	if err != nil {
		return MutationResult{}, err
	}

	if err := insertAssessmentSourceTx(ctx, tx, recordID, incidentID, request, assessor, assessedAt, now.UTC()); err != nil {
		return MutationResult{}, err
	}

	for _, supportRef := range uniqueUUIDs(request.SupportRefs) {
		if _, _, err := s.linkStore.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
			IncidentID:  incidentID,
			SrcRecordID: recordID,
			DstRecordID: supportRef,
			LinkType:    links.LinkType(links.LinkTypeSupportedBy),
			Provenance:  links.LinkProvenance(links.LinkProvenanceManual),
			OwnerUserID: actor.ID,
			Now:         now.UTC(),
		}); err != nil {
			return MutationResult{}, err
		}
	}

	if err := s.rowProjector.RefreshRowTx(ctx, tx, projectionadapters.AssessmentsViewSchemaID, recordID); err != nil {
		return MutationResult{}, err
	}
	projected, err := loadProjectionRecordTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}

	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      assessmentCreateRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}

	afterRow := BuildAssessmentRow(projected)
	afterVersionID := assessmentVersionID(recordID, projected.RowVersion)
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "assessment",
		TargetID:       recordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		BeforeValue:    nil,
		AfterValue:     afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  projected.RowVersion,
		BeforeValue: nil,
		AfterValue:  afterRow,
	}); err != nil {
		return MutationResult{}, err
	}

	payload := BuildMutationPayload(changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit assessment create transaction: %w", err)
	}

	return MutationResult{
		Payload:     payload,
		StatusCode:  http.StatusCreated,
		RecordID:    recordID,
		ChangeSetID: changeSetID,
		RowVersion:  projected.RowVersion,
	}, nil
}

func insertAssessmentSourceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, request CreateRequest, assessor uuid.UUID, assessedAt time.Time, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO assessments (
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
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
`, recordID, incidentID, *request.SubjectRef, request.SubjectType, request.AssessmentState, request.ConfidenceScore, request.Rationale, assessor, assessedAt.UTC(), now.UTC())
	if err != nil {
		return fmt.Errorf("insert assessment: %w", err)
	}
	return nil
}

func validateCreateRequestShape(request CreateRequest) error {
	if request.SubjectRef == nil {
		return &CreateValidationError{Field: "assessment.subject_ref", ReasonCode: "missing_required_field"}
	}
	if request.SubjectType != "host" && request.SubjectType != "identity" {
		return &CreateValidationError{Field: "assessment.subject_type", ReasonCode: "invalid_value"}
	}
	if !validAssessmentState(request.AssessmentState) {
		return &CreateValidationError{Field: "assessment.assessment_state", ReasonCode: "invalid_value"}
	}
	if request.Rationale == "" {
		return &CreateValidationError{Field: "assessment.rationale", ReasonCode: "missing_required_field"}
	}
	if request.ConfidenceScore != nil && (*request.ConfidenceScore < 0 || *request.ConfidenceScore > 100) {
		return &CreateValidationError{Field: "assessment.confidence_score", ReasonCode: "invalid_value"}
	}
	return nil
}

func validAssessmentState(value string) bool {
	switch value {
	case "unknown", "suspected", "confirmed", "disproven", "cleared":
		return true
	default:
		return false
	}
}

func validateSubjectTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, subjectRef uuid.UUID, subjectType string) error {
	var exists bool
	var err error
	switch subjectType {
	case "host":
		err = tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM hosts h
      JOIN records r
        ON r.incident_id = h.incident_id
       AND r.record_id = h.record_id
     WHERE h.incident_id = $1
       AND h.record_id = $2
       AND h.host_state IN ('stub', 'canonical')
       AND r.deleted_at IS NULL
)
`, incidentID, subjectRef).Scan(&exists)
	case "identity":
		err = tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM identities i
      JOIN records r
        ON r.incident_id = i.incident_id
       AND r.record_id = i.record_id
     WHERE i.incident_id = $1
       AND i.record_id = $2
       AND i.identity_state IN ('stub', 'canonical')
       AND r.deleted_at IS NULL
)
`, incidentID, subjectRef).Scan(&exists)
	default:
		return &CreateValidationError{Field: "assessment.subject_type", ReasonCode: "invalid_value"}
	}
	if err != nil {
		return fmt.Errorf("validate assessment subject: %w", err)
	}
	if !exists {
		return &CreateValidationError{Field: "assessment.subject_ref", ReasonCode: "invalid_value"}
	}
	return nil
}

func validateSupportRefsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, refs []uuid.UUID) error {
	for _, ref := range uniqueUUIDs(refs) {
		var exists bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND deleted_at IS NULL
)
`, incidentID, ref).Scan(&exists); err != nil {
			return fmt.Errorf("validate assessment support ref: %w", err)
		}
		if !exists {
			return &CreateValidationError{Field: "assessment.support_refs", ReasonCode: "invalid_value"}
		}
	}
	return nil
}

func validateAssessorTx(ctx context.Context, tx pgx.Tx, assessor uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM users
     WHERE id = $1
       AND is_active = true
)
`, assessor).Scan(&exists); err != nil {
		return fmt.Errorf("validate assessment assessor: %w", err)
	}
	if !exists {
		return &CreateValidationError{Field: "assessment.assessor", ReasonCode: "invalid_value"}
	}
	return nil
}

func loadProjectionRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (ProjectionRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT
    p.record_id,
    p.incident_id,
    p.row_version,
    p.subject_ref,
    p.subject_type,
    p.assessment_state,
    p.confidence_score,
    p.confidence_band,
    p.rationale,
    p.assessor,
    p.assessed_at,
    p.supporting_link_count
  FROM assessment_grid_projection p
 WHERE p.record_id = $1
`, recordID)
	record, err := scanProjectionRecord(row)
	if err != nil {
		return ProjectionRecord{}, fmt.Errorf("load assessment projection row: %w", err)
	}
	refs, err := loadSupportRefsTx(ctx, tx, recordID)
	if err != nil {
		return ProjectionRecord{}, err
	}
	record.SupportRefs = refs
	return record, nil
}

func scanProjectionRecord(row pgx.Row) (ProjectionRecord, error) {
	var (
		record          ProjectionRecord
		confidenceScore pgtype.Int4
	)
	if err := row.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.RowVersion,
		&record.SubjectRef,
		&record.SubjectType,
		&record.AssessmentState,
		&confidenceScore,
		&record.ConfidenceBand,
		&record.Rationale,
		&record.Assessor,
		&record.AssessedAt,
		&record.SupportingLinkCount,
	); err != nil {
		return ProjectionRecord{}, err
	}
	if confidenceScore.Valid {
		value := int(confidenceScore.Int32)
		record.ConfidenceScore = &value
	}
	record.AssessedAt = record.AssessedAt.UTC()
	return record, nil
}

func loadSupportRefsTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) ([]SupportRef, error) {
	rows, err := tx.Query(ctx, `
SELECT rl.dst_record_id, dst.record_type
  FROM active_record_links_v1 rl
  JOIN records dst
    ON dst.incident_id = rl.incident_id
   AND dst.record_id = rl.dst_record_id
   AND dst.deleted_at IS NULL
 WHERE rl.src_record_id = $1
   AND rl.link_type = 'supported_by'
   AND rl.deleted_at IS NULL
 ORDER BY dst.record_type ASC, rl.dst_record_id ASC
`, recordID)
	if err != nil {
		return nil, fmt.Errorf("load assessment support refs: %w", err)
	}
	defer rows.Close()
	refs := make([]SupportRef, 0)
	for rows.Next() {
		var ref SupportRef
		if err := rows.Scan(&ref.LinkedRecordID, &ref.RecordType); err != nil {
			return nil, fmt.Errorf("scan assessment support ref: %w", err)
		}
		refs = append(refs, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assessment support refs: %w", err)
	}
	return refs, nil
}

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{}
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

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractUUIDFromPayload(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok || text == "" {
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsed, nil
}

func assessmentVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("assessment:%s:%d", recordID.String(), rowVersion)
}
