package imports

import (
	"context"
	"encoding/json"
	"errors"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/records"
)

type committedSourceRecord struct {
	SchemaID          string    `json:"schema_id"`
	CommittedSourceID uuid.UUID `json:"committed_source_id"`
	SourceRowRef      int       `json:"source_row_ref"`
	RecordID          uuid.UUID `json:"record_id"`
}

func recordBindingFailure() error {
	return committedSourceDescriptor().DeclaredFailure("committed_sources.record_binding_exact")
}

func walkCommittedSourceRecords(ctx context.Context, members incidentportability.MemberReader, visit func(committedSourceRecord) error) error {
	return incidentportability.WalkPath(ctx, members, committedSourceRecordsPath, func(row map[string]any) error {
		if len(row) != 4 || row["schema_id"] != "cartulary.imports.committed_source_record.v1" {
			return recordBindingFailure()
		}
		sourceText, sourceOK := row["committed_source_id"].(string)
		recordText, recordOK := row["record_id"].(string)
		sourceID, sourceErr := uuid.Parse(sourceText)
		recordID, recordErr := uuid.Parse(recordText)
		rowNumber, rowOK := row["source_row_ref"].(json.Number)
		rowRef, rowErr := rowNumber.Int64()
		if !sourceOK || !recordOK || sourceErr != nil || recordErr != nil || sourceID == uuid.Nil || recordID == uuid.Nil || sourceID.String() != sourceText || recordID.String() != recordText || !rowOK || rowErr != nil || rowRef < 2 || rowRef > 2147483647 {
			return recordBindingFailure()
		}
		return visit(committedSourceRecord{SchemaID: "cartulary.imports.committed_source_record.v1", CommittedSourceID: sourceID, SourceRowRef: int(rowRef), RecordID: recordID})
	})
}

// applyRecordBindings checks every staged binding against its source's view
// scope and the applied record envelopes, one envelope read per batch, then
// writes the bindings with one COPY.
func (p *PreparedCommittedSources) applyRecordBindings(ctx context.Context, tx pgx.Tx, store incidentportability.MemberReader) error {
	envelopes := records.NewStore()
	batch := make([]committedSourceRecord, 0, committedSourceBindingBatch)
	check := func() error {
		if len(batch) == 0 {
			return nil
		}
		ids := make([]uuid.UUID, len(batch))
		for index, value := range batch {
			ids[index] = value.RecordID
		}
		loaded, err := envelopes.LoadEnvelopesTx(ctx, tx, ids, false)
		if err != nil {
			return err
		}
		for _, value := range batch {
			view, ok := p.views[value.CommittedSourceID]
			if !ok || !view.public || value.SourceRowRef < view.firstRow {
				return recordBindingFailure()
			}
			envelope, ok := loaded[value.RecordID]
			if !ok || envelope.IncidentID != p.incidentID || !slices.Contains(view.types, envelope.RecordType) {
				return recordBindingFailure()
			}
		}
		batch = batch[:0]
		return nil
	}
	if err := incidentportability.ReadPreparedRows(ctx, store, p.bindings, committedSourceBindingCodec{}, func(value committedSourceRecord) error {
		batch = append(batch, value)
		if len(batch) < committedSourceBindingBatch {
			return nil
		}
		return check()
	}); err != nil {
		return err
	}
	if err := check(); err != nil {
		return err
	}
	source, err := incidentportability.OpenPreparedCopy(ctx, store, p.bindings, committedSourceBindingCodec{incidentID: p.incidentID})
	if err != nil {
		return err
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"import_committed_source_records"}, committedSourceBindingColumns, source)
	return source.Finish(written, err)
}

// validateRecordBindings compares the stored bindings with the staged
// bindings in one ordered pass.
func (p *PreparedCommittedSources) validateRecordBindings(ctx context.Context, tx pgx.Tx, store incidentportability.MemberReader) error {
	var count int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM import_committed_source_records WHERE incident_id=$1`, p.incidentID).Scan(&count); err != nil {
		return err
	}
	if count != p.recordCount {
		return recordBindingFailure()
	}
	if !p.staged || store == nil {
		return errors.New("committed sources: incident bundle preparation is required")
	}
	rows, err := tx.Query(ctx, `SELECT committed_source_id,source_row_ref,record_id FROM import_committed_source_records WHERE incident_id=$1 ORDER BY committed_source_id,source_row_ref`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = incidentportability.ReadPreparedRows(ctx, store, p.bindings, committedSourceBindingCodec{}, func(value committedSourceRecord) error {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return recordBindingFailure()
		}
		var sourceID, recordID uuid.UUID
		var rowRef int
		if err := rows.Scan(&sourceID, &rowRef, &recordID); err != nil {
			return err
		}
		if sourceID != value.CommittedSourceID || rowRef != value.SourceRowRef || recordID != value.RecordID {
			return recordBindingFailure()
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rows.Next() {
		return recordBindingFailure()
	}
	return rows.Err()
}
