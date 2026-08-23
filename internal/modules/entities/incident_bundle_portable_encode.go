package entities

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func encodePortableHosts(ctx context.Context, exportContext sourceport.ExportContext, rows []portableHostRow) ([]byte, error) {
	ids := entityHostIDs(rows)
	created := entityPortableActorMap(ctx, exportContext, "hosts", "created_by_user_id", ids)
	updated := entityPortableActorMap(ctx, exportContext, "hosts", "updated_by_user_id", ids)
	if created.err != nil || updated.err != nil {
		return nil, errors.New("entity host portability attribution resolution failed")
	}
	var payload bytes.Buffer
	for _, row := range rows {
		createdBy, err := entityPortableActorID(created.values[row.RecordID.String()], row.RuntimeCreatedByID)
		if err != nil {
			return nil, err
		}
		updatedBy, err := entityPortableActorID(updated.values[row.RecordID.String()], row.RuntimeUpdatedByID)
		if err != nil {
			return nil, err
		}
		if err := appendEntityPortableRow(&payload, map[string]any{
			"record_id": row.RecordID.String(), "incident_id": row.IncidentID.String(), "display_name": row.DisplayName,
			"hostname": row.Hostname, "aad_device_id": row.AADDeviceID, "fqdn": row.FQDN, "entity_origin": row.EntityOrigin,
			"seed_entity_mention_id": entityNullableUUIDString(row.SeedMentionID), "host_state": row.State,
			"merged_into_record_id": entityNullableUUIDString(row.MergedIntoRecordID), "row_version": row.RowVersion,
			"created_at": entityFormatTimestamp(row.CreatedAt), "updated_at": entityFormatTimestamp(row.UpdatedAt),
			"created_by_user_id": createdBy, "updated_by_user_id": updatedBy, "location": row.Location,
			"os_platform": row.OSPlatform, "business_owner": row.BusinessOwner, "criticality": row.Criticality,
			"containment_status": row.ContainmentStatus,
		}); err != nil {
			return nil, err
		}
	}
	return payload.Bytes(), nil
}

func encodePortableIdentities(ctx context.Context, exportContext sourceport.ExportContext, rows []portableIdentityRow) ([]byte, error) {
	ids := entityIdentityIDs(rows)
	created := entityPortableActorMap(ctx, exportContext, "identities", "created_by_user_id", ids)
	updated := entityPortableActorMap(ctx, exportContext, "identities", "updated_by_user_id", ids)
	if created.err != nil || updated.err != nil {
		return nil, errors.New("entity identity portability attribution resolution failed")
	}
	var payload bytes.Buffer
	for _, row := range rows {
		createdBy, err := entityPortableActorID(created.values[row.RecordID.String()], row.RuntimeCreatedByID)
		if err != nil {
			return nil, err
		}
		updatedBy, err := entityPortableActorID(updated.values[row.RecordID.String()], row.RuntimeUpdatedByID)
		if err != nil {
			return nil, err
		}
		if err := appendEntityPortableRow(&payload, map[string]any{
			"record_id": row.RecordID.String(), "incident_id": row.IncidentID.String(), "display_name": row.DisplayName,
			"upn": row.UPN, "email": row.Email, "sam_account_name": row.SAMAccountName, "aad_object_id": row.AADObjectID,
			"sid": row.SID, "entity_origin": row.EntityOrigin, "seed_entity_mention_id": entityNullableUUIDString(row.SeedMentionID),
			"identity_state": row.State, "merged_into_record_id": entityNullableUUIDString(row.MergedIntoRecordID),
			"row_version": row.RowVersion, "created_at": entityFormatTimestamp(row.CreatedAt),
			"updated_at": entityFormatTimestamp(row.UpdatedAt), "created_by_user_id": createdBy,
			"updated_by_user_id": updatedBy, "privilege_level": row.PrivilegeLevel, "mfa_state": row.MFAState,
			"reset_status": row.ResetStatus,
		}); err != nil {
			return nil, err
		}
	}
	return payload.Bytes(), nil
}

func encodePortableMentions(ctx context.Context, exportContext sourceport.ExportContext, rows []portableMentionRow) ([]byte, error) {
	created := entityPortableActorMap(ctx, exportContext, "entity_mentions", "created_by_user_id", entityMentionIDs(rows, false))
	resolved := entityPortableActorMap(ctx, exportContext, "entity_mentions", "resolved_by_user_id", entityMentionIDs(rows, true))
	if created.err != nil || resolved.err != nil {
		return nil, errors.New("entity mention portability attribution resolution failed")
	}
	var payload bytes.Buffer
	for _, row := range rows {
		createdBy, err := entityPortableActorID(created.values[row.MentionID.String()], row.RuntimeCreatedByID)
		if err != nil {
			return nil, err
		}
		var resolvedBy any
		if row.RuntimeResolvedByID != nil {
			resolvedBy, err = entityPortableActorID(resolved.values[row.MentionID.String()], *row.RuntimeResolvedByID)
			if err != nil {
				return nil, err
			}
		}
		if err := appendEntityPortableRow(&payload, map[string]any{
			"entity_mention_id": row.MentionID.String(), "source_record_id": row.SourceRecordID.String(),
			"entity_type": row.EntityType, "source_field_key": row.SourceFieldKey, "origin_kind": row.OriginKind,
			"origin_locator": row.OriginLocator, "raw_text": row.RawText, "normalized_text": row.NormalizedText,
			"resolution_status": row.ResolutionStatus, "row_version": row.RowVersion, "ordinal": row.Ordinal,
			"created_by_user_id": createdBy, "created_at": entityFormatTimestamp(row.CreatedAt),
			"resolved_record_id": entityNullableUUIDString(row.ResolvedRecordID), "resolved_by_user_id": resolvedBy,
			"resolved_at": entityNullableTimestampString(row.ResolvedAt), "resolution_method": row.ResolutionMethod,
		}); err != nil {
			return nil, err
		}
	}
	return payload.Bytes(), nil
}

func encodePortablePreserved(ctx context.Context, exportContext sourceport.ExportContext, rows []portablePreservedIdentifierRow) ([]byte, error) {
	created := entityPortableActorMap(ctx, exportContext, "entity_preserved_identifiers", "created_by_user_id", entityPreservedIDs(rows))
	if created.err != nil {
		return nil, errors.New("entity preserved identifier portability attribution resolution failed")
	}
	var payload bytes.Buffer
	for _, row := range rows {
		createdBy, err := entityPortableActorID(created.values[row.PreservedID.String()], row.RuntimeCreatedByID)
		if err != nil {
			return nil, err
		}
		if err := appendEntityPortableRow(&payload, map[string]any{
			"entity_preserved_identifier_id": row.PreservedID.String(), "incident_id": row.IncidentID.String(),
			"record_id": row.RecordID.String(), "entity_type": row.EntityType, "identifier_type": row.IdentifierType,
			"raw_value": row.RawValue, "normalized_value": row.NormalizedValue, "classification": row.Classification,
			"created_by_user_id": createdBy, "created_at": entityFormatTimestamp(row.CreatedAt),
			"deleted_at": entityNullableTimestampString(row.DeletedAt),
		}); err != nil {
			return nil, err
		}
	}
	return payload.Bytes(), nil
}

func encodePortableAliases(ctx context.Context, exportContext sourceport.ExportContext, rows []portableAliasRow) ([]byte, error) {
	created := entityPortableActorMap(ctx, exportContext, "entity_aliases", "created_by_user_id", entityAliasIDs(rows))
	if created.err != nil {
		return nil, errors.New("entity alias portability attribution resolution failed")
	}
	var payload bytes.Buffer
	for _, row := range rows {
		createdBy, err := entityPortableActorID(created.values[row.AliasID.String()], row.RuntimeCreatedByID)
		if err != nil {
			return nil, err
		}
		if err := appendEntityPortableRow(&payload, map[string]any{
			"entity_alias_id": row.AliasID.String(), "incident_id": row.IncidentID.String(),
			"record_id": row.RecordID.String(), "entity_type": row.EntityType, "raw_text": row.RawText,
			"normalized_text": row.NormalizedText, "classification": row.Classification,
			"created_by_user_id": createdBy, "created_at": entityFormatTimestamp(row.CreatedAt),
			"deleted_at": entityNullableTimestampString(row.DeletedAt),
		}); err != nil {
			return nil, err
		}
	}
	return payload.Bytes(), nil
}

type entityActorMapResult struct {
	values map[string]string
	err    error
}

func entityPortableActorMap(ctx context.Context, exportContext sourceport.ExportContext, table, column string, rowIDs []string) entityActorMapResult {
	if exportContext.PortableAttributions == nil || len(rowIDs) == 0 {
		return entityActorMapResult{values: map[string]string{}}
	}
	values, err := exportContext.PortableAttributions.ResolvePortableSourceActors(
		ctx, exportContext.Query, exportContext.IncidentID, table, column, rowIDs,
	)
	return entityActorMapResult{values: values, err: err}
}

func entityPortableActorID(sourceActorID string, runtimeActorID uuid.UUID) (string, error) {
	if sourceActorID == "" {
		if runtimeActorID == uuid.Nil {
			return "", errors.New("entity portability export actor is invalid")
		}
		return runtimeActorID.String(), nil
	}
	parsed, err := uuid.Parse(sourceActorID)
	if err != nil || parsed.String() != sourceActorID {
		return "", errors.New("entity portability export attribution is invalid")
	}
	return sourceActorID, nil
}

func appendEntityPortableRow(payload *bytes.Buffer, row map[string]any) error {
	encoded, err := incidentportability.CanonicalJSONString(row)
	if err != nil {
		return errors.New("entity portability export encoding failed")
	}
	payload.Write(encoded)
	return nil
}

func entityTextFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func entityUUIDFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	result := uuid.UUID(value.Bytes)
	return &result
}

func entityTimeFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func entityNullableUUIDString(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func entityNullableTimestampString(value *time.Time) any {
	if value == nil {
		return nil
	}
	return entityFormatTimestamp(*value)
}

func entityHostIDs(rows []portableHostRow) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.RecordID.String())
	}
	return ids
}

func entityIdentityIDs(rows []portableIdentityRow) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.RecordID.String())
	}
	return ids
}

func entityMentionIDs(rows []portableMentionRow, resolvedOnly bool) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if !resolvedOnly || row.RuntimeResolvedByID != nil {
			ids = append(ids, row.MentionID.String())
		}
	}
	return ids
}

func entityPreservedIDs(rows []portablePreservedIdentifierRow) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.PreservedID.String())
	}
	return ids
}

func entityAliasIDs(rows []portableAliasRow) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.AliasID.String())
	}
	return ids
}
