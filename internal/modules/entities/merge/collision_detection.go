package merge

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
)

func (s *Store) findThirdPartyExactMatchConflictTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, identifierClass string, normalizedValue string, survivorRecordID uuid.UUID, loserRecordID uuid.UUID) (uuid.UUID, bool, error) {
	switch entityType {
	case "host":
		rows, err := tx.Query(ctx, `
SELECT record_id, aad_device_id, fqdn, hostname
  FROM hosts
 WHERE incident_id = $1
   AND host_state IN ('stub', 'canonical')
   AND record_id <> $2
   AND record_id <> $3
`, incidentID, survivorRecordID, loserRecordID)
		if err != nil {
			return uuid.UUID{}, false, fmt.Errorf("load host conflict candidates: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var (
				recordID    uuid.UUID
				aadDeviceID pgtype.Text
				fqdn        pgtype.Text
				hostname    pgtype.Text
			)
			if err := rows.Scan(&recordID, &aadDeviceID, &fqdn, &hostname); err != nil {
				return uuid.UUID{}, false, fmt.Errorf("scan host conflict candidate: %w", err)
			}
			record := hostidentity.HostRecord{
				RecordID:    recordID,
				AADDeviceID: textPointer(aadDeviceID),
				FQDN:        textPointer(fqdn),
				Hostname:    textPointer(hostname),
			}
			if s.hostIdentity.HostCanonicalNormalized(record, identifierClass) == normalizedValue {
				return recordID, true, nil
			}
		}
		if err := rows.Err(); err != nil {
			return uuid.UUID{}, false, fmt.Errorf("iterate host conflict candidates: %w", err)
		}
	case "identity":
		rows, err := tx.Query(ctx, `
SELECT record_id, aad_object_id, sid, upn, email::text, sam_account_name
  FROM identities
 WHERE incident_id = $1
   AND identity_state IN ('stub', 'canonical')
   AND record_id <> $2
   AND record_id <> $3
`, incidentID, survivorRecordID, loserRecordID)
		if err != nil {
			return uuid.UUID{}, false, fmt.Errorf("load identity conflict candidates: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var (
				recordID       uuid.UUID
				aadObjectID    pgtype.Text
				sid            pgtype.Text
				upn            pgtype.Text
				email          pgtype.Text
				samAccountName pgtype.Text
			)
			if err := rows.Scan(&recordID, &aadObjectID, &sid, &upn, &email, &samAccountName); err != nil {
				return uuid.UUID{}, false, fmt.Errorf("scan identity conflict candidate: %w", err)
			}
			record := hostidentity.IdentityRecord{
				RecordID:       recordID,
				AADObjectID:    textPointer(aadObjectID),
				SID:            textPointer(sid),
				UPN:            textPointer(upn),
				Email:          textPointer(email),
				SamAccountName: textPointer(samAccountName),
			}
			if s.hostIdentity.IdentityCanonicalNormalized(record, identifierClass) == normalizedValue {
				return recordID, true, nil
			}
		}
		if err := rows.Err(); err != nil {
			return uuid.UUID{}, false, fmt.Errorf("iterate identity conflict candidates: %w", err)
		}
	}

	rows, err := tx.Query(ctx, `
SELECT record_id
  FROM entity_preserved_identifiers
 WHERE incident_id = $1
   AND entity_type = $2
   AND identifier_type = $3
   AND normalized_value = $4
   AND classification = 'exact_match_reuse'
   AND deleted_at IS NULL
   AND record_id <> $5
   AND record_id <> $6
`, incidentID, entityType, identifierClass, normalizedValue, survivorRecordID, loserRecordID)
	if err != nil {
		return uuid.UUID{}, false, fmt.Errorf("load preserved identifier conflicts: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var recordID uuid.UUID
		if err := rows.Scan(&recordID); err != nil {
			return uuid.UUID{}, false, fmt.Errorf("scan preserved identifier conflict: %w", err)
		}
		active, err := isEntityRecordActiveTx(ctx, tx, entityType, recordID)
		if err != nil {
			return uuid.UUID{}, false, err
		}
		if active {
			return recordID, true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return uuid.UUID{}, false, fmt.Errorf("iterate preserved identifier conflicts: %w", err)
	}
	return uuid.UUID{}, false, nil
}

func isEntityRecordActiveTx(ctx context.Context, tx pgx.Tx, entityType string, recordID uuid.UUID) (bool, error) {
	switch entityType {
	case "host":
		var exists bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM hosts
     WHERE record_id = $1
       AND host_state IN ('stub', 'canonical')
)`, recordID).Scan(&exists); err != nil {
			return false, fmt.Errorf("query active host record: %w", err)
		}
		return exists, nil
	case "identity":
		var exists bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM identities
     WHERE record_id = $1
       AND identity_state IN ('stub', 'canonical')
)`, recordID).Scan(&exists); err != nil {
			return false, fmt.Errorf("query active identity record: %w", err)
		}
		return exists, nil
	default:
		return false, nil
	}
}
