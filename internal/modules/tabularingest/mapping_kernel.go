package tabularingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	MappingKernelContractV1 = "cartulary.tabular_mapping_kernel.v1"

	UnknownColumnPreserveRawCaptureV1  = "preserve_raw_capture"
	UnknownColumnPreserveCustomAttrsV1 = "preserve_custom_attrs"
	UnknownColumnRejectIfUnmappedV1    = "reject_if_unmapped"

	TransformTrimV1               = "trim_v1"
	TransformCollapseWhitespaceV1 = "collapse_whitespace_v1"
	TransformLowercaseV1          = "lowercase_v1"
	TransformSplitDelimitedV1     = "split_delimited_v1"

	EmptyValueOmitFieldV1 = "omit_field"
	EmptyValueWriteNullV1 = "write_null"

	MappingKernelCellScalarV1 = "scalar"
	MappingKernelCellEmptyV1  = "empty"

	MappingKernelValueV1 = "value"
	MappingKernelOmitV1  = "omit"
	MappingKernelNullV1  = "null"

	MappingKernelWarningUnmappedV1 = "unmapped_source_value"
)

var ErrMappingKernelCanceled = errors.New("tabular mapping kernel canceled")

// MappingKernelTargetFieldV1 is the complete owner-neutral field fact needed
// to validate and order a view-target mapping. Order is the immutable schema
// order; output remains in approved source-column order for v1 compatibility.
type MappingKernelTargetFieldV1 struct {
	FieldKey          string
	Order             int
	Writable          bool
	CreateWritable    bool
	Clearable         bool
	EntityBindingMode *string
}

type MappingKernelSourceColumnV1 struct {
	SourceColumnOrdinal int
	SourceHeaderText    any
	FieldKey            *string
	EntityBindingMode   *string
	TransformID         *string
	TransformOptions    map[string]any
	EmptyValuePolicy    string
}

type MappingKernelScalarCellV1 struct {
	SourceColumnOrdinal int
	RawValue            string
	CellKind            string
	Classification      string
	Present             bool
}

type MappingKernelSourceRowV1 struct {
	SourceRowOrdinal int
	Cells            []MappingKernelScalarCellV1
}

type MappingKernelRequestV1 struct {
	TargetFields        []MappingKernelTargetFieldV1
	SourceColumns       []MappingKernelSourceColumnV1
	Rows                []MappingKernelSourceRowV1
	UnknownColumnPolicy string
}

type MappingKernelValuePlanV1 struct {
	FieldKey            string
	RawValue            string
	TransformedValue    string
	Disposition         string
	SourceColumnOrdinal int
	SourceHeaderText    any
	CellKind            string
	EntityBindingMode   *string
	TransformID         *string
	EmptyValuePolicy    string
}

type MappingKernelUnknownValueV1 struct {
	SourceColumnOrdinal int
	SourceHeaderText    any
	RawValue            string
	CellKind            string
}

type MappingKernelRowPlanV1 struct {
	SourceRowOrdinal int
	Values           []MappingKernelValuePlanV1
	UnknownValues    []MappingKernelUnknownValueV1
}

type MappingKernelWarningV1 struct {
	Code                string
	SourceRowOrdinal    int
	SourceColumnOrdinal int
}

type MappingKernelPlanV1 struct {
	ContractID string
	Rows       []MappingKernelRowPlanV1
	Warnings   []MappingKernelWarningV1
}

// BuildMappingKernelPlanV1 validates and evaluates one parser-neutral mapping
// without I/O or owner-specific normalization. It builds the result locally so
// cancellation or validation failure cannot expose a partial applicable plan.
func BuildMappingKernelPlanV1(ctx context.Context, request MappingKernelRequestV1) (MappingKernelPlanV1, error) {
	if err := mappingKernelCanceled(ctx); err != nil {
		return MappingKernelPlanV1{}, err
	}
	fields, err := validateMappingKernelTargetFields(request.TargetFields)
	if err != nil {
		return MappingKernelPlanV1{}, err
	}
	columns, err := validateMappingKernelSourceColumns(request.SourceColumns, fields, request.UnknownColumnPolicy)
	if err != nil {
		return MappingKernelPlanV1{}, err
	}
	plan := MappingKernelPlanV1{
		ContractID: MappingKernelContractV1,
		Rows:       make([]MappingKernelRowPlanV1, 0, len(request.Rows)),
		Warnings:   []MappingKernelWarningV1{},
	}
	previousRowOrdinal := 0
	for rowIndex, row := range request.Rows {
		if err := mappingKernelCanceled(ctx); err != nil {
			return MappingKernelPlanV1{}, err
		}
		if row.SourceRowOrdinal <= previousRowOrdinal {
			return MappingKernelPlanV1{}, fmt.Errorf("tabular mapping kernel source rows are not strictly ordered at index %d", rowIndex)
		}
		previousRowOrdinal = row.SourceRowOrdinal
		if len(row.Cells) != len(columns) {
			return MappingKernelPlanV1{}, fmt.Errorf("tabular mapping kernel row %d has %d cells, want %d", row.SourceRowOrdinal, len(row.Cells), len(columns))
		}
		rowPlan := MappingKernelRowPlanV1{
			SourceRowOrdinal: row.SourceRowOrdinal,
			Values:           []MappingKernelValuePlanV1{},
			UnknownValues:    []MappingKernelUnknownValueV1{},
		}
		for columnIndex, cell := range row.Cells {
			if err := mappingKernelCanceled(ctx); err != nil {
				return MappingKernelPlanV1{}, err
			}
			column := columns[columnIndex]
			if cell.SourceColumnOrdinal != column.SourceColumnOrdinal {
				return MappingKernelPlanV1{}, fmt.Errorf("tabular mapping kernel row %d cells are not in source-column order", row.SourceRowOrdinal)
			}
			if !cell.Present {
				if cell.RawValue != "" || cell.CellKind != "" || cell.Classification != "" {
					return MappingKernelPlanV1{}, fmt.Errorf("tabular mapping kernel absent cell %d carries a value", cell.SourceColumnOrdinal)
				}
				continue
			}
			if cell.CellKind == "" || (cell.Classification != MappingKernelCellScalarV1 && cell.Classification != MappingKernelCellEmptyV1) {
				return MappingKernelPlanV1{}, fmt.Errorf("tabular mapping kernel cell %d has invalid classification", cell.SourceColumnOrdinal)
			}
			if column.FieldKey == nil {
				unknown := MappingKernelUnknownValueV1{
					SourceColumnOrdinal: column.SourceColumnOrdinal,
					SourceHeaderText:    cloneMappingKernelJSONValue(column.SourceHeaderText),
					RawValue:            cell.RawValue,
					CellKind:            cell.CellKind,
				}
				rowPlan.UnknownValues = append(rowPlan.UnknownValues, unknown)
				plan.Warnings = append(plan.Warnings, MappingKernelWarningV1{
					Code:                MappingKernelWarningUnmappedV1,
					SourceRowOrdinal:    row.SourceRowOrdinal,
					SourceColumnOrdinal: column.SourceColumnOrdinal,
				})
				continue
			}
			transformed, err := mappingKernelTransform(cell.RawValue, column)
			if err != nil {
				return MappingKernelPlanV1{}, err
			}
			disposition := MappingKernelValueV1
			semanticEmpty := cell.Classification == MappingKernelCellEmptyV1 || (cell.RawValue != "" && transformed == "")
			if semanticEmpty {
				switch column.EmptyValuePolicy {
				case EmptyValueOmitFieldV1:
					disposition = MappingKernelOmitV1
				case EmptyValueWriteNullV1:
					disposition = MappingKernelNullV1
				}
			}
			rowPlan.Values = append(rowPlan.Values, MappingKernelValuePlanV1{
				FieldKey:            *column.FieldKey,
				RawValue:            cell.RawValue,
				TransformedValue:    transformed,
				Disposition:         disposition,
				SourceColumnOrdinal: column.SourceColumnOrdinal,
				SourceHeaderText:    cloneMappingKernelJSONValue(column.SourceHeaderText),
				CellKind:            cell.CellKind,
				EntityBindingMode:   cloneMappingKernelString(column.EntityBindingMode),
				TransformID:         cloneMappingKernelString(column.TransformID),
				EmptyValuePolicy:    column.EmptyValuePolicy,
			})
		}
		plan.Rows = append(plan.Rows, rowPlan)
	}
	return plan, nil
}

func validateMappingKernelTargetFields(fields []MappingKernelTargetFieldV1) (map[string]MappingKernelTargetFieldV1, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("tabular mapping kernel target fields are empty")
	}
	byKey := make(map[string]MappingKernelTargetFieldV1, len(fields))
	orders := make(map[int]struct{}, len(fields))
	previousOrder := 0
	for index, field := range fields {
		if field.FieldKey == "" || field.Order <= previousOrder {
			return nil, fmt.Errorf("tabular mapping kernel target fields are not in schema order at index %d", index)
		}
		if _, duplicate := byKey[field.FieldKey]; duplicate {
			return nil, fmt.Errorf("tabular mapping kernel duplicate target field %q", field.FieldKey)
		}
		if _, duplicate := orders[field.Order]; duplicate {
			return nil, fmt.Errorf("tabular mapping kernel duplicate target field order %d", field.Order)
		}
		previousOrder = field.Order
		field.EntityBindingMode = cloneMappingKernelString(field.EntityBindingMode)
		byKey[field.FieldKey] = field
		orders[field.Order] = struct{}{}
	}
	return byKey, nil
}

func validateMappingKernelSourceColumns(
	columns []MappingKernelSourceColumnV1,
	fields map[string]MappingKernelTargetFieldV1,
	unknownPolicy string,
) ([]MappingKernelSourceColumnV1, error) {
	switch unknownPolicy {
	case UnknownColumnPreserveRawCaptureV1, UnknownColumnPreserveCustomAttrsV1, UnknownColumnRejectIfUnmappedV1:
	default:
		return nil, fmt.Errorf("tabular mapping kernel invalid unknown-column policy %q", unknownPolicy)
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("tabular mapping kernel source columns are empty")
	}
	validated := make([]MappingKernelSourceColumnV1, 0, len(columns))
	mappedFields := make(map[string]struct{}, len(columns))
	for index, source := range columns {
		if source.SourceColumnOrdinal != index+1 {
			return nil, fmt.Errorf("tabular mapping kernel source columns are not contiguous")
		}
		if _, err := json.Marshal(source.SourceHeaderText); err != nil {
			return nil, fmt.Errorf("tabular mapping kernel source header %d is not a scalar JSON value", source.SourceColumnOrdinal)
		}
		if source.TransformOptions == nil {
			return nil, fmt.Errorf("tabular mapping kernel transform options are nil for column %d", source.SourceColumnOrdinal)
		}
		if source.FieldKey == nil {
			if source.EntityBindingMode != nil || source.TransformID != nil || len(source.TransformOptions) != 0 || source.EmptyValuePolicy != EmptyValueOmitFieldV1 {
				return nil, fmt.Errorf("tabular mapping kernel unmapped column %d carries mapped-field behavior", source.SourceColumnOrdinal)
			}
			if unknownPolicy == UnknownColumnRejectIfUnmappedV1 {
				return nil, fmt.Errorf("tabular mapping kernel rejects unmapped column %d", source.SourceColumnOrdinal)
			}
		} else {
			field, exists := fields[*source.FieldKey]
			if !exists || (!field.Writable && !field.CreateWritable) {
				return nil, fmt.Errorf("tabular mapping kernel field %q is not import writable", *source.FieldKey)
			}
			if _, duplicate := mappedFields[*source.FieldKey]; duplicate {
				return nil, fmt.Errorf("tabular mapping kernel duplicate mapped field %q", *source.FieldKey)
			}
			if !mappingKernelOptionalStringEqual(source.EntityBindingMode, field.EntityBindingMode) {
				return nil, fmt.Errorf("tabular mapping kernel entity binding mode drift for %q", *source.FieldKey)
			}
			switch source.EmptyValuePolicy {
			case EmptyValueOmitFieldV1:
			case EmptyValueWriteNullV1:
				if !field.Clearable {
					return nil, fmt.Errorf("tabular mapping kernel field %q is not clearable", *source.FieldKey)
				}
			default:
				return nil, fmt.Errorf("tabular mapping kernel invalid empty-value policy %q", source.EmptyValuePolicy)
			}
			if err := validateMappingKernelTransform(source); err != nil {
				return nil, err
			}
			mappedFields[*source.FieldKey] = struct{}{}
		}
		source.FieldKey = cloneMappingKernelString(source.FieldKey)
		source.EntityBindingMode = cloneMappingKernelString(source.EntityBindingMode)
		source.TransformID = cloneMappingKernelString(source.TransformID)
		source.TransformOptions = cloneMappingKernelOptions(source.TransformOptions)
		source.SourceHeaderText = cloneMappingKernelJSONValue(source.SourceHeaderText)
		validated = append(validated, source)
	}
	return validated, nil
}

func validateMappingKernelTransform(column MappingKernelSourceColumnV1) error {
	if column.TransformID == nil {
		if len(column.TransformOptions) != 0 {
			return fmt.Errorf("tabular mapping kernel transform options require a transform")
		}
		return nil
	}
	if *column.TransformID != TransformSplitDelimitedV1 {
		if len(column.TransformOptions) != 0 {
			return fmt.Errorf("tabular mapping kernel transform %q does not accept options", *column.TransformID)
		}
		switch *column.TransformID {
		case TransformTrimV1, TransformCollapseWhitespaceV1, TransformLowercaseV1:
			return nil
		default:
			return fmt.Errorf("tabular mapping kernel unsupported transform %q", *column.TransformID)
		}
	}
	for key, value := range column.TransformOptions {
		switch key {
		case "delimiter":
			if delimiter, ok := value.(string); !ok || !mappingKernelDelimiterAllowed(delimiter) {
				return fmt.Errorf("tabular mapping kernel invalid split delimiter")
			}
		case "trim_items", "drop_empty_items":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("tabular mapping kernel split option %q is not Boolean", key)
			}
		default:
			return fmt.Errorf("tabular mapping kernel unsupported split option %q", key)
		}
	}
	delimiter, ok := column.TransformOptions["delimiter"].(string)
	if !ok || !mappingKernelDelimiterAllowed(delimiter) {
		return fmt.Errorf("tabular mapping kernel split delimiter is required")
	}
	return nil
}

func mappingKernelTransform(value string, column MappingKernelSourceColumnV1) (string, error) {
	if column.TransformID == nil {
		return value, nil
	}
	switch *column.TransformID {
	case TransformTrimV1:
		return strings.TrimSpace(value), nil
	case TransformCollapseWhitespaceV1:
		return strings.Join(strings.Fields(value), " "), nil
	case TransformLowercaseV1:
		return strings.ToLower(value), nil
	case TransformSplitDelimitedV1:
		delimiter := column.TransformOptions["delimiter"].(string)
		trimItems, _ := column.TransformOptions["trim_items"].(bool)
		dropEmpty, _ := column.TransformOptions["drop_empty_items"].(bool)
		parts := strings.Split(value, delimiter)
		transformed := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimItems {
				part = strings.TrimSpace(part)
			}
			if dropEmpty && part == "" {
				continue
			}
			transformed = append(transformed, part)
		}
		return strings.Join(transformed, delimiter), nil
	default:
		return "", fmt.Errorf("tabular mapping kernel unsupported transform %q", *column.TransformID)
	}
}

func mappingKernelDelimiterAllowed(value string) bool {
	switch value {
	case ",", ";", "|", "\n", "\t":
		return true
	default:
		return false
	}
}

func mappingKernelCanceled(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("tabular mapping kernel cancellation context is nil")
	}
	select {
	case <-ctx.Done():
		return ErrMappingKernelCanceled
	default:
		return nil
	}
}

func mappingKernelOptionalStringEqual(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func cloneMappingKernelString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneMappingKernelOptions(options map[string]any) map[string]any {
	cloned := make(map[string]any, len(options))
	for key, value := range options {
		cloned[key] = value
	}
	return cloned
}

func cloneMappingKernelJSONValue(value any) any {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var cloned any
	if err := json.Unmarshal(encoded, &cloned); err != nil {
		return nil
	}
	return cloned
}
