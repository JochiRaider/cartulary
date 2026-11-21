package incidentportability

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
)

// ErrAttributionStream reports an attribution row that is not admissible for
// staging or an attribution stream that does not match its preparation.
var ErrAttributionStream = errors.New("incident portability attribution stream is invalid")

// AttributionRow is one precomputed historical-actor attribution of an
// imported source row. The importing actor and incident are bound by the
// attribution owner when the stream is flushed.
type AttributionRow struct {
	SourceTable   string
	SourceRowID   string
	SourceColumn  string
	SourceActorID string
}

// AttributionStream is a sealed preparation member of grouped attribution
// lines with its expected per-source-table aggregates (Core 01 REQ-01-641).
type AttributionStream struct {
	PreparedRowStream
	// Tables holds the expected attribution count per source table. It is
	// bounded by the owner's declared relations, never by row count.
	Tables map[string]int64
}

// AttributionSink is implemented only by the attribution owner. It flushes
// precomputed streams with one COPY and reads set-based aggregates.
type AttributionSink interface {
	AttributionRecorder
	CountImportedAttributions(ctx context.Context, tables []string) (map[string]int64, error)
}

// attributionGroupRows bounds one grouped line: consecutive rows sharing a
// source table, column and historical actor share one line.
const attributionGroupRows = 256

// StageAttributions writes the attribution rows emitted by produce as one
// preparation member. Every row is checked for complete identity and for
// actor admission (admitted) once, at staging. Duplicate identities are the
// owner's stable-identity concern; the attribution owner's unique key remains
// the database backstop.
func StageAttributions(ctx context.Context, destination MemberWriter, path string, admitted func(actorID string) bool, produce func(emit func(AttributionRow) error) error) (AttributionStream, error) {
	if destination == nil || admitted == nil || produce == nil {
		return AttributionStream{}, ErrAttributionStream
	}
	tables := map[string]int64{}
	var rows int64
	var group struct {
		table, column, actor string
		ids                  []string
	}
	member, err := writeStagedMemberFrom(ctx, destination, path, func(ctx context.Context, out io.Writer) error {
		flush := func() error {
			if len(group.ids) == 0 {
				return nil
			}
			line, err := json.Marshal([]any{group.table, group.column, group.actor, group.ids})
			if err != nil {
				return err
			}
			if len(line) > MaxNDJSONLogicalLineBytes {
				return ErrAttributionStream
			}
			if _, err := out.Write(append(line, '\n')); err != nil {
				return err
			}
			group.ids = group.ids[:0]
			return nil
		}
		err := produce(func(row AttributionRow) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if strings.TrimSpace(row.SourceTable) == "" || strings.TrimSpace(row.SourceRowID) == "" ||
				strings.TrimSpace(row.SourceColumn) == "" || !admitted(row.SourceActorID) {
				return ErrAttributionStream
			}
			if len(group.ids) == attributionGroupRows || row.SourceTable != group.table || row.SourceColumn != group.column || row.SourceActorID != group.actor {
				if err := flush(); err != nil {
					return err
				}
				group.table, group.column, group.actor = row.SourceTable, row.SourceColumn, row.SourceActorID
			}
			group.ids = append(group.ids, row.SourceRowID)
			tables[row.SourceTable]++
			rows++
			return nil
		})
		if err != nil {
			return err
		}
		return flush()
	})
	if err != nil {
		return AttributionStream{}, err
	}
	return AttributionStream{PreparedRowStream: PreparedRowStream{Member: member, Rows: rows}, Tables: tables}, nil
}

// AttributionTables returns the union of source tables of streams, sorted.
func AttributionTables(streams ...AttributionStream) []string {
	seen := map[string]struct{}{}
	for _, stream := range streams {
		for table := range stream.Tables {
			seen[table] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for table := range seen {
		result = append(result, table)
	}
	sort.Strings(result)
	return result
}

// AttributionReader expands sealed attribution streams in order. It holds one
// grouped line at a time and verifies each stream's seal and row count.
type AttributionReader struct {
	ctx     context.Context
	source  MemberReader
	streams []AttributionStream
	current *preparedRowReader[[]AttributionRow]
	pending []AttributionRow
	row     AttributionRow
	index   int
	count   int64
	err     error
}

// OpenAttributions starts reading streams from source, the operation's
// preparation store.
func OpenAttributions(ctx context.Context, source MemberReader, streams ...AttributionStream) *AttributionReader {
	return &AttributionReader{ctx: ctx, source: source, streams: streams}
}

func (r *AttributionReader) Next() bool {
	for r.err == nil {
		if len(r.pending) > 0 {
			r.row, r.pending = r.pending[0], r.pending[1:]
			r.count++
			return true
		}
		if r.current != nil {
			if r.current.next() {
				r.pending = r.current.row
				continue
			}
			r.err = errors.Join(r.current.finish(), r.current.close())
			r.current = nil
			r.index++
			continue
		}
		if r.index >= len(r.streams) {
			return false
		}
		stream := r.streams[r.index]
		if stream.Rows == 0 {
			if err := requireEmptyStream(stream.PreparedRowStream); err != nil {
				r.err = err
			}
			r.index++
			continue
		}
		reader, err := openPreparedRows[[]AttributionRow](r.ctx, r.source, stream.PreparedRowStream, groupedAttributionCodec{})
		if err != nil {
			r.err = err
			return false
		}
		reader.weight = func(rows []AttributionRow) int64 { return int64(len(rows)) }
		r.current = reader
	}
	return false
}

func (r *AttributionReader) Row() AttributionRow { return r.row }
func (r *AttributionReader) Count() int64        { return r.count }
func (r *AttributionReader) Err() error          { return r.err }

func (r *AttributionReader) Close() error {
	if r.current == nil {
		return nil
	}
	err := r.current.close()
	r.current = nil
	return err
}

// Expected is the total row count the preparation recorded for streams.
func (r *AttributionReader) Expected() int64 {
	var total int64
	for _, stream := range r.streams {
		total += stream.Rows
	}
	return total
}

type groupedAttributionCodec struct{}

func (groupedAttributionCodec) Encode([]AttributionRow) ([]byte, error) {
	return nil, ErrAttributionStream
}
func (groupedAttributionCodec) Columns() []string { return nil }
func (groupedAttributionCodec) Values([]AttributionRow) ([]any, error) {
	return nil, ErrAttributionStream
}
func (groupedAttributionCodec) Decode(line []byte) ([]AttributionRow, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil || len(raw) != 4 {
		return nil, ErrAttributionStream
	}
	var table, column, actor string
	var ids []string
	if json.Unmarshal(raw[0], &table) != nil || json.Unmarshal(raw[1], &column) != nil ||
		json.Unmarshal(raw[2], &actor) != nil || json.Unmarshal(raw[3], &ids) != nil || len(ids) == 0 {
		return nil, ErrAttributionStream
	}
	rows := make([]AttributionRow, len(ids))
	for index, id := range ids {
		rows[index] = AttributionRow{SourceTable: table, SourceRowID: id, SourceColumn: column, SourceActorID: actor}
	}
	return rows, nil
}
