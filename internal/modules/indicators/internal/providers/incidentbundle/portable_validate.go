package incidentbundle

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
)

type portableRecordEnvelope struct {
	recordID        uuid.UUID
	incidentID      uuid.UUID
	recordType      string
	createdAt       time.Time
	createdByUserID uuid.UUID
	updatedAt       time.Time
	updatedByUserID uuid.UUID
	rowVersion      int64
	deletedAt       *time.Time
	deletedByUserID *uuid.UUID
}

func validatePreparedIndicatorImportTx(evaluate identity.Evaluator, ctx context.Context, tx pgx.Tx, prepared preparedIndicatorImport, input sourceport.ImportContext) error {
	if tx == nil || !prepared.binding.matches(input) || input.Attributions == nil {
		return indicatorSourceFailure(representationInvariant)
	}
	prepared.evaluate = evaluate
	var failures indicatorFailureAccumulator
	if err := prepared.indicators(ctx, input, func(row portableIndicatorRow) error {
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
	if sameIDs, sameValues, err := compareIndicatorSource(ctx, tx, input, func(visit func(portableIndicatorRow) error) error { return prepared.indicators(ctx, input, visit) }, func(ctx context.Context, input sourceport.ExportContext, after *uuid.UUID) ([]portableIndicatorRow, error) {
		return loadPortableIndicatorsPage(ctx, evaluate, input, after)
	}, func(row portableIndicatorRow) uuid.UUID { return row.RecordID }, portableIndicatorEqual); err != nil {
		return err
	} else {
		if !sameIDs {
			failures.add(representationInvariant)
		}
		if !sameValues {
			failures.add(representationInvariant)
		}
	}
	if sameIDs, sameValues, err := compareIndicatorSource(ctx, tx, input, func(visit func(portableObservationRow) error) error { return prepared.observations(ctx, input, visit) }, func(ctx context.Context, input sourceport.ExportContext, after *uuid.UUID) ([]portableObservationRow, error) {
		return loadPortableObservationsPage(ctx, evaluate, input, after)
	}, func(row portableObservationRow) uuid.UUID { return row.ObservationID }, portableObservationEqual); err != nil {
		return err
	} else {
		if !sameIDs {
			failures.add(repeatedObservationInvariant)
		}
		if !sameValues {
			failures.add(observationCoherentInvariant)
		}
	}
	if sameIDs, sameValues, err := compareIndicatorSource(ctx, tx, input, func(visit func(portableIntervalRow) error) error { return prepared.intervals(ctx, input, visit) }, loadPortableIntervalsPage, func(row portableIntervalRow) uuid.UUID { return row.IntervalID }, portableIntervalEqual); err != nil {
		return err
	} else {
		if !sameIDs {
			failures.add(intervalCoherentInvariant)
		}
		if !sameValues {
			failures.add(intervalCoherentInvariant)
		}
	}
	for _, family := range []struct{ path, table, identity, invariant string }{
		{indicatorsBundlePath, "indicators", "record_id", representationInvariant},
		{observationsBundlePath, "indicator_observations", "indicator_observation_id", observationCoherentInvariant},
		{intervalsBundlePath, "indicator_state_intervals", "indicator_state_interval_id", intervalCoherentInvariant},
	} {
		if err := incidentportability.ValidateTopLevelAttributions(ctx, prepared.members, family.path, family.table, family.identity, input.Attributions); err != nil {
			failures.add(family.invariant)
		}
	}
	return failures.err()
}

func loadPortableRecordEnvelope(ctx context.Context, tx pgx.Tx, incidentID, recordID uuid.UUID) (portableRecordEnvelope, bool, error) {
	var envelope portableRecordEnvelope
	var deletedAt pgtype.Timestamptz
	var deletedBy pgtype.UUID
	err := tx.QueryRow(ctx, `SELECT record_id,incident_id,record_type,created_at,created_by_user_id,updated_at,updated_by_user_id,row_version,deleted_at,deleted_by_user_id FROM records WHERE incident_id=$1 AND record_id=$2`, incidentID, recordID).Scan(&envelope.recordID, &envelope.incidentID, &envelope.recordType, &envelope.createdAt, &envelope.createdByUserID, &envelope.updatedAt, &envelope.updatedByUserID, &envelope.rowVersion, &deletedAt, &deletedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return envelope, false, nil
	}
	if err != nil {
		return envelope, false, err
	}
	envelope.deletedAt = timeFromPG(deletedAt)
	envelope.deletedByUserID = uuidFromPG(deletedBy)
	return envelope, true, nil
}

// Ordered merge distinguishes missing/repeated observations from a changed
// value on a surviving observation, without collection-sized identity maps.
func compareIndicatorSource[T any](ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, walk func(func(T) error) error, load func(context.Context, sourceport.ExportContext, *uuid.UUID) ([]T, error), identity func(T) uuid.UUID, equal func(T, T) bool) (bool, bool, error) {
	var page []T
	var after *uuid.UUID
	index := 0
	exhausted := false
	peek := func() (T, bool, error) {
		if index == len(page) && !exhausted {
			var err error
			page, err = load(ctx, sourceport.ExportContext{Query: tx, IncidentID: input.IncidentID}, after)
			if err != nil {
				var zero T
				return zero, false, err
			}
			index = 0
			if len(page) == 0 {
				exhausted = true
			} else {
				last := identity(page[len(page)-1])
				after = &last
			}
		}
		if exhausted {
			var zero T
			return zero, false, nil
		}
		return page[index], true, nil
	}
	sameIDs, sameValues := true, true
	err := walk(func(expected T) error {
		actual, present, err := peek()
		if err != nil {
			return err
		}
		for present && identity(actual).String() < identity(expected).String() {
			sameIDs = false
			index++
			actual, present, err = peek()
			if err != nil {
				return err
			}
		}
		if !present || identity(actual) != identity(expected) {
			sameIDs = false
			return nil
		}
		if !equal(expected, actual) {
			sameValues = false
		}
		index++
		return nil
	})
	if err != nil {
		return false, false, err
	}
	_, extra, err := peek()
	return sameIDs && !extra, sameValues, err
}

func indicatorEnvelopeEqual(row portableIndicatorRow, envelope portableRecordEnvelope) bool {
	return envelope.recordID == row.RecordID && envelope.incidentID == row.IncidentID &&
		envelope.recordType == "indicator" && envelope.rowVersion == row.RowVersion &&
		envelope.createdAt.Equal(row.CreatedAt) && envelope.updatedAt.Equal(row.UpdatedAt) &&
		envelope.createdByUserID == row.RuntimeCreatedByID &&
		envelope.updatedByUserID == row.RuntimeUpdatedByID &&
		portableTimePointersEqual(envelope.deletedAt, row.DeletedAt) &&
		portableUUIDPointersEqual(envelope.deletedByUserID, row.RuntimeDeletedByID)
}

func indicatorEnvelopeAttributionsEqual(row portableIndicatorRow, importContext sourceport.ImportContext) bool {
	want := map[string]string{
		"created_by_user_id": row.PortableCreatedByID.String(),
		"updated_by_user_id": row.PortableUpdatedByID.String(),
	}
	if row.PortableDeletedByID != nil {
		want["deleted_by_user_id"] = row.PortableDeletedByID.String()
	}
	found := map[string]string{}

	if err := importContext.Attributions.VisitImportedAttributions("records", row.RecordID.String(), func(attribution incidentportability.ImportedAttribution) error {
		found[attribution.SourceColumn] = attribution.SourceActorID
		return nil
	}); err != nil {
		return false
	}

	if len(found) != len(want) {
		return false
	}
	for column, actorID := range want {
		if found[column] != actorID {
			return false
		}
	}
	return true
}

func portableIndicatorEqual(left, right portableIndicatorRow) bool {
	return left.RecordID == right.RecordID && left.IncidentID == right.IncidentID &&
		left.IndicatorType == right.IndicatorType && left.ValueKind == right.ValueKind &&
		left.DisplayValue == right.DisplayValue &&
		portableStringPointersEqual(left.NormalizedValue, right.NormalizedValue) &&
		left.DedupeKey == right.DedupeKey &&
		portableStringPointersEqual(left.DefangedValue, right.DefangedValue) &&
		portableStringPointersEqual(left.HashAlgorithm, right.HashAlgorithm) &&
		portableStringPointersEqual(left.HashValue, right.HashValue) &&
		portableStringPointersEqual(left.STIXPattern, right.STIXPattern) &&
		left.RowVersion == right.RowVersion && left.CreatedAt.Equal(right.CreatedAt) &&
		left.UpdatedAt.Equal(right.UpdatedAt) &&
		left.RuntimeCreatedByID == right.RuntimeCreatedByID &&
		left.RuntimeUpdatedByID == right.RuntimeUpdatedByID &&
		portableTimePointersEqual(left.DeletedAt, right.DeletedAt) &&
		portableUUIDPointersEqual(left.RuntimeDeletedByID, right.RuntimeDeletedByID)
}

func portableObservationEqual(left, right portableObservationRow) bool {
	return left.ObservationID == right.ObservationID && left.IncidentID == right.IncidentID &&
		left.SourceRecordID == right.SourceRecordID && left.SourceFieldKey == right.SourceFieldKey &&
		left.OriginKind == right.OriginKind && left.OriginLocator == right.OriginLocator &&
		left.ObservedText == right.ObservedText &&
		portableStringPointersEqual(left.ParsedIndicatorType, right.ParsedIndicatorType) &&
		portableStringPointersEqual(left.NormalizedCandidate, right.NormalizedCandidate) &&
		left.ResolutionStatus == right.ResolutionStatus &&
		portableUUIDPointersEqual(left.ResolvedIndicatorID, right.ResolvedIndicatorID) &&
		left.RowVersion == right.RowVersion && left.RuntimeCreatedByID == right.RuntimeCreatedByID &&
		left.CreatedAt.Equal(right.CreatedAt) &&
		portableUUIDPointersEqual(left.RuntimeResolvedByID, right.RuntimeResolvedByID) &&
		portableTimePointersEqual(left.ResolvedAt, right.ResolvedAt) &&
		portableStringPointersEqual(left.ResolutionMethod, right.ResolutionMethod) &&
		portableTimePointersEqual(left.DeletedAt, right.DeletedAt) &&
		portableUUIDPointersEqual(left.RuntimeDeletedByID, right.RuntimeDeletedByID)
}

func portableIntervalEqual(left, right portableIntervalRow) bool {
	return left.IntervalID == right.IntervalID && left.IncidentID == right.IncidentID &&
		left.IndicatorRecordID == right.IndicatorRecordID &&
		left.LifecycleState == right.LifecycleState && left.ValidFrom.Equal(right.ValidFrom) &&
		portableTimePointersEqual(left.ValidTo, right.ValidTo) &&
		portableIntPointersEqual(left.Confidence, right.Confidence) &&
		portableStringPointersEqual(left.Rationale, right.Rationale) &&
		portableUUIDSlicesEqual(left.SupportRefs, right.SupportRefs) &&
		portableStringPointersEqual(left.Assessor, right.Assessor) &&
		left.AssessedAt.Equal(right.AssessedAt) && left.RowVersion == right.RowVersion &&
		left.RuntimeCreatedByID == right.RuntimeCreatedByID && left.CreatedAt.Equal(right.CreatedAt) &&
		portableTimePointersEqual(left.DeletedAt, right.DeletedAt) &&
		portableUUIDPointersEqual(left.RuntimeDeletedByID, right.RuntimeDeletedByID)
}

func portableIntPointersEqual(left, right *int) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func portableUUIDSlicesEqual(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
