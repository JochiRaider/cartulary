package imports

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Approval stages rows inside its existing transaction. A failed scan leaves
// neither rows nor a seal; mapping publication commits the complete projection.
func (s *store) stageGenericRowsTx(ctx context.Context, tx pgx.Tx, sessionID, unitID uuid.UUID, expected discoveredUnit) (string, error) {
	capability, err := sourceCapabilityForUnitTx(ctx, tx, sessionID, unitID)
	if err != nil {
		return "", err
	}
	port := extensionSourcePort{db: s.db}
	stream, err := port.OpenSourceStreamTx(ctx, tx, capability.SourceStreamRef)
	if err != nil {
		return "", err
	}
	defer stream.Reader.Close()
	if _, err := tx.Exec(ctx, `DELETE FROM import_unit_projection_rows WHERE import_unit_id=$1`, unitID); err != nil {
		return "", err
	}
	digest := sha256.New()
	actual, err := scanGenericCSV(ctx, stream.Reader, func(row map[string]any) error {
		raw, err := json.Marshal(row)
		if err != nil {
			return err
		}
		ref, _ := intFromAny(row["source_row_ref"])
		if _, err := tx.Exec(ctx, `INSERT INTO import_unit_projection_rows(import_unit_id,source_row_ref,row_json) VALUES($1,$2,$3)`, unitID, ref, raw); err != nil {
			return err
		}
		_, _ = digest.Write(raw)
		_, _ = digest.Write([]byte{'\n'})
		return nil
	})
	if err != nil {
		return "", err
	}
	if actual.SourceRectA1 != expected.SourceRectA1 || actual.InferredRowCount != expected.InferredRowCount || !reflect.DeepEqual(actual.Columns, expected.Columns) {
		return "", importApplyBlockedError("source_changed")
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

type genericRowIterator struct {
	ctx          context.Context
	tx           pgx.Tx
	unit         applyUnitData
	index, after int
	row          map[string]any
	digest       hash.Hash
	err          error
	done         bool
}

func newGenericRowIterator(ctx context.Context, tx pgx.Tx, unit applyUnitData) *genericRowIterator {
	return &genericRowIterator{ctx: ctx, tx: tx, unit: unit, digest: sha256.New()}
}

func (r *genericRowIterator) Next() bool {
	if r.done || r.err != nil {
		return false
	}
	if r.err = r.ctx.Err(); r.err != nil {
		return false
	}
	if r.unit.LocatorKind != "csv_file" {
		if r.index == len(r.unit.SourceRows) {
			r.done = true
			return false
		}
		r.row = r.unit.SourceRows[r.index]
		r.index++
		return true
	}
	var raw []byte
	var ref int
	err := r.tx.QueryRow(r.ctx, `SELECT source_row_ref,row_json FROM import_unit_projection_rows WHERE import_unit_id=$1 AND source_row_ref>$2 ORDER BY source_row_ref LIMIT 1`, r.unit.UnitID, r.after).Scan(&ref, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		r.done = true
		if r.index != r.unit.SourceProjectionCount || hex.EncodeToString(r.digest.Sum(nil)) != r.unit.SourceProjectionDigest {
			r.err = importApplyBlockedError("source_changed")
		}
		return false
	}
	if err != nil {
		r.err = err
		return false
	}
	r.row = nil
	if err := json.Unmarshal(raw, &r.row); err != nil {
		r.err = err
		return false
	}
	rowRef, ok := intFromAny(r.row["source_row_ref"])
	if !ok || rowRef != ref || ref != r.index+2 {
		r.err = importApplyBlockedError("source_changed")
		return false
	}
	_, _ = r.digest.Write(raw)
	_, _ = r.digest.Write([]byte{'\n'})
	r.after = ref
	r.index++
	return true
}

func (r *genericRowIterator) Row() map[string]any { return r.row }
func (r *genericRowIterator) Err() error          { return r.err }
