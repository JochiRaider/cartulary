package producer

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var sysmonPublisher = regexp.MustCompile(`<provider\s+name="(Microsoft-Windows-Sysmon)"\s+guid="([^"]+)"`)

// Read the Windows declaration as inert text. In particular, never run the T4
// generator, install Sysmon, or treat the Linux publisher as a Windows alias.
func windowsSysmonPublisher(s WindowsSource, sources map[string]WindowsAcquired, blobs map[string][]byte) (string, *string, string, error) {
	name, citation := "", ""
	var guid *string
	for id, a := range sources {
		if a.Source.Kind != "sysmon_template" {
			continue
		}
		if s.Kind == "sysmon_manifest" && (s.Revision == nil || a.Source.Revision == nil || *s.Revision != *a.Source.Revision) {
			continue
		}
		matches := sysmonPublisher.FindAllSubmatch(blobs[id], -1)
		if len(matches) != 1 || name != "" {
			return "", nil, "", errors.New("one retained Windows Sysmon publisher declaration required")
		}
		name = string(matches[0][1])
		var e error
		guid, e = sourceGUID(string(matches[0][2]))
		if e != nil {
			return "", nil, "", e
		}
		citation = "Windows publisher declaration in retained archive member " + a.Evidence.Path + "; SHA-256 " + a.Evidence.SHA256 + ". Template retained as inert text, not executed."
	}
	if name == "" {
		return "", nil, "", errors.New("matching Sysmon publisher declaration absent")
	}
	return name, guid, citation, nil
}

func extractWindowsSysmonManifest(s WindowsSource, b []byte, sources map[string]WindowsAcquired, blobs map[string][]byte) ([]WindowsRecord, []WindowsDisposition, error) {
	provider, guid, citation, e := windowsSysmonPublisher(s, sources, blobs)
	if e != nil {
		return nil, nil, e
	}
	root, e := parseWindowsXML(b)
	if e != nil {
		return nil, nil, e
	}
	if root.Name != "manifests" {
		return nil, nil, errors.New("expected a Sysmon manifest collection")
	}
	rows := []WindowsRecord{}
	dispositions := []WindowsDisposition{}
	for mi, manifest := range xmlChildren(root, "manifest") {
		for ei, event := range xmlChildren(xmlOne(manifest, "events"), "event") {
			loc := fmt.Sprintf("xml/manifest/%d/events/event/%d", mi, ei)
			switch event.Attr["target"] {
			case "linux":
				dispositions = append(dispositions, WindowsDisposition{s.ID, loc, "excluded", "linux_only_definition", nil})
				continue
			case "all", "windows":
			default:
				return nil, nil, errors.New("unreviewed Sysmon event target")
			}
			id, e := numericSource(event.Attr["value"], 65535)
			if e != nil || id == nil {
				return nil, nil, errors.New("explicit Sysmon event ID required")
			}
			version, e := numericSource(event.Attr["version"], 255)
			if e != nil || version == nil {
				return nil, nil, errors.New("explicit Sysmon event version required")
			}
			context := "Shared Sysmon source manifest at commit " + valueOrEmpty(s.Revision) + ". Configuration schema " + manifest.Attr["schemaversion"] + "; binary version " + manifest.Attr["binaryversion"] + ". These are source context, not installation applicability or event versions.\n" + citation
			context += "\nSymbol: " + event.Attr["name"] + "; symbolic level: " + event.Attr["level"]
			row := WindowsRecord{SourceID: s.ID, Locator: loc, FragmentSHA256: Digest([]byte(event.Raw)), Provider: provider, GUID: guid, EventID: *id, Version: version, Kind: "manifest_definition", Context: context, Documentation: []WindowsDocumentation{{Kind: "title", Scope: "event_version", Locator: loc, Content: map[string]any{"text": event.Attr["template"]}}, {Kind: "field_reference", Scope: "event_version", Locator: loc, Content: map[string]any{"text": event.Raw}}}}
			// Named levels and generated task/opcode expressions are preserved
			// above, not evaluated into invented numeric descriptor defaults.
			if level, err := numericSource(event.Attr["level"], 255); err == nil {
				row.Level = level
			}
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return nil, nil, errors.New("no Windows Sysmon definitions")
	}
	return rows, dispositions, nil
}

func extractWindowsSysmonDocument(s WindowsSource, b []byte, sources map[string]WindowsAcquired, blobs map[string][]byte) ([]WindowsRecord, []WindowsDisposition, error) {
	provider, guid, citation, e := windowsSysmonPublisher(s, sources, blobs)
	if e != nil {
		return nil, nil, e
	}
	main, e := htmlSource(b)
	if e != nil {
		return nil, nil, e
	}
	rows := []WindowsRecord{}
	for i, header := range htmlElements(main, "h3") {
		title := strings.TrimSpace(htmlNodeText(header))
		matches := eventMention.FindAllStringSubmatch(title, -1)
		if len(matches) != 1 || !strings.HasPrefix(title, "Event ID ") {
			continue
		}
		id, e := numericSource(matches[0][1], 65535)
		if e != nil {
			return nil, nil, e
		}
		loc, fragment := fmt.Sprintf("html/h3/%d/section()", i), windowsHTMLSection(header)
		row := catalogRecord(s, loc, fragment, provider, *id, "Sysinternals Sysmon documentary definition; event version unspecified.\n"+citation)
		row.GUID = guid
		row.Documentation[0].Scope = "event"
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil, nil, errors.New("no Sysmon documentary event sections")
	}
	return rows, nil, nil
}

// Resolve catalog locators to the exact retained element text. Suffixes identify
// assertions within a block without discarding the surrounding source context.
func windowsCatalogFragment(main *html.Node, locator string) (string, bool) {
	for _, tag := range []string{"blockquote", "pre", "ul", "h3"} {
		var i int
		if _, e := fmt.Sscanf(locator, "html/"+tag+"/%d/", &i); e != nil {
			continue
		}
		nodes := htmlElements(main, tag)
		if i < 0 || i >= len(nodes) {
			return "", false
		}
		if locator == fmt.Sprintf("html/h3/%d/section()", i) {
			return windowsHTMLSection(nodes[i]), true
		}
		if strings.HasPrefix(locator, fmt.Sprintf("html/%s/%d/text()", tag, i)) {
			return htmlNodeText(nodes[i]), true
		}
	}
	return "", false
}
