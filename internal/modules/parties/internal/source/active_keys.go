package source

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/parties/internal/policy"
)

const (
	MatchAmbiguousExactMatch = "ambiguous_exact_match"
	MatchCrossKeyExactMatch  = "cross_key_exact_match"
	MatchExactKeyClaimed     = "exact_match_key_claimed"
)

// MatchConflictError is deliberately value-free. The Parties root re-exports
// this closed semantic error; boundary adapters may expose only its reason and
// sorted field keys.
type MatchConflictError struct {
	ReasonCode           string
	ConflictingFieldKeys []string
}

func (e *MatchConflictError) Error() string { return "parties: exact-match conflict" }

type activeKeyTuple struct {
	keyKind         string
	normalizedValue string
	fieldKey        string
}

func FindReusablePartyTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	params CreateParams,
) (uuid.UUID, bool, error) {
	tuples := activeKeyTuples(params.Values)
	if err := lockActiveKeyTuplesTx(ctx, tx, incidentID, tuples); err != nil {
		return uuid.UUID{}, false, err
	}

	matches := make(map[uuid.UUID][]string)
	for _, tuple := range tuples {
		var recordID uuid.UUID
		err := tx.QueryRow(ctx, `
SELECT party_record_id
  FROM party_active_key_claims
 WHERE incident_id = $1
   AND key_kind = $2
   AND normalized_value = $3
`, incidentID, tuple.keyKind, tuple.normalizedValue).Scan(&recordID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return uuid.UUID{}, false, fmt.Errorf("resolve Party active-key claim: %w", err)
		}
		matches[recordID] = append(matches[recordID], tuple.fieldKey)
	}

	if len(matches) == 0 {
		return uuid.UUID{}, false, nil
	}
	if len(matches) > 1 {
		return uuid.UUID{}, false, newMatchConflict(MatchCrossKeyExactMatch, matchFields(matches))
	}
	for recordID := range matches {
		return recordID, true, nil
	}
	panic("unreachable Party claim resolution")
}

func PreparePatchActiveKeysTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	changes []PatchChange,
) error {
	var currentEmail, currentExternalRef *string
	if err := tx.QueryRow(ctx, `
SELECT primary_email, external_ref
  FROM parties
 WHERE incident_id = $1
   AND record_id = $2
 FOR UPDATE
`, incidentID, recordID).Scan(&currentEmail, &currentExternalRef); err != nil {
		return fmt.Errorf("load Party active keys: %w", err)
	}

	currentEmailValue, admissionErr := policy.AdmitStored("party.primary_email", currentEmail)
	if admissionErr != nil {
		return fmt.Errorf("validate stored Party primary email: %w", admissionErr)
	}
	currentExternalRefValue, admissionErr := policy.AdmitStored("party.external_ref", currentExternalRef)
	if admissionErr != nil {
		return fmt.Errorf("validate stored Party external reference: %w", admissionErr)
	}
	current := map[string]policy.Value{
		"party.primary_email": currentEmailValue,
		"party.external_ref":  currentExternalRefValue,
	}
	proposed := map[string]policy.Value{
		"party.primary_email": currentEmailValue,
		"party.external_ref":  currentExternalRefValue,
	}
	for _, change := range changes {
		switch change.FieldKey {
		case "party.primary_email", "party.external_ref":
			proposed[change.FieldKey] = change.Value
		}
	}

	allTuples := append(activeKeyTuples(current), activeKeyTuples(proposed)...)
	if err := lockActiveKeyTuplesTx(ctx, tx, incidentID, allTuples); err != nil {
		return err
	}

	var claimedFields []string
	for _, tuple := range activeKeyTuples(proposed) {
		var ownerID uuid.UUID
		err := tx.QueryRow(ctx, `
SELECT party_record_id
  FROM party_active_key_claims
 WHERE incident_id = $1
   AND key_kind = $2
   AND normalized_value = $3
`, incidentID, tuple.keyKind, tuple.normalizedValue).Scan(&ownerID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("validate proposed Party active-key claim: %w", err)
		}
		if ownerID != recordID {
			claimedFields = append(claimedFields, tuple.fieldKey)
		}
	}
	if len(claimedFields) > 0 {
		return newMatchConflict(MatchExactKeyClaimed, claimedFields)
	}
	return nil
}

func activeKeyTuples(values map[string]policy.Value) []activeKeyTuple {
	tuples := make([]activeKeyTuple, 0, 2)
	for _, descriptor := range []struct {
		fieldKey string
		keyKind  string
	}{
		{fieldKey: "party.primary_email", keyKind: "primary_email"},
		{fieldKey: "party.external_ref", keyKind: "external_ref"},
	} {
		value, ok := values[descriptor.fieldKey]
		if !ok {
			continue
		}
		normalized, present := value.ExactMatchClaimValue()
		if !present {
			continue
		}
		tuples = append(tuples, activeKeyTuple{
			keyKind: descriptor.keyKind, normalizedValue: normalized, fieldKey: descriptor.fieldKey,
		})
	}
	slices.SortFunc(tuples, compareActiveKeyTuples)
	return slices.CompactFunc(tuples, func(left, right activeKeyTuple) bool {
		return left.keyKind == right.keyKind && left.normalizedValue == right.normalizedValue
	})
}

func compareActiveKeyTuples(left, right activeKeyTuple) int {
	if compared := strings.Compare(left.keyKind, right.keyKind); compared != 0 {
		return compared
	}
	return strings.Compare(left.normalizedValue, right.normalizedValue)
}

func lockActiveKeyTuplesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, tuples []activeKeyTuple) error {
	slices.SortFunc(tuples, compareActiveKeyTuples)
	tuples = slices.CompactFunc(tuples, func(left, right activeKeyTuple) bool {
		return left.keyKind == right.keyKind && left.normalizedValue == right.normalizedValue
	})
	for _, tuple := range tuples {
		lockIdentity := incidentID.String() + "\x1f" + tuple.keyKind + "\x1f" + tuple.normalizedValue
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, lockIdentity); err != nil {
			return fmt.Errorf("lock Party active-key tuple: %w", err)
		}
	}
	return nil
}

func newMatchConflict(reason string, fields []string) *MatchConflictError {
	fields = append([]string(nil), fields...)
	slices.Sort(fields)
	fields = slices.Compact(fields)
	return &MatchConflictError{ReasonCode: reason, ConflictingFieldKeys: fields}
}

func matchFields(matches map[uuid.UUID][]string) []string {
	fields := make([]string, 0, 2)
	for _, matchedFields := range matches {
		fields = append(fields, matchedFields...)
	}
	return fields
}

func adaptActiveKeyClaimError(err error, fields []string) error {
	if err == nil {
		return nil
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		switch postgresError.ConstraintName {
		case "party_active_key_claims_pkey", "party_active_key_claims_record_key_unique":
			return newMatchConflict(MatchExactKeyClaimed, fields)
		}
	}
	return err
}
