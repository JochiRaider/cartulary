package tabularingest

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	MaxClipboardRows = 500
	MaxClipboardCols = 64
)

type SourceColumn struct {
	SourceKind          string
	SourceClientTxnID   string
	SourceRowOrdinal    int
	SourceColumnOrdinal int
	SourceHeaderText    any
	RawValue            string
}

type CellPlan struct {
	FieldKey            string
	RawValue            string
	SourceColumnOrdinal int
	EntityBindingMode   *string
}

type RowPlan struct {
	RowOrdinal int
	Cells      []CellPlan
	Unknown    []SourceColumn
}

type BatchPlan struct {
	ViewSchemaID string
	ClientTxnID  string
	SourceKind   string
	Rows         []RowPlan
}

type MappingRequest struct {
	ViewSchemaID   string
	ClientTxnID    string
	SourceKind     string
	Text           string
	Format         string
	StartFieldKey  string
	Columns        []string
	SourceHeaders  []any
	RequireTargets int
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

func BuildBatchPlan(request MappingRequest) (BatchPlan, error) {
	if request.SourceKind == "" {
		return BatchPlan{}, fmt.Errorf("missing source kind")
	}
	schema, ok := viewschema.Lookup(request.ViewSchemaID)
	if !ok {
		return BatchPlan{}, fmt.Errorf("unknown view schema")
	}
	rows, err := ParseTable(request.Text, request.Format)
	if err != nil {
		return BatchPlan{}, err
	}
	if len(rows) == 0 || len(rows) > MaxClipboardRows {
		return BatchPlan{}, fmt.Errorf("invalid tabular row count")
	}
	if request.RequireTargets >= 0 && request.RequireTargets != len(rows) {
		return BatchPlan{}, fmt.Errorf("target count must equal tabular row count")
	}
	if len(request.Columns) == 0 || len(request.Columns) > MaxClipboardCols {
		return BatchPlan{}, fmt.Errorf("invalid target column count")
	}
	columnStart := slices.Index(request.Columns, request.StartFieldKey)
	if columnStart < 0 {
		return BatchPlan{}, fmt.Errorf("start field not present in target columns")
	}
	fields := schema.Fields()
	for _, fieldKey := range request.Columns {
		if _, ok := fields[fieldKey]; !ok {
			return BatchPlan{}, fmt.Errorf("unsupported field key %s", fieldKey)
		}
	}

	plan := BatchPlan{
		ViewSchemaID: request.ViewSchemaID,
		ClientTxnID:  request.ClientTxnID,
		SourceKind:   request.SourceKind,
		Rows:         make([]RowPlan, 0, len(rows)),
	}
	for rowIndex, values := range rows {
		rowPlan := RowPlan{RowOrdinal: rowIndex + 1}
		seenFields := make(map[string]struct{})
		for columnOffset, rawValue := range values {
			columnIndex := columnStart + columnOffset
			sourceColumnOrdinal := columnOffset + 1
			if columnIndex >= len(request.Columns) {
				rowPlan.Unknown = append(rowPlan.Unknown, rawColumn(request, rowIndex+1, sourceColumnOrdinal, nil, rawValue))
				continue
			}
			fieldKey := request.Columns[columnIndex]
			field, ok := fields[fieldKey]
			if !ok || !field.Writable {
				rowPlan.Unknown = append(rowPlan.Unknown, rawColumn(request, rowIndex+1, sourceColumnOrdinal, &fieldKey, rawValue))
				continue
			}
			if _, duplicate := seenFields[fieldKey]; duplicate {
				return BatchPlan{}, fmt.Errorf("duplicate field key %s", fieldKey)
			}
			seenFields[fieldKey] = struct{}{}
			rowPlan.Cells = append(rowPlan.Cells, CellPlan{
				FieldKey:            fieldKey,
				RawValue:            rawValue,
				SourceColumnOrdinal: sourceColumnOrdinal,
				EntityBindingMode:   field.EntityBindingMode,
			})
		}
		plan.Rows = append(plan.Rows, rowPlan)
	}
	return plan, nil
}

func ReadAll(reader io.Reader, format string) ([][]string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return ParseTable(string(data), format)
}

func rawColumn(request MappingRequest, rowOrdinal int, columnOrdinal int, header *string, value string) SourceColumn {
	var headerValue any
	if header != nil {
		headerValue = *header
	} else if columnOrdinal > 0 && columnOrdinal <= len(request.SourceHeaders) {
		headerValue = request.SourceHeaders[columnOrdinal-1]
	}
	return SourceColumn{
		SourceKind:          request.SourceKind,
		SourceClientTxnID:   request.ClientTxnID,
		SourceRowOrdinal:    rowOrdinal,
		SourceColumnOrdinal: columnOrdinal,
		SourceHeaderText:    headerValue,
		RawValue:            value,
	}
}
