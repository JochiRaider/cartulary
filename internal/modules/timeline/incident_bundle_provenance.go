package timeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/internal/provenancevalue"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/internal/sourcevalue"
)

func (p preparedTimelineImport) provenance(ctx context.Context, d sourceport.Descriptor, visit func(provenancevalue.Row) error) error {
	return incidentportability.WalkPath(ctx, p.members, timelineBundleProvenancePath, func(raw map[string]any) error {
		row, ok := provenancevalue.Parse(raw)
		if !ok {
			return d.DeclaredFailure("timeline.version_shape_exact")
		}
		return visit(row)
	})
}

func (p preparedTimelineImport) prepareProvenance(ctx context.Context, d sourceport.Descriptor) error {
	var previousID, previousDigest string
	var previousRow, previousColumn int64
	if err := incidentportability.WalkPath(ctx, p.members, timelineBundleProvenancePath, func(raw map[string]any) error {
		id, row, column, digest, ok := provenancevalue.Identity(raw)
		if !ok {
			return d.DeclaredFailure("timeline.source_identity_admitted")
		}
		text := id.String()
		order := strings.Compare(text, previousID)
		if order == 0 {
			if row < previousRow {
				order = -1
			} else if row > previousRow {
				order = 1
			} else if column < previousColumn {
				order = -1
			} else if column > previousColumn {
				order = 1
			} else {
				order = strings.Compare(digest, previousDigest)
			}
		}
		if previousID != "" && order <= 0 {
			return d.DeclaredFailure("timeline.source_identity_admitted")
		}
		previousID, previousRow, previousColumn, previousDigest = text, row, column, digest
		return nil
	}); err != nil {
		return err
	}
	if err := p.provenance(ctx, d, func(provenancevalue.Row) error { return nil }); err != nil {
		return err
	}
	var record uuid.UUID
	count := 0
	if err := p.provenance(ctx, d, func(row provenancevalue.Row) error {
		if row.RecordID != record {
			record = row.RecordID
			count = 0
		}
		count++
		if !row.Coherent() || count > provenancevalue.MaxEntriesPerRecord {
			return d.DeclaredFailure("timeline.provenance_coherent")
		}
		return nil
	}); err != nil {
		return err
	}
	return p.admitProvenanceEventBindings(ctx, d)
}

// Both immutable members have already passed strict syntax, shape and sorted
// identity admission. Merge their identities with one decoded event at a time,
// draining the verified event stream through EOF before preparation succeeds.
func (p preparedTimelineImport) admitProvenanceEventBindings(ctx context.Context, d sourceport.Descriptor) (err error) {
	stream, _, err := p.members.OpenMember(ctx, timelineBundleRecordsPath)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, stream.Close()) }()
	decoder := json.NewDecoder(stream)
	decoder.UseNumber()
	var current string
	exhausted := false
	next := func() error {
		var raw map[string]any
		if err := decoder.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				exhausted = true
				current = ""
				return nil
			}
			return err
		}
		row, ok := sourcevalue.Parse(raw)
		if !ok {
			return d.DeclaredFailure("timeline.version_shape_exact")
		}
		current = row.RecordID.String()
		return ctx.Err()
	}
	if err = next(); err != nil {
		return err
	}
	if err = p.provenance(ctx, d, func(row provenancevalue.Row) error {
		for !exhausted && current < row.RecordID.String() {
			if err := next(); err != nil {
				return err
			}
		}
		if exhausted || current != row.RecordID.String() {
			return d.DeclaredFailure("timeline.provenance_non_orphaned")
		}
		return nil
	}); err != nil {
		return err
	}
	for !exhausted {
		if err = next(); err != nil {
			return err
		}
	}
	return nil
}

// stagedProvenanceBatches visits the staged provenance rows in admitted
// identity order, at most timelineSetBatch at a time.
func (p preparedTimelineImport) stagedProvenanceBatches(ctx context.Context, store incidentportability.MemberReader, visit func([]provenancevalue.Row) error) error {
	batch := make([]provenancevalue.Row, 0, timelineSetBatch)
	if err := incidentportability.ReadPreparedRows(ctx, store, p.provenanceRows, timelineProvenanceCodec{}, func(row provenancevalue.Row) error {
		batch = append(batch, row)
		if len(batch) < timelineSetBatch {
			return nil
		}
		err := visit(batch)
		batch = batch[:0]
		return err
	}); err != nil {
		return err
	}
	if len(batch) == 0 {
		return nil
	}
	return visit(batch)
}

// provenanceBindings checks each staged row's event binding and, before
// writes, its identity occupancy with one statement per batch. The first
// failing row in identity order selects the invariant.
func (p preparedTimelineImport) provenanceBindings(ctx context.Context, tx pgx.Tx, store incidentportability.MemberReader, d sourceport.Descriptor, collisions bool) error {
	return p.stagedProvenanceBatches(ctx, store, func(batch []provenancevalue.Row) error {
		ids, identities := make([]uuid.UUID, len(batch)), make([][]byte, len(batch))
		rowOrdinals, columnOrdinals := make([]int64, len(batch)), make([]int64, len(batch))
		for index, row := range batch {
			ids[index], identities[index] = row.RecordID, row.SourceIdentity[:]
			rowOrdinals[index], columnOrdinals[index] = row.RowOrdinal, row.ColumnOrdinal
		}
		rows, err := tx.Query(ctx, `
SELECT EXISTS(SELECT 1 FROM timeline_events t JOIN records r ON r.record_id=t.record_id WHERE t.record_id=u.id AND t.incident_id=$5 AND r.incident_id=$5 AND r.record_type='timeline_event'),
       EXISTS(SELECT 1 FROM timeline_source_provenance WHERE record_id=u.id AND source_identity_hash=u.identity AND source_row_ordinal=u.row_ordinal AND source_column_ordinal=u.column_ordinal)
  FROM unnest($1::uuid[],$2::bytea[],$3::bigint[],$4::bigint[]) WITH ORDINALITY AS u(id,identity,row_ordinal,column_ordinal,ordinal)
 ORDER BY u.ordinal`, ids, identities, rowOrdinals, columnOrdinals, p.incidentID)
		if err != nil {
			return err
		}
		checks, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) ([2]bool, error) {
			var check [2]bool
			return check, row.Scan(&check[0], &check[1])
		})
		if err != nil {
			return err
		}
		if len(checks) != len(batch) {
			return incidentportability.ErrPreparedStream
		}
		for _, check := range checks {
			if !check[0] {
				return d.DeclaredFailure("timeline.provenance_non_orphaned")
			}
			if collisions && check[1] {
				return d.DeclaredFailure("timeline.provenance_unique")
			}
		}
		return nil
	})
}

func (p preparedTimelineImport) applyProvenance(ctx context.Context, tx pgx.Tx, store incidentportability.MemberReader, d sourceport.Descriptor) error {
	if err := p.provenanceBindings(ctx, tx, store, d, true); err != nil {
		return err
	}
	source, err := incidentportability.OpenPreparedCopy(ctx, store, p.provenanceRows, timelineProvenanceCodec{})
	if err != nil {
		return err
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"timeline_source_provenance"}, timelineProvenanceColumns, source)
	return source.Finish(written, err)
}

// validateProvenance compares the stored provenance with the staged rows in
// one ordered pass. A missing, changed or extra row is an exact-state failure.
func (p preparedTimelineImport) validateProvenance(ctx context.Context, tx pgx.Tx, store incidentportability.MemberReader, d sourceport.Descriptor) error {
	if err := p.provenanceBindings(ctx, tx, store, d, false); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT jsonb_build_object('record_id',p.record_id,'source_identity_sha256',encode(p.source_identity_hash,'hex'),'source_row_ordinal',p.source_row_ordinal,'source_column_ordinal',p.source_column_ordinal,'source_kind',p.source_kind,'source_metadata',p.source_metadata,'source_header',p.source_header_json,'raw_value',p.raw_value,'cell_kind',p.cell_kind,'created_at',p.created_at)
FROM timeline_source_provenance p JOIN timeline_events t ON t.record_id=p.record_id
WHERE t.incident_id=$1
ORDER BY p.record_id, p.source_row_ordinal, p.source_column_ordinal, p.source_identity_hash`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = incidentportability.ReadPreparedRows(ctx, store, p.provenanceRows, timelineProvenanceCodec{}, func(expected provenancevalue.Row) error {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return d.DeclaredFailure("timeline.provenance_state_exact")
		}
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return err
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var object map[string]any
		if err := decoder.Decode(&object); err != nil {
			return err
		}
		actual, ok := provenancevalue.Parse(object)
		if !ok || !reflect.DeepEqual(actual, expected) {
			return d.DeclaredFailure("timeline.provenance_state_exact")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rows.Next() {
		return d.DeclaredFailure("timeline.provenance_state_exact")
	}
	return rows.Err()
}
