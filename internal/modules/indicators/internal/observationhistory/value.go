// Package observationhistory owns complete retained Indicator observation values.
package observationhistory

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/vocabulary"
	"github.com/google/uuid"
)

var ErrInvalidValue = errors.New("invalid retained Indicator observation")

type Value struct {
	ObservationID             uuid.UUID
	IncidentID                uuid.UUID
	SourceRecordID            uuid.UUID
	SourceFieldKey            string
	OriginKind                string
	OriginLocator             string
	ObservedText              string
	ParsedIndicatorType       *string
	NormalizedCandidate       *string
	ResolutionStatus          string
	ResolvedIndicatorRecordID *uuid.UUID
	RowVersion                int64
	CreatedByUserID           uuid.UUID
	CreatedAt                 time.Time
	ResolvedByUserID          *uuid.UUID
	ResolvedAt                *time.Time
	ResolutionMethod          *string
	DeletedAt                 *time.Time
	DeletedByUserID           *uuid.UUID
}

func (v Value) Map() map[string]any {
	return map[string]any{
		"indicator_observation_id": v.ObservationID.String(), "incident_id": v.IncidentID.String(), "source_record_id": v.SourceRecordID.String(),
		"source_field_key": v.SourceFieldKey, "origin_kind": v.OriginKind, "origin_locator": v.OriginLocator, "observed_text": v.ObservedText,
		"parsed_indicator_type": nullable(v.ParsedIndicatorType), "normalized_candidate": nullable(v.NormalizedCandidate),
		"resolution_status": v.ResolutionStatus, "resolved_indicator_record_id": uuidPointer(v.ResolvedIndicatorRecordID), "row_version": v.RowVersion,
		"created_by_user_id": v.CreatedByUserID.String(), "created_at": v.CreatedAt.UTC().Format(time.RFC3339Nano),
		"resolved_by_user_id": uuidPointer(v.ResolvedByUserID), "resolved_at": clockPointer(v.ResolvedAt), "resolution_method": nullable(v.ResolutionMethod),
		"deleted_at": clockPointer(v.DeletedAt), "deleted_by_user_id": uuidPointer(v.DeletedByUserID),
	}
}

func Decode(input map[string]any) (Value, error) {
	fields := []string{"indicator_observation_id", "incident_id", "source_record_id", "source_field_key", "origin_kind", "origin_locator", "observed_text", "parsed_indicator_type", "normalized_candidate", "resolution_status", "resolved_indicator_record_id", "row_version", "created_by_user_id", "created_at", "resolved_by_user_id", "resolved_at", "resolution_method", "deleted_at", "deleted_by_user_id"}
	if len(input) != len(fields) {
		return Value{}, ErrInvalidValue
	}
	for _, key := range fields {
		if _, ok := input[key]; !ok {
			return Value{}, ErrInvalidValue
		}
	}
	var v Value
	for key, out := range map[string]*uuid.UUID{"indicator_observation_id": &v.ObservationID, "incident_id": &v.IncidentID, "source_record_id": &v.SourceRecordID, "created_by_user_id": &v.CreatedByUserID} {
		id, ok := canonicalUUID(input[key])
		if !ok {
			return Value{}, ErrInvalidValue
		}
		*out = id
	}
	for key, out := range map[string]**uuid.UUID{"resolved_indicator_record_id": &v.ResolvedIndicatorRecordID, "resolved_by_user_id": &v.ResolvedByUserID, "deleted_by_user_id": &v.DeletedByUserID} {
		if input[key] != nil {
			id, ok := canonicalUUID(input[key])
			if !ok {
				return Value{}, ErrInvalidValue
			}
			*out = &id
		}
	}
	for key, out := range map[string]*string{"source_field_key": &v.SourceFieldKey, "origin_kind": &v.OriginKind, "origin_locator": &v.OriginLocator, "observed_text": &v.ObservedText, "resolution_status": &v.ResolutionStatus} {
		text, ok := nonemptyText(input[key])
		if !ok {
			return Value{}, ErrInvalidValue
		}
		*out = text
	}
	for key, out := range map[string]**string{"parsed_indicator_type": &v.ParsedIndicatorType, "normalized_candidate": &v.NormalizedCandidate, "resolution_method": &v.ResolutionMethod} {
		if input[key] != nil {
			text, ok := nonemptyText(input[key])
			if !ok {
				return Value{}, ErrInvalidValue
			}
			*out = &text
		}
	}
	if _, err := origin.Parse(v.OriginKind); err != nil {
		return Value{}, ErrInvalidValue
	}
	if v.ParsedIndicatorType != nil && !vocabulary.IsIndicatorType(*v.ParsedIndicatorType) || v.NormalizedCandidate != nil && v.ParsedIndicatorType == nil {
		return Value{}, ErrInvalidValue
	}
	var ok bool
	if v.CreatedAt, ok = canonicalTime(input["created_at"]); !ok {
		return Value{}, ErrInvalidValue
	}
	for key, out := range map[string]**time.Time{"resolved_at": &v.ResolvedAt, "deleted_at": &v.DeletedAt} {
		if input[key] != nil {
			at, ok := canonicalTime(input[key])
			if !ok {
				return Value{}, ErrInvalidValue
			}
			*out = &at
		}
	}
	if v.RowVersion, ok = integer(input["row_version"]); !ok || v.RowVersion < 1 {
		return Value{}, ErrInvalidValue
	}
	switch v.ResolutionStatus {
	case "unresolved":
		if v.ResolvedIndicatorRecordID != nil || v.ResolvedByUserID != nil || v.ResolvedAt != nil || v.ResolutionMethod != nil {
			return Value{}, ErrInvalidValue
		}
	case "resolved", "dismissed":
		if v.ResolvedByUserID == nil || v.ResolvedAt == nil || v.ResolutionMethod == nil || (v.ResolutionStatus == "resolved") != (v.ResolvedIndicatorRecordID != nil) {
			return Value{}, ErrInvalidValue
		}
	default:
		return Value{}, ErrInvalidValue
	}
	if (v.DeletedAt == nil) != (v.DeletedByUserID == nil) || v.ResolvedAt != nil && v.ResolvedAt.Before(v.CreatedAt) || v.DeletedAt != nil && (v.DeletedAt.Before(v.CreatedAt) || v.ResolvedAt != nil && v.DeletedAt.Before(*v.ResolvedAt)) {
		return Value{}, ErrInvalidValue
	}
	return v, nil
}

func ValidateMutation(target, operation string, beforeValue, afterValue map[string]any) error {
	after, err := Decode(afterValue)
	if err != nil || after.ObservationID.String() != target {
		return ErrInvalidValue
	}
	if operation == "create" {
		if beforeValue != nil || after.DeletedAt != nil || after.RowVersion != 1 || after.ResolutionStatus == "dismissed" {
			return ErrInvalidValue
		}
		return nil
	}
	before, err := Decode(beforeValue)
	if err != nil || before.ObservationID.String() != target || before.DeletedAt != nil || before.RowVersion == math.MaxInt64 || after.RowVersion != before.RowVersion+1 {
		return ErrInvalidValue
	}
	if operation == "rollback" && after.DeletedAt != nil {
		before.RowVersion, after.RowVersion = 0, 0
		after.DeletedAt, after.DeletedByUserID = nil, nil
		if !reflect.DeepEqual(before, after) {
			return ErrInvalidValue
		}
		return nil
	}
	if after.DeletedAt != nil {
		return ErrInvalidValue
	}
	switch operation {
	case "resolve":
		if before.ResolutionStatus == "dismissed" || after.ResolutionStatus != "resolved" || reflect.DeepEqual(before.ResolvedIndicatorRecordID, after.ResolvedIndicatorRecordID) {
			return ErrInvalidValue
		}
	case "dismiss":
		if before.ResolutionStatus == "dismissed" || after.ResolutionStatus != "dismissed" {
			return ErrInvalidValue
		}
	case "restore":
		if before.ResolutionStatus != "dismissed" || after.ResolutionStatus != "unresolved" {
			return ErrInvalidValue
		}
	case "rollback":
	default:
		return ErrInvalidValue
	}
	if sameResolution(before, after) {
		return ErrInvalidValue
	}
	clearResolution := func(v *Value) {
		v.ResolutionStatus = ""
		v.ResolvedIndicatorRecordID = nil
		v.ResolvedByUserID = nil
		v.ResolvedAt = nil
		v.ResolutionMethod = nil
		v.RowVersion = 0
	}
	clearResolution(&before)
	clearResolution(&after)
	if !reflect.DeepEqual(before, after) {
		return ErrInvalidValue
	}
	return nil
}

func sameResolution(a, b Value) bool {
	return a.ResolutionStatus == b.ResolutionStatus && reflect.DeepEqual(a.ResolvedIndicatorRecordID, b.ResolvedIndicatorRecordID) && reflect.DeepEqual(a.ResolvedByUserID, b.ResolvedByUserID) && reflect.DeepEqual(a.ResolvedAt, b.ResolvedAt) && reflect.DeepEqual(a.ResolutionMethod, b.ResolutionMethod)
}

// Scan captures explicit columns without an intermediate JSON number conversion.
func Scan(scanner interface{ Scan(...any) error }) (Value, error) {
	var v Value
	if err := scanner.Scan(&v.ObservationID, &v.IncidentID, &v.SourceRecordID, &v.SourceFieldKey, &v.OriginKind, &v.OriginLocator, &v.ObservedText, &v.ParsedIndicatorType, &v.NormalizedCandidate, &v.ResolutionStatus, &v.ResolvedIndicatorRecordID, &v.RowVersion, &v.CreatedByUserID, &v.CreatedAt, &v.ResolvedByUserID, &v.ResolvedAt, &v.ResolutionMethod, &v.DeletedAt, &v.DeletedByUserID); err != nil {
		return Value{}, err
	}
	return Decode(v.Map())
}
func nonemptyText(raw any) (string, bool) {
	s, ok := raw.(string)
	return s, ok && s != "" && utf8.ValidString(s) && !strings.ContainsRune(s, 0)
}
func canonicalUUID(raw any) (uuid.UUID, bool) {
	s, ok := raw.(string)
	id, err := uuid.Parse(s)
	return id, ok && err == nil && id != uuid.Nil && id.String() == s
}
func canonicalTime(raw any) (time.Time, bool) {
	s, ok := raw.(string)
	at, err := time.Parse(time.RFC3339Nano, s)
	return at, ok && err == nil && at.Year() >= 1 && at.Year() <= 9999 && at.Nanosecond()%1000 == 0 && at.UTC().Format(time.RFC3339Nano) == s
}
func integer(raw any) (int64, bool) {
	switch n := raw.(type) {
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case json.Number:
		v, err := n.Int64()
		return v, err == nil && strconv.FormatInt(v, 10) == string(n)
	default:
		return 0, false
	}
}
func nullable[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}
func clockPointer(p *time.Time) any {
	if p == nil {
		return nil
	}
	return p.UTC().Format(time.RFC3339Nano)
}
func uuidPointer(p *uuid.UUID) any {
	if p == nil {
		return nil
	}
	return p.String()
}
