package incidents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/text/unicode/norm"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/internal/sourcestate"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

const (
	maxIncidentSourceBytes      = 128 * 1024
	incidentSourceContractMajor = 2
)

const (
	incidentSourceIdentityInvariant = iota
	incidentExactShapeInvariant
	incidentIdentityLifecycleInvariant
	incidentAttributionVersionInvariant
)

var (
	canonicalSourceInteger           = regexp.MustCompile(`^[1-9][0-9]*$`)
	canonicalSourceTime              = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,6})?\+00:00$`)
	errIncidentSourceCatalog         = errors.New("incident source catalog is invalid")
	errIncidentSourcePreparedBinding = errors.New("prepared incident source belongs to another operation")
)

type incidentSourceContract struct {
	source  sourcestate.SourceDescriptor
	columns []string
}

func newIncidentSourceContract() (incidentSourceContract, error) {
	source, err := sourcestate.Source()
	if err != nil {
		return incidentSourceContract{}, err
	}
	if source.ContractMajor != incidentSourceContractMajor {
		return incidentSourceContract{}, fmt.Errorf("%w: contract generation mismatch", errIncidentSourceCatalog)
	}
	columns, err := sourcestate.IncidentColumns()
	if err != nil {
		return incidentSourceContract{}, err
	}
	return incidentSourceContract{source: source, columns: columns}, nil
}

func (contract incidentSourceContract) failure(index int) error {
	if index < 0 || index >= len(contract.source.InvariantIDs) {
		return fmt.Errorf("%w: Incidents invariant precedence index %d", errIncidentSourceCatalog, index)
	}
	return &incidentSourceInvariantFailure{invariantID: contract.source.InvariantIDs[index]}
}

type incidentSourceInvariantFailure struct {
	invariantID string
}

func (failure *incidentSourceInvariantFailure) Error() string {
	if failure == nil {
		return "incident source invariant failed"
	}
	return "incident source invariant failed: " + failure.invariantID
}

type incidentSourceImportContext struct {
	incidentID    uuid.UUID
	actorUserID   uuid.UUID
	bundleVersion int
	operationID   string
	attributions  incidentportability.AttributionRecorder
	actorAdmitted func(string) bool
}

type incidentSourceRow struct {
	ID                     uuid.UUID
	IncidentKey            string
	IncidentKeyCanonical   string
	Title                  string
	Description            *string
	Status                 string
	Severity               *string
	TLP                    *string
	CurrentPhase           *string
	PrimaryExternalCaseRef *string
	CreatedByUserID        uuid.UUID
	CreatedAt              time.Time
	UpdatedAt              time.Time
	UpdatedByUserID        uuid.UUID
	IncidentVersion        int64
	ClosedAt               *time.Time
}

type preparedIncidentSource struct {
	portKey       string
	logicalPath   string
	schemaID      string
	operationID   string
	incidentID    uuid.UUID
	bundleVersion int
	contractMajor int
	row           incidentSourceRow
}

func (prepared preparedIncidentSource) matches(
	contract incidentSourceContract,
	importContext incidentSourceImportContext,
) bool {
	return prepared.portKey == contract.source.OwnerID+":"+contract.source.FamilyID &&
		prepared.logicalPath == contract.source.Path.LogicalPath &&
		prepared.schemaID == contract.source.Path.SchemaID &&
		prepared.operationID != "" && prepared.operationID == importContext.operationID &&
		prepared.incidentID != uuid.Nil && prepared.incidentID == importContext.incidentID &&
		prepared.bundleVersion == 3 && prepared.bundleVersion == importContext.bundleVersion &&
		prepared.contractMajor == incidentSourceContractMajor &&
		prepared.contractMajor == contract.source.ContractMajor
}

type incidentSourceDatabase interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

const incidentSourceExportSQL = `
SELECT jsonb_build_object(
    'id', i.id,
    'incident_key', i.incident_key,
    'incident_key_canonical', i.incident_key_canonical,
    'title', i.title,
    'description', i.description,
    'status', i.status,
    'severity', i.severity,
    'tlp', i.tlp,
    'current_phase', i.current_phase,
    'primary_external_case_ref', i.primary_external_case_ref,
    'created_by_user_id', i.created_by_user_id,
    'created_at', i.created_at,
    'updated_at', i.updated_at,
    'updated_by_user_id', i.updated_by_user_id,
    'incident_version', i.incident_version,
    'closed_at', i.closed_at
), i.incident_key
FROM incidents i
WHERE i.id = $1`

func exportIncidentBundleIncident(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID) ([]byte, string, error) {
	var raw []byte
	var incidentKey string
	err := q.QueryRow(ctx, incidentSourceExportSQL, incidentID).Scan(&raw, &incidentKey)
	if err != nil {
		return nil, "", err
	}
	canonical, err := incidentportability.CanonicalRawJSON(raw)
	return canonical, incidentKey, err
}

func prepareIncidentBundleIncident(
	contract incidentSourceContract,
	payload []byte,
	importContext incidentSourceImportContext,
) (preparedIncidentSource, error) {
	if len(payload) == 0 || importContext.operationID == "" || importContext.incidentID == uuid.Nil ||
		importContext.actorUserID == uuid.Nil || importContext.bundleVersion != 3 || importContext.actorAdmitted == nil ||
		contract.source.ContractMajor != incidentSourceContractMajor || len(contract.columns) != 16 {
		return preparedIncidentSource{}, fmt.Errorf("%w: Incidents source binding is invalid", errIncidentSourceCatalog)
	}
	row, err := decodeIncidentSourceRow(contract, payload, importContext)
	if err != nil {
		return preparedIncidentSource{}, err
	}
	return preparedIncidentSource{
		portKey:       contract.source.OwnerID + ":" + contract.source.FamilyID,
		logicalPath:   contract.source.Path.LogicalPath,
		schemaID:      contract.source.Path.SchemaID,
		operationID:   importContext.operationID,
		incidentID:    importContext.incidentID,
		bundleVersion: importContext.bundleVersion,
		contractMajor: contract.source.ContractMajor,
		row:           row,
	}, nil
}

func decodeIncidentSourceRow(
	contract incidentSourceContract,
	payload []byte,
	importContext incidentSourceImportContext,
) (incidentSourceRow, error) {
	if len(payload) == 0 || len(payload) > maxIncidentSourceBytes || !utf8.Valid(payload) {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	raw, err := strictjson.DecodeObject(bytes.NewReader(payload))
	if err != nil {
		if member, duplicate := strictjson.DuplicateMember(err); duplicate && member == "id" {
			return incidentSourceRow{}, contract.failure(incidentSourceIdentityInvariant)
		}
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}

	idText, ok := requiredSourceString(raw["id"])
	if !ok {
		return incidentSourceRow{}, contract.failure(incidentSourceIdentityInvariant)
	}
	id, ok := canonicalSourceUUID(idText)
	if !ok || id != importContext.incidentID {
		return incidentSourceRow{}, contract.failure(incidentSourceIdentityInvariant)
	}
	if !exactIncidentSourceMembers(raw, contract.columns) {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}

	row := incidentSourceRow{ID: id}
	var version json.Number
	if row.IncidentKey, ok = requiredSourceString(raw["incident_key"]); !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	if row.IncidentKeyCanonical, ok = requiredSourceString(raw["incident_key_canonical"]); !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	if row.Title, ok = requiredSourceString(raw["title"]); !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	if row.Description, ok = nullableSourceString(raw["description"]); !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	if row.Status, ok = requiredSourceString(raw["status"]); !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	if row.Severity, ok = nullableSourceString(raw["severity"]); !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	if row.TLP, ok = nullableSourceString(raw["tlp"]); !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	if row.CurrentPhase, ok = nullableSourceString(raw["current_phase"]); !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	if row.PrimaryExternalCaseRef, ok = nullableSourceString(raw["primary_external_case_ref"]); !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	createdByText, ok := requiredSourceString(raw["created_by_user_id"])
	if !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	createdAtText, ok := requiredSourceString(raw["created_at"])
	if !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	updatedAtText, ok := requiredSourceString(raw["updated_at"])
	if !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	updatedByText, ok := requiredSourceString(raw["updated_by_user_id"])
	if !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	if version, ok = requiredSourceNumber(raw["incident_version"]); !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}
	closedAtText, closedAtPresent, ok := nullableSourceStringValue(raw["closed_at"])
	if !ok {
		return incidentSourceRow{}, contract.failure(incidentExactShapeInvariant)
	}

	if !validIncidentKey(row.IncidentKey) || row.IncidentKeyCanonical != row.IncidentKey ||
		!validSourceText(row.Title, 512, false) || !validOptionalDescription(row.Description) ||
		!validOptionalMetadata(row.Severity) || !validOptionalMetadata(row.CurrentPhase) ||
		!validOptionalMetadata(row.PrimaryExternalCaseRef) || !validOptionalTLP(row.TLP) ||
		(row.Status != "active" && row.Status != "closed") ||
		(row.Status == "active" && closedAtPresent) || (row.Status == "closed" && !closedAtPresent) {
		return incidentSourceRow{}, contract.failure(incidentIdentityLifecycleInvariant)
	}

	if row.CreatedByUserID, ok = canonicalSourceUUID(createdByText); !ok {
		return incidentSourceRow{}, contract.failure(incidentAttributionVersionInvariant)
	}
	if !importContext.actorAdmitted(createdByText) {
		return incidentSourceRow{}, contract.failure(incidentAttributionVersionInvariant)
	}
	if row.UpdatedByUserID, ok = canonicalSourceUUID(updatedByText); !ok {
		return incidentSourceRow{}, contract.failure(incidentAttributionVersionInvariant)
	}
	if !importContext.actorAdmitted(updatedByText) {
		return incidentSourceRow{}, contract.failure(incidentAttributionVersionInvariant)
	}
	if row.CreatedAt, ok = parseCanonicalSourceTime(createdAtText); !ok {
		return incidentSourceRow{}, contract.failure(incidentAttributionVersionInvariant)
	}
	if row.UpdatedAt, ok = parseCanonicalSourceTime(updatedAtText); !ok || row.CreatedAt.After(row.UpdatedAt) {
		return incidentSourceRow{}, contract.failure(incidentAttributionVersionInvariant)
	}
	if row.IncidentVersion, ok = parseCanonicalSourceVersion(version.String()); !ok {
		return incidentSourceRow{}, contract.failure(incidentAttributionVersionInvariant)
	}
	if closedAtPresent {
		closedAt, valid := parseCanonicalSourceTime(closedAtText)
		if !valid || closedAt.Before(row.CreatedAt) || closedAt.After(row.UpdatedAt) {
			return incidentSourceRow{}, contract.failure(incidentAttributionVersionInvariant)
		}
		row.ClosedAt = &closedAt
	}
	return row, nil
}

func exactIncidentSourceMembers(raw map[string]json.RawMessage, columns []string) bool {
	if len(raw) != len(columns) {
		return false
	}
	for _, column := range columns {
		if _, present := raw[column]; !present {
			return false
		}
	}
	return true
}

func decodeSourceScalar(raw json.RawMessage) (any, bool) {
	if len(raw) == 0 {
		return nil, false
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, false
	}
	return value, true
}

func requiredSourceString(raw json.RawMessage) (string, bool) {
	value, ok := decodeSourceScalar(raw)
	if !ok {
		return "", false
	}
	result, ok := value.(string)
	return result, ok
}

func nullableSourceString(raw json.RawMessage) (*string, bool) {
	value, present, ok := nullableSourceStringValue(raw)
	if !ok {
		return nil, false
	}
	if !present {
		return nil, true
	}
	return &value, true
}

func nullableSourceStringValue(raw json.RawMessage) (string, bool, bool) {
	value, ok := decodeSourceScalar(raw)
	if !ok {
		return "", false, false
	}
	if value == nil {
		return "", false, true
	}
	result, ok := value.(string)
	return result, true, ok
}

func requiredSourceNumber(raw json.RawMessage) (json.Number, bool) {
	value, ok := decodeSourceScalar(raw)
	if !ok {
		return "", false
	}
	result, ok := value.(json.Number)
	return result, ok
}

func canonicalSourceUUID(value string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return uuid.Nil, false
	}
	return parsed, true
}

func validIncidentKey(value string) bool {
	return len(value) <= 128 && validSourceText(value, 0, false)
}

func validSourceText(value string, maxRunes int, allowDescriptionControls bool) bool {
	if value == "" || norm.NFC.String(value) != value || (maxRunes > 0 && utf8.RuneCountInString(value) > maxRunes) {
		return false
	}
	for _, char := range value {
		if allowDescriptionControls && (char == '\n' || char == '\t') {
			continue
		}
		if unicode.IsControl(char) || char == '\u2028' || char == '\u2029' {
			return false
		}
	}
	return true
}

func validOptionalDescription(value *string) bool {
	return value == nil || validSourceText(*value, 16384, true)
}

func validOptionalMetadata(value *string) bool {
	return value == nil || validSourceText(*value, 128, false)
}

func validOptionalTLP(value *string) bool {
	if value == nil {
		return true
	}
	switch *value {
	case "TLP:CLEAR", "TLP:GREEN", "TLP:AMBER", "TLP:AMBER+STRICT", "TLP:RED":
		return true
	default:
		return false
	}
}

func parseCanonicalSourceTime(value string) (time.Time, bool) {
	if !canonicalSourceTime.MatchString(value) {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func parseCanonicalSourceVersion(value string) (int64, bool) {
	if !canonicalSourceInteger.MatchString(value) {
		return 0, false
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 1 {
		return 0, false
	}
	return parsed, true
}

const insertIncidentSourceSQL = `
INSERT INTO incidents (
    id,
    incident_key,
    incident_key_canonical,
    title,
    description,
    status,
    severity,
    tlp,
    current_phase,
    primary_external_case_ref,
    created_by_user_id,
    created_at,
    updated_at,
    updated_by_user_id,
    incident_version,
    closed_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

func applyIncidentBundleIncidentTx(
	ctx context.Context,
	db incidentSourceDatabase,
	value any,
	importContext incidentSourceImportContext,
	contract incidentSourceContract,
) error {
	prepared, ok := value.(preparedIncidentSource)
	if !ok {
		return fmt.Errorf("%w: Incidents prepared import has type %T", errIncidentSourceCatalog, value)
	}
	if db == nil || !prepared.matches(contract, importContext) || importContext.actorUserID == uuid.Nil ||
		importContext.attributions == nil {
		return fmt.Errorf("%w: Incidents prepared import binding mismatch", errIncidentSourcePreparedBinding)
	}
	row := prepared.row
	rowID := row.ID.String()
	if err := importContext.attributions.RecordImportedAttribution(
		"incidents", rowID, "created_by_user_id", row.CreatedByUserID.String(),
	); err != nil {
		return err
	}
	if err := importContext.attributions.RecordImportedAttribution(
		"incidents", rowID, "updated_by_user_id", row.UpdatedByUserID.String(),
	); err != nil {
		return err
	}
	tag, err := db.Exec(
		ctx,
		insertIncidentSourceSQL,
		row.ID,
		row.IncidentKey,
		row.IncidentKeyCanonical,
		row.Title,
		nullableSourceStringArgument(row.Description),
		row.Status,
		nullableSourceStringArgument(row.Severity),
		nullableSourceStringArgument(row.TLP),
		nullableSourceStringArgument(row.CurrentPhase),
		nullableSourceStringArgument(row.PrimaryExternalCaseRef),
		importContext.actorUserID,
		row.CreatedAt,
		row.UpdatedAt,
		importContext.actorUserID,
		row.IncidentVersion,
		nullableSourceTimeArgument(row.ClosedAt),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return incidentportability.FixedImportFailure(prepared.logicalPath)
	}
	return nil
}

func nullableSourceStringArgument(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableSourceTimeArgument(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

const selectIncidentSourceCountSQL = `SELECT count(*) FROM incidents WHERE id = $1`

const selectIncidentSourceValidationSQL = `
SELECT
    id,
    incident_key,
    incident_key_canonical,
    title,
    description,
    status,
    severity,
    tlp,
    current_phase,
    primary_external_case_ref,
    created_by_user_id,
    created_at,
    updated_at,
    updated_by_user_id,
    incident_version,
    closed_at
FROM incidents
WHERE id = $1`

func validateIncidentBundleIncidentTx(
	ctx context.Context,
	db incidentSourceDatabase,
	value any,
	importContext incidentSourceImportContext,
	contract incidentSourceContract,
) error {
	prepared, ok := value.(preparedIncidentSource)
	if !ok {
		return fmt.Errorf("%w: Incidents prepared validation has type %T", errIncidentSourceCatalog, value)
	}
	if db == nil || !prepared.matches(contract, importContext) || importContext.actorUserID == uuid.Nil {
		return fmt.Errorf("%w: Incidents prepared validation binding mismatch", errIncidentSourcePreparedBinding)
	}
	var count int
	if err := db.QueryRow(ctx, selectIncidentSourceCountSQL, prepared.incidentID).Scan(&count); err != nil {
		return err
	}
	if count != 1 {
		return contract.failure(incidentSourceIdentityInvariant)
	}
	stored, err := scanStoredIncidentSourceRow(db.QueryRow(ctx, selectIncidentSourceValidationSQL, prepared.incidentID))
	if err != nil {
		return err
	}
	expected := prepared.row
	expected.CreatedByUserID = importContext.actorUserID
	expected.UpdatedByUserID = importContext.actorUserID
	if stored.ID != expected.ID {
		return contract.failure(incidentSourceIdentityInvariant)
	}
	if stored.IncidentKey != expected.IncidentKey ||
		stored.IncidentKeyCanonical != expected.IncidentKeyCanonical ||
		stored.Title != expected.Title || !equalOptionalString(stored.Description, expected.Description) ||
		stored.Status != expected.Status || !equalOptionalString(stored.Severity, expected.Severity) ||
		!equalOptionalString(stored.TLP, expected.TLP) ||
		!equalOptionalString(stored.CurrentPhase, expected.CurrentPhase) ||
		!equalOptionalString(stored.PrimaryExternalCaseRef, expected.PrimaryExternalCaseRef) ||
		!equalOptionalTime(stored.ClosedAt, expected.ClosedAt) {
		return contract.failure(incidentIdentityLifecycleInvariant)
	}
	if stored.CreatedByUserID != expected.CreatedByUserID ||
		!stored.CreatedAt.Equal(expected.CreatedAt) || !stored.UpdatedAt.Equal(expected.UpdatedAt) ||
		stored.UpdatedByUserID != expected.UpdatedByUserID || stored.IncidentVersion != expected.IncidentVersion {
		return contract.failure(incidentAttributionVersionInvariant)
	}
	return nil
}

func scanStoredIncidentSourceRow(row pgx.Row) (incidentSourceRow, error) {
	var (
		id                     pgtype.UUID
		description            pgtype.Text
		severity               pgtype.Text
		tlp                    pgtype.Text
		currentPhase           pgtype.Text
		primaryExternalCaseRef pgtype.Text
		createdByUserID        pgtype.UUID
		createdAt              pgtype.Timestamptz
		updatedAt              pgtype.Timestamptz
		updatedByUserID        pgtype.UUID
		closedAt               pgtype.Timestamptz
		result                 incidentSourceRow
	)
	if err := row.Scan(
		&id,
		&result.IncidentKey,
		&result.IncidentKeyCanonical,
		&result.Title,
		&description,
		&result.Status,
		&severity,
		&tlp,
		&currentPhase,
		&primaryExternalCaseRef,
		&createdByUserID,
		&createdAt,
		&updatedAt,
		&updatedByUserID,
		&result.IncidentVersion,
		&closedAt,
	); err != nil {
		return incidentSourceRow{}, err
	}
	var err error
	if result.ID, err = uuidFromPG(id); err != nil {
		return incidentSourceRow{}, fmt.Errorf("incident source id: %w", err)
	}
	if result.CreatedByUserID, err = uuidFromPG(createdByUserID); err != nil {
		return incidentSourceRow{}, fmt.Errorf("incident source created actor: %w", err)
	}
	if result.CreatedAt, err = timeFromPG(createdAt); err != nil {
		return incidentSourceRow{}, fmt.Errorf("incident source created time: %w", err)
	}
	if result.UpdatedAt, err = timeFromPG(updatedAt); err != nil {
		return incidentSourceRow{}, fmt.Errorf("incident source updated time: %w", err)
	}
	if result.UpdatedByUserID, err = uuidFromPG(updatedByUserID); err != nil {
		return incidentSourceRow{}, fmt.Errorf("incident source updated actor: %w", err)
	}
	result.Description = optionalStringFromPG(description)
	result.Severity = optionalStringFromPG(severity)
	result.TLP = optionalStringFromPG(tlp)
	result.CurrentPhase = optionalStringFromPG(currentPhase)
	result.PrimaryExternalCaseRef = optionalStringFromPG(primaryExternalCaseRef)
	result.ClosedAt = optionalTimeFromPG(closedAt)
	return result, nil
}

func equalOptionalString(left *string, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func equalOptionalTime(left *time.Time, right *time.Time) bool {
	return left == nil && right == nil || left != nil && right != nil && left.Equal(*right)
}
