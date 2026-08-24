package rollbackprovider

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
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

type claimTuple struct {
	identifierType  string
	normalizedValue string
}

func prepareIdentifierClaimRestoreTx(ctx context.Context, tx pgx.Tx, entityType string, request rollbackcontract.IdentifierClaimRestoreRequest) error {
	uniqueTuples := make(map[string]claimTuple)
	recordIDs := make([]uuid.UUID, 0, len(request.Records))
	for _, record := range request.Records {
		if record.RecordID == uuid.Nil {
			return rollbackcontract.ErrTargetNotReversible
		}
		recordIDs = append(recordIDs, record.RecordID)
		tuples, err := rollbackIdentifierTuplesTx(ctx, tx, entityType, record.RecordID, record.RetainedValue)
		if err != nil {
			return err
		}
		for _, tuple := range tuples {
			uniqueTuples[tuple.identifierType+"\x1f"+tuple.normalizedValue] = tuple
		}
	}
	tuples := make([]claimTuple, 0, len(uniqueTuples))
	for _, tuple := range uniqueTuples {
		tuples = append(tuples, tuple)
	}
	slices.SortFunc(tuples, func(left claimTuple, right claimTuple) int {
		if compared := strings.Compare(left.identifierType, right.identifierType); compared != 0 {
			return compared
		}
		return strings.Compare(left.normalizedValue, right.normalizedValue)
	})
	affected := make(map[uuid.UUID]struct{}, len(request.AffectedRecordIDs))
	for _, recordID := range request.AffectedRecordIDs {
		affected[recordID] = struct{}{}
	}
	for _, tuple := range tuples {
		lockKey := request.IncidentID.String() + "\x1f" + entityType + "\x1f" + tuple.identifierType + "\x1f" + tuple.normalizedValue
		if _, err := tx.Exec(ctx, `SELECT pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended($1, 0))`, lockKey); err != nil {
			return fmt.Errorf("lock rollback entity identifier claim: %w", err)
		}
		var claimedRecordID uuid.UUID
		err := tx.QueryRow(ctx, `
SELECT record_id
  FROM entity_active_identifier_claims
 WHERE incident_id = $1
   AND entity_type = $2
   AND identifier_type = $3
   AND normalized_value = $4
`, request.IncidentID, entityType, tuple.identifierType, tuple.normalizedValue).Scan(&claimedRecordID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("validate rollback entity identifier claim: %w", err)
		}
		if _, ok := affected[claimedRecordID]; !ok {
			return rollbackcontract.ErrEntityIdentifierConflict
		}
	}
	if _, err := tx.Exec(ctx, `SELECT pg_catalog.set_config('cartulary.entities_defer_active_identifier_claims', 'on', true)`); err != nil {
		return fmt.Errorf("defer rollback entity identifier claim refresh: %w", err)
	}
	slices.SortFunc(recordIDs, func(left uuid.UUID, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	for _, recordID := range recordIDs {
		if _, err := tx.Exec(ctx, `SELECT public.entities_release_active_identifier_claims_v1($1)`, recordID); err != nil {
			return fmt.Errorf("release rollback entity identifier claims: %w", err)
		}
	}
	return nil
}

func finalizeIdentifierClaimRestoreTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	for _, recordID := range recordIDs {
		if _, err := tx.Exec(ctx, `SELECT public.entities_refresh_active_identifier_claims_v1($1)`, recordID); err != nil {
			var postgresError *pgconn.PgError
			if errors.As(err, &postgresError) && postgresError.Code == "23505" {
				return rollbackcontract.ErrEntityIdentifierConflict
			}
			return fmt.Errorf("refresh rollback entity identifier claims: %w", err)
		}
	}
	return nil
}

func rollbackIdentifierTuplesTx(ctx context.Context, tx pgx.Tx, entityType string, recordID uuid.UUID, retained map[string]any) ([]claimTuple, error) {
	source, hasRetainedSource := objectMap(retained, "source")
	if retained != nil && !hasRetainedSource {
		return nil, rollbackcontract.ErrTargetNotReversible
	}
	values := map[string]*string{}
	switch entityType {
	case "host":
		var aadDeviceID, fqdn, hostname pgtype.Text
		if err := tx.QueryRow(ctx, `SELECT aad_device_id, fqdn, hostname FROM hosts WHERE record_id = $1`, recordID).Scan(&aadDeviceID, &fqdn, &hostname); err != nil {
			return nil, err
		}
		values["aad_device_id"] = pgTextPointer(aadDeviceID)
		values["fqdn"] = pgTextPointer(fqdn)
		values["hostname"] = pgTextPointer(hostname)
	case "identity":
		var aadObjectID, sid, upn, email, samAccountName pgtype.Text
		if err := tx.QueryRow(ctx, `SELECT aad_object_id, sid, upn, email::text, sam_account_name FROM identities WHERE record_id = $1`, recordID).Scan(&aadObjectID, &sid, &upn, &email, &samAccountName); err != nil {
			return nil, err
		}
		values["aad_object_id"] = pgTextPointer(aadObjectID)
		values["sid"] = pgTextPointer(sid)
		values["upn"] = pgTextPointer(upn)
		values["email"] = pgTextPointer(email)
		values["sam_account_name"] = pgTextPointer(samAccountName)
	default:
		return nil, rollbackcontract.ErrTargetNotReversible
	}
	for identifierType := range values {
		if raw, present := source[identifierType]; hasRetainedSource && present {
			if raw == nil {
				values[identifierType] = nil
				continue
			}
			text, valid := raw.(string)
			if !valid {
				return nil, rollbackcontract.ErrTargetNotReversible
			}
			values[identifierType] = &text
		}
	}
	unique := map[string]claimTuple{}
	for identifierType, raw := range values {
		if raw == nil {
			continue
		}
		normalized, valid := fieldnorm.NormalizeIdentifier(identifierType, *raw)
		if !valid {
			return nil, rollbackcontract.ErrTargetNotReversible
		}
		tuple := claimTuple{identifierType: identifierType, normalizedValue: normalized}
		unique[identifierType+"\x1f"+normalized] = tuple
	}
	rows, err := tx.Query(ctx, `
SELECT identifier_type, normalized_value
  FROM entity_preserved_identifiers
 WHERE record_id = $1
   AND entity_type = $2
   AND classification = 'exact_match_reuse'
   AND deleted_at IS NULL
`, recordID, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tuple claimTuple
		if err := rows.Scan(&tuple.identifierType, &tuple.normalizedValue); err != nil {
			return nil, err
		}
		unique[tuple.identifierType+"\x1f"+tuple.normalizedValue] = tuple
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	tuples := make([]claimTuple, 0, len(unique))
	for _, tuple := range unique {
		tuples = append(tuples, tuple)
	}
	slices.SortFunc(tuples, func(left claimTuple, right claimTuple) int {
		if compared := strings.Compare(left.identifierType, right.identifierType); compared != 0 {
			return compared
		}
		return strings.Compare(left.normalizedValue, right.normalizedValue)
	})
	return tuples, nil
}

func pgTextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
