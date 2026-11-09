package native

import (
	"strconv"
	"strings"
	"unicode"
)

func policy(Manifest) hooks      { return hooks{} }
func httpsReference(string) bool { return false }
func entry(m map[string]any, c *semanticChecks, manifest Manifest) {
	id := m["entry_id"].(string)
	provider := m["provider_name"].(string)
	c.require(singleLine(provider, 256), ".provider_name")
	c.require(len(m["aliases"].([]any)) == 0, ".aliases")
	event, _ := integer(m["event_id"], 0, 65535)
	var version *int64
	if m["event_version"] != nil {
		n, _ := integer(m["event_version"], 0, 255)
		version = &n
	}
	c.require(id == EventIdentity(provider, event, version), ".entry_id")
	c.require(nullableText(m["level"], 128, false), ".level")
	c.require(nullableText(m["task"], 256, false), ".task")
	c.require(nullableText(m["opcode"], 128, false), ".opcode")
	for _, v := range m["documentation"].([]any) {
		block := v.(map[string]any)
		c.require(windowsDocumentationPresentation(block["content"]), ".documentation")
		previous := ""
		for _, v := range block["source_refs"].([]any) {
			ref := v.(map[string]any)
			index, ok := integer(ref["artifact_index"], 0, int64(len(manifest.Artifacts))-1)
			locator := ref["locator"].(string)
			order := strconv.FormatInt(index+100, 10) + "\x00" + locator
			c.require(ok && singleLine(locator, 1024) && order > previous, ".documentation.source_refs")
			previous = order
		}
	}
}
func windowsDocumentationPresentation(value any) bool {
	switch v := value.(type) {
	case string:
		return strings.TrimRightFunc(v, unicode.IsSpace) == v
	case map[string]any:
		for _, x := range v {
			if !windowsDocumentationPresentation(x) {
				return false
			}
		}
	case []any:
		for _, x := range v {
			if !windowsDocumentationPresentation(x) {
				return false
			}
		}
	}
	return true
}
func EventIdentity(provider string, event int64, version *int64) string {
	v := "*"
	if version != nil {
		v = strconv.FormatInt(*version, 10)
	}
	return asciiLower(provider) + ":" + strconv.FormatInt(event, 10) + ":" + v
}
func ParseEventIdentity(id string) (string, int64, *int64, bool) {
	if len(id) > 266 {
		return "", 0, nil, false
	}
	last := strings.LastIndexByte(id, ':')
	if last < 0 {
		return "", 0, nil, false
	}
	second := strings.LastIndexByte(id[:last], ':')
	if second < 0 {
		return "", 0, nil, false
	}
	provider := id[:second]
	event, err := strconv.ParseInt(id[second+1:last], 10, 64)
	if err != nil || event < 0 || event > 65535 || strconv.FormatInt(event, 10) != id[second+1:last] || provider != asciiLower(provider) || !isASCII(provider) || !singleLine(provider, 256) {
		return "", 0, nil, false
	}
	var version *int64
	if id[last+1:] != "*" {
		n, err := strconv.ParseInt(id[last+1:], 10, 64)
		if err != nil || n < 0 || n > 255 || strconv.FormatInt(n, 10) != id[last+1:] {
			return "", 0, nil, false
		}
		version = &n
	}
	return provider, event, version, true
}
