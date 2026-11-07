package producer

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type windowsOriginalSummaries struct {
	SchemaID  string            `json:"schema_id"`
	Summaries map[string]string `json:"summaries"`
}

var windowsFactIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.:-]{0,159}$`)
var windowsFactLabel = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 /&(),_.:-]{0,127}$`)
var windowsOSName = regexp.MustCompile(`^Windows (?:Server )?(?:Vista|7|8|8\.1|10|11|2008|2012|2016|2019|2022|2025)(?: R2)?(?: SP[12])?(?: \[Version [0-9]{4}\])?$`)

// Public entries are an allowlist projection of facts, not a redaction pass
// over upstream prose. Research records preserve originals independently.
func windowsPublicFacts(record WindowsRecord, fragment string) []WindowsDocumentation {
	result := []WindowsDocumentation{}
	text := func(scope, value string) {
		result = append(result, WindowsDocumentation{Kind: "explanation", Scope: scope, Content: map[string]any{"text": value}})
	}
	if record.Kind == "manifest_definition" {
		if node, err := parseWindowsXML([]byte(fragment)); err == nil {
			for _, key := range []string{"symbol", "name"} {
				if value := node.Attr[key]; windowsFactIdentifier.MatchString(value) {
					text("event_version", "Manifest event identifier: "+value+".")
				}
			}
		}
	}
	for _, doc := range record.Documentation {
		switch doc.Kind {
		case "classification":
			values := map[string]any{"category": nil, "subcategory": nil, "outcomes": []any{}, "minimum_os": nil}
			for _, name := range []string{"category", "subcategory"} {
				if value, ok := doc.Content[name].(string); ok && windowsFactLabel.MatchString(value) {
					values[name] = value
				}
			}
			if outcomes, ok := doc.Content["outcomes"].([]any); ok {
				for _, outcome := range outcomes {
					if value, ok := outcome.(string); ok && slices.Contains([]string{"S", "F", "Success", "Failure"}, value) {
						values["outcomes"] = append(values["outcomes"].([]any), value)
					}
				}
			}
			if values["category"] != nil || values["subcategory"] != nil || len(values["outcomes"].([]any)) != 0 {
				result = append(result, WindowsDocumentation{Kind: "classification", Scope: doc.Scope, Content: values})
			}
		case "applicability":
			var minimum, roles any
			if value, ok := doc.Content["minimum_os"].(string); ok {
				parts := strings.Split(strings.TrimSuffix(strings.TrimSpace(value), "."), ", ")
				valid := len(parts) > 0
				for _, part := range parts {
					valid = valid && windowsOSName.MatchString(part)
				}
				if valid {
					minimum = strings.Join(parts, ", ")
				}
			}
			if value, ok := doc.Content["required_roles"].(string); ok {
				switch strings.ToLower(strings.TrimSuffix(strings.TrimSpace(value), ".")) {
				case "none", "no specific requirements":
					roles = "No server role required"
				case "active directory domain controller":
					roles = "Active Directory domain controller"
				case "for an active directory object, the domain controller role is required. for a sam object, there is no required role":
					roles = "AD objects require a domain controller; local SAM objects have no server-role requirement"
				}
			}
			if minimum != nil || roles != nil {
				result = append(result, WindowsDocumentation{Kind: "applicability", Scope: doc.Scope, Content: map[string]any{"minimum_os": minimum, "required_roles": roles}})
			}
		case "version_history":
			if values, ok := doc.Content["documented_versions"].([]any); ok {
				versions := []any{}
				labels := []string{}
				for _, value := range values {
					if n, err := strconv.ParseUint(fmt.Sprint(value), 10, 8); err == nil {
						versions = append(versions, n)
						labels = append(labels, strconv.FormatUint(n, 10))
					}
				}
				if len(versions) != 0 {
					result = append(result, WindowsDocumentation{Kind: "version_history", Scope: "event", Content: map[string]any{"text": "Documented event versions: " + strings.Join(labels, ", ") + ".", "documented_versions": versions}})
				}
			}
		case "quality_issue":
			if code, ok := doc.Content["code"].(string); ok && windowsFactIdentifier.MatchString(code) {
				text("source_context", "Source review flag: "+code+". The original discrepancy remains in local research retention.")
			}
		}
	}
	return result
}

func windowsOriginalDescription(record WindowsRecord, summaries windowsOriginalSummaries) string {
	key := strings.ToLower(record.Provider) + ":" + strconv.FormatInt(record.EventID, 10)
	text := summaries.Summaries[key]
	if text == "" {
		text = fmt.Sprintf("%s defines event %d.", record.Provider, record.EventID)
	}
	if record.Version == nil {
		text += " The reference does not specify an event version."
	} else {
		text += " This entry describes event version " + strconv.FormatInt(*record.Version, 10) + "."
	}
	return windowsText(text)
}

// Repeated source assertions share a presentation block while retaining every
// source reference. Distinct scopes or factual variants remain separate.
func windowsCoalesceDocumentation(blocks []map[string]any) ([]map[string]any, error) {
	result := []map[string]any{}
	byContent := map[string]int{}
	for _, block := range blocks {
		key, err := canonicaljson.Marshal(map[string]any{"kind": block["kind"], "scope": block["scope"], "content": block["content"]})
		if err != nil {
			return nil, err
		}
		if index, exists := byContent[string(key)]; exists {
			refs := result[index]["source_refs"].([]map[string]any)
			for _, ref := range block["source_refs"].([]map[string]any) {
				if !slices.ContainsFunc(refs, func(existing map[string]any) bool {
					return existing["artifact_index"] == ref["artifact_index"] && existing["locator"] == ref["locator"]
				}) {
					refs = append(refs, ref)
				}
			}
			result[index]["source_refs"] = refs
		} else {
			byContent[string(key)] = len(result)
			result = append(result, block)
		}
	}
	for _, block := range result {
		slices.SortFunc(block["source_refs"].([]map[string]any), func(a, b map[string]any) int {
			return strings.Compare(a["locator"].(string), b["locator"].(string))
		})
	}
	return result, nil
}
