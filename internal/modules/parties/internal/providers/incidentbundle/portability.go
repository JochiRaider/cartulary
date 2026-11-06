package incidentbundle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"hash"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/parties/internal/policy"
)

const partyIncidentBundlePath = "data/parties.ndjson"

var partyIncidentBundleMembers = []string{
	"record_id",
	"incident_id",
	"display_name",
	"party_kind",
	"organization_name",
	"role_title",
	"primary_email",
	"timezone_name",
	"external_ref",
	"notes",
}

type partyPortableRow struct {
	RecordID         uuid.UUID
	IncidentID       uuid.UUID
	DisplayName      string
	PartyKind        string
	OrganizationName *string
	RoleTitle        *string
	PrimaryEmail     *string
	TimezoneName     *string
	ExternalRef      *string
	Notes            *string
}

type preparedPartyImport struct {
	operationID   string
	incidentID    uuid.UUID
	bundleVersion int
	contractMajor int
	members       *incidentportability.BoundMembers
}

func exportIncidentBundleFiles(
	ctx context.Context,
	q incidentportability.Queryer,
	incidentID uuid.UUID,
	members incidentportability.MemberWriter,
) error {
	err := incidentportability.ExportNDJSONTo(ctx, q, incidentID, partyIncidentBundlePath, `
SELECT jsonb_build_object(
           'record_id', record_id,
           'incident_id', incident_id,
           'display_name', display_name,
           'party_kind', party_kind,
           'organization_name', organization_name,
           'role_title', role_title,
           'primary_email', primary_email,
           'timezone_name', timezone_name,
           'external_ref', external_ref,
           'notes', notes
       )
  FROM parties
 WHERE incident_id = $1
 ORDER BY record_id`, members)
	if err != nil {
		return err
	}
	return nil
}

func preparePartyImport(
	ctx context.Context,
	bundle sourceport.Bundle,
	importContext sourceport.ImportContext,
	descriptor sourceport.Descriptor,
) (preparedPartyImport, error) {
	if err := ctx.Err(); err != nil {
		return preparedPartyImport{}, err
	}
	if bundle == nil || importContext.OperationID == "" ||
		importContext.IncidentID == uuid.Nil || importContext.BundleVersion != 7 ||
		descriptor.ContractMajor != sourceport.ContractMajor {
		return preparedPartyImport{}, sourceport.ErrPreparedBinding
	}

	members, err := incidentportability.BindMembers(bundle, []string{partyIncidentBundlePath})
	if err != nil {
		return preparedPartyImport{}, err
	}
	prepared := preparedPartyImport{operationID: importContext.OperationID, incidentID: importContext.IncidentID, bundleVersion: importContext.BundleVersion, contractMajor: descriptor.ContractMajor, members: members}
	previous := ""
	err = incidentportability.WalkPath(ctx, members, partyIncidentBundlePath, func(raw map[string]any) error {
		recordID, ok := canonicalPartyPortableUUID(raw["record_id"])
		incidentID, inScope := canonicalPartyPortableUUID(raw["incident_id"])
		if !ok || !inScope || incidentID != importContext.IncidentID || recordID.String() <= previous {
			return descriptor.DeclaredFailure("parties.source_identity_admitted")
		}
		previous = recordID.String()
		return nil
	})
	if err != nil {
		return preparedPartyImport{}, partyStreamFailure(descriptor, err)
	}
	// Each bounded pass preserves the owner's semantic precedence without sorting
	// or retaining the complete source collection.
	if err := prepared.walk(ctx, descriptor, func(partyPortableRow) error { return nil }); err != nil {
		return preparedPartyImport{}, err
	}
	if err := prepared.walk(ctx, descriptor, func(row partyPortableRow) error {
		if row.DisplayName == "" || !partyKindAdmitted(row.PartyKind) {
			return descriptor.DeclaredFailure("parties.identity_lifecycle")
		}
		return nil
	}); err != nil {
		return preparedPartyImport{}, err
	}
	if err := prepared.walk(ctx, descriptor, func(row partyPortableRow) error {
		if !partyPortableNormalizationExact(row) {
			return descriptor.DeclaredFailure("parties.normalization_exact")
		}
		return nil
	}); err != nil {
		return preparedPartyImport{}, err
	}
	return prepared, nil
}
func (p preparedPartyImport) walk(ctx context.Context, descriptor sourceport.Descriptor, visit func(partyPortableRow) error) error {
	err := incidentportability.WalkPath(ctx, p.members, partyIncidentBundlePath, func(raw map[string]any) error {
		row, valid := partyPortableRowShape(raw)
		if !valid {
			return descriptor.DeclaredFailure("parties.version_shape_exact")
		}
		return visit(row)
	})
	return partyStreamFailure(descriptor, err)
}
func partyStreamFailure(descriptor sourceport.Descriptor, err error) error {
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return descriptor.DeclaredFailure("parties.version_shape_exact")
	}
	return err
}

func applyPreparedPartyImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedPartyImport,
	importContext sourceport.ImportContext,
	descriptor sourceport.Descriptor,
) error {
	if !prepared.matches(importContext, descriptor) || tx == nil {
		return sourceport.ErrPreparedBinding
	}
	return prepared.walk(ctx, descriptor, func(row partyPortableRow) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `
INSERT INTO parties (
    record_id,
    incident_id,
    display_name,
    party_kind,
    organization_name,
    role_title,
    primary_email,
    timezone_name,
    external_ref,
    notes
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`, row.RecordID, row.IncidentID, row.DisplayName, row.PartyKind,
			row.OrganizationName, row.RoleTitle, row.PrimaryEmail,
			row.TimezoneName, row.ExternalRef, row.Notes)
		if err != nil {
			return classifyPartyPortableApplyError(descriptor, err)
		}
		if tag.RowsAffected() != 1 {
			return descriptor.DeclaredFailure("parties.source_identity_admitted")
		}
		return nil
	})
}

func validatePreparedPartyImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedPartyImport,
	importContext sourceport.ImportContext,
	descriptor sourceport.Descriptor,
) error {
	if !prepared.matches(importContext, descriptor) || tx == nil {
		return sourceport.ErrPreparedBinding
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	var envelopeInvalid bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM parties party
      LEFT JOIN records record
        ON record.record_id = party.record_id
       AND record.incident_id = party.incident_id
     WHERE party.incident_id = $1
       AND (record.record_id IS NULL OR record.record_type <> 'party')
    UNION ALL
    SELECT 1
      FROM records record
      LEFT JOIN parties party
        ON party.record_id = record.record_id
       AND party.incident_id = record.incident_id
     WHERE record.incident_id = $1
       AND record.record_type = 'party'
       AND party.record_id IS NULL
)`, importContext.IncidentID).Scan(&envelopeInvalid); err != nil {
		return err
	}
	if envelopeInvalid {
		return descriptor.DeclaredFailure("parties.envelope_type_scope")
	}

	var lifecycleInvalid bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM parties
     WHERE incident_id = $1
       AND (
           display_name IS NULL OR display_name = '' OR
           party_kind IS NULL OR party_kind NOT IN (
               'person', 'team', 'organization', 'distribution_list', 'other'
           )
       )
)`, importContext.IncidentID).Scan(&lifecycleInvalid); err != nil {
		return err
	}
	if lifecycleInvalid {
		return descriptor.DeclaredFailure("parties.identity_lifecycle")
	}

	actualRows, err := tx.Query(ctx, `
 SELECT party.record_id,party.incident_id,party.display_name,party.party_kind,
 party.organization_name,party.role_title,party.primary_email,party.timezone_name,
 party.external_ref,party.notes,record.deleted_at IS NULL
 FROM parties party JOIN records record ON record.record_id=party.record_id
 WHERE party.incident_id=$1 ORDER BY party.record_id
 `, importContext.IncidentID)
	if err != nil {
		return err
	}
	defer actualRows.Close()
	expectedClaims := newPartyClaimInventory()
	err = prepared.walk(ctx, descriptor, func(expected partyPortableRow) error {
		if !actualRows.Next() {
			if err := actualRows.Err(); err != nil {
				return err
			}
			return descriptor.DeclaredFailure("parties.source_identity_admitted")
		}
		var actual partyPortableRow
		var active bool
		if err := actualRows.Scan(&actual.RecordID, &actual.IncidentID, &actual.DisplayName, &actual.PartyKind, &actual.OrganizationName, &actual.RoleTitle, &actual.PrimaryEmail, &actual.TimezoneName, &actual.ExternalRef, &actual.Notes, &active); err != nil {
			return err
		}
		if actual.RecordID != expected.RecordID || actual.IncidentID != expected.IncidentID {
			return descriptor.DeclaredFailure("parties.source_identity_admitted")
		}
		if actual.PartyKind != expected.PartyKind {
			return descriptor.DeclaredFailure("parties.identity_lifecycle")
		}
		if !partyPortableNormalizationExact(actual) ||
			actual.DisplayName != expected.DisplayName ||
			!optionalPartyStringEqual(actual.OrganizationName, expected.OrganizationName) ||
			!optionalPartyStringEqual(actual.RoleTitle, expected.RoleTitle) ||
			!optionalPartyStringEqual(actual.PrimaryEmail, expected.PrimaryEmail) ||
			!optionalPartyStringEqual(actual.TimezoneName, expected.TimezoneName) ||
			!optionalPartyStringEqual(actual.ExternalRef, expected.ExternalRef) ||
			!optionalPartyStringEqual(actual.Notes, expected.Notes) {
			return descriptor.DeclaredFailure("parties.normalization_exact")
		}
		if active {
			for _, candidate := range []struct {
				field, key string
				value      *string
			}{
				{"party.external_ref", "external_ref", actual.ExternalRef},
				{"party.primary_email", "primary_email", actual.PrimaryEmail},
			} {
				admitted, err := policy.AdmitStored(candidate.field, candidate.value)
				if err != nil {
					return descriptor.DeclaredFailure("parties.normalization_exact")
				}
				if normalized, present := admitted.ExactMatchClaimValue(); present {
					expectedClaims.add(importContext.IncidentID, actual.RecordID, candidate.key, normalized)
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if actualRows.Next() {
		return descriptor.DeclaredFailure("parties.source_identity_admitted")
	}
	if err := actualRows.Err(); err != nil {
		return err
	}
	actualRows.Close()
	claimRows, err := tx.Query(ctx, `SELECT party_record_id,key_kind,normalized_value FROM party_active_key_claims WHERE incident_id=$1 ORDER BY party_record_id,key_kind,normalized_value`, importContext.IncidentID)
	if err != nil {
		return err
	}
	defer claimRows.Close()
	actualClaims := newPartyClaimInventory()
	for claimRows.Next() {
		var recordID uuid.UUID
		var key, value string
		if err := claimRows.Scan(&recordID, &key, &value); err != nil {
			return err
		}
		actualClaims.add(importContext.IncidentID, recordID, key, value)
	}
	if err := claimRows.Err(); err != nil {
		return err
	}
	if expectedClaims.count != actualClaims.count || !bytes.Equal(expectedClaims.digest.Sum(nil), actualClaims.digest.Sum(nil)) {
		return descriptor.DeclaredFailure("parties.identity_lifecycle")
	}
	return nil
}

type partyClaimInventory struct {
	count  uint64
	digest hash.Hash
}

func newPartyClaimInventory() *partyClaimInventory { return &partyClaimInventory{digest: sha256.New()} }
func (i *partyClaimInventory) add(incidentID, recordID uuid.UUID, key, value string) {
	encoded, _ := json.Marshal([]string{incidentID.String(), recordID.String(), key, value})
	_, _ = i.digest.Write(encoded)
	_, _ = i.digest.Write([]byte{'\n'})
	i.count++
}

func (prepared preparedPartyImport) matches(
	importContext sourceport.ImportContext,
	descriptor sourceport.Descriptor,
) bool {
	return prepared.members != nil && prepared.operationID != "" &&
		prepared.operationID == importContext.OperationID &&
		prepared.incidentID != uuid.Nil &&
		prepared.incidentID == importContext.IncidentID &&
		prepared.bundleVersion == 7 &&
		prepared.bundleVersion == importContext.BundleVersion &&
		prepared.contractMajor == sourceport.ContractMajor &&
		prepared.contractMajor == descriptor.ContractMajor
}

func canonicalPartyPortableUUID(value any) (uuid.UUID, bool) {
	raw, ok := value.(string)
	if !ok {
		return uuid.UUID{}, false
	}
	parsed, err := uuid.Parse(raw)
	return parsed, err == nil && parsed.String() == raw
}

func partyPortableRowShape(raw map[string]any) (partyPortableRow, bool) {
	if len(raw) != len(partyIncidentBundleMembers) {
		return partyPortableRow{}, false
	}
	for _, member := range partyIncidentBundleMembers {
		if _, present := raw[member]; !present {
			return partyPortableRow{}, false
		}
	}
	recordID, recordAdmitted := canonicalPartyPortableUUID(raw["record_id"])
	incidentID, incidentAdmitted := canonicalPartyPortableUUID(raw["incident_id"])
	displayName, displayAdmitted := raw["display_name"].(string)
	partyKind, kindAdmitted := raw["party_kind"].(string)
	organizationName, organizationAdmitted := nullablePartyPortableString(raw["organization_name"])
	roleTitle, roleAdmitted := nullablePartyPortableString(raw["role_title"])
	primaryEmail, emailAdmitted := nullablePartyPortableString(raw["primary_email"])
	timezoneName, timezoneAdmitted := nullablePartyPortableString(raw["timezone_name"])
	externalRef, externalAdmitted := nullablePartyPortableString(raw["external_ref"])
	notes, notesAdmitted := nullablePartyPortableString(raw["notes"])
	if !recordAdmitted || !incidentAdmitted || !displayAdmitted || !kindAdmitted ||
		!organizationAdmitted || !roleAdmitted || !emailAdmitted ||
		!timezoneAdmitted || !externalAdmitted || !notesAdmitted {
		return partyPortableRow{}, false
	}
	return partyPortableRow{
		RecordID: recordID, IncidentID: incidentID, DisplayName: displayName,
		PartyKind: partyKind, OrganizationName: organizationName,
		RoleTitle: roleTitle, PrimaryEmail: primaryEmail, TimezoneName: timezoneName,
		ExternalRef: externalRef, Notes: notes,
	}, true
}

func nullablePartyPortableString(value any) (*string, bool) {
	if value == nil {
		return nil, true
	}
	text, ok := value.(string)
	if !ok {
		return nil, false
	}
	return &text, true
}

func partyKindAdmitted(value string) bool {
	admitted, admissionErr := policy.AdmitStored("party.party_kind", &value)
	if admissionErr != nil {
		return false
	}
	stored, present := admitted.StoredValue()
	return present && stored == value
}

func partyPortableNormalizationExact(row partyPortableRow) bool {
	fields := []struct {
		key   string
		value *string
	}{
		{key: "party.display_name", value: &row.DisplayName},
		{key: "party.organization_name", value: row.OrganizationName},
		{key: "party.role_title", value: row.RoleTitle},
		{key: "party.primary_email", value: row.PrimaryEmail},
		{key: "party.timezone_name", value: row.TimezoneName},
		{key: "party.external_ref", value: row.ExternalRef},
		{key: "party.notes", value: row.Notes},
	}
	for _, field := range fields {
		admitted, admissionErr := policy.AdmitStored(field.key, field.value)
		if admissionErr != nil {
			return false
		}
		stored, present := admitted.StoredValue()
		if field.value == nil {
			if present {
				return false
			}
			continue
		}
		if !present || stored != *field.value {
			return false
		}
	}
	return true
}

func optionalPartyStringEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func classifyPartyPortableApplyError(descriptor sourceport.Descriptor, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return err
	}
	switch postgresError.ConstraintName {
	case "parties_pkey":
		return descriptor.DeclaredFailure("parties.source_identity_admitted")
	case "party_active_key_claims_pkey", "party_active_key_claims_record_key_unique":
		return descriptor.DeclaredFailure("parties.identity_lifecycle")
	case "parties_record_envelope_fkey", "parties_incident_id_fkey":
		return descriptor.DeclaredFailure("parties.envelope_type_scope")
	default:
		if postgresError.Code == "23514" || postgresError.Code == "23502" {
			return descriptor.DeclaredFailure("parties.identity_lifecycle")
		}
		return err
	}
}
