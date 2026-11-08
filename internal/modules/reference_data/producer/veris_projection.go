package producer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	verisinputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/veris"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type VERISRule struct {
	ID            string   `json:"rule_id"`
	Category      string   `json:"category"`
	Statement     string   `json:"statement"`
	AppliesTo     []string `json:"applies_to"`
	SourcePath    string   `json:"source_path"`
	Locator       string   `json:"locator"`
	Qualification *string  `json:"qualification"`
}
type VERISAudit struct {
	SchemaID            string         `json:"schema_id"`
	SourceLockSHA256    string         `json:"source_lock_sha256"`
	SourceProfileSHA256 string         `json:"source_profile_sha256"`
	RulesSHA256         string         `json:"rules_sha256"`
	DispositionsSHA256  string         `json:"dispositions_sha256"`
	Builder             Builder        `json:"builder"`
	Counts              map[string]int `json:"counts"`
	ObjectsSHA256       string         `json:"objects_sha256"`
	RelationshipsSHA256 string         `json:"relationships_sha256"`
}
type VERISProjection struct {
	Members   packageSource
	Artifacts []packformat.SourceArtifact
	Report    VERISAudit
	Research  map[string][]byte
}

func verisEqual(a, b any) bool {
	x, e := canonicaljson.Marshal(a)
	y, f := canonicaljson.Marshal(b)
	return e == nil && f == nil && bytes.Equal(x, y)
}

func VERISProject(ctx context.Context, root *os.Root, inventoryBytes []byte, builder Builder) (VERISProjection, error) {
	result := VERISProjection{Members: packageSource{}, Research: map[string][]byte{}}
	lock, err := VERISCheck(root, inventoryBytes)
	if err != nil {
		return result, err
	}
	inventory, err := verisInventory(inventoryBytes)
	if err != nil {
		return result, err
	}
	profileBytes, err := verisinputs.Files.ReadFile("source-profile.v1.json")
	if err != nil {
		return result, err
	}
	profileValue, err := canonicaljson.DecodeStrict(profileBytes)
	if err != nil {
		return result, err
	}
	profile := profileValue.(map[string]any)
	if profile["inventory_sha256"] != Digest(inventoryBytes) || profile["source_profile_id"] != VERISSourceProfile || profile["status"] != "adopted_frozen" {
		return result, errors.New("VERIS frozen profile mismatch")
	}
	for name, hash := range profile["adopted_schema_sha256"].(map[string]any) {
		if packformat.ProjectionDigest(name) != hash {
			return result, errors.New("VERIS native schema binding mismatch")
		}
	}
	rulesBytes, err := verisinputs.Files.ReadFile("rules.v1.json")
	if err != nil {
		return result, err
	}
	dispositions, err := verisinputs.Files.ReadFile("dispositions.v1.json")
	if err != nil {
		return result, err
	}
	if Digest(rulesBytes) != profile["rules_sha256"] || Digest(dispositions) != profile["dispositions_sha256"] {
		return result, errors.New("VERIS interpretation profile mismatch")
	}
	continuity, err := verisinputs.Files.ReadFile("continuity.v1.json")
	if err != nil {
		return result, err
	}
	if Digest(continuity) != profile["continuity_sha256"] {
		return result, errors.New("VERIS continuity decision mismatch")
	}
	var rules struct {
		SchemaID string      `json:"schema_id"`
		Rules    []VERISRule `json:"rules"`
	}
	if err := DecodeClosed(rulesBytes, &rules); err != nil {
		return result, err
	}
	if err := verisDispositions(dispositions, inventory, rules.Rules); err != nil {
		return result, err
	}
	data := map[string]any{}
	indexes := map[string]int{}
	for i, s := range lock.Sources {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		result.Artifacts = append(result.Artifacts, packformat.SourceArtifact{Ref: s.Source.URL, Version: inventory.Version + "@" + inventory.Commit, SHA256: s.Source.SHA256})
		indexes[s.Source.Path] = i
		if strings.HasSuffix(s.Source.Path, ".json") {
			b, err := ReadEvidence(root, s.Evidence)
			if err != nil {
				return result, err
			}
			v, err := canonicaljson.DecodeStrict(b)
			if err != nil {
				return result, err
			}
			data[s.Source.Path] = v
		}
	}
	// HTTPS artifacts precede producer URNs; indexes above stay stable.
	for _, input := range []struct {
		name string
		b    []byte
	}{{"continuity", continuity}, {"dispositions", dispositions}, {"rules", rulesBytes}} {
		result.Artifacts = append(result.Artifacts, packformat.SourceArtifact{Ref: "urn:cartulary:veris:" + input.name, Version: VERISSourceProfile, SHA256: Digest(input.b)})
	}
	refs := func(file, loc string) []any {
		return []any{map[string]any{"artifact_index": indexes[file], "locator": loc}}
	}
	objects := []map[string]any{}
	edges := []map[string]any{}
	counts := map[string]int{"schema_nodes": 0, "enum_locations": 0, "enum_values": 0, "property_edges": 0, "item_edges": 0, "rules": len(rules.Rules)}
	addObject := func(source map[string]any, provenance []any) string {
		kind := source["kind"].(string)
		key, _ := source["pointer"].(string)
		values := []string{}
		if kind == "semantic-rule" {
			key = source["rule_id"].(string)
		}
		if kind == "enum-value" {
			values = append(values, source["value"].(string))
		}
		id := packformat.VERISIdentity(kind, key, values...)
		label, desc := packformat.VERISDisplay(source)
		objects = append(objects, map[string]any{"schema_id": "cartulary.reference_pack.framework.veris.object.v2", "object_id": id, "object_type": kind, "display_label": label, "description": desc, "aliases": []string{}, "deprecated": false, "replacement_object_id": nil, "external_refs": []any{}, "source_refs": provenance, "extensions": map[string]any{}, "source": source})
		return id
	}
	addEdge := func(kind, from, to string, provenance []any) error {
		id, err := packformat.RelationshipID(packformat.VERISProfile, kind, from, to, provenance)
		if err != nil {
			return err
		}
		edges = append(edges, map[string]any{"schema_id": "cartulary.reference_pack.framework.veris.relationship.v2", "relationship_id": id, "relationship_type": kind, "source_object_id": from, "target_object_id": to, "description": nil, "source_refs": provenance, "extensions": map[string]any{}})
		return nil
	}
	keywords := map[string]bool{}
	for _, k := range profile["schema_keywords"].([]any) {
		keywords[k.(string)] = true
	}
	seenEnums, seenLabels := map[string]any{}, map[string]any{}
	var walk func(any, string, int) (string, any, error)
	walk = func(value any, pointer string, depth int) (string, any, error) {
		if err := ctx.Err(); err != nil {
			return "", nil, err
		}
		if depth > 32 {
			return "", nil, errors.New("VERIS structural depth overflow")
		}
		node, ok := value.(map[string]any)
		if !ok {
			return "", nil, errors.New("VERIS node must be an object")
		}
		kind := "schema-node"
		if pointer == "" {
			kind = "schema-root"
		}
		id := packformat.VERISIdentity(kind, pointer)
		native, base := map[string]any{}, map[string]any{}
		counts["schema_nodes"]++
		for k, v := range node {
			if !keywords[k] {
				return "", nil, fmt.Errorf("unsupported VERIS keyword %s", k)
			}
			switch k {
			case "properties":
				props, ok := v.(map[string]any)
				if !ok {
					return "", nil, errors.New("VERIS properties must be an object")
				}
				np, bp := map[string]any{}, map[string]any{}
				for name, child := range props {
					p := pointer + "/properties/" + packformat.VERISEscape(name)
					cid, cb, err := walk(child, p, depth+1)
					if err != nil {
						return "", nil, err
					}
					np[name], bp[name] = cid, cb
					counts["property_edges"]++
					if err := addEdge("contains", id, cid, refs("verisc-merged.json", "#"+p)); err != nil {
						return "", nil, err
					}
				}
				native[k], base[k] = np, bp
			case "items":
				p := pointer + "/items"
				cid, cb, err := walk(v, p, depth+1)
				if err != nil {
					return "", nil, err
				}
				native[k], base[k] = cid, cb
				counts["item_edges"]++
				if err := addEdge("has-item-schema", id, cid, refs("verisc-merged.json", "#"+p)); err != nil {
					return "", nil, err
				}
			case "enum":
				values, ok := v.([]any)
				if !ok {
					return "", nil, errors.New("VERIS enum must be an array")
				}
				lp := packformat.VERISLabelPointer(pointer)
				if !verisEqual(verisPointer(data["verisc-enum.json"], lp), values) {
					return "", nil, errors.New("VERIS enumeration mismatch")
				}
				defs, ok := verisPointer(data["verisc-labels.json"], lp).(map[string]any)
				if !ok || len(defs) != len(values) {
					return "", nil, errors.New("VERIS contextual definition coverage mismatch")
				}
				seenEnums[lp], seenLabels[lp] = values, defs
				counts["enum_locations"]++
				for i, raw := range values {
					text, ok := raw.(string)
					if !ok {
						return "", nil, errors.New("VERIS enum must be a string")
					}
					definition, ok := defs[text].(string)
					if !ok {
						return "", nil, errors.New("VERIS definition must be a string")
					}
					provenance := refs("verisc-labels.json", "#"+lp+"/"+packformat.VERISEscape(text))
					provenance = append(provenance, refs("verisc-merged.json", "#"+pointer+"/enum/"+strconv.Itoa(i))...)
					eid := addObject(map[string]any{"kind": "enum-value", "schema_version": inventory.Version, "pointer": pointer, "value": text, "definition": definition}, provenance)
					if err := addEdge("allows-value", id, eid, provenance); err != nil {
						return "", nil, err
					}
					counts["enum_values"]++
				}
				native[k] = v
			default:
				native[k], base[k] = v, v
			}
		}
		addObject(map[string]any{"kind": kind, "schema_version": inventory.Version, "pointer": pointer, "keywords": native}, refs("verisc-merged.json", "#"+pointer))
		return id, base, nil
	}
	_, base, err := walk(data["verisc-merged.json"], "", 0)
	if err != nil {
		return result, err
	}
	if !verisEqual(base, data["verisc.json"]) || !verisEqual(verisFlatten(data["verisc-enum.json"], false), seenEnums) || !verisEqual(verisFlatten(data["verisc-labels.json"], true), seenLabels) {
		return result, errors.New("VERIS base, enum or label accounting mismatch")
	}
	for _, rule := range rules.Rules {
		if _, ok := indexes[rule.SourcePath]; !ok {
			return result, errors.New("unlisted VERIS rule source")
		}
		provenance := refs(rule.SourcePath, rule.Locator)
		id := addObject(map[string]any{"kind": "semantic-rule", "schema_version": inventory.Version, "rule_id": rule.ID, "category": rule.Category, "statement": rule.Statement, "applies_to": rule.AppliesTo, "qualification": rule.Qualification}, provenance)
		for _, p := range rule.AppliesTo {
			kind := "schema-node"
			if p == "" {
				kind = "schema-root"
			}
			if err := addEdge("governed-by-rule", packformat.VERISIdentity(kind, p), id, provenance); err != nil {
				return result, err
			}
		}
	}
	aliases := map[string]int{}
	fold := func(s string) string {
		return strings.Map(func(r rune) rune {
			if r >= 'A' && r <= 'Z' {
				return r + 32
			}
			return r
		}, s)
	}
	for _, o := range objects {
		aliases[fold(o["display_label"].(string))]++
	}
	for _, o := range objects {
		label := o["display_label"].(string)
		if aliases[fold(label)] == 1 {
			o["aliases"] = []string{label}
		}
	}
	slices.SortFunc(objects, func(a, b map[string]any) int {
		return strings.Compare(a["object_id"].(string), b["object_id"].(string))
	})
	order := func(e map[string]any) string {
		return e["relationship_type"].(string) + "\x00" + e["source_object_id"].(string) + "\x00" + e["target_object_id"].(string) + "\x00" + e["relationship_id"].(string)
	}
	slices.SortFunc(edges, func(a, b map[string]any) int { return strings.Compare(order(a), order(b)) })
	for name, rows := range map[string][]map[string]any{"objects": objects, "relationships": edges} {
		var buf bytes.Buffer
		for _, row := range rows {
			b, err := canonicaljson.Marshal(row)
			if err != nil {
				return result, err
			}
			if len(b)+1 > 1048576 {
				return result, errors.New("VERIS row overflow")
			}
			buf.Write(b)
			buf.WriteByte('\n')
		}
		result.Members["payload/"+name+".ndjson"] = buf.Bytes()
	}
	counts["objects"], counts["relationships"] = len(objects), len(edges)
	lockBytes, err := canonicaljson.Marshal(lock)
	if err != nil {
		return result, err
	}
	result.Report = VERISAudit{"cartulary.veris_audit.v1", Digest(lockBytes), Digest(profileBytes), Digest(rulesBytes), Digest(dispositions), builder, counts, Digest(result.Members["payload/objects.ndjson"]), Digest(result.Members["payload/relationships.ndjson"])}
	result.Research = map[string][]byte{"source-lock.json": lockBytes, "source-profile.json": profileBytes, "inventory.json": inventoryBytes, "rules.json": rulesBytes, "dispositions.json": dispositions}
	result.Research["continuity.json"] = continuity
	return result, nil
}

func verisPointer(v any, pointer string) any {
	if pointer == "" {
		return v
	}
	for _, part := range strings.Split(pointer, "/")[1:] {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")]
	}
	return v
}

func verisFlatten(v any, labels bool) map[string]any {
	out := map[string]any{}
	var walk func(any, string)
	walk = func(v any, p string) {
		if _, ok := v.([]any); ok && !labels {
			out[p] = v
			return
		}
		m, ok := v.(map[string]any)
		if !ok {
			out[p] = v
			return
		}
		if labels {
			for _, x := range m {
				if _, ok := x.(string); ok {
					out[p] = v
					return
				}
			}
		}
		for k, x := range m {
			walk(x, p+"/"+packformat.VERISEscape(k))
		}
	}
	walk(v, "")
	return out
}

func verisDispositions(b []byte, inventory VERISInventory, rules []VERISRule) error {
	type item struct {
		Path        string   `json:"source_path"`
		Disposition string   `json:"disposition"`
		RuleIDs     []string `json:"rule_ids"`
		Reason      string   `json:"reason"`
	}
	type unit struct {
		Path        string   `json:"source_path"`
		Locator     string   `json:"locator"`
		Disposition string   `json:"disposition"`
		RuleIDs     []string `json:"rule_ids"`
		Reason      string   `json:"reason"`
	}
	var ledger struct {
		SchemaID string   `json:"schema_id"`
		Units    []unit   `json:"units"`
		Excluded []string `json:"excluded_content"`
		Items    []item   `json:"source_items"`
	}
	if err := DecodeClosed(b, &ledger); err != nil {
		return err
	}
	if ledger.SchemaID != "cartulary.veris_dispositions.v1" || len(ledger.Excluded) == 0 || len(ledger.Items) != len(inventory.Sources) {
		return errors.New("incomplete VERIS source disposition ledger")
	}
	sources, seen, rulesByID, covered := map[string]bool{}, map[string]bool{}, map[string]string{}, map[string]bool{}
	for _, s := range inventory.Sources {
		sources[s.Path] = true
	}
	for _, r := range rules {
		if rulesByID[r.ID] != "" || !sources[r.SourcePath] {
			return errors.New("duplicate or unsourced VERIS rule")
		}
		rulesByID[r.ID] = r.SourcePath
	}
	for _, item := range ledger.Items {
		if !sources[item.Path] || seen[item.Path] || item.Reason == "" || !slices.Contains([]string{"contextual", "included_with_context", "excluded"}, item.Disposition) {
			return errors.New("invalid VERIS source disposition")
		}
		seen[item.Path] = true
		for _, id := range item.RuleIDs {
			if rulesByID[id] != item.Path || covered[id] {
				return errors.New("invalid VERIS rule disposition")
			}
			covered[id] = true
		}
	}
	if len(covered) != len(rules) {
		return errors.New("unaccounted VERIS rules")
	}
	units, pages := map[string]bool{}, map[int]bool{}
	for _, u := range ledger.Units {
		key := u.Path + "\x00" + u.Locator
		if !sources[u.Path] || u.Locator == "" || u.Reason == "" || units[key] || !slices.Contains([]string{"documentation_only", "included_guidance_with_documentation"}, u.Disposition) {
			return errors.New("invalid VERIS unit disposition")
		}
		units[key] = true
		for _, id := range u.RuleIDs {
			if rulesByID[id] != u.Path {
				return errors.New("VERIS unit cites a different source's rule")
			}
		}
		if strings.HasSuffix(u.Path, ".pdf") {
			n, err := strconv.Atoi(strings.TrimPrefix(u.Locator, "page:"))
			if err != nil || n < 1 || n > 172 || pages[n] {
				return errors.New("invalid VERIS guide page disposition")
			}
			pages[n] = true
		}
	}
	if len(pages) != 172 {
		return errors.New("incomplete VERIS coding guide dispositions")
	}
	return nil
}
