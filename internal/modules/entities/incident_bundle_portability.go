package entities

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func exportEntityIncidentBundleFiles(ctx context.Context, exportContext sourceport.ExportContext) ([]incidentportability.File, error) {
	mentions, err := loadPortableEntityMentions(ctx, exportContext)
	if err != nil {
		return nil, err
	}
	hosts, err := loadPortableHosts(ctx, exportContext)
	if err != nil {
		return nil, err
	}
	identities, err := loadPortableIdentities(ctx, exportContext)
	if err != nil {
		return nil, err
	}
	preserved, err := loadPortablePreservedIdentifiers(ctx, exportContext)
	if err != nil {
		return nil, err
	}
	aliases, err := loadPortableAliases(ctx, exportContext)
	if err != nil {
		return nil, err
	}
	mentionPayload, err := encodePortableMentions(ctx, exportContext, mentions)
	if err != nil {
		return nil, err
	}
	hostPayload, err := encodePortableHosts(ctx, exportContext, hosts)
	if err != nil {
		return nil, err
	}
	identityPayload, err := encodePortableIdentities(ctx, exportContext, identities)
	if err != nil {
		return nil, err
	}
	preservedPayload, err := encodePortablePreserved(ctx, exportContext, preserved)
	if err != nil {
		return nil, err
	}
	aliasPayload, err := encodePortableAliases(ctx, exportContext, aliases)
	if err != nil {
		return nil, err
	}
	return []incidentportability.File{
		{Path: entityMentionsBundlePath, Payload: mentionPayload},
		{Path: hostsBundlePath, Payload: hostPayload},
		{Path: identitiesBundlePath, Payload: identityPayload},
		{Path: preservedIDsBundlePath, Payload: preservedPayload},
		{Path: entityAliasesBundlePath, Payload: aliasPayload},
	}, nil
}

func loadPortableHosts(ctx context.Context, exportContext sourceport.ExportContext) ([]portableHostRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT record_id, incident_id, display_name, hostname, aad_device_id, fqdn,
       entity_origin, seed_entity_mention_id, host_state, merged_into_record_id,
       row_version, created_at, updated_at, created_by_user_id, updated_by_user_id,
       location, os_platform, business_owner, criticality, containment_status
  FROM hosts
 WHERE incident_id = $1
 ORDER BY record_id
`, exportContext.IncidentID)
	if err != nil {
		return nil, errors.New("entity host portability export query failed")
	}
	defer rows.Close()
	var result []portableHostRow
	for rows.Next() {
		var row portableHostRow
		var hostname, aad, fqdn, location, platform, owner, criticality, containment pgtype.Text
		var seed, merged pgtype.UUID
		if err := rows.Scan(&row.RecordID, &row.IncidentID, &row.DisplayName, &hostname,
			&aad, &fqdn, &row.EntityOrigin, &seed, &row.State, &merged, &row.RowVersion,
			&row.CreatedAt, &row.UpdatedAt, &row.RuntimeCreatedByID, &row.RuntimeUpdatedByID,
			&location, &platform, &owner, &criticality, &containment); err != nil {
			return nil, errors.New("entity host portability export scan failed")
		}
		row.Hostname, row.AADDeviceID, row.FQDN = entityTextFromPG(hostname), entityTextFromPG(aad), entityTextFromPG(fqdn)
		row.SeedMentionID, row.MergedIntoRecordID = entityUUIDFromPG(seed), entityUUIDFromPG(merged)
		row.Location, row.OSPlatform, row.BusinessOwner = entityTextFromPG(location), entityTextFromPG(platform), entityTextFromPG(owner)
		row.Criticality, row.ContainmentStatus = entityTextFromPG(criticality), entityTextFromPG(containment)
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("entity host portability export iteration failed")
	}
	return result, nil
}

func loadPortableIdentities(ctx context.Context, exportContext sourceport.ExportContext) ([]portableIdentityRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT record_id, incident_id, display_name, upn, email::text, sam_account_name,
       aad_object_id, sid, entity_origin, seed_entity_mention_id, identity_state,
       merged_into_record_id, row_version, created_at, updated_at,
       created_by_user_id, updated_by_user_id, privilege_level, mfa_state, reset_status
  FROM identities
 WHERE incident_id = $1
 ORDER BY record_id
`, exportContext.IncidentID)
	if err != nil {
		return nil, errors.New("entity identity portability export query failed")
	}
	defer rows.Close()
	var result []portableIdentityRow
	for rows.Next() {
		var row portableIdentityRow
		var upn, email, sam, aad, sid, privilege, mfa, reset pgtype.Text
		var seed, merged pgtype.UUID
		if err := rows.Scan(&row.RecordID, &row.IncidentID, &row.DisplayName, &upn, &email,
			&sam, &aad, &sid, &row.EntityOrigin, &seed, &row.State, &merged,
			&row.RowVersion, &row.CreatedAt, &row.UpdatedAt, &row.RuntimeCreatedByID,
			&row.RuntimeUpdatedByID, &privilege, &mfa, &reset); err != nil {
			return nil, errors.New("entity identity portability export scan failed")
		}
		row.UPN, row.Email, row.SAMAccountName = entityTextFromPG(upn), entityTextFromPG(email), entityTextFromPG(sam)
		row.AADObjectID, row.SID = entityTextFromPG(aad), entityTextFromPG(sid)
		row.SeedMentionID, row.MergedIntoRecordID = entityUUIDFromPG(seed), entityUUIDFromPG(merged)
		row.PrivilegeLevel, row.MFAState, row.ResetStatus = entityTextFromPG(privilege), entityTextFromPG(mfa), entityTextFromPG(reset)
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("entity identity portability export iteration failed")
	}
	return result, nil
}

func loadPortableEntityMentions(ctx context.Context, exportContext sourceport.ExportContext) ([]portableMentionRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT mention.entity_mention_id, mention.source_record_id, mention.entity_type,
       mention.source_field_key, mention.origin_kind, mention.origin_locator,
       mention.raw_text, mention.normalized_text, mention.resolution_status,
       mention.row_version, mention.ordinal, mention.created_by_user_id,
       mention.created_at, mention.resolved_record_id, mention.resolved_by_user_id,
       mention.resolved_at, mention.resolution_method
  FROM entity_mentions AS mention
  JOIN records AS source_record
    ON source_record.record_id = mention.source_record_id
 WHERE source_record.incident_id = $1
 ORDER BY mention.entity_mention_id
`, exportContext.IncidentID)
	if err != nil {
		return nil, errors.New("entity mention portability export query failed")
	}
	defer rows.Close()
	var result []portableMentionRow
	for rows.Next() {
		var row portableMentionRow
		var resolvedRecord, resolvedBy pgtype.UUID
		var resolvedAt pgtype.Timestamptz
		var method pgtype.Text
		if err := rows.Scan(&row.MentionID, &row.SourceRecordID, &row.EntityType,
			&row.SourceFieldKey, &row.OriginKind, &row.OriginLocator, &row.RawText,
			&row.NormalizedText, &row.ResolutionStatus, &row.RowVersion, &row.Ordinal,
			&row.RuntimeCreatedByID, &row.CreatedAt, &resolvedRecord, &resolvedBy,
			&resolvedAt, &method); err != nil {
			return nil, errors.New("entity mention portability export scan failed")
		}
		row.ResolvedRecordID, row.RuntimeResolvedByID = entityUUIDFromPG(resolvedRecord), entityUUIDFromPG(resolvedBy)
		row.ResolvedAt, row.ResolutionMethod = entityTimeFromPG(resolvedAt), entityTextFromPG(method)
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("entity mention portability export iteration failed")
	}
	return result, nil
}

func loadPortablePreservedIdentifiers(ctx context.Context, exportContext sourceport.ExportContext) ([]portablePreservedIdentifierRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT entity_preserved_identifier_id, incident_id, record_id, entity_type,
       identifier_type, raw_value, normalized_value, classification,
       created_by_user_id, created_at, deleted_at
  FROM entity_preserved_identifiers
 WHERE incident_id = $1
 ORDER BY entity_preserved_identifier_id
`, exportContext.IncidentID)
	if err != nil {
		return nil, errors.New("entity preserved identifier portability export query failed")
	}
	defer rows.Close()
	var result []portablePreservedIdentifierRow
	for rows.Next() {
		var row portablePreservedIdentifierRow
		var deleted pgtype.Timestamptz
		if err := rows.Scan(&row.PreservedID, &row.IncidentID, &row.RecordID, &row.EntityType,
			&row.IdentifierType, &row.RawValue, &row.NormalizedValue, &row.Classification,
			&row.RuntimeCreatedByID, &row.CreatedAt, &deleted); err != nil {
			return nil, errors.New("entity preserved identifier portability export scan failed")
		}
		row.DeletedAt = entityTimeFromPG(deleted)
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("entity preserved identifier portability export iteration failed")
	}
	return result, nil
}

func loadPortableAliases(ctx context.Context, exportContext sourceport.ExportContext) ([]portableAliasRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT entity_alias_id, incident_id, record_id, entity_type, raw_text,
       normalized_text::text, classification, created_by_user_id, created_at, deleted_at
  FROM entity_aliases
 WHERE incident_id = $1
 ORDER BY entity_alias_id
`, exportContext.IncidentID)
	if err != nil {
		return nil, errors.New("entity alias portability export query failed")
	}
	defer rows.Close()
	var result []portableAliasRow
	for rows.Next() {
		var row portableAliasRow
		var deleted pgtype.Timestamptz
		if err := rows.Scan(&row.AliasID, &row.IncidentID, &row.RecordID, &row.EntityType,
			&row.RawText, &row.NormalizedText, &row.Classification,
			&row.RuntimeCreatedByID, &row.CreatedAt, &deleted); err != nil {
			return nil, errors.New("entity alias portability export scan failed")
		}
		row.DeletedAt = entityTimeFromPG(deleted)
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("entity alias portability export iteration failed")
	}
	return result, nil
}
