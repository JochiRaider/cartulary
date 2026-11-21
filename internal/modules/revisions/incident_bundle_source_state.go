package revisions

import (
	"context"
	"reflect"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// validatePreparedRevisionsSourceStateTx compares the stored rows with the
// staged write-ready rows, never with bundle members. Imported rows must be
// exactly what preparation staged, after database normalization.
func validatePreparedRevisionsSourceStateTx(ctx context.Context, tx pgx.Tx, prepared preparedRevisionsImport, store incidentportability.MemberReader) error {
	if err := prepared.validateChangeSetState(ctx, tx, store); err != nil {
		return err
	}
	if err := prepared.validateMutationState(ctx, tx, store); err != nil {
		return err
	}
	return prepared.validateRevisionState(ctx, tx, store)
}

func (p preparedRevisionsImport) validateChangeSetState(ctx context.Context, tx pgx.Tx, store incidentportability.MemberReader) error {
	rows, err := tx.Query(ctx, `
SELECT change_set_id, incident_id, actor_user_id, source, reason, client_txn_id, request_id, created_at, actor_import_contexts, history_order
FROM change_sets WHERE incident_id=$1 ORDER BY change_set_id
`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = incidentportability.ReadPreparedRows(ctx, store, p.changeSets, changeSetRowCodec{}, func(staged stagedChangeSet) error {
		expected := portableChangeSet{
			HistoryOrder: staged.HistoryOrder, ActorImportContexts: staged.Contexts,
			ChangeSetID: staged.ChangeSetID, IncidentID: staged.IncidentID, RuntimeActorID: p.runtimeActor,
			Source: staged.Source, Reason: staged.Reason, ClientTxnID: staged.ClientTxnID,
			RequestID: staged.RequestID, CreatedAt: staged.CreatedAt,
		}
		if err := requireRevisionSourceRow(rows); err != nil {
			return err
		}
		var actual portableChangeSet
		var contexts []byte
		var historyOrder pgtype.Int8
		if err := rows.Scan(&actual.ChangeSetID, &actual.IncidentID, &actual.RuntimeActorID, &actual.Source,
			&actual.Reason, &actual.ClientTxnID, &actual.RequestID, &actual.CreatedAt, &contexts, &historyOrder); err != nil {
			return err
		}
		// Imported rows carry explicit orders. A row without one was appended
		// natively inside this import transaction and is not bundle source state.
		if !historyOrder.Valid {
			return revisionsFailure(revisionsSourceStateInvariant)
		}
		actual.HistoryOrder = historyOrder.Int64
		actual.ActorImportContexts, err = decodeRevisionActorContexts(contexts)
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

func (p preparedRevisionsImport) validateMutationState(ctx context.Context, tx pgx.Tx, store incidentportability.MemberReader) error {
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
	err = incidentportability.ReadPreparedRows(ctx, store, p.mutations, mutationRowCodec{}, func(staged stagedMutation) error {
		expected := portableChangeSetMutation{
			ChangeSetID: staged.ChangeSetID, SequenceNo: staged.SequenceNo, TargetKind: staged.TargetKind,
			TargetID: staged.TargetID, OperationKind: staged.OperationKind,
			BeforeVersionID: staged.BeforeVersionID, AfterVersionID: staged.AfterVersionID,
		}
		var err error
		if expected.BeforeValue, err = decodeDatabaseJSON(staged.BeforeValue); err != nil {
			return incidentportability.ErrPreparedStream
		}
		if expected.AfterValue, err = decodeDatabaseJSON(staged.AfterValue); err != nil {
			return incidentportability.ErrPreparedStream
		}
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
		// Preparation derived the history identities; a mutation whose values
		// could not be described keeps that verdict at this rank.
		if staged.DescribeFailed {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		if !reflect.DeepEqual(actual, expected) || !slices.Equal(historyIDs, staged.HistoryRecordIDs) || !slices.Equal(entryIDs, staged.HistoryEntryRecordIDs) {
			return revisionsFailure(revisionsSourceStateInvariant)
		}
		return nil
	})
	return finishRevisionSourceRows(rows, err)
}

func (p preparedRevisionsImport) validateRevisionState(ctx context.Context, tx pgx.Tx, store incidentportability.MemberReader) error {
	rows, err := tx.Query(ctx, `
SELECT r.revision_id,r.change_set_id,r.record_id,r.row_version,r.before_json,r.after_json,r.created_at
FROM record_revisions r JOIN change_sets c USING(change_set_id)
WHERE c.incident_id=$1 ORDER BY r.revision_id
`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = incidentportability.ReadPreparedRows(ctx, store, p.revisions, revisionRowCodec{}, func(staged stagedRevision) error {
		expected := portableRecordRevision{
			RevisionID: staged.RevisionID, ChangeSetID: staged.ChangeSetID, RecordID: staged.RecordID,
			RowVersion: staged.RowVersion, CreatedAt: staged.CreatedAt,
		}
		var err error
		if expected.BeforeJSON, err = decodeDatabaseJSONObject(staged.BeforeJSON); err != nil {
			return incidentportability.ErrPreparedStream
		}
		if expected.AfterJSON, err = decodeDatabaseJSONObject(staged.AfterJSON); err != nil {
			return incidentportability.ErrPreparedStream
		}
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
