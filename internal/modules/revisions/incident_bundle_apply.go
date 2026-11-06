package revisions

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func applyPreparedRevisionsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	importContext sourceport.ImportContext,
	validation *incidentBundleValidationCatalog,
) error {
	if importContext.Attributions == nil || importContext.ActorUserID == uuid.Nil ||
		importContext.IncidentID != prepared.incidentID || importContext.ActorUserID != prepared.runtimeActor || validation == nil {
		return revisionsFailure(revisionsReferencesInvariant)
	}
	if err := validatePreparedRevisionsBeforeWriteTx(ctx, tx, prepared, validation); err != nil {
		return err
	}
	if err := prepared.walkChangeSets(ctx, func(row portableChangeSet) error {
		tag, err := tx.Exec(ctx, `
INSERT INTO change_sets (
    change_set_id, incident_id, actor_user_id, source, reason,
    client_txn_id, request_id, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, row.ChangeSetID, row.IncidentID, importContext.ActorUserID, row.Source,
			row.Reason, row.ClientTxnID, row.RequestID, row.CreatedAt)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return revisionsFailure(revisionsReferencesInvariant)
		}
		if err := importContext.Attributions.RecordImportedAttribution(
			"change_sets",
			row.ChangeSetID.String(),
			"actor_user_id",
			row.PortableActorID.String(),
		); err != nil {
			return revisionsFailure(revisionsActorsInvariant)
		}
		return nil
	}); err != nil {
		return classifyRevisionsImportError(err)
	}

	if err := prepared.walkMutations(ctx, func(row portableChangeSetMutation) error {
		if err := validateImportedChangeSetTx(ctx, tx, prepared.incidentID, row.ChangeSetID); err != nil {
			return err
		}
		history, err := validation.targets.DescribeValues(row.TargetKind, row.TargetID, row.OperationKind, row.BeforeValue, row.AfterValue)
		if err != nil {
			return revisionsFailure(revisionsHistoryInvariant)
		}
		tag, err := tx.Exec(ctx, `
INSERT INTO change_set_mutations (
    change_set_id, sequence_no, target_kind, target_id, operation_kind,
    before_version_id, after_version_id, before_value, after_value,
    history_record_ids, history_entry_record_ids
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`, row.ChangeSetID, row.SequenceNo, row.TargetKind, row.TargetID,
			row.OperationKind, row.BeforeVersionID, row.AfterVersionID,
			jsonOrNil(row.BeforeValue), jsonOrNil(row.AfterValue),
			history.HistoryRecordIDs, history.HistoryEntryRecordIDs)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return revisionsFailure(revisionsSequenceInvariant)
		}
		return nil
	}); err != nil {
		return classifyRevisionsImportError(err)
	}
	if err := prepared.walkRevisions(ctx, func(row portableRecordRevision) error {
		if err := validateImportedChangeSetTx(ctx, tx, prepared.incidentID, row.ChangeSetID); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `
INSERT INTO record_revisions (
    revision_id, change_set_id, record_id, row_version,
    before_json, after_json, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
`, row.RevisionID, row.ChangeSetID, row.RecordID, row.RowVersion,
			jsonOrNil(row.BeforeJSON), jsonOrNil(row.AfterJSON), row.CreatedAt)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return revisionsFailure(revisionsRecordVersionInvariant)
		}
		return nil
	}); err != nil {
		return classifyRevisionsImportError(err)
	}
	return nil
}

func classifyRevisionsImportError(err error) error {
	var failure *pgconn.PgError
	if !errors.As(err, &failure) {
		return err
	}
	switch failure.ConstraintName {
	case "record_revisions_pkey", "record_revisions_record_id_row_version_key":
		return revisionsFailure(revisionsRecordVersionInvariant)
	case "change_set_mutations_pkey":
		return revisionsFailure(revisionsSequenceInvariant)
	default:
		return revisionsFailure(revisionsReferencesInvariant)
	}
}
