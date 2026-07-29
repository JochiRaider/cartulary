package imports

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"math"
	"path"
	"sort"
	"strconv"
	"strings"

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
	Sheets       []xlsxSheet       `xml:"sheets>sheet"`
	DefinedNames []xlsxDefinedName `xml:"definedNames>definedName"`
	Protection   *struct{}         `xml:"workbookProtection"`
}

type xlsxDefinedName struct {
	Name         string `xml:"name,attr"`
	LocalSheetID *int   `xml:"localSheetId,attr"`
	Formula      string `xml:",chardata"`
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
	ID         string `xml:"Id,attr"`
	Type       string `xml:"Type,attr"`
	Target     string `xml:"Target,attr"`
	TargetMode string `xml:"TargetMode,attr"`
}

type xlsxWorksheet struct {
	Rows       []xlsxRow       `xml:"sheetData>row"`
	Columns    []xlsxColumn    `xml:"cols>col"`
	Merges     []xlsxMergeCell `xml:"mergeCells>mergeCell"`
	TableParts []xlsxTablePart `xml:"tableParts>tablePart"`
	AutoFilter *struct{}       `xml:"autoFilter"`
	SortState  *struct{}       `xml:"sortState"`
	Protection *struct{}       `xml:"sheetProtection"`
}

type xlsxColumn struct {
	Hidden bool `xml:"hidden,attr"`
}

type xlsxRow struct {
	Ref    int        `xml:"r,attr"`
	Hidden bool       `xml:"hidden,attr"`
	Cells  []xlsxCell `xml:"c"`
}

type xlsxCell struct {
	Ref          string           `xml:"r,attr"`
	Type         string           `xml:"t,attr"`
	Style        string           `xml:"s,attr"`
	Formula      *string          `xml:"f"`
	Value        *string          `xml:"v"`
	InlineString xlsxInlineString `xml:"is"`
}

type xlsxInlineString struct {
	Text string          `xml:"t"`
	Runs []xlsxStringRun `xml:"r"`
}

type xlsxMergeCell struct {
	Ref string `xml:"ref,attr"`
}

type xlsxTablePart struct {
	RID string `xml:"id,attr"`
}

type xlsxTable struct {
	Name        string `xml:"name,attr"`
	DisplayName string `xml:"displayName,attr"`
	Ref         string `xml:"ref,attr"`
}

type xlsxCellPosition struct {
	row    int
	column int
}

type xlsxSheetIndex struct {
	name         string
	order        int
	path         string
	cells        map[xlsxCellPosition]xlsxCell
	usedRange    *sourceRectangle
	mergedRanges []sourceRectangle
	tables       []xlsxLocatedRange
	warnings     map[string]struct{}
}

type xlsxLocatedRange struct {
	kind        string
	sheet       *xlsxSheetIndex
	rect        sourceRectangle
	definedName string
	tableName   string
}

type xlsxWorkbookIndex struct {
	sheets       []*xlsxSheetIndex
	sheetsByName map[string]*xlsxSheetIndex
	ranges       []xlsxLocatedRange
	shared       []string
	warnings     map[string]struct{}
}

type xlsxDecodedRectangle struct {
	rows                   [][]tabularCell
	warningCodes           []string
	blockingColumnOrdinals []int
}

func indexXLSXWorkbook(data []byte, importLimits Limits, archiveLimits ArchiveLimits) (*xlsxWorkbookIndex, *httpapi.APIError) {
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
	warnings := make(map[string]struct{})
	for _, rel := range rels.Relationships {
		if strings.EqualFold(rel.TargetMode, "External") {
			warnings["external_links_ignored"] = struct{}{}
			continue
		}
		if rel.ID == "" || rel.Target == "" {
			continue
		}
		if target, targetOK := resolveWorkbookRelationshipTarget(rel.Target); targetOK {
			relTargets[rel.ID] = target
		}
	}
	for name := range files {
		switch {
		case strings.HasPrefix(name, "xl/externalLinks/"):
			warnings["external_links_ignored"] = struct{}{}
		case strings.HasPrefix(name, "xl/charts/"):
			warnings["chart_ignored"] = struct{}{}
		case strings.HasPrefix(name, "xl/pivot"):
			warnings["pivot_ignored"] = struct{}{}
		}
	}
	if workbook.Protection != nil {
		warnings["workbook_protection_metadata_only"] = struct{}{}
	}

	sharedStrings, apiErr := parseXLSXSharedStrings(files)
	if apiErr != nil {
		return nil, apiErr
	}
	index := &xlsxWorkbookIndex{
		sheetsByName: make(map[string]*xlsxSheetIndex, len(workbook.Sheets)),
		shared:       sharedStrings,
		warnings:     warnings,
	}
	for sheetOrder, sheet := range workbook.Sheets {
		if sheet.RID == "" || sheet.Name == "" {
			return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
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
		sheetIndex := &xlsxSheetIndex{
			name:     sheet.Name,
			order:    sheetOrder,
			path:     worksheetPath,
			cells:    make(map[xlsxCellPosition]xlsxCell),
			warnings: make(map[string]struct{}),
		}
		if sheet.State == "hidden" || sheet.State == "veryHidden" {
			sheetIndex.warnings["filtered_or_hidden_state_ignored"] = struct{}{}
		}
		if worksheet.AutoFilter != nil || worksheet.SortState != nil {
			sheetIndex.warnings["filtered_or_hidden_state_ignored"] = struct{}{}
		}
		if worksheet.Protection != nil {
			sheetIndex.warnings["sheet_protection_metadata_only"] = struct{}{}
		}
		for _, column := range worksheet.Columns {
			if column.Hidden {
				sheetIndex.warnings["filtered_or_hidden_state_ignored"] = struct{}{}
			}
		}
		for _, row := range worksheet.Rows {
			if row.Hidden {
				sheetIndex.warnings["filtered_or_hidden_state_ignored"] = struct{}{}
			}
			for cellIndex, cell := range row.Cells {
				cellRow, cellColumn, refOK := parseCellRef(cell.Ref)
				if !refOK {
					cellRow = row.Ref
					cellColumn = cellIndex + 1
				}
				if cellRow <= 0 || cellColumn <= 0 {
					return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
				}
				if !xlsxCellHasSemanticContent(cell) {
					continue
				}
				position := xlsxCellPosition{row: cellRow, column: cellColumn}
				sheetIndex.cells[position] = cell
				sheetIndex.usedRange = expandSourceRectangle(sheetIndex.usedRange, position)
			}
		}
		for _, merge := range worksheet.Merges {
			rect, rectErr := parseSourceRectangle(strings.ToUpper(strings.ReplaceAll(merge.Ref, "$", "")))
			if rectErr != nil {
				return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
			}
			sheetIndex.mergedRanges = append(sheetIndex.mergedRanges, rect)
			sheetIndex.warnings["merged_layout_downgraded"] = struct{}{}
		}
		if apiErr := indexWorksheetTables(files, worksheet, sheetIndex); apiErr != nil {
			return nil, apiErr
		}
		for _, table := range sheetIndex.tables {
			sheetIndex.usedRange = expandRectangle(sheetIndex.usedRange, table.rect)
		}
		index.sheets = append(index.sheets, sheetIndex)
		index.sheetsByName[sheet.Name] = sheetIndex
	}

	for _, sheet := range index.sheets {
		if sheet.usedRange == nil {
			continue
		}
		if apiErr := validateXLSXRectangle(*sheet.usedRange, importLimits); apiErr != nil {
			return nil, apiErr
		}
		index.ranges = append(index.ranges, xlsxLocatedRange{kind: "xlsx_used_range", sheet: sheet, rect: *sheet.usedRange})
		index.ranges = append(index.ranges, sheet.tables...)
	}
	for _, defined := range workbook.DefinedNames {
		if strings.HasPrefix(strings.ToLower(defined.Name), "_xlnm.") {
			continue
		}
		sheet, rect, parseOK := parseStaticDefinedName(defined, index.sheets, index.sheetsByName)
		if !parseOK {
			return nil, importSourceUnsupported("unsupported_named_range")
		}
		if apiErr := validateXLSXRectangle(rect, importLimits); apiErr != nil {
			return nil, apiErr
		}
		index.ranges = append(index.ranges, xlsxLocatedRange{
			kind:        "xlsx_named_range",
			sheet:       sheet,
			rect:        rect,
			definedName: defined.Name,
		})
	}
	sort.SliceStable(index.ranges, func(left, right int) bool {
		a, b := index.ranges[left], index.ranges[right]
		if a.sheet.order != b.sheet.order {
			return a.sheet.order < b.sheet.order
		}
		if a.rect.top != b.rect.top {
			return a.rect.top < b.rect.top
		}
		if a.rect.left != b.rect.left {
			return a.rect.left < b.rect.left
		}
		if a.rect.bottom != b.rect.bottom {
			return a.rect.bottom < b.rect.bottom
		}
		if a.rect.right != b.rect.right {
			return a.rect.right < b.rect.right
		}
		if xlsxRangeKindOrder(a.kind) != xlsxRangeKindOrder(b.kind) {
			return xlsxRangeKindOrder(a.kind) < xlsxRangeKindOrder(b.kind)
		}
		if a.tableName != b.tableName {
			return a.tableName < b.tableName
		}
		return a.definedName < b.definedName
	})
	if len(index.ranges) == 0 {
		return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
	return index, nil
}

func indexWorksheetTables(files map[string]*zip.File, worksheet xlsxWorksheet, sheet *xlsxSheetIndex) *httpapi.APIError {
	if len(worksheet.TableParts) == 0 {
		return nil
	}
	relsPath := path.Join(path.Dir(sheet.path), "_rels", path.Base(sheet.path)+".rels")
	relsBytes, ok := readZipFile(files, relsPath)
	if !ok {
		return importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
	var rels xlsxRelationships
	if err := xml.Unmarshal(relsBytes, &rels); err != nil {
		return importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
	relationships := make(map[string]xlsxRelationship, len(rels.Relationships))
	for _, rel := range rels.Relationships {
		relationships[rel.ID] = rel
	}
	for _, part := range worksheet.TableParts {
		rel, ok := relationships[part.RID]
		if !ok || strings.EqualFold(rel.TargetMode, "External") {
			return importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		target, ok := resolvePartRelationshipTarget(sheet.path, rel.Target)
		if !ok {
			return importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		tableBytes, ok := readZipFile(files, target)
		if !ok {
			return importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		var table xlsxTable
		if err := xml.Unmarshal(tableBytes, &table); err != nil {
			return importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		name := table.DisplayName
		if name == "" {
			name = table.Name
		}
		rect, err := parseSourceRectangle(strings.ToUpper(strings.ReplaceAll(table.Ref, "$", "")))
		if err != nil || name == "" {
			return importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		sheet.tables = append(sheet.tables, xlsxLocatedRange{
			kind:      "xlsx_table",
			sheet:     sheet,
			rect:      rect,
			tableName: name,
		})
	}
	return nil
}

func (index *xlsxWorkbookIndex) decodeRectangle(sheet *xlsxSheetIndex, rect sourceRectangle, limits Limits) (xlsxDecodedRectangle, *httpapi.APIError) {
	if sheet == nil {
		return xlsxDecodedRectangle{}, importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
	if apiErr := validateXLSXRectangle(rect, limits); apiErr != nil {
		return xlsxDecodedRectangle{}, apiErr
	}
	height := rect.bottom - rect.top + 1
	width := rect.right - rect.left + 1
	rows := make([][]tabularCell, height)
	warnings := cloneStringSet(index.warnings)
	for warning := range sheet.warnings {
		warnings[warning] = struct{}{}
	}
	blocked := make(map[int]struct{})
	for rowOffset := 0; rowOffset < height; rowOffset++ {
		rows[rowOffset] = make([]tabularCell, width)
		for columnOffset := 0; columnOffset < width; columnOffset++ {
			rows[rowOffset][columnOffset] = tabularCell{CellKind: "blank"}
			position := xlsxCellPosition{row: rect.top + rowOffset, column: rect.left + columnOffset}
			cell, ok := sheet.cells[position]
			if !ok || mergedCellIsCovered(sheet.mergedRanges, position) {
				continue
			}
			value, missingCache, apiErr := xlsxCellValue(cell, index.shared)
			if apiErr != nil {
				return xlsxDecodedRectangle{}, apiErr
			}
			rows[rowOffset][columnOffset] = value
			if cell.Formula != nil {
				if missingCache {
					warnings["formula_cached_value_missing"] = struct{}{}
					blocked[columnOffset+1] = struct{}{}
				} else {
					warnings["formula_inert"] = struct{}{}
				}
			}
		}
	}
	return xlsxDecodedRectangle{
		rows:                   rows,
		warningCodes:           sortedStringSet(warnings),
		blockingColumnOrdinals: sortedIntSet(blocked),
	}, nil
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

func xlsxCellHasSemanticContent(cell xlsxCell) bool {
	return cell.Formula != nil || cell.Value != nil || cell.Type == "inlineStr"
}

func xlsxCellValue(cell xlsxCell, sharedStrings []string) (tabularCell, bool, *httpapi.APIError) {
	if cell.Formula != nil && cell.Value == nil && cell.Type != "inlineStr" {
		return tabularCell{CellKind: "blank"}, true, nil
	}
	rawValue := ""
	if cell.Value != nil {
		rawValue = *cell.Value
	}
	value := strings.TrimSpace(rawValue)
	kind := "number"
	switch cell.Type {
	case "s":
		index, err := strconv.Atoi(value)
		if err != nil || index < 0 || index >= len(sharedStrings) {
			return tabularCell{}, false, importSourceUnsupported("encrypted_or_unparseable_workbook")
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
	if cell.Formula != nil {
		kind = "formula_cached"
	}
	return tabularCell{DisplayText: value, CellKind: kind}, false, nil
}

func parseStaticDefinedName(
	defined xlsxDefinedName,
	sheets []*xlsxSheetIndex,
	byName map[string]*xlsxSheetIndex,
) (*xlsxSheetIndex, sourceRectangle, bool) {
	if strings.TrimSpace(defined.Name) == "" {
		return nil, sourceRectangle{}, false
	}
	formula := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(defined.Formula), "="))
	if formula == "" || strings.ContainsAny(formula, "[],;()") {
		return nil, sourceRectangle{}, false
	}
	bang := strings.LastIndex(formula, "!")
	if bang <= 0 || bang == len(formula)-1 {
		return nil, sourceRectangle{}, false
	}
	sheetName := strings.TrimSpace(formula[:bang])
	if strings.HasPrefix(sheetName, "'") && strings.HasSuffix(sheetName, "'") && len(sheetName) >= 2 {
		sheetName = strings.ReplaceAll(sheetName[1:len(sheetName)-1], "''", "'")
	}
	sheet := byName[sheetName]
	if sheet == nil {
		return nil, sourceRectangle{}, false
	}
	if defined.LocalSheetID != nil {
		if *defined.LocalSheetID < 0 || *defined.LocalSheetID >= len(sheets) || sheets[*defined.LocalSheetID] != sheet {
			return nil, sourceRectangle{}, false
		}
	}
	ref := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(formula[bang+1:]), "$", ""))
	if !strings.Contains(ref, ":") {
		ref += ":" + ref
	}
	rect, err := parseSourceRectangle(ref)
	if err != nil {
		return nil, sourceRectangle{}, false
	}
	return sheet, rect, true
}

func validateXLSXRectangle(rect sourceRectangle, limits Limits) *httpapi.APIError {
	width := int64(rect.right - rect.left + 1)
	height := int64(rect.bottom - rect.top + 1)
	dataRows := height - 1
	if dataRows < 0 {
		dataRows = 0
	}
	if width > limits.MaxColumns {
		return importSourceRejected("import_columns_exceeded", width, limits.MaxColumns)
	}
	if dataRows > limits.MaxRows {
		return importSourceRejected("import_rows_exceeded", dataRows, limits.MaxRows)
	}
	if dataRows != 0 && width > math.MaxInt64/dataRows {
		return importSourceRejected("import_cells_exceeded", math.MaxInt64, limits.MaxCells)
	}
	if dataRows*width > limits.MaxCells {
		return importSourceRejected("import_cells_exceeded", dataRows*width, limits.MaxCells)
	}
	return nil
}

func expandSourceRectangle(existing *sourceRectangle, position xlsxCellPosition) *sourceRectangle {
	return expandRectangle(existing, sourceRectangle{
		left: position.column, top: position.row, right: position.column, bottom: position.row,
	})
}

func expandRectangle(existing *sourceRectangle, next sourceRectangle) *sourceRectangle {
	if existing == nil {
		copy := next
		return &copy
	}
	if next.left < existing.left {
		existing.left = next.left
	}
	if next.top < existing.top {
		existing.top = next.top
	}
	if next.right > existing.right {
		existing.right = next.right
	}
	if next.bottom > existing.bottom {
		existing.bottom = next.bottom
	}
	return existing
}

func mergedCellIsCovered(merges []sourceRectangle, position xlsxCellPosition) bool {
	for _, merge := range merges {
		if position.column >= merge.left && position.column <= merge.right &&
			position.row >= merge.top && position.row <= merge.bottom {
			return position.column != merge.left || position.row != merge.top
		}
	}
	return false
}

func xlsxRangeKindOrder(kind string) int {
	switch kind {
	case "xlsx_used_range":
		return 0
	case "xlsx_table":
		return 1
	default:
		return 2
	}
}

func sourceRectangleA1(rect sourceRectangle) string {
	return sourceColumnLetters(rect.left) + strconv.Itoa(rect.top) + ":" + sourceColumnLetters(rect.right) + strconv.Itoa(rect.bottom)
}

func sourceColumnLetters(column int) string {
	if column <= 0 {
		return ""
	}
	var result []byte
	for column > 0 {
		column--
		result = append(result, byte('A'+column%26))
		column /= 26
	}
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return string(result)
}

func cloneStringSet(source map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{}, len(source))
	for value := range source {
		result[value] = struct{}{}
	}
	return result
}

func sortedStringSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func sortedIntSet(values map[int]struct{}) []int {
	result := make([]int, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Ints(result)
	return result
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

func resolvePartRelationshipTarget(sourcePart string, target string) (string, bool) {
	if strings.HasPrefix(target, "/") {
		return cleanXLSXPath(target)
	}
	return cleanXLSXPath(path.Join(path.Dir(sourcePart), target))
}

func parseCellRef(ref string) (int, int, bool) {
	if ref == "" {
		return 0, 0, false
	}
	ref = strings.ReplaceAll(ref, "$", "")
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
