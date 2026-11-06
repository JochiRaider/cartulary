package entities

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func exportEntityIncidentBundleFiles(ctx context.Context, input sourceport.ExportContext) error {
	if err := exportSourceMember(ctx, input, hostsBundlePath, loadPortableHostsPage, func(row portableHostRow) uuid.UUID { return row.RecordID }, encodePortableHosts); err != nil {
		return err
	}
	if err := exportSourceMember(ctx, input, identitiesBundlePath, loadPortableIdentitiesPage, func(row portableIdentityRow) uuid.UUID { return row.RecordID }, encodePortableIdentities); err != nil {
		return err
	}
	if err := exportSourceMember(ctx, input, entityMentionsBundlePath, loadPortableEntityMentionsPage, func(row portableMentionRow) uuid.UUID { return row.MentionID }, encodePortableMentions); err != nil {
		return err
	}
	if err := exportSourceMember(ctx, input, preservedIDsBundlePath, loadPortablePreservedIdentifiersPage, func(row portablePreservedIdentifierRow) uuid.UUID { return row.PreservedID }, encodePortablePreserved); err != nil {
		return err
	}
	if err := exportSourceMember(ctx, input, entityAliasesBundlePath, loadPortableAliasesPage, func(row portableAliasRow) uuid.UUID { return row.AliasID }, encodePortableAliases); err != nil {
		return err
	}
	return nil
}

func loadPortableHostsPage(ctx context.Context, exportContext sourceport.ExportContext, after *uuid.UUID) ([]portableHostRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT record_id, incident_id, display_name, hostname, aad_device_id, fqdn,
       entity_origin, seed_entity_mention_id, host_state, merged_into_record_id,
       row_version, created_at, updated_at, created_by_user_id, updated_by_user_id,
       location, os_platform, business_owner, criticality, containment_status
  FROM hosts
 WHERE incident_id = $1
 AND ($2::uuid IS NULL OR record_id > $2)
 ORDER BY record_id
 LIMIT 128
`, exportContext.IncidentID, after)
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

func loadPortableIdentitiesPage(ctx context.Context, exportContext sourceport.ExportContext, after *uuid.UUID) ([]portableIdentityRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT record_id, incident_id, display_name, upn, email::text, sam_account_name,
       aad_object_id, sid, entity_origin, seed_entity_mention_id, identity_state,
       merged_into_record_id, row_version, created_at, updated_at,
       created_by_user_id, updated_by_user_id, privilege_level, mfa_state, reset_status
  FROM identities
 WHERE incident_id = $1
 AND ($2::uuid IS NULL OR record_id > $2)
 ORDER BY record_id
 LIMIT 128
`, exportContext.IncidentID, after)
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

func loadPortableEntityMentionsPage(ctx context.Context, exportContext sourceport.ExportContext, after *uuid.UUID) ([]portableMentionRow, error) {
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
 AND ($2::uuid IS NULL OR mention.entity_mention_id > $2)
 ORDER BY mention.entity_mention_id
 LIMIT 128
`, exportContext.IncidentID, after)
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

func loadPortablePreservedIdentifiersPage(ctx context.Context, exportContext sourceport.ExportContext, after *uuid.UUID) ([]portablePreservedIdentifierRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT entity_preserved_identifier_id, incident_id, record_id, entity_type,
       identifier_type, raw_value, normalized_value, classification,
       created_by_user_id, created_at, deleted_at
  FROM entity_preserved_identifiers
 WHERE incident_id = $1
 AND ($2::uuid IS NULL OR entity_preserved_identifier_id > $2)
 ORDER BY entity_preserved_identifier_id
 LIMIT 128
`, exportContext.IncidentID, after)
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

func loadPortableAliasesPage(ctx context.Context, exportContext sourceport.ExportContext, after *uuid.UUID) ([]portableAliasRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT entity_alias_id, incident_id, record_id, entity_type, raw_text,
       normalized_text::text, classification, created_by_user_id, created_at, deleted_at
  FROM entity_aliases
 WHERE incident_id = $1
 AND ($2::uuid IS NULL OR entity_alias_id > $2)
 ORDER BY entity_alias_id
 LIMIT 128
`, exportContext.IncidentID, after)
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

func exportSourceMember[T any](ctx context.Context, input sourceport.ExportContext, path string,
	load func(context.Context, sourceport.ExportContext, *uuid.UUID) ([]T, error),
	identity func(T) uuid.UUID,
	encode func(context.Context, sourceport.ExportContext, []T, io.Writer) error,
) error {
	return incidentportability.WriteMemberFrom(ctx, input.Members, path, func(ctx context.Context, output io.Writer) error {
		var after *uuid.UUID
		for {
			rows, err := load(ctx, input, after)
			if err != nil {
				return err
			}
			if len(rows) == 0 {
				return nil
			}
			if err := encode(ctx, input, rows, output); err != nil {
				return err
			}
			last := identity(rows[len(rows)-1])
			after = &last
		}
	})
}
