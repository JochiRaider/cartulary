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

type normalizedIdentifierTuple struct {
	IdentifierType  string
	NormalizedValue string
}

type ActiveIdentifierTransitionConflict struct {
	IdentifierClass  string
	NormalizedValue  string
	BlockingRecordID uuid.UUID
}

func PrepareActiveIdentifierStateTransitionTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, recordID uuid.UUID, releasing bool) (*ActiveIdentifierTransitionConflict, error) {
	tuples, err := recordIdentifierTuplesTx(ctx, tx, incidentID, entityType, recordID)
	if err != nil {
		return nil, err
	}
	if err := lockIdentifierTuplesTx(ctx, tx, incidentID, entityType, tuples); err != nil {
		return nil, err
	}
	if releasing {
		if _, err := tx.Exec(ctx, `SELECT public.entities_release_active_identifier_claims_v1($1)`, recordID); err != nil {
			return nil, fmt.Errorf("release active entity identifier claims: %w", err)
		}
		return nil, nil
	}
	for _, tuple := range tuples {
		var blockingRecordID uuid.UUID
		err := tx.QueryRow(ctx, `
SELECT record_id
  FROM entity_active_identifier_claims
 WHERE incident_id = $1
   AND entity_type = $2
   AND identifier_type = $3
   AND normalized_value = $4
`, incidentID, entityType, tuple.IdentifierType, tuple.NormalizedValue).Scan(&blockingRecordID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("validate restored active entity identifier claim: %w", err)
		}
		if blockingRecordID != recordID {
			return &ActiveIdentifierTransitionConflict{
				IdentifierClass:  tuple.IdentifierType,
				NormalizedValue:  tuple.NormalizedValue,
				BlockingRecordID: blockingRecordID,
			}, nil
		}
	}
	return nil, nil
}

func recordIdentifierTuplesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, recordID uuid.UUID) ([]normalizedIdentifierTuple, error) {
	seeds := make([]identifierSeed, 0, 8)
	switch entityType {
	case "host":
		var aadDeviceID, fqdn, hostname pgtype.Text
		if err := tx.QueryRow(ctx, `
SELECT aad_device_id, fqdn, hostname
  FROM hosts
 WHERE incident_id = $1
   AND record_id = $2
`, incidentID, recordID).Scan(&aadDeviceID, &fqdn, &hostname); err != nil {
			return nil, fmt.Errorf("load Host transition identifiers: %w", err)
		}
		for identifierType, value := range map[string]pgtype.Text{
			"aad_device_id": aadDeviceID,
			"fqdn":          fqdn,
			"hostname":      hostname,
		} {
			if value.Valid {
				seeds = append(seeds, identifierSeed{IdentifierType: identifierType, RawValue: value.String})
			}
		}
	case "identity":
		var aadObjectID, sid, upn, email, samAccountName pgtype.Text
		if err := tx.QueryRow(ctx, `
SELECT aad_object_id, sid, upn, email::text, sam_account_name
  FROM identities
 WHERE incident_id = $1
   AND record_id = $2
`, incidentID, recordID).Scan(&aadObjectID, &sid, &upn, &email, &samAccountName); err != nil {
			return nil, fmt.Errorf("load Identity transition identifiers: %w", err)
		}
		for identifierType, value := range map[string]pgtype.Text{
			"aad_object_id":    aadObjectID,
			"sid":              sid,
			"upn":              upn,
			"email":            email,
			"sam_account_name": samAccountName,
		} {
			if value.Valid {
				seeds = append(seeds, identifierSeed{IdentifierType: identifierType, RawValue: value.String})
			}
		}
	default:
		return nil, fmt.Errorf("unsupported active entity identifier type %q", entityType)
	}
	tuples := normalizedIdentifierTuples(seeds)
	rows, err := tx.Query(ctx, `
SELECT identifier_type, normalized_value
  FROM entity_preserved_identifiers
 WHERE incident_id = $1
   AND entity_type = $2
   AND record_id = $3
   AND classification = 'exact_match_reuse'
   AND deleted_at IS NULL
`, incidentID, entityType, recordID)
	if err != nil {
		return nil, fmt.Errorf("load transition preserved identifiers: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var tuple normalizedIdentifierTuple
		if err := rows.Scan(&tuple.IdentifierType, &tuple.NormalizedValue); err != nil {
			return nil, fmt.Errorf("scan transition preserved identifier: %w", err)
		}
		tuples = append(tuples, tuple)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transition preserved identifiers: %w", err)
	}
	unique := make(map[string]normalizedIdentifierTuple, len(tuples))
	for _, tuple := range tuples {
		unique[tuple.IdentifierType+"\x1f"+tuple.NormalizedValue] = tuple
	}
	tuples = tuples[:0]
	for _, tuple := range unique {
		tuples = append(tuples, tuple)
	}
	slices.SortFunc(tuples, func(left normalizedIdentifierTuple, right normalizedIdentifierTuple) int {
		if compared := strings.Compare(left.IdentifierType, right.IdentifierType); compared != 0 {
			return compared
		}
		return strings.Compare(left.NormalizedValue, right.NormalizedValue)
	})
	return tuples, nil
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
	recordID, matched, err := matchClaimedRecordTx(
		ctx,
		tx,
		incidentID,
		"host",
		hostExactMatchPrecedence,
		normalizedIdentifierTuples(hostIdentifierSeeds(input)),
	)
	if err != nil {
		return HostRecord{}, false, err
	}
	if !matched {
		return HostRecord{}, false, nil
	}
	record, err := loadHostByRecordIDTx(ctx, tx, recordID)
	if err != nil {
		return HostRecord{}, false, err
	}
	if record.IncidentID != incidentID || record.HostState == "merged" {
		return HostRecord{}, false, fmt.Errorf("active Host claim points to an inactive source record")
	}
	return record, true, nil
}

func matchIdentityTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, input identityUpsertInput) (IdentityRecord, bool, error) {
	recordID, matched, err := matchClaimedRecordTx(
		ctx,
		tx,
		incidentID,
		"identity",
		identityExactMatchPrecedence,
		normalizedIdentifierTuples(identityIdentifierSeeds(input)),
	)
	if err != nil {
		return IdentityRecord{}, false, err
	}
	if !matched {
		return IdentityRecord{}, false, nil
	}
	record, err := loadIdentityByRecordIDTx(ctx, tx, recordID)
	if err != nil {
		return IdentityRecord{}, false, err
	}
	if record.IncidentID != incidentID || record.IdentityState == "merged" {
		return IdentityRecord{}, false, fmt.Errorf("active Identity claim points to an inactive source record")
	}
	return record, true, nil
}

func normalizedIdentifierTuples(seeds []identifierSeed) []normalizedIdentifierTuple {
	unique := make(map[string]normalizedIdentifierTuple, len(seeds))
	for _, seed := range seeds {
		normalized, ok := fieldnorm.NormalizeIdentifier(seed.IdentifierType, seed.RawValue)
		if !ok {
			continue
		}
		tuple := normalizedIdentifierTuple{
			IdentifierType:  seed.IdentifierType,
			NormalizedValue: normalized,
		}
		unique[tuple.IdentifierType+"\x1f"+tuple.NormalizedValue] = tuple
	}
	tuples := make([]normalizedIdentifierTuple, 0, len(unique))
	for _, tuple := range unique {
		tuples = append(tuples, tuple)
	}
	slices.SortFunc(tuples, func(left normalizedIdentifierTuple, right normalizedIdentifierTuple) int {
		if compared := strings.Compare(left.IdentifierType, right.IdentifierType); compared != 0 {
			return compared
		}
		return strings.Compare(left.NormalizedValue, right.NormalizedValue)
	})
	return tuples
}

func lockIdentifierTuplesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, tuples []normalizedIdentifierTuple) error {
	for _, tuple := range tuples {
		lockKey := incidentID.String() + "\x1f" + entityType + "\x1f" + tuple.IdentifierType + "\x1f" + tuple.NormalizedValue
		if _, err := tx.Exec(ctx, `
SELECT pg_catalog.pg_advisory_xact_lock(
    pg_catalog.hashtextextended($1, 0)
)
`, lockKey); err != nil {
			return fmt.Errorf("lock active entity identifier claim: %w", err)
		}
	}
	return nil
}

func matchClaimedRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, precedence []string, tuples []normalizedIdentifierTuple) (uuid.UUID, bool, error) {
	if err := lockIdentifierTuplesTx(ctx, tx, incidentID, entityType, tuples); err != nil {
		return uuid.Nil, false, err
	}
	byType := make(map[string]normalizedIdentifierTuple, len(tuples))
	for _, tuple := range tuples {
		byType[tuple.IdentifierType] = tuple
	}
	matched := make(map[uuid.UUID]struct{})
	firstMatchedClass := ""
	for _, identifierType := range precedence {
		tuple, ok := byType[identifierType]
		if !ok {
			continue
		}
		var recordID uuid.UUID
		err := tx.QueryRow(ctx, `
SELECT record_id
  FROM entity_active_identifier_claims
 WHERE incident_id = $1
   AND entity_type = $2
   AND identifier_type = $3
   AND normalized_value = $4
`, incidentID, entityType, tuple.IdentifierType, tuple.NormalizedValue).Scan(&recordID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return uuid.Nil, false, fmt.Errorf("lookup active entity identifier claim: %w", err)
		}
		if firstMatchedClass == "" {
			firstMatchedClass = identifierType
		}
		matched[recordID] = struct{}{}
	}
	if len(matched) == 0 {
		return uuid.Nil, false, nil
	}
	if len(matched) > 1 {
		return uuid.Nil, false, &ExactMatchConflictError{
			EntityType:       entityType,
			IdentifierClass:  firstMatchedClass,
			CandidateRecords: sortedRecordIDs(matched),
		}
	}
	for recordID := range matched {
		return recordID, true, nil
	}
	return uuid.Nil, false, nil
}

func prepareIdentifierMutationTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, recordID uuid.UUID, tuples []normalizedIdentifierTuple) error {
	if err := lockIdentifierTuplesTx(ctx, tx, incidentID, entityType, tuples); err != nil {
		return err
	}
	for _, tuple := range tuples {
		var claimedRecordID uuid.UUID
		err := tx.QueryRow(ctx, `
SELECT record_id
  FROM entity_active_identifier_claims
 WHERE incident_id = $1
   AND entity_type = $2
   AND identifier_type = $3
   AND normalized_value = $4
`, incidentID, entityType, tuple.IdentifierType, tuple.NormalizedValue).Scan(&claimedRecordID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("validate active entity identifier claim: %w", err)
		}
		if claimedRecordID != recordID {
			return &ExactMatchConflictError{
				EntityType:       entityType,
				IdentifierClass:  tuple.IdentifierType,
				CandidateRecords: []uuid.UUID{claimedRecordID},
			}
		}
	}
	return nil
}

func hostPatchIdentifierTuples(changes []PatchChange) []normalizedIdentifierTuple {
	seeds := make([]identifierSeed, 0, 1)
	for _, change := range changes {
		if change.FieldKey == "host.hostname" && change.Value != nil {
			seeds = append(seeds, identifierSeed{IdentifierType: "hostname", RawValue: *change.Value})
		}
	}
	return normalizedIdentifierTuples(seeds)
}

func identityPatchIdentifierTuples(changes []PatchChange) []normalizedIdentifierTuple {
	seeds := make([]identifierSeed, 0, 3)
	for _, change := range changes {
		identifierType := ""
		switch change.FieldKey {
		case "identity.upn":
			identifierType = "upn"
		case "identity.email":
			identifierType = "email"
		case "identity.sam_account_name":
			identifierType = "sam_account_name"
		}
		if identifierType != "" && change.Value != nil {
			seeds = append(seeds, identifierSeed{IdentifierType: identifierType, RawValue: *change.Value})
		}
	}
	return normalizedIdentifierTuples(seeds)
}

func mergeNormalizedIdentifierTuples(groups ...[]normalizedIdentifierTuple) []normalizedIdentifierTuple {
	unique := make(map[string]normalizedIdentifierTuple)
	for _, group := range groups {
		for _, tuple := range group {
			unique[tuple.IdentifierType+"\x1f"+tuple.NormalizedValue] = tuple
		}
	}
	merged := make([]normalizedIdentifierTuple, 0, len(unique))
	for _, tuple := range unique {
		merged = append(merged, tuple)
	}
	slices.SortFunc(merged, func(left normalizedIdentifierTuple, right normalizedIdentifierTuple) int {
		if compared := strings.Compare(left.IdentifierType, right.IdentifierType); compared != 0 {
			return compared
		}
		return strings.Compare(left.NormalizedValue, right.NormalizedValue)
	})
	return merged
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
