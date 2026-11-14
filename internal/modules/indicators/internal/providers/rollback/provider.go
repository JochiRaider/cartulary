package rollback

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type Provider struct {
	assignments reference_data.RegistryAssignments
}

var _ rollbackcontract.RowSourceProvider = Provider{}

func NewProvider(assignments reference_data.RegistryAssignments) Provider {
	return Provider{assignments: assignments}
}

func (Provider) ValidateRollbackValue(value map[string]any) error {
	_, err := parseCompleteIndicatorSource(value)
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

func (p Provider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	state, err := parseCompleteIndicatorSource(request.RetainedValue)
	if err != nil {
		return err
	}
	if state.recordID != request.RecordID {
		return rollbackcontract.ErrTargetNotReversible
	}
	assignment, err := p.assignments.BeginTx(ctx, tx)
	if err != nil {
		return err
	}
	evaluate, err := identity.FromConsumer(ctx, assignment)
	if err != nil {
		return err
	}
	var currentIncidentID uuid.UUID
	var previousType string
	if err := tx.QueryRow(ctx, `
SELECT incident_id, indicator_type
  FROM indicators
 WHERE record_id = $1
 FOR UPDATE
`, request.RecordID).Scan(&currentIncidentID, &previousType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackcontract.ErrTargetNotFound
		}
		return err
	}
	if state.incidentID != currentIncidentID {
		return rollbackcontract.ErrTargetNotReversible
	}
	canonical, err := identity.Canonicalize(evaluate, identity.Input{
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
	if state.dedupeKey != canonical.DedupeKey {
		return rollbackcontract.ErrTargetNotReversible
	}
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
	if previousType != state.indicatorType {
		return assignment.RecordUsage(ctx, "type_registry.indicator")
	}
	return nil
}

func identityMatchesCanonical(state rowState, canonical identity.Canonical) bool {
	// Retained presentation fields preserve explicit nulls. Creation may derive
	// a defanged suggestion, but rollback must not require or materialize it.
	return state.indicatorType == canonical.IndicatorType &&
		state.valueKind == canonical.ValueKind &&
		state.displayValue == canonical.DisplayValue &&
		equalStringPointers(state.normalized, canonical.NormalizedValue) &&
		equalStringPointers(state.hashAlgorithm, canonical.HashAlgorithm) &&
		equalStringPointers(state.hashValue, canonical.HashValue) &&
		equalStringPointers(state.stixPattern, canonical.STIXPattern)
}
