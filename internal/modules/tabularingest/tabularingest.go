package tabularingest

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	MaxClipboardRows       = 500
	MaxClipboardCols       = 64
	TabularRowPlanV1Kind   = "tabular_row_plan_v1"
	SourceFormatCSV        = "csv"
	SourceFormatTSV        = "tsv"
	WarningUnmappedValueV1 = "unmapped_source_value"
)

type SourceColumnV1 struct {
	Ordinal    int
	HeaderText any
}

type FieldMappingV1 struct {
	SourceColumnOrdinal int
	FieldKey            string
	EntityBindingMode   *string
}

type UnmappedRawValueV1 struct {
	SourceKind          string
	SourceClientTxnID   string
	SourceRowOrdinal    int
	SourceColumnOrdinal int
	SourceHeaderText    any
	RawValue            string
}

type CellPlanV1 struct {
	FieldKey            string
	RawValue            string
	SourceColumnOrdinal int
	EntityBindingMode   *string
}

type RowPlanV1 struct {
	RowOrdinal int
	Cells      []CellPlanV1
	Unmapped   []UnmappedRawValueV1
}

type WarningV1 struct {
	Code                string
	SourceRowOrdinal    int
	SourceColumnOrdinal int
}

type TabularRowPlanV1 struct {
	Kind               string
	ViewSchemaID       string
	ClientTxnID        string
	SourceKind         string
	SourceFormat       string
	SourceColumns      []SourceColumnV1
	FieldMappings      []FieldMappingV1
	Rows               []RowPlanV1
	Warnings           []WarningV1
	MappingFingerprint string
}

type MappingRequest struct {
	ViewSchemaID         string
	ClientTxnID          string
	SourceKind           string
	Text                 string
	Format               string
	StartFieldKey        string
	Columns              []string
	SourceHeaders        []any
	ExactHeaderLabels    []string
	ExactHeaderFieldKeys []string
	RequireTargets       int
}

type Dimensions struct {
	RowCount    int
	ColumnCount int
}

func ParseTable(text string, format string) ([][]string, error) {
	return ParseTableWithMaxColumns(text, format, MaxClipboardCols)
}

func ParseTableWithMaxColumns(text string, format string, maxColumns int) ([][]string, error) {
	normalized := strings.TrimRight(text, "\r\n")
	if normalized == "" {
		return nil, fmt.Errorf("empty tabular payload")
	}
	delimiter := ','
	if format == "tsv" || (format == "auto" && strings.Contains(normalized, "\t")) {
		delimiter = '\t'
	}
	reader := csv.NewReader(bytes.NewBufferString(normalized))
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = false
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse tabular payload: %w", err)
	}
	if maxColumns > 0 {
		for _, row := range rows {
			if len(row) > maxColumns {
				return nil, fmt.Errorf("tabular column count exceeded")
			}
		}
	}
	return rows, nil
}

func DimensionsForText(text string, format string) (Dimensions, error) {
	rows, err := ParseTable(text, format)
	if err != nil {
		return Dimensions{}, err
	}
	maxCols := 1
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	return Dimensions{RowCount: len(rows), ColumnCount: maxCols}, nil
}

func BuildTabularRowPlanV1(request MappingRequest) (TabularRowPlanV1, error) {
	if request.SourceKind == "" {
		return TabularRowPlanV1{}, fmt.Errorf("missing source kind")
	}
	schema, ok := viewschema.Lookup(request.ViewSchemaID)
	if !ok {
		return TabularRowPlanV1{}, fmt.Errorf("unknown view schema")
	}
	rows, err := ParseTable(request.Text, request.Format)
	if err != nil {
		return TabularRowPlanV1{}, err
	}
	sourceFormat, err := normalizedSourceFormat(request.Text, request.Format)
	if err != nil {
		return TabularRowPlanV1{}, err
	}

	targetFieldKeys := request.Columns
	sourceHeaders := request.SourceHeaders
	exactHeader := exactHeaderMatches(rows, request.ExactHeaderLabels, request.ExactHeaderFieldKeys)
	if exactHeader {
		sourceHeaders = stringHeaders(request.ExactHeaderLabels)
		targetFieldKeys = append([]string(nil), request.ExactHeaderFieldKeys...)
		rows = rows[1:]
		request.StartFieldKey = targetFieldKeys[0]
		for rowIndex := range rows {
			if missing := len(targetFieldKeys) - len(rows[rowIndex]); missing > 0 {
				rows[rowIndex] = append(rows[rowIndex], make([]string, missing)...)
			}
		}
	}
	if len(rows) == 0 || len(rows) > MaxClipboardRows {
		return TabularRowPlanV1{}, fmt.Errorf("invalid tabular row count")
	}
	if request.RequireTargets >= 0 && request.RequireTargets != len(rows) {
		return TabularRowPlanV1{}, fmt.Errorf("target count must equal tabular row count")
	}
	if len(targetFieldKeys) == 0 || len(targetFieldKeys) > MaxClipboardCols {
		return TabularRowPlanV1{}, fmt.Errorf("invalid target column count")
	}
	columnStart := slices.Index(targetFieldKeys, request.StartFieldKey)
	if columnStart < 0 {
		return TabularRowPlanV1{}, fmt.Errorf("start field not present in target columns")
	}
	fields := schema.Fields()
	for _, fieldKey := range targetFieldKeys {
		if _, ok := fields[fieldKey]; !ok {
			return TabularRowPlanV1{}, fmt.Errorf("unsupported field key %s", fieldKey)
		}
	}

	maxSourceColumns := 0
	for _, values := range rows {
		maxSourceColumns = max(maxSourceColumns, len(values))
	}
	plan := TabularRowPlanV1{
		Kind:          TabularRowPlanV1Kind,
		ViewSchemaID:  request.ViewSchemaID,
		ClientTxnID:   request.ClientTxnID,
		SourceKind:    request.SourceKind,
		SourceFormat:  sourceFormat,
		SourceColumns: make([]SourceColumnV1, 0, maxSourceColumns),
		FieldMappings: make([]FieldMappingV1, 0, min(maxSourceColumns, len(targetFieldKeys)-columnStart)),
		Rows:          make([]RowPlanV1, 0, len(rows)),
		Warnings:      []WarningV1{},
	}
	mappings := make(map[int]FieldMappingV1, maxSourceColumns)
	for sourceColumnIndex := 0; sourceColumnIndex < maxSourceColumns; sourceColumnIndex++ {
		ordinal := sourceColumnIndex + 1
		var header any
		if sourceColumnIndex < len(sourceHeaders) {
			header = sourceHeaders[sourceColumnIndex]
		}
		plan.SourceColumns = append(plan.SourceColumns, SourceColumnV1{Ordinal: ordinal, HeaderText: header})
		targetColumnIndex := columnStart + sourceColumnIndex
		if targetColumnIndex >= len(targetFieldKeys) {
			continue
		}
		fieldKey := targetFieldKeys[targetColumnIndex]
		field := fields[fieldKey]
		if !field.Writable {
			continue
		}
		mapping := FieldMappingV1{
			SourceColumnOrdinal: ordinal,
			FieldKey:            fieldKey,
			EntityBindingMode:   field.EntityBindingMode,
		}
		mappings[ordinal] = mapping
		plan.FieldMappings = append(plan.FieldMappings, mapping)
	}

	for rowIndex, values := range rows {
		rowPlan := RowPlanV1{RowOrdinal: rowIndex + 1}
		seenFields := make(map[string]struct{})
		for sourceColumnIndex, rawValue := range values {
			sourceColumnOrdinal := sourceColumnIndex + 1
			mapping, mapped := mappings[sourceColumnOrdinal]
			if !mapped {
				rowPlan.Unmapped = append(rowPlan.Unmapped, unmappedRawValue(request, sourceHeaders, rowIndex+1, sourceColumnOrdinal, rawValue))
				plan.Warnings = append(plan.Warnings, WarningV1{
					Code:                WarningUnmappedValueV1,
					SourceRowOrdinal:    rowIndex + 1,
					SourceColumnOrdinal: sourceColumnOrdinal,
				})
				continue
			}
			if _, duplicate := seenFields[mapping.FieldKey]; duplicate {
				return TabularRowPlanV1{}, fmt.Errorf("duplicate field key %s", mapping.FieldKey)
			}
			seenFields[mapping.FieldKey] = struct{}{}
			rowPlan.Cells = append(rowPlan.Cells, CellPlanV1{
				FieldKey:            mapping.FieldKey,
				RawValue:            rawValue,
				SourceColumnOrdinal: sourceColumnOrdinal,
				EntityBindingMode:   mapping.EntityBindingMode,
			})
		}
		plan.Rows = append(plan.Rows, rowPlan)
	}
	plan.MappingFingerprint, err = mappingFingerprint(plan)
	if err != nil {
		return TabularRowPlanV1{}, err
	}
	if err := plan.Validate(); err != nil {
		return TabularRowPlanV1{}, err
	}
	return plan, nil
}

func (plan TabularRowPlanV1) Validate() error {
	if plan.Kind != TabularRowPlanV1Kind {
		return fmt.Errorf("unsupported tabular row plan kind %q", plan.Kind)
	}
	if strings.TrimSpace(plan.ClientTxnID) == "" || strings.TrimSpace(plan.SourceKind) == "" {
		return fmt.Errorf("tabular row plan identity is incomplete")
	}
	if plan.SourceFormat != SourceFormatCSV && plan.SourceFormat != SourceFormatTSV {
		return fmt.Errorf("unsupported tabular row plan source format %q", plan.SourceFormat)
	}
	schema, ok := viewschema.Lookup(plan.ViewSchemaID)
	if !ok {
		return fmt.Errorf("unknown tabular row plan view schema")
	}
	if len(plan.Rows) == 0 || len(plan.Rows) > MaxClipboardRows || len(plan.SourceColumns) == 0 || len(plan.SourceColumns) > MaxClipboardCols {
		return fmt.Errorf("invalid tabular row plan dimensions")
	}

	sourceColumns := make(map[int]SourceColumnV1, len(plan.SourceColumns))
	for index, column := range plan.SourceColumns {
		if column.Ordinal != index+1 {
			return fmt.Errorf("tabular source columns are not contiguous")
		}
		sourceColumns[column.Ordinal] = column
	}
	fields := schema.Fields()
	mappings := make(map[int]FieldMappingV1, len(plan.FieldMappings))
	mappedFields := make(map[string]struct{}, len(plan.FieldMappings))
	for _, mapping := range plan.FieldMappings {
		field, exists := fields[mapping.FieldKey]
		if !exists || !field.Writable {
			return fmt.Errorf("invalid tabular field mapping %q", mapping.FieldKey)
		}
		if _, exists := sourceColumns[mapping.SourceColumnOrdinal]; !exists {
			return fmt.Errorf("tabular field mapping references an unknown source column")
		}
		if _, duplicate := mappings[mapping.SourceColumnOrdinal]; duplicate {
			return fmt.Errorf("duplicate tabular source-column mapping")
		}
		if _, duplicate := mappedFields[mapping.FieldKey]; duplicate {
			return fmt.Errorf("duplicate tabular target-field mapping")
		}
		if !equalOptionalString(mapping.EntityBindingMode, field.EntityBindingMode) {
			return fmt.Errorf("tabular entity binding mode drift for %q", mapping.FieldKey)
		}
		mappings[mapping.SourceColumnOrdinal] = mapping
		mappedFields[mapping.FieldKey] = struct{}{}
	}

	warnings := make(map[[2]int]string, len(plan.Warnings))
	for _, warning := range plan.Warnings {
		if warning.Code != WarningUnmappedValueV1 {
			return fmt.Errorf("unsupported tabular row plan warning %q", warning.Code)
		}
		key := [2]int{warning.SourceRowOrdinal, warning.SourceColumnOrdinal}
		if _, duplicate := warnings[key]; duplicate {
			return fmt.Errorf("duplicate tabular row plan warning")
		}
		warnings[key] = warning.Code
	}
	seenWarnings := make(map[[2]int]struct{}, len(warnings))
	for rowIndex, row := range plan.Rows {
		if row.RowOrdinal != rowIndex+1 {
			return fmt.Errorf("tabular rows are not contiguous")
		}
		seenCells := make(map[int]struct{}, len(row.Cells))
		for _, cell := range row.Cells {
			mapping, exists := mappings[cell.SourceColumnOrdinal]
			if !exists || mapping.FieldKey != cell.FieldKey || !equalOptionalString(mapping.EntityBindingMode, cell.EntityBindingMode) {
				return fmt.Errorf("tabular cell does not match its field mapping")
			}
			if _, duplicate := seenCells[cell.SourceColumnOrdinal]; duplicate {
				return fmt.Errorf("duplicate tabular cell source column")
			}
			seenCells[cell.SourceColumnOrdinal] = struct{}{}
		}
		for _, unmapped := range row.Unmapped {
			if unmapped.SourceKind != plan.SourceKind ||
				unmapped.SourceClientTxnID != plan.ClientTxnID ||
				unmapped.SourceRowOrdinal != row.RowOrdinal {
				return fmt.Errorf("tabular unmapped value provenance mismatch")
			}
			column, exists := sourceColumns[unmapped.SourceColumnOrdinal]
			if !exists || !equalJSONValue(column.HeaderText, unmapped.SourceHeaderText) {
				return fmt.Errorf("tabular unmapped value source-column mismatch")
			}
			key := [2]int{row.RowOrdinal, unmapped.SourceColumnOrdinal}
			if warnings[key] != WarningUnmappedValueV1 {
				return fmt.Errorf("tabular unmapped value lacks its warning")
			}
			seenWarnings[key] = struct{}{}
		}
	}
	if len(seenWarnings) != len(warnings) {
		return fmt.Errorf("tabular warning lacks an unmapped value")
	}
	fingerprint, err := mappingFingerprint(plan)
	if err != nil {
		return err
	}
	if plan.MappingFingerprint == "" || plan.MappingFingerprint != fingerprint {
		return fmt.Errorf("tabular mapping fingerprint mismatch")
	}
	return nil
}

func ReadAll(reader io.Reader, format string) ([][]string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return ParseTable(string(data), format)
}

func unmappedRawValue(request MappingRequest, headers []any, rowOrdinal int, columnOrdinal int, value string) UnmappedRawValueV1 {
	var headerValue any
	if columnOrdinal > 0 && columnOrdinal <= len(headers) {
		headerValue = headers[columnOrdinal-1]
	}
	return UnmappedRawValueV1{
		SourceKind:          request.SourceKind,
		SourceClientTxnID:   request.ClientTxnID,
		SourceRowOrdinal:    rowOrdinal,
		SourceColumnOrdinal: columnOrdinal,
		SourceHeaderText:    headerValue,
		RawValue:            value,
	}
}

func normalizedSourceFormat(text string, format string) (string, error) {
	switch format {
	case "", "auto":
		if strings.Contains(strings.TrimRight(text, "\r\n"), "\t") {
			return SourceFormatTSV, nil
		}
		return SourceFormatCSV, nil
	case SourceFormatCSV, SourceFormatTSV:
		return format, nil
	default:
		return "", fmt.Errorf("unsupported tabular format")
	}
}

func exactHeaderMatches(rows [][]string, labels []string, fieldKeys []string) bool {
	if len(labels) == 0 && len(fieldKeys) == 0 {
		return false
	}
	if len(labels) == 0 || len(labels) != len(fieldKeys) || len(rows) == 0 || len(rows[0]) != len(labels) {
		return false
	}
	return slices.Equal(rows[0], labels)
}

func stringHeaders(values []string) []any {
	headers := make([]any, 0, len(values))
	for _, value := range values {
		headers = append(headers, value)
	}
	return headers
}

func equalOptionalString(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func equalJSONValue(left any, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func mappingFingerprint(plan TabularRowPlanV1) (string, error) {
	payload := struct {
		Kind          string
		ViewSchemaID  string
		SourceKind    string
		SourceFormat  string
		SourceColumns []SourceColumnV1
		FieldMappings []FieldMappingV1
	}{
		Kind:          plan.Kind,
		ViewSchemaID:  plan.ViewSchemaID,
		SourceKind:    plan.SourceKind,
		SourceFormat:  plan.SourceFormat,
		SourceColumns: plan.SourceColumns,
		FieldMappings: plan.FieldMappings,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode tabular mapping fingerprint: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}
