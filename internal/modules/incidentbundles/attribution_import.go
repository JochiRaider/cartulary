package incidentbundles

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Attribution is authoritative incident state and is written within the same
// publication transaction as its source rows. Preparation retains no row buffer.
type transactionAttributions struct {
	tx                      pgx.Tx
	incidentID, localUserID uuid.UUID
}

// FlushImportedAttributions writes precomputed attribution streams with one
// COPY into the attribution owner's relation (Core 01 REQ-01-641 step 9). The
// rows were admitted at staging; the database must acknowledge exactly the
// staged count, and the unique identity key remains the backstop.
func (r *transactionAttributions) FlushImportedAttributions(ctx context.Context, source incidentportability.MemberReader, streams ...incidentportability.AttributionStream) error {
	if r == nil || r.tx == nil || r.incidentID == uuid.Nil || r.localUserID == uuid.Nil || source == nil {
		return errors.New("incident bundle attribution flush is invalid")
	}
	reader := incidentportability.OpenAttributions(ctx, source, streams...)
	written, err := r.tx.CopyFrom(ctx, pgx.Identifier{"incident_bundle_imported_attributions"},
		[]string{"incident_id", "source_table", "source_row_id", "source_column", "source_actor_id", "local_user_id"},
		attributionCopySource{reader: reader, incidentID: r.incidentID, localUserID: r.localUserID})
	err = errors.Join(err, reader.Err(), reader.Close())
	if err != nil {
		return err
	}
	if written != reader.Expected() || reader.Count() != reader.Expected() {
		return errors.New("incident bundle attribution flush count differs from preparation")
	}
	return nil
}

// CountImportedAttributions is the set-based database side of attribution
// validation: one aggregate read per flush, never a per-row lookup.
func (r *transactionAttributions) CountImportedAttributions(ctx context.Context, tables []string) (map[string]int64, error) {
	if r == nil || r.tx == nil || r.incidentID == uuid.Nil {
		return nil, errors.New("incident bundle attribution query is invalid")
	}
	rows, err := r.tx.Query(ctx, `
 SELECT source_table, count(*)
 FROM incident_bundle_imported_attributions
 WHERE incident_id=$1 AND source_table = ANY($2::text[])
 GROUP BY source_table
 ORDER BY source_table
 `, r.incidentID, tables)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := make(map[string]int64, len(tables))
	for rows.Next() {
		var table string
		var count int64
		if err := rows.Scan(&table, &count); err != nil {
			return nil, err
		}
		counts[table] = count
	}
	return counts, rows.Err()
}

type attributionCopySource struct {
	reader                  *incidentportability.AttributionReader
	incidentID, localUserID uuid.UUID
}

func (s attributionCopySource) Next() bool { return s.reader.Next() }
func (s attributionCopySource) Values() ([]any, error) {
	row := s.reader.Row()
	return []any{s.incidentID, row.SourceTable, row.SourceRowID, row.SourceColumn, row.SourceActorID, s.localUserID}, nil
}
func (s attributionCopySource) Err() error { return s.reader.Err() }

// NewImportedAttributionRecorder binds a participant's historical attribution to
// the same final transaction. Only application composition supplies the handle.
// Actor admission happened when each participant staged its attribution.
func NewImportedAttributionRecorder(ctx context.Context, tx pgx.Tx, incidentID, actorID uuid.UUID) (incidentportability.AttributionRecorder, error) {
	if ctx == nil || tx == nil || incidentID == uuid.Nil || actorID == uuid.Nil {
		return nil, ErrPortabilityUnavailable
	}
	return &transactionAttributions{tx: tx, incidentID: incidentID, localUserID: actorID}, nil
}
