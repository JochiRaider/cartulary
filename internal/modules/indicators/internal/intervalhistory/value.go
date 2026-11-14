// Package intervalhistory owns complete retained Indicator interval values.
package intervalhistory

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/vocabulary"
	"github.com/google/uuid"
)

var ErrInvalidValue = errors.New("invalid retained Indicator interval")

type Value struct {
	IntervalID        uuid.UUID
	IncidentID        uuid.UUID
	IndicatorRecordID uuid.UUID
	LifecycleState    string
	ValidFrom         time.Time
	ValidTo           *time.Time
	Confidence        *int
	Rationale         *string
	SupportRefs       []uuid.UUID
	Assessor          *string
	AssessedAt        time.Time
	RowVersion        int64
	CreatedByUserID   uuid.UUID
	CreatedAt         time.Time
	DeletedAt         *time.Time
	DeletedByUserID   *uuid.UUID
}

func (v Value) Map() map[string]any {
	support := make([]string, len(v.SupportRefs))
	for i, id := range v.SupportRefs {
		support[i] = id.String()
	}
	return map[string]any{"indicator_state_interval_id": v.IntervalID.String(), "incident_id": v.IncidentID.String(), "indicator_record_id": v.IndicatorRecordID.String(),
		"lifecycle_state": v.LifecycleState, "valid_from": v.ValidFrom.UTC().Format(time.RFC3339Nano), "valid_to": clockPointer(v.ValidTo), "confidence": nullable(v.Confidence),
		"rationale": nullable(v.Rationale), "support_refs": support, "assessor": nullable(v.Assessor), "assessed_at": v.AssessedAt.UTC().Format(time.RFC3339Nano),
		"row_version": v.RowVersion, "created_by_user_id": v.CreatedByUserID.String(), "created_at": v.CreatedAt.UTC().Format(time.RFC3339Nano),
		"deleted_at": clockPointer(v.DeletedAt), "deleted_by_user_id": uuidPointer(v.DeletedByUserID)}
}

func Decode(input map[string]any) (Value, error) {
	fields := []string{"indicator_state_interval_id", "incident_id", "indicator_record_id", "lifecycle_state", "valid_from", "valid_to", "confidence", "rationale", "support_refs", "assessor", "assessed_at", "row_version", "created_by_user_id", "created_at", "deleted_at", "deleted_by_user_id"}
	if len(input) != len(fields) {
		return Value{}, ErrInvalidValue
	}
	for _, key := range fields {
		if _, ok := input[key]; !ok {
			return Value{}, ErrInvalidValue
		}
	}
	v := Value{}
	for key, out := range map[string]*uuid.UUID{"indicator_state_interval_id": &v.IntervalID, "incident_id": &v.IncidentID, "indicator_record_id": &v.IndicatorRecordID, "created_by_user_id": &v.CreatedByUserID} {
		id, ok := canonicalUUID(input[key])
		if !ok {
			return Value{}, ErrInvalidValue
		}
		*out = id
	}
	state, ok := input["lifecycle_state"].(string)
	if !ok || !vocabulary.IsLifecycleState(state) {
		return Value{}, ErrInvalidValue
	}
	v.LifecycleState = state
	for key, out := range map[string]*time.Time{"valid_from": &v.ValidFrom, "assessed_at": &v.AssessedAt, "created_at": &v.CreatedAt} {
		at, ok := canonicalTime(input[key])
		if !ok {
			return Value{}, ErrInvalidValue
		}
		*out = at
	}
	for key, out := range map[string]**time.Time{"valid_to": &v.ValidTo, "deleted_at": &v.DeletedAt} {
		if input[key] != nil {
			at, ok := canonicalTime(input[key])
			if !ok {
				return Value{}, ErrInvalidValue
			}
			*out = &at
		}
	}
	if v.ValidTo != nil && v.ValidTo.Before(v.ValidFrom) || v.DeletedAt != nil && v.DeletedAt.Before(v.CreatedAt) {
		return Value{}, ErrInvalidValue
	}
	if input["deleted_by_user_id"] != nil {
		id, ok := canonicalUUID(input["deleted_by_user_id"])
		if !ok {
			return Value{}, ErrInvalidValue
		}
		v.DeletedByUserID = &id
	}
	if (v.DeletedAt == nil) != (v.DeletedByUserID == nil) {
		return Value{}, ErrInvalidValue
	}
	version, ok := integer(input["row_version"])
	if !ok || version < 1 {
		return Value{}, ErrInvalidValue
	}
	v.RowVersion = version
	if input["confidence"] != nil {
		n, ok := integer(input["confidence"])
		if !ok || n < 0 || n > 100 {
			return Value{}, ErrInvalidValue
		}
		number := int(n)
		v.Confidence = &number
	}
	for key, out := range map[string]**string{"rationale": &v.Rationale, "assessor": &v.Assessor} {
		if input[key] != nil {
			text, ok := input[key].(string)
			if !ok || !utf8.ValidString(text) || strings.ContainsRune(text, 0) {
				return Value{}, ErrInvalidValue
			}
			*out = &text
		}
	}
	var support []any
	switch values := input["support_refs"].(type) {
	case []any:
		if values == nil {
			return Value{}, ErrInvalidValue
		}
		support = values
	case []string:
		if values == nil {
			return Value{}, ErrInvalidValue
		}
		support = make([]any, len(values))
		for i, text := range values {
			support[i] = text
		}
	default:
		return Value{}, ErrInvalidValue
	}
	v.SupportRefs = make([]uuid.UUID, 0, len(support))
	seen := map[uuid.UUID]bool{}
	for _, raw := range support {
		id, ok := canonicalUUID(raw)
		if !ok || seen[id] {
			return Value{}, ErrInvalidValue
		}
		seen[id] = true
		v.SupportRefs = append(v.SupportRefs, id)
	}
	return v, nil
}

func ValidateMutation(target, operation string, beforeValue, afterValue map[string]any) error {
	after, err := Decode(afterValue)
	if err != nil || after.IntervalID.String() != target {
		return ErrInvalidValue
	}
	if operation == "create" {
		if beforeValue != nil || after.DeletedAt != nil || after.RowVersion != 1 {
			return ErrInvalidValue
		}
		return nil
	}
	if operation != "rollback" {
		return ErrInvalidValue
	}
	before, err := Decode(beforeValue)
	if err != nil || before.IntervalID.String() != target || before.DeletedAt != nil || after.DeletedAt == nil || before.RowVersion == math.MaxInt64 || after.RowVersion != before.RowVersion+1 {
		return ErrInvalidValue
	}
	before.RowVersion, after.RowVersion = 0, 0
	after.DeletedAt, after.DeletedByUserID = nil, nil
	if !reflect.DeepEqual(before, after) {
		return ErrInvalidValue
	}
	return nil
}

func canonicalUUID(raw any) (uuid.UUID, bool) {
	text, ok := raw.(string)
	id, err := uuid.Parse(text)
	return id, ok && err == nil && id != uuid.Nil && id.String() == text
}
func canonicalTime(raw any) (time.Time, bool) {
	text, ok := raw.(string)
	at, err := time.Parse(time.RFC3339Nano, text)
	return at, ok && err == nil && at.Year() >= 1 && at.Year() <= 9999 && at.Nanosecond()%1000 == 0 && at.UTC().Format(time.RFC3339Nano) == text
}
func integer(raw any) (int64, bool) {
	switch value := raw.(type) {
	case int:
		return int64(value), true
	case int32:
		return int64(value), true
	case int64:
		return value, true
	case json.Number:
		n, err := value.Int64()
		return n, err == nil && strconv.FormatInt(n, 10) == string(value)
	default:
		return 0, false
	}
}
func nullable[T any](value *T) any {
	if value == nil {
		return nil
	}
	return *value
}
func clockPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}
func uuidPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

// Scan captures the explicit source columns without a JSON numeric round trip.
func Scan(scanner interface{ Scan(...any) error }) (Value, error) {
	var value Value
	var supportJSON []byte
	if err := scanner.Scan(&value.IntervalID, &value.IncidentID, &value.IndicatorRecordID, &value.LifecycleState, &value.ValidFrom, &value.ValidTo,
		&value.Confidence, &value.Rationale, &supportJSON, &value.Assessor, &value.AssessedAt, &value.RowVersion, &value.CreatedByUserID, &value.CreatedAt, &value.DeletedAt, &value.DeletedByUserID); err != nil {
		return Value{}, err
	}
	var support []string
	if err := json.Unmarshal(supportJSON, &support); err != nil || support == nil {
		return Value{}, ErrInvalidValue
	}
	value.SupportRefs = make([]uuid.UUID, 0, len(support))
	for _, text := range support {
		id, ok := canonicalUUID(text)
		if !ok {
			return Value{}, ErrInvalidValue
		}
		value.SupportRefs = append(value.SupportRefs, id)
	}
	return Decode(value.Map())
}
