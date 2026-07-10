package networkflow

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const previewRecordLimit = 50

var unsignedDecimalRE = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

type ParsedCSV struct {
	SourceContentSHA256  string
	SourceColumns        []SourceColumnDescriptor
	Records              []CSVRecord
	Diagnostics          []RejectedRowDiagnostic
	DiagnosticsTruncated bool
}

type CSVRecord struct {
	SourceRowNumber int64
	Fields          []string
	RawFieldCount   int
	FieldCountOK    bool
}

type parseMode int

const (
	parseModePreview parseMode = iota
	parseModeApply
)

func ParseCSVPreview(reader io.Reader, expectedSHA256 string, limits Limits) (ParsedCSV, error) {
	return parseCSV(reader, expectedSHA256, limits.normalized(), parseModePreview)
}

func ParseCSVApply(reader io.Reader, expectedSHA256 string, limits Limits) (ParsedCSV, error) {
	return parseCSV(reader, expectedSHA256, limits.normalized(), parseModeApply)
}

func parseCSV(reader io.Reader, expectedSHA256 string, limits Limits, mode parseMode) (ParsedCSV, error) {
	sourceBytes, err := io.ReadAll(reader)
	if err != nil {
		return ParsedCSV{}, err
	}
	actualSHA256 := sha256Hex(sourceBytes)
	if expectedSHA256 != "" && actualSHA256 != expectedSHA256 {
		return ParsedCSV{}, ErrSourceChanged
	}
	sourceBytes, err = normalizeCSVBytes(sourceBytes)
	if err != nil {
		return ParsedCSV{}, err
	}
	if len(sourceBytes) == 0 {
		return ParsedCSV{}, &SourceValidationError{Code: "network_flow_csv_empty_file", ReasonCode: "zero_bytes"}
	}
	csvReader := csv.NewReader(bytes.NewReader(sourceBytes))
	csvReader.FieldsPerRecord = -1
	csvReader.ReuseRecord = false
	header, err := csvReader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return ParsedCSV{}, &SourceValidationError{Code: "network_flow_csv_empty_file", ReasonCode: "zero_bytes"}
		}
		return ParsedCSV{}, csvParseError(err)
	}
	if int64(len(header)) > limits.MaxColumnsPerCSV {
		return ParsedCSV{}, &SourceValidationError{Code: "network_flow_resource_limit_exceeded", ReasonCode: "column_limit_exceeded"}
	}
	sourceColumns, err := sourceColumnsFromHeader(header, limits)
	if err != nil {
		return ParsedCSV{}, err
	}
	records := []CSVRecord{}
	diagnostics := []RejectedRowDiagnostic{}
	sourceRowNumber := int64(1)
	for {
		row, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		sourceRowNumber++
		if err != nil {
			return ParsedCSV{}, csvParseError(err)
		}
		if sourceRowNumber-1 > limits.MaxRowsPerCSV {
			return ParsedCSV{}, &SourceValidationError{Code: "network_flow_resource_limit_exceeded", ReasonCode: "row_limit_exceeded"}
		}
		record := CSVRecord{
			SourceRowNumber: sourceRowNumber,
			Fields:          append([]string(nil), row...),
			RawFieldCount:   len(row),
			FieldCountOK:    len(row) == len(header),
		}
		if !record.FieldCountOK {
			diagnostics = appendDiagnostic(diagnostics, limits, fieldCountDiagnostic(record, len(header)))
		} else if len(records) < previewRecordLimit {
			addColumnSamples(sourceColumns, row)
		}
		records = append(records, record)
		if mode == parseModePreview && len(records) >= previewRecordLimit {
			break
		}
	}
	if len(records) == 0 {
		return ParsedCSV{}, &SourceValidationError{Code: "network_flow_no_data_rows", ReasonCode: "header_only"}
	}
	return ParsedCSV{
		SourceContentSHA256:  actualSHA256,
		SourceColumns:        sourceColumns,
		Records:              records,
		Diagnostics:          diagnostics,
		DiagnosticsTruncated: int64(len(diagnostics)) >= limits.MaxRejectedRowDiagnostics && limits.MaxRejectedRowDiagnostics >= 0,
	}, nil
}

func normalizeCSVBytes(sourceBytes []byte) ([]byte, error) {
	if len(sourceBytes) >= 3 && sourceBytes[0] == 0xef && sourceBytes[1] == 0xbb && sourceBytes[2] == 0xbf {
		sourceBytes = sourceBytes[3:]
	}
	if bytes.Contains(sourceBytes, []byte{0xef, 0xbb, 0xbf}) {
		return nil, &SourceValidationError{Code: "network_flow_invalid_utf8", ReasonCode: "bom_not_at_offset_zero"}
	}
	if !utf8.Valid(sourceBytes) {
		return nil, &SourceValidationError{Code: "network_flow_invalid_utf8", ReasonCode: "invalid_utf8_sequence"}
	}
	return sourceBytes, nil
}

func sourceColumnsFromHeader(header []string, limits Limits) ([]SourceColumnDescriptor, error) {
	columns := make([]SourceColumnDescriptor, 0, len(header))
	for index, value := range header {
		if invalidHeaderText(value, limits) {
			return nil, &SourceValidationError{Code: "network_flow_invalid_header", ReasonCode: "invalid_header_text"}
		}
		columns = append(columns, SourceColumnDescriptor{
			SourceColumnOrdinal:           index + 1,
			RawHeaderText:                 value,
			NormalizedHeaderForSuggestion: SourceAliasMatchKey(value),
			RawHeaderSHA256:               sha256Hex([]byte(value)),
			SampleValues:                  []SafeSample{},
		})
	}
	return columns, nil
}

func invalidHeaderText(value string, limits Limits) bool {
	if int64(utf8.RuneCountInString(value)) > limits.MaxHeaderScalarLength {
		return true
	}
	for _, r := range value {
		if r == '\t' {
			continue
		}
		if isC0C1Control(r) {
			return true
		}
	}
	return false
}

func addColumnSamples(columns []SourceColumnDescriptor, row []string) {
	for index := range columns {
		value := row[index]
		if value == "" {
			columns[index].DetectedEmptyCount++
		}
		if len(columns[index].SampleValues) >= previewRecordLimit {
			continue
		}
		columns[index].SampleValues = append(columns[index].SampleValues, sampleForValue(value))
	}
}

func csvParseError(err error) error {
	var parseErr *csv.ParseError
	if errors.As(err, &parseErr) {
		return &SourceValidationError{Code: "network_flow_csv_malformed_quote", ReasonCode: "csv_parse_error"}
	}
	return err
}

func ValidateRows(parsed ParsedCSV, mapping ApprovedMapping, mappingFingerprint string, limits Limits) ([]FlowRow, []RejectedRowDiagnostic, bool, error) {
	fieldMappings := sourceFieldMappings(mapping)
	accepted := []FlowRow{}
	diagnostics := append([]RejectedRowDiagnostic(nil), parsed.Diagnostics...)
	diagnosticsTruncated := parsed.DiagnosticsTruncated
	limits = limits.normalized()
	for _, record := range parsed.Records {
		if !record.FieldCountOK {
			continue
		}
		row, rowDiagnostics := validateRecord(record, mapping, mappingFingerprint, fieldMappings)
		if len(rowDiagnostics) > 0 {
			for _, diagnostic := range rowDiagnostics {
				diagnostics = appendDiagnostic(diagnostics, limits, diagnostic)
			}
			if limits.MaxRejectedRowDiagnostics >= 0 && int64(len(diagnostics)) >= limits.MaxRejectedRowDiagnostics {
				diagnosticsTruncated = true
			}
			continue
		}
		accepted = append(accepted, row)
	}
	return accepted, diagnostics, diagnosticsTruncated, nil
}

func validateRecord(record CSVRecord, mapping ApprovedMapping, mappingFingerprint string, fieldMappings map[string]FieldMapping) (FlowRow, []RejectedRowDiagnostic) {
	diagnostics := []RejectedRowDiagnostic{}
	values := map[string]any{}
	var flowStart time.Time
	var flowEnd time.Time
	for _, fieldKey := range requiredCiscoFields() {
		fieldMapping := fieldMappings[fieldKey]
		value, diagnostic := mappedValue(record, mapping, fieldMapping, fieldKey)
		if diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
			continue
		}
		values[fieldKey] = value
		switch fieldKey {
		case FieldFlowStartUTC:
			flowStart = value.(time.Time)
		case FieldFlowEndUTC:
			flowEnd = value.(time.Time)
		}
	}
	for _, fieldKey := range []string{FieldInputInterface, FieldOutputInterface} {
		if fieldMapping, ok := fieldMappings[fieldKey]; ok {
			value, diagnostic := mappedValue(record, mapping, fieldMapping, fieldKey)
			if diagnostic != nil {
				diagnostics = append(diagnostics, *diagnostic)
				continue
			}
			values[fieldKey] = value
		} else {
			values[fieldKey] = nil
		}
	}
	if len(diagnostics) == 0 && flowEnd.Before(flowStart) {
		diagnostics = append(diagnostics, diagnostic(record.SourceRowNumber, sourceColumnOrdinalPtr(fieldMappings[FieldFlowEndUTC].SourceColumnOrdinal), headerHashPtr(mapping, fieldMappings[FieldFlowEndUTC].SourceColumnOrdinal), stringPtr(FieldFlowEndUTC), "network_flow_end_before_start", "cross_field_semantics", ""))
	}
	if len(diagnostics) > 0 {
		return FlowRow{}, diagnostics
	}
	unmappedRaw := unmappedRawValues(record, mapping, fieldMappings)
	sourceRowDigest := SourceRowDigest(mapping.ParserProfileID, record.SourceRowNumber, record.Fields)
	normalizedValues := map[string]any{
		FieldFlowStartUTC:         formatTimestamp(values[FieldFlowStartUTC].(time.Time)),
		FieldFlowEndUTC:           formatTimestamp(values[FieldFlowEndUTC].(time.Time)),
		FieldSrcIP:                values[FieldSrcIP],
		FieldDstIP:                values[FieldDstIP],
		FieldSrcPort:              values[FieldSrcPort],
		FieldDstPort:              values[FieldDstPort],
		FieldIPProtocol:           values[FieldIPProtocol],
		FieldBytesCount:           values[FieldBytesCount],
		FieldPacketsCount:         values[FieldPacketsCount],
		FieldExporterID:           nil,
		FieldInputInterface:       values[FieldInputInterface],
		FieldOutputInterface:      values[FieldOutputInterface],
		FieldTCPFlags:             nil,
		FieldApplicationLabel:     nil,
		FieldObservationSourceRef: map[string]any{},
	}
	normalizedDigest := NormalizedRowDigest(mappingFingerprint, normalizedValues, unmappedRaw)
	unmappedJSON, _ := json.Marshal(unmappedRaw)
	srcPort := int32(values[FieldSrcPort].(int))
	dstPort := int32(values[FieldDstPort].(int))
	return FlowRow{
		SourceRowNumber:           record.SourceRowNumber,
		SourceRowDigestSHA256:     sourceRowDigest,
		NormalizedRowDigestSHA256: normalizedDigest,
		FlowStartUTC:              values[FieldFlowStartUTC].(time.Time),
		FlowEndUTC:                values[FieldFlowEndUTC].(time.Time),
		SrcIP:                     values[FieldSrcIP].(string),
		DstIP:                     values[FieldDstIP].(string),
		SrcPort:                   &srcPort,
		DstPort:                   &dstPort,
		IPProtocol:                int32(values[FieldIPProtocol].(int)),
		BytesCount:                values[FieldBytesCount].(string),
		PacketsCount:              values[FieldPacketsCount].(string),
		InputInterface:            stringPtrFromAny(values[FieldInputInterface]),
		OutputInterface:           stringPtrFromAny(values[FieldOutputInterface]),
		UnmappedRaw:               unmappedJSON,
	}, nil
}

func mappedValue(record CSVRecord, mapping ApprovedMapping, fieldMapping FieldMapping, fieldKey string) (any, *RejectedRowDiagnostic) {
	if fieldMapping.SourceColumnOrdinal <= 0 || fieldMapping.SourceColumnOrdinal > len(record.Fields) {
		diag := diagnostic(record.SourceRowNumber, nil, nil, stringPtr(fieldKey), errorCodeForField(fieldKey), "missing_or_empty", "")
		return nil, &diag
	}
	raw := record.Fields[fieldMapping.SourceColumnOrdinal-1]
	transformed := raw
	if fieldMapping.TransformID == TransformTrimASCIISpace {
		transformed = strings.Trim(transformed, " ")
	}
	if transformed == "" {
		if fieldMapping.EmptyValuePolicy == EmptyPolicyNull {
			return nil, nil
		}
		diag := diagnostic(record.SourceRowNumber, sourceColumnOrdinalPtr(fieldMapping.SourceColumnOrdinal), headerHashPtr(mapping, fieldMapping.SourceColumnOrdinal), stringPtr(fieldKey), errorCodeForField(fieldKey), "missing_or_empty", raw)
		return nil, &diag
	}
	var (
		value any
		err   error
	)
	switch fieldMapping.TransformID {
	case TransformTimestampProfile:
		value, err = parseTimestamp(transformed, mapping.TimestampProfile)
	case TransformIPLiteral:
		value, err = parseIPLiteral(transformed)
	case TransformPortNumber:
		value, err = parseBoundedDecimalInt(transformed, 65535)
	case TransformProtocol:
		value, err = parseProtocol(transformed)
	case TransformUint64Decimal:
		value, err = parseUint64Decimal(transformed)
	case TransformTrimASCIISpace:
		value, err = parseBoundedText256(transformed)
	default:
		err = fmt.Errorf("unsupported transform")
	}
	if err != nil {
		reason := "invalid_syntax"
		if errors.Is(err, errOutOfRange) {
			reason = "out_of_range"
		}
		diag := diagnostic(record.SourceRowNumber, sourceColumnOrdinalPtr(fieldMapping.SourceColumnOrdinal), headerHashPtr(mapping, fieldMapping.SourceColumnOrdinal), stringPtr(fieldKey), errorCodeForField(fieldKey), reason, raw)
		return nil, &diag
	}
	return value, nil
}

var errOutOfRange = errors.New("out of range")

func parseTimestamp(value string, profile TimestampProfile) (time.Time, error) {
	switch profile.Mode {
	case "epoch_seconds":
		seconds, err := parseUint64(value)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(int64(seconds), 0).UTC(), nil
	case "epoch_milliseconds":
		millis, err := parseUint64(value)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(int64(millis/1000), int64(millis%1000)*int64(time.Millisecond)).UTC(), nil
	default:
		if strings.Contains(value, "t") || strings.Contains(value, "z") || strings.Contains(value, " ") || strings.Contains(value, ",") {
			return time.Time{}, fmt.Errorf("invalid timestamp")
		}
		parsed, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return time.Time{}, err
		}
		if parsed.Location() != time.UTC && !strings.HasSuffix(value, "Z") && !strings.Contains(value[len("2006-01-02T15:04:05"):], "+") && !strings.Contains(value[len("2006-01-02T15:04:05"):], "-") {
			return time.Time{}, fmt.Errorf("invalid timestamp")
		}
		if strings.Contains(value, "-00:00") {
			return time.Time{}, fmt.Errorf("invalid timestamp")
		}
		return parsed.UTC(), nil
	}
}

func parseIPLiteral(value string) (string, error) {
	if strings.Contains(value, "%") {
		return "", fmt.Errorf("zone not allowed")
	}
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return "", err
	}
	if addr.Is4() {
		parts := strings.Split(value, ".")
		if len(parts) != 4 {
			return "", fmt.Errorf("invalid ipv4")
		}
		for _, part := range parts {
			if len(part) > 1 && strings.HasPrefix(part, "0") {
				return "", fmt.Errorf("leading zero")
			}
		}
	}
	return addr.String(), nil
}

func parseBoundedDecimalInt(value string, max uint64) (int, error) {
	parsed, err := parseUint64(value)
	if err != nil {
		return 0, err
	}
	if parsed > max {
		return 0, errOutOfRange
	}
	return int(parsed), nil
}

func parseProtocol(value string) (int, error) {
	trimmed := strings.Trim(value, " ")
	upper := strings.ToUpper(trimmed)
	switch upper {
	case "TCP":
		return 6, nil
	case "UDP":
		return 17, nil
	case "ICMP":
		return 1, nil
	case "ICMPV6", "ICMP_V6":
		return 58, nil
	case "GRE":
		return 47, nil
	case "ESP":
		return 50, nil
	case "AH":
		return 51, nil
	}
	return parseBoundedDecimalInt(trimmed, 255)
}

func parseUint64Decimal(value string) (string, error) {
	parsed, err := parseUint64(value)
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(parsed, 10), nil
}

func parseUint64(value string) (uint64, error) {
	if !unsignedDecimalRE.MatchString(value) {
		return 0, fmt.Errorf("invalid decimal")
	}
	parsed, ok := new(big.Int).SetString(value, 10)
	if !ok || parsed.BitLen() > 64 {
		return 0, errOutOfRange
	}
	return parsed.Uint64(), nil
}

func parseBoundedText256(value string) (string, error) {
	if utf8.RuneCountInString(value) > 256 || containsC0C1Control(value) {
		return "", errOutOfRange
	}
	return value, nil
}

func sourceFieldMappings(mapping ApprovedMapping) map[string]FieldMapping {
	result := map[string]FieldMapping{}
	for _, fieldMapping := range mapping.FieldMappings {
		if fieldMapping.MappingKind == MappingKindSourceColumn {
			result[fieldMapping.FieldKey] = fieldMapping
		}
	}
	return result
}

func unmappedRawValues(record CSVRecord, mapping ApprovedMapping, fieldMappings map[string]FieldMapping) map[string]any {
	used := map[int]struct{}{}
	for _, fieldMapping := range fieldMappings {
		used[fieldMapping.SourceColumnOrdinal] = struct{}{}
	}
	if mapping.UnknownColumnPolicy != UnknownColumnPolicyPreserve {
		return map[string]any{}
	}
	result := map[string]any{}
	for _, column := range mapping.SourceColumns {
		if _, ok := used[column.SourceColumnOrdinal]; ok {
			continue
		}
		if column.SourceColumnOrdinal <= 0 || column.SourceColumnOrdinal > len(record.Fields) {
			continue
		}
		value := record.Fields[column.SourceColumnOrdinal-1]
		result[strconv.Itoa(column.SourceColumnOrdinal)] = map[string]any{
			"source_column_ordinal": column.SourceColumnOrdinal,
			"raw_header_text":       column.RawHeaderText,
			"raw_header_sha256":     column.RawHeaderSHA256,
			"decoded_value":         value,
			"decoded_value_sha256":  sha256Hex([]byte(value)),
		}
	}
	return result
}

func fieldCountDiagnostic(record CSVRecord, want int) RejectedRowDiagnostic {
	return diagnostic(record.SourceRowNumber, nil, nil, nil, "network_flow_csv_field_count_mismatch", "field_count_mismatch", fmt.Sprintf("%d", record.RawFieldCount-want))
}

func diagnostic(sourceRowNumber int64, sourceColumnOrdinal *int64, rawHeaderSHA256 *string, fieldKey *string, errorCode string, reasonCode string, rawValue string) RejectedRowDiagnostic {
	sample := sampleForValue(rawValue)
	diagnostic := RejectedRowDiagnostic{
		SourceRowNumber:     sourceRowNumber,
		SourceColumnOrdinal: sourceColumnOrdinal,
		RawHeaderSHA256:     rawHeaderSHA256,
		FieldKey:            fieldKey,
		ErrorCode:           errorCode,
		ReasonCode:          reasonCode,
		SafeSample:          sample.SafeSample,
		RawValueSHA256:      sample.RawValueSHA256,
		MessageKey:          "network_flow.diagnostic." + strings.TrimPrefix(errorCode, "network_flow_") + "." + reasonCode,
		MessageArgs:         json.RawMessage(`{}`),
		Message:             errorCode + ": " + reasonCode,
	}
	diagnostic.DiagnosticID = DiagnosticID(sourceRowNumber, sourceColumnOrdinal, rawHeaderSHA256, fieldKey, errorCode, reasonCode)
	return diagnostic
}

func appendDiagnostic(diagnostics []RejectedRowDiagnostic, limits Limits, diagnostic RejectedRowDiagnostic) []RejectedRowDiagnostic {
	if limits.MaxRejectedRowDiagnostics >= 0 && int64(len(diagnostics)) >= limits.MaxRejectedRowDiagnostics {
		return diagnostics
	}
	return append(diagnostics, diagnostic)
}

func sampleForValue(value string) SafeSample {
	hash := sha256Hex([]byte(value))
	sample := SafeSample{RawValueSHA256: &hash}
	if unsignedDecimalRE.MatchString(value) && len(value) <= 32 {
		v := value
		sample.SafeSample = &v
	}
	if value == "" {
		empty := ""
		sample.SafeSample = &empty
	}
	return sample
}

func sourceColumnOrdinalPtr(value int) *int64 {
	v := int64(value)
	return &v
}

func headerHashPtr(mapping ApprovedMapping, ordinal int) *string {
	if ordinal <= 0 || ordinal > len(mapping.SourceColumns) {
		return nil
	}
	value := mapping.SourceColumns[ordinal-1].RawHeaderSHA256
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func stringPtrFromAny(value any) *string {
	if value == nil {
		return nil
	}
	text, ok := value.(string)
	if !ok {
		return nil
	}
	return &text
}

func errorCodeForField(fieldKey string) string {
	switch fieldKey {
	case FieldFlowStartUTC, FieldFlowEndUTC:
		return "network_flow_invalid_timestamp"
	case FieldSrcIP, FieldDstIP:
		return "network_flow_invalid_ip"
	case FieldSrcPort, FieldDstPort:
		return "network_flow_invalid_port"
	case FieldIPProtocol:
		return "network_flow_invalid_protocol"
	case FieldBytesCount, FieldPacketsCount:
		return "network_flow_invalid_counter"
	default:
		return "network_flow_invalid_request"
	}
}

func formatTimestamp(value time.Time) string {
	value = value.UTC()
	if value.Nanosecond() == 0 {
		return value.Format(time.RFC3339)
	}
	return strings.TrimRight(strings.TrimRight(value.Format("2006-01-02T15:04:05.999999Z"), "0"), ".")
}
