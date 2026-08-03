package incidentbundle

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
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

func validatePreparedIndicatorImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedIndicatorImport,
	importContext sourceport.ImportContext,
) error {
	if tx == nil || !prepared.binding.matches(importContext) || importContext.Attributions == nil {
		return indicatorSourceFailure(representationInvariant)
	}
	envelopes, err := loadPortableRecordEnvelopes(ctx, tx, importContext.IncidentID)
	if err != nil {
		return indicatorSourceFailure(representationInvariant)
	}
	actualIndicators, err := loadPortableIndicators(ctx, sourceport.ExportContext{
		Query: tx, IncidentID: importContext.IncidentID,
	})
	if err != nil {
		return indicatorSourceFailure(representationInvariant)
	}
	actualObservations, err := loadPortableObservations(ctx, sourceport.ExportContext{
		Query: tx, IncidentID: importContext.IncidentID,
	})
	if err != nil {
		return indicatorSourceFailure(observationCoherentInvariant)
	}
	actualIntervals, err := loadPortableIntervals(ctx, sourceport.ExportContext{
		Query: tx, IncidentID: importContext.IncidentID,
	})
	if err != nil {
		return indicatorSourceFailure(intervalCoherentInvariant)
	}

	var failures []indicatorFailureCandidate
	indicatorByID := make(map[uuid.UUID]portableIndicatorRow, len(prepared.indicators))
	for _, row := range prepared.indicators {
		indicatorByID[row.RecordID] = row
		envelope, present := envelopes[row.RecordID]
		if !present || !indicatorEnvelopeEqual(row, envelope) ||
			!indicatorEnvelopeAttributionsEqual(row, importContext) {
			failures = append(failures, indicatorFailure(
				representationInvariant, indicatorsBundlePath, row.RecordID.String(), row.RecordID,
			))
		}
	}
	portableIndicatorSetsEqual(prepared.indicators, actualIndicators, &failures)

	for _, row := range prepared.observations {
		if _, present := envelopes[row.SourceRecordID]; !present {
			failures = append(failures, indicatorFailure(
				observationIncidentInvariant, observationsBundlePath,
				row.ObservationID.String(), row.SourceRecordID,
			))
		}
		if row.ResolvedIndicatorID != nil {
			if _, present := indicatorByID[*row.ResolvedIndicatorID]; !present {
				failures = append(failures, indicatorFailure(
					observationIncidentInvariant, observationsBundlePath,
					row.ObservationID.String(), row.ResolvedIndicatorID,
				))
			}
		}
	}
	portableObservationSetsEqual(prepared.observations, actualObservations, &failures)

	for _, row := range prepared.intervals {
		if _, present := indicatorByID[row.IndicatorRecordID]; !present {
			failures = append(failures, indicatorFailure(
				intervalIncidentInvariant, intervalsBundlePath,
				row.IntervalID.String(), row.IndicatorRecordID,
			))
		}
		for _, supportID := range row.SupportRefs {
			if _, present := envelopes[supportID]; !present {
				failures = append(failures, indicatorFailure(
					intervalIncidentInvariant, intervalsBundlePath,
					row.IntervalID.String(), supportID,
				))
			}
		}
	}
	portableIntervalSetsEqual(prepared.intervals, actualIntervals, &failures)
	return selectedIndicatorFailure(failures)
}

func loadPortableRecordEnvelopes(
	ctx context.Context,
	query incidentportability.Queryer,
	incidentID uuid.UUID,
) (map[uuid.UUID]portableRecordEnvelope, error) {
	rows, err := query.Query(ctx, `
SELECT record_id, incident_id, record_type, created_at, created_by_user_id,
       updated_at, updated_by_user_id, row_version, deleted_at, deleted_by_user_id
  FROM records
 WHERE incident_id = $1
 ORDER BY record_id
`, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[uuid.UUID]portableRecordEnvelope)
	for rows.Next() {
		var envelope portableRecordEnvelope
		var deletedAt pgtype.Timestamptz
		var deletedBy pgtype.UUID
		if err := rows.Scan(
			&envelope.recordID, &envelope.incidentID, &envelope.recordType,
			&envelope.createdAt, &envelope.createdByUserID, &envelope.updatedAt,
			&envelope.updatedByUserID, &envelope.rowVersion, &deletedAt, &deletedBy,
		); err != nil {
			return nil, err
		}
		envelope.deletedAt = timeFromPG(deletedAt)
		envelope.deletedByUserID = uuidFromPG(deletedBy)
		result[envelope.recordID] = envelope
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
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
	for _, attribution := range importContext.Attributions.ImportedAttributions() {
		if attribution.SourceTable == "records" && attribution.SourceRowID == row.RecordID.String() {
			found[attribution.SourceColumn] = attribution.SourceActorID
		}
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

func portableIndicatorSetsEqual(
	expected []portableIndicatorRow,
	actual []portableIndicatorRow,
	failures *[]indicatorFailureCandidate,
) {
	want := make(map[uuid.UUID]portableIndicatorRow, len(expected))
	for _, row := range expected {
		want[row.RecordID] = row
	}
	seen := make(map[uuid.UUID]struct{}, len(actual))
	for _, row := range actual {
		expectedRow, present := want[row.RecordID]
		if !present || !portableIndicatorEqual(expectedRow, row) {
			*failures = append(*failures, indicatorFailure(
				representationInvariant, indicatorsBundlePath, row.RecordID.String(), row.RecordID,
			))
		}
		seen[row.RecordID] = struct{}{}
	}
	for _, row := range expected {
		if _, present := seen[row.RecordID]; !present {
			*failures = append(*failures, indicatorFailure(
				representationInvariant, indicatorsBundlePath, row.RecordID.String(), row.RecordID,
			))
		}
	}
}

func portableObservationSetsEqual(
	expected []portableObservationRow,
	actual []portableObservationRow,
	failures *[]indicatorFailureCandidate,
) {
	want := make(map[uuid.UUID]portableObservationRow, len(expected))
	for _, row := range expected {
		want[row.ObservationID] = row
	}
	seen := make(map[uuid.UUID]struct{}, len(actual))
	for _, row := range actual {
		expectedRow, present := want[row.ObservationID]
		if !present {
			*failures = append(*failures, indicatorFailure(
				repeatedObservationInvariant, observationsBundlePath,
				row.ObservationID.String(), row.ObservationID,
			))
		} else if !portableObservationEqual(expectedRow, row) {
			*failures = append(*failures, indicatorFailure(
				observationCoherentInvariant, observationsBundlePath,
				row.ObservationID.String(), row.ObservationID,
			))
		}
		seen[row.ObservationID] = struct{}{}
	}
	for _, row := range expected {
		if _, present := seen[row.ObservationID]; !present {
			*failures = append(*failures, indicatorFailure(
				repeatedObservationInvariant, observationsBundlePath,
				row.ObservationID.String(), row.ObservationID,
			))
		}
	}
}

func portableIntervalSetsEqual(
	expected []portableIntervalRow,
	actual []portableIntervalRow,
	failures *[]indicatorFailureCandidate,
) {
	want := make(map[uuid.UUID]portableIntervalRow, len(expected))
	for _, row := range expected {
		want[row.IntervalID] = row
	}
	seen := make(map[uuid.UUID]struct{}, len(actual))
	for _, row := range actual {
		expectedRow, present := want[row.IntervalID]
		if !present || !portableIntervalEqual(expectedRow, row) {
			*failures = append(*failures, indicatorFailure(
				intervalCoherentInvariant, intervalsBundlePath,
				row.IntervalID.String(), row.IntervalID,
			))
		}
		seen[row.IntervalID] = struct{}{}
	}
	for _, row := range expected {
		if _, present := seen[row.IntervalID]; !present {
			*failures = append(*failures, indicatorFailure(
				intervalCoherentInvariant, intervalsBundlePath,
				row.IntervalID.String(), row.IntervalID,
			))
		}
	}
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
