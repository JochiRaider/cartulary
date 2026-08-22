package hostidentity

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

var (
	hostExactMatchPrecedence     = []string{"aad_device_id", "fqdn", "hostname"}
	identityExactMatchPrecedence = []string{"aad_object_id", "sid", "upn", "email", "sam_account_name"}
	ErrInvalidAliasReference     = errors.New("entities: invalid alias reference")
)

const (
	entityOriginEntitySheet = "entity_sheet"
)

type ExactMatchConflictError struct {
	EntityType       string
	IdentifierClass  string
	CandidateRecords []uuid.UUID
}

func (e *ExactMatchConflictError) Error() string {
	return fmt.Sprintf("%s exact match conflict on %s", e.EntityType, e.IdentifierClass)
}

// EntityMatchConflictDetails exposes only the closed, client-safe conflict
// facts needed by application adapters. The returned IDs are owned by the
// caller so an adapter cannot mutate the owner's error after construction.
func (e *ExactMatchConflictError) EntityMatchConflictDetails() (string, string, []uuid.UUID) {
	if e == nil {
		return "", "", nil
	}
	return e.EntityType, e.IdentifierClass, append([]uuid.UUID(nil), e.CandidateRecords...)
}

type identifierSeed struct {
	IdentifierType string
	RawValue       string
	Classification string
}

type hostUpsertInput struct {
	DisplayName            string
	AADDeviceID            *string
	FQDN                   *string
	Hostname               *string
	Location               *string
	OSPlatform             *string
	BusinessOwner          *string
	Criticality            *string
	ContainmentStatus      *string
	AliasAdds              []CollectionAction
	EntityOrigin           string
	SeedMentionID          *uuid.UUID
	AllowDisplayNameUpdate bool
}

type identityUpsertInput struct {
	DisplayName            string
	AADObjectID            *string
	SID                    *string
	UPN                    *string
	Email                  *string
	SamAccountName         *string
	PrivilegeLevel         *string
	MFAState               *string
	ResetStatus            *string
	AliasAdds              []CollectionAction
	EntityOrigin           string
	SeedMentionID          *uuid.UUID
	AllowDisplayNameUpdate bool
}

type preservedIdentifierRecord struct {
	RecordID        uuid.UUID
	EntityType      string
	IdentifierType  string
	NormalizedValue string
	Classification  string
}

func hostInputFromCreateRequest(request CreateRequest) (hostUpsertInput, error) {
	input := hostUpsertInput{
		DisplayName:            request.Values["host.display_name"],
		AADDeviceID:            optionalValue(request.Values, "host.aad_device_id"),
		FQDN:                   optionalValue(request.Values, "host.fqdn"),
		Hostname:               optionalValue(request.Values, "host.hostname"),
		Location:               optionalValue(request.Values, "host.location"),
		OSPlatform:             optionalValue(request.Values, "host.os_platform"),
		BusinessOwner:          optionalValue(request.Values, "host.business_owner"),
		Criticality:            optionalValue(request.Values, "host.criticality"),
		ContainmentStatus:      optionalValue(request.Values, "host.containment_status"),
		AliasAdds:              append([]CollectionAction(nil), request.AliasAdds["host.aliases"]...),
		EntityOrigin:           entityOriginEntitySheet,
		AllowDisplayNameUpdate: true,
	}
	if strings.TrimSpace(input.DisplayName) == "" {
		switch {
		case input.Hostname != nil:
			input.DisplayName = *input.Hostname
		case input.FQDN != nil:
			input.DisplayName = *input.FQDN
		case input.AADDeviceID != nil:
			input.DisplayName = *input.AADDeviceID
		default:
			return hostUpsertInput{}, ErrInvalidCreateRequest
		}
	}
	return input, nil
}

func identityInputFromCreateRequest(request CreateRequest) (identityUpsertInput, error) {
	input := identityUpsertInput{
		DisplayName:            request.Values["identity.display_name"],
		AADObjectID:            optionalValue(request.Values, "identity.aad_object_id"),
		SID:                    optionalValue(request.Values, "identity.sid"),
		UPN:                    optionalValue(request.Values, "identity.upn"),
		Email:                  optionalValue(request.Values, "identity.email"),
		SamAccountName:         optionalValue(request.Values, "identity.sam_account_name"),
		PrivilegeLevel:         optionalValue(request.Values, "identity.privilege_level"),
		MFAState:               optionalValue(request.Values, "identity.mfa_state"),
		ResetStatus:            optionalValue(request.Values, "identity.reset_status"),
		AliasAdds:              append([]CollectionAction(nil), request.AliasAdds["identity.aliases"]...),
		EntityOrigin:           entityOriginEntitySheet,
		AllowDisplayNameUpdate: true,
	}
	if strings.TrimSpace(input.DisplayName) == "" {
		switch {
		case input.UPN != nil:
			input.DisplayName = *input.UPN
		case input.Email != nil:
			input.DisplayName = *input.Email
		case input.SamAccountName != nil:
			input.DisplayName = *input.SamAccountName
		case input.SID != nil:
			input.DisplayName = *input.SID
		case input.AADObjectID != nil:
			input.DisplayName = *input.AADObjectID
		default:
			return identityUpsertInput{}, ErrInvalidCreateRequest
		}
	}
	return input, nil
}

func hostIdentifierSeeds(input hostUpsertInput) []identifierSeed {
	seeds := make([]identifierSeed, 0, 3)
	if input.AADDeviceID != nil {
		seeds = append(seeds, identifierSeed{IdentifierType: "aad_device_id", RawValue: *input.AADDeviceID, Classification: "exact_match_reuse"})
	}
	if input.FQDN != nil {
		seeds = append(seeds, identifierSeed{IdentifierType: "fqdn", RawValue: *input.FQDN, Classification: "exact_match_reuse"})
	}
	if input.Hostname != nil {
		seeds = append(seeds, identifierSeed{IdentifierType: "hostname", RawValue: *input.Hostname, Classification: "exact_match_reuse"})
	}
	return seeds
}

func identityIdentifierSeeds(input identityUpsertInput) []identifierSeed {
	seeds := make([]identifierSeed, 0, 5)
	if input.AADObjectID != nil {
		seeds = append(seeds, identifierSeed{IdentifierType: "aad_object_id", RawValue: *input.AADObjectID, Classification: "exact_match_reuse"})
	}
	if input.SID != nil {
		seeds = append(seeds, identifierSeed{IdentifierType: "sid", RawValue: *input.SID, Classification: "exact_match_reuse"})
	}
	if input.UPN != nil {
		seeds = append(seeds, identifierSeed{IdentifierType: "upn", RawValue: *input.UPN, Classification: "exact_match_reuse"})
	}
	if input.Email != nil {
		seeds = append(seeds, identifierSeed{IdentifierType: "email", RawValue: *input.Email, Classification: "exact_match_reuse"})
	}
	if input.SamAccountName != nil {
		seeds = append(seeds, identifierSeed{IdentifierType: "sam_account_name", RawValue: *input.SamAccountName, Classification: "exact_match_reuse"})
	}
	return seeds
}

func matchHostTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, input hostUpsertInput) (HostRecord, bool, error) {
	records, err := loadActiveHostsTx(ctx, tx, incidentID)
	if err != nil {
		return HostRecord{}, false, err
	}
	preserved, err := loadActivePreservedIdentifiersTx(ctx, tx, incidentID, "host")
	if err != nil {
		return HostRecord{}, false, err
	}
	recordIndex := make(map[uuid.UUID]HostRecord, len(records))
	for _, record := range records {
		recordIndex[record.RecordID] = record
	}

	candidates := map[string]*string{
		"aad_device_id": input.AADDeviceID,
		"fqdn":          input.FQDN,
		"hostname":      input.Hostname,
	}
	for _, identifierClass := range hostExactMatchPrecedence {
		raw := candidates[identifierClass]
		if raw == nil {
			continue
		}
		normalized, ok := fieldnorm.NormalizeIdentifier(identifierClass, *raw)
		if !ok {
			continue
		}
		matched := make(map[uuid.UUID]HostRecord)
		for _, record := range records {
			if hostCanonicalNormalized(record, identifierClass) == normalized {
				matched[record.RecordID] = record
			}
		}
		for _, identifier := range preserved {
			if identifier.IdentifierType != identifierClass || identifier.Classification != "exact_match_reuse" || identifier.NormalizedValue != normalized {
				continue
			}
			record, ok := recordIndex[identifier.RecordID]
			if ok {
				matched[record.RecordID] = record
			}
		}
		switch len(matched) {
		case 0:
			continue
		case 1:
			for _, record := range matched {
				return record, true, nil
			}
		default:
			return HostRecord{}, false, &ExactMatchConflictError{
				EntityType:       "host",
				IdentifierClass:  identifierClass,
				CandidateRecords: sortedRecordIDs(matched),
			}
		}
	}
	return HostRecord{}, false, nil
}

func matchIdentityTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, input identityUpsertInput) (IdentityRecord, bool, error) {
	records, err := loadActiveIdentitiesTx(ctx, tx, incidentID)
	if err != nil {
		return IdentityRecord{}, false, err
	}
	preserved, err := loadActivePreservedIdentifiersTx(ctx, tx, incidentID, "identity")
	if err != nil {
		return IdentityRecord{}, false, err
	}
	recordIndex := make(map[uuid.UUID]IdentityRecord, len(records))
	for _, record := range records {
		recordIndex[record.RecordID] = record
	}

	candidates := map[string]*string{
		"aad_object_id":    input.AADObjectID,
		"sid":              input.SID,
		"upn":              input.UPN,
		"email":            input.Email,
		"sam_account_name": input.SamAccountName,
	}
	for _, identifierClass := range identityExactMatchPrecedence {
		raw := candidates[identifierClass]
		if raw == nil {
			continue
		}
		normalized, ok := fieldnorm.NormalizeIdentifier(identifierClass, *raw)
		if !ok {
			continue
		}
		matched := make(map[uuid.UUID]IdentityRecord)
		for _, record := range records {
			if identityCanonicalNormalized(record, identifierClass) == normalized {
				matched[record.RecordID] = record
			}
		}
		for _, identifier := range preserved {
			if identifier.IdentifierType != identifierClass || identifier.Classification != "exact_match_reuse" || identifier.NormalizedValue != normalized {
				continue
			}
			record, ok := recordIndex[identifier.RecordID]
			if ok {
				matched[record.RecordID] = record
			}
		}
		switch len(matched) {
		case 0:
			continue
		case 1:
			for _, record := range matched {
				return record, true, nil
			}
		default:
			return IdentityRecord{}, false, &ExactMatchConflictError{
				EntityType:       "identity",
				IdentifierClass:  identifierClass,
				CandidateRecords: sortedRecordIDs(matched),
			}
		}
	}
	return IdentityRecord{}, false, nil
}

func loadActiveHostsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) ([]HostRecord, error) {
	rows, err := tx.Query(ctx, `
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
 WHERE h.incident_id = $1
   AND h.host_state IN ('stub', 'canonical')
 FOR UPDATE OF h, r
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("load active hosts: %w", err)
	}
	defer rows.Close()

	records := make([]HostRecord, 0)
	for rows.Next() {
		record, err := scanHostRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active hosts: %w", err)
	}
	return records, nil
}

func loadActiveIdentitiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) ([]IdentityRecord, error) {
	rows, err := tx.Query(ctx, `
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
 WHERE i.incident_id = $1
   AND i.identity_state IN ('stub', 'canonical')
 FOR UPDATE OF i, r
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("load active identities: %w", err)
	}
	defer rows.Close()

	records := make([]IdentityRecord, 0)
	for rows.Next() {
		record, err := scanIdentityRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active identities: %w", err)
	}
	return records, nil
}

func loadActivePreservedIdentifiersTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string) ([]preservedIdentifierRecord, error) {
	query := `
SELECT epi.record_id, epi.entity_type, epi.identifier_type, epi.normalized_value, epi.classification
  FROM entity_preserved_identifiers epi
`
	switch entityType {
	case "host":
		query += `
  JOIN hosts h
    ON h.record_id = epi.record_id
   AND h.incident_id = epi.incident_id
`
	case "identity":
		query += `
  JOIN identities i
    ON i.record_id = epi.record_id
   AND i.incident_id = epi.incident_id
`
	default:
		return nil, fmt.Errorf("load preserved identifiers: unsupported entity type %q", entityType)
	}
	query += `
 WHERE epi.incident_id = $1
   AND epi.entity_type = $2
   AND epi.deleted_at IS NULL
   AND epi.classification = 'exact_match_reuse'
`
	switch entityType {
	case "host":
		query += `
   AND h.host_state IN ('stub', 'canonical')
 FOR UPDATE OF epi, h
`
	case "identity":
		query += `
   AND i.identity_state IN ('stub', 'canonical')
 FOR UPDATE OF epi, i
`
	}
	rows, err := tx.Query(ctx, query, incidentID, entityType)
	if err != nil {
		return nil, fmt.Errorf("load preserved identifiers: %w", err)
	}
	defer rows.Close()

	records := make([]preservedIdentifierRecord, 0)
	for rows.Next() {
		var record preservedIdentifierRecord
		if err := rows.Scan(&record.RecordID, &record.EntityType, &record.IdentifierType, &record.NormalizedValue, &record.Classification); err != nil {
			return nil, fmt.Errorf("scan preserved identifier: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate preserved identifiers: %w", err)
	}
	return records, nil
}

func hasPendingIdentifierSeedsTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string, seeds []identifierSeed) (bool, error) {
	for _, seed := range seeds {
		normalized, ok := fieldnorm.NormalizeIdentifier(seed.IdentifierType, seed.RawValue)
		if !ok {
			continue
		}
		var exists bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM entity_preserved_identifiers
     WHERE record_id = $1
       AND entity_type = $2
       AND identifier_type = $3
       AND normalized_value = $4
       AND classification = $5
       AND deleted_at IS NULL
)`, recordID, entityType, seed.IdentifierType, normalized, seed.Classification).Scan(&exists); err != nil {
			return false, fmt.Errorf("query preserved identifier existence: %w", err)
		}
		if !exists {
			return true, nil
		}
	}
	return false, nil
}

func hasPendingAliasAddsTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string, actions []CollectionAction) (bool, error) {
	for _, action := range actions {
		var exists bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM entity_aliases
     WHERE record_id = $1
       AND entity_type = $2
       AND normalized_text = $3
       AND deleted_at IS NULL
)`, recordID, entityType, action.NormalizedText).Scan(&exists); err != nil {
			return false, fmt.Errorf("query alias existence: %w", err)
		}
		if !exists {
			return true, nil
		}
	}
	return false, nil
}

func hostCanonicalNormalized(record HostRecord, identifierClass string) string {
	switch identifierClass {
	case "aad_device_id":
		return normalizeOptionalIdentifier(identifierClass, record.AADDeviceID)
	case "fqdn":
		return normalizeOptionalIdentifier(identifierClass, record.FQDN)
	case "hostname":
		return normalizeOptionalIdentifier(identifierClass, record.Hostname)
	default:
		return ""
	}
}

func identityCanonicalNormalized(record IdentityRecord, identifierClass string) string {
	switch identifierClass {
	case "aad_object_id":
		return normalizeOptionalIdentifier(identifierClass, record.AADObjectID)
	case "sid":
		return normalizeOptionalIdentifier(identifierClass, record.SID)
	case "upn":
		return normalizeOptionalIdentifier(identifierClass, record.UPN)
	case "email":
		return normalizeOptionalIdentifier(identifierClass, record.Email)
	case "sam_account_name":
		return normalizeOptionalIdentifier(identifierClass, record.SamAccountName)
	default:
		return ""
	}
}

func normalizeOptionalIdentifier(identifierClass string, value *string) string {
	if value == nil {
		return ""
	}
	normalized, ok := fieldnorm.NormalizeIdentifier(identifierClass, *value)
	if !ok {
		return ""
	}
	return normalized
}

func sortedRecordIDs[T any](records map[uuid.UUID]T) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(records))
	for recordID := range records {
		ids = append(ids, recordID)
	}
	slices.SortFunc(ids, func(left uuid.UUID, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	return ids
}

func scanHostRecord(scanner interface {
	Scan(dest ...any) error
}) (HostRecord, error) {
	var (
		record           HostRecord
		rawAADDeviceID   pgtype.Text
		rawFQDN          pgtype.Text
		rawHostname      pgtype.Text
		rawLocation      pgtype.Text
		rawOSPlatform    pgtype.Text
		rawBusinessOwner pgtype.Text
		rawCriticality   pgtype.Text
		rawContainment   pgtype.Text
		rawMergedInto    pgtype.UUID
		rawSeedMention   pgtype.UUID
	)
	if err := scanner.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DisplayName,
		&rawAADDeviceID,
		&rawFQDN,
		&rawHostname,
		&rawLocation,
		&rawOSPlatform,
		&rawBusinessOwner,
		&rawCriticality,
		&rawContainment,
		&record.HostState,
		&rawMergedInto,
		&record.EntityOrigin,
		&rawSeedMention,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedByUser,
		&record.UpdatedByUser,
	); err != nil {
		return HostRecord{}, fmt.Errorf("scan host record: %w", err)
	}
	record.AADDeviceID = textPointer(rawAADDeviceID)
	record.FQDN = textPointer(rawFQDN)
	record.Hostname = textPointer(rawHostname)
	record.Location = textPointer(rawLocation)
	record.OSPlatform = textPointer(rawOSPlatform)
	record.BusinessOwner = textPointer(rawBusinessOwner)
	record.Criticality = textPointer(rawCriticality)
	record.ContainmentStatus = textPointer(rawContainment)
	record.MergedIntoRecordID = uuidPointerFromPG(rawMergedInto)
	record.SeedMentionID = uuidPointerFromPG(rawSeedMention)
	return record, nil
}

func scanIdentityRecord(scanner interface {
	Scan(dest ...any) error
}) (IdentityRecord, error) {
	var (
		record            IdentityRecord
		rawAADObjectID    pgtype.Text
		rawSID            pgtype.Text
		rawUPN            pgtype.Text
		rawEmail          pgtype.Text
		rawSamAccountName pgtype.Text
		rawPrivilegeLevel pgtype.Text
		rawMFAState       pgtype.Text
		rawResetStatus    pgtype.Text
		rawMergedInto     pgtype.UUID
		rawSeedMention    pgtype.UUID
	)
	if err := scanner.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DisplayName,
		&rawAADObjectID,
		&rawSID,
		&rawUPN,
		&rawEmail,
		&rawSamAccountName,
		&rawPrivilegeLevel,
		&rawMFAState,
		&rawResetStatus,
		&record.IdentityState,
		&rawMergedInto,
		&record.EntityOrigin,
		&rawSeedMention,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedByUser,
		&record.UpdatedByUser,
	); err != nil {
		return IdentityRecord{}, fmt.Errorf("scan identity record: %w", err)
	}
	record.AADObjectID = textPointer(rawAADObjectID)
	record.SID = textPointer(rawSID)
	record.UPN = textPointer(rawUPN)
	record.Email = textPointer(rawEmail)
	record.SamAccountName = textPointer(rawSamAccountName)
	record.PrivilegeLevel = textPointer(rawPrivilegeLevel)
	record.MFAState = textPointer(rawMFAState)
	record.ResetStatus = textPointer(rawResetStatus)
	record.MergedIntoRecordID = uuidPointerFromPG(rawMergedInto)
	record.SeedMentionID = uuidPointerFromPG(rawSeedMention)
	return record, nil
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func uuidPointerFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.Must(uuid.FromBytes(value.Bytes[:]))
	return &parsed
}

const (
	httpStatusCreated = 201
	httpStatusOK      = 200
)
