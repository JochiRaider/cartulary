package indicators

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CanonicalIPIndicatorType supplies Core's atomic IP policy to source owners.
// It recognizes exact canonical values; it never repairs a confirmation.
func CanonicalIPIndicatorType(value string) (string, bool) {
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return "", false
	}
	kind := "ipv6_addr"
	if addr.Is4() {
		kind = "ipv4_addr"
	}
	canonical, err := identity.Canonicalize(identity.Input{IndicatorType: kind, ValueKind: "atomic", DisplayValue: value})
	if err != nil || canonical.DisplayValue != value {
		return "", false
	}
	return canonical.IndicatorType, true
}

type indicatorRecordQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// GetActiveIndicatorParticipant reads an indicator through the Indicator
// owner. Consumers never query the indicators or records tables directly.
func (s *Application) GetActiveIndicatorParticipant(ctx context.Context, incidentID uuid.UUID, indicatorID uuid.UUID) (IndicatorReference, error) {
	return getActiveIndicatorParticipant(ctx, s.pool, incidentID, indicatorID)
}

// GetActiveIndicatorParticipantTx participates in a consumer-owned atomic
// operation without taking ownership of the outer transaction.
func (*Application) GetActiveIndicatorParticipantTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, indicatorID uuid.UUID) (IndicatorReference, error) {
	return getActiveIndicatorParticipant(ctx, tx, incidentID, indicatorID)
}

func getActiveIndicatorParticipant(ctx context.Context, querier indicatorRecordQuerier, incidentID uuid.UUID, indicatorID uuid.UUID) (IndicatorReference, error) {
	record, err := scanIndicatorRecord(querier.QueryRow(ctx, `
SELECT
    i.record_id, i.incident_id, i.indicator_type, i.value_kind, i.display_value,
    i.normalized_value, i.dedupe_key, i.defanged_value, i.hash_algorithm,
    i.hash_value, i.stix_pattern, r.row_version, r.created_at, r.updated_at,
    r.created_by_user_id, r.updated_by_user_id, r.deleted_at, r.deleted_by_user_id
  FROM indicators i
  JOIN records r ON r.record_id = i.record_id
 WHERE i.incident_id = $1
   AND i.record_id = $2
   AND r.deleted_at IS NULL
`, incidentID, indicatorID))
	if errors.Is(err, pgx.ErrNoRows) {
		return IndicatorReference{}, ErrIndicatorNotFound
	}
	if err != nil {
		return IndicatorReference{}, fmt.Errorf("get active indicator participant: %w", err)
	}
	return referenceFromRecord(record), nil
}
