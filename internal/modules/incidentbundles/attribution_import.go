package incidentbundles

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"strings"
)

// Attribution is authoritative incident state and is written within the same
// publication transaction as its source rows. Preparation retains no row buffer.
type transactionAttributions struct {
	ctx                     context.Context
	tx                      pgx.Tx
	incidentID, localUserID uuid.UUID
	actors                  sourceport.ActorCatalog
}

func (r *transactionAttributions) RecordImportedAttribution(table, rowID, column, actorID string) error {
	if r == nil || r.tx == nil || r.incidentID == uuid.Nil || r.localUserID == uuid.Nil || strings.TrimSpace(table) == "" || strings.TrimSpace(rowID) == "" || strings.TrimSpace(column) == "" {
		return errors.New("incident bundle attribution is invalid")
	}
	if _, admitted := r.actors.Lookup(actorID); !admitted {
		return actorReferenceFailure()
	}
	tag, err := r.tx.Exec(r.ctx, `
 INSERT INTO incident_bundle_imported_attributions AS existing
 (incident_id,source_table,source_row_id,source_column,source_actor_id,local_user_id)
 VALUES ($1,$2,$3,$4,$5,$6)
 ON CONFLICT (incident_id,source_table,source_row_id,source_column)
 DO UPDATE SET source_actor_id=EXCLUDED.source_actor_id
 WHERE existing.source_actor_id=EXCLUDED.source_actor_id AND existing.local_user_id=EXCLUDED.local_user_id
 `, r.incidentID, table, rowID, column, actorID, r.localUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errors.New("incident bundle attribution conflicts")
	}
	return nil
}
func (r *transactionAttributions) VisitImportedAttributions(table, rowID string, visit func(incidentportability.ImportedAttribution) error) error {
	if r == nil || r.tx == nil || strings.TrimSpace(table) == "" || visit == nil {
		return errors.New("incident bundle attribution query is invalid")
	}
	rows, err := r.tx.Query(r.ctx, `
 SELECT source_table,source_row_id,source_column,source_actor_id,local_user_id
 FROM incident_bundle_imported_attributions
 WHERE incident_id=$1 AND source_table=$2 AND ($3='' OR source_row_id=$3)
 ORDER BY source_row_id,source_column
 `, r.incidentID, table, rowID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var row incidentportability.ImportedAttribution
		if err := rows.Scan(&row.SourceTable, &row.SourceRowID, &row.SourceColumn, &row.SourceActorID, &row.LocalUserID); err != nil {
			return err
		}
		if err := visit(row); err != nil {
			return err
		}
	}
	return rows.Err()
}

// NewImportedAttributionRecorder binds a participant's historical attribution to
// the same final transaction. Only application composition supplies the handle.
func NewImportedAttributionRecorder(ctx context.Context, tx pgx.Tx, incidentID, actorID uuid.UUID, actors sourceport.ActorCatalog) (incidentportability.AttributionRecorder, error) {
	if ctx == nil || tx == nil || incidentID == uuid.Nil || actorID == uuid.Nil {
		return nil, ErrPortabilityUnavailable
	}
	return &transactionAttributions{ctx: ctx, tx: tx, incidentID: incidentID, localUserID: actorID, actors: actors}, nil
}
