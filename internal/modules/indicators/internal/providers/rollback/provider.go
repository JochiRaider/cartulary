package rollback

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
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
	for _, key := range []string{"indicator_type", "value_kind", "display_value", "dedupe_key"} {
		if raw, present := source[key]; present {
			text, valid := raw.(string)
			if !valid || strings.TrimSpace(text) == "" {
				return rollbackcontract.ErrTargetNotReversible
			}
		}
	}
	if raw, present := source["indicator_type"]; present {
		if _, err := identity.NormalizeIndicatorType(raw.(string)); err != nil {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	if raw, present := source["value_kind"]; present {
		if _, err := identity.NormalizeValueKind(raw.(string)); err != nil {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	return nil
}

type rowState struct {
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

func (Provider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := sourceForRollbackValue(request.RetainedValue)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if err := (Provider{}).ValidateRollbackValue(request.RetainedValue); err != nil {
		return err
	}
	var state rowState
	if err := tx.QueryRow(ctx, `
SELECT indicator_type, value_kind, display_value, normalized_value, dedupe_key,
       defanged_value, hash_algorithm, hash_value, stix_pattern
  FROM indicators
 WHERE record_id = $1
`, request.RecordID).Scan(
		&state.indicatorType, &state.valueKind, &state.displayValue, &state.normalized, &state.dedupeKey,
		&state.defanged, &state.hashAlgorithm, &state.hashValue, &state.stixPattern,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackcontract.ErrTargetNotFound
		}
		return err
	}
	applyText(source, "indicator_type", &state.indicatorType)
	applyText(source, "value_kind", &state.valueKind)
	applyText(source, "display_value", &state.displayValue)
	applyNullableText(source, "normalized_value", &state.normalized)
	applyNullableText(source, "defanged_value", &state.defanged)
	applyNullableText(source, "hash_algorithm", &state.hashAlgorithm)
	applyNullableText(source, "hash_value", &state.hashValue)
	applyNullableText(source, "stix_pattern", &state.stixPattern)
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
	if raw, present := source["dedupe_key"]; present && raw.(string) != canonical.DedupeKey {
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

func sourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	if source, ok := objectMap(value, "source"); ok {
		return source, len(source) > 0
	}
	if cells, ok := objectMap(value, "cells"); ok {
		source := map[string]any{}
		mapping := map[string]string{
			"indicator.indicator_type":   "indicator_type",
			"indicator.value_kind":       "value_kind",
			"indicator.display_value":    "display_value",
			"indicator.normalized_value": "normalized_value",
			"indicator.defanged_value":   "defanged_value",
			"indicator.hash_algorithm":   "hash_algorithm",
			"indicator.hash_value":       "hash_value",
			"indicator.stix_pattern":     "stix_pattern",
		}
		for fieldKey, sourceKey := range mapping {
			if cell, present := objectMap(cells, fieldKey); present {
				source[sourceKey] = cell["value"]
			}
		}
		return source, len(source) > 0
	}
	if _, ok := value["record_id"]; ok {
		for _, key := range []string{
			"indicator_type", "value_kind", "display_value", "normalized_value", "dedupe_key",
			"defanged_value", "hash_algorithm", "hash_value", "stix_pattern",
		} {
			if _, present := value[key]; present {
				return value, true
			}
		}
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

func applyText(source map[string]any, key string, destination *string) {
	if raw, present := source[key]; present {
		*destination, _ = raw.(string)
	}
}

func applyNullableText(source map[string]any, key string, destination **string) {
	raw, present := source[key]
	if !present {
		return
	}
	text, valid := raw.(string)
	if !valid || strings.TrimSpace(text) == "" {
		*destination = nil
		return
	}
	*destination = &text
}

func identityMatchesCanonical(state rowState, canonical identity.Canonical) bool {
	return state.indicatorType == canonical.IndicatorType &&
		state.valueKind == canonical.ValueKind &&
		state.displayValue == canonical.DisplayValue &&
		equalStringPointers(state.normalized, canonical.NormalizedValue) &&
		equalStringPointers(state.hashAlgorithm, canonical.HashAlgorithm) &&
		equalStringPointers(state.hashValue, canonical.HashValue)
}
