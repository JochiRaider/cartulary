package incidentbundle

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func validateIndicatorSourceReferences(ctx context.Context, tx pgx.Tx, prepared preparedIndicatorImport, input sourceport.ImportContext) error {
	// Owner-local transaction staging supplies indexed source membership and exact
	// active-identity checks before any authoritative row or sidecar is inserted.
	var failures indicatorFailureAccumulator
	if err := prepared.indicators(ctx, input, func(row portableIndicatorRow) error {
		_, err := tx.Exec(ctx, `INSERT INTO public.indicators_bundle_validation_sources VALUES ($1,$2,$3,$4)`, row.RecordID, row.IndicatorType, row.DedupeKey, row.DeletedAt == nil)
		if err != nil {
			return err
		}
		envelope, present, err := loadPortableRecordEnvelope(ctx, tx, input.IncidentID, row.RecordID)
		if err != nil {
			return err
		}
		if !present || !indicatorEnvelopeEqual(row, envelope) || !indicatorEnvelopeAttributionsEqual(row, input) {
			failures.add(representationInvariant)
		}
		return nil
	}); err != nil {
		return err
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
	recordPresent := func(id uuid.UUID, invariant string) error {
		var present bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM records WHERE incident_id=$1 AND record_id=$2)`, input.IncidentID, id).Scan(&present); err != nil {
			return err
		}
		if !present {
			failures.add(invariant)
		}
		return nil
	}
	indicatorPresent := func(id uuid.UUID, invariant string) error {
		var present bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM public.indicators_bundle_validation_sources WHERE id=$1)`, id).Scan(&present); err != nil {
			return err
		}
		if !present {
			failures.add(invariant)
		}
		return nil
	}
	if err := prepared.observations(ctx, input, func(row portableObservationRow) error {
		if err := recordPresent(row.SourceRecordID, observationIncidentInvariant); err != nil {
			return err
		}
		if row.ResolvedIndicatorID != nil {
			return indicatorPresent(*row.ResolvedIndicatorID, observationIncidentInvariant)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := prepared.intervals(ctx, input, func(row portableIntervalRow) error {
		if err := indicatorPresent(row.IndicatorRecordID, intervalIncidentInvariant); err != nil {
			return err
		}
		for _, id := range row.SupportRefs {
			if err := recordPresent(id, intervalIncidentInvariant); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := failures.err(); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `DELETE FROM public.indicators_bundle_validation_sources WHERE transaction_id=pg_current_xact_id()`)
	return err
}
