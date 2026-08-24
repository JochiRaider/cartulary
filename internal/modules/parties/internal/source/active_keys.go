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
	MatchCrossKeyExactMatch = "cross_key_exact_match"
	MatchExactKeyClaimed    = "exact_match_key_claimed"
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
	return uuid.Nil, false, fmt.Errorf("resolve Party active-key claim: matched owner set was empty")
}

func validateActiveKeyTransitionTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	current map[string]policy.Value,
	proposed map[string]policy.Value,
) error {
	allTuples := append(activeKeyTuples(current), activeKeyTuples(proposed)...)
	if err := lockActiveKeyTuplesTx(ctx, tx, incidentID, allTuples); err != nil {
		return err
	}
	return validateActiveKeyTupleOwnersTx(
		ctx,
		tx,
		incidentID,
		activeKeyTuples(proposed),
		map[uuid.UUID]struct{}{recordID: {}},
	)
}

// ValidateActiveKeyClaimsTx locks the union of every proposed claim set in
// UTF-8 tuple order and admits only owners in allowedOwnerIDs.
func ValidateActiveKeyClaimsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	proposedValueSets []map[string]policy.Value,
	allowedOwnerIDs map[uuid.UUID]struct{},
) error {
	var tuples []activeKeyTuple
	for _, values := range proposedValueSets {
		tuples = append(tuples, activeKeyTuples(values)...)
	}
	if err := lockActiveKeyTuplesTx(ctx, tx, incidentID, tuples); err != nil {
		return err
	}
	return validateActiveKeyTupleOwnersTx(ctx, tx, incidentID, tuples, allowedOwnerIDs)
}

func validateActiveKeyTupleOwnersTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	tuples []activeKeyTuple,
	allowedOwnerIDs map[uuid.UUID]struct{},
) error {
	var claimedFields []string
	for _, tuple := range uniqueActiveKeyTuples(tuples) {
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
		if _, allowed := allowedOwnerIDs[ownerID]; !allowed {
			claimedFields = append(claimedFields, tuple.fieldKey)
		}
	}
	if len(claimedFields) > 0 {
		return newMatchConflict(MatchExactKeyClaimed, claimedFields)
	}
	return nil
}

func SetActiveKeyClaimsDeferredTx(ctx context.Context, tx pgx.Tx, deferred bool) error {
	value := "off"
	if deferred {
		value = "on"
	}
	if _, err := tx.Exec(ctx, `SELECT set_config('cartulary.parties_defer_active_key_claims', $1, true)`, value); err != nil {
		return fmt.Errorf("set Party active-key refresh mode: %w", err)
	}
	return nil
}

func ReleaseActiveKeyClaimsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	for _, recordID := range sortedRecordIDs(recordIDs) {
		if _, err := tx.Exec(ctx, `SELECT public.parties_release_active_key_claims_v1($1)`, recordID); err != nil {
			return fmt.Errorf("release Party active-key claims: %w", err)
		}
	}
	return nil
}

func RefreshActiveKeyClaimsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	for _, recordID := range sortedRecordIDs(recordIDs) {
		if _, err := tx.Exec(ctx, `SELECT public.parties_refresh_active_key_claims_v1($1)`, recordID); err != nil {
			return adaptActiveKeyClaimError(err, []string{"party.external_ref", "party.primary_email"})
		}
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
	return uniqueActiveKeyTuples(tuples)
}

func uniqueActiveKeyTuples(tuples []activeKeyTuple) []activeKeyTuple {
	tuples = append([]activeKeyTuple(nil), tuples...)
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
	for _, tuple := range uniqueActiveKeyTuples(tuples) {
		lockIdentity := incidentID.String() + "\x1f" + tuple.keyKind + "\x1f" + tuple.normalizedValue
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, lockIdentity); err != nil {
			return fmt.Errorf("lock Party active-key tuple: %w", err)
		}
	}
	return nil
}

func sortedRecordIDs(recordIDs []uuid.UUID) []uuid.UUID {
	result := append([]uuid.UUID(nil), recordIDs...)
	slices.SortFunc(result, func(left, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	return slices.Compact(result)
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
