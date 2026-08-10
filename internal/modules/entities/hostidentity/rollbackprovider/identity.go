package rollbackprovider

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type IdentityProvider struct{}

var _ rollbackcontract.RowSourceProvider = IdentityProvider{}

func NewIdentityProvider() IdentityProvider { return IdentityProvider{} }

func (IdentityProvider) ValidateRollbackValue(value map[string]any) error {
	source, ok := identitySourceForRollbackValue(value)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if raw, present := source["display_name"]; present {
		text, valid := raw.(string)
		if !valid || strings.TrimSpace(text) == "" {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	if raw, present := source["identity_state"]; present {
		state, valid := raw.(string)
		if !valid || !validIdentityState(state) {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	if _, _, err := nullableUUID(source, "merged_into_record_id"); err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
}

func (IdentityProvider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := identitySourceForRollbackValue(request.RetainedValue)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if err := (IdentityProvider{}).ValidateRollbackValue(request.RetainedValue); err != nil {
		return err
	}
	var currentState string
	var currentMergedInto *uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT identity_state, merged_into_record_id FROM identities WHERE record_id = $1`, request.RecordID).Scan(&currentState, &currentMergedInto); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackcontract.ErrTargetNotFound
		}
		return err
	}
	state := currentState
	if raw, present := source["identity_state"]; present {
		state = raw.(string)
	}
	mergedInto := currentMergedInto
	parsedMergedInto, hasMergedInto, err := nullableUUID(source, "merged_into_record_id")
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	if hasMergedInto {
		mergedInto = nil
		if parsed, valid := parsedMergedInto.(uuid.UUID); valid {
			mergedInto = &parsed
		}
	}
	if (state == "merged") != (mergedInto != nil) {
		return rollbackcontract.ErrTargetNotReversible
	}
	values := make([]any, 0, 18)
	for _, key := range []string{
		"display_name", "upn", "email", "sam_account_name", "aad_object_id", "sid",
		"privilege_level", "mfa_state", "reset_status",
	} {
		value, present := source[key]
		values = append(values, present, value)
	}
	_, err = tx.Exec(ctx, `
UPDATE identities
   SET display_name = CASE WHEN $2 THEN $3::text ELSE display_name END,
       upn = CASE WHEN $4 THEN $5::text ELSE upn END,
       email = CASE WHEN $6 THEN $7::text ELSE email END,
       sam_account_name = CASE WHEN $8 THEN $9::text ELSE sam_account_name END,
       aad_object_id = CASE WHEN $10 THEN $11::text ELSE aad_object_id END,
       sid = CASE WHEN $12 THEN $13::text ELSE sid END,
       privilege_level = CASE WHEN $14 THEN $15::text ELSE privilege_level END,
       mfa_state = CASE WHEN $16 THEN $17::text ELSE mfa_state END,
       reset_status = CASE WHEN $18 THEN $19::text ELSE reset_status END,
       identity_state = CASE WHEN $20 THEN $21::text ELSE identity_state END,
       merged_into_record_id = CASE WHEN $22 THEN $23::uuid ELSE merged_into_record_id END,
       row_version = $24,
       updated_at = $25,
       updated_by_user_id = $26
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values,
		mapHas(source, "identity_state"), state,
		hasMergedInto, mergedInto,
		request.NextRowVersion, request.Now.UTC(), request.ActorUserID,
	)...)...)
	return err
}

func identitySourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	if source, ok := objectMap(value, "source"); ok {
		return source, len(source) > 0
	}
	return nil, false
}

func validIdentityState(value string) bool {
	switch value {
	case "stub", "canonical", "merged":
		return true
	default:
		return false
	}
}

func mapHas(value map[string]any, key string) bool {
	_, present := value[key]
	return present
}
