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

type parsedCSV struct {
	SourceContentSHA256  string
	SourceColumns        []sourceColumnDescriptor
	Records              []csvRecord
	Diagnostics          []rejectedRowDiagnostic
	DiagnosticsTruncated bool
}

type csvRecord struct {
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

func parseCSVPreview(reader io.Reader, expectedSHA256 string, limits EffectiveLimits) (parsedCSV, error) {
	return parseCSV(reader, expectedSHA256, limits, parseModePreview)
}

func parseCSVApply(reader io.Reader, expectedSHA256 string, limits EffectiveLimits) (parsedCSV, error) {
	return parseCSV(reader, expectedSHA256, limits, parseModeApply)
}

func parseCSV(reader io.Reader, expectedSHA256 string, limits EffectiveLimits, mode parseMode) (parsedCSV, error) {
	sourceBytes, err := io.ReadAll(reader)
	if err != nil {
		return parsedCSV{}, err
	}
	actualSHA256 := sha256Hex(sourceBytes)
	if expectedSHA256 != "" && actualSHA256 != expectedSHA256 {
		return parsedCSV{}, errSourceChanged
	}
	sourceBytes, err = normalizeCSVBytes(sourceBytes)
	if err != nil {
		return parsedCSV{}, err
	}
	if len(sourceBytes) == 0 {
		return parsedCSV{}, &sourceValidationError{Code: "network_flow_csv_empty_file", ReasonCode: "zero_bytes"}
	}
	csvReader := csv.NewReader(bytes.NewReader(sourceBytes))
	csvReader.FieldsPerRecord = -1
	csvReader.ReuseRecord = false
	header, err := csvReader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return parsedCSV{}, &sourceValidationError{Code: "network_flow_csv_empty_file", ReasonCode: "zero_bytes"}
		}
		return parsedCSV{}, csvParseError(err)
	}
	if int64(len(header)) > limits.MaxColumnsPerCSV {
		return parsedCSV{}, &sourceValidationError{Code: "network_flow_resource_limit_exceeded", ReasonCode: "column_limit_exceeded"}
	}
	sourceColumns, err := sourceColumnsFromHeader(header, limits)
	if err != nil {
		return parsedCSV{}, err
	}
	records := []csvRecord{}
	diagnostics := []rejectedRowDiagnostic{}
	sourceRowNumber := int64(1)
	for {
		row, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		sourceRowNumber++
		if err != nil {
			return parsedCSV{}, csvParseError(err)
		}
		if sourceRowNumber-1 > limits.MaxRowsPerCSV {
			return parsedCSV{}, &sourceValidationError{Code: "network_flow_resource_limit_exceeded", ReasonCode: "row_limit_exceeded"}
		}
		record := csvRecord{
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
		return parsedCSV{}, &sourceValidationError{Code: "network_flow_no_data_rows", ReasonCode: "header_only"}
	}
	return parsedCSV{
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
		return nil, &sourceValidationError{Code: "network_flow_invalid_utf8", ReasonCode: "bom_not_at_offset_zero"}
	}
	if !utf8.Valid(sourceBytes) {
		return nil, &sourceValidationError{Code: "network_flow_invalid_utf8", ReasonCode: "invalid_utf8_sequence"}
	}
	return sourceBytes, nil
}

func sourceColumnsFromHeader(header []string, limits EffectiveLimits) ([]sourceColumnDescriptor, error) {
	columns := make([]sourceColumnDescriptor, 0, len(header))
	for index, value := range header {
		if invalidHeaderText(value, limits) {
			return nil, &sourceValidationError{Code: "network_flow_invalid_header", ReasonCode: "invalid_header_text"}
		}
		columns = append(columns, sourceColumnDescriptor{
			SourceColumnOrdinal:           index + 1,
			RawHeaderText:                 value,
			NormalizedHeaderForSuggestion: sourceAliasMatchKey(value),
			RawHeaderSHA256:               sha256Hex([]byte(value)),
			SampleValues:                  []safeSample{},
		})
	}
	return columns, nil
}

func invalidHeaderText(value string, limits EffectiveLimits) bool {
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

func addColumnSamples(columns []sourceColumnDescriptor, row []string) {
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
		return &sourceValidationError{Code: "network_flow_csv_malformed_quote", ReasonCode: "csv_parse_error"}
	}
	return err
}

func validateRows(parsed parsedCSV, mapping approvedMapping, mappingFingerprint string, limits EffectiveLimits) ([]flowRow, []rejectedRowDiagnostic, bool, error) {
	fieldMappings := sourceFieldMappings(mapping)
	accepted := []flowRow{}
	diagnostics := append([]rejectedRowDiagnostic(nil), parsed.Diagnostics...)
	diagnosticsTruncated := parsed.DiagnosticsTruncated
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

func validateRecord(record csvRecord, mapping approvedMapping, mappingFingerprint string, fieldMappings map[string]fieldMapping) (flowRow, []rejectedRowDiagnostic) {
	diagnostics := []rejectedRowDiagnostic{}
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
		case fieldFlowStartUTC:
			flowStart = value.(time.Time)
		case fieldFlowEndUTC:
			flowEnd = value.(time.Time)
		}
	}
	for _, fieldKey := range []string{fieldInputInterface, fieldOutputInterface} {
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
		diagnostics = append(diagnostics, diagnostic(record.SourceRowNumber, sourceColumnOrdinalPtr(fieldMappings[fieldFlowEndUTC].SourceColumnOrdinal), headerHashPtr(mapping, fieldMappings[fieldFlowEndUTC].SourceColumnOrdinal), stringPtr(fieldFlowEndUTC), "network_flow_end_before_start", "cross_field_semantics", ""))
	}
	if len(diagnostics) > 0 {
		return flowRow{}, diagnostics
	}
	unmappedRaw := unmappedRawValues(record, mapping, fieldMappings)
	sourceRowDigest := sourceRowDigest(mapping.ParserProfileID, record.SourceRowNumber, record.Fields)
	normalizedValues := map[string]any{
		fieldFlowStartUTC:         formatTimestamp(values[fieldFlowStartUTC].(time.Time)),
		fieldFlowEndUTC:           formatTimestamp(values[fieldFlowEndUTC].(time.Time)),
		fieldSrcIP:                values[fieldSrcIP],
		fieldDstIP:                values[fieldDstIP],
		fieldSrcPort:              values[fieldSrcPort],
		fieldDstPort:              values[fieldDstPort],
		fieldIPProtocol:           values[fieldIPProtocol],
		fieldBytesCount:           values[fieldBytesCount],
		fieldPacketsCount:         values[fieldPacketsCount],
		fieldExporterID:           nil,
		fieldInputInterface:       values[fieldInputInterface],
		fieldOutputInterface:      values[fieldOutputInterface],
		fieldTCPFlags:             nil,
		fieldApplicationLabel:     nil,
		fieldObservationSourceRef: map[string]any{},
	}
	normalizedDigest := normalizedRowDigest(mappingFingerprint, normalizedValues, unmappedRaw)
	unmappedJSON, _ := json.Marshal(unmappedRaw)
	srcPort := int32(values[fieldSrcPort].(int))
	dstPort := int32(values[fieldDstPort].(int))
	return flowRow{
		SourceRowNumber:           record.SourceRowNumber,
		SourceRowDigestSHA256:     sourceRowDigest,
		NormalizedRowDigestSHA256: normalizedDigest,
		FlowStartUTC:              values[fieldFlowStartUTC].(time.Time),
		FlowEndUTC:                values[fieldFlowEndUTC].(time.Time),
		SrcIP:                     values[fieldSrcIP].(string),
		DstIP:                     values[fieldDstIP].(string),
		SrcPort:                   &srcPort,
		DstPort:                   &dstPort,
		IPProtocol:                int32(values[fieldIPProtocol].(int)),
		BytesCount:                values[fieldBytesCount].(string),
		PacketsCount:              values[fieldPacketsCount].(string),
		InputInterface:            stringPtrFromAny(values[fieldInputInterface]),
		OutputInterface:           stringPtrFromAny(values[fieldOutputInterface]),
		UnmappedRaw:               unmappedJSON,
	}, nil
}

func mappedValue(record csvRecord, mapping approvedMapping, fieldMapping fieldMapping, fieldKey string) (any, *rejectedRowDiagnostic) {
	if fieldMapping.SourceColumnOrdinal <= 0 || fieldMapping.SourceColumnOrdinal > len(record.Fields) {
		diag := diagnostic(record.SourceRowNumber, nil, nil, stringPtr(fieldKey), errorCodeForField(fieldKey), "missing_or_empty", "")
		return nil, &diag
	}
	raw := record.Fields[fieldMapping.SourceColumnOrdinal-1]
	transformed := raw
	if fieldMapping.TransformID == transformTrimASCIISpace {
		transformed = strings.Trim(transformed, " ")
	}
	if transformed == "" {
		if fieldMapping.EmptyValuePolicy == emptyPolicyNull {
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
	case transformTimestampProfile:
		value, err = parseTimestampForRecord(transformed, mapping.TimestampProfile, &record)
	case transformIPLiteral:
		value, err = parseIPLiteral(transformed)
	case transformPortNumber:
		value, err = parseBoundedDecimalInt(transformed, 65535)
	case transformProtocol:
		value, err = parseProtocol(transformed)
	case transformUint64Decimal:
		value, err = parseUint64Decimal(transformed)
	case transformTrimASCIISpace:
		value, err = parseBoundedText256(transformed)
	default:
		err = fmt.Errorf("unsupported transform")
	}
	if err != nil {
		reason := "invalid_syntax"
		if fieldMapping.TransformID == transformTimestampProfile {
			reason = timestampReason(err)
		} else if errors.Is(err, errOutOfRange) {
			reason = "out_of_range"
		}
		diag := diagnostic(record.SourceRowNumber, sourceColumnOrdinalPtr(fieldMapping.SourceColumnOrdinal), headerHashPtr(mapping, fieldMapping.SourceColumnOrdinal), stringPtr(fieldKey), errorCodeForField(fieldKey), reason, raw)
		return nil, &diag
	}
	return value, nil
}

var errOutOfRange = errors.New("out of range")

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
	if utf8.RuneCountInString(value) > 256 || containsForbiddenBoundedTextControl(value) {
		return "", errOutOfRange
	}
	return value, nil
}

func containsForbiddenBoundedTextControl(value string) bool {
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

func sourceFieldMappings(mapping approvedMapping) map[string]fieldMapping {
	result := map[string]fieldMapping{}
	for _, fieldMapping := range mapping.FieldMappings {
		if fieldMapping.MappingKind == mappingKindSourceColumn {
			result[fieldMapping.FieldKey] = fieldMapping
		}
	}
	return result
}

func unmappedRawValues(record csvRecord, mapping approvedMapping, fieldMappings map[string]fieldMapping) map[string]any {
	used := map[int]struct{}{}
	for _, fieldMapping := range fieldMappings {
		used[fieldMapping.SourceColumnOrdinal] = struct{}{}
	}
	if mapping.UnknownColumnPolicy != unknownColumnPolicyPreserve {
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

func fieldCountDiagnostic(record csvRecord, want int) rejectedRowDiagnostic {
	return diagnostic(record.SourceRowNumber, nil, nil, nil, "network_flow_csv_field_count_mismatch", "field_count_mismatch", fmt.Sprintf("%d", record.RawFieldCount-want))
}

func diagnostic(sourceRowNumber int64, sourceColumnOrdinal *int64, rawHeaderSHA256 *string, fieldKey *string, errorCode string, reasonCode string, rawValue string) rejectedRowDiagnostic {
	sample := sampleForValue(rawValue)
	diagnostic := rejectedRowDiagnostic{
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
	diagnostic.DiagnosticID = diagnosticID(sourceRowNumber, sourceColumnOrdinal, rawHeaderSHA256, fieldKey, errorCode, reasonCode)
	return diagnostic
}

func appendDiagnostic(diagnostics []rejectedRowDiagnostic, limits EffectiveLimits, diagnostic rejectedRowDiagnostic) []rejectedRowDiagnostic {
	if limits.MaxRejectedRowDiagnostics >= 0 && int64(len(diagnostics)) >= limits.MaxRejectedRowDiagnostics {
		return diagnostics
	}
	return append(diagnostics, diagnostic)
}

func sampleForValue(value string) safeSample {
	hash := sha256Hex([]byte(value))
	sample := safeSample{RawValueSHA256: &hash}
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

func headerHashPtr(mapping approvedMapping, ordinal int) *string {
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
	case fieldFlowStartUTC, fieldFlowEndUTC:
		return "network_flow_invalid_timestamp"
	case fieldSrcIP, fieldDstIP:
		return "network_flow_invalid_ip"
	case fieldSrcPort, fieldDstPort:
		return "network_flow_invalid_port"
	case fieldIPProtocol:
		return "network_flow_invalid_protocol"
	case fieldBytesCount, fieldPacketsCount:
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
