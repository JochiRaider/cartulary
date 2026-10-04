package packformat

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"golang.org/x/text/unicode/norm"
)

// ContentSource exposes only admitted members, never host paths. Each stream
// is consumed and closed before the next member is opened.
type ContentSource interface {
	Open(context.Context, string) (io.ReadCloser, error)
}
type LookupKey struct{ Kind, Value, Order string }
type ContentRow struct {
	ID, Shape string
	Canonical []byte
	Keys      []LookupKey
}

// ContentSink writes an unpublished index. Publication belongs to the owner
// coordinator; a validation or storage error discards the entire index.
type ContentSink interface {
	Append(context.Context, ContentRow) error
}

var contentShapes = loadContentShapes()

func loadContentShapes() map[string]*shape {
	result := map[string]*shape{}
	for _, p := range profiles {
		kinds := []string{"entry"}
		if p.Shape == "objects_relationships" {
			kinds = []string{"object", "relationship"}
		}
		for _, kind := range kinds {
			result[p.Key+"/"+kind] = compileProjection("profiles/" + p.Key + "." + kind + ".v1.schema.json")
		}
	}
	return result
}

type contentState struct {
	manifest     Manifest
	ids          map[string]bool
	aliases      map[string]bool
	replacements map[string]string
	emit         FindingSink
	previous     string
}

// ValidateContent is the complete content boundary for already admitted local
// release assets and restore. Live attempts interleave these checks with the
// other registry phases before starting any index writes.
func ValidateContent(ctx context.Context, m Manifest, source ContentSource, sink ContentSink) error {
	if err := ValidateContentSchema(ctx, m, source); err != nil {
		return err
	}
	if err := ValidateContentSemantics(ctx, m, source, nil); err != nil {
		return err
	}
	if err := ValidateRegistryCompatibility(ctx, m, source); err != nil {
		return err
	}
	if sink != nil {
		return ValidateContentSemantics(ctx, m, source, sink)
	}
	return nil
}

type contentPhase uint8

const (
	contentSchema contentPhase = iota
	contentSemantics
	contentRegistry
)

func ValidateContentSchema(ctx context.Context, m Manifest, source ContentSource) error {
	return validateContentPhase(ctx, m, source, nil, contentSchema, nil)
}

func ValidateContentSemantics(ctx context.Context, m Manifest, source ContentSource, sink ContentSink) error {
	return validateContentPhase(ctx, m, source, sink, contentSemantics, nil)
}

func ValidateRegistryCompatibility(ctx context.Context, m Manifest, source ContentSource) error {
	if !strings.HasPrefix(m.Key, "type_registry.") {
		return nil
	}
	return validateContentPhase(ctx, m, source, nil, contentRegistry, nil)
}

// Each pass streams one bounded row at a time. Schema admission completes for
// every file before semantic traversal starts, so file or row order cannot
// choose between failures belonging to different checks.
func validateContentPhase(ctx context.Context, m Manifest, source ContentSource, sink ContentSink, phase contentPhase, emit FindingSink) error {
	p, ok := profiles[m.Key]
	if !ok {
		return fail("contract_incompatible")
	}
	state := contentState{manifest: m, emit: emit, ids: map[string]bool{}, aliases: map[string]bool{}, replacements: map[string]string{}}
	kinds := []string{"entry"}
	if p.Shape == "objects_relationships" {
		kinds = []string{"object", "relationship"}
	}
	for _, kind := range kinds {
		path := "payload/entries.ndjson"
		expected := m.Summary.Entries
		if kind == "object" {
			path = "payload/objects.ndjson"
			expected = m.Summary.Objects
		}
		if kind == "relationship" {
			path = "payload/relationships.ndjson"
			expected = m.Summary.Relationships
		}
		if expected == nil || !admittedContentCount(kind, *expected) {
			return fail("content_schema_invalid")
		}
		stream, err := source.Open(ctx, path)
		if err != nil {
			return ErrStorage
		}
		state.previous = ""
		err = state.read(ctx, stream, kind, *expected, sink, phase)
		closeErr := stream.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return ErrStorage
		}
		if phase == contentSemantics && kind != "relationship" {
			if err := state.validateReplacements(); err != nil {
				return err
			}
		}
	}
	if phase != contentRegistry {
		return nil
	}
	c := semanticChecks{path: "$.entries", emit: emit}
	if m.Key == "type_registry.indicator" {
		c.require(len(state.ids) == 9, "")
		for id := range indicatorPolicies {
			c.id = &id
			c.require(state.ids[id], "")
		}
	}
	if (m.Key == "type_registry.host" || m.Key == "type_registry.evidence") && m.Repository == nil {
		id := "unknown"
		c.id = &id
		c.require(state.ids[id], "")
	}
	if emit != nil {
		return c.err
	}
	return c.finish("type_registry_incompatible")
}

// The relationship ceiling is independently enforceable even when the
// current payload-byte bound makes a complete at-count fixture unreachable.
func admittedContentCount(kind string, count int64) bool {
	limit := int64(2000000)
	if kind == "relationship" {
		limit = 5000000
	}
	return count >= 0 && count <= limit
}

func (s *contentState) read(ctx context.Context, stream io.Reader, kind string, expected int64, sink ContentSink, phase contentPhase) error {
	reader := bufio.NewReaderSize(contextReader{ctx, stream}, 1048576)
	count := int64(0)
	for {
		line, err := reader.ReadSlice('\n')
		if errors.Is(err, io.EOF) && len(line) == 0 {
			break
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(err, bufio.ErrBufferFull) || errors.Is(err, io.EOF) || len(line) < 2 {
			return fail("content_schema_invalid")
		}
		if err != nil {
			return ErrStorage
		}
		count++
		if count > expected {
			return fail("content_schema_invalid")
		}
		data := line[:len(line)-1]
		value, err := canonicaljson.DecodeStrict(data)
		if err != nil || !validPackNumbers(value) || !contentShapes[s.manifest.Key+"/"+kind].matches(value) {
			return fail("content_schema_invalid")
		}
		canonical, err := canonicaljson.Marshal(value)
		if err != nil || !bytes.Equal(canonical, data) {
			return fail("content_schema_invalid")
		}
		if phase == contentSchema {
			continue
		}
		entry := value.(map[string]any)
		idName, collection := "entry_id", "entries"
		if kind == "object" {
			idName, collection = "object_id", "objects"
		}
		if kind == "relationship" {
			idName, collection = "relationship_id", "relationships"
		}
		id := entry[idName].(string)
		check := &semanticChecks{path: "$." + collection + "[" + strconv.FormatInt(count-1, 10) + "]", id: &id, emit: s.emit}
		if phase == contentRegistry {
			err := validateRegistryEntry(s.manifest.Key, entry, s.manifest.Repository == nil, check)
			s.ids[id] = true
			if err != nil && (s.emit == nil || check.err != nil || !check.invalid) {
				return err
			}
			continue
		}
		row, err := s.row(entry, kind, check)
		if err != nil {
			if s.emit == nil || check.err != nil || !check.invalid {
				return err
			}
			continue
		}
		row.Canonical = bytes.Clone(data)
		if sink != nil {
			if err := sink.Append(ctx, row); err != nil {
				return ErrStorage
			}
		}
	}
	if count != expected {
		return fail("content_schema_invalid")
	}
	return nil
}
func multiline(value string) bool {
	return value != "" && utf8.RuneCountInString(value) <= 8192 && norm.NFC.IsNormalString(value) && TrimIndicatorInput(value) == value && noControls(value, true)
}
func validExtensions(value any) bool {
	m, ok := object(value)
	return ok && len(m) == 0
}
func nullableText(value any, max int, multi bool) bool {
	if value == nil {
		return true
	}
	if multi {
		return multiline(value.(string))
	}
	return singleLine(value.(string), max)
}
func unchanged(v string) string { return v }
func commonLine(v string) bool  { return singleLine(v, 256) }
func httpsReference(v string) bool {
	out, ok := NormalizeURL(v)
	return ok && out == v && strings.HasPrefix(v, "https://") && !strings.Contains(v, "#")
}
func dateValue(value any) (string, bool) {
	if value == nil {
		return "", true
	}
	v := value.(string)
	t, err := time.Parse("2006-01-02", v)
	return v, err == nil && t.Year() >= 1 && t.Format("2006-01-02") == v
}
func aliasNormalizer(key string) func(string) string {
	if key == "type_registry.host" || key == "type_registry.evidence" {
		return unchanged
	}
	if key == "type_registry.indicator" {
		return func(v string) string { return asciiLower(norm.NFC.String(TrimIndicatorInput(v))) }
	}
	return asciiLower
}
func profileAllowsAliases(key string) bool {
	return strings.HasPrefix(key, "framework.") || strings.HasPrefix(key, "type_registry.") || slices.Contains([]string{"enrichment.ms_portals", "enrichment.entra_app_ids", "enrichment.lolbas", "enrichment.loldrivers", "enrichment.lolesxi", "enrichment.windows_sids"}, key)
}
func (s *contentState) row(m map[string]any, kind string, c *semanticChecks) (ContentRow, error) {
	c.require(nullableText(m["description"], 8192, true), ".description")
	c.require(validExtensions(m["extensions"]), ".extensions")
	previousRef := ""
	for i, value := range m["source_refs"].([]any) {
		ref := value.(map[string]any)
		index, ok := integer(ref["artifact_index"], 0, int64(len(s.manifest.Artifacts))-1)
		locator := ref["locator"].(string)
		order := strconv.FormatInt(index+100, 10) + "\x00" + locator
		path := ".source_refs[" + strconv.Itoa(i) + "]"
		c.require(ok, path+".artifact_index")
		c.require(singleLine(locator, 1024), path+".locator")
		c.require(order > previousRef, path)
		previousRef = order
	}
	if kind == "relationship" {
		return s.relationship(m, c)
	}
	idName, replacementName := "entry_id", "replacement_entry_id"
	if kind == "object" {
		idName, replacementName = "object_id", "replacement_object_id"
	}
	id := m[idName].(string)
	c.require(id > s.previous && !s.ids[id], "."+idName)
	c.require(singleLine(m["display_label"].(string), 256), ".display_label")
	s.previous = id
	s.ids[id] = true
	if replacement := m[replacementName]; replacement != nil {
		c.require(m["deprecated"].(bool), ".deprecated")
		c.require(replacement.(string) != id, "."+replacementName)
		// A malformed duplicate source has no unambiguous graph meaning. Keep
		// its first declaration for the independent bounded graph check.
		if _, exists := s.replacements[id]; !exists {
			s.replacements[id] = replacement.(string)
		}
	}
	normalize := aliasNormalizer(s.manifest.Key)
	c.require(profileAllowsAliases(s.manifest.Key) || len(m["aliases"].([]any)) == 0, ".aliases")
	c.ordered(m["aliases"], ".aliases", normalize, commonLine)
	row := ContentRow{ID: id, Shape: kind, Keys: []LookupKey{}}
	for i, item := range m["aliases"].([]any) {
		alias := item.(string)
		key := normalize(alias)
		path := ".aliases[" + strconv.Itoa(i) + "]"
		if s.manifest.Key == "type_registry.host" || s.manifest.Key == "type_registry.evidence" {
			c.require(key == asciiLower(key) && isASCII(key), path)
		}
		c.require(!s.aliases[key], path)
		s.aliases[key] = true
		row.Keys = append(row.Keys, LookupKey{"alias", key, id})
	}
	if kind == "object" {
		previous := ""
		row.Keys = append(row.Keys, LookupKey{"object_id", id, id})
		for i, value := range m["external_refs"].([]any) {
			ref := value.(map[string]any)
			source, externalID := ref["source_name"].(string), ref["external_id"].(string)
			path := ".external_refs[" + strconv.Itoa(i) + "]"
			url := ""
			if ref["url"] != nil {
				url = ref["url"].(string)
				c.require(httpsReference(url), path+".url")
			}
			c.require(singleLine(source, 256), path+".source_name")
			c.require(singleLine(externalID, 256), path+".external_id")
			order := source + "\x00" + externalID + "\x00" + url
			c.require(order > previous, path)
			previous = order
			row.Keys = append(row.Keys, LookupKey{"external_id", externalID, id})
		}
		return row, c.finish("content_semantic_invalid")
	}
	keys, err := validateProfileEntry(s.manifest.Key, m, c)
	row.Keys = append(row.Keys, keys...)
	return row, err
}
func (s *contentState) validateReplacements() error {
	path := "$.entries"
	if profiles[s.manifest.Key].Shape == "objects_relationships" {
		path = "$.objects"
	}
	for start := range s.replacements {
		seen := map[string]bool{start: true}
		current := start
		count := 1
		for s.replacements[current] != "" {
			current = s.replacements[current]
			count++
			if count > 32 || seen[current] || !s.ids[current] {
				if s.emit == nil {
					return fail("content_semantic_invalid")
				}
				if err := s.emit(Finding{Path: path, EntryID: &start}); err != nil {
					return err
				}
				break
			}
			seen[current] = true
		}
	}
	return nil
}
func (s *contentState) relationship(m map[string]any, c *semanticChecks) (ContentRow, error) {
	typ, source, target, id := m["relationship_type"].(string), m["source_object_id"].(string), m["target_object_id"].(string), m["relationship_id"].(string)
	refs, err := canonicaljson.Marshal(m["source_refs"])
	if err != nil {
		return ContentRow{}, err
	}
	expected := "rpr_" + Digest(append([]byte(s.manifest.ProfileID+"\x00"+typ+"\x00"+source+"\x00"+target+"\x00"), refs...))
	order := typ + "\x00" + source + "\x00" + target + "\x00" + id
	c.require(id == expected, ".relationship_id")
	c.require(s.ids[source], ".source_object_id")
	c.require(s.ids[target], ".target_object_id")
	c.require(order > s.previous, "")
	s.previous = order
	return ContentRow{ID: id, Shape: "relationship", Keys: []LookupKey{}}, c.finish("content_semantic_invalid")
}

// EventIdentity includes the nullable version and remains unambiguous even
// when provider names contain colons.
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
func ValidSID(value string) bool {
	if len(value) > 184 || !strings.HasPrefix(value, "S-1-") {
		return false
	}
	parts := strings.Split(value[4:], "-")
	if len(parts) < 2 || len(parts) > 16 {
		return false
	}
	for i, part := range parts {
		bits := 32
		if i == 0 {
			bits = 48
		}
		n, err := strconv.ParseUint(part, 10, bits)
		if err != nil || strconv.FormatUint(n, 10) != part {
			return false
		}
	}
	return true
}
func WindowsPathLookup(value string) string {
	v := strings.ReplaceAll(norm.NFC.String(TrimIndicatorInput(value)), "/", `\`)
	for strings.Contains(v, `\\`) {
		v = strings.ReplaceAll(v, `\\`, `\`)
	}
	return asciiLower(v)
}
