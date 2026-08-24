package rollback

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/parties/internal/policy"
	partysource "github.com/JochiRaider/cartulary/internal/modules/parties/internal/source"
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
	_, err := admitRollbackSourceValues(source)
	return err
}

func (Provider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := partySourceForRollbackValue(request.RetainedValue)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if err := (Provider{}).ValidateRollbackValue(request.RetainedValue); err != nil {
		return err
	}
	admitted, err := admitRollbackSourceValues(source)
	if err != nil {
		return err
	}
	values := make([]any, 0, 16)
	for _, fieldKey := range []string{
		"party.display_name", "party.party_kind", "party.organization_name", "party.role_title",
		"party.primary_email", "party.timezone_name", "party.external_ref", "party.notes",
	} {
		value, present := admitted[fieldKey]
		values = append(values, present, rollbackDBValue(value))
	}
	_, err = tx.Exec(ctx, `
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

func (Provider) PrepareIdentifierClaimRestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.IdentifierClaimRestoreRequest) error {
	recordIDs := make([]uuid.UUID, 0, len(request.Records))
	proposedValueSets := make([]map[string]policy.Value, 0, len(request.Records))
	for _, record := range request.Records {
		if record.RecordID == uuid.Nil {
			return rollbackcontract.ErrTargetNotReversible
		}
		recordIDs = append(recordIDs, record.RecordID)
		values, err := partyRollbackValuesTx(ctx, tx, request.IncidentID, record.RecordID, record.RetainedValue)
		if err != nil {
			return err
		}
		proposedValueSets = append(proposedValueSets, values)
	}
	affected := make(map[uuid.UUID]struct{}, len(request.AffectedRecordIDs))
	for _, recordID := range request.AffectedRecordIDs {
		affected[recordID] = struct{}{}
	}
	if err := partysource.ValidateActiveKeyClaimsTx(ctx, tx, request.IncidentID, proposedValueSets, affected); err != nil {
		var matchConflict *partysource.MatchConflictError
		if errors.As(err, &matchConflict) {
			return rollbackcontract.ErrPartyExactMatchKeyClaimed
		}
		return err
	}
	if err := partysource.SetActiveKeyClaimsDeferredTx(ctx, tx, true); err != nil {
		return err
	}
	return partysource.ReleaseActiveKeyClaimsTx(ctx, tx, recordIDs)
}

func (Provider) FinalizeIdentifierClaimRestoreTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	if err := partysource.SetActiveKeyClaimsDeferredTx(ctx, tx, false); err != nil {
		return err
	}
	if err := partysource.RefreshActiveKeyClaimsTx(ctx, tx, recordIDs); err != nil {
		var matchConflict *partysource.MatchConflictError
		if errors.As(err, &matchConflict) {
			return rollbackcontract.ErrPartyExactMatchKeyClaimed
		}
		return err
	}
	return nil
}

func partyRollbackValuesTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	retained map[string]any,
) (map[string]policy.Value, error) {
	source, ok := partySourceForRollbackValue(retained)
	if !ok {
		return nil, rollbackcontract.ErrTargetNotReversible
	}
	loadedIncidentID, current, err := partysource.LoadPartyValuesForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return nil, err
	}
	if loadedIncidentID != incidentID {
		return nil, rollbackcontract.ErrTargetNotReversible
	}
	admitted, err := admitRollbackSourceValues(source)
	if err != nil {
		return nil, err
	}
	for fieldKey, value := range admitted {
		current[fieldKey] = value
	}
	return current, nil
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

func admitRollbackSourceValues(source map[string]any) (map[string]policy.Value, error) {
	admitted := make(map[string]policy.Value)
	for _, fieldKey := range policy.FieldKeys() {
		field, _ := policy.LookupField(fieldKey)
		raw, present := source[field.SourceColumn]
		if !present {
			continue
		}
		var text *string
		if raw != nil {
			value, ok := raw.(string)
			if !ok {
				return nil, rollbackcontract.ErrTargetNotReversible
			}
			text = &value
		}
		value, admissionErr := policy.AdmitStored(fieldKey, text)
		if admissionErr != nil || !canonicalStoredValue(value, text) {
			return nil, rollbackcontract.ErrTargetNotReversible
		}
		admitted[fieldKey] = value
	}
	return admitted, nil
}

func canonicalStoredValue(value policy.Value, raw *string) bool {
	stored, present := value.StoredValue()
	return (raw != nil) == present && (!present || stored == *raw)
}

func rollbackDBValue(value policy.Value) any {
	if stored, present := value.StoredValue(); present {
		return stored
	}
	return nil
}
