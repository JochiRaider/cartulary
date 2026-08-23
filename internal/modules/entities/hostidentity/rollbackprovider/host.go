package rollbackprovider

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type HostProvider struct{}

var _ rollbackcontract.RowSourceProvider = HostProvider{}
var _ rollbackcontract.IdentifierClaimRestoreProvider = HostProvider{}

func NewHostProvider() HostProvider { return HostProvider{} }

func (HostProvider) PrepareIdentifierClaimRestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.IdentifierClaimRestoreRequest) error {
	return prepareIdentifierClaimRestoreTx(ctx, tx, "host", request)
}

func (HostProvider) FinalizeIdentifierClaimRestoreTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	return finalizeIdentifierClaimRestoreTx(ctx, tx, recordIDs)
}

func (HostProvider) ValidateRollbackValue(value map[string]any) error {
	source, ok := hostSourceForRollbackValue(value)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if raw, present := source["display_name"]; present {
		text, valid := raw.(string)
		if !valid || strings.TrimSpace(text) == "" {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	if _, _, err := nullableUUID(source, "merged_into_record_id"); err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
}

func (HostProvider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := hostSourceForRollbackValue(request.RetainedValue)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if err := (HostProvider{}).ValidateRollbackValue(request.RetainedValue); err != nil {
		return err
	}
	display, hasDisplay := source["display_name"]
	hostname, hasHostname := source["hostname"]
	aadDeviceID, hasAADDeviceID := source["aad_device_id"]
	fqdn, hasFQDN := source["fqdn"]
	hostState, hasHostState := source["host_state"]
	mergedInto, hasMergedInto, err := nullableUUID(source, "merged_into_record_id")
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	if _, err := tx.Exec(ctx, `
UPDATE hosts
   SET display_name = CASE WHEN $2 THEN $3::text ELSE display_name END,
       hostname = CASE WHEN $4 THEN $5::text ELSE hostname END,
       aad_device_id = CASE WHEN $6 THEN $7::text ELSE aad_device_id END,
       fqdn = CASE WHEN $8 THEN $9::text ELSE fqdn END,
       host_state = CASE WHEN $10 THEN $11::text ELSE host_state END,
       merged_into_record_id = CASE WHEN $12 THEN $13::uuid ELSE merged_into_record_id END,
       row_version = $14,
       updated_at = $15,
       updated_by_user_id = $16
 WHERE record_id = $1
`, request.RecordID, hasDisplay, display, hasHostname, hostname, hasAADDeviceID, aadDeviceID, hasFQDN, fqdn, hasHostState, hostState, hasMergedInto, mergedInto, request.NextRowVersion, request.Now.UTC(), request.ActorUserID); err != nil {
		return err
	}
	return nil
}

func hostSourceForRollbackValue(value map[string]any) (map[string]any, bool) {
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

func nullableUUID(value map[string]any, key string) (any, bool, error) {
	raw, present := value[key]
	if !present || raw == nil {
		return nil, present, nil
	}
	text, ok := raw.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return nil, true, nil
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return nil, true, err
	}
	return parsed, true, nil
}
