package workbookassembly

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

const assessmentCreateResultSchemaID = "cartulary.assessments.create_result.v2"

var assessmentLegacyCreateResultMembers = []string{
	"schema_id",
	"record_id",
	"change_set_id",
	"row_version",
	"row",
}

var assessmentCreateResultMembers = append(append([]string(nil), assessmentLegacyCreateResultMembers...), "default_assessed_at")

type storedAssessmentCreateResult struct {
	DefaultAssessedAt string         `json:"default_assessed_at"`
	SchemaID          string         `json:"schema_id"`
	RecordID          string         `json:"record_id"`
	ChangeSetID       string         `json:"change_set_id"`
	RowVersion        int64          `json:"row_version"`
	CanonicalRow      map[string]any `json:"row"`
}

func encodeAssessmentCreateResult(scopeKey string, receipt assessments.CreateIdempotencyRecord) ([]byte, error) {
	if receipt.DefaultAssessedAt == nil || receipt.DefaultAssessedAt.IsZero() {
		return nil, errors.New("encode assessment create result: default assessment time is required")
	}
	result := receipt.Result
	incidentID, err := assessmentIncidentIDFromScope(scopeKey)
	if err != nil {
		return nil, fmt.Errorf("encode assessment create result: %w", err)
	}
	storedRow := make(map[string]any, len(result.CanonicalRow)+1)
	for key, value := range result.CanonicalRow {
		storedRow[key] = value
	}
	if value, present := storedRow["incident_id"]; present && value != incidentID.String() {
		return nil, errors.New("encode assessment create result: row incident_id mismatch")
	}
	storedRow["incident_id"] = incidentID.String()
	payload, err := json.Marshal(storedAssessmentCreateResult{
		SchemaID:          assessmentCreateResultSchemaID,
		DefaultAssessedAt: receipt.DefaultAssessedAt.UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano),
		RecordID:          result.RecordID.String(),
		ChangeSetID:       result.ChangeSetID.String(),
		RowVersion:        result.RowVersion,
		CanonicalRow:      storedRow,
	})
	if err != nil {
		return nil, fmt.Errorf("encode assessment create result: %w", err)
	}
	if _, err := decodeAssessmentCreateResult(payload, scopeKey, http.StatusCreated, receipt.RequestHash); err != nil {
		return nil, fmt.Errorf("encode assessment create result: %w", err)
	}
	return payload, nil
}

func decodeAssessmentCreateResult(data []byte, scopeKey string, statusCode int, requestHash []byte) (assessments.CreateIdempotencyRecord, error) {
	if statusCode != http.StatusCreated {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid stored status")
	}
	if len(requestHash) != sha256.Size {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid request hash")
	}
	incidentID, err := assessmentIncidentIDFromScope(scopeKey)
	if err != nil {
		return assessments.CreateIdempotencyRecord{}, err
	}
	if err := strictjson.ValidateObject(data); err != nil {
		return assessments.CreateIdempotencyRecord{}, fmt.Errorf("decode assessment create result: %w", err)
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(data, &envelope); err != nil {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid object")
	}
	if _, present := envelope["schema_id"]; !present {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid members")
	}
	schemaID, err := decodeJSONString(envelope["schema_id"])
	if err != nil {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid schema_id")
	}
	var defaultAssessedAt *time.Time
	switch schemaID {
	case "cartulary.assessments.create_result.v1":
		if !hasExactJSONMembers(envelope, assessmentLegacyCreateResultMembers) {
			return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid members")
		}
	case assessmentCreateResultSchemaID:
		if !hasExactJSONMembers(envelope, assessmentCreateResultMembers) {
			return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid members")
		}
		text, err := decodeJSONString(envelope["default_assessed_at"])
		if err != nil {
			return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid default assessment time")
		}
		parsed, err := time.Parse(time.RFC3339Nano, text)
		if err != nil || parsed.IsZero() || parsed.UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano) != text {
			return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid default assessment time")
		}
		defaultAssessedAt = &parsed
	default:
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid schema_id")
	}
	recordID, err := decodeCanonicalUUID(envelope["record_id"])
	if err != nil {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid record_id")
	}
	changeSetID, err := decodeCanonicalUUID(envelope["change_set_id"])
	if err != nil {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid change_set_id")
	}
	rowVersion, err := decodePositiveInt64(envelope["row_version"])
	if err != nil {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid row_version")
	}

	var rowRaw map[string]json.RawMessage
	if err := json.Unmarshal(envelope["row"], &rowRaw); err != nil || rowRaw == nil {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: row must be an object")
	}
	rowRecordID, err := decodeCanonicalUUID(rowRaw["record_id"])
	if err != nil || rowRecordID != recordID {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: row record_id mismatch")
	}
	rowIncidentID, err := decodeCanonicalUUID(rowRaw["incident_id"])
	if err != nil || rowIncidentID != incidentID {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: row incident_id mismatch")
	}
	rowVersionValue, err := decodePositiveInt64(rowRaw["row_version"])
	if err != nil || rowVersionValue != rowVersion {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: row_version mismatch")
	}

	var row map[string]any
	if err := json.Unmarshal(envelope["row"], &row); err != nil {
		return assessments.CreateIdempotencyRecord{}, errors.New("decode assessment create result: invalid row")
	}
	row["row_version"] = rowVersion
	delete(row, "incident_id")
	return assessments.CreateIdempotencyRecord{
		RequestHash:       append([]byte(nil), requestHash...),
		DefaultAssessedAt: defaultAssessedAt,
		Result: assessments.CreateResult{
			Outcome:      assessments.CreateOutcomeCommitted,
			CanonicalRow: row,
			RecordID:     recordID,
			ChangeSetID:  changeSetID,
			RowVersion:   rowVersion,
		},
	}, nil
}

func assessmentIncidentIDFromScope(scopeKey string) (uuid.UUID, error) {
	suffix := ":" + assessments.AssessmentsViewSchemaID
	if !strings.HasSuffix(scopeKey, suffix) {
		return uuid.Nil, errors.New("decode assessment create result: invalid scope")
	}
	incidentText := strings.TrimSuffix(scopeKey, suffix)
	incidentID, err := uuid.Parse(incidentText)
	if err != nil || incidentID == uuid.Nil || incidentID.String() != incidentText {
		return uuid.Nil, errors.New("decode assessment create result: invalid scope")
	}
	return incidentID, nil
}

func hasExactJSONMembers(object map[string]json.RawMessage, members []string) bool {
	if len(object) != len(members) {
		return false
	}
	for _, member := range members {
		if _, ok := object[member]; !ok {
			return false
		}
	}
	return true
}

func decodeJSONString(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", errors.New("missing string")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	return value, nil
}

func decodeCanonicalUUID(raw json.RawMessage) (uuid.UUID, error) {
	value, err := decodeJSONString(raw)
	if err != nil {
		return uuid.Nil, err
	}
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return uuid.Nil, errors.New("invalid canonical uuid")
	}
	return parsed, nil
}

func decodePositiveInt64(raw json.RawMessage) (int64, error) {
	if len(raw) == 0 {
		return 0, errors.New("missing number")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return 0, err
	}
	number, ok := value.(json.Number)
	if !ok || strings.ContainsAny(number.String(), ".eE+-") {
		return 0, errors.New("invalid positive integer")
	}
	parsed, err := strconv.ParseInt(number.String(), 10, 64)
	if err != nil || parsed <= 0 {
		return 0, errors.New("invalid positive integer")
	}
	return parsed, nil
}
