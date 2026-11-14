package indicators

import (
	"fmt"
	"reflect"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/intervalhistory"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/observationhistory"
	"github.com/google/uuid"
)

func referenceFromRecord(record indicatorRecord) IndicatorReference {
	return IndicatorReference{
		RecordID: record.RecordID, IncidentID: record.IncidentID,
		IndicatorType: record.IndicatorType, ValueKind: record.ValueKind,
		DisplayValue:    record.DisplayValue,
		NormalizedValue: cloneStringPointer(record.NormalizedValue),
		DedupeKey:       record.DedupeKey,
	}
}

func entityVersionID(prefix string, recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("%s:%s:%d", prefix, recordID.String(), rowVersion)
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func buildIndicatorObservationValue(record IndicatorObservationRecord) map[string]any {
	return observationhistory.Value(record).Map()
}

func buildIndicatorLifecycleValue(record IndicatorLifecycleIntervalRecord) map[string]any {
	return intervalhistory.Value(record).Map()
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func formatTimestampPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func derefString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func derefInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func stringPointer(value string) *string {
	cloned := value
	return &cloned
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func jsonEqual(left map[string]any, right map[string]any) bool {
	return reflect.DeepEqual(left, right)
}
