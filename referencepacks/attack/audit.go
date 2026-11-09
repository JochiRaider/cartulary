package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/referencepacks/attack/internal/native"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"golang.org/x/text/unicode/norm"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type FieldRule struct {
	Kinds       []string `json:"kinds"`
	Disposition string   `json:"disposition"`
	Target      string   `json:"target"`
	Reason      string   `json:"reason"`
	Reference   bool     `json:"reference"`
}
type Policy struct {
	RelationshipTypes []string                        `json:"relationship_types"`
	SchemaID          string                          `json:"schema_id"`
	CandidateID       string                          `json:"candidate_id"`
	Status            string                          `json:"status"`
	Common            map[string]FieldRule            `json:"common_fields"`
	Types             map[string]map[string]FieldRule `json:"object_types"`
}

func ReadPolicy(b []byte) (Policy, error) {
	var p Policy
	if err := DecodeClosed(b, &p); err != nil {
		return p, err
	}
	if p.SchemaID != "cartulary.attack_field_policy.v1" || p.CandidateID == "" || p.Status != "adopted_current" || len(p.Types) == 0 || len(p.Common) == 0 || len(p.RelationshipTypes) == 0 || !slices.IsSorted(p.RelationshipTypes) {
		return p, errors.New("field policy identity invalid")
	}
	groups := []map[string]FieldRule{p.Common}
	for _, fields := range p.Types {
		groups = append(groups, fields)
	}
	for _, fields := range groups {
		for key, r := range fields {
			if key == "" || len(r.Kinds) == 0 || r.Target == "" || r.Reason == "" || !slices.Contains([]string{"included_natively", "transformed", "source_only", "excluded"}, r.Disposition) {
				return p, errors.New("field policy rule invalid")
			}
			for _, kind := range r.Kinds {
				if !slices.Contains([]string{"object", "array", "string", "boolean", "number", "null"}, kind) {
					return p, errors.New("field kind invalid")
				}
			}
		}
	}
	return p, nil
}

type Locator struct {
	Artifact string `json:"artifact"`
	Pointer  string `json:"pointer"`
}
type FieldStats struct {
	Count           int            `json:"count"`
	Kinds           map[string]int `json:"kinds"`
	MaxScalars      int            `json:"max_scalars"`
	MaxArray        int            `json:"max_array_length"`
	NonNFC          int            `json:"non_nfc"`
	OuterWhitespace int            `json:"outer_whitespace"`
	Controls        int            `json:"control_strings"`
	Disposition     string         `json:"disposition"`
	Target          string         `json:"target"`
	Reason          string         `json:"reason"`
	Example         Locator        `json:"example"`
}
type DomainStats struct {
	Records              int            `json:"records"`
	Types                map[string]int `json:"types"`
	SROTypes             map[string]int `json:"sro_types"`
	MaxDescription       int            `json:"max_description_scalars"`
	OverlongDescriptions int            `json:"descriptions_over_8192"`
	ActiveSubtechniques  int            `json:"active_subtechniques"`
	RetiredSROs          int            `json:"retired_sros"`
	CollectionID         string         `json:"collection_id"`
	CollectionModified   string         `json:"collection_modified"`
}
type AliasCollision struct {
	Key             string   `json:"normalized_alias"`
	ObjectIDs       []string `json:"object_ids"`
	OriginalAliases []string `json:"original_aliases"`
}
type RefEdge struct {
	SourceID string  `json:"source_id"`
	TargetID string  `json:"target_id"`
	Property string  `json:"property"`
	Locator  Locator `json:"locator"`
}
type AuditReport struct {
	SourceProfileID               string                    `json:"source_profile_id"`
	SourceProfileSHA256           string                    `json:"source_profile_sha256"`
	SchemaID                      string                    `json:"schema_id"`
	Status                        string                    `json:"status"`
	BuiltAt                       string                    `json:"built_at"`
	LockSHA256                    string                    `json:"source_lock_sha256"`
	PolicySHA256                  string                    `json:"field_policy_sha256"`
	Builder                       Builder                   `json:"builder"`
	Domains                       map[string]*DomainStats   `json:"domains"`
	UniqueRecords                 int                       `json:"unique_records"`
	UniqueTypes                   map[string]int            `json:"unique_types"`
	DuplicateIDs                  int                       `json:"cross_domain_duplicate_ids"`
	Properties                    map[string]*FieldStats    `json:"properties"`
	RecordLocators                map[string][]Locator      `json:"record_locators"`
	Collisions                    []AliasCollision          `json:"alias_collisions"`
	LookupAliases                 map[string][]string       `json:"lookup_aliases"`
	TacticMemberships             []RefEdge                 `json:"tactic_memberships"`
	AliasOccurrenceCount          int                       `json:"alias_occurrences"`
	MaxAliases                    int                       `json:"max_aliases_per_object"`
	MaxExternalRefs               int                       `json:"max_external_refs_per_record"`
	CitationsWithoutID            int                       `json:"citations_without_external_id"`
	URLsWithFragments             int                       `json:"urls_with_fragments"`
	NonHTTPSURLs                  int                       `json:"non_https_urls"`
	VersionValues                 map[string]map[string]int `json:"source_version_values"`
	Lifecycle                     map[string]int            `json:"lifecycle_flags"`
	EmbeddedReferences            map[string]int            `json:"embedded_reference_counts"`
	References                    []RefEdge                 `json:"embedded_references"`
	MaxCanonicalSourceRecordBytes int                       `json:"max_canonical_source_record_bytes"`
	Issues                        []string                  `json:"issues"`
}

type sourceRecord struct {
	value     map[string]any
	canonical []byte
}
type auditState struct {
	policy              Policy
	report              *AuditReport
	records             map[string]sourceRecord
	aliases             map[string]map[string]bool
	originals           map[string]map[string]bool
	replacements        map[string]string
	replacementLocators map[string]Locator
}

var stixID = regexp.MustCompile(`^[a-z][a-z0-9-]*--[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Audit examines retained exact source bytes. Output is producer evidence, not
// runtime payload, a source-profile adoption, or a product-verification result.
func Audit(ctx context.Context, root *os.Root, lockBytes, profileBytes, policyBytes []byte, builtAt string, builder Builder) (AuditReport, error) {
	inputs, err := admitATTACKInputs(lockBytes, profileBytes, policyBytes)
	if err != nil {
		return AuditReport{}, err
	}
	return auditATTACKSources(ctx, root, inputs, builtAt, builder)
}

func auditATTACKSources(ctx context.Context, root *os.Root, inputs attackInputs, builtAt string, builder Builder) (r AuditReport, auditErr error) {
	if err := ctx.Err(); err != nil {
		return AuditReport{}, err
	}
	l, p := inputs.lock, inputs.policy
	lockBytes, policyBytes := inputs.lockBytes, inputs.policyBytes
	stamp, err := time.Parse(time.RFC3339, builtAt)
	pub, _ := time.Parse(time.RFC3339, l.PublishedAt)
	if err != nil || stamp.Before(pub) || stamp.UTC().Format(time.RFC3339) != builtAt {
		return AuditReport{}, errors.New("explicit UTC built_at must be at or after source publication")
	}
	if !digestPattern.MatchString(builder.SHA256) || builder.Size < 1 || builder.ID == "" || builder.Version == "" {
		return AuditReport{}, errors.New("immutable builder identity required")
	}
	a, err := Check(root, lockBytes)
	if err != nil {
		return AuditReport{}, err
	}
	r = AuditReport{SchemaID: "cartulary.attack_source_audit.v2", Status: "source_audit_failed", SourceProfileID: inputs.profileID, SourceProfileSHA256: Digest(inputs.profileBytes), BuiltAt: builtAt, LockSHA256: Digest(lockBytes), PolicySHA256: Digest(policyBytes), Builder: builder, Domains: map[string]*DomainStats{}, UniqueTypes: map[string]int{}, Properties: map[string]*FieldStats{}, RecordLocators: map[string][]Locator{}, Collisions: []AliasCollision{}, LookupAliases: map[string][]string{}, TacticMemberships: []RefEdge{}, VersionValues: map[string]map[string]int{}, Lifecycle: map[string]int{}, EmbeddedReferences: map[string]int{}, References: []RefEdge{}, Issues: []string{}}
	s := auditState{policy: p, report: &r, records: map[string]sourceRecord{}, aliases: map[string]map[string]bool{}, originals: map[string]map[string]bool{}, replacements: map[string]string{}, replacementLocators: map[string]Locator{}}
	defer s.summarize()
	for i, source := range l.Artifacts {
		if err := ctx.Err(); err != nil {
			return r, err
		}
		if source.Domain == "license" {
			continue
		}
		b, err := ReadEvidence(root, a.Sources[i])
		if err != nil {
			return r, err
		}
		if err = s.bundle(ctx, b, source, l.CatalogVersion); err != nil {
			return r, err
		}
	}
	s.finish()
	if err := ctx.Err(); err != nil {
		return r, err
	}
	if len(r.Issues) > 0 {
		r.Status = "source_audit_failed"
		return r, errors.New("source audit failed; inspect retained issues")
	}
	r.Status = "source_audit_passed"
	return r, nil
}

func (s *auditState) issue(code, id, field string) {
	s.report.Issues = append(s.report.Issues, code+":"+id+":"+field)
}

func (s *auditState) bundle(ctx context.Context, b []byte, source SourceArtifact, version string) error {
	v, err := canonicaljson.DecodeStrict(b)
	if err != nil {
		return errors.New("malformed source JSON")
	}
	bundle, ok := v.(map[string]any)
	if !ok || len(bundle) != 3 || bundle["type"] != "bundle" {
		return errors.New("malformed STIX bundle")
	}
	id, ok := bundle["id"].(string)
	if !ok || !strings.HasPrefix(id, "bundle--") || !stixID.MatchString(id) {
		return errors.New("invalid STIX bundle identity")
	}
	objects, ok := bundle["objects"].([]any)
	if !ok || len(objects) == 0 {
		return errors.New("STIX objects array required")
	}
	d := &DomainStats{Records: len(objects), Types: map[string]int{}, SROTypes: map[string]int{}}
	s.report.Domains[source.Domain] = d
	seen := map[string]bool{}
	collections := 0
	for index, item := range objects {
		if err := ctx.Err(); err != nil {
			return err
		}
		o, ok := item.(map[string]any)
		if !ok {
			return errors.New("STIX object must be an object")
		}
		typ, _ := o["type"].(string)
		oid, _ := o["id"].(string)
		if !stixID.MatchString(oid) || !strings.HasPrefix(oid, typ+"--") || typ == "" {
			return errors.New("STIX object identity malformed")
		}
		if seen[oid] {
			s.issue("duplicate_source_id", oid, source.Domain)
		}
		seen[oid] = true
		if _, known := s.policy.Types[typ]; !known {
			s.issue("unclassified_type", oid, typ)
		}
		loc := Locator{Artifact: source.Name, Pointer: fmt.Sprintf("/objects/%d", index)}
		if err := native.ValidateATTACKNativeSource(o); err != nil {
			s.issue("native_source_invalid", oid, loc.Artifact+":"+loc.Pointer)
		}
		s.report.RecordLocators[oid] = append(s.report.RecordLocators[oid], loc)
		encoded, err := canonicaljson.Marshal(o)
		if err != nil {
			return err
		}
		s.report.MaxCanonicalSourceRecordBytes = max(s.report.MaxCanonicalSourceRecordBytes, len(encoded))
		d.Types[typ]++
		n := utf8.RuneCountInString(text(o, "description"))
		d.MaxDescription = max(d.MaxDescription, n)
		if n > 8192 {
			d.OverlongDescriptions++
		}
		retired := flag(o, "revoked") || flag(o, "x_mitre_deprecated")
		if typ == "attack-pattern" && flag(o, "x_mitre_is_subtechnique") && !retired {
			d.ActiveSubtechniques++
		}
		if typ == "relationship" {
			d.SROTypes[text(o, "relationship_type")]++
			if retired {
				d.RetiredSROs++
			}
		}
		if typ == "x-mitre-collection" {
			collections++
			d.CollectionID = oid
			d.CollectionModified = text(o, "modified")
			if source.CollectionID == nil || oid != *source.CollectionID || o["x_mitre_version"] != version {
				s.issue("collection_identity_mismatch", oid, source.Domain)
			}
		}
		if previous, exists := s.records[oid]; exists {
			if !bytes.Equal(previous.canonical, encoded) {
				s.issue("same_id_content_conflict", oid, source.Domain)
			}
			continue
		}
		s.records[oid] = sourceRecord{o, encoded}
		s.report.UniqueTypes[typ]++
		for key, value := range o {
			s.walk(typ, oid, key, value, Locator{Artifact: source.Name, Pointer: loc.Pointer + "/" + pointerToken(key)})
		}
		s.object(o, loc)
	}
	if collections != 1 {
		s.issue("collection_count_invalid", source.Domain, fmt.Sprint(collections))
	}
	return nil
}

func kind(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case string:
		return "string"
	case bool:
		return "boolean"
	case json.Number:
		return "number"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	}
	return "invalid"
}
func pointerToken(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "~", "~0"), "/", "~1")
}
func text(o map[string]any, k string) string { v, _ := o[k].(string); return v }
func flag(o map[string]any, k string) bool   { v, _ := o[k].(bool); return v }
func array(o map[string]any, k string) []any { v, _ := o[k].([]any); return v }
func fold(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + 32
		}
		return r
	}, norm.NFC.String(s))
}

func (s *auditState) walk(typ, id, field string, value any, loc Locator) {
	rule, known := s.policy.Types[typ][field]
	if !known {
		rule, known = s.policy.Common[field]
	}
	key := typ + "/" + field
	stats := s.report.Properties[key]
	if stats == nil {
		stats = &FieldStats{Kinds: map[string]int{}, Disposition: rule.Disposition, Target: rule.Target, Reason: rule.Reason, Example: loc}
		s.report.Properties[key] = stats
	}
	stats.Count++
	k := kind(value)
	stats.Kinds[k]++
	if !known {
		s.issue("unclassified_property", id, field)
	} else if !slices.Contains(rule.Kinds, k) {
		s.issue("property_type_mismatch", id, field)
	}
	if rule.Reference {
		if target, ok := value.(string); ok {
			s.report.References = append(s.report.References, RefEdge{SourceID: id, TargetID: target, Property: field, Locator: loc})
			s.report.EmbeddedReferences[typ+"/"+field]++
		}
	}
	switch v := value.(type) {
	case string:
		stats.MaxScalars = max(stats.MaxScalars, utf8.RuneCountInString(v))
		if !norm.NFC.IsNormalString(v) {
			stats.NonNFC++
		}
		if strings.TrimSpace(v) != v {
			stats.OuterWhitespace++
		}
		if strings.ContainsFunc(v, func(r rune) bool { return unicode.IsControl(r) && r != '\n' && r != '\t' }) {
			stats.Controls++
		}
	case []any:
		stats.MaxArray = max(stats.MaxArray, len(v))
		for i, item := range v {
			s.walk(typ, id, field+"[]", item, Locator{Artifact: loc.Artifact, Pointer: fmt.Sprintf("%s/%d", loc.Pointer, i)})
		}
	case map[string]any:
		for key, item := range v {
			s.walk(typ, id, field+"."+key, item, Locator{Artifact: loc.Artifact, Pointer: loc.Pointer + "/" + pointerToken(key)})
		}
	}
}

func (s *auditState) object(o map[string]any, loc Locator) {
	id, typ := text(o, "id"), text(o, "type")
	for _, field := range []string{"spec_version", "x_mitre_attack_spec_version", "x_mitre_version"} {
		v, exists := o[field]
		if !exists {
			continue
		}
		key := typ + "/" + field
		if s.report.VersionValues[key] == nil {
			s.report.VersionValues[key] = map[string]int{}
		}
		s.report.VersionValues[key][fmt.Sprint(v)]++
	}
	flagValue := func(key string) string {
		v, ok := o[key]
		if !ok {
			return "absent"
		}
		return fmt.Sprint(v)
	}
	s.report.Lifecycle[typ+"/revoked="+flagValue("revoked")+"/deprecated="+flagValue("x_mitre_deprecated")]++
	aliases := append(slices.Clone(array(o, "aliases")), array(o, "x_mitre_aliases")...)
	s.report.MaxAliases = max(s.report.MaxAliases, len(aliases))
	s.report.AliasOccurrenceCount += len(aliases)
	for _, v := range aliases {
		a, ok := v.(string)
		if !ok {
			continue
		}
		key := fold(a)
		if s.aliases[key] == nil {
			s.aliases[key] = map[string]bool{}
			s.originals[key] = map[string]bool{}
		}
		s.aliases[key][id] = true
		s.originals[key][a] = true
	}
	refs := array(o, "external_references")
	s.report.MaxExternalRefs = max(s.report.MaxExternalRefs, len(refs))
	for _, v := range refs {
		r, ok := v.(map[string]any)
		if !ok {
			continue
		}
		if text(r, "external_id") == "" {
			s.report.CitationsWithoutID++
		}
		u := text(r, "url")
		if strings.Contains(u, "#") {
			s.report.URLsWithFragments++
		}
		if u != "" && !strings.HasPrefix(u, "https://") {
			s.report.NonHTTPSURLs++
		}
	}
	if typ == "relationship" {
		if !slices.Contains(s.policy.RelationshipTypes, text(o, "relationship_type")) {
			s.issue("unclassified_relationship_type", id, text(o, "relationship_type"))
		}
		for _, f := range []string{"relationship_type", "source_ref", "target_ref"} {
			if text(o, f) == "" {
				s.issue("missing_relationship_member", id, f)
			}
		}
		if text(o, "relationship_type") == "revoked-by" && !flag(o, "revoked") && !flag(o, "x_mitre_deprecated") {
			source, target := text(o, "source_ref"), text(o, "target_ref")
			if old, ok := s.replacements[source]; ok && old != target {
				s.issue("conflicting_replacement", source, target)
			}
			s.replacements[source] = target
			s.replacementLocators[source] = loc
		}
	}
}

func sortedKeys[M ~map[string]V, V any](m M) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
func (s *auditState) finish() {
	r := s.report
	for _, key := range sortedKeys(s.aliases) {
		if len(s.aliases[key]) > 1 {
			r.Collisions = append(r.Collisions, AliasCollision{Key: key, ObjectIDs: sortedKeys(s.aliases[key]), OriginalAliases: sortedKeys(s.originals[key])})
		}
	}
	// Remove every colliding key from every subject; no traversal-order winner.
	for _, id := range sortedKeys(s.records) {
		o := s.records[id].value
		values := append(slices.Clone(array(o, "aliases")), array(o, "x_mitre_aliases")...)
		if len(values) == 0 {
			continue
		}
		keys := map[string]bool{}
		for _, v := range values {
			if a, ok := v.(string); ok && len(s.aliases[fold(a)]) == 1 {
				keys[fold(a)] = true
			}
		}
		r.LookupAliases[id] = sortedKeys(keys)
	}
	// ATT&CK kill-chain membership names a tactic within an explicit domain.
	// Resolve only a unique direct named assertion; never a transitive path.
	tactics := map[string][]string{}
	for _, id := range sortedKeys(s.records) {
		o := s.records[id].value
		if text(o, "type") != "x-mitre-tactic" {
			continue
		}
		for _, domain := range array(o, "x_mitre_domains") {
			if d, ok := domain.(string); ok {
				key := d + "/" + text(o, "x_mitre_shortname")
				tactics[key] = append(tactics[key], id)
			}
		}
	}
	for _, id := range sortedKeys(s.records) {
		o := s.records[id].value
		for i, v := range array(o, "kill_chain_phases") {
			phase, ok := v.(map[string]any)
			if !ok {
				continue
			}
			domain := text(phase, "kill_chain_name")
			if domain == "mitre-attack" {
				domain = "enterprise-attack"
			} else {
				domain = strings.TrimPrefix(domain, "mitre-")
			}
			key := domain + "/" + text(phase, "phase_name")
			targets := tactics[key]
			if len(targets) != 1 {
				s.issue("unresolved_tactic_membership", id, key)
				continue
			}
			loc := r.RecordLocators[id][0]
			loc.Pointer += fmt.Sprintf("/kill_chain_phases/%d", i)
			r.TacticMemberships = append(r.TacticMemberships, RefEdge{SourceID: id, TargetID: targets[0], Property: "kill_chain_phases[]", Locator: loc})
		}
	}
	for _, ref := range r.References {
		if _, ok := s.records[ref.TargetID]; !ok {
			s.issue("unresolved_reference", ref.SourceID, ref.Property+"="+ref.TargetID)
		}
	}
	for _, start := range sortedKeys(s.replacements) {
		seen := map[string]bool{}
		current := start
		for current != "" {
			if seen[current] {
				s.issue("replacement_cycle", start, current)
				break
			}
			seen[current] = true
			if len(seen) > 32 {
				s.issue("replacement_chain_exceeds_32", start, current)
				break
			}
			next := s.replacements[current]
			if next != "" {
				if _, ok := s.records[next]; !ok {
					s.issue("replacement_target_missing", start, next)
					break
				}
			}
			current = next
		}
	}
	// Collection object_modified is a source-version witness, not a release date.
	for id, record := range s.records {
		if text(record.value, "type") != "x-mitre-collection" {
			continue
		}
		for _, v := range array(record.value, "x_mitre_contents") {
			item, ok := v.(map[string]any)
			if !ok {
				continue
			}
			target, ok := s.records[text(item, "object_ref")]
			if ok && target.value["modified"] != item["object_modified"] {
				s.issue("collection_version_mismatch", id, text(item, "object_ref"))
			}
		}
	}
}

// summarize also runs for partial failed audits, before report retention.
func (s *auditState) summarize() {
	r := s.report
	r.UniqueRecords = len(s.records)
	for _, locs := range r.RecordLocators {
		domains := map[string]bool{}
		for _, loc := range locs {
			domains[loc.Artifact] = true
		}
		if len(domains) > 1 {
			r.DuplicateIDs++
		}
	}

	slices.SortFunc(r.References, func(a, b RefEdge) int {
		for _, pair := range [][2]string{{a.SourceID, b.SourceID}, {a.Property, b.Property}, {a.TargetID, b.TargetID}, {a.Locator.Artifact, b.Locator.Artifact}, {a.Locator.Pointer, b.Locator.Pointer}} {
			if n := strings.Compare(pair[0], pair[1]); n != 0 {
				return n
			}
		}
		return 0
	})
	slices.Sort(r.Issues)
	r.Issues = slices.Compact(r.Issues)
}
