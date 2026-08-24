package rollback

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type Provider struct{}

var _ rollbackcontract.RowSourceProvider = Provider{}
var _ rollbackcontract.IdentifierClaimRestoreProvider = Provider{}

func NewProvider() Provider { return Provider{} }

func (Provider) ValidateRollbackValue(value map[string]any) error {
	source, ok := partySourceForRollbackValue(value)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if raw, present := source["display_name"]; present {
		text, valid := raw.(string)
		if !valid || strings.TrimSpace(text) == "" {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	if raw, present := source["party_kind"]; present {
		text, valid := raw.(string)
		if !valid || !validPartyKind(text) {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	return nil
}

func (Provider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := partySourceForRollbackValue(request.RetainedValue)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if err := (Provider{}).ValidateRollbackValue(request.RetainedValue); err != nil {
		return err
	}
	values := make([]any, 0, 18)
	for _, key := range []string{
		"display_name", "party_kind", "organization_name", "role_title", "primary_email",
		"timezone_name", "external_ref", "notes",
	} {
		value, present := source[key]
		values = append(values, present, value)
	}
	_, err := tx.Exec(ctx, `
UPDATE parties
   SET display_name = CASE WHEN $2 THEN $3::text ELSE display_name END,
       party_kind = CASE WHEN $4 THEN $5::text ELSE party_kind END,
       organization_name = CASE WHEN $6 THEN $7::text ELSE organization_name END,
       role_title = CASE WHEN $8 THEN $9::text ELSE role_title END,
       primary_email = CASE WHEN $10 THEN $11::text ELSE primary_email END,
       timezone_name = CASE WHEN $12 THEN $13::text ELSE timezone_name END,
       external_ref = CASE WHEN $14 THEN $15::text ELSE external_ref END,
       notes = CASE WHEN $16 THEN $17::text ELSE notes END,
       updated_at = $18
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	return err
}

type partyClaimTuple struct {
	keyKind         string
	normalizedValue string
}

func (Provider) PrepareIdentifierClaimRestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.IdentifierClaimRestoreRequest) error {
	unique := make(map[string]partyClaimTuple)
	recordIDs := make([]uuid.UUID, 0, len(request.Records))
	for _, record := range request.Records {
		if record.RecordID == uuid.Nil {
			return rollbackcontract.ErrTargetNotReversible
		}
		recordIDs = append(recordIDs, record.RecordID)
		tuples, err := partyRollbackClaimTuplesTx(ctx, tx, record.RecordID, record.RetainedValue)
		if err != nil {
			return err
		}
		for _, tuple := range tuples {
			unique[tuple.keyKind+"\x1f"+tuple.normalizedValue] = tuple
		}
	}
	tuples := make([]partyClaimTuple, 0, len(unique))
	for _, tuple := range unique {
		tuples = append(tuples, tuple)
	}
	slices.SortFunc(tuples, comparePartyClaimTuples)
	affected := make(map[uuid.UUID]struct{}, len(request.AffectedRecordIDs))
	for _, recordID := range request.AffectedRecordIDs {
		affected[recordID] = struct{}{}
	}
	for _, tuple := range tuples {
		lockKey := request.IncidentID.String() + "\x1f" + tuple.keyKind + "\x1f" + tuple.normalizedValue
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, lockKey); err != nil {
			return fmt.Errorf("lock rollback Party claim: %w", err)
		}
		var ownerID uuid.UUID
		err := tx.QueryRow(ctx, `
SELECT party_record_id
  FROM party_active_key_claims
 WHERE incident_id = $1
   AND key_kind = $2
   AND normalized_value = $3
`, request.IncidentID, tuple.keyKind, tuple.normalizedValue).Scan(&ownerID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("validate rollback Party claim: %w", err)
		}
		if _, ok := affected[ownerID]; !ok {
			return rollbackcontract.ErrPartyExactMatchKeyClaimed
		}
	}
	if _, err := tx.Exec(ctx, `SELECT set_config('cartulary.parties_defer_active_key_claims', 'on', true)`); err != nil {
		return fmt.Errorf("defer rollback Party claim refresh: %w", err)
	}
	slices.SortFunc(recordIDs, func(left, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	for _, recordID := range recordIDs {
		if _, err := tx.Exec(ctx, `SELECT public.parties_release_active_key_claims_v1($1)`, recordID); err != nil {
			return fmt.Errorf("release rollback Party claims: %w", err)
		}
	}
	return nil
}

func (Provider) FinalizeIdentifierClaimRestoreTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	if _, err := tx.Exec(ctx, `SELECT set_config('cartulary.parties_defer_active_key_claims', 'off', true)`); err != nil {
		return fmt.Errorf("enable rollback Party claim refresh: %w", err)
	}
	for _, recordID := range recordIDs {
		if _, err := tx.Exec(ctx, `SELECT public.parties_refresh_active_key_claims_v1($1)`, recordID); err != nil {
			var postgresError *pgconn.PgError
			if errors.As(err, &postgresError) && postgresError.Code == "23505" {
				return rollbackcontract.ErrPartyExactMatchKeyClaimed
			}
			return fmt.Errorf("refresh rollback Party claims: %w", err)
		}
	}
	return nil
}

func partyRollbackClaimTuplesTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, retained map[string]any) ([]partyClaimTuple, error) {
	source, hasRetainedSource := objectMap(retained, "source")
	if retained != nil && !hasRetainedSource {
		return nil, rollbackcontract.ErrTargetNotReversible
	}
	var email, externalRef pgtype.Text
	if err := tx.QueryRow(ctx, `SELECT primary_email, external_ref FROM parties WHERE record_id = $1`, recordID).Scan(&email, &externalRef); err != nil {
		return nil, err
	}
	values := map[string]*string{
		"primary_email": pgTextPointer(email),
		"external_ref":  pgTextPointer(externalRef),
	}
	for keyKind := range values {
		if raw, present := source[keyKind]; hasRetainedSource && present {
			if raw == nil {
				values[keyKind] = nil
				continue
			}
			text, ok := raw.(string)
			if !ok {
				return nil, rollbackcontract.ErrTargetNotReversible
			}
			values[keyKind] = &text
		}
	}
	var tuples []partyClaimTuple
	for keyKind, raw := range values {
		if raw == nil {
			continue
		}
		var normalized *string
		if err := tx.QueryRow(ctx, `SELECT public.parties_normalize_active_key_v1($1, $2)`, keyKind, *raw).Scan(&normalized); err != nil {
			return nil, err
		}
		if normalized == nil {
			return nil, rollbackcontract.ErrTargetNotReversible
		}
		tuples = append(tuples, partyClaimTuple{keyKind: keyKind, normalizedValue: *normalized})
	}
	slices.SortFunc(tuples, comparePartyClaimTuples)
	return tuples, nil
}

func comparePartyClaimTuples(left, right partyClaimTuple) int {
	if compared := strings.Compare(left.keyKind, right.keyKind); compared != 0 {
		return compared
	}
	return strings.Compare(left.normalizedValue, right.normalizedValue)
}

func pgTextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func partySourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	if source, ok := objectMap(value, "source"); ok {
		return source, len(source) > 0
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

func validPartyKind(value string) bool {
	switch value {
	case "person", "team", "organization", "distribution_list", "other":
		return true
	default:
		return false
	}
}
