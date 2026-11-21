package incidentbundle

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// Staged Indicators streams (Core 01 REQ-01-641). Preparation admits and
// evaluates every row under the pack set it resolved; the final transaction
// reads these streams and checks only database-side facts. Paths are relative
// to the port's preparation scope.
const (
	preparedIndicatorsPath    = "indicators.rows"
	preparedObservationsPath  = "indicator_observations.rows"
	preparedIntervalsPath     = "indicator_state_intervals.rows"
	preparedIndicatorAttrPath = "attribution.rows"
	// indicatorSetBatch bounds one set-based statement's parameter arrays.
	indicatorSetBatch = 512
)

var (
	portableIndicatorColumns = []string{
		"record_id", "incident_id", "indicator_type", "value_kind", "display_value",
		"normalized_value", "dedupe_key", "defanged_value", "hash_algorithm", "hash_value", "stix_pattern",
	}
	portableObservationColumns = []string{
		"indicator_observation_id", "incident_id", "source_record_id", "source_field_key",
		"origin_kind", "origin_locator", "observed_text", "parsed_indicator_type",
		"normalized_candidate", "resolution_status", "resolved_indicator_record_id",
		"row_version", "created_by_user_id", "created_at", "resolved_by_user_id",
		"resolved_at", "resolution_method", "deleted_at", "deleted_by_user_id",
	}
	portableIntervalColumns = []string{
		"indicator_state_interval_id", "incident_id", "indicator_record_id",
		"lifecycle_state", "valid_from", "valid_to", "confidence", "rationale", "support_refs",
		"assessor", "assessed_at", "row_version", "created_by_user_id", "created_at",
		"deleted_at", "deleted_by_user_id",
	}
	indicatorValidationSourceColumns = []string{"id", "kind", "dedupe_key", "active"}
)

func decodeIndicatorRow(line []byte, target any) error {
	if err := json.Unmarshal(line, target); err != nil {
		return incidentportability.ErrPreparedStream
	}
	return nil
}

type indicatorRowCodec struct{}

func (indicatorRowCodec) Encode(row portableIndicatorRow) ([]byte, error) { return json.Marshal(row) }
func (indicatorRowCodec) Decode(line []byte) (portableIndicatorRow, error) {
	var row portableIndicatorRow
	return row, decodeIndicatorRow(line, &row)
}
func (indicatorRowCodec) Columns() []string { return portableIndicatorColumns }
func (indicatorRowCodec) Values(row portableIndicatorRow) ([]any, error) {
	return []any{row.RecordID, row.IncidentID, row.IndicatorType, row.ValueKind,
		row.DisplayValue, row.NormalizedValue, row.DedupeKey, row.DefangedValue,
		row.HashAlgorithm, row.HashValue, row.STIXPattern}, nil
}

// indicatorSourceCodec projects staged indicators into the transaction-local
// validation index.
type indicatorSourceCodec struct{ indicatorRowCodec }

func (indicatorSourceCodec) Columns() []string { return indicatorValidationSourceColumns }
func (indicatorSourceCodec) Values(row portableIndicatorRow) ([]any, error) {
	return []any{row.RecordID, row.IndicatorType, row.DedupeKey, row.DeletedAt == nil}, nil
}

type observationRowCodec struct{}

func (observationRowCodec) Encode(row portableObservationRow) ([]byte, error) {
	return json.Marshal(row)
}
func (observationRowCodec) Decode(line []byte) (portableObservationRow, error) {
	var row portableObservationRow
	return row, decodeIndicatorRow(line, &row)
}
func (observationRowCodec) Columns() []string { return portableObservationColumns }
func (observationRowCodec) Values(row portableObservationRow) ([]any, error) {
	return []any{row.ObservationID, row.IncidentID, row.SourceRecordID, row.SourceFieldKey,
		row.OriginKind.String(), row.OriginLocator, row.ObservedText,
		row.ParsedIndicatorType, row.NormalizedCandidate, row.ResolutionStatus,
		row.ResolvedIndicatorID, row.RowVersion, row.RuntimeCreatedByID,
		row.CreatedAt, row.RuntimeResolvedByID, row.ResolvedAt,
		row.ResolutionMethod, row.DeletedAt, row.RuntimeDeletedByID}, nil
}

type intervalRowCodec struct{}

func (intervalRowCodec) Encode(row portableIntervalRow) ([]byte, error) { return json.Marshal(row) }
func (intervalRowCodec) Decode(line []byte) (portableIntervalRow, error) {
	var row portableIntervalRow
	return row, decodeIndicatorRow(line, &row)
}
func (intervalRowCodec) Columns() []string { return portableIntervalColumns }
func (intervalRowCodec) Values(row portableIntervalRow) ([]any, error) {
	supportRefs, err := json.Marshal(uuidStrings(row.SupportRefs))
	if err != nil {
		return nil, indicatorSourceFailure(intervalCoherentInvariant)
	}
	return []any{row.IntervalID, row.IncidentID, row.IndicatorRecordID, row.LifecycleState,
		row.ValidFrom, row.ValidTo, row.Confidence, row.Rationale, supportRefs,
		row.Assessor, row.AssessedAt, row.RowVersion, row.RuntimeCreatedByID,
		row.CreatedAt, row.DeletedAt, row.RuntimeDeletedByID}, nil
}

func readIndicatorRows[T any](ctx context.Context, p preparedIndicatorImport, input sourceport.ImportContext, stream incidentportability.PreparedRowStream, codec incidentportability.PreparedRowCodec[T], visit func(T) error) error {
	if !p.binding.matches(input) || !p.staged || input.Preparation == nil {
		return indicatorSourceFailure(representationInvariant)
	}
	return incidentportability.ReadPreparedRows(ctx, input.Preparation, stream, codec, visit)
}

// readIndicatorBatches visits a staged stream at most indicatorSetBatch rows at
// a time, for one set-based statement per batch.
func readIndicatorBatches[T any](ctx context.Context, p preparedIndicatorImport, input sourceport.ImportContext, stream incidentportability.PreparedRowStream, codec incidentportability.PreparedRowCodec[T], visit func([]T) error) error {
	batch := make([]T, 0, indicatorSetBatch)
	if err := readIndicatorRows(ctx, p, input, stream, codec, func(row T) error {
		batch = append(batch, row)
		if len(batch) < indicatorSetBatch {
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

// stage writes the admitted, evaluated rows and their historical attribution.
func (p *preparedIndicatorImport) stage(ctx context.Context, input sourceport.ImportContext) error {
	store := input.Preparation
	if store == nil {
		return errors.New("indicators: incident bundle preparation staging is required")
	}
	var err error
	if p.indicatorRows, err = incidentportability.StagePreparedRows(ctx, store, preparedIndicatorsPath, indicatorRowCodec{}, func(emit func(portableIndicatorRow) error) error {
		return p.walkIndicators(ctx, input, emit)
	}); err != nil {
		return err
	}
	if p.observationRows, err = incidentportability.StagePreparedRows(ctx, store, preparedObservationsPath, observationRowCodec{}, func(emit func(portableObservationRow) error) error {
		return p.walkObservations(ctx, input, emit)
	}); err != nil {
		return err
	}
	if p.intervalRows, err = incidentportability.StagePreparedRows(ctx, store, preparedIntervalsPath, intervalRowCodec{}, func(emit func(portableIntervalRow) error) error {
		return p.walkIntervals(ctx, input, emit)
	}); err != nil {
		return err
	}
	admitted := func(actorID string) bool { _, ok := input.Actors.Lookup(actorID); return ok }
	p.attribution, err = incidentportability.StageAttributions(ctx, store, preparedIndicatorAttrPath, admitted, func(emit func(incidentportability.AttributionRow) error) error {
		row := func(table, id, column string, actor *uuid.UUID) error {
			if actor == nil {
				return nil
			}
			return emit(incidentportability.AttributionRow{SourceTable: table, SourceRowID: id, SourceColumn: column, SourceActorID: actor.String()})
		}
		for _, column := range []string{"created_by_user_id", "updated_by_user_id", "deleted_by_user_id"} {
			if err := incidentportability.ReadPreparedRows(ctx, store, p.indicatorRows, indicatorRowCodec{}, func(indicator portableIndicatorRow) error {
				actor := map[string]*uuid.UUID{"created_by_user_id": &indicator.PortableCreatedByID, "updated_by_user_id": &indicator.PortableUpdatedByID, "deleted_by_user_id": indicator.PortableDeletedByID}[column]
				return row("indicators", indicator.RecordID.String(), column, actor)
			}); err != nil {
				return err
			}
		}
		for _, column := range []string{"created_by_user_id", "resolved_by_user_id", "deleted_by_user_id"} {
			if err := incidentportability.ReadPreparedRows(ctx, store, p.observationRows, observationRowCodec{}, func(observation portableObservationRow) error {
				actor := map[string]*uuid.UUID{"created_by_user_id": &observation.PortableCreatedByID, "resolved_by_user_id": observation.PortableResolvedByID, "deleted_by_user_id": observation.PortableDeletedByID}[column]
				return row("indicator_observations", observation.ObservationID.String(), column, actor)
			}); err != nil {
				return err
			}
		}
		for _, column := range []string{"created_by_user_id", "deleted_by_user_id"} {
			if err := incidentportability.ReadPreparedRows(ctx, store, p.intervalRows, intervalRowCodec{}, func(interval portableIntervalRow) error {
				actor := map[string]*uuid.UUID{"created_by_user_id": &interval.PortableCreatedByID, "deleted_by_user_id": interval.PortableDeletedByID}[column]
				return row("indicator_state_intervals", interval.IntervalID.String(), column, actor)
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	p.staged = true
	return nil
}
