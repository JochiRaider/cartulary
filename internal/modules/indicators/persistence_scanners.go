package indicators

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
)

type indicatorRecord struct {
	RecordID        uuid.UUID
	IncidentID      uuid.UUID
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue *string
	DedupeKey       string
	DefangedValue   *string
	HashAlgorithm   *string
	HashValue       *string
	STIXPattern     *string
	RowVersion      int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedByUser   uuid.UUID
	UpdatedByUser   uuid.UUID
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
}

func scanIndicatorRecord(scanner interface{ Scan(dest ...any) error }) (indicatorRecord, error) {
	var (
		record           indicatorRecord
		rawRecordID      pgtype.UUID
		rawIncidentID    pgtype.UUID
		rawNormalized    pgtype.Text
		rawDefanged      pgtype.Text
		rawHashAlgorithm pgtype.Text
		rawHashValue     pgtype.Text
		rawSTIXPattern   pgtype.Text
		rawCreatedBy     pgtype.UUID
		rawUpdatedBy     pgtype.UUID
		rawDeletedAt     pgtype.Timestamptz
		rawDeletedBy     pgtype.UUID
	)
	if err := scanner.Scan(
		&rawRecordID,
		&rawIncidentID,
		&record.IndicatorType,
		&record.ValueKind,
		&record.DisplayValue,
		&rawNormalized,
		&record.DedupeKey,
		&rawDefanged,
		&rawHashAlgorithm,
		&rawHashValue,
		&rawSTIXPattern,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&rawCreatedBy,
		&rawUpdatedBy,
		&rawDeletedAt,
		&rawDeletedBy,
	); err != nil {
		return indicatorRecord{}, err
	}
	if !rawRecordID.Valid || !rawIncidentID.Valid || !rawCreatedBy.Valid || !rawUpdatedBy.Valid {
		return indicatorRecord{}, fmt.Errorf("scan indicator: required UUID is NULL")
	}
	record.RecordID = uuid.UUID(rawRecordID.Bytes)
	record.IncidentID = uuid.UUID(rawIncidentID.Bytes)
	record.NormalizedValue = textPointer(rawNormalized)
	record.DefangedValue = textPointer(rawDefanged)
	record.HashAlgorithm = textPointer(rawHashAlgorithm)
	record.HashValue = textPointer(rawHashValue)
	record.STIXPattern = textPointer(rawSTIXPattern)
	record.CreatedByUser = uuid.UUID(rawCreatedBy.Bytes)
	record.UpdatedByUser = uuid.UUID(rawUpdatedBy.Bytes)
	record.DeletedAt = timePointerFromPG(rawDeletedAt)
	record.DeletedByUserID = uuidPointerFromPG(rawDeletedBy)
	return record, nil
}

func scanIndicatorObservationRecord(scanner interface{ Scan(dest ...any) error }) (IndicatorObservationRecord, error) {
	var (
		record              IndicatorObservationRecord
		rawObservationID    pgtype.UUID
		rawIncidentID       pgtype.UUID
		rawSourceRecordID   pgtype.UUID
		rawOriginKind       string
		rawParsedType       pgtype.Text
		rawNormalized       pgtype.Text
		rawResolvedID       pgtype.UUID
		rawCreatedBy        pgtype.UUID
		rawResolvedBy       pgtype.UUID
		rawResolvedAt       pgtype.Timestamptz
		rawResolutionMethod pgtype.Text
		rawDeletedAt        pgtype.Timestamptz
		rawDeletedBy        pgtype.UUID
	)
	if err := scanner.Scan(
		&rawObservationID,
		&rawIncidentID,
		&rawSourceRecordID,
		&record.SourceFieldKey,
		&rawOriginKind,
		&record.OriginLocator,
		&record.ObservedText,
		&rawParsedType,
		&rawNormalized,
		&record.ResolutionStatus,
		&rawResolvedID,
		&record.RowVersion,
		&rawCreatedBy,
		&record.CreatedAt,
		&rawResolvedBy,
		&rawResolvedAt,
		&rawResolutionMethod,
		&rawDeletedAt,
		&rawDeletedBy,
	); err != nil {
		return IndicatorObservationRecord{}, err
	}
	originKind, err := indicatororigin.Parse(rawOriginKind)
	if err != nil {
		return IndicatorObservationRecord{}, err
	}
	record.OriginKind = originKind.String()
	if !rawObservationID.Valid || !rawIncidentID.Valid || !rawSourceRecordID.Valid || !rawCreatedBy.Valid {
		return IndicatorObservationRecord{}, fmt.Errorf("scan indicator observation: required UUID is NULL")
	}
	record.ObservationID = uuid.UUID(rawObservationID.Bytes)
	record.IncidentID = uuid.UUID(rawIncidentID.Bytes)
	record.SourceRecordID = uuid.UUID(rawSourceRecordID.Bytes)
	record.ParsedIndicatorType = textPointer(rawParsedType)
	record.NormalizedCandidate = textPointer(rawNormalized)
	record.ResolvedIndicatorRecordID = uuidPointerFromPG(rawResolvedID)
	record.CreatedByUserID = uuid.UUID(rawCreatedBy.Bytes)
	record.ResolvedByUserID = uuidPointerFromPG(rawResolvedBy)
	record.ResolvedAt = timePointerFromPG(rawResolvedAt)
	record.ResolutionMethod = textPointer(rawResolutionMethod)
	record.DeletedAt = timePointerFromPG(rawDeletedAt)
	record.DeletedByUserID = uuidPointerFromPG(rawDeletedBy)
	return record, nil
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func uuidPointerFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.UUID(value.Bytes)
	return &parsed
}

func timePointerFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}
