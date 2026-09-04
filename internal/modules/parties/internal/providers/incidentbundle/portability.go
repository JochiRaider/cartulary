package incidentbundle

import (
	"context"
	"errors"
	"sort"

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
	rows          []partyPortableRow
}

type partyActiveKeyClaim struct {
	incidentID      uuid.UUID
	keyKind         string
	normalizedValue string
	recordID        uuid.UUID
}

func exportIncidentBundleFiles(
	ctx context.Context,
	q incidentportability.Queryer,
	incidentID uuid.UUID,
) ([]incidentportability.File, error) {
	file, err := incidentportability.ExportNDJSON(ctx, q, incidentID, partyIncidentBundlePath, `
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
 ORDER BY record_id`)
	if err != nil {
		return nil, err
	}
	return []incidentportability.File{file}, nil
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
		importContext.IncidentID == uuid.Nil || importContext.BundleVersion != 3 ||
		descriptor.ContractMajor != sourceport.ContractMajor {
		return preparedPartyImport{}, sourceport.ErrPreparedBinding
	}
	payload, present := bundle.File(partyIncidentBundlePath)
	if !present {
		return preparedPartyImport{}, &incidentportability.VerificationFailure{ReasonCode: "missing_required_file"}
	}
	rawRows, err := incidentportability.DecodeStrictNDJSONObjects(payload, partyIncidentBundlePath)
	if err != nil {
		return preparedPartyImport{}, descriptor.DeclaredFailure("parties.version_shape_exact")
	}

	// Stable identity is admitted before every other member so malformed rows
	// cannot change invariant selection through member or row ordering.
	seen := make(map[uuid.UUID]struct{}, len(rawRows))
	type identifiedRow struct {
		recordID uuid.UUID
		raw      map[string]any
	}
	ordered := make([]identifiedRow, 0, len(rawRows))
	for _, raw := range rawRows {
		recordID, admitted := canonicalPartyPortableUUID(raw["record_id"])
		if !admitted {
			return preparedPartyImport{}, descriptor.DeclaredFailure("parties.source_identity_admitted")
		}
		incidentID, admitted := canonicalPartyPortableUUID(raw["incident_id"])
		if !admitted || incidentID != importContext.IncidentID {
			return preparedPartyImport{}, descriptor.DeclaredFailure("parties.source_identity_admitted")
		}
		if _, duplicate := seen[recordID]; duplicate {
			return preparedPartyImport{}, descriptor.DeclaredFailure("parties.source_identity_admitted")
		}
		seen[recordID] = struct{}{}
		ordered = append(ordered, identifiedRow{recordID: recordID, raw: raw})
	}
	sort.Slice(ordered, func(left, right int) bool {
		return ordered[left].recordID.String() < ordered[right].recordID.String()
	})

	rows := make([]partyPortableRow, 0, len(ordered))
	for _, candidate := range ordered {
		row, admitted := partyPortableRowShape(candidate.raw)
		if !admitted {
			return preparedPartyImport{}, descriptor.DeclaredFailure("parties.version_shape_exact")
		}
		rows = append(rows, row)
	}
	for _, row := range rows {
		if row.DisplayName == "" || !partyKindAdmitted(row.PartyKind) {
			return preparedPartyImport{}, descriptor.DeclaredFailure("parties.identity_lifecycle")
		}
	}
	for _, row := range rows {
		if !partyPortableNormalizationExact(row) {
			return preparedPartyImport{}, descriptor.DeclaredFailure("parties.normalization_exact")
		}
	}
	return preparedPartyImport{
		operationID:   importContext.OperationID,
		incidentID:    importContext.IncidentID,
		bundleVersion: importContext.BundleVersion,
		contractMajor: descriptor.ContractMajor,
		rows:          rows,
	}, nil
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
	for _, row := range prepared.rows {
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
	}
	return nil
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

	actualRows, err := loadPortablePartyRowsTx(ctx, tx, importContext.IncidentID)
	if err != nil {
		return err
	}
	if len(actualRows) != len(prepared.rows) {
		return descriptor.DeclaredFailure("parties.source_identity_admitted")
	}
	for index, actual := range actualRows {
		expected := prepared.rows[index]
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
	}
	claimsValid, err := partyActiveKeyClaimsValidTx(ctx, tx, importContext.IncidentID, actualRows)
	if err != nil {
		return err
	}
	if !claimsValid {
		return descriptor.DeclaredFailure("parties.identity_lifecycle")
	}
	return nil
}

func (prepared preparedPartyImport) matches(
	importContext sourceport.ImportContext,
	descriptor sourceport.Descriptor,
) bool {
	return prepared.operationID != "" &&
		prepared.operationID == importContext.OperationID &&
		prepared.incidentID != uuid.Nil &&
		prepared.incidentID == importContext.IncidentID &&
		prepared.bundleVersion == 3 &&
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

func loadPortablePartyRowsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]partyPortableRow, error) {
	rows, err := tx.Query(ctx, `
SELECT record_id, incident_id, display_name, party_kind,
       organization_name, role_title, primary_email, timezone_name,
       external_ref, notes
  FROM parties
 WHERE incident_id = $1
 ORDER BY record_id`, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]partyPortableRow, 0)
	for rows.Next() {
		var row partyPortableRow
		if err := rows.Scan(
			&row.RecordID, &row.IncidentID, &row.DisplayName, &row.PartyKind,
			&row.OrganizationName, &row.RoleTitle, &row.PrimaryEmail,
			&row.TimezoneName, &row.ExternalRef, &row.Notes,
		); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func partyActiveKeyClaimsValidTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	partyRows []partyPortableRow,
) (bool, error) {
	activeRecordRows, err := tx.Query(ctx, `
SELECT record_id
  FROM records
 WHERE incident_id = $1
   AND record_type = 'party'
   AND deleted_at IS NULL`, incidentID)
	if err != nil {
		return false, err
	}
	activeRecordIDs := make(map[uuid.UUID]struct{})
	for activeRecordRows.Next() {
		var recordID uuid.UUID
		if err := activeRecordRows.Scan(&recordID); err != nil {
			activeRecordRows.Close()
			return false, err
		}
		activeRecordIDs[recordID] = struct{}{}
	}
	if err := activeRecordRows.Err(); err != nil {
		activeRecordRows.Close()
		return false, err
	}
	activeRecordRows.Close()

	expected := make(map[partyActiveKeyClaim]struct{})
	for _, row := range partyRows {
		if _, active := activeRecordIDs[row.RecordID]; !active {
			continue
		}
		for _, candidate := range []struct {
			fieldKey string
			keyKind  string
			value    *string
		}{
			{fieldKey: "party.primary_email", keyKind: "primary_email", value: row.PrimaryEmail},
			{fieldKey: "party.external_ref", keyKind: "external_ref", value: row.ExternalRef},
		} {
			admitted, admissionErr := policy.AdmitStored(candidate.fieldKey, candidate.value)
			if admissionErr != nil {
				return false, nil
			}
			normalizedValue, present := admitted.ExactMatchClaimValue()
			if !present {
				continue
			}
			expected[partyActiveKeyClaim{
				incidentID:      incidentID,
				keyKind:         candidate.keyKind,
				normalizedValue: normalizedValue,
				recordID:        row.RecordID,
			}] = struct{}{}
		}
	}

	claimRows, err := tx.Query(ctx, `
SELECT incident_id, key_kind, normalized_value, party_record_id
  FROM party_active_key_claims
 WHERE incident_id = $1`, incidentID)
	if err != nil {
		return false, err
	}
	actual := make(map[partyActiveKeyClaim]struct{})
	for claimRows.Next() {
		var claim partyActiveKeyClaim
		if err := claimRows.Scan(
			&claim.incidentID,
			&claim.keyKind,
			&claim.normalizedValue,
			&claim.recordID,
		); err != nil {
			claimRows.Close()
			return false, err
		}
		actual[claim] = struct{}{}
	}
	if err := claimRows.Err(); err != nil {
		claimRows.Close()
		return false, err
	}
	claimRows.Close()
	if len(expected) != len(actual) {
		return false, nil
	}
	for claim := range expected {
		if _, present := actual[claim]; !present {
			return false, nil
		}
	}
	return true, nil
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
