package packformat

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

// ValidateATTACKNativeSource applies the same closed native source shapes used
// by canonical content validation. It never loads producer or upstream inputs.
func ValidateATTACKNativeSource(source map[string]any) error {
	kind := "object"
	if source["type"] == "relationship" {
		kind = "relationship"
	}
	if !contentShapes["framework.attack/"+kind].properties["source"].matches(source) {
		return fail("content_schema_invalid")
	}
	return nil
}

// ATTACKText is the common display projection. Native values are never passed
// through it for storage: whitespace, optional presence and URLs remain exact.
func ATTACKText(value string) string {
	value = strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\r", "\n")
	return norm.NFC.String(TrimIndicatorInput(value))
}

func ATTACKDescription(source map[string]any) any {
	value, _ := source["description"].(string)
	if value = ATTACKText(value); value != "" {
		return value
	}
	return nil
}

func ATTACKLabel(source map[string]any) string {
	if source["type"] == "marking-definition" {
		return "MITRE ATT&CK marking " + source["id"].(string)
	}
	value, _ := source["name"].(string)
	return ATTACKText(value)
}

func ATTACKDeprecated(source map[string]any) bool {
	return source["revoked"] == true || source["x_mitre_deprecated"] == true
}

func ATTACKAliases(source map[string]any) []string {
	values := []string{}
	for _, key := range []string{"aliases", "x_mitre_aliases"} {
		for _, value := range attackArray(source[key]) {
			values = append(values, asciiLower(norm.NFC.String(value.(string))))
		}
	}
	slices.Sort(values)
	return slices.Compact(values)
}

func ATTACKExternalRefs(source map[string]any) []map[string]any {
	result := []map[string]any{}
	seen := map[string]bool{}
	for _, value := range attackArray(source["external_references"]) {
		ref := value.(map[string]any)
		id, _ := ref["external_id"].(string)
		if id == "" {
			continue
		}
		name, _ := ref["source_name"].(string)
		var url any
		if v, ok := ref["url"].(string); ok && httpsReference(v) {
			url = v
		}
		row := map[string]any{"source_name": ATTACKText(name), "external_id": ATTACKText(id), "url": url}
		b, _ := canonicaljson.Marshal(row)
		if !seen[string(b)] {
			result = append(result, row)
			seen[string(b)] = true
		}
	}
	slices.SortFunc(result, func(a, b map[string]any) int {
		for _, k := range []string{"source_name", "external_id", "url"} {
			av, _ := a[k].(string)
			bv, _ := b[k].(string)
			if c := strings.Compare(av, bv); c != 0 {
				return c
			}
		}
		return 0
	})
	return result
}

func RelationshipID(profile, typ, source, target string, refs any) (string, error) {
	b, err := canonicaljson.Marshal(refs)
	if err != nil {
		return "", err
	}
	return "rpr_" + Digest(append([]byte(profile+"\x00"+typ+"\x00"+source+"\x00"+target+"\x00"), b...)), nil
}

func attackArray(value any) []any { values, _ := value.([]any); return values }

type attackObject struct {
	native      map[string]any
	refs        []any
	domains     []any
	aliases     []any
	replacement any
}

type attackState struct {
	objects      map[string]attackObject
	aliasOwners  map[string]map[string]bool
	edges        map[string]map[string]any
	seenEdges    map[string]bool
	replacements map[string]string
	sroIDs       map[string]bool
	records      map[string]map[string]any
}

func newAttackState() *attackState {
	return &attackState{objects: map[string]attackObject{}, aliasOwners: map[string]map[string]bool{}, edges: map[string]map[string]any{}, seenEdges: map[string]bool{}, replacements: map[string]string{}, sroIDs: map[string]bool{}, records: map[string]map[string]any{}}
}

func (a *attackState) object(m map[string]any, c *semanticChecks) {
	source := m["source"].(map[string]any)
	id := m["object_id"].(string)
	c.require(source["id"] == id, ".source.id")
	c.require(source["type"] == m["object_type"], ".source.type")
	c.require(m["display_label"] == ATTACKLabel(source), ".display_label")
	c.require(m["description"] == ATTACKDescription(source), ".description")
	c.require(m["deprecated"] == ATTACKDeprecated(source), ".deprecated")
	c.require(equalJSON(m["external_refs"], ATTACKExternalRefs(source)), ".external_refs")
	c.ordered(m["source_domains"], ".source_domains", unchanged, func(s string) bool {
		return slices.Contains([]string{"enterprise-attack", "ics-attack", "mobile-attack"}, s)
	})
	if _, exists := a.objects[id]; exists {
		return
	}
	a.objects[id] = attackObject{source, m["source_refs"].([]any), m["source_domains"].([]any), m["aliases"].([]any), m["replacement_object_id"]}
	a.records[id] = source
	for _, alias := range ATTACKAliases(source) {
		if a.aliasOwners[alias] == nil {
			a.aliasOwners[alias] = map[string]bool{}
		}
		a.aliasOwners[alias][id] = true
	}
}

type embeddedRule struct {
	field, predicate, sourceType, targetType, member string
	tacticMembership                                 bool
}

// This closed table describes only native direct occurrences. Provenance fields
// are deliberately absent; they cannot become threat relationships.
var embeddedRules = []embeddedRule{
	{field: "tactic_refs", predicate: "has-tactic", sourceType: "x-mitre-matrix", targetType: "x-mitre-tactic"},
	{field: "kill_chain_phases", predicate: "member-of-tactic", sourceType: "attack-pattern", targetType: "x-mitre-tactic", tacticMembership: true},
	{field: "x_mitre_analytic_refs", predicate: "has-analytic", sourceType: "x-mitre-detection-strategy", targetType: "x-mitre-analytic"},
	{field: "x_mitre_log_source_references", predicate: "uses-data-component", sourceType: "x-mitre-analytic", targetType: "x-mitre-data-component", member: "x_mitre_data_component_ref"},
}

func embeddedNativeObject(id string, object map[string]any) (map[string]any, bool) {
	source, ok := object["source"].(map[string]any)
	if !ok || source["id"] != id {
		return nil, false
	}
	kind, suffix, found := strings.Cut(id, "--")
	parsed, err := uuid.Parse(suffix)
	if !found || err != nil || parsed.String() != suffix || source["type"] != kind {
		return nil, false
	}
	return source, true
}

func (r embeddedRule) target(value any, tactics map[string][]string) (string, error) {
	if r.member != "" {
		entry, ok := value.(map[string]any)
		if !ok {
			return "", fail("content_semantic_invalid")
		}
		value = entry[r.member]
	}
	if !r.tacticMembership {
		target, ok := value.(string)
		if !ok {
			return "", fail("content_semantic_invalid")
		}
		return target, nil
	}
	phase, ok := value.(map[string]any)
	if !ok {
		return "", fail("content_semantic_invalid")
	}
	domain := map[string]string{"mitre-attack": "enterprise-attack", "mitre-mobile-attack": "mobile-attack", "mitre-ics-attack": "ics-attack"}[stringValue(phase["kill_chain_name"])]
	name, ok := phase["phase_name"].(string)
	targets := tactics[domain+"\x00"+name]
	if domain == "" || !ok || len(targets) != 1 {
		return "", fail("content_semantic_invalid")
	}
	return targets[0], nil
}

func stringValue(v any) string { s, _ := v.(string); return s }

// ATTACKEmbedded validates typed endpoints before deriving each occurrence.
// Retired endpoints remain valid: lifecycle does not change the native type.
func ATTACKEmbedded(objects map[string]map[string]any) ([]map[string]any, error) {
	tactics := map[string][]string{}
	for id, object := range objects {
		source, ok := embeddedNativeObject(id, object)
		if !ok {
			return nil, fail("content_semantic_invalid")
		}
		if source["type"] != "x-mitre-tactic" {
			continue
		}
		name, ok := source["x_mitre_shortname"].(string)
		if !ok {
			return nil, fail("content_semantic_invalid")
		}
		for _, domain := range attackArray(object["source_domains"]) {
			key := stringValue(domain) + "\x00" + name
			tactics[key] = append(tactics[key], id)
		}
	}
	result := []map[string]any{}
	for id, object := range objects {
		source := object["source"].(map[string]any)
		for _, rule := range embeddedRules {
			value, present := source[rule.field]
			if !present {
				continue
			}
			values, ok := value.([]any)
			if !ok || source["type"] != rule.sourceType {
				return nil, fail("content_semantic_invalid")
			}
			for i, value := range values {
				target, err := rule.target(value, tactics)
				if err != nil {
					return nil, err
				}
				endpoint, ok := objects[target]
				if !ok {
					return nil, fail("content_semantic_invalid")
				}
				native, ok := embeddedNativeObject(target, endpoint)
				if !ok || native["type"] != rule.targetType {
					return nil, fail("content_semantic_invalid")
				}
				suffix := "/" + rule.field + "/" + strconv.Itoa(i)
				if rule.member != "" {
					suffix += "/" + rule.member
				}
				refs := []map[string]any{}
				locators, ok := object["source_refs"].([]any)
				if !ok || len(locators) == 0 {
					return nil, fail("content_semantic_invalid")
				}
				for _, locator := range locators {
					ref, ok := locator.(map[string]any)
					if !ok {
						return nil, fail("content_semantic_invalid")
					}
					path, ok := ref["locator"].(string)
					if !ok {
						return nil, fail("content_semantic_invalid")
					}
					refs = append(refs, map[string]any{"artifact_index": ref["artifact_index"], "locator": path + suffix})
				}
				rid, err := RelationshipID("cartulary.reference_pack.framework.attack.v2", rule.predicate, id, target, refs)
				if err != nil {
					return nil, err
				}
				result = append(result, map[string]any{"schema_id": "cartulary.reference_pack.framework.attack.relationship.v2", "relationship_id": rid, "relationship_type": rule.predicate, "source_object_id": id, "target_object_id": target, "description": nil, "source_refs": refs, "extensions": map[string]any{}, "source_kind": "embedded_reference", "source": nil})
			}
		}
	}
	return result, nil
}

func (a *attackState) finishObjects() error {
	objects := map[string]map[string]any{}
	for id, o := range a.objects {
		aliases := []string{}
		for _, key := range ATTACKAliases(o.native) {
			if len(a.aliasOwners[key]) == 1 {
				aliases = append(aliases, key)
			}
		}
		if !equalJSON(aliases, o.aliases) {
			return fail("content_semantic_invalid")
		}
		objects[id] = map[string]any{"source": o.native, "source_refs": o.refs, "source_domains": o.domains}
	}
	edges, err := ATTACKEmbedded(objects)
	if err != nil {
		return err
	}
	for _, e := range edges {
		a.edges[e["relationship_id"].(string)] = e
	}
	return nil
}

func (a *attackState) relationship(m map[string]any, c *semanticChecks) {
	id := m["relationship_id"].(string)
	if m["source_kind"] == "embedded_reference" {
		expected, ok := a.edges[id]
		c.require(ok && equalJSON(expected, m), ".source_refs")
		a.seenEdges[id] = true
		return
	}
	s, ok := m["source"].(map[string]any)
	c.require(ok, ".source")
	if !ok {
		return
	}
	sid := s["id"].(string)
	c.require(!a.sroIDs[sid], ".source.id")
	a.sroIDs[sid] = true
	a.records[sid] = s
	for _, p := range []struct{ canonical, native string }{{"source_object_id", "source_ref"}, {"target_object_id", "target_ref"}, {"relationship_type", "relationship_type"}} {
		c.require(m[p.canonical] == s[p.native], "."+p.canonical)
	}
	c.require(m["description"] == ATTACKDescription(s), ".description")
	if s["relationship_type"] == "revoked-by" && !ATTACKDeprecated(s) {
		from, to := s["source_ref"].(string), s["target_ref"].(string)
		c.require(a.replacements[from] == "" || a.replacements[from] == to, ".source.target_ref")
		a.replacements[from] = to
	}
}

func (a *attackState) finish() error {
	if len(a.edges) != len(a.seenEdges) {
		return fail("content_semantic_invalid")
	}
	for id, o := range a.objects {
		var expected any
		if replacement := a.replacements[id]; replacement != "" {
			expected = replacement
		}
		if o.replacement != expected {
			return fail("content_semantic_invalid")
		}
	}
	for _, s := range a.records {
		for _, key := range []string{"created_by_ref", "x_mitre_modified_by_ref", "x_mitre_data_source_ref"} {
			if target, ok := s[key].(string); ok && a.records[target] == nil {
				return fail("content_semantic_invalid")
			}
		}
		for _, target := range attackArray(s["object_marking_refs"]) {
			if a.records[target.(string)] == nil {
				return fail("content_semantic_invalid")
			}
		}
		for _, value := range attackArray(s["x_mitre_contents"]) {
			item := value.(map[string]any)
			target := a.records[item["object_ref"].(string)]
			if target == nil || target["modified"] != item["object_modified"] {
				return fail("content_semantic_invalid")
			}
		}
	}
	return nil
}

// CanonicalMap bridges typed builder rows without changing native JSON values.
func CanonicalMap(value any) (map[string]any, error) {
	b, err := canonicaljson.Marshal(value)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	return m, err
}
