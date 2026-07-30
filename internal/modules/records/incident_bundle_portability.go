package records

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
)

const recordsBundlePath = "data/records.ndjson"

var positiveDecimalInteger = regexp.MustCompile(`^[1-9][0-9]*$`)

type recordsPortableAttributionResolver interface {
	ResolvePortableSourceActors(
		context.Context,
		incidentportability.Queryer,
		uuid.UUID,
		string,
		string,
		[]string,
	) (map[string]string, error)
}

type recordsExportContext struct {
	Query                incidentportability.Queryer
	IncidentID           uuid.UUID
	PortableAttributions recordsPortableAttributionResolver
}

type recordsImportContext struct {
	IncidentID    uuid.UUID
	ActorUserID   uuid.UUID
	Attributions  incidentportability.AttributionRecorder
	ActorAdmitted func(string) bool
}

type recordsInvariantError struct {
	InvariantID string
}

func (e *recordsInvariantError) Error() string {
	return "records failed invariant " + e.InvariantID
}

type recordsPortableRow struct {
	RecordID            uuid.UUID
	IncidentID          uuid.UUID
	RecordType          subtypepresence.RecordType
	CreatedAt           time.Time
	PortableCreatedByID uuid.UUID
	RuntimeCreatedByID  uuid.UUID
	UpdatedAt           time.Time
	PortableUpdatedByID uuid.UUID
	RuntimeUpdatedByID  uuid.UUID
	RowVersion          int64
	DeletedAt           *time.Time
	PortableDeletedByID *uuid.UUID
	RuntimeDeletedByID  *uuid.UUID
}

type preparedRecordsImport struct {
	rows []recordsPortableRow
}

func exportIncidentBundleFiles(
	ctx context.Context,
	exportContext recordsExportContext,
) ([]incidentportability.File, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT record_id,
       incident_id,
       record_type,
       created_at,
       created_by_user_id,
       updated_at,
       updated_by_user_id,
       row_version,
       deleted_at,
       deleted_by_user_id
  FROM records
 WHERE incident_id = $1
 ORDER BY record_id
`, exportContext.IncidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exportRows := make([]recordsPortableRow, 0)
	allRowIDs := make([]string, 0)
	deletedRowIDs := make([]string, 0)
	for rows.Next() {
		var (
			row             recordsPortableRow
			recordTypeText  string
			deletedAt       pgtype.Timestamptz
			deletedByUserID pgtype.UUID
		)
		if err := rows.Scan(
			&row.RecordID,
			&row.IncidentID,
			&recordTypeText,
			&row.CreatedAt,
			&row.RuntimeCreatedByID,
			&row.UpdatedAt,
			&row.RuntimeUpdatedByID,
			&row.RowVersion,
			&deletedAt,
			&deletedByUserID,
		); err != nil {
			return nil, err
		}
		row.RecordType = subtypepresence.RecordType(recordTypeText)
		if deletedAt.Valid {
			value := deletedAt.Time.UTC()
			row.DeletedAt = &value
		}
		if deletedByUserID.Valid {
			value, err := uuid.FromBytes(deletedByUserID.Bytes[:])
			if err != nil {
				return nil, errors.New("records export delete actor is invalid")
			}
			row.RuntimeDeletedByID = &value
		}
		if err := validateRuntimePortableRow(row, exportContext.IncidentID); err != nil {
			return nil, err
		}
		rowID := row.RecordID.String()
		allRowIDs = append(allRowIDs, rowID)
		if row.DeletedAt != nil {
			deletedRowIDs = append(deletedRowIDs, rowID)
		}
		exportRows = append(exportRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	createdActors, err := resolvePortableActors(
		ctx, exportContext, "created_by_user_id", allRowIDs,
	)
	if err != nil {
		return nil, err
	}
	updatedActors, err := resolvePortableActors(
		ctx, exportContext, "updated_by_user_id", allRowIDs,
	)
	if err != nil {
		return nil, err
	}
	deletedActors, err := resolvePortableActors(
		ctx, exportContext, "deleted_by_user_id", deletedRowIDs,
	)
	if err != nil {
		return nil, err
	}

	var payload bytes.Buffer
	for _, row := range exportRows {
		rowID := row.RecordID.String()
		createdBy, err := portableActor(createdActors[rowID], row.RuntimeCreatedByID)
		if err != nil {
			return nil, err
		}
		updatedBy, err := portableActor(updatedActors[rowID], row.RuntimeUpdatedByID)
		if err != nil {
			return nil, err
		}
		var deletedAt any
		var deletedBy any
		if row.DeletedAt != nil {
			deletedAt = row.DeletedAt.UTC().Format(time.RFC3339Nano)
			if row.RuntimeDeletedByID == nil {
				return nil, errors.New("records export delete tuple is invalid")
			}
			deletedBy, err = portableActor(deletedActors[rowID], *row.RuntimeDeletedByID)
			if err != nil {
				return nil, err
			}
		}
		encoded, err := incidentportability.CanonicalJSONString(map[string]any{
			"record_id":          row.RecordID.String(),
			"incident_id":        row.IncidentID.String(),
			"record_type":        string(row.RecordType),
			"created_at":         row.CreatedAt.UTC().Format(time.RFC3339Nano),
			"created_by_user_id": createdBy,
			"updated_at":         row.UpdatedAt.UTC().Format(time.RFC3339Nano),
			"updated_by_user_id": updatedBy,
			"row_version":        row.RowVersion,
			"deleted_at":         deletedAt,
			"deleted_by_user_id": deletedBy,
		})
		if err != nil {
			return nil, err
		}
		payload.Write(encoded)
	}
	return []incidentportability.File{{
		Path:    recordsBundlePath,
		Payload: payload.Bytes(),
	}}, nil
}

func resolvePortableActors(
	ctx context.Context,
	exportContext recordsExportContext,
	sourceColumn string,
	sourceRowIDs []string,
) (map[string]string, error) {
	if exportContext.PortableAttributions == nil || len(sourceRowIDs) == 0 {
		return map[string]string{}, nil
	}
	return exportContext.PortableAttributions.ResolvePortableSourceActors(
		ctx,
		exportContext.Query,
		exportContext.IncidentID,
		"records",
		sourceColumn,
		sourceRowIDs,
	)
}

func portableActor(sourceActorID string, runtimeActorID uuid.UUID) (string, error) {
	if sourceActorID == "" {
		if runtimeActorID == uuid.Nil {
			return "", errors.New("records export actor is invalid")
		}
		return runtimeActorID.String(), nil
	}
	parsed, err := uuid.Parse(sourceActorID)
	if err != nil || parsed.String() != sourceActorID {
		return "", errors.New("records export attribution is invalid")
	}
	return sourceActorID, nil
}

func validateRuntimePortableRow(row recordsPortableRow, incidentID uuid.UUID) error {
	if row.RecordID == uuid.Nil ||
		row.IncidentID != incidentID ||
		!admittedRecordType(row.RecordType) ||
		row.RuntimeCreatedByID == uuid.Nil ||
		row.RuntimeUpdatedByID == uuid.Nil ||
		row.RowVersion < 1 ||
		row.UpdatedAt.Before(row.CreatedAt) ||
		(row.DeletedAt == nil) != (row.RuntimeDeletedByID == nil) {
		return errors.New("records export envelope is invalid")
	}
	if row.DeletedAt != nil &&
		(row.DeletedAt.Before(row.CreatedAt) || row.DeletedAt.After(row.UpdatedAt)) {
		return errors.New("records export delete tuple is invalid")
	}
	return nil
}

func prepareRecordsImport(
	bundle interface{ File(string) ([]byte, bool) },
	importContext recordsImportContext,
) (preparedRecordsImport, error) {
	payload, ok := bundle.File(recordsBundlePath)
	if !ok {
		return preparedRecordsImport{}, recordsInvariantFailure("records.envelope_legal")
	}
	rows, err := incidentportability.DecodeStrictNDJSONObjects(payload, recordsBundlePath)
	if err != nil {
		return preparedRecordsImport{}, recordsInvariantFailure("records.envelope_legal")
	}
	prepared := preparedRecordsImport{rows: make([]recordsPortableRow, 0, len(rows))}
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		parsed, err := prepareRecordsRow(row, importContext)
		if err != nil {
			return preparedRecordsImport{}, err
		}
		if _, duplicate := seen[parsed.RecordID]; duplicate {
			return preparedRecordsImport{}, recordsInvariantFailure("records.envelope_legal")
		}
		seen[parsed.RecordID] = struct{}{}
		prepared.rows = append(prepared.rows, parsed)
	}
	return prepared, nil
}

func prepareRecordsRow(
	row map[string]any,
	importContext recordsImportContext,
) (recordsPortableRow, error) {
	required := []string{
		"record_id", "incident_id", "record_type", "created_at",
		"created_by_user_id", "updated_at", "updated_by_user_id",
		"row_version", "deleted_at", "deleted_by_user_id",
	}
	if len(row) != len(required) {
		return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
	}
	for _, member := range required {
		if _, present := row[member]; !present {
			return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
		}
	}
	recordID, ok := parseCanonicalUUIDMember(row["record_id"])
	if !ok {
		return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
	}
	incidentID, ok := parseCanonicalUUIDMember(row["incident_id"])
	if !ok {
		return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
	}
	if incidentID != importContext.IncidentID {
		return recordsPortableRow{}, recordsInvariantFailure("records.incident_scope")
	}
	recordTypeText, ok := row["record_type"].(string)
	recordType := subtypepresence.RecordType(recordTypeText)
	if !ok || !admittedRecordType(recordType) {
		return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
	}
	createdAt, ok := parseCanonicalTimestampMember(row["created_at"])
	if !ok {
		return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
	}
	createdBy, ok := parseAdmittedActorMember(row["created_by_user_id"], importContext)
	if !ok {
		return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
	}
	updatedAt, ok := parseCanonicalTimestampMember(row["updated_at"])
	if !ok || updatedAt.Before(createdAt) {
		return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
	}
	updatedBy, ok := parseAdmittedActorMember(row["updated_by_user_id"], importContext)
	if !ok {
		return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
	}
	rowVersionNumber, ok := row["row_version"].(json.Number)
	if !ok || !positiveDecimalInteger.MatchString(rowVersionNumber.String()) {
		return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
	}
	rowVersion, err := strconv.ParseInt(rowVersionNumber.String(), 10, 64)
	if err != nil || rowVersion < 1 {
		return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
	}

	var deletedAt *time.Time
	var portableDeletedBy *uuid.UUID
	var runtimeDeletedBy *uuid.UUID
	switch rawDeletedAt := row["deleted_at"].(type) {
	case nil:
		if row["deleted_by_user_id"] != nil {
			return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
		}
	case string:
		parsedDeletedAt, ok := parseCanonicalTimestamp(rawDeletedAt)
		if !ok || parsedDeletedAt.Before(createdAt) || parsedDeletedAt.After(updatedAt) {
			return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
		}
		deletedActor, ok := parseAdmittedActorMember(row["deleted_by_user_id"], importContext)
		if !ok {
			return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
		}
		deletedAt = &parsedDeletedAt
		portableDeletedBy = &deletedActor
		runtimeActor := importContext.ActorUserID
		runtimeDeletedBy = &runtimeActor
	default:
		return recordsPortableRow{}, recordsInvariantFailure("records.envelope_legal")
	}
	return recordsPortableRow{
		RecordID:            recordID,
		IncidentID:          incidentID,
		RecordType:          recordType,
		CreatedAt:           createdAt,
		PortableCreatedByID: createdBy,
		RuntimeCreatedByID:  importContext.ActorUserID,
		UpdatedAt:           updatedAt,
		PortableUpdatedByID: updatedBy,
		RuntimeUpdatedByID:  importContext.ActorUserID,
		RowVersion:          rowVersion,
		DeletedAt:           deletedAt,
		PortableDeletedByID: portableDeletedBy,
		RuntimeDeletedByID:  runtimeDeletedBy,
	}, nil
}

func parseCanonicalUUIDMember(value any) (uuid.UUID, bool) {
	text, ok := value.(string)
	if !ok {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(text)
	return parsed, err == nil && parsed.String() == text
}

func parseAdmittedActorMember(
	value any,
	importContext recordsImportContext,
) (uuid.UUID, bool) {
	actorID, ok := parseCanonicalUUIDMember(value)
	if !ok || importContext.ActorAdmitted == nil ||
		!importContext.ActorAdmitted(actorID.String()) {
		return uuid.Nil, false
	}
	return actorID, true
}

func parseCanonicalTimestampMember(value any) (time.Time, bool) {
	text, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	return parseCanonicalTimestamp(text)
}

func parseCanonicalTimestamp(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.UTC().Format(time.RFC3339Nano) != value {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func admittedRecordType(recordType subtypepresence.RecordType) bool {
	switch recordType {
	case subtypepresence.RecordTypeTimelineEvent,
		subtypepresence.RecordTypeHost,
		subtypepresence.RecordTypeIdentity,
		subtypepresence.RecordTypeParty,
		subtypepresence.RecordTypeIndicator,
		subtypepresence.RecordTypeArtifact,
		subtypepresence.RecordTypeTaskRequest,
		subtypepresence.RecordTypeDecision,
		subtypepresence.RecordTypeEvidence,
		subtypepresence.RecordTypeAssessment:
		return true
	default:
		return false
	}
}

func applyPreparedRecordsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRecordsImport,
	importContext recordsImportContext,
) error {
	if importContext.ActorUserID == uuid.Nil || importContext.Attributions == nil {
		return recordsInvariantFailure("records.envelope_legal")
	}
	for _, row := range prepared.rows {
		tag, err := tx.Exec(ctx, `
INSERT INTO records (
    record_id,
    incident_id,
    record_type,
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    row_version,
    deleted_at,
    deleted_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`,
			row.RecordID,
			row.IncidentID,
			string(row.RecordType),
			row.RuntimeCreatedByID,
			row.CreatedAt,
			row.RuntimeUpdatedByID,
			row.UpdatedAt,
			row.RowVersion,
			row.DeletedAt,
			row.RuntimeDeletedByID,
		)
		if err != nil {
			return classifyRecordsApplyError(err)
		}
		if tag.RowsAffected() != 1 {
			return recordsInvariantFailure("records.envelope_legal")
		}
		for _, attribution := range []struct {
			column   string
			portable *uuid.UUID
		}{
			{column: "created_by_user_id", portable: &row.PortableCreatedByID},
			{column: "updated_by_user_id", portable: &row.PortableUpdatedByID},
			{column: "deleted_by_user_id", portable: row.PortableDeletedByID},
		} {
			if attribution.portable == nil {
				continue
			}
			if err := importContext.Attributions.RecordImportedAttribution(
				"records",
				row.RecordID.String(),
				attribution.column,
				attribution.portable.String(),
			); err != nil {
				return recordsInvariantFailure("records.envelope_legal")
			}
		}
	}
	return nil
}

func classifyRecordsApplyError(err error) error {
	var postgresFailure *pgconn.PgError
	if errors.As(err, &postgresFailure) {
		return recordsInvariantFailure("records.envelope_legal")
	}
	return err
}

func validatePreparedRecordsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRecordsImport,
	importContext recordsImportContext,
	subtypeCatalog *subtypepresence.Catalog,
) error {
	if subtypeCatalog == nil {
		return recordsInvariantFailure("records.subtype_complete")
	}
	expected := make(map[uuid.UUID]recordsPortableRow, len(prepared.rows))
	for _, row := range prepared.rows {
		expected[row.RecordID] = row
	}
	rows, err := tx.Query(ctx, `
SELECT record_id, incident_id, record_type
  FROM records
 WHERE incident_id = $1
 ORDER BY record_id
`, importContext.IncidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	envelopes := make([]subtypepresence.Envelope, 0, len(expected))
	seen := make(map[uuid.UUID]struct{}, len(expected))
	for rows.Next() {
		var envelope subtypepresence.Envelope
		if err := rows.Scan(
			&envelope.RecordID,
			&envelope.IncidentID,
			&envelope.RecordType,
		); err != nil {
			return err
		}
		expectedRow, present := expected[envelope.RecordID]
		if !present ||
			expectedRow.IncidentID != envelope.IncidentID ||
			expectedRow.RecordType != envelope.RecordType {
			return recordsInvariantFailure("records.envelope_legal")
		}
		if _, duplicate := seen[envelope.RecordID]; duplicate {
			return recordsInvariantFailure("records.envelope_legal")
		}
		seen[envelope.RecordID] = struct{}{}
		envelopes = append(envelopes, envelope)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(seen) != len(expected) {
		return recordsInvariantFailure("records.envelope_legal")
	}
	if err := subtypeCatalog.ValidateTx(ctx, tx, importContext.IncidentID, envelopes); err != nil {
		if errors.Is(err, subtypepresence.ErrIncompleteBindings) ||
			errors.Is(err, subtypepresence.ErrInvalidCatalog) {
			return recordsInvariantFailure("records.subtype_complete")
		}
		return err
	}
	return nil
}

func recordsInvariantFailure(invariantID string) error {
	return &recordsInvariantError{InvariantID: invariantID}
}
