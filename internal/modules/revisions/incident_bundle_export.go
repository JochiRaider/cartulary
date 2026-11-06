package revisions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func exportIncidentBundleFiles(ctx context.Context, input sourceport.ExportContext) error {
	for _, member := range []struct {
		path  string
		write func(context.Context, sourceport.ExportContext, io.Writer) error
	}{
		{changeSetsBundlePath, exportPortableChangeSets},
		{changeSetMutationsPath, exportPortableMutations},
		{recordRevisionsBundlePath, exportPortableRecordRevisions},
	} {
		if err := incidentportability.WriteMemberFrom(ctx, input.Members, member.path, func(ctx context.Context, output io.Writer) error { return member.write(ctx, input, output) }); err != nil {
			return err
		}
	}
	return nil
}

func exportPortableChangeSets(ctx context.Context, input sourceport.ExportContext, output io.Writer) error {
	var after *uuid.UUID
	for {
		next, err := exportPortableChangeSetPage(ctx, input, after, output)
		if err != nil {
			return err
		}
		if next == nil {
			return nil
		}
		after = next
	}
}

func exportPortableChangeSetPage(ctx context.Context, exportContext sourceport.ExportContext, after *uuid.UUID, output io.Writer) (*uuid.UUID, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT change_set_id, incident_id, actor_user_id, source, reason,
       client_txn_id, request_id, created_at
  FROM change_sets
 WHERE incident_id = $1 AND ($2::uuid IS NULL OR change_set_id > $2)
 ORDER BY change_set_id
 LIMIT 128
`, exportContext.IncidentID, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changeSets := make([]portableChangeSet, 0)
	rowIDs := make([]string, 0)
	for rows.Next() {
		var row portableChangeSet
		var reason, clientTxnID, requestID pgtype.Text
		if err := rows.Scan(
			&row.ChangeSetID,
			&row.IncidentID,
			&row.RuntimeActorID,
			&row.Source,
			&reason,
			&clientTxnID,
			&requestID,
			&row.CreatedAt,
		); err != nil {
			return nil, err
		}
		if row.ChangeSetID == uuid.Nil || row.IncidentID != exportContext.IncidentID ||
			row.RuntimeActorID == uuid.Nil || !validRequiredPortableString(row.Source) {
			return nil, errors.New("revisions export change set is invalid")
		}
		row.Reason = textPointer(reason)
		row.ClientTxnID = textPointer(clientTxnID)
		row.RequestID = textPointer(requestID)
		if !validNullablePortableString(row.Reason) ||
			!validNullablePortableString(row.ClientTxnID) ||
			!validNullablePortableString(row.RequestID) {
			return nil, errors.New("revisions export change set text is invalid")
		}
		changeSets = append(changeSets, row)
		rowIDs = append(rowIDs, row.ChangeSetID.String())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	portableActors := map[string]string{}
	if exportContext.PortableAttributions != nil && len(rowIDs) > 0 {
		portableActors, err = exportContext.PortableAttributions.ResolvePortableSourceActors(
			ctx,
			exportContext.Query,
			exportContext.IncidentID,
			"change_sets",
			"actor_user_id",
			rowIDs,
		)
		if err != nil {
			return nil, err
		}
	}
	rows.Close()
	for _, row := range changeSets {
		actorID, err := incidentportability.ExportActorID(exportContext.Actors, portableActors[row.ChangeSetID.String()], row.RuntimeActorID)
		if err != nil {
			return nil, err
		}
		err = incidentportability.WriteNDJSONRecord(output, map[string]any{
			"change_set_id": row.ChangeSetID.String(),
			"incident_id":   row.IncidentID.String(),
			"actor_user_id": actorID,
			"source":        row.Source,
			"reason":        nullableStringValue(row.Reason),
			"client_txn_id": nullableStringValue(row.ClientTxnID),
			"request_id":    nullableStringValue(row.RequestID),
			"created_at":    row.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
		if err != nil {
			return nil, err
		}
	}
	if len(changeSets) == 0 {
		return nil, nil
	}
	last := changeSets[len(changeSets)-1].ChangeSetID
	return &last, nil
}

func exportPortableMutations(
	ctx context.Context,
	exportContext sourceport.ExportContext,
	output io.Writer,
) error {
	rows, err := exportContext.Query.Query(ctx, `
SELECT mutation.change_set_id, mutation.sequence_no, mutation.target_kind,
       mutation.target_id, mutation.operation_kind, mutation.before_version_id,
       mutation.after_version_id, mutation.before_value, mutation.after_value
  FROM change_set_mutations mutation
  JOIN change_sets change_set
    ON change_set.change_set_id = mutation.change_set_id
 WHERE change_set.incident_id = $1
 ORDER BY mutation.change_set_id, mutation.sequence_no
`, exportContext.IncidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var row portableChangeSetMutation
		var beforeVersion, afterVersion pgtype.Text
		var beforeRaw, afterRaw []byte
		if err := rows.Scan(
			&row.ChangeSetID,
			&row.SequenceNo,
			&row.TargetKind,
			&row.TargetID,
			&row.OperationKind,
			&beforeVersion,
			&afterVersion,
			&beforeRaw,
			&afterRaw,
		); err != nil {
			return err
		}
		row.BeforeVersionID = textPointer(beforeVersion)
		row.AfterVersionID = textPointer(afterVersion)
		row.BeforeValue, err = decodeDatabaseJSON(beforeRaw)
		if err != nil {
			return err
		}
		row.AfterValue, err = decodeDatabaseJSON(afterRaw)
		if err != nil {
			return err
		}
		err = incidentportability.WriteNDJSONRecord(output, map[string]any{
			"change_set_id":     row.ChangeSetID.String(),
			"sequence_no":       row.SequenceNo,
			"target_kind":       row.TargetKind,
			"target_id":         row.TargetID,
			"operation_kind":    row.OperationKind,
			"before_version_id": nullableStringValue(row.BeforeVersionID),
			"after_version_id":  nullableStringValue(row.AfterVersionID),
			"before_value":      row.BeforeValue,
			"after_value":       row.AfterValue,
		})
		if err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func exportPortableRecordRevisions(
	ctx context.Context,
	exportContext sourceport.ExportContext,
	output io.Writer,
) error {
	rows, err := exportContext.Query.Query(ctx, `
SELECT revision.revision_id, revision.change_set_id, revision.record_id,
       revision.row_version, revision.before_json, revision.after_json,
       revision.created_at
  FROM record_revisions revision
  JOIN change_sets change_set
    ON change_set.change_set_id = revision.change_set_id
 WHERE change_set.incident_id = $1
 ORDER BY revision.revision_id
`, exportContext.IncidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var row portableRecordRevision
		var beforeRaw, afterRaw []byte
		if err := rows.Scan(
			&row.RevisionID,
			&row.ChangeSetID,
			&row.RecordID,
			&row.RowVersion,
			&beforeRaw,
			&afterRaw,
			&row.CreatedAt,
		); err != nil {
			return err
		}
		before, err := decodeDatabaseJSONObject(beforeRaw)
		if err != nil {
			return err
		}
		after, err := decodeDatabaseJSONObject(afterRaw)
		if err != nil {
			return err
		}
		err = incidentportability.WriteNDJSONRecord(output, map[string]any{
			"revision_id":   row.RevisionID,
			"change_set_id": row.ChangeSetID.String(),
			"record_id":     row.RecordID.String(),
			"row_version":   row.RowVersion,
			"before_json":   before,
			"after_json":    after,
			"created_at":    row.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
		if err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func nullableStringValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func decodeDatabaseJSON(raw []byte) (any, error) {
	if raw == nil {
		return nil, nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func decodeDatabaseJSONObject(raw []byte) (map[string]any, error) {
	value, err := decodeDatabaseJSON(raw)
	if err != nil || value == nil {
		return nil, err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("revisions history snapshot is not an object")
	}
	return object, nil
}
