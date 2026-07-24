package imports

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"math"
	"path"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type tabularCell struct {
	DisplayText string
	CellKind    string
}

func tabularCellsFromStrings(rows [][]string) [][]tabularCell {
	table := make([][]tabularCell, 0, len(rows))
	for _, row := range rows {
		cells := make([]tabularCell, 0, len(row))
		for _, value := range row {
			kind := "string"
			if value == "" {
				kind = "blank"
			}
			cells = append(cells, tabularCell{DisplayText: value, CellKind: kind})
		}
		table = append(table, cells)
	}
	return table
}

type xlsxSharedStrings struct {
	Items []xlsxSharedString `xml:"si"`
}

type xlsxSharedString struct {
	Text string          `xml:"t"`
	Runs []xlsxStringRun `xml:"r"`
}

type xlsxStringRun struct {
	Text string `xml:"t"`
}

type xlsxWorkbook struct {
	Sheets []xlsxSheet `xml:"sheets>sheet"`
}

type xlsxSheet struct {
	Name  string `xml:"name,attr"`
	State string `xml:"state,attr"`
	RID   string `xml:"id,attr"`
}

type xlsxRelationships struct {
	Relationships []xlsxRelationship `xml:"Relationship"`
}

type xlsxRelationship struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type xlsxWorksheet struct {
	Rows []xlsxRow `xml:"sheetData>row"`
}

type xlsxRow struct {
	Ref   int        `xml:"r,attr"`
	Cells []xlsxCell `xml:"c"`
}

type xlsxCell struct {
	Ref          string           `xml:"r,attr"`
	Type         string           `xml:"t,attr"`
	Formula      *string          `xml:"f"`
	Value        string           `xml:"v"`
	InlineString xlsxInlineString `xml:"is"`
}

type xlsxInlineString struct {
	Text string          `xml:"t"`
	Runs []xlsxStringRun `xml:"r"`
}

type xlsxDiscoveredTable struct {
	SheetName string
	Rows      [][]tabularCell
}

func parseXLSXTables(data []byte, importLimits config.ImportLimits, archiveLimits config.ArchiveLimits) ([]xlsxDiscoveredTable, *httpapi.APIError) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
	files := make(map[string]*zip.File, len(reader.File))
	var memberCount int64
	var extractedBytes uint64
	var compressedBytes uint64
	for _, file := range reader.File {
		cleanName, ok := cleanXLSXPath(file.Name)
		if !ok {
			return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		if file.FileInfo().IsDir() {
			continue
		}
		memberCount++
		if memberCount > archiveLimits.MaxMembers {
			return nil, importSourceRejected("archive_member_count_exceeded", memberCount, archiveLimits.MaxMembers)
		}
		extractedBytes += file.UncompressedSize64
		compressedBytes += file.CompressedSize64
		if extractedBytes > uint64(archiveLimits.DefaultMaxExtractedBytes) {
			return nil, importSourceRejected("archive_extracted_bytes_exceeded", int64MinUint64(extractedBytes), archiveLimits.DefaultMaxExtractedBytes)
		}
		if archiveCompressionRatioExceeded(extractedBytes, compressedBytes, archiveLimits.MaxCompressionRatio) {
			return nil, importSourceRejected("archive_compression_ratio_exceeded", int64MinUint64(extractedBytes), archiveLimits.MaxCompressionRatio)
		}
		files[cleanName] = file
	}

	workbookBytes, ok := readZipFile(files, "xl/workbook.xml")
	if !ok {
		return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
	var workbook xlsxWorkbook
	if err := xml.Unmarshal(workbookBytes, &workbook); err != nil {
		return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
	}

	relsBytes, ok := readZipFile(files, "xl/_rels/workbook.xml.rels")
	if !ok {
		return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
	var rels xlsxRelationships
	if err := xml.Unmarshal(relsBytes, &rels); err != nil {
		return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
	relTargets := make(map[string]string, len(rels.Relationships))
	for _, rel := range rels.Relationships {
		if rel.ID == "" || rel.Target == "" {
			continue
		}
		if target, ok := resolveWorkbookRelationshipTarget(rel.Target); ok {
			relTargets[rel.ID] = target
		}
	}

	sharedStrings, apiErr := parseXLSXSharedStrings(files)
	if apiErr != nil {
		return nil, apiErr
	}
	tables := make([]xlsxDiscoveredTable, 0, len(workbook.Sheets))
	var totalRows int64
	var totalCells int64
	for _, sheet := range workbook.Sheets {
		if sheet.RID == "" || sheet.Name == "" || sheet.State == "hidden" || sheet.State == "veryHidden" {
			continue
		}
		worksheetPath := relTargets[sheet.RID]
		if worksheetPath == "" {
			return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		worksheetBytes, ok := readZipFile(files, worksheetPath)
		if !ok {
			return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		var worksheet xlsxWorksheet
		if err := xml.Unmarshal(worksheetBytes, &worksheet); err != nil {
			return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		rows, rowErr := xlsxRowsToTable(worksheet.Rows, sharedStrings, importLimits)
		if rowErr != nil {
			return nil, rowErr
		}
		if len(rows) == 0 {
			continue
		}
		totalRows += int64(len(rows) - 1)
		totalCells += int64((len(rows) - 1) * len(rows[0]))
		if totalRows > importLimits.MaxRows {
			return nil, importSourceRejected("import_rows_exceeded", totalRows, importLimits.MaxRows)
		}
		if totalCells > importLimits.MaxCells {
			return nil, importSourceRejected("import_cells_exceeded", totalCells, importLimits.MaxCells)
		}
		tables = append(tables, xlsxDiscoveredTable{SheetName: sheet.Name, Rows: rows})
	}
	if len(tables) == 0 {
		return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
	return tables, nil
}

func parseXLSXTable(data []byte, importLimits config.ImportLimits, archiveLimits config.ArchiveLimits) ([][]tabularCell, *httpapi.APIError) {
	tables, apiErr := parseXLSXTables(data, importLimits, archiveLimits)
	if apiErr != nil {
		return nil, apiErr
	}
	return tables[0].Rows, nil
}

func parseXLSXSharedStrings(files map[string]*zip.File) ([]string, *httpapi.APIError) {
	data, ok := readZipFile(files, "xl/sharedStrings.xml")
	if !ok {
		return nil, nil
	}
	var shared xlsxSharedStrings
	if err := xml.Unmarshal(data, &shared); err != nil {
		return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
	values := make([]string, 0, len(shared.Items))
	for _, item := range shared.Items {
		values = append(values, xlsxStringValue(item.Text, item.Runs))
	}
	return values, nil
}

func xlsxRowsToTable(rows []xlsxRow, sharedStrings []string, importLimits config.ImportLimits) ([][]tabularCell, *httpapi.APIError) {
	maxRow := 0
	maxColumn := 0
	type keyedCell struct {
		row    int
		column int
		cell   tabularCell
	}
	cells := make([]keyedCell, 0)
	for _, row := range rows {
		rowRef := row.Ref
		for cellIndex, cell := range row.Cells {
			cellRow, cellColumn, ok := parseCellRef(cell.Ref)
			if !ok {
				cellRow = rowRef
				cellColumn = cellIndex + 1
			}
			if cellRow <= 0 {
				cellRow = rowRef
			}
			if cellRow <= 0 || cellColumn <= 0 {
				return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
			}
			value, apiErr := xlsxCellValue(cell, sharedStrings)
			if apiErr != nil {
				return nil, apiErr
			}
			if cellRow > maxRow {
				maxRow = cellRow
			}
			if cellColumn > maxColumn {
				maxColumn = cellColumn
			}
			cells = append(cells, keyedCell{row: cellRow, column: cellColumn, cell: value})
		}
	}
	if maxRow == 0 {
		return nil, nil
	}
	dataRows := maxRow - 1
	if int64(dataRows) > importLimits.MaxRows {
		return nil, importSourceRejected("import_rows_exceeded", int64(dataRows), importLimits.MaxRows)
	}
	if int64(maxColumn) > importLimits.MaxColumns {
		return nil, importSourceRejected("import_columns_exceeded", int64(maxColumn), importLimits.MaxColumns)
	}
	if int64(dataRows*maxColumn) > importLimits.MaxCells {
		return nil, importSourceRejected("import_cells_exceeded", int64(dataRows*maxColumn), importLimits.MaxCells)
	}
	table := make([][]tabularCell, maxRow)
	for rowIndex := range table {
		table[rowIndex] = make([]tabularCell, maxColumn)
		for columnIndex := range table[rowIndex] {
			table[rowIndex][columnIndex] = tabularCell{CellKind: "blank"}
		}
	}
	for _, keyed := range cells {
		table[keyed.row-1][keyed.column-1] = keyed.cell
	}
	return table, nil
}

func xlsxCellValue(cell xlsxCell, sharedStrings []string) (tabularCell, *httpapi.APIError) {
	if cell.Formula != nil && cell.Value == "" && cell.Type != "inlineStr" {
		return tabularCell{}, importSourceUnsupported("formula_cached_value_missing")
	}
	value := strings.TrimSpace(cell.Value)
	kind := "number"
	switch cell.Type {
	case "s":
		index, err := strconv.Atoi(value)
		if err != nil || index < 0 || index >= len(sharedStrings) {
			return tabularCell{}, importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		value = sharedStrings[index]
		kind = "string"
	case "inlineStr":
		value = xlsxStringValue(cell.InlineString.Text, cell.InlineString.Runs)
		kind = "string"
	case "str":
		kind = "string"
	case "b":
		if value == "1" {
			value = "true"
		} else if value == "0" {
			value = "false"
		}
		kind = "boolean"
	case "e":
		kind = "error_literal"
	case "":
		if value == "" {
			kind = "blank"
		}
	default:
		if value == "" {
			kind = "blank"
		} else {
			kind = "string"
		}
	}
	if cell.Formula != nil && kind != "blank" {
		kind = "formula_cached"
	}
	return tabularCell{DisplayText: value, CellKind: kind}, nil
}

func xlsxStringValue(text string, runs []xlsxStringRun) string {
	if len(runs) == 0 {
		return text
	}
	var builder strings.Builder
	for _, run := range runs {
		builder.WriteString(run.Text)
	}
	return builder.String()
}

func readZipFile(files map[string]*zip.File, name string) ([]byte, bool) {
	file, ok := files[name]
	if !ok {
		return nil, false
	}
	reader, err := file.Open()
	if err != nil {
		return nil, false
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, false
	}
	return data, true
}

func cleanXLSXPath(name string) (string, bool) {
	cleaned := path.Clean(strings.TrimPrefix(name, "/"))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", false
	}
	return cleaned, true
}

func resolveWorkbookRelationshipTarget(target string) (string, bool) {
	cleaned := path.Clean(strings.TrimPrefix(target, "/"))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", false
	}
	if strings.HasPrefix(cleaned, "xl/") {
		return cleaned, true
	}
	return path.Clean("xl/" + cleaned), true
}

func parseCellRef(ref string) (int, int, bool) {
	if ref == "" {
		return 0, 0, false
	}
	column := 0
	index := 0
	for index < len(ref) {
		ch := ref[index]
		if ch >= 'a' && ch <= 'z' {
			ch = ch - 'a' + 'A'
		}
		if ch < 'A' || ch > 'Z' {
			break
		}
		column = column*26 + int(ch-'A'+1)
		index++
	}
	if column == 0 || index == len(ref) {
		return 0, 0, false
	}
	row, err := strconv.Atoi(ref[index:])
	if err != nil || row <= 0 {
		return 0, 0, false
	}
	return row, column, true
}

func archiveCompressionRatioExceeded(extracted uint64, compressed uint64, maxRatio int64) bool {
	if extracted == 0 {
		return false
	}
	if compressed == 0 {
		return true
	}
	ratio := uint64(maxRatio)
	if compressed > math.MaxUint64/ratio {
		return false
	}
	return extracted > compressed*ratio
}

func int64MinUint64(value uint64) int64 {
	if value > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(value)
}
