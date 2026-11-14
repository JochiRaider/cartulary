// Package sourcevalue owns the complete stored Assessment value independently
// of transport, database-record JSON and derived workbook fields.
package sourcevalue

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/assessments/internal/policy"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/google/uuid"
)

type Row struct {
	RecordID, IncidentID, SubjectRecordID, AssessorUserID uuid.UUID
	SubjectType, AssessmentState, Rationale               string
	ConfidenceScore                                       *int
	AssessedAt, CreatedAt, UpdatedAt                      time.Time
	DeletedAt                                             *time.Time
	DeletedByUserID                                       *uuid.UUID
}

// Parse admits only the exact scalar shape. Semantic checks run in separate
// whole-member passes so a later row cannot invert the owner's failure order.
func Parse(raw map[string]any) (Row, bool) {
	var row Row
	fields := strings.Fields("record_id incident_id subject_record_id subject_type assessment_state confidence_score assessor_user_id assessed_at created_at updated_at deleted_at deleted_by_user_id rationale")
	if len(raw) != len(fields) {
		return row, false
	}
	for _, field := range fields {
		if _, ok := raw[field]; !ok {
			return row, false
		}
	}
	var ok bool
	for _, pair := range []struct {
		name   string
		target *uuid.UUID
	}{
		{"record_id", &row.RecordID}, {"incident_id", &row.IncidentID},
		{"subject_record_id", &row.SubjectRecordID}, {"assessor_user_id", &row.AssessorUserID},
	} {
		if *pair.target, ok = CanonicalUUID(raw[pair.name]); !ok {
			return Row{}, false
		}
	}
	for _, pair := range []struct {
		name   string
		target *string
	}{
		{"subject_type", &row.SubjectType}, {"assessment_state", &row.AssessmentState}, {"rationale", &row.Rationale},
	} {
		if *pair.target, ok = raw[pair.name].(string); !ok {
			return Row{}, false
		}
	}
	if raw["confidence_score"] != nil {
		number, ok := raw["confidence_score"].(json.Number)
		if !ok {
			return Row{}, false
		}
		value, err := number.Int64()
		score := int(value)
		if err != nil || int64(score) != value || strconv.FormatInt(value, 10) != number.String() {
			return Row{}, false
		}
		row.ConfidenceScore = &score
	}
	for _, pair := range []struct {
		name   string
		target *time.Time
	}{
		{"assessed_at", &row.AssessedAt}, {"created_at", &row.CreatedAt}, {"updated_at", &row.UpdatedAt},
	} {
		if *pair.target, ok = canonicalTime(raw[pair.name]); !ok {
			return Row{}, false
		}
	}
	if raw["deleted_at"] != nil {
		value, ok := canonicalTime(raw["deleted_at"])
		if !ok {
			return Row{}, false
		}
		row.DeletedAt = &value
	}
	if raw["deleted_by_user_id"] != nil {
		value, ok := CanonicalUUID(raw["deleted_by_user_id"])
		if !ok {
			return Row{}, false
		}
		row.DeletedByUserID = &value
	}
	return row, true
}

func CanonicalUUID(raw any) (uuid.UUID, bool) {
	text, ok := raw.(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(text)
	return id, err == nil && id != uuid.Nil && id.String() == text
}

func canonicalTime(raw any) (time.Time, bool) {
	text, ok := raw.(string)
	if !ok {
		return time.Time{}, false
	}
	value, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || value.Year() < 1 || value.Nanosecond()%1000 != 0 {
		return time.Time{}, false
	}
	canonical := value.UTC().Format(time.RFC3339Nano)
	return value.UTC(), text == canonical || text == strings.TrimSuffix(canonical, "Z")+"+00:00"
}

func (r Row) SubjectLegal(incidentID uuid.UUID) bool {
	return r.IncidentID == incidentID && policy.ValidSubjectType(r.SubjectType) && r.SubjectRecordID != r.RecordID
}
func (r Row) JudgmentLegal() bool {
	_, valid := policy.ConfidenceBand(r.ConfidenceScore)
	rationale, canonical := fieldnorm.NormalizeNote(r.Rationale)
	return valid && policy.ValidState(r.AssessmentState) && canonical && rationale == r.Rationale
}
func (r Row) LifecycleLegal() bool {
	if r.UpdatedAt.Before(r.CreatedAt) || (r.DeletedAt == nil) != (r.DeletedByUserID == nil) {
		return false
	}
	return r.DeletedAt == nil || (!r.DeletedAt.Before(r.CreatedAt) && !r.DeletedAt.After(r.UpdatedAt))
}
