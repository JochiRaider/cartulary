package timeline

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/internal/provenancevalue"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/internal/sourcevalue"
)

// Staged Timeline streams (Core 01 REQ-01-641). Preparation admits every
// bundle-only fact and stages the admitted rows; the final transaction checks
// only database-side facts and writes from these streams. Paths are relative to
// the port's preparation scope.
const (
	preparedTimelineEventsPath      = "timeline_events.rows"
	preparedTimelineProvenancePath  = "timeline_source_provenance.rows"
	preparedTimelineAttributionPath = "attribution.rows"
	// timelineSetBatch bounds one set-based statement's parameter arrays.
	timelineSetBatch = 512
)

// timelineEventCodec stages the admitted logical event exactly as preparation
// parsed it; JSON round-trips every field of the closed shape.
type timelineEventCodec struct{}

func (timelineEventCodec) Encode(row sourcevalue.Row) ([]byte, error) { return json.Marshal(row) }
func (timelineEventCodec) Decode(line []byte) (sourcevalue.Row, error) {
	var row sourcevalue.Row
	if err := json.Unmarshal(line, &row); err != nil {
		return sourcevalue.Row{}, incidentportability.ErrPreparedStream
	}
	row.NormalizeClocks()
	return row, nil
}
func (timelineEventCodec) Columns() []string { return nil }
func (timelineEventCodec) Values(sourcevalue.Row) ([]any, error) {
	return nil, incidentportability.ErrPreparedStream
}

var timelineProvenanceColumns = []string{
	"record_id", "source_identity_hash", "source_row_ordinal", "source_column_ordinal", "source_kind",
	"source_metadata", "source_header_json", "raw_value", "cell_kind", "created_at",
}

// timelineProvenanceCodec stages one admitted provenance row; its COPY values
// are the live import's insert values.
type timelineProvenanceCodec struct{}

func (timelineProvenanceCodec) Encode(row provenancevalue.Row) ([]byte, error) {
	return json.Marshal(row)
}
func (timelineProvenanceCodec) Decode(line []byte) (provenancevalue.Row, error) {
	var row provenancevalue.Row
	if err := json.Unmarshal(line, &row); err != nil {
		return provenancevalue.Row{}, incidentportability.ErrPreparedStream
	}
	row.CreatedAt = row.CreatedAt.UTC()
	return row, nil
}
func (timelineProvenanceCodec) Columns() []string { return timelineProvenanceColumns }
func (timelineProvenanceCodec) Values(row provenancevalue.Row) ([]any, error) {
	return []any{row.RecordID, row.SourceIdentity[:], row.RowOrdinal, row.ColumnOrdinal, row.Values.SourceKind,
		row.Values.MetadataJSON, row.Values.HeaderJSON, row.Values.RawValue, row.Values.CellKind, row.CreatedAt}, nil
}

// stage writes the admitted events, provenance and historical attribution.
// Every bundle fact was proven before this point; the transaction reads only
// these streams.
func (p *preparedTimelineImport) stage(ctx context.Context, store sourceport.PreparationStore, input sourceport.ImportContext, d sourceport.Descriptor) error {
	events, err := incidentportability.StagePreparedRows(ctx, store, preparedTimelineEventsPath, timelineEventCodec{}, func(emit func(sourcevalue.Row) error) error {
		return p.events(ctx, d, emit)
	})
	if err != nil {
		return err
	}
	provenance, err := incidentportability.StagePreparedRows(ctx, store, preparedTimelineProvenancePath, timelineProvenanceCodec{}, func(emit func(provenancevalue.Row) error) error {
		return p.provenance(ctx, d, emit)
	})
	if err != nil {
		return err
	}
	admitted := func(actorID string) bool { _, ok := input.Actors.Lookup(actorID); return ok }
	attribution, err := incidentportability.StageAttributions(ctx, store, preparedTimelineAttributionPath, admitted, func(emit func(incidentportability.AttributionRow) error) error {
		if p.profile != nil && p.profile.UpdatedByUserID != nil {
			if err := emit(incidentportability.AttributionRow{SourceTable: "timeline_time_conversion_profiles", SourceRowID: p.incidentID.String(), SourceColumn: "updated_by_user_id", SourceActorID: p.profile.UpdatedByUserID.String()}); err != nil {
				return err
			}
		}
		for _, column := range []string{"reviewed_by_user_id", "superseded_by_user_id"} {
			if err := incidentportability.ReadPreparedRows(ctx, store, events, timelineEventCodec{}, func(row sourcevalue.Row) error {
				actor := row.ReviewedByUserID
				if column == "superseded_by_user_id" {
					actor = row.SupersededByUserID
				}
				if actor == nil {
					return nil
				}
				return emit(incidentportability.AttributionRow{SourceTable: "timeline_events", SourceRowID: row.RecordID.String(), SourceColumn: column, SourceActorID: actor.String()})
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	p.staged, p.eventRows, p.provenanceRows, p.attribution = true, events, provenance, attribution
	return nil
}

// importedEventActor replaces an admitted historical actor with the importing
// actor, as the live import stores it.
func (p preparedTimelineImport) importedEventActor(actor *uuid.UUID) *uuid.UUID {
	if actor == nil {
		return nil
	}
	local := p.actorID
	return &local
}
