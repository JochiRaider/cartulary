package rollbackprovider

import (
	"context"
	"errors"
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
	if raw, present := source["host_state"]; present {
		state, valid := raw.(string)
		if !valid || !validEntityState(state) {
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
	location, hasLocation := source["location"]
	osPlatform, hasOSPlatform := source["os_platform"]
	businessOwner, hasBusinessOwner := source["business_owner"]
	criticality, hasCriticality := source["criticality"]
	containmentStatus, hasContainmentStatus := source["containment_status"]
	hostState, hasHostState := source["host_state"]
	mergedInto, hasMergedInto, err := nullableUUID(source, "merged_into_record_id")
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	var currentState string
	var currentMergedInto *uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT host_state, merged_into_record_id FROM hosts WHERE record_id = $1`, request.RecordID).Scan(&currentState, &currentMergedInto); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackcontract.ErrTargetNotFound
		}
		return err
	}
	state := currentState
	if hasHostState {
		state = hostState.(string)
	}
	hasTarget := currentMergedInto != nil
	if hasMergedInto {
		hasTarget = mergedInto != nil
	}
	if (state == "merged") != hasTarget {
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
       location = CASE WHEN $14 THEN $15::text ELSE location END,
       os_platform = CASE WHEN $16 THEN $17::text ELSE os_platform END,
       business_owner = CASE WHEN $18 THEN $19::text ELSE business_owner END,
       criticality = CASE WHEN $20 THEN $21::text ELSE criticality END,
       containment_status = CASE WHEN $22 THEN $23::text ELSE containment_status END,
       row_version = $24,
       updated_at = $25,
       updated_by_user_id = $26
 WHERE record_id = $1
`, request.RecordID, hasDisplay, display, hasHostname, hostname, hasAADDeviceID, aadDeviceID, hasFQDN, fqdn, hasHostState, hostState, hasMergedInto, mergedInto,
		hasLocation, location, hasOSPlatform, osPlatform, hasBusinessOwner, businessOwner, hasCriticality, criticality, hasContainmentStatus, containmentStatus,
		request.NextRowVersion, request.Now.UTC(), request.ActorUserID); err != nil {
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
		return nil, true, rollbackcontract.ErrTargetNotReversible
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return nil, true, err
	}
	if parsed == uuid.Nil {
		return nil, true, rollbackcontract.ErrTargetNotReversible
	}
	return parsed, true, nil
}
