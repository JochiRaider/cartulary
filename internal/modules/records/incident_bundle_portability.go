package records

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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
	Actors               incidentportability.ActorReferences
	Query                incidentportability.Queryer
	IncidentID           uuid.UUID
	PortableAttributions recordsPortableAttributionResolver
}

type recordsImportContext struct {
	IncidentID    uuid.UUID
	ActorUserID   uuid.UUID
	Attributions  incidentportability.AttributionRecorder
	ActorAdmitted func(string) bool
	// Preparation holds the port's staged write-ready rows and attribution.
	Preparation recordsPreparationStore
}

type recordsPreparationStore interface {
	incidentportability.MemberReader
	incidentportability.MemberWriter
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

// preparedRecordsImport is the only state the final transaction receives:
// sealed write-ready rows and attribution, never a bundle member reader
// (Core 01 REQ-01-641).
type preparedRecordsImport struct {
	rows        incidentportability.PreparedRowStream
	attribution incidentportability.AttributionStream
	envelopes   subtypepresence.Summary
	incidentID  uuid.UUID
	actorID     uuid.UUID
}

const (
	preparedRecordsPath            = "records.rows"
	preparedRecordsAttributionPath = "attribution.rows"
)

var portableRecordColumns = []string{
	"record_id", "incident_id", "record_type", "created_by_user_id", "created_at",
	"updated_by_user_id", "updated_at", "row_version", "deleted_at", "deleted_by_user_id",
}

// recordRowCodec stages one admitted envelope with both its historical and
// importing actors; COPY writes the importing actors.
type recordRowCodec struct{}

func (recordRowCodec) Encode(row recordsPortableRow) ([]byte, error) {
	return json.Marshal([]any{row.RecordID, row.IncidentID, row.RecordType, row.CreatedAt, row.PortableCreatedByID,
		row.RuntimeCreatedByID, row.UpdatedAt, row.PortableUpdatedByID, row.RuntimeUpdatedByID, row.RowVersion,
		row.DeletedAt, row.PortableDeletedByID, row.RuntimeDeletedByID})
}

func (recordRowCodec) Decode(line []byte) (recordsPortableRow, error) {
	var row recordsPortableRow
	var raw []json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil || len(raw) != 13 {
		return row, incidentportability.ErrPreparedStream
	}
	for index, target := range []any{&row.RecordID, &row.IncidentID, &row.RecordType, &row.CreatedAt, &row.PortableCreatedByID,
		&row.RuntimeCreatedByID, &row.UpdatedAt, &row.PortableUpdatedByID, &row.RuntimeUpdatedByID, &row.RowVersion,
		&row.DeletedAt, &row.PortableDeletedByID, &row.RuntimeDeletedByID} {
		if err := json.Unmarshal(raw[index], target); err != nil {
			return row, incidentportability.ErrPreparedStream
		}
	}
	return row, nil
}

func (recordRowCodec) Columns() []string { return portableRecordColumns }

func (recordRowCodec) Values(row recordsPortableRow) ([]any, error) {
	return []any{row.RecordID, row.IncidentID, string(row.RecordType), row.RuntimeCreatedByID, row.CreatedAt,
		row.RuntimeUpdatedByID, row.UpdatedAt, row.RowVersion, row.DeletedAt, row.RuntimeDeletedByID}, nil
}

func exportIncidentBundleFiles(ctx context.Context, exportContext recordsExportContext, members incidentportability.MemberWriter) error {
	return incidentportability.WriteMemberFrom(ctx, members, recordsBundlePath, func(ctx context.Context, output io.Writer) error {
		var after *uuid.UUID
		for {
			next, err := exportIncidentBundlePage(ctx, exportContext, after, output)
			if err != nil {
				return err
			}
			if next == nil {
				return nil
			}
			after = next
		}
	})
}

func exportIncidentBundlePage(ctx context.Context, exportContext recordsExportContext, after *uuid.UUID, output io.Writer) (*uuid.UUID, error) {
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
 WHERE incident_id = $1 AND ($2::uuid IS NULL OR record_id > $2)
 ORDER BY record_id
 LIMIT 128
`, exportContext.IncidentID, after)
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

	rows.Close()
	for _, row := range exportRows {
		rowID := row.RecordID.String()
		createdBy, err := incidentportability.ExportActorID(exportContext.Actors, createdActors[rowID], row.RuntimeCreatedByID)
		if err != nil {
			return nil, err
		}
		updatedBy, err := incidentportability.ExportActorID(exportContext.Actors, updatedActors[rowID], row.RuntimeUpdatedByID)
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
			deletedBy, err = incidentportability.ExportActorID(exportContext.Actors, deletedActors[rowID], *row.RuntimeDeletedByID)
			if err != nil {
				return nil, err
			}
		}
		err = incidentportability.WriteNDJSONRecord(output, map[string]any{
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
	}
	if len(exportRows) == 0 {
		return nil, nil
	}
	last := exportRows[len(exportRows)-1].RecordID
	return &last, nil
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

func prepareRecordsImport(ctx context.Context, bundle incidentportability.MemberReader, importContext recordsImportContext) (preparedRecordsImport, error) {
	if bundle == nil {
		return preparedRecordsImport{}, recordsInvariantFailure("records.envelope_legal")
	}
	var member incidentportability.Member
	for _, candidate := range bundle.Members() {
		if candidate.Path == recordsBundlePath {
			member = candidate
			break
		}
	}
	if member.Path == "" {
		return preparedRecordsImport{}, recordsInvariantFailure("records.envelope_legal")
	}
	store := importContext.Preparation
	if store == nil {
		return preparedRecordsImport{}, errors.New("records: incident bundle preparation staging is required")
	}
	inventory := subtypepresence.NewInventory(importContext.IncidentID)
	previous := ""
	rows, err := incidentportability.StagePreparedRows(ctx, store, preparedRecordsPath, recordRowCodec{}, func(emit func(recordsPortableRow) error) error {
		return incidentportability.WalkMember(ctx, bundle, member, func(raw map[string]any) error {
			row, err := prepareRecordsRow(raw, importContext)
			if err != nil {
				return err
			}
			id := row.RecordID.String()
			if id <= previous {
				return recordsInvariantFailure("records.envelope_legal")
			}
			previous = id
			if err := inventory.Add(subtypepresence.Binding{RecordID: row.RecordID, IncidentID: row.IncidentID, RecordType: row.RecordType}); err != nil {
				return recordsInvariantFailure("records.envelope_legal")
			}
			return emit(row)
		})
	})
	if err != nil {
		if ctx.Err() != nil {
			return preparedRecordsImport{}, ctx.Err()
		}
		var malformed *incidentportability.StrictNDJSONDecodeError
		if errors.As(err, &malformed) {
			return preparedRecordsImport{}, recordsInvariantFailure("records.envelope_legal")
		}
		return preparedRecordsImport{}, err
	}
	// Historical actors are admitted while parsing; one pass per column keeps
	// the grouped attribution lines dense.
	admitted := func(actorID string) bool {
		return importContext.ActorAdmitted != nil && importContext.ActorAdmitted(actorID)
	}
	attribution, err := incidentportability.StageAttributions(ctx, store, preparedRecordsAttributionPath, admitted, func(emit func(incidentportability.AttributionRow) error) error {
		for _, column := range []string{"created_by_user_id", "updated_by_user_id", "deleted_by_user_id"} {
			if err := incidentportability.ReadPreparedRows(ctx, store, rows, recordRowCodec{}, func(row recordsPortableRow) error {
				portable := map[string]*uuid.UUID{"created_by_user_id": &row.PortableCreatedByID, "updated_by_user_id": &row.PortableUpdatedByID, "deleted_by_user_id": row.PortableDeletedByID}[column]
				if portable == nil {
					return nil
				}
				return emit(incidentportability.AttributionRow{SourceTable: "records", SourceRowID: row.RecordID.String(), SourceColumn: column, SourceActorID: portable.String()})
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return preparedRecordsImport{}, err
	}
	return preparedRecordsImport{rows: rows, attribution: attribution, envelopes: inventory.Summary(), incidentID: importContext.IncidentID, actorID: importContext.ActorUserID}, nil
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
	if err != nil || parsed.Year() < 1 || parsed.Year() > 9999 || parsed.Nanosecond()%1_000 != 0 || parsed.UTC().Format(time.RFC3339Nano) != value {
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
	if importContext.ActorUserID == uuid.Nil || importContext.Attributions == nil || prepared.incidentID != importContext.IncidentID || prepared.actorID != importContext.ActorUserID {
		return recordsInvariantFailure("records.envelope_legal")
	}
	if importContext.Preparation == nil {
		return errors.New("records: incident bundle preparation is required")
	}
	// One COPY of the staged envelopes; the acknowledgement must equal the
	// staged count. Attribution is flushed by the core in step 9.
	source, err := incidentportability.OpenPreparedCopy(ctx, importContext.Preparation, prepared.rows, recordRowCodec{})
	if err != nil {
		return err
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"records"}, portableRecordColumns, source)
	if err := source.Finish(written, err); err != nil {
		return classifyRecordsApplyError(err)
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
	if prepared.incidentID != importContext.IncidentID || prepared.actorID != importContext.ActorUserID {
		return recordsInvariantFailure("records.envelope_legal")
	}
	actual := subtypepresence.NewInventory(importContext.IncidentID)
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
	for rows.Next() {
		var envelope subtypepresence.Binding
		if err := rows.Scan(
			&envelope.RecordID,
			&envelope.IncidentID,
			&envelope.RecordType,
		); err != nil {
			return err
		}
		if err := actual.Add(envelope); err != nil {
			return recordsInvariantFailure("records.envelope_legal")
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()
	if !prepared.envelopes.Equal(actual.Summary()) {
		return recordsInvariantFailure("records.envelope_legal")
	}
	if err := subtypeCatalog.ValidateTx(ctx, tx, importContext.IncidentID, prepared.envelopes); err != nil {
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
