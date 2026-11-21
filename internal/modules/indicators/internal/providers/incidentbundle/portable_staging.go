package incidentbundle

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// validateIndicatorSourceReferences checks the staged rows' database-side
// facts before any authoritative row is written: exact envelopes, active
// identity uniqueness and source references. Every check is set-based; the
// failure accumulator keeps the declared precedence independent of row order.
func validateIndicatorSourceReferences(ctx context.Context, tx pgx.Tx, prepared preparedIndicatorImport, input sourceport.ImportContext) error {
	// Owner-local transaction staging supplies indexed source membership and exact
	// active-identity checks before any authoritative row or sidecar is inserted.
	sources, err := incidentportability.OpenPreparedCopy(ctx, input.Preparation, prepared.indicatorRows, indicatorSourceCodec{})
	if err != nil {
		return err
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"public", "indicators_bundle_validation_sources"}, indicatorValidationSourceColumns, sources)
	if err := sources.Finish(written, err); err != nil {
		return err
	}
	var failures indicatorFailureAccumulator
	if err := checkIndicatorEnvelopes(ctx, tx, prepared, input, &failures); err != nil {
		return err
	}
	// Preparation proved the historical envelope actors against the records
	// family's published attribution; the verdict keeps this rank.
	if prepared.verdicts.At(rankApplyEnvelopeActors) != nil {
		failures.add(representationInvariant)
	}
	var duplicate bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM public.indicators_bundle_validation_sources WHERE active GROUP BY kind,dedupe_key HAVING count(*)>1)
 OR EXISTS (SELECT 1 FROM public.indicators_bundle_validation_sources s JOIN indicators i ON i.record_id=s.id)
 OR EXISTS (SELECT 1 FROM public.indicators_bundle_validation_sources s JOIN indicators i ON i.incident_id=$1 AND i.indicator_type=s.kind AND i.dedupe_key=s.dedupe_key JOIN records r ON r.record_id=i.record_id WHERE s.active AND r.deleted_at IS NULL)`, input.IncidentID).Scan(&duplicate); err != nil {
		return err
	}
	if duplicate {
		failures.add(identityUniqueInvariant)
	}
	// references reports whether any record is absent from this incident or any
	// indicator is absent from the staged collection.
	references := func(records, indicators []uuid.UUID) (bool, error) {
		var missingRecord, missingIndicator bool
		if err := tx.QueryRow(ctx, `SELECT
  EXISTS (SELECT 1 FROM unnest($2::uuid[]) AS r(id) WHERE NOT EXISTS (SELECT 1 FROM records WHERE incident_id=$1 AND record_id=r.id)),
  EXISTS (SELECT 1 FROM unnest($3::uuid[]) AS i(id) WHERE NOT EXISTS (SELECT 1 FROM public.indicators_bundle_validation_sources WHERE id=i.id))`, input.IncidentID, records, indicators).Scan(&missingRecord, &missingIndicator); err != nil {
			return false, err
		}
		return missingRecord || missingIndicator, nil
	}
	if err := readIndicatorBatches(ctx, prepared, input, prepared.observationRows, observationRowCodec{}, func(batch []portableObservationRow) error {
		records, indicators := make([]uuid.UUID, 0, len(batch)), []uuid.UUID{}
		for _, row := range batch {
			records = append(records, row.SourceRecordID)
			if row.ResolvedIndicatorID != nil {
				indicators = append(indicators, *row.ResolvedIndicatorID)
			}
		}
		missing, err := references(records, indicators)
		if missing {
			failures.add(observationIncidentInvariant)
		}
		return err
	}); err != nil {
		return err
	}
	if err := readIndicatorBatches(ctx, prepared, input, prepared.intervalRows, intervalRowCodec{}, func(batch []portableIntervalRow) error {
		records, indicators := []uuid.UUID{}, make([]uuid.UUID, 0, len(batch))
		for _, row := range batch {
			indicators = append(indicators, row.IndicatorRecordID)
			records = append(records, row.SupportRefs...)
		}
		missing, err := references(records, indicators)
		if missing {
			failures.add(intervalIncidentInvariant)
		}
		return err
	}); err != nil {
		return err
	}
	if err := failures.err(); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM public.indicators_bundle_validation_sources WHERE transaction_id=pg_current_xact_id()`)
	return err
}

// checkIndicatorEnvelopes compares each staged indicator with its applied
// record envelope, one envelope read per batch.
func checkIndicatorEnvelopes(ctx context.Context, tx pgx.Tx, prepared preparedIndicatorImport, input sourceport.ImportContext, failures *indicatorFailureAccumulator) error {
	return readIndicatorBatches(ctx, prepared, input, prepared.indicatorRows, indicatorRowCodec{}, func(batch []portableIndicatorRow) error {
		ids := make([]uuid.UUID, len(batch))
		for index, row := range batch {
			ids[index] = row.RecordID
		}
		envelopes, err := loadPortableRecordEnvelopes(ctx, tx, input.IncidentID, ids)
		if err != nil {
			return err
		}
		for _, row := range batch {
			envelope, present := envelopes[row.RecordID]
			if !present || !indicatorEnvelopeEqual(row, envelope) {
				failures.add(representationInvariant)
			}
		}
		return nil
	})
}
