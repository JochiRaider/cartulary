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
	recordenvelope "github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
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

func (p *PreparedCommittedSources) applyRecordBindings(ctx context.Context, tx pgx.Tx) error {
	return walkCommittedSourceRecords(ctx, p.members, func(value committedSourceRecord) error {
		var viewID string
		var firstRow int
		err := tx.QueryRow(ctx, `SELECT target_view_schema_id,data_start_row_ref FROM import_committed_sources
WHERE incident_id=$1 AND committed_source_id=$2 AND target_kind=$3`, p.incidentID, value.CommittedSourceID, ImportTargetKindViewSchema).Scan(&viewID, &firstRow)
		if errors.Is(err, pgx.ErrNoRows) {
			return recordBindingFailure()
		}
		if err != nil {
			return err
		}
		resource, ok := viewschema.LookupPublicResource(viewID)
		if !ok || value.SourceRowRef < firstRow {
			return recordBindingFailure()
		}
		envelope, err := records.NewStore().LoadEnvelopeTx(ctx, tx, value.RecordID, false)
		if errors.Is(err, recordenvelope.ErrEnvelopeNotFound) {
			return recordBindingFailure()
		}
		if err != nil {
			return err
		}
		if envelope.IncidentID != p.incidentID || !slices.Contains(resource.SourceRecordTypes, envelope.RecordType) {
			return recordBindingFailure()
		}
		_, err = tx.Exec(ctx, `INSERT INTO import_committed_source_records(incident_id,committed_source_id,source_row_ref,record_id) VALUES($1,$2,$3,$4)`, p.incidentID, value.CommittedSourceID, value.SourceRowRef, value.RecordID)
		return err
	})
}

func (p *PreparedCommittedSources) validateRecordBindings(ctx context.Context, tx pgx.Tx) error {
	var count int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM import_committed_source_records WHERE incident_id=$1`, p.incidentID).Scan(&count); err != nil {
		return err
	}
	if count != p.recordCount {
		return recordBindingFailure()
	}
	return walkCommittedSourceRecords(ctx, p.members, func(value committedSourceRecord) error {
		var recordID uuid.UUID
		err := tx.QueryRow(ctx, `SELECT record_id FROM import_committed_source_records WHERE incident_id=$1 AND committed_source_id=$2 AND source_row_ref=$3`, p.incidentID, value.CommittedSourceID, value.SourceRowRef).Scan(&recordID)
		if errors.Is(err, pgx.ErrNoRows) || err == nil && recordID != value.RecordID {
			return recordBindingFailure()
		}
		return err
	})
}
