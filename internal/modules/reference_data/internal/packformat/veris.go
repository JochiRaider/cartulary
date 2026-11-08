package packformat

import (
	"crypto/sha256"
	"encoding/hex"
	"html"
	"regexp"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"golang.org/x/text/unicode/norm"
)

const VERISProfile = "cartulary.reference_pack.framework.veris.v2"

// VERISIdentity binds a source subject, never a display label or release.
func VERISIdentity(kind, key string, value ...string) string {
	parts := []string{"veris-community", kind, key}
	parts = append(parts, value...)
	b, _ := canonicaljson.Marshal(parts)
	d := sha256.Sum256(b)
	return "veris-" + kind + "-" + hex.EncodeToString(d[:])
}

func VERISEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "~", "~0"), "/", "~1")
}

var verisTags = regexp.MustCompile(`<[^>]*>`)

func VERISText(s string) string {
	return norm.NFC.String(strings.Join(strings.Fields(html.UnescapeString(verisTags.ReplaceAllString(s, " "))), " "))
}

func VERISPath(pointer string) string {
	parts := strings.Split(pointer, "/")
	out := ""
	for i := 1; i < len(parts); i++ {
		if parts[i] == "properties" && i+1 < len(parts) {
			i++
			if out != "" {
				out += "."
			}
			out += strings.ReplaceAll(strings.ReplaceAll(parts[i], "~1", "/"), "~0", "~")
		} else if parts[i] == "items" {
			out += "[]"
		}
	}
	return out
}

// VERISDisplay derives only presentation. Original text stays in source.
func VERISDisplay(source map[string]any) (string, any) {
	kind := source["kind"].(string)
	label, description := "", ""
	switch kind {
	case "schema-root", "schema-node":
		label = VERISPath(source["pointer"].(string))
		if kind == "schema-root" {
			label = "VERIS Community schema"
		}
		description, _ = source["keywords"].(map[string]any)["description"].(string)
	case "enum-value":
		label = VERISPath(source["pointer"].(string)) + ": " + source["value"].(string)
		description = source["definition"].(string)
	case "semantic-rule":
		label = "VERIS rule: " + source["rule_id"].(string)
		description = source["statement"].(string)
	}
	description = VERISText(description)
	if description == "" {
		return VERISText(label), nil
	}
	return VERISText(label), description
}

type verisState struct {
	manifest Manifest
	objects  map[string]map[string]any
	nodes    map[string]string
	edges    map[string]any
	seen     map[string]bool
}

func newVERISState(m Manifest) *verisState {
	return &verisState{m, map[string]map[string]any{}, map[string]string{}, map[string]any{}, map[string]bool{}}
}

func (s *verisState) object(m map[string]any, c *semanticChecks) {
	source := m["source"].(map[string]any)
	kind := source["kind"].(string)
	id := m["object_id"].(string)
	key, _ := source["pointer"].(string)
	var values []string
	if kind == "semantic-rule" {
		key = source["rule_id"].(string)
	}
	if kind == "enum-value" {
		values = []string{source["value"].(string)}
	}
	c.require(id == VERISIdentity(kind, key, values...), ".object_id")
	c.require(m["object_type"] == kind, ".object_type")
	c.require(source["schema_version"] == s.manifest.SourceVersion, ".source.schema_version")
	label, description := VERISDisplay(source)
	c.require(m["display_label"] == label, ".display_label")
	c.require(m["description"] == description, ".description")
	c.require(m["deprecated"] == false && m["replacement_object_id"] == nil, ".deprecated")
	c.require(len(m["external_refs"].([]any)) == 0, ".external_refs")
	s.objects[id] = m
	if kind == "schema-root" || kind == "schema-node" {
		_, exists := s.nodes[key]
		c.require(!exists, ".source.pointer")
		s.nodes[key] = id
		c.require(equalJSON(m["source_refs"], s.refs("verisc-merged.json", "#"+key)), ".source_refs")
	}
}

func (s *verisState) refs(file, locator string) []any {
	for i, a := range s.manifest.Artifacts {
		if strings.HasSuffix(a.Ref, "/"+file) {
			return []any{map[string]any{"artifact_index": i, "locator": locator}}
		}
	}
	return nil
}

func (s *verisState) addEdge(kind, from, to string, refs any) bool {
	id, err := RelationshipID(VERISProfile, kind, from, to, refs)
	if err != nil || s.edges[id] != nil {
		return false
	}
	s.edges[id] = refs
	return true
}

func (s *verisState) finishObjects() error {
	invalid := func() error { return fail("content_semantic_invalid") }
	root, ok := s.nodes[""]
	if !ok || s.objects[root]["object_type"] != "schema-root" {
		return invalid()
	}
	aliasOwners := map[string]int{}
	for _, m := range s.objects {
		aliasOwners[asciiLower(m["display_label"].(string))]++
	}
	visited := map[string]bool{}
	var walk func(string, string, int) bool
	walk = func(id, pointer string, depth int) bool {
		if depth > 32 || visited[id] || s.nodes[pointer] != id {
			return false
		}
		visited[id] = true
		m := s.objects[id]
		if m == nil {
			return false
		}
		source := m["source"].(map[string]any)
		kind := "schema-node"
		if pointer == "" {
			kind = "schema-root"
		}
		if source["kind"] != kind {
			return false
		}
		kw := source["keywords"].(map[string]any)
		props, _ := kw["properties"].(map[string]any)
		if props != nil && kw["type"] != "object" {
			return false
		}
		for name, child := range props {
			p := pointer + "/properties/" + VERISEscape(name)
			if !walk(child.(string), p, depth+1) || !s.addEdge("contains", id, child.(string), s.refs("verisc-merged.json", "#"+p)) {
				return false
			}
		}
		if names, ok := kw["required"].([]any); ok {
			if kw["type"] != "object" {
				return false
			}
			for _, name := range names {
				if props[name.(string)] == nil {
					return false
				}
			}
		}
		if child, ok := kw["items"].(string); ok {
			p := pointer + "/items"
			if kw["type"] != "array" || !walk(child, p, depth+1) || !s.addEdge("has-item-schema", id, child, s.refs("verisc-merged.json", "#"+p)) {
				return false
			}
		}
		if values, ok := kw["enum"].([]any); ok {
			if kw["type"] != "string" {
				return false
			}
			for i, v := range values {
				value := v.(string)
				eid := VERISIdentity("enum-value", pointer, value)
				enum := s.objects[eid]
				if enum == nil {
					return false
				}
				refs := s.refs("verisc-labels.json", "#"+VERISLabelPointer(pointer)+"/"+VERISEscape(value))
				refs = append(refs, s.refs("verisc-merged.json", "#"+pointer+"/enum/"+strconv.Itoa(i))...)
				if len(refs) != 2 || !equalJSON(enum["source_refs"], refs) || !s.addEdge("allows-value", id, eid, refs) {
					return false
				}
				visited[eid] = true
			}
		}
		return true
	}
	if !walk(root, "", 0) {
		return invalid()
	}
	for id, m := range s.objects {
		source := m["source"].(map[string]any)
		if source["kind"] == "semantic-rule" {
			for _, v := range source["applies_to"].([]any) {
				from, ok := s.nodes[v.(string)]
				if !ok || !s.addEdge("governed-by-rule", from, id, m["source_refs"]) {
					return invalid()
				}
			}
			visited[id] = true
		}
		if !visited[id] {
			return invalid()
		}
		aliases := []string{}
		label := m["display_label"].(string)
		if aliasOwners[asciiLower(label)] == 1 {
			aliases = append(aliases, label)
		}
		if !equalJSON(m["aliases"], aliases) {
			return invalid()
		}
	}
	return nil
}

// Label paths omit schema structural keywords, including array item nodes.
func VERISLabelPointer(pointer string) string {
	parts := strings.Split(pointer, "/")
	out := []string{}
	for i := 1; i < len(parts); i++ {
		if parts[i] == "properties" && i+1 < len(parts) {
			i++
			out = append(out, parts[i])
		}
	}
	return "/" + strings.Join(out, "/")
}

func (s *verisState) relationship(m map[string]any, c *semanticChecks) {
	id := m["relationship_id"].(string)
	refs, exists := s.edges[id]
	c.require(exists && !s.seen[id] && equalJSON(refs, m["source_refs"]), ".source_refs")
	c.require(m["description"] == nil, ".description")
	s.seen[id] = true
}

func (s *verisState) finish() error {
	if len(s.seen) != len(s.edges) {
		return fail("content_semantic_invalid")
	}
	for id := range s.edges {
		if !s.seen[id] {
			return fail("content_semantic_invalid")
		}
	}
	return nil
}
