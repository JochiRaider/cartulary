package source

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/parties/internal/policy"
)

type CreateParams struct {
	Values map[string]policy.Value
}

type PatchChange struct {
	FieldKey string
	Value    policy.Value
}

func InsertPartyTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	incidentID uuid.UUID,
	params CreateParams,
	now time.Time,
) error {
	_, err := tx.Exec(ctx, `
INSERT INTO parties (
    record_id, incident_id, display_name, party_kind, organization_name, role_title,
    primary_email, timezone_name, external_ref, notes, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
`, recordID, incidentID,
		textValue(params.Values, "party.display_name"),
		textValue(params.Values, "party.party_kind"),
		nullableTextValue(params.Values, "party.organization_name"),
		nullableTextValue(params.Values, "party.role_title"),
		nullableTextValue(params.Values, "party.primary_email"),
		nullableTextValue(params.Values, "party.timezone_name"),
		nullableTextValue(params.Values, "party.external_ref"),
		nullableTextValue(params.Values, "party.notes"),
		now)
	err = adaptActiveKeyClaimError(err, fieldKeysForClaims(params.Values))
	if err != nil {
		return fmt.Errorf("insert party: %w", err)
	}
	return nil
}

func ApplyDirectChangeTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	change PatchChange,
	now time.Time,
) (bool, error) {
	field, ok := policy.LookupField(change.FieldKey)
	if !ok || change.Value.FieldKey() != change.FieldKey {
		return false, fmt.Errorf("parties source: unsupported field key %q", change.FieldKey)
	}
	column := field.SourceColumn
	var current *string
	if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT %s FROM parties WHERE record_id = $1`, column), recordID).Scan(&current); err != nil {
		return false, fmt.Errorf("load current Party field %s: %w", change.FieldKey, err)
	}
	currentValue, admissionErr := policy.AdmitStored(change.FieldKey, current)
	if admissionErr != nil {
		return false, fmt.Errorf("validate current Party field %s: %w", change.FieldKey, admissionErr)
	}
	if sameOptionalString(currentValue.EqualityValue, change.Value.EqualityValue) {
		return false, nil
	}
	dbValue := directDBValue(change.Value)
	tag, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE parties SET %s = $2, updated_at = $3 WHERE record_id = $1 AND %s IS DISTINCT FROM $2`, column, column), recordID, dbValue, now)
	err = adaptActiveKeyClaimError(err, []string{change.FieldKey})
	if err != nil {
		return false, fmt.Errorf("apply party direct change: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func TouchPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `UPDATE parties SET updated_at = $2 WHERE record_id = $1`, recordID, now); err != nil {
		return fmt.Errorf("touch party row: %w", err)
	}
	return nil
}

func HasStoredText(values map[string]policy.Value, field string) bool {
	value, ok := values[field]
	if !ok {
		return false
	}
	_, present := value.StoredValue()
	return present
}

func sameOptionalString(left func() (string, bool), right func() (string, bool)) bool {
	leftValue, leftPresent := left()
	rightValue, rightPresent := right()
	return leftPresent == rightPresent && (!leftPresent || leftValue == rightValue)
}

func fieldKeysForClaims(values map[string]policy.Value) []string {
	fields := make([]string, 0, 2)
	for _, field := range []string{"party.primary_email", "party.external_ref"} {
		if value, ok := values[field]; ok {
			if _, present := value.ExactMatchClaimValue(); !present {
				continue
			}
			fields = append(fields, field)
		}
	}
	return fields
}

func directDBValue(value policy.Value) any {
	if stored, present := value.StoredValue(); present {
		return stored
	}
	return nil
}

func textValue(values map[string]policy.Value, field string) string {
	if value, ok := values[field]; ok {
		if stored, present := value.StoredValue(); present {
			return stored
		}
	}
	return ""
}

func nullableTextValue(values map[string]policy.Value, field string) any {
	if value, ok := values[field]; ok {
		if stored, present := value.StoredValue(); present {
			return stored
		}
	}
	return nil
}
