package indicators

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/observationhistory"
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
	value, err := observationhistory.Scan(scanner)
	return IndicatorObservationRecord(value), err
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
