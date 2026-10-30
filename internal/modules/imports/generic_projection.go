package imports

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// genericCSVPreview cannot retain a full approval projection.
type genericCSVPreview struct {
	SourceRectA1        string
	HeaderRowRef        int
	DataStartRowRef     int
	InferredRowCount    int
	InferredColumnCount int
	WarningCodes        []string
	Columns             []map[string]any
	PreviewRows         []map[string]any
}

func analyzeGenericCSV(ctx context.Context, source io.Reader) (genericCSVPreview, error) {
	return scanGenericCSV(ctx, source, nil)
}

func materializeGenericCSV(ctx context.Context, source io.Reader) (discoveredUnit, error) {
	rows := make([]map[string]any, 0)
	preview, err := scanGenericCSV(ctx, source, func(row map[string]any) { rows = append(rows, row) })
	if err != nil {
		return discoveredUnit{}, err
	}
	return discoveredUnit{
		LocatorKind: "csv_file", Locator: "file", SourceRectA1: preview.SourceRectA1,
		HeaderRowRef: preview.HeaderRowRef, DataStartRowRef: preview.DataStartRowRef,
		InferredRowCount: preview.InferredRowCount, InferredColumnCount: preview.InferredColumnCount,
		WarningCodes: preview.WarningCodes, BlockingColumns: []int{}, Columns: preview.Columns,
		PreviewRows: preview.PreviewRows, SourceRows: rows,
	}, nil
}

// A nil consumer keeps only header/preview values while validating every record.
func scanGenericCSV(ctx context.Context, source io.Reader, consume func(map[string]any)) (genericCSVPreview, error) {
	decoder, err := tabularingest.NewDelimitedDecoder(ctx, source, "csv", 0)
	if err != nil {
		return genericCSVPreview{}, err
	}
	result := genericCSVPreview{HeaderRowRef: 1, DataStartRowRef: 2, WarningCodes: []string{}, Columns: []map[string]any{}, PreviewRows: []map[string]any{}}
	var header []string
	recordCount, width := 0, 0
	for {
		retain := recordCount <= 50 || consume != nil
		record, err := decoder.Next(retain)
		if err == io.EOF {
			break
		}
		if err != nil {
			return genericCSVPreview{}, err
		}
		recordCount++
		width = max(width, record.FieldCount)
		if recordCount == 1 {
			header = record.Fields
			continue
		}
		if !retain {
			continue
		}
		cells := make([]map[string]any, 0, record.FieldCount)
		for i, value := range record.Fields {
			if err := ctx.Err(); err != nil {
				return genericCSVPreview{}, err
			}
			kind := "string"
			if value == "" {
				kind = "blank"
			}
			cells = append(cells, map[string]any{"source_column_ordinal": i + 1, "display_text": value, "cell_kind": kind})
		}
		row := map[string]any{"source_row_ref": recordCount, "cells": cells}
		if len(result.PreviewRows) < 50 {
			result.PreviewRows = append(result.PreviewRows, row)
		}
		if consume != nil {
			consume(row)
		}
	}
	if recordCount == 0 {
		return genericCSVPreview{}, fmt.Errorf("%w: empty_table", tabularingest.ErrInvalidDelimited)
	}
	for i := 0; i < width; i++ {
		if err := ctx.Err(); err != nil {
			return genericCSVPreview{}, err
		}
		var text any
		if i < len(header) && header[i] != "" {
			text = header[i]
		}
		result.Columns = append(result.Columns, map[string]any{"source_column_ordinal": i + 1, "source_header_text": text})
	}
	if err := ctx.Err(); err != nil {
		return genericCSVPreview{}, err
	}
	result.SourceRectA1 = sourceRectangleA1(sourceRectangle{left: 1, top: 1, right: width, bottom: recordCount})
	result.InferredRowCount, result.InferredColumnCount = recordCount-1, width
	return result, nil
}

func (s *service) openGenericCSV(ctx context.Context, sessionID, unitID uuid.UUID) (ImportSourceStream, *httpapi.APIError) {
	capability, err := s.store.sourceCapabilityForUnit(ctx, sessionID, unitID)
	if err != nil {
		return ImportSourceStream{}, internalAPIError(err)
	}
	port := extensionSourcePort{db: s.store.db}
	stream, err := port.OpenSourceStream(ctx, capability.SourceStreamRef)
	if err != nil {
		return ImportSourceStream{}, internalAPIError(err)
	}
	if stream.ImportSourceCapability != capability {
		_ = stream.Reader.Close()
		return ImportSourceStream{}, invalidImportRequest("source", "owner_preview_validation_failed")
	}
	return stream, nil
}

func genericCSVError(err error) *httpapi.APIError {
	if errors.Is(err, tabularingest.ErrInvalidDelimited) {
		return importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
	return internalAPIError(err)
}

func (s *service) genericPreview(ctx context.Context, sessionID, unitID uuid.UUID, unit map[string]any) (map[string]any, error, *httpapi.APIError) {
	if unit["source_rect_a1"] != nil {
		preview, _, err := s.store.getPreview(ctx, sessionID, unitID)
		return preview, err, nil
	}
	stream, apiErr := s.openGenericCSV(ctx, sessionID, unitID)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	defer func() { _ = stream.Reader.Close() }()
	projection, err := analyzeGenericCSV(ctx, stream.Reader)
	if err != nil {
		return nil, nil, genericCSVError(err)
	}
	preview := map[string]any{
		"import_session_id": sessionID.String(), "import_unit_id": unitID.String(),
		"locator_kind": unit["locator_kind"], "locator": unit["locator"],
		"source_rect_a1": projection.SourceRectA1, "header_row_ref": projection.HeaderRowRef,
		"data_start_row_ref": projection.DataStartRowRef, "inferred_row_count": projection.InferredRowCount,
		"inferred_column_count": projection.InferredColumnCount, "warning_codes": projection.WarningCodes,
		"unit_status": unit["unit_status"], "columns": projection.Columns, "preview_rows": projection.PreviewRows,
		"truncated": projection.InferredRowCount > len(projection.PreviewRows),
	}
	if fingerprint, ok := unit["mapping_fingerprint"]; ok {
		preview["mapping_fingerprint"] = fingerprint
	}
	return preview, nil, nil
}

func validateMappingColumns(mapping approvedMapping, columns []map[string]any) *httpapi.APIError {
	if len(mapping.SourceColumns) != len(columns) {
		return invalidImportRequest("source_columns", "invalid_source_columns")
	}
	for i, column := range mapping.SourceColumns {
		if column.SourceColumnOrdinal != i+1 || !sourceHeaderEqual(column.SourceHeaderText, columns[i]["source_header_text"]) {
			return invalidImportRequest("source_columns", "invalid_source_columns")
		}
	}
	return nil
}

func (s *service) prepareGenericMapping(ctx context.Context, route importSessionRoute, request mappingRequest) (mappingRequest, *httpapi.APIError) {
	unit, _, err := s.store.getUnit(ctx, route.SessionID, route.UnitID)
	if err != nil {
		return mappingRequest{}, internalAPIError(err)
	}
	var columns []map[string]any
	if unit["source_rect_a1"] == nil {
		stream, apiErr := s.openGenericCSV(ctx, route.SessionID, route.UnitID)
		if apiErr != nil {
			return mappingRequest{}, apiErr
		}
		defer func() { _ = stream.Reader.Close() }()
		projection, err := materializeGenericCSV(ctx, stream.Reader)
		if err != nil {
			return mappingRequest{}, genericCSVError(err)
		}
		request.Projection, request.SourceCapability = &projection, &stream.ImportSourceCapability
		columns = projection.Columns
	} else {
		var raw []byte
		if err := s.store.db.QueryRow(ctx, `SELECT columns_json FROM import_units WHERE import_session_id=$1 AND import_unit_id=$2`, route.SessionID, route.UnitID).Scan(&raw); err != nil {
			return mappingRequest{}, internalAPIError(err)
		}
		if err := json.Unmarshal(raw, &columns); err != nil {
			return mappingRequest{}, internalAPIError(err)
		}
	}
	if apiErr := validateMappingColumns(request.approvedMapping, columns); apiErr != nil {
		return mappingRequest{}, apiErr
	}
	return request, nil
}

func saveGenericProjectionTx(ctx context.Context, tx pgx.Tx, sessionID, unitID uuid.UUID, projection discoveredUnit) error {
	columns, err := json.Marshal(projection.Columns)
	if err != nil {
		return err
	}
	rows, err := json.Marshal(projection.SourceRows)
	if err != nil {
		return err
	}
	preview, err := json.Marshal(projection.PreviewRows)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE import_units SET source_rect_a1=$3, header_row_ref=$4, data_start_row_ref=$5,
inferred_row_count=$6, inferred_column_count=$7, columns_json=$8, source_rows_json=$9, preview_rows_json=$10
WHERE import_session_id=$1 AND import_unit_id=$2`, sessionID, unitID, projection.SourceRectA1,
		projection.HeaderRowRef, projection.DataStartRowRef, projection.InferredRowCount, projection.InferredColumnCount,
		columns, rows, preview)
	return err
}
