package networkflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/netip"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const previewRecordLimit = 50

var unsignedDecimalRE = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

type parsedCSV struct {
	Mode                parseMode
	SourceContentSHA256 string
	SourceColumns       []sourceColumnDescriptor
	Records             []csvRecord
}

type csvRecord struct {
	SourceRowNumber int64
	Fields          []string
	Cells           []csvCell
	RawFieldCount   int
	FieldCountOK    bool
}

type parseMode int

const (
	parseModePreview parseMode = iota
	parseModeApply
)

func parseCSVPreview(reader io.Reader, expectedSHA256 string, limits EffectiveLimits) (parsedCSV, error) {
	decoder, err := newCSVDecoder(reader, expectedSHA256, limits, parseModePreview)
	if err != nil {
		return parsedCSV{}, err
	}
	records := []csvRecord{}
	for {
		record, err := decoder.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return parsedCSV{}, err
		}
		records = append(records, record)
	}
	return parsedCSV{Mode: parseModePreview, SourceContentSHA256: decoder.SourceSHA256, SourceColumns: decoder.Columns, Records: records}, nil
}

func sourceColumnsFromHeader(header []string, limits EffectiveLimits) ([]sourceColumnDescriptor, error) {
	columns := make([]sourceColumnDescriptor, 0, len(header))
	for index, value := range header {
		if invalidHeaderText(value, limits) {
			return nil, &sourceValidationError{Code: "network_flow_invalid_header", ReasonCode: "forbidden_header_control"}
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

func addColumnSamples(columns []sourceColumnDescriptor, record csvRecord) {
	for index := range columns {
		cell := record.Cells[index]
		if cell.ScalarCount == 0 {
			columns[index].DetectedEmptyCount++
		}
		if len(columns[index].SampleValues) >= previewRecordLimit {
			continue
		}
		sample := safeSample{RawValueSHA256: stringPtr(cell.SHA256)}
		if !cell.OverLimit && safeNumericSample(record.Fields[index]) {
			sample.SafeSample = stringPtr(record.Fields[index])
		}
		columns[index].SampleValues = append(columns[index].SampleValues, sample)
	}
}

type rowValidationResult struct {
	Rows                 []flowRow
	RejectedRowCount     int64
	Diagnostics          []rejectedRowDiagnostic
	DiagnosticsTruncated bool
	DiagnosticCount      int
	DiagnosticSample     []rejectedRowDiagnostic
}

func validateRows(parsed parsedCSV, mapping approvedMapping, fingerprint string, limits EffectiveLimits) (rowValidationResult, error) {
	result := rowValidationResult{Rows: []flowRow{}, Diagnostics: []rejectedRowDiagnostic{}, DiagnosticSample: []rejectedRowDiagnostic{}}
	fieldMappings := sourceFieldMappings(mapping)
	for _, record := range parsed.Records {
		row, diagnostics := validateCSVRecord(record, parsed.SourceColumns, mapping, fingerprint, parsed.SourceContentSHA256, limits, fieldMappings)
		if len(diagnostics) == 0 {
			result.Rows = append(result.Rows, row)
			continue
		}
		result.RejectedRowCount++
		for _, d := range diagnostics {
			result.DiagnosticCount++
			if len(result.DiagnosticSample) < previewRecordLimit {
				result.DiagnosticSample = append(result.DiagnosticSample, d)
			}
			// Preview is bounded by 50 records and the closed target field set. Apply
			// retains the first N in owner order and records actual omissions only.
			if parsed.Mode == parseModePreview || int64(len(result.Diagnostics)) < limits.MaxRejectedRowDiagnostics {
				result.Diagnostics = append(result.Diagnostics, d)
			} else {
				result.DiagnosticsTruncated = true
			}
		}
	}
	return result, nil
}

func validateCSVRecord(record csvRecord, columns []sourceColumnDescriptor, mapping approvedMapping, fingerprint, sourceSHA string, limits EffectiveLimits, fieldMappings map[string]fieldMapping) (flowRow, []rejectedRowDiagnostic) {
	var row flowRow
	var diagnostics []rejectedRowDiagnostic
	if !record.FieldCountOK {
		diagnostics = []rejectedRowDiagnostic{fieldCountDiagnostic(record, len(columns))}
	} else {
		row, diagnostics = validateRecord(record, mapping, fingerprint, fieldMappings, limits)
	}
	for i := range diagnostics {
		d := &diagnostics[i]
		d.DiagnosticID = diagnosticID(sourceSHA, fingerprint, d.SourceRowNumber, d.SourceColumnOrdinal, d.FieldKey, d.ErrorCode, d.ReasonCode)
	}
	slices.SortFunc(diagnostics, compareDiagnostics)
	return row, diagnostics
}

func validateRecord(record csvRecord, mapping approvedMapping, mappingFingerprint string, fieldMappings map[string]fieldMapping, limits EffectiveLimits) (flowRow, []rejectedRowDiagnostic) {
	diagnostics := []rejectedRowDiagnostic{}
	for index, cell := range record.Cells {
		if !cell.OverLimit {
			continue
		}
		var fieldKey *string
		for _, mapped := range fieldMappings {
			if mapped.SourceColumnOrdinal == index+1 {
				fieldKey = stringPtr(mapped.FieldKey)
				break
			}
		}
		d := diagnostic(record.SourceRowNumber, sourceColumnOrdinalPtr(index+1), headerHashPtr(mapping, index+1), fieldKey, "network_flow_resource_limit_exceeded", "cell_scalar_limit_exceeded", "")
		d.SafeSample = nil
		d.RawValueSHA256 = stringPtr(cell.SHA256)
		d.LimitName = stringPtr("network_flow.max_raw_cell_scalar_length")
		limit, actual := limits.MaxRawCellScalarLength, cell.ScalarCount
		d.LimitValue, d.ActualValue = &limit, &actual
		diagnostics = append(diagnostics, d)
	}
	values := map[string]any{}
	var flowStart time.Time
	var flowEnd time.Time
	for _, fieldKey := range requiredCiscoFields() {
		fieldMapping := fieldMappings[fieldKey]
		if cellValidationBlocked(record, mapping, fieldMapping, fieldKey) {
			continue
		}
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
			if cellValidationBlocked(record, mapping, fieldMapping, fieldKey) {
				continue
			}
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
	_, startValid := values[fieldFlowStartUTC]
	_, endValid := values[fieldFlowEndUTC]
	if startValid && endValid && flowEnd.Before(flowStart) {
		diagnostics = append(diagnostics, diagnostic(record.SourceRowNumber, sourceColumnOrdinalPtr(fieldMappings[fieldFlowEndUTC].SourceColumnOrdinal), headerHashPtr(mapping, fieldMappings[fieldFlowEndUTC].SourceColumnOrdinal), stringPtr(fieldFlowEndUTC), "network_flow_end_before_start", "end_before_start", record.Fields[fieldMappings[fieldFlowEndUTC].SourceColumnOrdinal-1]))
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

func cellValidationBlocked(record csvRecord, mapping approvedMapping, field fieldMapping, key string) bool {
	oversized := func(ordinal int) bool {
		return ordinal > 0 && ordinal <= len(record.Cells) && record.Cells[ordinal-1].OverLimit
	}
	if oversized(field.SourceColumnOrdinal) {
		return true
	}
	if (key == fieldFlowStartUTC || key == fieldFlowEndUTC) && mapping.TimestampProfile.Mode == "netflow_sys_uptime_milliseconds" {
		profile := mapping.TimestampProfile
		return profile.NetFlowExportTimeColumnOrdinal != nil && oversized(*profile.NetFlowExportTimeColumnOrdinal) || profile.NetFlowExporterUptimeAtExportColumnOrdinal != nil && oversized(*profile.NetFlowExporterUptimeAtExportColumnOrdinal)
	}
	return false
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
		if fieldMapping.EmptyValuePolicy == emptyPolicyNull && (fieldKey == fieldInputInterface || fieldKey == fieldOutputInterface) {
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
	return diagnostic(record.SourceRowNumber, nil, nil, nil, "network_flow_csv_field_count_mismatch", "field_count_mismatch", "")
}

func diagnostic(sourceRowNumber int64, sourceColumnOrdinal *int64, rawHeaderSHA256 *string, fieldKey *string, errorCode string, reasonCode string, rawValue string) rejectedRowDiagnostic {
	sample := sampleForValue(rawValue)
	if sourceColumnOrdinal == nil {
		sample = safeSample{}
	} else if fieldKey == nil || !numericDiagnosticField(*fieldKey) {
		sample.SafeSample = nil
	}
	diagnostic := rejectedRowDiagnostic{
		SourceRowNumber:     sourceRowNumber,
		SourceColumnOrdinal: sourceColumnOrdinal,
		RawHeaderSHA256:     rawHeaderSHA256,
		FieldKey:            fieldKey,
		ErrorCode:           errorCode,
		ReasonCode:          reasonCode,
		SafeSample:          sample.SafeSample,
		RawValueSHA256:      sample.RawValueSHA256,
		MessageKey:          "network_flow.diagnostic." + errorCode + "." + reasonCode,
		MessageArgs:         json.RawMessage(`{}`),
		Message:             "network_flow.diagnostic." + errorCode + "." + reasonCode,
	}
	return diagnostic
}

func sampleForValue(value string) safeSample {
	hash := sha256Hex([]byte(value))
	sample := safeSample{RawValueSHA256: &hash}
	if safeNumericSample(value) {
		v := value
		sample.SafeSample = &v
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
	case fieldInputInterface, fieldOutputInterface:
		return "network_flow_invalid_text"
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
