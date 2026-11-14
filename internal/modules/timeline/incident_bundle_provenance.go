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

func (p preparedTimelineImport) provenanceBindings(ctx context.Context, tx pgx.Tx, d sourceport.Descriptor, collisions bool) error {
	return p.provenance(ctx, d, func(row provenancevalue.Row) error {
		var bound, occupied bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM timeline_events t JOIN records r ON r.record_id=t.record_id WHERE t.record_id=$1 AND t.incident_id=$2 AND r.incident_id=$2 AND r.record_type='timeline_event'),EXISTS(SELECT 1 FROM timeline_source_provenance WHERE record_id=$1 AND source_identity_hash=$3 AND source_row_ordinal=$4 AND source_column_ordinal=$5)`, row.RecordID, p.incidentID, row.SourceIdentity[:], row.RowOrdinal, row.ColumnOrdinal).Scan(&bound, &occupied); err != nil {
			return err
		}
		if !bound {
			return d.DeclaredFailure("timeline.provenance_non_orphaned")
		}
		if collisions && occupied {
			return d.DeclaredFailure("timeline.provenance_unique")
		}
		return nil
	})
}

func (p preparedTimelineImport) applyProvenance(ctx context.Context, tx pgx.Tx, d sourceport.Descriptor) error {
	if err := p.provenanceBindings(ctx, tx, d, true); err != nil {
		return err
	}
	return p.provenance(ctx, d, func(row provenancevalue.Row) error {
		tag, err := tx.Exec(ctx, `INSERT INTO timeline_source_provenance(record_id,source_identity_hash,source_row_ordinal,source_column_ordinal,source_kind,source_metadata,source_header_json,raw_value,cell_kind,created_at) VALUES($1,$2,$3,$4,$5,$6::jsonb,$7::jsonb,$8,$9,$10)`, row.RecordID, row.SourceIdentity[:], row.RowOrdinal, row.ColumnOrdinal, row.Values.SourceKind, row.Values.MetadataJSON, row.Values.HeaderJSON, row.Values.RawValue, row.Values.CellKind, row.CreatedAt)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return d.DeclaredFailure("timeline.provenance_unique")
		}
		return nil
	})
}

func (p preparedTimelineImport) validateProvenance(ctx context.Context, tx pgx.Tx, d sourceport.Descriptor) error {
	if err := p.provenanceBindings(ctx, tx, d, false); err != nil {
		return err
	}
	var count int64
	if err := p.provenance(ctx, d, func(expected provenancevalue.Row) error {
		count++
		var raw []byte
		err := tx.QueryRow(ctx, `SELECT jsonb_build_object('record_id',record_id,'source_identity_sha256',encode(source_identity_hash,'hex'),'source_row_ordinal',source_row_ordinal,'source_column_ordinal',source_column_ordinal,'source_kind',source_kind,'source_metadata',source_metadata,'source_header',source_header_json,'raw_value',raw_value,'cell_kind',cell_kind,'created_at',created_at) FROM timeline_source_provenance WHERE record_id=$1 AND source_identity_hash=$2 AND source_row_ordinal=$3 AND source_column_ordinal=$4`, expected.RecordID, expected.SourceIdentity[:], expected.RowOrdinal, expected.ColumnOrdinal).Scan(&raw)
		if errors.Is(err, pgx.ErrNoRows) {
			return d.DeclaredFailure("timeline.provenance_state_exact")
		}
		if err != nil {
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
	}); err != nil {
		return err
	}
	var actual int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM timeline_source_provenance p JOIN timeline_events t ON t.record_id=p.record_id WHERE t.incident_id=$1`, p.incidentID).Scan(&actual); err != nil {
		return err
	}
	if count != actual {
		return d.DeclaredFailure("timeline.provenance_state_exact")
	}
	return nil
}
