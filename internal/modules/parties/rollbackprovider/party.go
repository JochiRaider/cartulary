package rollbackprovider

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type PartyProvider struct{}

var _ rollbackcontract.RowSourceProvider = PartyProvider{}

func NewPartyProvider() PartyProvider { return PartyProvider{} }

func (PartyProvider) ValidateRollbackValue(value map[string]any) error {
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

func (PartyProvider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := partySourceForRollbackValue(request.RetainedValue)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if err := (PartyProvider{}).ValidateRollbackValue(request.RetainedValue); err != nil {
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
