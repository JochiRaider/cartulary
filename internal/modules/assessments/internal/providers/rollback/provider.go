package rollback

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	assessmentpolicy "github.com/JochiRaider/cartulary/internal/modules/assessments/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type Provider struct{}

var _ rollbackcontract.RowSourceProvider = Provider{}

func NewProvider() Provider { return Provider{} }

func (Provider) ValidateRollbackValue(value map[string]any) error {
	source, ok := sourceForRollbackValue(value)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if raw, present := source["subject_type"]; present {
		text, valid := raw.(string)
		if !valid || !assessmentpolicy.ValidSubjectType(text) {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	if raw, present := source["assessment_state"]; present {
		text, valid := raw.(string)
		if !valid || !assessmentpolicy.ValidState(text) {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	for _, key := range []string{"subject_record_id", "assessor_user_id"} {
		if _, present, err := requiredUUID(source, key); present && err != nil {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	if raw, present := source["confidence_score"]; present && raw != nil {
		score, valid := numericInt(raw)
		if !valid {
			return rollbackcontract.ErrTargetNotReversible
		}
		if _, valid := assessmentpolicy.ConfidenceBand(&score); !valid {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	if raw, present := source["rationale"]; present {
		if _, valid := raw.(string); !valid {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	if raw, present := source["assessed_at"]; present && raw == nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
}

func (Provider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := sourceForRollbackValue(request.RetainedValue)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if err := (Provider{}).ValidateRollbackValue(request.RetainedValue); err != nil {
		return err
	}
	var incidentID uuid.UUID
	var subjectID uuid.UUID
	var subjectType string
	if err := tx.QueryRow(ctx, `SELECT incident_id, subject_record_id, subject_type FROM assessments WHERE record_id = $1`, request.RecordID).Scan(&incidentID, &subjectID, &subjectType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackcontract.ErrTargetNotFound
		}
		return err
	}
	if parsed, present, _ := requiredUUID(source, "subject_record_id"); present {
		subjectID = parsed
	}
	if raw, present := source["subject_type"]; present {
		subjectType = raw.(string)
	}
	var subjectExists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND record_type = $3
)
`, incidentID, subjectID, subjectType).Scan(&subjectExists); err != nil {
		return err
	}
	if !subjectExists {
		return rollbackcontract.ErrTargetNotReversible
	}
	values := make([]any, 0, 14)
	for _, key := range []string{
		"subject_record_id", "subject_type", "assessment_state", "confidence_score",
		"rationale", "assessor_user_id", "assessed_at",
	} {
		value, present := source[key]
		if key == "subject_record_id" || key == "assessor_user_id" {
			value, _, _ = requiredUUID(source, key)
		}
		values = append(values, present, value)
	}
	_, err := tx.Exec(ctx, `
UPDATE assessments
   SET subject_record_id = CASE WHEN $2 THEN $3::uuid ELSE subject_record_id END,
       subject_type = CASE WHEN $4 THEN $5::text ELSE subject_type END,
       assessment_state = CASE WHEN $6 THEN $7::text ELSE assessment_state END,
       confidence_score = CASE WHEN $8 THEN $9::integer ELSE confidence_score END,
       rationale = CASE WHEN $10 THEN $11::text ELSE rationale END,
       assessor_user_id = CASE WHEN $12 THEN $13::uuid ELSE assessor_user_id END,
       assessed_at = CASE WHEN $14 THEN $15::timestamptz ELSE assessed_at END,
       updated_at = $16,
       deleted_at = NULL,
       deleted_by_user_id = NULL
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	return err
}

func sourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	if source, ok := objectMap(value, "source"); ok {
		return source, len(source) > 0
	}
	return nil, false
}

func objectMap(value map[string]any, key string) (map[string]any, bool) {
	raw, ok := value[key]
	if !ok || raw == nil {
		return nil, false
	}
	typed, ok := raw.(map[string]any)
	return typed, ok
}

func requiredUUID(value map[string]any, key string) (uuid.UUID, bool, error) {
	raw, present := value[key]
	if !present {
		return uuid.Nil, false, nil
	}
	text, valid := raw.(string)
	if !valid || strings.TrimSpace(text) == "" {
		return uuid.Nil, true, errors.New("missing uuid")
	}
	parsed, err := uuid.Parse(text)
	return parsed, true, err
}

func numericInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), math.Trunc(typed) == typed
	default:
		return 0, false
	}
}
