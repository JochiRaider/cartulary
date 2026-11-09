package producer

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

var catalogSource = regexp.MustCompile(`(?im)^\s*Source:\s*([^\r\n]+)`)
var catalogID = regexp.MustCompile(`(?im)^\s*(?:Event ID|ID):\s*([0-9]{1,5})\s*$`)
var catalogLog = regexp.MustCompile(`(?im)^\s*Log name:`)
var catalogProviderQuery = regexp.MustCompile(`^\*\[System\[Provider\[@Name=['"]([^'"]+)['"]\] and \(([^()]*)\)\]\]$`)
var catalogEquality = regexp.MustCompile(`^EventID\s*=\s*([0-9]{1,5})$`)
var catalogRange = regexp.MustCompile(`^EventID\s*>=\s*([0-9]{1,5})\s+and\s+EventID\s*<=\s*([0-9]{1,5})$`)
var catalogOutput = regexp.MustCompile(`^(?:Information|Error|Warning)\s+\S+\s+([0-9]{1,5})\s+(Microsoft-Windows-[A-Za-z0-9-]+)\s+N/A\s+`)

func catalogRecord(s WindowsSource, locator, fragment, provider string, id int64, context string) WindowsRecord {
	return WindowsRecord{SourceID: s.ID, Locator: locator, FragmentSHA256: Digest([]byte(fragment)), Provider: provider, EventID: id, Kind: "documented_unspecified", Context: context + "\nSource applicability: " + s.URL, Documentation: []WindowsDocumentation{{Kind: "explanation", Scope: "source_context", Locator: locator, Content: map[string]any{"text": fragment}}}}
}

// Catalog blocks bind source and ID locally. A publisher mentioned elsewhere
// on a mixed-provider page cannot qualify this block. These are documentary
// examples, not observed incidents or universal message templates.
func extractWindowsCatalog(s WindowsSource, b []byte) ([]WindowsRecord, []WindowsDisposition, error) {
	main, err := htmlSource(b)
	if err != nil {
		return nil, nil, err
	}
	rows := []WindowsRecord{}
	dispositions := []WindowsDisposition{}
	for _, tag := range []string{"blockquote", "pre", "ul"} {
		for i, node := range htmlElements(main, tag) {
			fragment := htmlNodeText(node)
			loc := fmt.Sprintf("html/%s/%d/text()", tag, i)
			if tag == "pre" && strings.Contains(fragment, "<QueryList>") {
				query, e := parseWindowsXML([]byte(strings.TrimSpace(fragment)))
				if e != nil {
					return nil, nil, e
				}
				for j, selectNode := range xmlDescendants(query, "Select") {
					selectorLoc := fmt.Sprintf("%s/select/%d", loc, j)
					channel := selectNode.Attr["Path"]
					if !slices.Contains([]string{"Security", "System", "Application", "Setup"}, channel) {
						dispositions = append(dispositions, WindowsDisposition{s.ID, selectorLoc, "excluded", "channel_outside_standard_log_catalog_scope", nil})
						continue
					}
					m := catalogProviderQuery.FindStringSubmatch(strings.TrimSpace(selectNode.Text))
					if m == nil {
						dispositions = append(dispositions, WindowsDisposition{s.ID, selectorLoc, "excluded", "selector_does_not_define_explicit_provider_and_finite_ids", nil})
						continue
					}
					if _, ok := windowsProvider(m[1]); !ok {
						return nil, nil, fmt.Errorf("unreviewed catalog provider %q", m[1])
					}
					ids, e := catalogQueryIDs(m[2])
					if e != nil {
						return nil, nil, e
					}
					for _, id := range ids {
						row := catalogRecord(s, fmt.Sprintf("%s/id/%d", selectorLoc, id), fragment, m[1], id, "Provider-qualified subscription selector; channel: "+channel+". Version and native numeric descriptors unspecified.")
						row.Documentation[0].Content["text"] = selectNode.Raw
						rows = append(rows, row)
					}
				}
				continue
			}
			// The reviewed Setup article includes rendered event-list rows.
			if tag == "pre" {
				for j, line := range strings.Split(fragment, "\n") {
					m := catalogOutput.FindStringSubmatch(line)
					if m == nil {
						continue
					}
					if _, ok := windowsProvider(m[2]); !ok {
						return nil, nil, fmt.Errorf("unreviewed catalog provider %q", m[2])
					}
					id, e := numericSource(m[1], 65535)
					if e != nil {
						return nil, nil, e
					}
					row := catalogRecord(s, fmt.Sprintf("%s/line/%d", loc, j), fragment, m[2], *id, "Documented event-list example. Version and native numeric descriptors unspecified; hosts, users and package instances are sample context.")
					row.Documentation[0].Content["text"] = line
					rows = append(rows, row)
				}
			}
			boundaries := catalogLog.FindAllStringIndex(fragment, -1)
			starts := []int{0}
			for _, boundary := range boundaries {
				if boundary[0] > 0 {
					starts = append(starts, boundary[0])
				}
			}
			starts = append(starts, len(fragment))
			for j := 0; j+1 < len(starts); j++ {
				block := fragment[starts[j]:starts[j+1]]
				publishers, ids := catalogSource.FindAllStringSubmatch(block, -1), catalogID.FindAllStringSubmatch(block, -1)
				if len(ids) == 0 {
					continue
				}
				blockLoc := fmt.Sprintf("%s/block/%d", loc, j)
				if len(publishers) != 1 || len(ids) != 1 {
					dispositions = append(dispositions, WindowsDisposition{s.ID, blockLoc, "excluded", "block_does_not_define_one_provider_and_event_id", nil})
					continue
				}
				provider := strings.TrimSpace(publishers[0][1])
				if _, ok := windowsProvider(provider); !ok {
					return nil, nil, fmt.Errorf("unreviewed catalog provider %q", provider)
				}
				id, e := numericSource(ids[0][1], 65535)
				if e != nil {
					return nil, nil, e
				}
				row := catalogRecord(s, blockLoc, fragment, provider, *id, "Provider-qualified documentary log example. Product/application versions and symbolic severity labels are not numeric event descriptors. Occurrence values remain sample context.")
				// Only an explicitly labelled event version can override null.
				for _, field := range []struct {
					label string
					dst   **int64
					max   int64
				}{{"Event Version", &row.Version, 255}, {"Level", &row.Level, 255}, {"Task", &row.Task, 65535}, {"Opcode", &row.Opcode, 255}} {
					m := regexp.MustCompile(`(?im)^\s*` + field.label + `:\s*([0-9]+)\s*$`).FindStringSubmatch(block)
					if m != nil {
						*field.dst, e = numericSource(m[1], field.max)
						if e != nil {
							return nil, nil, e
						}
					}
				}
				row.Documentation[0].Content["text"] = block
				rows = append(rows, row)
			}
		}
	}
	if len(rows) == 0 && len(dispositions) == 0 {
		dispositions = append(dispositions, WindowsDisposition{s.ID, "document", "excluded", "no_explicit_catalog_definitions", nil})
	}
	return rows, dispositions, nil
}

func catalogQueryIDs(expression string) ([]int64, error) {
	if m := catalogRange.FindStringSubmatch(strings.TrimSpace(expression)); m != nil {
		lo, e := numericSource(m[1], 65535)
		if e != nil {
			return nil, e
		}
		hi, e := numericSource(m[2], 65535)
		if e != nil {
			return nil, e
		}
		if *lo > *hi || *hi-*lo >= 10000 {
			return nil, errors.New("catalog event range exceeds bound")
		}
		ids := []int64{}
		for i := *lo; i <= *hi; i++ {
			ids = append(ids, i)
		}
		return ids, nil
	}
	ids := []int64{}
	for _, item := range strings.Split(expression, " or ") {
		m := catalogEquality.FindStringSubmatch(strings.TrimSpace(item))
		if m == nil {
			return nil, errors.New("unsupported catalog event predicate")
		}
		id, e := numericSource(m[1], 65535)
		if e != nil {
			return nil, e
		}
		ids = append(ids, *id)
	}
	return ids, nil
}

func windowsHTMLSection(header *html.Node) string {
	var out strings.Builder
	for n := header; n != nil; n = n.NextSibling {
		if n != header && n.Type == html.ElementNode && slices.Contains([]string{"h1", "h2", "h3"}, n.Data) {
			break
		}
		out.WriteString(htmlNodeText(n))
	}
	return out.String()
}
