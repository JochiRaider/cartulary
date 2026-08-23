package hostidentity

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func insertHostTx(ctx context.Context, tx pgx.Tx, record *HostRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO hosts (
    record_id,
    incident_id,
    display_name,
    aad_device_id,
    fqdn,
    hostname,
    location,
    os_platform,
    business_owner,
    criticality,
    containment_status,
    host_state,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $16, $17, $17)
RETURNING record_id
`, record.RecordID, record.IncidentID, record.DisplayName, record.AADDeviceID, record.FQDN, record.Hostname, record.Location, record.OSPlatform, record.BusinessOwner, record.Criticality, record.ContainmentStatus, record.HostState, record.EntityOrigin, record.SeedMentionID, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
}

func updateHostTx(ctx context.Context, tx pgx.Tx, record HostRecord) error {
	_, err := tx.Exec(ctx, `
UPDATE hosts
   SET display_name = $2,
       aad_device_id = $3,
       fqdn = $4,
       hostname = $5,
       location = $6,
       os_platform = $7,
       business_owner = $8,
       criticality = $9,
       containment_status = $10,
       host_state = $11,
       merged_into_record_id = $12,
       row_version = $13,
       updated_at = $14,
       updated_by_user_id = $15
 WHERE record_id = $1
`, record.RecordID, record.DisplayName, record.AADDeviceID, record.FQDN, record.Hostname, record.Location, record.OSPlatform, record.BusinessOwner, record.Criticality, record.ContainmentStatus, record.HostState, record.MergedIntoRecordID, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
	if err != nil {
		return fmt.Errorf("update host: %w", err)
	}
	return nil
}

func insertIdentityTx(ctx context.Context, tx pgx.Tx, record *IdentityRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO identities (
    record_id,
    incident_id,
    display_name,
    aad_object_id,
    sid,
    upn,
    email,
    sam_account_name,
    privilege_level,
    mfa_state,
    reset_status,
    identity_state,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $16, $17, $17)
RETURNING record_id
`, record.RecordID, record.IncidentID, record.DisplayName, record.AADObjectID, record.SID, record.UPN, record.Email, record.SamAccountName, record.PrivilegeLevel, record.MFAState, record.ResetStatus, record.IdentityState, record.EntityOrigin, record.SeedMentionID, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
}

func updateIdentityTx(ctx context.Context, tx pgx.Tx, record IdentityRecord) error {
	_, err := tx.Exec(ctx, `
UPDATE identities
   SET display_name = $2,
       aad_object_id = $3,
       sid = $4,
       upn = $5,
       email = $6,
       sam_account_name = $7,
       privilege_level = $8,
       mfa_state = $9,
       reset_status = $10,
       identity_state = $11,
       merged_into_record_id = $12,
       row_version = $13,
       updated_at = $14,
       updated_by_user_id = $15
 WHERE record_id = $1
`, record.RecordID, record.DisplayName, record.AADObjectID, record.SID, record.UPN, record.Email, record.SamAccountName, record.PrivilegeLevel, record.MFAState, record.ResetStatus, record.IdentityState, record.MergedIntoRecordID, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
	if err != nil {
		return fmt.Errorf("update identity: %w", err)
	}
	return nil
}

func loadHostByRecordIDTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (HostRecord, error) {
	record, err := scanHostRecord(tx.QueryRow(ctx, `
SELECT
    h.record_id,
    h.incident_id,
    h.display_name,
    h.aad_device_id,
    h.fqdn,
    h.hostname,
    h.location,
    h.os_platform,
    h.business_owner,
    h.criticality,
    h.containment_status,
    h.host_state,
    h.merged_into_record_id,
    h.entity_origin,
    h.seed_entity_mention_id,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id
  FROM hosts h
 JOIN records r
    ON r.record_id = h.record_id
 WHERE h.record_id = $1
 FOR UPDATE OF h, r
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return HostRecord{}, ErrHostIdentityRecordNotFound
	}
	if err != nil {
		return HostRecord{}, fmt.Errorf("load host by record id: %w", err)
	}
	return record, nil
}

func loadIdentityByRecordIDTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (IdentityRecord, error) {
	record, err := scanIdentityRecord(tx.QueryRow(ctx, `
SELECT
    i.record_id,
    i.incident_id,
    i.display_name,
    i.aad_object_id,
    i.sid,
    i.upn,
    i.email::text,
    i.sam_account_name,
    i.privilege_level,
    i.mfa_state,
    i.reset_status,
    i.identity_state,
    i.merged_into_record_id,
    i.entity_origin,
    i.seed_entity_mention_id,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id
  FROM identities i
 JOIN records r
    ON r.record_id = i.record_id
 WHERE i.record_id = $1
 FOR UPDATE OF i, r
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return IdentityRecord{}, ErrHostIdentityRecordNotFound
	}
	if err != nil {
		return IdentityRecord{}, fmt.Errorf("load identity by record id: %w", err)
	}
	return record, nil
}

type entityAliasQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

type entityRowHydration struct {
	AliasesByRecord             map[uuid.UUID][]AliasValue
	ReusableIdentifiersByRecord map[uuid.UUID][]ReusableIdentifier
}

func loadEntityRowHydrationByRecord(ctx context.Context, querier entityAliasQueryer, incidentID uuid.UUID, entityType string, recordIDs []uuid.UUID) (entityRowHydration, error) {
	aliasesByRecord, err := loadEntityAliasesByRecord(ctx, querier, incidentID, entityType, recordIDs)
	if err != nil {
		return entityRowHydration{}, err
	}
	reusableIdentifiersByRecord, err := loadEntityReusableIdentifiersByRecord(ctx, querier, incidentID, entityType, recordIDs)
	if err != nil {
		return entityRowHydration{}, err
	}
	return entityRowHydration{
		AliasesByRecord:             aliasesByRecord,
		ReusableIdentifiersByRecord: reusableIdentifiersByRecord,
	}, nil
}

func applyHostRowHydration(record *HostRecord, hydration entityRowHydration) {
	record.SuggestionOnlyAliases = hydration.AliasesByRecord[record.RecordID]
	record.ReusableIdentifiers = hydration.ReusableIdentifiersByRecord[record.RecordID]
}

func applyIdentityRowHydration(record *IdentityRecord, hydration entityRowHydration) {
	record.SuggestionOnlyAliases = hydration.AliasesByRecord[record.RecordID]
	record.ReusableIdentifiers = hydration.ReusableIdentifiersByRecord[record.RecordID]
}

func hydrateHostRecordTx(ctx context.Context, tx pgx.Tx, record *HostRecord) error {
	aliases, err := loadEntityAliasesTx(ctx, tx, record.RecordID, "host")
	if err != nil {
		return err
	}
	reusableIdentifiers, err := loadEntityReusableIdentifiersTx(ctx, tx, record.RecordID, "host")
	if err != nil {
		return err
	}
	record.SuggestionOnlyAliases = aliases
	record.ReusableIdentifiers = reusableIdentifiers
	return nil
}

func hydrateIdentityRecordTx(ctx context.Context, tx pgx.Tx, record *IdentityRecord) error {
	aliases, err := loadEntityAliasesTx(ctx, tx, record.RecordID, "identity")
	if err != nil {
		return err
	}
	reusableIdentifiers, err := loadEntityReusableIdentifiersTx(ctx, tx, record.RecordID, "identity")
	if err != nil {
		return err
	}
	record.SuggestionOnlyAliases = aliases
	record.ReusableIdentifiers = reusableIdentifiers
	return nil
}

func loadEntityAliasesByRecord(ctx context.Context, querier entityAliasQueryer, incidentID uuid.UUID, entityType string, recordIDs []uuid.UUID) (map[uuid.UUID][]AliasValue, error) {
	if len(recordIDs) == 0 {
		return map[uuid.UUID][]AliasValue{}, nil
	}
	args := []any{incidentID, entityType}
	rows, err := querier.Query(ctx, `
SELECT record_id, entity_alias_id, normalized_text::text
  FROM entity_aliases
 WHERE incident_id = $1
   AND entity_type = $2
   AND record_id IN (`+bindUUIDList(&args, recordIDs)+`)
   AND deleted_at IS NULL
 ORDER BY record_id ASC, normalized_text ASC, created_at ASC, entity_alias_id ASC
`, args...)
	if err != nil {
		return nil, fmt.Errorf("query entity aliases by record: %w", err)
	}
	defer rows.Close()

	aliasesByRecord := make(map[uuid.UUID][]AliasValue)
	for rows.Next() {
		var (
			recordID uuid.UUID
			alias    AliasValue
		)
		if err := rows.Scan(&recordID, &alias.EntityAliasID, &alias.AliasText); err != nil {
			return nil, fmt.Errorf("scan entity alias by record: %w", err)
		}
		aliasesByRecord[recordID] = append(aliasesByRecord[recordID], alias)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity aliases by record: %w", err)
	}

	return aliasesByRecord, nil
}

func loadEntityReusableIdentifiersByRecord(ctx context.Context, querier entityAliasQueryer, incidentID uuid.UUID, entityType string, recordIDs []uuid.UUID) (map[uuid.UUID][]ReusableIdentifier, error) {
	if len(recordIDs) == 0 {
		return map[uuid.UUID][]ReusableIdentifier{}, nil
	}
	args := []any{incidentID, entityType}
	rows, err := querier.Query(ctx, `
SELECT
    record_id,
    entity_preserved_identifier_id,
    identifier_type,
    raw_value,
    normalized_value
  FROM entity_preserved_identifiers
 WHERE incident_id = $1
   AND entity_type = $2
   AND record_id IN (`+bindUUIDList(&args, recordIDs)+`)
   AND classification = 'exact_match_reuse'
   AND deleted_at IS NULL
 ORDER BY record_id ASC, identifier_type ASC, normalized_value ASC, created_at ASC, entity_preserved_identifier_id ASC
`, args...)
	if err != nil {
		return nil, fmt.Errorf("query entity reusable identifiers by record: %w", err)
	}
	defer rows.Close()

	identifiersByRecord := make(map[uuid.UUID][]ReusableIdentifier)
	for rows.Next() {
		var (
			recordID uuid.UUID
			item     ReusableIdentifier
		)
		if err := rows.Scan(&recordID, &item.EntityPreservedIdentifierID, &item.IdentifierClass, &item.RawValue, &item.NormalizedValue); err != nil {
			return nil, fmt.Errorf("scan entity reusable identifier by record: %w", err)
		}
		identifiersByRecord[recordID] = append(identifiersByRecord[recordID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity reusable identifiers by record: %w", err)
	}
	return identifiersByRecord, nil
}

func bindUUIDList(args *[]any, recordIDs []uuid.UUID) string {
	placeholders := make([]string, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		*args = append(*args, recordID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(*args)))
	}
	return strings.Join(placeholders, ", ")
}

func loadEntityReusableIdentifiersTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) ([]ReusableIdentifier, error) {
	rows, err := tx.Query(ctx, `
SELECT
    entity_preserved_identifier_id,
    identifier_type,
    raw_value,
    normalized_value
  FROM entity_preserved_identifiers
 WHERE record_id = $1
   AND entity_type = $2
   AND classification = 'exact_match_reuse'
   AND deleted_at IS NULL
 ORDER BY identifier_type ASC, normalized_value ASC, created_at ASC, entity_preserved_identifier_id ASC
`, recordID, entityType)
	if err != nil {
		return nil, fmt.Errorf("load entity reusable identifiers: %w", err)
	}
	defer rows.Close()

	identifiers := make([]ReusableIdentifier, 0)
	for rows.Next() {
		var item ReusableIdentifier
		if err := rows.Scan(&item.EntityPreservedIdentifierID, &item.IdentifierClass, &item.RawValue, &item.NormalizedValue); err != nil {
			return nil, fmt.Errorf("scan entity reusable identifier: %w", err)
		}
		identifiers = append(identifiers, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity reusable identifiers: %w", err)
	}
	return identifiers, nil
}
