package rollbackprovider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

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
	if raw, present := source["indicator_type"]; present && !validIndicatorType(raw.(string)) {
		return rollbackcontract.ErrTargetNotReversible
	}
	if raw, present := source["value_kind"]; present && !validValueKind(raw.(string)) {
		return rollbackcontract.ErrTargetNotReversible
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
	if raw, present := source["dedupe_key"]; present {
		state.dedupeKey = raw.(string)
	} else if indicatorIdentityRepresented(source) {
		state.dedupeKey = buildDedupeKey(state)
	}
	if !validIndicatorType(state.indicatorType) || !validValueKind(state.valueKind) || strings.TrimSpace(state.displayValue) == "" || strings.TrimSpace(state.dedupeKey) == "" {
		return rollbackcontract.ErrTargetNotReversible
	}
	if (state.hashAlgorithm == nil) != (state.hashValue == nil) {
		return rollbackcontract.ErrTargetNotReversible
	}
	_, err := tx.Exec(ctx, `
UPDATE indicators
   SET indicator_type = $2,
       value_kind = $3,
       display_value = $4,
       normalized_value = $5,
       dedupe_key = $6,
       defanged_value = $7,
       hash_algorithm = $8,
       hash_value = $9,
       stix_pattern = $10,
       row_version = $11,
       updated_at = $12,
       updated_by_user_id = $13,
       deleted_at = NULL,
       deleted_by_user_id = NULL
 WHERE record_id = $1
`, request.RecordID, state.indicatorType, state.valueKind, state.displayValue, state.normalized, state.dedupeKey, state.defanged, state.hashAlgorithm, state.hashValue, state.stixPattern, request.NextRowVersion, request.Now.UTC(), request.ActorUserID)
	return err
}

func (Provider) TouchTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.TouchRequest) error {
	_, err := tx.Exec(ctx, `UPDATE indicators SET row_version = $2, updated_at = $3, updated_by_user_id = $4 WHERE record_id = $1`, request.RecordID, request.NextRowVersion, request.Now.UTC(), request.ActorUserID)
	return err
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

func indicatorIdentityRepresented(source map[string]any) bool {
	for _, key := range []string{"indicator_type", "value_kind", "display_value", "normalized_value", "hash_algorithm", "hash_value"} {
		if _, present := source[key]; present {
			return true
		}
	}
	return false
}

func buildDedupeKey(state rowState) string {
	parts := []string{
		state.indicatorType,
		state.valueKind,
		state.displayValue,
		derefString(state.normalized),
		derefString(state.hashAlgorithm),
		derefString(state.hashValue),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x1f")))
	return hex.EncodeToString(sum[:])
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func validIndicatorType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "ipv4_addr", "ipv6_addr", "domain_name", "url", "sha256", "email_addr", "registry_key", "process_name", "text":
		return true
	default:
		return false
	}
}

func validValueKind(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "atomic", "pattern", "reference":
		return true
	default:
		return false
	}
}
