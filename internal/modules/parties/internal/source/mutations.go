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

// LoadPartyValuesForUpdateTx loads and admits the complete Party source row
// under the caller's transaction. Every writer uses this path before
// overlaying trusted or public changes.
func LoadPartyValuesForUpdateTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (uuid.UUID, map[string]policy.Value, error) {
	var incidentID uuid.UUID
	var displayName, partyKind, organizationName, roleTitle, primaryEmail, timezoneName, externalRef, notes *string
	if err := tx.QueryRow(ctx, `
SELECT incident_id, display_name, party_kind, organization_name, role_title,
       primary_email, timezone_name, external_ref, notes
  FROM parties
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(
		&incidentID, &displayName, &partyKind, &organizationName, &roleTitle,
		&primaryEmail, &timezoneName, &externalRef, &notes,
	); err != nil {
		return uuid.Nil, nil, fmt.Errorf("load Party source row: %w", err)
	}
	raw := map[string]*string{
		"party.display_name":      displayName,
		"party.party_kind":        partyKind,
		"party.organization_name": organizationName,
		"party.role_title":        roleTitle,
		"party.primary_email":     primaryEmail,
		"party.timezone_name":     timezoneName,
		"party.external_ref":      externalRef,
		"party.notes":             notes,
	}
	values := make(map[string]policy.Value, len(raw))
	for _, fieldKey := range policy.FieldKeys() {
		value, admissionErr := policy.AdmitStored(fieldKey, raw[fieldKey])
		if admissionErr != nil {
			return uuid.Nil, nil, fmt.Errorf("validate stored Party field %s: %w", fieldKey, admissionErr)
		}
		values[fieldKey] = value
	}
	return incidentID, values, nil
}

// ApplyPatchTx overlays admitted changes on one admitted source load, locks the
// complete current/proposed claim set, and performs exactly one fixed-column
// Party update when the effective value changes.
func ApplyPatchTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	changes []PatchChange,
	now time.Time,
) (bool, error) {
	loadedIncidentID, current, err := LoadPartyValuesForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return false, err
	}
	if loadedIncidentID != incidentID {
		return false, fmt.Errorf("parties source: Party incident does not match record envelope")
	}
	proposed := cloneValues(current)
	for _, change := range changes {
		if _, ok := policy.LookupField(change.FieldKey); !ok || change.Value.FieldKey() != change.FieldKey {
			return false, fmt.Errorf("parties source: unsupported field key %q", change.FieldKey)
		}
		proposed[change.FieldKey] = change.Value
	}
	if valuesEqual(current, proposed) {
		return false, nil
	}
	if err := validateActiveKeyTransitionTx(ctx, tx, incidentID, recordID, current, proposed); err != nil {
		return false, err
	}
	tag, err := tx.Exec(ctx, `
UPDATE parties
   SET display_name = $3,
       party_kind = $4,
       organization_name = $5,
       role_title = $6,
       primary_email = $7,
       timezone_name = $8,
       external_ref = $9,
       notes = $10,
       updated_at = $11
 WHERE incident_id = $1
   AND record_id = $2
`, incidentID, recordID,
		textValue(proposed, "party.display_name"),
		textValue(proposed, "party.party_kind"),
		nullableTextValue(proposed, "party.organization_name"),
		nullableTextValue(proposed, "party.role_title"),
		nullableTextValue(proposed, "party.primary_email"),
		nullableTextValue(proposed, "party.timezone_name"),
		nullableTextValue(proposed, "party.external_ref"),
		nullableTextValue(proposed, "party.notes"),
		now,
	)
	err = adaptActiveKeyClaimError(err, fieldKeysForClaims(proposed))
	if err != nil {
		return false, fmt.Errorf("apply Party patch: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return false, fmt.Errorf("apply Party patch: source row disappeared")
	}
	return true, nil
}

func sameOptionalString(left func() (string, bool), right func() (string, bool)) bool {
	leftValue, leftPresent := left()
	rightValue, rightPresent := right()
	return leftPresent == rightPresent && (!leftPresent || leftValue == rightValue)
}

func cloneValues(values map[string]policy.Value) map[string]policy.Value {
	cloned := make(map[string]policy.Value, len(values))
	for fieldKey, value := range values {
		cloned[fieldKey] = value
	}
	return cloned
}

func valuesEqual(left map[string]policy.Value, right map[string]policy.Value) bool {
	for _, fieldKey := range policy.FieldKeys() {
		if !sameOptionalString(left[fieldKey].EqualityValue, right[fieldKey].EqualityValue) {
			return false
		}
	}
	return true
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
