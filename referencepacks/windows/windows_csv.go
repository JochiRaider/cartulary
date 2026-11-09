package producer

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

type windowsCSVRow struct {
	Fields map[string]string
	JSON   map[string][]map[string]any
	Raw    string
	Number int
}
type WindowsAssertion struct {
	SourceID   string `json:"source_id"`
	Locator    string `json:"locator"`
	SHA256     string `json:"fragment_sha256"`
	Role       string `json:"role"`
	CSVRecords []int  `json:"csv_records"`
}
type windowsCSVResult struct {
	Records      []WindowsRecord
	Dispositions []WindowsDisposition
	Issues       []WindowsError
	Assertions   []WindowsAssertion
}

func parseWindowsCSV(b []byte) ([]windowsCSVRow, error) {
	if !utf8.Valid(b) || bytes.HasPrefix(b, []byte{0xef, 0xbb, 0xbf}) || int64(len(b)) > windowsArtifactLimit {
		return nil, errors.New("strict bounded UTF-8 CSV without BOM required")
	}
	r := csv.NewReader(bytes.NewReader(b))
	header, e := r.Read()
	if e != nil {
		return nil, e
	}
	seen := map[string]bool{}
	for _, h := range header {
		if seen[h] {
			return nil, errors.New("duplicate CSV header")
		}
		seen[h] = true
	}
	var rows []windowsCSVRow
	keys := map[string]bool{}
	offset := r.InputOffset()
	for {
		values, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		row := windowsCSVRow{Fields: map[string]string{}, JSON: map[string][]map[string]any{}, Raw: string(b[offset:r.InputOffset()]), Number: len(rows) + 1}
		offset = r.InputOffset()
		for i, k := range header {
			row.Fields[k] = values[i]
		}
		raw, e := canonicaljson.Marshal(row.Fields)
		if e != nil {
			return nil, e
		}
		var check map[string]string
		if e = windowsDecode(raw, "csv_row", &check); e != nil {
			return nil, e
		}
		for k, v := range row.Fields {
			if !strings.HasSuffix(k, "_json") || v == "" {
				continue
			}
			var a []map[string]any
			if e = windowsDecode([]byte(v), k, &a); e != nil {
				return nil, fmt.Errorf("CSV record %d %s: %w", row.Number, k, e)
			}
			row.JSON[k] = a
		}
		id, e := numericSource(row.Fields["event_id"], 65535)
		if e != nil || id == nil {
			return nil, errors.New("CSV event ID must be explicit")
		}
		version, e := numericSource(row.Fields["event_version"], 255)
		if e != nil || version == nil {
			return nil, errors.New("CSV event version must be explicit")
		}
		kind := row.Fields["record_kind"]
		if kind != "event_reference" && kind != "source_boundary_marker" {
			return nil, errors.New("unknown CSV record kind")
		}
		key := kind + ":" + row.Fields["event_id"] + ":" + row.Fields["event_version"]
		if keys[key] {
			return nil, errors.New("duplicate CSV identity")
		}
		keys[key] = true
		rows = append(rows, row)
		if len(rows) > 10000 {
			return nil, errors.New("CSV record budget")
		}
	}
	if len(rows) == 0 {
		return nil, errors.New("empty CSV")
	}
	return rows, nil
}
func csvNumber(v any) int {
	if n, ok := v.(float64); ok {
		return int(n)
	}
	n, _ := v.(json.Number)
	i, _ := strconv.Atoi(string(n))
	return i
}
func csvString(v any) string { s, _ := v.(string); return s }
func csvSourceRange(v any) (int, int) {
	m, _ := v.(map[string]any)
	return csvNumber(m["body_child_start"]), csvNumber(m["body_child_end"])
}

var csvSheetRecord = regexp.MustCompile(`^xlsx:s([12]):r([1-9][0-9]*)$`)
var wordEventHeading = regexp.MustCompile(`^([0-9]{1,5})(\([^\r\n)]{0,32}\))?:`)
var wordVersions = regexp.MustCompile(`(?m)^\s*Event Versions:\s*([0-9]+(?:[ \t]*,[ \t]*[0-9]+)*)`)

func extractWindowsCSV(s WindowsSource, b []byte, sources map[string]WindowsAcquired, blobs map[string][]byte, bindings ...WindowsProviderBinding) (windowsCSVResult, error) {
	var out windowsCSVResult
	rows, e := parseWindowsCSV(b)
	if e != nil {
		return out, e
	}
	offices := map[string]windowsOffice{}
	officeIDs := map[string]string{}
	for id, a := range sources {
		if !windowsBinary(a.Source.Kind) {
			continue
		}
		if _, ok := offices[a.Source.Kind]; ok {
			return out, errors.New("at most one retained original per OOXML kind allowed in a CSV inventory")
		}
		o, e := readWindowsOffice(a.Source.Kind, blobs[id])
		if e != nil {
			return out, e
		}
		offices[a.Source.Kind] = o
		officeIDs[a.Source.Kind] = id
	}
	book, bok := offices["original_xlsx"]
	word, wok := offices["original_docx"]
	if !bok && !wok {
		return out, errors.New("CSV requires retained original evidence")
	}
	originalSections := map[int64][][2]int{}
	for i, n := range word.Body {
		p := xmlOne(n, "pPr")
		if p == nil {
			continue
		}
		style := xmlOne(p, "pStyle")
		text := strings.TrimSpace(officeText(n))
		if style == nil || style.Attr["val"] != "Heading3" || !wordEventHeading.MatchString(text) {
			continue
		}
		id, _ := strconv.ParseInt(wordEventHeading.FindStringSubmatch(text)[1], 10, 64)
		end := len(word.Body)
		for j := i + 1; j < len(word.Body); j++ {
			p := xmlOne(word.Body[j], "pPr")
			if p == nil {
				continue
			}
			s := xmlOne(p, "pStyle")
			if s != nil && slices.Contains([]string{"Heading1", "Heading2", "Heading3"}, s.Attr["val"]) {
				end = j
				break
			}
		}
		originalSections[id] = append(originalSections[id], [2]int{i + 1, end})
	}
	assertions := map[string]*WindowsAssertion{}
	sections := map[int]bool{}
	bind := func(source, loc, raw, role string, row int) {
		key := source + "\x00" + loc
		if a, ok := assertions[key]; ok {
			if role < a.Role {
				a.Role = role
			}
			if !slices.Contains(a.CSVRecords, row) {
				a.CSVRecords = append(a.CSVRecords, row)
			}
			return
		}
		assertions[key] = &WindowsAssertion{source, loc, Digest([]byte(raw)), role, []int{row}}
	}
	for _, row := range rows {
		loc := fmt.Sprintf("csv/record/%d", row.Number)
		id, _ := strconv.ParseInt(row.Fields["event_id"], 10, 64)
		version, _ := strconv.ParseInt(row.Fields["event_version"], 10, 64)
		issues := []string{}
		for _, meta := range row.JSON["source_artifact_metadata_json"] {
			kind := "original_docx"
			if strings.HasSuffix(csvString(meta["filename"]), ".xlsx") {
				kind = "original_xlsx"
			}
			if csvString(meta["sha256"]) != Digest(blobs[officeIDs[kind]]) {
				issues = append(issues, "original artifact digest differs")
			}
		}
		// Recover worksheet associations by exact ID/version, then compare every
		// literal message/OS cell. Recovery is separate evidence, never a CSV edit.
		messageRows := []int{}
		catalogRows := []int{}
		for n := range book.Sheets["Complete Event Messages"].Rows {
			sh := book.Sheets["Complete Event Messages"]
			if sh.Cells[fmt.Sprintf("A%d", n)] == row.Fields["event_id"] && sh.Cells[fmt.Sprintf("B%d", n)] == row.Fields["event_version"] {
				messageRows = append(messageRows, n)
			}
		}
		for n := range book.Sheets["Security Audit Events"].Rows {
			if book.Sheets["Security Audit Events"].Cells[fmt.Sprintf("C%d", n)] == row.Fields["event_id"] {
				catalogRows = append(catalogRows, n)
			}
		}
		slices.Sort(messageRows)
		slices.Sort(catalogRows)
		for _, n := range messageRows {
			sh := book.Sheets["Complete Event Messages"]
			if sh.Cells[fmt.Sprintf("C%d", n)] != row.Fields["xlsx_message_template"] || sh.Cells[fmt.Sprintf("D%d", n)] != row.Fields["xlsx_minimum_os_for_message_version"] {
				issues = append(issues, fmt.Sprintf("message/OS cells differ at row %d", n))
			}
			bind(officeIDs["original_xlsx"], fmt.Sprintf("ooxml/%s/row/%d", sh.Part, n), sh.Rows[n], "explicit_event_id_version_message", row.Number)
		}
		if row.Fields["xlsx_message_template"] != "" && len(messageRows) != 1 {
			issues = append(issues, "versioned message does not resolve uniquely")
		}
		for _, n := range catalogRows {
			sh := book.Sheets["Security Audit Events"]
			bind(officeIDs["original_xlsx"], fmt.Sprintf("ooxml/%s/row/%d", sh.Part, n), sh.Rows[n], "event_id_catalog_context", row.Number)
		}
		for _, a := range row.JSON["xlsx_catalog_rows_json"] {
			match := csvSheetRecord.FindStringSubmatch(csvString(a["record_id"]))
			if len(match) != 3 {
				issues = append(issues, "invalid catalog record locator")
				continue
			}
			n, _ := strconv.Atoi(match[2])
			sh := book.Sheets["Security Audit Events"]
			for field, col := range map[string]string{"category": "A", "subcategory": "B", "message_summary": "D", "minimum_os_requirement": "E"} {
				if csvString(a[field]) != sh.Cells[fmt.Sprintf("%s%d", col, n)] {
					issues = append(issues, "catalog field differs: "+field)
				}
			}
		}
		for _, a := range row.JSON["source_locators_json"] {
			if a["source"] == "xlsx" {
				m := csvSheetRecord.FindStringSubmatch(csvString(a["record_id"]))
				if len(m) != 3 {
					issues = append(issues, "invalid worksheet citation")
					continue
				}
				n, _ := strconv.Atoi(m[2])
				valid := slices.Contains(catalogRows, n)
				if m[1] == "2" {
					valid = slices.Contains(messageRows, n)
				}
				if !valid {
					issues = append(issues, "worksheet citation identifies another event/version")
				}
				continue
			}
			start, end := csvSourceRange(a)
			raw, _, e := word.fragment(start, end)
			if e != nil {
				issues = append(issues, e.Error())
				continue
			}
			bind(officeIDs["original_docx"], officeRangeLocator(start, end), raw, csvString(a["scope"]), row.Number)
		}
		// Check all nested Word field/table/link/image/sample locators as well as
		// the top-level sections. Their immutable raw ranges remain resolvable.
		for column, items := range row.JSON {
			if !strings.HasPrefix(column, "docx_") {
				continue
			}
			for _, a := range items {
				if a["source"] == nil {
					continue
				}
				start, end := csvSourceRange(a["source"])
				raw, text, e := word.fragment(start, end)
				if e != nil {
					issues = append(issues, e.Error())
					continue
				}
				bind(officeIDs["original_docx"], officeRangeLocator(start, end), raw, column, row.Number)
				formatted := officeReferenceText(word.Body[start-1 : end])
				if column == "docx_sample_metadata_for_row_version_json" || column == "docx_other_sample_metadata_json" {
					if e := verifyCSVSample(text, a); e != nil {
						issues = append(issues, e.Error())
					}
				}
				if column == "docx_section_contexts_json" {
					sections[start] = true
					if !strings.HasPrefix(strings.TrimSpace(text), strings.TrimSpace(csvString(a["heading"]))) {
						issues = append(issues, "section heading differs")
					}
				}
				if column == "docx_field_definitions_json" {
					for _, key := range []string{"label", "type_as_documented", "definition"} {
						if value := csvString(a[key]); value != "" && !strings.Contains(officeComparable(text), officeComparable(value)) {
							issues = append(issues, "field text differs from original: "+key)
						}
					}
					if values, ok := a["continuation_text"].([]any); ok {
						for _, v := range values {
							if !strings.Contains(officeComparable(formatted), officeComparable(csvString(v))) {
								issues = append(issues, "field continuation differs from original")
							}
						}
					}
				}
				if column == "docx_subcategory_guidance_json" && !strings.Contains(officeComparable(formatted), officeComparable(csvString(a["text"]))) {
					issues = append(issues, "subcategory guidance differs from original")
				}
				if column == "docx_reference_tables_json" {
					if e := verifyOfficeTable(word.Body[start-1:end], a); e != nil {
						issues = append(issues, e.Error())
					}
				}
				if column == "docx_image_references_json" {
					if parts, ok := a["media_parts"].([]any); ok {
						for _, part := range parts {
							if _, ok := word.Members[csvString(part)]; !ok {
								issues = append(issues, "image member absent")
							}
						}
					}
				}
			}
		}
		r := WindowsRecord{SourceID: s.ID, Locator: loc, FragmentSHA256: Digest([]byte(row.Raw)), EventID: id, Version: &version, Kind: "combined_reference", Context: "Supplied combined reference; available original-source provenance audited independently. Event-level explanations are not a per-version payload schema.", Documentation: csvDocumentation(row)}
		providers := map[string]string{}
		matching := []*windowsXML{}
		versionSupported := len(messageRows) == 1
		for _, span := range originalSections[id] {
			start, end := span[0], span[1]
			raw, text, e := word.fragment(start, end)
			if e != nil {
				continue
			}
			sections[start] = true
			bind(officeIDs["original_docx"], officeRangeLocator(start, end), raw, "recovered_event_section", row.Number)
			for _, m := range wordVersions.FindAllStringSubmatch(text, -1) {
				for _, v := range strings.Split(m[1], ",") {
					if strings.TrimSpace(v) == row.Fields["event_version"] {
						versionSupported = true
					}
				}
			}
			system, e := windowsSampleSystem(text)
			if e != nil {
				continue
			}
			event := xmlOne(system, "EventID")
			provider := xmlOne(system, "Provider")
			if event == nil || provider == nil {
				continue
			}
			n, e := numericSource(strings.TrimSpace(event.Text), 65535)
			if e != nil || n == nil || *n != id {
				continue
			}
			p, ok := windowsProvider(provider.Attr["Name"])
			if !ok {
				continue
			}
			guid := provider.Attr["Guid"]
			if previous := providers[p]; previous != "" && guid != "" && !strings.EqualFold(previous, guid) {
				issues = append(issues, "contradictory provider GUID in original samples")
			}
			if guid != "" || providers[p] == "" {
				providers[p] = guid
			}
			v := xmlOne(system, "Version")
			if v != nil && strings.TrimSpace(v.Text) == row.Fields["event_version"] {
				matching = append(matching, system)
				versionSupported = true
			}
		}
		if len(providers) == 1 {
			for p, g := range providers {
				r.Provider = p
				r.GUID, e = sourceGUID(g)
				if e != nil {
					issues = append(issues, "invalid provider GUID")
				}
			}
		}
		if extracted := row.Fields["provider_names_from_docx_examples"]; extracted != "" && extracted != r.Provider {
			issues = append(issues, "extracted provider differs from original matching sample")
		}
		for _, binding := range bindings {
			if binding.CSVSourceID != s.ID || binding.CSVLocator != loc {
				continue
			}
			provider, guid, err := providerFromWindowsBinding(row, binding, sources, blobs)
			if err != nil {
				issues = append(issues, "reviewed provider binding: "+err.Error())
				continue
			}
			if len(providers) > 1 || r.Provider != "" && r.Provider != provider || r.GUID != nil && guid != nil && *r.GUID != *guid {
				issues = append(issues, "reviewed provider binding conflicts with original sample")
				continue
			}
			r.Provider = provider
			versionSupported = true // The selected retained System explicitly names this version.
			if r.GUID == nil {
				r.GUID = guid
			}
			fragment, _ := windowsFragment(sources[binding.SourceID].Source, blobs[binding.SourceID], binding.Locator)
			bind(binding.SourceID, binding.Locator, fragment, "reviewed_provider_system", row.Number)
			bind(binding.SourceID, "document", string(blobs[binding.SourceID]), "reviewed_provider_document", row.Number)
			r.Context += " Provider identity supplied by one reviewed retained definition; see the archive's assertion index."
		}
		if !versionSupported {
			issues = append(issues, "event version lacks original worksheet, declaration or matching sample evidence")
		}
		for _, system := range matching {
			for name, dst := range map[string]**int64{"Level": &r.Level, "Task": &r.Task, "Opcode": &r.Opcode} {
				n := xmlOne(system, name)
				if n == nil {
					continue
				}
				max := int64(255)
				if name == "Task" {
					max = 65535
				}
				value, e := numericSource(strings.TrimSpace(n.Text), max)
				if e != nil {
					issues = append(issues, "invalid sample descriptor")
					continue
				}
				if *dst != nil && value != nil && **dst != *value {
					issues = append(issues, "contradictory sample descriptor")
				}
				*dst = value
			}
		}
		if row.Fields["record_kind"] == "source_boundary_marker" {
			if id != 8191 || strings.TrimSpace(row.Fields["xlsx_message_template"]) != "Highest System-Defined Audit Message Value." {
				issues = append(issues, "unrecognized boundary marker")
			}
			out.Dispositions = append(out.Dispositions, WindowsDisposition{s.ID, loc, "excluded", "source_boundary_marker_not_event", nil})
		} else if r.Provider == "" {
			out.Issues = append(out.Issues, WindowsError{"identity_conflict", s.ID + "/" + loc, fmt.Sprintf("event %d version %d: missing an explicit canonical Provider Name for this definition in retained original evidence, or one explicitly reviewed provider-qualified retained definition; ID/channel/category alone and samples naming other event IDs cannot supply it", id, version)})
			out.Dispositions = append(out.Dispositions, WindowsDisposition{s.ID, loc, "blocking", "provider_evidence_missing", nil})
		} else {
			out.Records = append(out.Records, r)
		}
		for _, issue := range issues {
			out.Issues = append(out.Issues, WindowsError{"source_accounting_failed", s.ID + "/" + loc, issue})
		}
	}
	for _, a := range assertions {
		out.Assertions = append(out.Assertions, *a)
	}
	slices.SortFunc(out.Assertions, func(a, b WindowsAssertion) int {
		return strings.Compare(a.SourceID+"/"+a.Locator, b.SourceID+"/"+b.Locator)
	})
	// Every literal source row must participate, not just the rows the CSV
	// happened to cite. Header rows are not event assertions.
	for _, name := range []string{"Security Audit Events", "Complete Event Messages"} {
		sh := book.Sheets[name]
		for n := range sh.Rows {
			if n == 1 {
				continue
			}
			loc := fmt.Sprintf("ooxml/%s/row/%d", sh.Part, n)
			if assertions[officeIDs["original_xlsx"]+"\x00"+loc] == nil {
				out.Issues = append(out.Issues, WindowsError{"source_accounting_failed", officeIDs["original_xlsx"], "unaccounted worksheet row: " + loc})
			}
		}
	}
	for i, n := range word.Body {
		props := xmlOne(n, "pPr")
		if props == nil {
			continue
		}
		style := xmlOne(props, "pStyle")
		if style != nil && style.Attr["val"] == "Heading3" && wordEventHeading.MatchString(strings.TrimSpace(officeText(n))) && !sections[i+1] {
			out.Issues = append(out.Issues, WindowsError{"source_accounting_failed", officeIDs["original_docx"], fmt.Sprintf("unaccounted Word event section at body child %d", i+1)})
		}
	}
	return out, nil
}

func officeComparable(s string) string { return strings.Join(strings.Fields(s), " ") }

// A damaged payload does not invalidate a complete, independently parsed System
// fragment. Extracted sample metadata must still agree with that fragment;
// documentary metadata never becomes an incident observation.
func verifyCSVSample(text string, sample map[string]any) error {
	system, e := windowsSampleSystem(text)
	if e != nil {
		return errors.New("sample System fragment cannot be independently verified")
	}
	for key, name := range map[string]string{"event_id": "EventID", "event_version": "Version", "level": "Level", "task": "Task", "opcode": "Opcode", "channel": "Channel", "keywords": "Keywords"} {
		if sample[key] == nil {
			continue
		}
		n := xmlOne(system, name)
		if n == nil || strings.TrimSpace(n.Text) != fmt.Sprint(sample[key]) {
			return fmt.Errorf("sample metadata differs from original: %s", key)
		}
	}
	p := xmlOne(system, "Provider")
	for key, attr := range map[string]string{"provider_name": "Name", "provider_guid": "Guid"} {
		if sample[key] == nil {
			continue
		}
		if p == nil || !strings.EqualFold(p.Attr[attr], csvString(sample[key])) {
			return fmt.Errorf("sample metadata differs from original: %s", key)
		}
	}
	return nil
}
func verifyOfficeTable(nodes []*windowsXML, table map[string]any) error {
	var tables []*windowsXML
	for _, n := range nodes {
		if n.Name == "tbl" {
			tables = append(tables, n)
		} else {
			tables = append(tables, xmlDescendants(n, "tbl")...)
		}
	}
	if len(tables) != 1 {
		return errors.New("table locator must identify one original table")
	}
	rows, ok := table["rows"].([]any)
	actual := xmlChildren(tables[0], "tr")
	if !ok || len(actual) != len(rows) {
		return errors.New("original table row count differs")
	}
	for i, r := range rows {
		cells, ok := r.([]any)
		got := xmlChildren(actual[i], "tc")
		if !ok || len(cells) != len(got) {
			return errors.New("original table cell count differs")
		}
		for j, v := range cells {
			cell, _ := v.(map[string]any)
			if officeComparable(csvString(cell["text"])) != officeComparable(officeText(got[j])) {
				return errors.New("original table cell text differs")
			}
			span := 1
			var merge any
			if p := xmlOne(got[j], "tcPr"); p != nil {
				if n := xmlOne(p, "gridSpan"); n != nil {
					span, _ = strconv.Atoi(n.Attr["val"])
				}
				if n := xmlOne(p, "vMerge"); n != nil {
					value := n.Attr["val"]
					if value == "" {
						value = "continue"
					}
					merge = value
				}
			}
			if csvNumber(cell["grid_span"]) != span || cell["vertical_merge"] != merge {
				return errors.New("original table merge metadata differs")
			}
		}
	}
	return nil
}

func windowsSampleSystem(text string) (*windowsXML, error) {
	start := strings.Index(text, "<System>")
	end := strings.Index(text, "</System>")
	if start < 0 || end < start || strings.Count(text, "<System>") != 1 || strings.Count(text, "</System>") != 1 {
		return nil, errors.New("one literal System fragment required")
	}
	system, e := parseWindowsXML([]byte(text[start : end+len("</System>")]))
	if e != nil {
		return nil, e
	}
	for _, field := range []string{"Provider", "EventID", "Version", "Level", "Task", "Opcode", "Channel", "Keywords"} {
		if len(xmlChildren(system, field)) > 1 {
			return nil, errors.New("ambiguous sample System field")
		}
	}
	return system, nil
}
