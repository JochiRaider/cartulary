package rollback

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/vocabulary"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type Provider struct{}

var _ rollbackcontract.RowSourceProvider = Provider{}

func NewProvider() Provider { return Provider{} }

func (Provider) ValidateRollbackValue(value map[string]any) error {
	_, err := parseIndicatorSourcePatch(value)
	return err
}

type rowState struct {
	recordID      uuid.UUID
	incidentID    uuid.UUID
	indicatorType string
	valueKind     string
	displayValue  string
	normalized    *string
	dedupeKey     string
	defanged      *string
	hashAlgorithm *string
	hashValue     *string
	stixPattern   *string
}

type nullableTextPatch struct {
	present bool
	value   *string
}

type indicatorSourcePatch struct {
	recordID      *uuid.UUID
	incidentID    *uuid.UUID
	indicatorType *string
	valueKind     *string
	displayValue  *string
	normalized    nullableTextPatch
	dedupeKey     *string
	defanged      nullableTextPatch
	hashAlgorithm nullableTextPatch
	hashValue     nullableTextPatch
	stixPattern   nullableTextPatch
}

func (Provider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	patch, err := parseIndicatorSourcePatch(request.RetainedValue)
	if err != nil {
		return err
	}
	var state rowState
	if err := tx.QueryRow(ctx, `
SELECT record_id, incident_id, indicator_type, value_kind, display_value, normalized_value, dedupe_key,
       defanged_value, hash_algorithm, hash_value, stix_pattern
  FROM indicators
 WHERE record_id = $1
 FOR UPDATE
`, request.RecordID).Scan(
		&state.recordID, &state.incidentID,
		&state.indicatorType, &state.valueKind, &state.displayValue, &state.normalized, &state.dedupeKey,
		&state.defanged, &state.hashAlgorithm, &state.hashValue, &state.stixPattern,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackcontract.ErrTargetNotFound
		}
		return err
	}
	if !patch.overlay(&state) {
		return rollbackcontract.ErrTargetNotReversible
	}
	canonical, err := identity.Canonicalize(identity.Input{
		IndicatorType:   state.indicatorType,
		ValueKind:       state.valueKind,
		DisplayValue:    state.displayValue,
		NormalizedValue: state.normalized,
		DefangedValue:   state.defanged,
		HashAlgorithm:   state.hashAlgorithm,
		HashValue:       state.hashValue,
		STIXPattern:     state.stixPattern,
	})
	if err != nil || !identityMatchesCanonical(state, canonical) {
		return rollbackcontract.ErrTargetNotReversible
	}
	if patch.dedupeKey != nil && *patch.dedupeKey != canonical.DedupeKey {
		return rollbackcontract.ErrTargetNotReversible
	}
	state.dedupeKey = canonical.DedupeKey
	tag, err := tx.Exec(ctx, `
UPDATE indicators
   SET indicator_type = $2,
       value_kind = $3,
       display_value = $4,
       normalized_value = $5,
       dedupe_key = $6,
       defanged_value = $7,
       hash_algorithm = $8,
       hash_value = $9,
       stix_pattern = $10
 WHERE record_id = $1
`, request.RecordID, state.indicatorType, state.valueKind, state.displayValue, state.normalized, state.dedupeKey, state.defanged, state.hashAlgorithm, state.hashValue, state.stixPattern)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return rollbackcontract.ErrStaleTarget
	}
	return nil
}

func parseIndicatorSourcePatch(value map[string]any) (indicatorSourcePatch, error) {
	rawSource, present := value["source"]
	if !present || rawSource == nil {
		return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
	}
	source, ok := rawSource.(map[string]any)
	if !ok || len(source) == 0 {
		return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
	}
	var patch indicatorSourcePatch
	for key, raw := range source {
		switch key {
		case "record_id":
			parsed, valid := parsePatchUUID(raw)
			if !valid {
				return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
			}
			patch.recordID = &parsed
		case "incident_id":
			parsed, valid := parsePatchUUID(raw)
			if !valid {
				return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
			}
			patch.incidentID = &parsed
		case "indicator_type":
			text, valid := parseRequiredPatchText(raw)
			if !valid || !vocabulary.IsIndicatorType(text) {
				return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
			}
			patch.indicatorType = &text
		case "value_kind":
			text, valid := parseRequiredPatchText(raw)
			if !valid || !vocabulary.IsValueKind(text) {
				return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
			}
			patch.valueKind = &text
		case "display_value":
			text, valid := parseRequiredPatchText(raw)
			if !valid {
				return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
			}
			patch.displayValue = &text
		case "dedupe_key":
			text, valid := parseRequiredPatchText(raw)
			if !valid {
				return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
			}
			patch.dedupeKey = &text
		case "normalized_value":
			parsed, valid := parseNullablePatchText(raw)
			if !valid {
				return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
			}
			patch.normalized = parsed
		case "defanged_value":
			parsed, valid := parseNullablePatchText(raw)
			if !valid {
				return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
			}
			patch.defanged = parsed
		case "hash_algorithm":
			parsed, valid := parseNullablePatchText(raw)
			if !valid {
				return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
			}
			patch.hashAlgorithm = parsed
		case "hash_value":
			parsed, valid := parseNullablePatchText(raw)
			if !valid {
				return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
			}
			patch.hashValue = parsed
		case "stix_pattern":
			parsed, valid := parseNullablePatchText(raw)
			if !valid {
				return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
			}
			patch.stixPattern = parsed
		default:
			return indicatorSourcePatch{}, rollbackcontract.ErrTargetNotReversible
		}
	}
	return patch, nil
}

func parsePatchUUID(raw any) (uuid.UUID, bool) {
	text, ok := raw.(string)
	if !ok || !validPatchText(text) {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(text)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == text
}

func parseRequiredPatchText(raw any) (string, bool) {
	text, ok := raw.(string)
	return text, ok && validPatchText(text)
}

func parseNullablePatchText(raw any) (nullableTextPatch, bool) {
	if raw == nil {
		return nullableTextPatch{present: true}, true
	}
	text, ok := raw.(string)
	if !ok || !validPatchText(text) {
		return nullableTextPatch{}, false
	}
	return nullableTextPatch{present: true, value: &text}, true
}

func validPatchText(text string) bool {
	return text != "" && strings.TrimSpace(text) != "" && !strings.ContainsRune(text, 0)
}

func (patch indicatorSourcePatch) overlay(state *rowState) bool {
	if state == nil || (patch.recordID != nil && *patch.recordID != state.recordID) ||
		(patch.incidentID != nil && *patch.incidentID != state.incidentID) {
		return false
	}
	if patch.indicatorType != nil {
		state.indicatorType = *patch.indicatorType
	}
	if patch.valueKind != nil {
		state.valueKind = *patch.valueKind
	}
	if patch.displayValue != nil {
		state.displayValue = *patch.displayValue
	}
	if patch.normalized.present {
		state.normalized = patch.normalized.value
	}
	if patch.defanged.present {
		state.defanged = patch.defanged.value
	}
	if patch.hashAlgorithm.present {
		state.hashAlgorithm = patch.hashAlgorithm.value
	}
	if patch.hashValue.present {
		state.hashValue = patch.hashValue.value
	}
	if patch.stixPattern.present {
		state.stixPattern = patch.stixPattern.value
	}
	return true
}

func identityMatchesCanonical(state rowState, canonical identity.Canonical) bool {
	return state.indicatorType == canonical.IndicatorType &&
		state.valueKind == canonical.ValueKind &&
		state.displayValue == canonical.DisplayValue &&
		equalStringPointers(state.normalized, canonical.NormalizedValue) &&
		equalStringPointers(state.defanged, canonical.DefangedValue) &&
		equalStringPointers(state.hashAlgorithm, canonical.HashAlgorithm) &&
		equalStringPointers(state.hashValue, canonical.HashValue) &&
		equalStringPointers(state.stixPattern, canonical.STIXPattern)
}
