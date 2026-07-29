package imports

import (
	"encoding/json"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func (s *Service) discoverImportUnits(envelope httpapi.UploadEnvelope, sourceFileKind string) ([]DiscoveredUnit, *httpapi.APIError) {
	switch sourceFileKind {
	case SourceFileKindCSV:
		if int64(len(envelope.File)) > s.limits.MaxCSVSourceBytes {
			return nil, importSourceRejected("csv_source_too_large", int64(len(envelope.File)), s.limits.MaxCSVSourceBytes)
		}
		maxColumns := int(s.limits.MaxColumns)
		parsedRows, err := tabularingest.ParseTableWithMaxColumns(string(envelope.File), "csv", maxColumns)
		if err != nil {
			if strings.Contains(err.Error(), "column count exceeded") {
				return nil, importSourceRejected("import_columns_exceeded", int64(maxColumns+1), s.limits.MaxColumns)
			}
			return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		unit, apiErr := s.discoveredImportUnit(tabularCellsFromStrings(parsedRows), "csv_file", "file")
		if apiErr != nil {
			return nil, apiErr
		}
		return []DiscoveredUnit{unit}, nil
	case SourceFileKindXLSX:
		if int64(len(envelope.File)) > s.limits.MaxXLSXSourceBytes {
			return nil, importSourceRejected("xlsx_source_too_large", int64(len(envelope.File)), s.limits.MaxXLSXSourceBytes)
		}
		workbook, apiErr := indexXLSXWorkbook(envelope.File, s.limits, s.archiveLimits)
		if apiErr != nil {
			return nil, apiErr
		}
		units := make([]DiscoveredUnit, 0, len(workbook.ranges))
		for _, located := range workbook.ranges {
			decoded, decodeErr := workbook.decodeRectangle(located.sheet, located.rect, s.limits)
			if decodeErr != nil {
				return nil, decodeErr
			}
			locator := map[string]any{
				"sheet_name": located.sheet.name,
				"rect_a1":    sourceRectangleA1(located.rect),
			}
			if located.tableName != "" {
				locator["table_name"] = located.tableName
			}
			if located.definedName != "" {
				locator["defined_name"] = located.definedName
			}
			locatorBytes, err := json.Marshal(locator)
			if err != nil {
				return nil, internalAPIError(err)
			}
			unit, unitErr := s.discoveredImportUnitAt(
				decoded.rows,
				located.kind,
				string(locatorBytes),
				located.rect,
				decoded.warningCodes,
				decoded.blockingColumnOrdinals,
			)
			if unitErr != nil {
				return nil, unitErr
			}
			units = append(units, unit)
		}
		return units, nil
	default:
		return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
}

func (s *Service) discoveredImportUnit(rows [][]tabularCell, locatorKind string, locator string) (DiscoveredUnit, *httpapi.APIError) {
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	rect := sourceRectangle{left: 1, top: 1, right: maxCols, bottom: len(rows)}
	if rect.right < 1 {
		rect.right = 1
	}
	if rect.bottom < 1 {
		rect.bottom = 1
	}
	return s.discoveredImportUnitAt(rows, locatorKind, locator, rect, nil, nil)
}

func (s *Service) discoveredImportUnitAt(
	rows [][]tabularCell,
	locatorKind string,
	locator string,
	rect sourceRectangle,
	warningCodes []string,
	blockingColumnOrdinals []int,
) (DiscoveredUnit, *httpapi.APIError) {
	if warningCodes == nil {
		warningCodes = []string{}
	}
	if blockingColumnOrdinals == nil {
		blockingColumnOrdinals = []int{}
	}
	maxCols := rect.right - rect.left + 1
	dataRows := len(rows) - 1
	if dataRows < 0 {
		dataRows = 0
	}
	if int64(dataRows) > s.limits.MaxRows {
		return DiscoveredUnit{}, importSourceRejected("import_rows_exceeded", int64(dataRows), s.limits.MaxRows)
	}
	if int64(dataRows*maxCols) > s.limits.MaxCells {
		return DiscoveredUnit{}, importSourceRejected("import_cells_exceeded", int64(dataRows*maxCols), s.limits.MaxCells)
	}
	columns := make([]map[string]any, 0, maxCols)
	if len(rows) > 0 {
		for index := 0; index < maxCols; index++ {
			var header any
			if index < len(rows[0]) && rows[0][index].DisplayText != "" {
				header = rows[0][index].DisplayText
			}
			columns = append(columns, map[string]any{
				"source_column_ordinal": index + 1,
				"source_header_text":    header,
			})
		}
	}
	previewRows := make([]map[string]any, 0)
	sourceRows := make([]map[string]any, 0)
	for rowIndex := 1; rowIndex < len(rows) && len(previewRows) < 50; rowIndex++ {
		sourceRowRef := rect.top + rowIndex
		cells := make([]map[string]any, 0, maxCols)
		for columnIndex := 0; columnIndex < maxCols; columnIndex++ {
			cell := tabularCell{CellKind: "blank"}
			if columnIndex < len(rows[rowIndex]) {
				cell = rows[rowIndex][columnIndex]
			}
			cells = append(cells, map[string]any{
				"source_column_ordinal": columnIndex + 1,
				"display_text":          cell.DisplayText,
				"cell_kind":             cell.CellKind,
			})
		}
		previewRows = append(previewRows, map[string]any{
			"source_row_ref": sourceRowRef,
			"cells":          cells,
		})
	}
	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		sourceRowRef := rect.top + rowIndex
		cells := make([]map[string]any, 0, maxCols)
		for columnIndex := 0; columnIndex < maxCols; columnIndex++ {
			cell := tabularCell{CellKind: "blank"}
			if columnIndex < len(rows[rowIndex]) {
				cell = rows[rowIndex][columnIndex]
			}
			cells = append(cells, map[string]any{
				"source_column_ordinal": columnIndex + 1,
				"display_text":          cell.DisplayText,
				"cell_kind":             cell.CellKind,
			})
		}
		sourceRows = append(sourceRows, map[string]any{
			"source_row_ref": sourceRowRef,
			"cells":          cells,
		})
	}
	return DiscoveredUnit{
		LocatorKind:         locatorKind,
		Locator:             locator,
		SourceRectA1:        sourceRectangleA1(rect),
		HeaderRowRef:        rect.top,
		DataStartRowRef:     rect.top + 1,
		InferredRowCount:    dataRows,
		InferredColumnCount: maxCols,
		WarningCodes:        warningCodes,
		BlockingColumns:     blockingColumnOrdinals,
		Columns:             columns,
		SourceRows:          sourceRows,
		PreviewRows:         previewRows,
	}, nil
}

func detectSourceFileKind(envelope httpapi.UploadEnvelope) string {
	if looksLikeXLSX(envelope.File) {
		return SourceFileKindXLSX
	}
	return SourceFileKindCSV
}

func looksLikeXLSX(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return (data[0] == 'P' && data[1] == 'K' && data[2] == 0x03 && data[3] == 0x04) ||
		(data[0] == 'P' && data[1] == 'K' && data[2] == 0x05 && data[3] == 0x06) ||
		(data[0] == 'P' && data[1] == 'K' && data[2] == 0x07 && data[3] == 0x08)
}
