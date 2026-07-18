package networkflow

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTimestampProfileExactGrammarPrecisionAndZoneTransitions(t *testing.T) {
	t.Parallel()
	assertTimestampProfileExactGrammarPrecisionAndZoneTransitions(t)
}

func TestTimestampProfileClosedJSONVariantsRejectNullMissingAndCrossVariantMembers(t *testing.T) {
	t.Parallel()
	approved := MarshalApprovedMapping(approvedMappingFixture(SourceProfileCiscoSNANetFlowCSV))
	for name, raw := range map[string]string{
		"explicit null precision": strings.Replace(string(approved), `"precision":"microseconds"`, `"precision":null`, 1),
		"missing rfc policy":      strings.Replace(string(approved), `,"local_time_gap_policy":"reject"`, ``, 1),
		"cross variant member":    strings.Replace(string(approved), `"mode":"rfc3339"`, `"mode":"epoch_seconds"`, 1),
		"duplicate member":        strings.Replace(string(approved), `"mode":"rfc3339"`, `"mode":"rfc3339","mode":"rfc3339"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeApprovedMapping([]byte(raw))
			var mappingErr *MappingValidationError
			if !errors.As(err, &mappingErr) || mappingErr.ReasonCode != "variant_member_conflict" {
				t.Fatalf("closed timestamp profile got %T %[1]v", err)
			}
		})
	}
}

func assertTimestampProfileExactGrammarPrecisionAndZoneTransitions(t *testing.T) {
	t.Helper()
	seconds := materializeTimestampProfile(TimestampProfile{SchemaID: timestampProfileSchemaID, Mode: "rfc3339", Precision: "seconds"})
	if got, err := parseTimestamp("2026-07-10T12:00:00+02:30", seconds); err != nil || !got.Equal(time.Date(2026, 7, 10, 9, 30, 0, 0, time.UTC)) {
		t.Fatalf("offset timestamp = %v, %v", got, err)
	}
	for _, value := range []string{
		"2026-07-10t12:00:00Z", "2026-07-10T12:00:00z", "2026-07-10 12:00:00Z",
		"2026-07-10T12:00:60Z", "2026-07-10T24:00:00Z", "2026-02-29T12:00:00Z",
		"2026-07-10T12:00:00-00:00", "2026-07-10T12:00:00.1Z",
	} {
		if _, err := parseTimestamp(value, seconds); err == nil {
			t.Fatalf("invalid timestamp %q was accepted", value)
		}
	}
	milliseconds := materializeTimestampProfile(TimestampProfile{SchemaID: timestampProfileSchemaID, Mode: "rfc3339", Precision: "milliseconds"})
	if _, err := parseTimestamp("2026-07-10T12:00:00.123Z", milliseconds); err != nil {
		t.Fatalf("millisecond timestamp: %v", err)
	}
	if _, err := parseTimestamp("2026-07-10T12:00:00.1234Z", milliseconds); timestampReason(err) != "precision_exceeded" {
		t.Fatalf("precision reason = %q, %v", timestampReason(err), err)
	}

	zone := "America/New_York"
	ruleset := timestampRulesetID
	reject := "reject"
	iana := TimestampProfile{
		SchemaID: timestampProfileSchemaID, Mode: "rfc3339", Precision: "microseconds",
		Timezone: &zone, TimezoneRulesetID: &ruleset,
		AmbiguousLocalTimePolicy: &reject, LocalTimeGapPolicy: &reject,
	}
	if err := validateTimestampProfile(iana, 0); err != nil {
		t.Fatalf("validate IANA profile: %v", err)
	}
	if _, err := parseTimestamp("2026-11-01T01:30:00", iana); timestampReason(err) != "ambiguous_local_time" {
		t.Fatalf("fold reason = %q, %v", timestampReason(err), err)
	}
	if _, err := parseTimestamp("2026-03-08T02:30:00", iana); timestampReason(err) != "nonexistent_local_time" {
		t.Fatalf("gap reason = %q, %v", timestampReason(err), err)
	}
	if got, err := parseTimestamp("2026-07-10T12:00:00", iana); err != nil || !got.Equal(time.Date(2026, 7, 10, 16, 0, 0, 0, time.UTC)) {
		t.Fatalf("IANA local timestamp = %v, %v", got, err)
	}
}

func TestNetFlowSystemUptimeTimestampAndMappingOrdinals(t *testing.T) {
	t.Parallel()
	assertNetFlowSystemUptimeTimestampAndMappingOrdinals(t)
}

func assertNetFlowSystemUptimeTimestampAndMappingOrdinals(t *testing.T) {
	t.Helper()
	exportOrdinal := 3
	exporterUptimeOrdinal := 4
	exportMode := "rfc3339"
	profile := materializeTimestampProfile(TimestampProfile{
		SchemaID: timestampProfileSchemaID, Mode: "netflow_sys_uptime_milliseconds",
		NetFlowExportTimeColumnOrdinal: &exportOrdinal, NetFlowExportTimeMode: &exportMode,
		NetFlowExporterUptimeAtExportColumnOrdinal: &exporterUptimeOrdinal,
	})
	record := CSVRecord{Fields: []string{"900", "800", "2026-07-10T12:00:00Z", "1000"}}
	got, err := parseTimestampForRecord(record.Fields[0], profile, &record)
	if err != nil || !got.Equal(time.Date(2026, 7, 10, 11, 59, 59, 900_000_000, time.UTC)) {
		t.Fatalf("uptime timestamp = %v, %v", got, err)
	}
	if _, err := parseTimestampForRecord("1001", profile, &record); timestampReason(err) != "sys_uptime_wrap_ambiguous" {
		t.Fatalf("wrap reason = %q, %v", timestampReason(err), err)
	}
	if _, err := parseTimestampForRecord("4294967296", profile, &record); timestampReason(err) != "sys_uptime_invalid" {
		t.Fatalf("uint32 reason = %q, %v", timestampReason(err), err)
	}

	mapping := approvedMappingFixture(SourceProfileCiscoSNANetFlowCSV)
	mapping.TimestampProfile = profile
	for len(mapping.SourceColumns) < 4 {
		ordinal := len(mapping.SourceColumns) + 1
		mapping.SourceColumns = append(mapping.SourceColumns, SourceColumnDescriptor{SourceColumnOrdinal: ordinal})
	}
	// The first required field is flow_start_utc at ordinal 1. Reusing it as
	// export time is rejected before row evaluation.
	mapping.TimestampProfile.NetFlowExportTimeColumnOrdinal = intPtr(1)
	err = validateApprovedMapping(mapping)
	var mappingErr *MappingValidationError
	if !errors.As(err, &mappingErr) || mappingErr.ReasonCode != "timestamp_column_reused" {
		t.Fatalf("ordinal reuse = %T %[1]v", err)
	}
}

func intPtr(value int) *int { return &value }
