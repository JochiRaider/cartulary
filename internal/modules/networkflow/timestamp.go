package networkflow

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	timestampProfileSchemaID = "cartulary.network_flow.timestamp_profile.v1"
	timestampRulesetID       = "tzdb-2026c"
	maxTimestampUnixSeconds  = int64(253402300799)
	maxTimestampUnixMillis   = uint64(253402300799999)
)

var exactRFC3339Pattern = regexp.MustCompile(`^([0-9]{4})-([0-9]{2})-([0-9]{2})T([0-9]{2}):([0-9]{2}):([0-9]{2})(?:\.([0-9]{1,6}))?(Z|[+-][0-9]{2}:[0-9]{2})?$`)

type timestampValidationError struct {
	reason string
}

func (e *timestampValidationError) Error() string {
	return "invalid Network Flow timestamp: " + e.reason
}

func timestampError(reason string) error { return &timestampValidationError{reason: reason} }

func timestampReason(err error) string {
	var validationErr *timestampValidationError
	if errors.As(err, &validationErr) {
		return validationErr.reason
	}
	if errors.Is(err, errOutOfRange) {
		return "out_of_range"
	}
	return "invalid_syntax"
}

func validateTimestampProfile(profile TimestampProfile, maxColumns int) error {
	conflict := func(reason string) error {
		return &MappingValidationError{Code: "network_flow_mapping_conflict", ReasonCode: reason}
	}
	if profile.SchemaID != timestampProfileSchemaID {
		return conflict("variant_member_conflict")
	}
	hasRFCMembers := profile.Timezone != nil || profile.TimezoneRulesetID != nil || profile.AmbiguousLocalTimePolicy != nil || profile.LocalTimeGapPolicy != nil
	hasUptimeMembers := profile.NetFlowExportTimeColumnOrdinal != nil || profile.NetFlowExportTimeMode != nil || profile.NetFlowExporterUptimeAtExportColumnOrdinal != nil
	switch profile.Mode {
	case "rfc3339":
		if hasUptimeMembers || profile.Precision != "seconds" && profile.Precision != "milliseconds" && profile.Precision != "microseconds" {
			return conflict("variant_member_conflict")
		}
		if profile.AmbiguousLocalTimePolicy == nil || *profile.AmbiguousLocalTimePolicy != "reject" || profile.LocalTimeGapPolicy == nil || *profile.LocalTimeGapPolicy != "reject" {
			return conflict("variant_member_conflict")
		}
		if profile.Timezone == nil || *profile.Timezone == "UTC" {
			if profile.TimezoneRulesetID != nil {
				return conflict("variant_member_conflict")
			}
			return nil
		}
		zone := *profile.Timezone
		if zone == "" || zone == "Local" || strings.HasPrefix(zone, "/") || strings.Contains(zone, "\\") || strings.Contains(zone, "..") {
			return conflict("variant_member_conflict")
		}
		if profile.TimezoneRulesetID == nil || *profile.TimezoneRulesetID != timestampRulesetID {
			return conflict("variant_member_conflict")
		}
		if _, err := loadPinnedTimezone(zone); err != nil {
			return conflict("variant_member_conflict")
		}
		return nil
	case "epoch_seconds":
		if profile.Precision != "seconds" || hasRFCMembers || hasUptimeMembers {
			return conflict("variant_member_conflict")
		}
		return nil
	case "epoch_milliseconds":
		if profile.Precision != "milliseconds" || hasRFCMembers || hasUptimeMembers {
			return conflict("variant_member_conflict")
		}
		return nil
	case "netflow_sys_uptime_milliseconds":
		if profile.Precision != "milliseconds" || hasRFCMembers || profile.NetFlowExportTimeColumnOrdinal == nil || profile.NetFlowExportTimeMode == nil || profile.NetFlowExporterUptimeAtExportColumnOrdinal == nil {
			return conflict("variant_member_conflict")
		}
		if maxColumns <= 0 || *profile.NetFlowExportTimeColumnOrdinal < 1 || *profile.NetFlowExportTimeColumnOrdinal > maxColumns || *profile.NetFlowExporterUptimeAtExportColumnOrdinal < 1 || *profile.NetFlowExporterUptimeAtExportColumnOrdinal > maxColumns {
			return conflict("variant_member_conflict")
		}
		switch *profile.NetFlowExportTimeMode {
		case "rfc3339", "epoch_seconds", "epoch_milliseconds":
			return nil
		default:
			return conflict("variant_member_conflict")
		}
	default:
		return conflict("variant_member_conflict")
	}
}

// MarshalJSON emits the selected closed-union variant, including the required
// null members for the RFC3339 variant and no members owned by another variant.
func (profile TimestampProfile) MarshalJSON() ([]byte, error) {
	object := map[string]any{
		"schema_id": profile.SchemaID,
		"mode":      profile.Mode,
		"precision": profile.Precision,
	}
	switch profile.Mode {
	case "rfc3339":
		object["timezone"] = profile.Timezone
		object["timezone_ruleset_id"] = profile.TimezoneRulesetID
		object["ambiguous_local_time_policy"] = profile.AmbiguousLocalTimePolicy
		object["local_time_gap_policy"] = profile.LocalTimeGapPolicy
	case "netflow_sys_uptime_milliseconds":
		object["netflow_export_time_column_ordinal"] = profile.NetFlowExportTimeColumnOrdinal
		object["netflow_export_time_mode"] = profile.NetFlowExportTimeMode
		object["netflow_exporter_uptime_at_export_column_ordinal"] = profile.NetFlowExporterUptimeAtExportColumnOrdinal
	}
	return json.Marshal(object)
}

func sourceFieldOrdinal(mappings []FieldMapping, fieldKey string) int {
	for _, mapping := range mappings {
		if mapping.MappingKind == MappingKindSourceColumn && mapping.FieldKey == fieldKey {
			return mapping.SourceColumnOrdinal
		}
	}
	return 0
}

func parseTimestamp(value string, profile TimestampProfile) (time.Time, error) {
	return parseTimestampForRecord(value, profile, nil)
}

func parseTimestampForRecord(value string, profile TimestampProfile, record *CSVRecord) (time.Time, error) {
	switch profile.Mode {
	case "rfc3339":
		return parseExactRFC3339(value, profile)
	case "epoch_seconds":
		return parseEpochTimestamp(value, false)
	case "epoch_milliseconds":
		return parseEpochTimestamp(value, true)
	case "netflow_sys_uptime_milliseconds":
		return parseNetFlowUptimeTimestamp(value, profile, record)
	default:
		return time.Time{}, timestampError("invalid_syntax")
	}
}

func parseEpochTimestamp(value string, milliseconds bool) (time.Time, error) {
	parsed, err := parseUint64(value)
	if err != nil {
		return time.Time{}, timestampError("invalid_syntax")
	}
	if milliseconds {
		if parsed > maxTimestampUnixMillis {
			return time.Time{}, timestampError("out_of_range")
		}
		return time.Unix(int64(parsed/1000), int64(parsed%1000)*int64(time.Millisecond)).UTC(), nil
	}
	if parsed > uint64(maxTimestampUnixSeconds) {
		return time.Time{}, timestampError("out_of_range")
	}
	return time.Unix(int64(parsed), 0).UTC(), nil
}

func parseExactRFC3339(value string, profile TimestampProfile) (time.Time, error) {
	matches := exactRFC3339Pattern.FindStringSubmatch(value)
	if matches == nil {
		return time.Time{}, timestampError("invalid_syntax")
	}
	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	day, _ := strconv.Atoi(matches[3])
	hour, _ := strconv.Atoi(matches[4])
	minute, _ := strconv.Atoi(matches[5])
	second, _ := strconv.Atoi(matches[6])
	fraction := matches[7]
	offsetText := matches[8]
	if year < 1 || month < 1 || month > 12 || day < 1 || day > 31 || hour > 23 || minute > 59 || second > 59 {
		return time.Time{}, timestampError("out_of_range")
	}
	precisionDigits := map[string]int{"seconds": 0, "milliseconds": 3, "microseconds": 6}[profile.Precision]
	if len(fraction) > precisionDigits {
		return time.Time{}, timestampError("precision_exceeded")
	}
	nanoseconds := 0
	if fraction != "" {
		fractionValue, _ := strconv.Atoi(fraction)
		for index := len(fraction); index < 9; index++ {
			fractionValue *= 10
		}
		nanoseconds = fractionValue
	}
	if offsetText != "" {
		offsetSeconds, err := parseTimestampOffset(offsetText)
		if err != nil {
			return time.Time{}, err
		}
		location := time.FixedZone("network-flow-source-offset", offsetSeconds)
		result := time.Date(year, time.Month(month), day, hour, minute, second, nanoseconds, location)
		if !sameWallTime(result, year, month, day, hour, minute, second, nanoseconds) {
			return time.Time{}, timestampError("out_of_range")
		}
		return result.UTC(), nil
	}
	if profile.Timezone == nil {
		return time.Time{}, timestampError("invalid_syntax")
	}
	if *profile.Timezone == "UTC" {
		result := time.Date(year, time.Month(month), day, hour, minute, second, nanoseconds, time.UTC)
		if !sameWallTime(result, year, month, day, hour, minute, second, nanoseconds) {
			return time.Time{}, timestampError("out_of_range")
		}
		return result, nil
	}
	location, err := loadPinnedTimezone(*profile.Timezone)
	if err != nil {
		return time.Time{}, timestampError("invalid_syntax")
	}
	result := time.Date(year, time.Month(month), day, hour, minute, second, nanoseconds, location)
	if !sameWallTime(result, year, month, day, hour, minute, second, nanoseconds) {
		return time.Time{}, timestampError("nonexistent_local_time")
	}
	if localTimeMatchCount(year, month, day, hour, minute, second, nanoseconds, location) != 1 {
		return time.Time{}, timestampError("ambiguous_local_time")
	}
	return result.UTC(), nil
}

func parseTimestampOffset(value string) (int, error) {
	if value == "Z" {
		return 0, nil
	}
	if value == "-00:00" || len(value) != 6 {
		return 0, timestampError("invalid_syntax")
	}
	hours, _ := strconv.Atoi(value[1:3])
	minutes, _ := strconv.Atoi(value[4:6])
	if hours > 23 || minutes > 59 {
		return 0, timestampError("out_of_range")
	}
	offset := hours*3600 + minutes*60
	if value[0] == '-' {
		offset = -offset
	}
	return offset, nil
}

func sameWallTime(value time.Time, year, month, day, hour, minute, second, nanosecond int) bool {
	return value.Year() == year && int(value.Month()) == month && value.Day() == day && value.Hour() == hour && value.Minute() == minute && value.Second() == second && value.Nanosecond() == nanosecond
}

func localTimeMatchCount(year, month, day, hour, minute, second, nanosecond int, location *time.Location) int {
	wallUTC := time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, time.UTC)
	offsets := map[int]struct{}{}
	for hours := -36; hours <= 36; hours += 3 {
		_, offset := wallUTC.Add(time.Duration(hours) * time.Hour).In(location).Zone()
		offsets[offset] = struct{}{}
	}
	matches := 0
	for offset := range offsets {
		candidate := wallUTC.Add(-time.Duration(offset) * time.Second).In(location)
		if sameWallTime(candidate, year, month, day, hour, minute, second, nanosecond) {
			matches++
		}
	}
	return matches
}

func parseNetFlowUptimeTimestamp(value string, profile TimestampProfile, record *CSVRecord) (time.Time, error) {
	if record == nil || profile.NetFlowExportTimeColumnOrdinal == nil || profile.NetFlowExportTimeMode == nil || profile.NetFlowExporterUptimeAtExportColumnOrdinal == nil {
		return time.Time{}, timestampError("sys_uptime_invalid")
	}
	exportOrdinal := *profile.NetFlowExportTimeColumnOrdinal
	uptimeOrdinal := *profile.NetFlowExporterUptimeAtExportColumnOrdinal
	if exportOrdinal < 1 || uptimeOrdinal < 1 || exportOrdinal > len(record.Fields) || uptimeOrdinal > len(record.Fields) {
		return time.Time{}, timestampError("sys_uptime_invalid")
	}
	eventUptime, err := parseUint32TimestampValue(value)
	if err != nil {
		return time.Time{}, err
	}
	exporterUptime, err := parseUint32TimestampValue(record.Fields[uptimeOrdinal-1])
	if err != nil {
		return time.Time{}, err
	}
	if eventUptime > exporterUptime {
		return time.Time{}, timestampError("sys_uptime_wrap_ambiguous")
	}
	exportProfile := materializeTimestampProfile(TimestampProfile{SchemaID: timestampProfileSchemaID, Mode: *profile.NetFlowExportTimeMode})
	if exportProfile.Mode == "rfc3339" {
		exportProfile.Timezone = nil
		exportProfile.TimezoneRulesetID = nil
	}
	exportTime, err := parseTimestampForRecord(record.Fields[exportOrdinal-1], exportProfile, nil)
	if err != nil {
		return time.Time{}, timestampError("sys_uptime_invalid")
	}
	delta := time.Duration(exporterUptime-eventUptime) * time.Millisecond
	result := exportTime.Add(-delta).UTC()
	if result.Year() < 1 || result.Year() > 9999 {
		return time.Time{}, timestampError("out_of_range")
	}
	return result, nil
}

func parseUint32TimestampValue(value string) (uint32, error) {
	parsed, err := parseUint64(value)
	if err != nil || parsed > uint64(^uint32(0)) {
		return 0, timestampError("sys_uptime_invalid")
	}
	return uint32(parsed), nil
}
