package hostidentity

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
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

type identifierSeed struct {
	IdentifierType string
	RawValue       string
	Classification string
}

type IdentifierSeed = identifierSeed

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

func (s *Store) upsertHostTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, now time.Time) (HostRecord, map[string]any, string, int, error) {
	input, err := hostInputFromCreateRequest(request)
	if err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	return s.upsertHostWithInputTx(ctx, tx, actor, incidentID, input, now)
}

func (s *Store) upsertIdentityTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, now time.Time) (IdentityRecord, map[string]any, string, int, error) {
	input, err := identityInputFromCreateRequest(request)
	if err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	return s.upsertIdentityWithInputTx(ctx, tx, actor, incidentID, input, now)
}

func (s *Store) captureHostSnapshotBeforeUpsertTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request CreateRequest) (*revisions.RecordSnapshot, error) {
	input, err := hostInputFromCreateRequest(request)
	if err != nil {
		return nil, err
	}
	current, matched, err := matchHostTx(ctx, tx, incidentID, input)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, nil
	}
	snapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (s *Store) captureIdentitySnapshotBeforeUpsertTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request CreateRequest) (*revisions.RecordSnapshot, error) {
	input, err := identityInputFromCreateRequest(request)
	if err != nil {
		return nil, err
	}
	current, matched, err := matchIdentityTx(ctx, tx, incidentID, input)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, nil
	}
	snapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (s *Store) upsertHostWithInputTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, input hostUpsertInput, now time.Time) (HostRecord, map[string]any, string, int, error) {
	current, matched, err := matchHostTx(ctx, tx, incidentID, input)
	if err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	if !matched {
		record := HostRecord{
			IncidentID:    incidentID,
			DisplayName:   input.DisplayName,
			AADDeviceID:   cloneStringPointer(input.AADDeviceID),
			FQDN:          cloneStringPointer(input.FQDN),
			Hostname:      cloneStringPointer(input.Hostname),
			Location:      cloneStringPointer(input.Location),
			OSPlatform:    cloneStringPointer(input.OSPlatform),
			BusinessOwner: cloneStringPointer(input.BusinessOwner),
			Criticality:   cloneStringPointer(input.Criticality),
			ContainmentStatus: cloneStringPointer(
				input.ContainmentStatus,
			),
			HostState:     "stub",
			EntityOrigin:  input.EntityOrigin,
			SeedMentionID: input.SeedMentionID,
			RowVersion:    1,
			CreatedAt:     now.UTC(),
			UpdatedAt:     now.UTC(),
			CreatedByUser: actor.ID,
			UpdatedByUser: actor.ID,
		}
		recordID, err := s.ports.records.InsertTx(ctx, tx, entityRecordInsertParams{
			IncidentID:      incidentID,
			RecordType:      "host",
			CreatedByUserID: actor.ID,
			CreatedAt:       record.CreatedAt,
			UpdatedByUserID: actor.ID,
			UpdatedAt:       record.UpdatedAt,
			RowVersion:      record.RowVersion,
		})
		if err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		record.RecordID = recordID
		if err := insertHostTx(ctx, tx, &record); err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		if _, err := syncPreservedIdentifiersTx(ctx, tx, incidentID, record.RecordID, "host", hostIdentifierSeeds(input), actor.ID, now); err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		aliasResult, err := syncEntityAliasesTx(ctx, tx, incidentID, record.RecordID, "host", input.AliasAdds, actor.ID, now)
		if err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		record.AliasMutations = aliasResult.Added
		if err := hydrateHostRecordTx(ctx, tx, &record); err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		return record, nil, "create", httpStatusCreated, nil
	}

	if err := hydrateHostRecordTx(ctx, tx, &current); err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	beforeRow := BuildHostRow(current)

	next := current
	fieldChanged := false
	if input.AllowDisplayNameUpdate && strings.TrimSpace(input.DisplayName) != "" && input.DisplayName != current.DisplayName {
		next.DisplayName = input.DisplayName
		fieldChanged = true
	}
	if current.AADDeviceID == nil && input.AADDeviceID != nil {
		next.AADDeviceID = cloneStringPointer(input.AADDeviceID)
		fieldChanged = true
	}
	if current.FQDN == nil && input.FQDN != nil {
		next.FQDN = cloneStringPointer(input.FQDN)
		fieldChanged = true
	}
	if current.Hostname == nil && input.Hostname != nil {
		next.Hostname = cloneStringPointer(input.Hostname)
		fieldChanged = true
	}
	if current.Location == nil && input.Location != nil {
		next.Location = cloneStringPointer(input.Location)
		fieldChanged = true
	}
	if current.OSPlatform == nil && input.OSPlatform != nil {
		next.OSPlatform = cloneStringPointer(input.OSPlatform)
		fieldChanged = true
	}
	if current.BusinessOwner == nil && input.BusinessOwner != nil {
		next.BusinessOwner = cloneStringPointer(input.BusinessOwner)
		fieldChanged = true
	}
	if current.Criticality == nil && input.Criticality != nil {
		next.Criticality = cloneStringPointer(input.Criticality)
		fieldChanged = true
	}
	if current.ContainmentStatus == nil && input.ContainmentStatus != nil {
		next.ContainmentStatus = cloneStringPointer(input.ContainmentStatus)
		fieldChanged = true
	}

	identifierChanged, err := hasPendingIdentifierSeedsTx(ctx, tx, current.RecordID, "host", hostIdentifierSeeds(input))
	if err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	aliasChanged, err := hasPendingAliasAddsTx(ctx, tx, current.RecordID, "host", input.AliasAdds)
	if err != nil {
		return HostRecord{}, nil, "", 0, err
	}

	if fieldChanged || identifierChanged || aliasChanged {
		next.RowVersion, err = s.ports.records.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
		if err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		next.UpdatedAt = now.UTC()
		next.UpdatedByUser = actor.ID
		if err := updateHostTx(ctx, tx, next); err != nil {
			return HostRecord{}, nil, "", 0, err
		}
	}
	if _, err := syncPreservedIdentifiersTx(ctx, tx, incidentID, current.RecordID, "host", hostIdentifierSeeds(input), actor.ID, now); err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	aliasResult, err := syncEntityAliasesTx(ctx, tx, incidentID, current.RecordID, "host", input.AliasAdds, actor.ID, now)
	if err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	next.AliasMutations = aliasResult.Added
	if err := hydrateHostRecordTx(ctx, tx, &next); err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	return next, beforeRow, "patch", httpStatusOK, nil
}

func (s *Store) upsertIdentityWithInputTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, input identityUpsertInput, now time.Time) (IdentityRecord, map[string]any, string, int, error) {
	current, matched, err := matchIdentityTx(ctx, tx, incidentID, input)
	if err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	if !matched {
		record := IdentityRecord{
			IncidentID:     incidentID,
			DisplayName:    input.DisplayName,
			AADObjectID:    cloneStringPointer(input.AADObjectID),
			SID:            cloneStringPointer(input.SID),
			UPN:            cloneStringPointer(input.UPN),
			Email:          cloneStringPointer(input.Email),
			SamAccountName: cloneStringPointer(input.SamAccountName),
			PrivilegeLevel: cloneStringPointer(input.PrivilegeLevel),
			MFAState:       cloneStringPointer(input.MFAState),
			ResetStatus:    cloneStringPointer(input.ResetStatus),
			IdentityState:  "stub",
			EntityOrigin:   input.EntityOrigin,
			SeedMentionID:  input.SeedMentionID,
			RowVersion:     1,
			CreatedAt:      now.UTC(),
			UpdatedAt:      now.UTC(),
			CreatedByUser:  actor.ID,
			UpdatedByUser:  actor.ID,
		}
		recordID, err := s.ports.records.InsertTx(ctx, tx, entityRecordInsertParams{
			IncidentID:      incidentID,
			RecordType:      "identity",
			CreatedByUserID: actor.ID,
			CreatedAt:       record.CreatedAt,
			UpdatedByUserID: actor.ID,
			UpdatedAt:       record.UpdatedAt,
			RowVersion:      record.RowVersion,
		})
		if err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		record.RecordID = recordID
		if err := insertIdentityTx(ctx, tx, &record); err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		if _, err := syncPreservedIdentifiersTx(ctx, tx, incidentID, record.RecordID, "identity", identityIdentifierSeeds(input), actor.ID, now); err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		aliasResult, err := syncEntityAliasesTx(ctx, tx, incidentID, record.RecordID, "identity", input.AliasAdds, actor.ID, now)
		if err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		record.AliasMutations = aliasResult.Added
		if err := hydrateIdentityRecordTx(ctx, tx, &record); err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		return record, nil, "create", httpStatusCreated, nil
	}

	if err := hydrateIdentityRecordTx(ctx, tx, &current); err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	beforeRow := BuildIdentityRow(current)

	next := current
	fieldChanged := false
	if input.AllowDisplayNameUpdate && strings.TrimSpace(input.DisplayName) != "" && input.DisplayName != current.DisplayName {
		next.DisplayName = input.DisplayName
		fieldChanged = true
	}
	if current.AADObjectID == nil && input.AADObjectID != nil {
		next.AADObjectID = cloneStringPointer(input.AADObjectID)
		fieldChanged = true
	}
	if current.SID == nil && input.SID != nil {
		next.SID = cloneStringPointer(input.SID)
		fieldChanged = true
	}
	if current.UPN == nil && input.UPN != nil {
		next.UPN = cloneStringPointer(input.UPN)
		fieldChanged = true
	}
	if current.Email == nil && input.Email != nil {
		next.Email = cloneStringPointer(input.Email)
		fieldChanged = true
	}
	if current.SamAccountName == nil && input.SamAccountName != nil {
		next.SamAccountName = cloneStringPointer(input.SamAccountName)
		fieldChanged = true
	}
	if current.PrivilegeLevel == nil && input.PrivilegeLevel != nil {
		next.PrivilegeLevel = cloneStringPointer(input.PrivilegeLevel)
		fieldChanged = true
	}
	if current.MFAState == nil && input.MFAState != nil {
		next.MFAState = cloneStringPointer(input.MFAState)
		fieldChanged = true
	}
	if current.ResetStatus == nil && input.ResetStatus != nil {
		next.ResetStatus = cloneStringPointer(input.ResetStatus)
		fieldChanged = true
	}

	identifierChanged, err := hasPendingIdentifierSeedsTx(ctx, tx, current.RecordID, "identity", identityIdentifierSeeds(input))
	if err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	aliasChanged, err := hasPendingAliasAddsTx(ctx, tx, current.RecordID, "identity", input.AliasAdds)
	if err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}

	if fieldChanged || identifierChanged || aliasChanged {
		next.RowVersion, err = s.ports.records.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
		if err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		next.UpdatedAt = now.UTC()
		next.UpdatedByUser = actor.ID
		if err := updateIdentityTx(ctx, tx, next); err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
	}
	if _, err := syncPreservedIdentifiersTx(ctx, tx, incidentID, current.RecordID, "identity", identityIdentifierSeeds(input), actor.ID, now); err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	aliasResult, err := syncEntityAliasesTx(ctx, tx, incidentID, current.RecordID, "identity", input.AliasAdds, actor.ID, now)
	if err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	next.AliasMutations = aliasResult.Added
	if err := hydrateIdentityRecordTx(ctx, tx, &next); err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	return next, beforeRow, "patch", httpStatusOK, nil
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

func syncPreservedIdentifiersTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, entityType string, seeds []identifierSeed, actorUserID uuid.UUID, now time.Time) (bool, error) {
	inserted := false
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
		if exists {
			continue
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO entity_preserved_identifiers (
    incident_id,
    record_id,
    entity_type,
    identifier_type,
    raw_value,
    normalized_value,
    classification,
    created_by_user_id,
    created_at,
    deleted_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULL)
`, incidentID, recordID, entityType, seed.IdentifierType, seed.RawValue, normalized, seed.Classification, actorUserID, now.UTC()); err != nil {
			return false, fmt.Errorf("insert preserved identifier: %w", err)
		}
		inserted = true
	}
	return inserted, nil
}

func syncEntityAliasesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, entityType string, actions []CollectionAction, actorUserID uuid.UUID, now time.Time) (AliasSyncResult, error) {
	result := AliasSyncResult{}
	for _, action := range actions {
		normalized, ok := fieldnorm.NormalizeAliasText(action.NormalizedText)
		if !ok {
			return AliasSyncResult{}, fmt.Errorf("invalid entity alias")
		}
		aliasID := uuid.New()
		var insertedID uuid.UUID
		err := tx.QueryRow(ctx, `
INSERT INTO entity_aliases (
	entity_alias_id,
    incident_id,
    record_id,
    entity_type,
    raw_text,
    normalized_text,
    classification,
    created_by_user_id,
    created_at,
    deleted_at
)
VALUES ($1, $2, $3, $4, $5::text, $5::citext, 'suggestion_only', $6, $7, NULL)
ON CONFLICT (record_id, entity_type, normalized_text) WHERE deleted_at IS NULL
DO NOTHING
RETURNING entity_alias_id
`, aliasID, incidentID, recordID, entityType, normalized, actorUserID, now.UTC()).Scan(&insertedID)
		if errors.Is(err, pgx.ErrNoRows) {
			result.DuplicateNoopCount++
			continue
		}
		if err != nil {
			return AliasSyncResult{}, fmt.Errorf("insert entity alias: %w", err)
		}
		result.Added = append(result.Added, AliasMutationValue{
			EntityAliasID: insertedID,
			IncidentID:    incidentID,
			RecordID:      recordID,
			EntityType:    entityType,
			AliasText:     normalized,
			CreatedByUser: actorUserID,
			CreatedAt:     now.UTC(),
		})
	}
	return result, nil
}

func applyEntityAliasActionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, entityType string, actions []CollectionAction, actorUserID uuid.UUID, now time.Time) ([]AliasAppliedMutation, error) {
	mutations := make([]AliasAppliedMutation, 0, len(actions))
	for _, action := range actions {
		switch action.Op {
		case "add_alias":
			result, err := syncEntityAliasesTx(ctx, tx, incidentID, recordID, entityType, []CollectionAction{action}, actorUserID, now)
			if err != nil {
				return nil, err
			}
			for _, added := range result.Added {
				mutations = append(mutations, AliasAppliedMutation{
					OperationKind: "create",
					TargetID:      "entity_alias:" + added.EntityAliasID.String(),
					AfterValue:    added.MutationValue(),
				})
			}
		case "remove_alias":
			aliasID, err := ParseEntityAliasItemRef(action.ItemRef)
			if err != nil {
				return nil, ErrInvalidAliasReference
			}
			var value AliasMutationValue
			var classification string
			err = tx.QueryRow(ctx, `
SELECT entity_alias_id, incident_id, record_id, entity_type, normalized_text::text,
       classification, created_by_user_id, created_at
  FROM entity_aliases
 WHERE entity_alias_id = $1
   AND deleted_at IS NULL
 FOR UPDATE
`, aliasID).Scan(
				&value.EntityAliasID,
				&value.IncidentID,
				&value.RecordID,
				&value.EntityType,
				&value.AliasText,
				&classification,
				&value.CreatedByUser,
				&value.CreatedAt,
			)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrInvalidAliasReference
			}
			if err != nil {
				return nil, err
			}
			if value.IncidentID != incidentID || value.RecordID != recordID || value.EntityType != entityType || classification != "suggestion_only" {
				return nil, ErrInvalidAliasReference
			}
			before := value.MutationValue()
			deletedAt := now.UTC()
			if tag, err := tx.Exec(ctx, `UPDATE entity_aliases SET deleted_at = $2 WHERE entity_alias_id = $1 AND deleted_at IS NULL`, aliasID, deletedAt); err != nil {
				return nil, err
			} else if tag.RowsAffected() != 1 {
				return nil, ErrInvalidAliasReference
			}
			value.DeletedAt = &deletedAt
			mutations = append(mutations, AliasAppliedMutation{
				OperationKind: "delete",
				TargetID:      action.ItemRef,
				BeforeValue:   before,
				AfterValue:    value.MutationValue(),
			})
		default:
			return nil, ErrInvalidAliasReference
		}
	}
	return mutations, nil
}

func loadEntityAliasesTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) ([]AliasValue, error) {
	rows, err := tx.Query(ctx, `
SELECT entity_alias_id, normalized_text::text
  FROM entity_aliases
 WHERE record_id = $1
   AND entity_type = $2
   AND deleted_at IS NULL
 ORDER BY normalized_text ASC, created_at ASC, entity_alias_id ASC
`, recordID, entityType)
	if err != nil {
		return nil, fmt.Errorf("load entity aliases: %w", err)
	}
	defer rows.Close()

	aliases := make([]AliasValue, 0)
	for rows.Next() {
		var alias AliasValue
		if err := rows.Scan(&alias.EntityAliasID, &alias.AliasText); err != nil {
			return nil, fmt.Errorf("scan entity alias: %w", err)
		}
		aliases = append(aliases, alias)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity aliases: %w", err)
	}
	return aliases, nil
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

func HostExactMatchPrecedence() []string {
	return append([]string(nil), hostExactMatchPrecedence...)
}

func IdentityExactMatchPrecedence() []string {
	return append([]string(nil), identityExactMatchPrecedence...)
}

func HostCanonicalNormalized(record HostRecord, identifierClass string) string {
	return hostCanonicalNormalized(record, identifierClass)
}

func IdentityCanonicalNormalized(record IdentityRecord, identifierClass string) string {
	return identityCanonicalNormalized(record, identifierClass)
}

func SyncPreservedIdentifiersTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, entityType string, seeds []IdentifierSeed, actorUserID uuid.UUID, now time.Time) (bool, error) {
	return syncPreservedIdentifiersTx(ctx, tx, incidentID, recordID, entityType, seeds, actorUserID, now)
}

func SyncEntityAliasesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, entityType string, actions []CollectionAction, actorUserID uuid.UUID, now time.Time) (AliasSyncResult, error) {
	return syncEntityAliasesTx(ctx, tx, incidentID, recordID, entityType, actions, actorUserID, now)
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
