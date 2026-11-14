package revisions

import (
	"context"
	"reflect"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func validatePreparedRevisionsSourceStateTx(ctx context.Context, tx pgx.Tx, prepared preparedRevisionsImport, validation *incidentBundleValidationCatalog) error {
	if err := prepared.validateChangeSetState(ctx, tx); err != nil {
		return err
	}
	if err := prepared.validateMutationState(ctx, tx, validation); err != nil {
		return err
	}
	return prepared.validateRevisionState(ctx, tx)
}

func (p preparedRevisionsImport) validateChangeSetState(ctx context.Context, tx pgx.Tx) error {
	rows, err := tx.Query(ctx, `
SELECT change_set_id, incident_id, actor_user_id, source, reason, client_txn_id, request_id, created_at, actor_import_contexts, history_order
FROM change_sets WHERE incident_id=$1 ORDER BY change_set_id
`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = p.walkChangeSets(ctx, func(expected portableChangeSet) error {
		if err := requireRevisionSourceRow(rows); err != nil {
			return err
		}
		var actual portableChangeSet
		var contexts []byte
		if err := rows.Scan(&actual.ChangeSetID, &actual.IncidentID, &actual.RuntimeActorID, &actual.Source,
			&actual.Reason, &actual.ClientTxnID, &actual.RequestID, &actual.CreatedAt, &contexts, &actual.HistoryOrder); err != nil {
			return err
		}
		actual.ActorImportContexts, err = decodeRevisionActorContexts(contexts)
		if err != nil {
			return revisionsFailure(revisionsSourceStateInvariant)
		}
		expected.ActorImportContexts = p.importedActorContexts(expected.ActorImportContexts)
		actual.CreatedAt = actual.CreatedAt.UTC()
		expected.PortableActorID = uuid.Nil
		expected.RuntimeActorID = p.runtimeActor
		if !reflect.DeepEqual(actual, expected) {
			return revisionsFailure(revisionsSourceStateInvariant)
		}
		return nil
	})
	return finishRevisionSourceRows(rows, err)
}

func (p preparedRevisionsImport) validateMutationState(ctx context.Context, tx pgx.Tx, validation *incidentBundleValidationCatalog) error {
	rows, err := tx.Query(ctx, `
SELECT m.change_set_id, m.sequence_no, m.target_kind, m.target_id, m.operation_kind,
       m.before_version_id, m.after_version_id, m.before_value, m.after_value,
       m.history_record_ids, m.history_entry_record_ids
FROM change_set_mutations m JOIN change_sets c USING(change_set_id)
WHERE c.incident_id=$1 ORDER BY m.change_set_id,m.sequence_no
`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = p.walkMutations(ctx, func(expected portableChangeSetMutation) error {
		if err := requireRevisionSourceRow(rows); err != nil {
			return err
		}
		var actual portableChangeSetMutation
		var before, after []byte
		var historyIDs, entryIDs []uuid.UUID
		if err := rows.Scan(&actual.ChangeSetID, &actual.SequenceNo, &actual.TargetKind, &actual.TargetID,
			&actual.OperationKind, &actual.BeforeVersionID, &actual.AfterVersionID, &before, &after, &historyIDs, &entryIDs); err != nil {
			return err
		}
		actual.BeforeValue, err = decodeDatabaseJSON(before)
		if err != nil {
			return err
		}
		actual.AfterValue, err = decodeDatabaseJSON(after)
		if err != nil {
			return err
		}
		history, err := validation.targets.DescribeValues(expected.TargetKind, expected.TargetID, expected.OperationKind, expected.BeforeValue, expected.AfterValue)
		if err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		if !reflect.DeepEqual(actual, expected) || !slices.Equal(historyIDs, history.HistoryRecordIDs) || !slices.Equal(entryIDs, history.HistoryEntryRecordIDs) {
			return revisionsFailure(revisionsSourceStateInvariant)
		}
		return nil
	})
	return finishRevisionSourceRows(rows, err)
}

func (p preparedRevisionsImport) validateRevisionState(ctx context.Context, tx pgx.Tx) error {
	rows, err := tx.Query(ctx, `
SELECT r.revision_id,r.change_set_id,r.record_id,r.row_version,r.before_json,r.after_json,r.created_at
FROM record_revisions r JOIN change_sets c USING(change_set_id)
WHERE c.incident_id=$1 ORDER BY r.revision_id
`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = p.walkRevisions(ctx, func(expected portableRecordRevision) error {
		if err := requireRevisionSourceRow(rows); err != nil {
			return err
		}
		var actual portableRecordRevision
		var before, after []byte
		if err := rows.Scan(&actual.RevisionID, &actual.ChangeSetID, &actual.RecordID, &actual.RowVersion, &before, &after, &actual.CreatedAt); err != nil {
			return err
		}
		actual.BeforeJSON, err = decodeDatabaseJSONObject(before)
		if err != nil {
			return revisionsFailure(revisionsSourceStateInvariant)
		}
		actual.AfterJSON, err = decodeDatabaseJSONObject(after)
		if err != nil {
			return revisionsFailure(revisionsSourceStateInvariant)
		}
		actual.CreatedAt = actual.CreatedAt.UTC()
		if !reflect.DeepEqual(actual, expected) {
			return revisionsFailure(revisionsSourceStateInvariant)
		}
		return nil
	})
	return finishRevisionSourceRows(rows, err)
}

func requireRevisionSourceRow(rows pgx.Rows) error {
	if rows.Next() {
		return nil
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return revisionsFailure(revisionsSourceStateInvariant)
}

func finishRevisionSourceRows(rows pgx.Rows, err error) error {
	if err != nil {
		return err
	}
	if rows.Next() {
		return revisionsFailure(revisionsSourceStateInvariant)
	}
	return rows.Err()
}
