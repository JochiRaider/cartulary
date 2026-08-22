package hostidentity

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func (s *Store) QueryHostRowsPage(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
	if s.pool == nil {
		return querypage.Result{}, fmt.Errorf("query host rows: store pool is nil")
	}
	if s.projectionReader == nil {
		return querypage.Result{}, fmt.Errorf("query host rows: projection reader is required")
	}
	projections, err := s.projectionReader.SelectHostQueryProjections(ctx, incidentID, query, window)
	if err != nil {
		return querypage.Result{}, err
	}
	recordIDs := make([]uuid.UUID, 0, len(projections))
	for _, projection := range projections {
		recordIDs = append(recordIDs, projection.RecordID)
	}
	recordsByID, err := loadHostSourceRecordsByID(ctx, s.pool, incidentID, recordIDs)
	if err != nil {
		return querypage.Result{}, err
	}
	hydration, err := loadEntityRowHydrationByRecord(ctx, s.pool, incidentID, "host", recordIDs)
	if err != nil {
		return querypage.Result{}, err
	}
	result := make([]map[string]any, 0, len(projections))
	for _, projection := range projections {
		record, ok := recordsByID[projection.RecordID]
		if !ok {
			return querypage.Result{}, fmt.Errorf("hydrate host projection %s: authoritative source row is missing", projection.RecordID)
		}
		applyHostProjection(&record, projection)
		applyHostRowHydration(&record, hydration)
		result = append(result, buildHostRow(record))
	}
	return querypage.Finish(result, window.Limit), nil
}

func loadHostSourceRecordsByID(
	ctx context.Context,
	querier entityAliasQueryer,
	incidentID uuid.UUID,
	recordIDs []uuid.UUID,
) (map[uuid.UUID]HostRecord, error) {
	if len(recordIDs) == 0 {
		return map[uuid.UUID]HostRecord{}, nil
	}
	rows, err := querier.Query(ctx, `
SELECT
    h.record_id,
    h.incident_id,
    h.display_name,
    h.aad_device_id,
    h.fqdn,
    h.hostname,
    h.merged_into_record_id,
    h.entity_origin,
    h.seed_entity_mention_id,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id
  FROM hosts h
  JOIN records r ON r.record_id = h.record_id
 WHERE h.incident_id = $1
   AND h.record_id = ANY($2::uuid[])
   AND r.deleted_at IS NULL
`, incidentID, recordIDs)
	if err != nil {
		return nil, fmt.Errorf("load bounded host source rows: %w", err)
	}
	defer rows.Close()

	records := make(map[uuid.UUID]HostRecord, len(recordIDs))
	for rows.Next() {
		record, scanErr := scanHostSourceRecord(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		records[record.RecordID] = record
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bounded host source rows: %w", err)
	}
	return records, nil
}

func scanHostSourceRecord(scanner interface{ Scan(...any) error }) (HostRecord, error) {
	var (
		record      HostRecord
		aadDeviceID pgtype.Text
		fqdn        pgtype.Text
		hostname    pgtype.Text
		mergedInto  pgtype.UUID
		seedMention pgtype.UUID
	)
	if err := scanner.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DisplayName,
		&aadDeviceID,
		&fqdn,
		&hostname,
		&mergedInto,
		&record.EntityOrigin,
		&seedMention,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedByUser,
		&record.UpdatedByUser,
	); err != nil {
		return HostRecord{}, fmt.Errorf("scan bounded host source row: %w", err)
	}
	record.AADDeviceID = textPointer(aadDeviceID)
	record.FQDN = textPointer(fqdn)
	record.Hostname = textPointer(hostname)
	record.MergedIntoRecordID = uuidPointerFromPG(mergedInto)
	record.SeedMentionID = uuidPointerFromPG(seedMention)
	return record, nil
}

func applyHostProjection(record *HostRecord, projection workbookprojection.HostQueryProjection) {
	record.HostState = projection.HostState
	record.LinkedEventCount = projection.LinkedEventCount
	record.EvidenceCount = projection.EvidenceCount
	record.Location = projection.Location
	record.OSPlatform = projection.OSPlatform
	record.BusinessOwner = projection.BusinessOwner
	record.Criticality = projection.Criticality
	record.ContainmentStatus = projection.ContainmentStatus
}

func (s *Store) QueryIdentityRowsPage(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
	if s.pool == nil {
		return querypage.Result{}, fmt.Errorf("query identity rows: store pool is nil")
	}
	if s.projectionReader == nil {
		return querypage.Result{}, fmt.Errorf("query identity rows: projection reader is required")
	}
	projections, err := s.projectionReader.SelectIdentityQueryProjections(ctx, incidentID, query, window)
	if err != nil {
		return querypage.Result{}, err
	}
	recordIDs := make([]uuid.UUID, 0, len(projections))
	for _, projection := range projections {
		recordIDs = append(recordIDs, projection.RecordID)
	}
	recordsByID, err := loadIdentitySourceRecordsByID(ctx, s.pool, incidentID, recordIDs)
	if err != nil {
		return querypage.Result{}, err
	}
	hydration, err := loadEntityRowHydrationByRecord(ctx, s.pool, incidentID, "identity", recordIDs)
	if err != nil {
		return querypage.Result{}, err
	}
	result := make([]map[string]any, 0, len(projections))
	for _, projection := range projections {
		record, ok := recordsByID[projection.RecordID]
		if !ok {
			return querypage.Result{}, fmt.Errorf("hydrate identity projection %s: authoritative source row is missing", projection.RecordID)
		}
		applyIdentityProjection(&record, projection)
		applyIdentityRowHydration(&record, hydration)
		result = append(result, buildIdentityRow(record))
	}
	return querypage.Finish(result, window.Limit), nil
}

func loadIdentitySourceRecordsByID(
	ctx context.Context,
	querier entityAliasQueryer,
	incidentID uuid.UUID,
	recordIDs []uuid.UUID,
) (map[uuid.UUID]IdentityRecord, error) {
	if len(recordIDs) == 0 {
		return map[uuid.UUID]IdentityRecord{}, nil
	}
	rows, err := querier.Query(ctx, `
SELECT
    i.record_id,
    i.incident_id,
    i.display_name,
    i.aad_object_id,
    i.sid,
    i.upn,
    i.email::text,
    i.sam_account_name,
    i.merged_into_record_id,
    i.entity_origin,
    i.seed_entity_mention_id,
    r.row_version,
    r.created_at,
    r.updated_at,
    r.created_by_user_id,
    r.updated_by_user_id
  FROM identities i
  JOIN records r ON r.record_id = i.record_id
 WHERE i.incident_id = $1
   AND i.record_id = ANY($2::uuid[])
   AND r.deleted_at IS NULL
`, incidentID, recordIDs)
	if err != nil {
		return nil, fmt.Errorf("load bounded identity source rows: %w", err)
	}
	defer rows.Close()

	records := make(map[uuid.UUID]IdentityRecord, len(recordIDs))
	for rows.Next() {
		record, scanErr := scanIdentitySourceRecord(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		records[record.RecordID] = record
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bounded identity source rows: %w", err)
	}
	return records, nil
}

func scanIdentitySourceRecord(scanner interface{ Scan(...any) error }) (IdentityRecord, error) {
	var (
		record         IdentityRecord
		aadObjectID    pgtype.Text
		sid            pgtype.Text
		upn            pgtype.Text
		email          pgtype.Text
		samAccountName pgtype.Text
		mergedInto     pgtype.UUID
		seedMention    pgtype.UUID
	)
	if err := scanner.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DisplayName,
		&aadObjectID,
		&sid,
		&upn,
		&email,
		&samAccountName,
		&mergedInto,
		&record.EntityOrigin,
		&seedMention,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedByUser,
		&record.UpdatedByUser,
	); err != nil {
		return IdentityRecord{}, fmt.Errorf("scan bounded identity source row: %w", err)
	}
	record.AADObjectID = textPointer(aadObjectID)
	record.SID = textPointer(sid)
	record.UPN = textPointer(upn)
	record.Email = textPointer(email)
	record.SamAccountName = textPointer(samAccountName)
	record.MergedIntoRecordID = uuidPointerFromPG(mergedInto)
	record.SeedMentionID = uuidPointerFromPG(seedMention)
	return record, nil
}

func applyIdentityProjection(record *IdentityRecord, projection workbookprojection.IdentityQueryProjection) {
	record.IdentityState = projection.IdentityState
	record.LinkedEventCount = projection.LinkedEventCount
	record.EvidenceCount = projection.EvidenceCount
	record.PrivilegeLevel = projection.PrivilegeLevel
	record.MFAState = projection.MFAState
	record.ResetStatus = projection.ResetStatus
}
