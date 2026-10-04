package reporting

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/artifactport"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ValidateRetainedArtifacts checks restored source evidence before readiness.
// It never requires source freshness at restore time, renders content, adopts
// source trust or resolves the destination's active selection.
func ValidateRetainedArtifacts(ctx context.Context, pool *pgxpool.Pool, references interface {
	ExportTx(context.Context, pgx.Tx, uuid.UUID, []reference_data.SetBinding) ([]byte, error)
}) error {
	if pool == nil || references == nil {
		return errors.New("reporting: incomplete artifact recovery dependencies")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	last := uuid.Nil
	for {
		var incident uuid.UUID
		err := tx.QueryRow(ctx, `SELECT incident_id FROM reporting_imported_artifact_files WHERE incident_id>$1 GROUP BY incident_id ORDER BY incident_id LIMIT 1`, last).Scan(&incident)
		if errors.Is(err, pgx.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		last = incident
		var operations int
		if err := tx.QueryRow(ctx, `SELECT count(DISTINCT operation_id) FROM reporting_imported_artifact_files WHERE incident_id=$1`, incident).Scan(&operations); err != nil {
			return err
		}
		if operations != 1 {
			return artifactport.ErrInvalid
		}
		files, err := loadImportedArtifactFiles(ctx, tx, incident)
		if err != nil {
			return err
		}
		refs, err := references.ExportTx(ctx, tx, incident, nil)
		if err != nil {
			return err
		}
		_, err = (PortableReferenceSource{}).PrepareArtifactImport(ctx, artifactport.ImportRequest{IncidentID: incident, OperationID: incident, References: refs, Bundle: artifactFiles(files)})
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
