package incidentbundle

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func applyPreparedIndicatorImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedIndicatorImport,
	importContext sourceport.ImportContext,
) error {
	if tx == nil || !prepared.binding.matches(importContext) ||
		importContext.ActorUserID == uuid.Nil || importContext.Attributions == nil ||
		!prepared.staged || importContext.Preparation == nil {
		return indicatorSourceFailure(representationInvariant)
	}
	if err := validateIndicatorSourceReferences(ctx, tx, prepared, importContext); err != nil {
		return err
	}
	// One COPY per family from the staged rows; every reference was checked
	// above. Attribution is flushed by the core in step 9.
	store := importContext.Preparation
	indicators, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.indicatorRows, indicatorRowCodec{})
	if err != nil {
		return classifyIndicatorApplyError(err, representationInvariant)
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"indicators"}, portableIndicatorColumns, indicators)
	if err := indicators.Finish(written, err); err != nil {
		return classifyIndicatorApplyError(err, representationInvariant)
	}
	observations, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.observationRows, observationRowCodec{})
	if err != nil {
		return classifyIndicatorApplyError(err, observationCoherentInvariant)
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"indicator_observations"}, portableObservationColumns, observations)
	if err := observations.Finish(written, err); err != nil {
		return classifyIndicatorApplyError(err, observationCoherentInvariant)
	}
	intervals, err := incidentportability.OpenPreparedCopy(ctx, store, prepared.intervalRows, intervalRowCodec{})
	if err != nil {
		return classifyIndicatorApplyError(err, intervalCoherentInvariant)
	}
	written, err = tx.CopyFrom(ctx, pgx.Identifier{"indicator_state_intervals"}, portableIntervalColumns, intervals)
	if err := intervals.Finish(written, err); err != nil {
		return classifyIndicatorApplyError(err, intervalCoherentInvariant)
	}
	return nil
}

func classifyIndicatorApplyError(err error, invariant string) error {
	if err == nil {
		return indicatorSourceFailure(invariant)
	}
	var postgresFailure *pgconn.PgError
	if errors.As(err, &postgresFailure) {
		return indicatorSourceFailure(invariant)
	}
	// Source failures are deliberately closed and do not disclose adapter,
	// relation, constraint, or hostile row details to the coordinator.
	return indicatorSourceFailure(invariant)
}
