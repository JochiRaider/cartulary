package producer

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/text/unicode/norm"
)

type windowsXML struct {
	Name     string
	Attr     map[string]string
	Children []*windowsXML
	Text     string
	Raw      string
}

func parseWindowsXML(b []byte) (*windowsXML, error) {
	if !utf8.Valid(b) {
		return nil, errors.New("invalid XML encoding")
	}
	d := xml.NewDecoder(bytes.NewReader(b))
	var root *windowsXML
	stack := []*windowsXML{}
	offsets := []int64{}
	for {
		start := d.InputOffset()
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := token.(type) {
		case xml.Directive:
			return nil, errors.New("XML directives, DTDs and entities are prohibited")
		case xml.StartElement:
			n := &windowsXML{Name: t.Name.Local, Attr: map[string]string{}}
			for _, a := range t.Attr {
				if _, exists := n.Attr[a.Name.Local]; exists {
					return nil, errors.New("duplicate XML attribute")
				}
				n.Attr[a.Name.Local] = a.Value
			}
			if len(stack) == 0 {
				if root != nil {
					return nil, errors.New("multiple XML roots")
				}
				root = n
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, n)
			}
			stack = append(stack, n)
			offsets = append(offsets, start)
			if len(stack) > 128 {
				return nil, errors.New("XML nesting limit")
			}
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, errors.New("unexpected XML end")
			}
			i := len(stack) - 1
			stack[i].Raw = string(b[offsets[i]:d.InputOffset()])
			stack = stack[:i]
			offsets = offsets[:i]
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].Text += string(t)
			} else if strings.TrimSpace(string(t)) != "" {
				return nil, errors.New("text outside XML root")
			}
		}
	}
	if root == nil || len(stack) != 0 {
		return nil, errors.New("incomplete XML")
	}
	return root, nil
}
func xmlChildren(n *windowsXML, name string) []*windowsXML {
	out := []*windowsXML{}
	if n != nil {
		for _, c := range n.Children {
			if c.Name == name {
				out = append(out, c)
			}
		}
	}
	return out
}
func xmlDescendants(n *windowsXML, name string) []*windowsXML {
	out := []*windowsXML{}
	if n == nil {
		return out
	}
	if n.Name == name {
		out = append(out, n)
	}
	for _, c := range n.Children {
		out = append(out, xmlDescendants(c, name)...)
	}
	return out
}
func xmlOne(n *windowsXML, name string) *windowsXML {
	v := xmlChildren(n, name)
	if len(v) == 1 {
		return v[0]
	}
	return nil
}
func numericSource(value string, max int64) (*int64, error) {
	if value == "" {
		return nil, nil
	}
	base := 10
	if strings.HasPrefix(value, "0x") {
		base = 16
		value = value[2:]
	}
	n, err := strconv.ParseInt(value, base, 64)
	if err != nil || n < 0 || n > max {
		return nil, errors.New("numeric source value outside domain")
	}
	return &n, nil
}
func sourceGUID(value string) (*string, error) {
	if value == "" {
		return nil, nil
	}
	v := strings.ToLower(strings.Trim(value, "{}"))
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).MatchString(v) {
		return nil, errors.New("invalid provider GUID")
	}
	return &v, nil
}
func windowsTrim(r rune) bool {
	return (r >= 9 && r <= 13) || r == 0x20 || r == 0x85 || r == 0xa0 || r == 0x1680 || (r >= 0x2000 && r <= 0x200a) || r == 0x2028 || r == 0x2029 || r == 0x202f || r == 0x205f || r == 0x3000
}
func windowsText(s string) string {
	return strings.TrimFunc(norm.NFC.String(strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")), windowsTrim)
}
func windowsXMLText(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

func extractWindowsManifest(s WindowsSource, b []byte) ([]WindowsRecord, []WindowsDisposition, error) {
	root, err := parseWindowsXML(b)
	if err != nil {
		return nil, nil, err
	}
	stringsByID := map[string]string{}
	for _, resources := range xmlDescendants(root, "resources") {
		if resources.Attr["culture"] != "en-US" {
			continue
		}
		for _, str := range xmlDescendants(resources, "string") {
			id, value := str.Attr["id"], str.Attr["value"]
			if old, ok := stringsByID[id]; ok && old != value {
				return nil, nil, errors.New("conflicting localized string")
			}
			stringsByID[id] = value
		}
	}
	records := []WindowsRecord{}
	dispositions := []WindowsDisposition{}
	for pi, p := range xmlDescendants(root, "provider") {
		if len(xmlChildren(xmlOne(p, "events"), "event")) == 0 {
			dispositions = append(dispositions, WindowsDisposition{s.ID, fmt.Sprintf("xml/provider/%d", pi), "excluded", "non_event_provider", nil})
			continue
		}
		_, ok := windowsProvider(p.Attr["name"])
		if !ok {
			return nil, nil, errors.New("manifest provider outside declared scope")
		}
		provider := p.Attr["name"]
		guid, e := sourceGUID(p.Attr["guid"])
		if e != nil {
			return nil, nil, e
		}
		templates := map[string]string{}
		for _, t := range xmlDescendants(p, "template") {
			id := t.Attr["tid"]
			raw := windowsXMLText(t.Raw)
			if old, ok := templates[id]; ok && old != raw {
				return nil, nil, errors.New("conflicting template identity")
			}
			templates[id] = raw
		}
		for i, event := range xmlChildren(xmlOne(p, "events"), "event") {
			r := WindowsRecord{SourceID: s.ID, Locator: fmt.Sprintf("xml/provider/%d/events/event/%d", pi, i), FragmentSHA256: Digest([]byte(event.Raw)), Provider: provider, GUID: guid, Kind: "manifest_definition", Context: "PowerShell source manifest at commit " + valueOrEmpty(s.Revision) + "; installation and Windows build applicability not measured."}
			id, e := numericSource(event.Attr["value"], 65535)
			if e != nil || id == nil {
				return nil, nil, errors.New("manifest event ID required")
			}
			r.EventID = *id
			if r.Version, e = numericSource(event.Attr["version"], 255); e != nil {
				return nil, nil, e
			}
			if r.Level, e = manifestNumber(p, event, "level", 255); e != nil {
				return nil, nil, e
			}
			if r.Task, e = manifestNumber(p, event, "task", 65535); e != nil {
				return nil, nil, e
			}
			if r.Opcode, e = manifestNumber(p, event, "opcode", 255); e != nil {
				return nil, nil, e
			}
			for _, key := range []string{"channel", "level", "task", "opcode", "keywords"} {
				if value := event.Attr[key]; value != "" {
					r.Context += "\nManifest " + key + ": " + value
				}
			}
			if ref := event.Attr["template"]; ref != "" {
				v, ok := templates[ref]
				if !ok {
					return nil, nil, errors.New("unresolved manifest template")
				}
				r.Template = &v
			}
			if ref := event.Attr["message"]; ref != "" {
				if !strings.HasPrefix(ref, "$(string.") || !strings.HasSuffix(ref, ")") {
					return nil, nil, errors.New("unsupported message reference")
				}
				v, ok := stringsByID[strings.TrimSuffix(strings.TrimPrefix(ref, "$(string."), ")")]
				if !ok {
					return nil, nil, errors.New("missing en-US manifest message")
				}
				r.Message = &v
			}
			records = append(records, r)
		}
	}
	if len(records) == 0 {
		return nil, nil, errors.New("manifest contains no scoped event definitions")
	}
	return records, dispositions, nil
}
func valueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
func manifestNumber(provider, event *windowsXML, kind string, max int64) (*int64, error) {
	name := event.Attr[kind]
	if name == "" {
		return nil, nil
	}
	// A symbolic Winmeta name is evidence, not an invented numeric default.
	if n, e := numericSource(name, max); e == nil {
		return n, nil
	}
	candidates := xmlChildren(xmlOne(provider, kind+"s"), kind)
	if kind == "opcode" {
		scoped := []*windowsXML{}
		for _, task := range xmlChildren(xmlOne(provider, "tasks"), "task") {
			if task.Attr["name"] == event.Attr["task"] {
				for _, o := range xmlChildren(xmlOne(task, "opcodes"), "opcode") {
					if o.Attr["name"] == name {
						scoped = append(scoped, o)
					}
				}
			}
		}
		if len(scoped) > 0 {
			candidates = scoped
		}
	}
	var result *int64
	for _, v := range candidates {
		if v.Attr["name"] != name {
			continue
		}
		n, e := numericSource(v.Attr["value"], max)
		if e != nil || n == nil {
			return nil, errors.New("invalid manifest numeric declaration")
		}
		if result != nil && *result != *n {
			return nil, errors.New("ambiguous manifest numeric declaration")
		}
		result = n
	}
	if result == nil && (name[0] >= '0' && name[0] <= '9' || strings.HasPrefix(name, "-")) {
		return nil, errors.New("invalid numeric manifest descriptor")
	}
	return result, nil
}

func htmlSource(b []byte) (*html.Node, error) {
	if !utf8.Valid(b) {
		return nil, errors.New("invalid HTML encoding")
	}
	root, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var main, content *html.Node
	walkHTML(root, func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}
		if n.Data == "main" && main == nil {
			main = n
		}
		for _, a := range n.Attr {
			if a.Key == "class" && slices.Contains(strings.Fields(a.Val), "content") && n.Data == "div" {
				content = n
			}
		}
	})
	if content != nil {
		return content, nil
	}
	if main != nil {
		return main, nil
	}
	return nil, errors.New("document main content absent")
}
func htmlNodeText(n *html.Node) string {
	var out strings.Builder
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.TextNode {
			out.WriteString(n.Data)
		}
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
		if slices.Contains([]string{"p", "div", "li", "h1", "h2", "h3", "pre", "tr", "br", "blockquote"}, n.Data) {
			out.WriteByte('\n')
		}
		if n.Data == "td" || n.Data == "th" {
			out.WriteByte('\t')
		}
	}
	visit(n)
	return out.String()
}
func htmlElements(root *html.Node, name string) []*html.Node {
	out := []*html.Node{}
	walkHTML(root, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == name {
			out = append(out, n)
		}
	})
	return out
}

var eventMention = regexp.MustCompile(`(?i)Event\s*I[dD]\s*:?\s*([0-9]{1,5})`)
var numericVersionLine = regexp.MustCompile(`^\s*([0-9]{1,3})(?:\s*[-–—:.]|\s*$)`)
var documentedSystem = regexp.MustCompile(`(?s)<System(?:\s[^>]*)?>.*?</System\s*>`)

func extractWindowsDocument(s WindowsSource, b []byte) ([]WindowsRecord, []WindowsDisposition, error) {
	main, err := htmlSource(b)
	if err != nil {
		return nil, nil, err
	}
	full := htmlNodeText(main)
	title := ""
	if h := htmlElements(main, "h1"); len(h) > 0 {
		title = windowsText(htmlNodeText(h[0]))
	}
	records := []WindowsRecord{}
	dispositions := []WindowsDisposition{}
	xmlIDs := map[int64]bool{}
	for i, pre := range htmlElements(main, "pre") {
		fragment := htmlNodeText(pre)
		if !strings.Contains(fragment, "<Event") || !strings.Contains(fragment, "<Provider") {
			continue
		}
		lines := strings.Split(fragment, "\n")
		for j, line := range lines {
			line = strings.TrimSpace(line)
			lines[j] = strings.TrimPrefix(line, "- ")
		}
		// Only System metadata is asserted. Documentation payload examples may
		// contain placeholders and unescaped occurrence values; they are retained
		// verbatim but are not parsed, repaired, or represented as event schemas.
		sample := strings.Join(lines, "\n")
		systems := documentedSystem.FindAllString(sample, -1)
		if len(systems) != 1 || strings.Contains(sample, "<!") {
			return nil, nil, errors.New("exactly one inert System fragment required")
		}
		x, e := parseWindowsXML([]byte(systems[0]))
		if e != nil {
			return nil, nil, fmt.Errorf("event XML sample: %w", e)
		}
		system := x
		p := xmlOne(system, "Provider")
		id := xmlOne(system, "EventID")
		if x.Name != "System" || p == nil || id == nil {
			return nil, nil, errors.New("event sample has incomplete identity")
		}
		_, ok := windowsProvider(p.Attr["Name"])
		loc := fmt.Sprintf("html/pre/%d/text()", i)
		if !ok {
			dispositions = append(dispositions, WindowsDisposition{s.ID, loc, "excluded", "provider_outside_declared_scope", nil})
			continue
		}
		provider := p.Attr["Name"]
		eventID, e := numericSource(strings.TrimSpace(id.Text), 65535)
		if e != nil || eventID == nil {
			return nil, nil, errors.New("documented event ID invalid")
		}
		guid, e := sourceGUID(p.Attr["Guid"])
		if e != nil {
			return nil, nil, e
		}
		r := WindowsRecord{SourceID: s.ID, Locator: loc, FragmentSHA256: Digest([]byte(fragment)), Provider: provider, GUID: guid, EventID: *eventID, Kind: "documented_xml", Context: title + "\nDocumented example metadata; event occurrence values are not catalog facts. Source applicability: " + s.URL}
		for _, field := range []struct {
			name string
			dst  **int64
			max  int64
		}{{"Version", &r.Version, 255}, {"Level", &r.Level, 255}, {"Task", &r.Task, 65535}, {"Opcode", &r.Opcode, 255}} {
			if n := xmlOne(system, field.name); n != nil {
				*field.dst, e = numericSource(strings.TrimSpace(n.Text), field.max)
				if e != nil {
					return nil, nil, e
				}
			}
		}
		records = append(records, r)
		xmlIDs[*eventID] = true
	}
	// Explicit documented version lists establish identities only. They do not
	// copy level/task/opcode from the page's example into other versions.
	if s.Kind == "event_document" && len(records) > 0 {
		p, id := records[0].Provider, records[0].EventID
		for _, r := range records {
			if r.Provider != p || r.EventID != id {
				return nil, nil, errors.New("event page identity is ambiguous")
			}
		}
		start := strings.Index(full, "Event Versions:")
		if start >= 0 {
			section := full[start+len("Event Versions:"):]
			if end := strings.Index(section, "Field Descriptions:"); end >= 0 {
				section = section[:end]
			}
			if len(section) > 16384 {
				return nil, nil, errors.New("unbounded event version section")
			}
			for i, line := range strings.Split(section, "\n") {
				match := numericVersionLine.FindStringSubmatch(line)
				if match == nil {
					continue
				}
				v, e := numericSource(match[1], 255)
				if e != nil {
					return nil, nil, e
				}
				records = append(records, WindowsRecord{SourceID: s.ID, Locator: fmt.Sprintf("html/event-versions/line/%d", i), FragmentSHA256: Digest([]byte(line)), Provider: p, GUID: records[0].GUID, EventID: id, Version: v, Kind: "documented_version", Context: title + "\nExplicit documented version: " + windowsText(line) + "\nOther version-specific metadata is unspecified. Source applicability: " + s.URL})
			}
		}
	}
	if s.Kind == "provider_article" {
		monitoringTable := false
		for _, header := range htmlElements(main, "th") {
			label := strings.ToLower(strings.Join(strings.Fields(htmlNodeText(header)), ""))
			if label == "currentwindowseventid" || label == "legacywindowseventid" {
				monitoringTable = true
			}
		}
		providers := []string{}
		for _, p := range windowsProviders {
			if strings.Contains(full, p) {
				providers = append(providers, p)
			}
		}
		// These reviewed documentation selectors bind the named provider to the
		// explicit provider GUID, not to a similarly named channel.
		if strings.Contains(s.URL, "/about_logging?") && strings.Contains(strings.ToLower(full), "a0c1853b-5c40-4b15-8766-3cf1c58f985a") {
			providers = []string{"Microsoft-Windows-PowerShell"}
		}
		if strings.Contains(s.URL, "/about_logging_windows?") && strings.Contains(strings.ToLower(full), "f90714a8-5509-434a-bf6d-b1624c8a19a2") {
			providers = []string{"PowerShellCore"}
		}
		if len(providers) != 1 || monitoringTable {
			// Monitoring catalogs can mix providers and legacy/current IDs without
			// naming their provider. Account for each table row, but never turn
			// an OS generation or a criticality rating into event metadata.
			for ti, table := range htmlElements(main, "table") {
				eventTable := false
				for ri, row := range htmlElements(table, "tr") {
					headers := htmlElements(row, "th")
					if len(headers) > 0 {
						for _, cell := range headers {
							label := strings.ToLower(strings.Join(strings.Fields(htmlNodeText(cell)), ""))
							if label == "eventid" || label == "currentwindowseventid" || label == "legacywindowseventid" {
								eventTable = true
							}
						}
						continue
					}
					if eventTable && len(htmlElements(row, "td")) > 0 {
						dispositions = append(dispositions, WindowsDisposition{s.ID, fmt.Sprintf("html/table/%d/row/%d", ti, ri), "excluded", "provider_not_stated_or_ambiguous_in_event_table", nil})
					}
				}
			}
		}
		if monitoringTable {
			// A provider named in an enumeration command or recommendation does
			// not qualify every row of a mixed current/legacy monitoring catalog.
			return records, dispositions, nil
		}
		if len(providers) == 1 {
			// Both vertical Field/Value and horizontal Event ID tables occur in
			// the selected first-party articles. Cell boundaries are significant.
			for ti, table := range htmlElements(main, "table") {
				column := -1
				for ri, row := range htmlElements(table, "tr") {
					cells := htmlElements(row, "th")
					header := len(cells) > 0
					if !header {
						cells = htmlElements(row, "td")
					}
					value := ""
					for ci, cell := range cells {
						label := strings.ToLower(strings.Join(strings.Fields(htmlNodeText(cell)), ""))
						if label == "eventid" {
							if header {
								column = ci
							} else if ci+1 < len(cells) {
								value = htmlNodeText(cells[ci+1])
							}
						}
					}
					if !header && column >= 0 && column < len(cells) {
						value = htmlNodeText(cells[column])
					}
					if value == "" {
						continue
					}
					match := regexp.MustCompile(`^\s*([0-9]{1,5})(?:\s|/|$)`).FindStringSubmatch(value)
					if match == nil {
						return nil, nil, errors.New("unrecognized documented event table ID")
					}
					id, e := numericSource(match[1], 65535)
					if e != nil {
						return nil, nil, e
					}
					if xmlIDs[*id] {
						continue
					}
					fragment := htmlNodeText(row)
					records = append(records, WindowsRecord{SourceID: s.ID, Locator: fmt.Sprintf("html/table/%d/row/%d", ti, ri), FragmentSHA256: Digest([]byte(fragment)), Provider: providers[0], EventID: *id, Kind: "documented_unspecified", Context: title + "\nProvider-qualified documentation table; version and numeric descriptors unspecified. Source applicability: " + s.URL})
				}
			}
			for i, line := range strings.Split(full, "\n") {
				for mi, match := range eventMention.FindAllStringSubmatch(line, -1) {
					id, e := numericSource(match[1], 65535)
					if e != nil {
						return nil, nil, e
					}
					if xmlIDs[*id] {
						continue
					}
					records = append(records, WindowsRecord{SourceID: s.ID, Locator: fmt.Sprintf("html/main/line/%d/mention/%d", i, mi), FragmentSHA256: Digest([]byte(line)), Provider: providers[0], EventID: *id, Kind: "documented_unspecified", Context: title + "\nOfficial documentation identifies this provider and event ID; event version and unreported descriptor values are unspecified. Source applicability: " + s.URL})
				}
			}
		}
	}
	if len(records) == 0 && len(dispositions) == 0 {
		dispositions = append(dispositions, WindowsDisposition{s.ID, "document", "excluded", "no_unambiguous_provider_event_identity", nil})
	}
	return records, dispositions, nil
}

// windowsFragment resolves producer locators independently of the rendered row.
func windowsFragment(s WindowsSource, b []byte, locator string) (string, error) {
	if s.Kind == "combined_csv" {
		rows, e := parseWindowsCSV(b)
		if e != nil {
			return "", e
		}
		var n int
		if _, e = fmt.Sscanf(locator, "csv/record/%d", &n); e != nil || n < 1 || n > len(rows) || locator != fmt.Sprintf("csv/record/%d", n) {
			return "", errors.New("invalid CSV record locator")
		}
		return rows[n-1].Raw, nil
	}
	if strings.HasPrefix(locator, "xml/manifest/") {
		root, e := parseWindowsXML(b)
		if e != nil {
			return "", e
		}
		var mi, ei int
		if _, e = fmt.Sscanf(locator, "xml/manifest/%d/events/event/%d", &mi, &ei); e != nil {
			return "", e
		}
		manifests := xmlChildren(root, "manifest")
		if mi < 0 || mi >= len(manifests) {
			return "", errors.New("manifest locator out of bounds")
		}
		events := xmlChildren(xmlOne(manifests[mi], "events"), "event")
		if ei < 0 || ei >= len(events) {
			return "", errors.New("event locator out of bounds")
		}
		return events[ei].Raw, nil
	}
	if strings.HasPrefix(locator, "xml/provider/") {
		root, err := parseWindowsXML(b)
		if err != nil {
			return "", err
		}
		var pi, ei int
		if _, err = fmt.Sscanf(locator, "xml/provider/%d/events/event/%d", &pi, &ei); err != nil {
			return "", err
		}
		providers := xmlDescendants(root, "provider")
		if pi < 0 || pi >= len(providers) {
			return "", errors.New("provider locator out of bounds")
		}
		events := xmlChildren(xmlOne(providers[pi], "events"), "event")
		if ei < 0 || ei >= len(events) {
			return "", errors.New("event locator out of bounds")
		}
		return events[ei].Raw, nil
	}
	main, err := htmlSource(b)
	if err != nil {
		return "", err
	}
	if fragment, ok := windowsCatalogFragment(main, locator); ok {
		return fragment, nil
	}
	var index int
	var tableIndex, rowIndex int
	if _, e := fmt.Sscanf(locator, "html/table/%d/row/%d", &tableIndex, &rowIndex); e == nil {
		tables := htmlElements(main, "table")
		if tableIndex >= 0 && tableIndex < len(tables) {
			rows := htmlElements(tables[tableIndex], "tr")
			if rowIndex >= 0 && rowIndex < len(rows) {
				return htmlNodeText(rows[rowIndex]), nil
			}
		}
	}
	if _, err = fmt.Sscanf(locator, "html/pre/%d/text()", &index); err == nil {
		nodes := htmlElements(main, "pre")
		if index >= 0 && index < len(nodes) {
			return htmlNodeText(nodes[index]), nil
		}
	}
	full := htmlNodeText(main)
	if _, err = fmt.Sscanf(locator, "html/main/line/%d", &index); err == nil {
		lines := strings.Split(full, "\n")
		if index >= 0 && index < len(lines) {
			return lines[index], nil
		}
	}
	if _, err = fmt.Sscanf(locator, "html/event-versions/line/%d", &index); err == nil {
		start := strings.Index(full, "Event Versions:")
		if start >= 0 {
			section := full[start+len("Event Versions:"):]
			if end := strings.Index(section, "Field Descriptions:"); end >= 0 {
				section = section[:end]
			}
			lines := strings.Split(section, "\n")
			if index >= 0 && index < len(lines) {
				return lines[index], nil
			}
		}
	}
	return "", errors.New("unresolved source fragment locator")
}
