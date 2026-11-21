package revisions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func applyPreparedRevisionsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedRevisionsImport,
	importContext sourceport.ImportContext,
	validation *incidentBundleValidationCatalog,
) error {
	if importContext.ActorUserID == uuid.Nil ||
		importContext.IncidentID != prepared.incidentID || importContext.ActorUserID != prepared.runtimeActor || validation == nil {
		return revisionsFailure(revisionsReferencesInvariant)
	}
	store := importContext.Preparation
	if store == nil || prepared.verdicts == nil {
		return errors.New("revisions: incident bundle preparation is required")
	}
	if err := validatePreparedRevisionsBeforeWriteTx(ctx, tx, prepared, store, validation, importContext.Attributions); err != nil {
		return err
	}
	// Each family is one COPY of its staged write-ready rows; the database
	// acknowledgement must equal the staged count.
	changeSets, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.changeSets, changeSetRowCodec{actor: importContext.ActorUserID})
	if err != nil {
		return err
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"change_sets"}, changeSetColumns, changeSets)
	if err := changeSets.Finish(written, err); err != nil {
		return classifyRevisionsImportError(err)
	}
	mutations, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.mutations, mutationRowCodec{})
	if err != nil {
		return err
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"change_set_mutations"}, mutationColumns, mutations)
	if err := mutations.Finish(written, err); err != nil {
		return classifyRevisionsImportError(err)
	}
	revisionRows, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.revisions, revisionRowCodec{})
	if err != nil {
		return err
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"record_revisions"}, revisionColumns, revisionRows)
	if err := revisionRows.Finish(written, err); err != nil {
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
	case "change_sets_incident_history_order_key":
		return revisionsFailure(revisionsHistoryInvariant)
	case "record_revisions_pkey", "record_revisions_record_id_row_version_key":
		return revisionsFailure(revisionsRecordVersionInvariant)
	case "change_set_mutations_pkey":
		return revisionsFailure(revisionsSequenceInvariant)
	default:
		return revisionsFailure(revisionsReferencesInvariant)
	}
}
