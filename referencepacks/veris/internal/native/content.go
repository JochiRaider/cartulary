package native

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
	"golang.org/x/text/unicode/norm"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Manifest = envelope.Manifest
type Finding = schema.Finding
type FindingSink = schema.FindingSink
type Source interface {
	Open(context.Context, string) (io.ReadCloser, error)
}

func fail(code string) error                  { return &envelope.Failure{Code: code} }
func Digest(b []byte) string                  { return envelope.Digest(b) }
func object(value any) (map[string]any, bool) { v, ok := value.(map[string]any); return v, ok }
func integer(value any, min, max int64) (int64, bool) {
	v, ok := value.(json.Number)
	if !ok || v.String() == "-0" || strings.ContainsAny(v.String(), ".eE+") {
		return 0, false
	}
	n, err := strconv.ParseInt(v.String(), 10, 64)
	return n, err == nil && n >= min && n <= max
}
func validPackNumbers(value any) bool {
	switch v := value.(type) {
	case json.Number:
		_, ok := integer(v, -9007199254740991, 9007199254740991)
		return ok
	case []any:
		for _, child := range v {
			if !validPackNumbers(child) {
				return false
			}
		}
	case map[string]any:
		for _, child := range v {
			if !validPackNumbers(child) {
				return false
			}
		}
	}
	return true
}
func asciiLower(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + 'a' - 'A'
		}
		return r
	}, s)
}
func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}
func noControls(s string, multi bool) bool {
	for _, r := range s {
		if multi && (r == '\n' || r == '\t') {
			continue
		}
		if r < 32 || r >= 127 && r <= 159 {
			return false
		}
	}
	return true
}
func TrimIndicatorInput(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return r >= 9 && r <= 13 || r == 32 || r == 0x85 || r == 0xa0 || r == 0x1680 || r >= 0x2000 && r <= 0x200a || r == 0x2028 || r == 0x2029 || r == 0x202f || r == 0x205f || r == 0x3000
	})
}
func singleLine(s string, max int) bool {
	return s != "" && utf8.ValidString(s) && norm.NFC.IsNormalString(s) && strings.TrimSpace(s) == s && utf8.RuneCountInString(s) <= max && noControls(s, false) && !strings.ContainsAny(s, "\u2028\u2029")
}
func multiline(s string) bool {
	return s != "" && utf8.RuneCountInString(s) <= 8192 && norm.NFC.IsNormalString(s) && TrimIndicatorInput(s) == s && noControls(s, true)
}
func nullableText(v any, max int, multi bool) bool {
	if v == nil {
		return true
	}
	if multi {
		return multiline(v.(string))
	}
	return singleLine(v.(string), max)
}
func commonLine(s string) bool  { return singleLine(s, 256) }
func unchanged(s string) string { return s }

type hooks struct {
	object, relationship  func(map[string]any, *semanticChecks)
	finishObjects, finish func() error
}
type state struct {
	manifest     Manifest
	hooks        hooks
	ids, aliases map[string]bool
	replacements map[string]string
	previous     string
}

// ValidateContent preserves the producer-owned source-model assertions before
// the native record is carried into inert generic data. It never indexes or
// publishes application state.
func ValidateContent(ctx context.Context, m Manifest, source Source, _ any) error {
	return validatePhases(ctx, m, source, []bool{false, true})
}
func ValidateStructure(ctx context.Context, m Manifest, source Source) error {
	return validatePhases(ctx, m, source, []bool{false})
}
func validatePhases(ctx context.Context, m Manifest, source Source, phases []bool) error {
	if m.Key != definition.Key {
		return fail("contract_incompatible")
	}
	for _, semantics := range phases {
		state := state{manifest: m, hooks: policy(m), ids: map[string]bool{}, aliases: map[string]bool{}, replacements: map[string]string{}}
		kinds := []string{"entry"}
		if definition.Shape == "objects_relationships" {
			kinds = []string{"object", "relationship"}
		}
		for _, kind := range kinds {
			name := "payload/entries.ndjson"
			count := m.Summary.Entries
			maximum := int64(2000000)
			if kind == "object" {
				name = "payload/objects.ndjson"
				count = m.Summary.Objects
			}
			if kind == "relationship" {
				name = "payload/relationships.ndjson"
				count = m.Summary.Relationships
				maximum = 5000000
			}
			if count == nil || *count < 0 || *count > maximum {
				return fail("content_schema_invalid")
			}
			stream, err := source.Open(ctx, name)
			if err != nil {
				return err
			}
			state.previous = ""
			err = state.read(ctx, stream, kind, *count, semantics)
			err = errors.Join(err, stream.Close())
			if err != nil {
				return err
			}
			if !semantics {
				continue
			}
			if kind != "relationship" {
				if state.hooks.finishObjects != nil {
					if err = state.hooks.finishObjects(); err != nil {
						return err
					}
				}
				if err = state.replacementsValid(); err != nil {
					return err
				}
			} else if state.hooks.finish != nil {
				if err = state.hooks.finish(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func (s *state) read(ctx context.Context, stream io.Reader, kind string, count int64, semantics bool) error {
	reader := bufio.NewReaderSize(stream, definition.MaxLineBytes)
	var actual int64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		line, err := reader.ReadSlice('\n')
		if errors.Is(err, io.EOF) && len(line) == 0 {
			break
		}
		if errors.Is(err, bufio.ErrBufferFull) || errors.Is(err, io.EOF) || len(line) < 2 {
			return fail("content_schema_invalid")
		}
		if err != nil {
			return err
		}
		actual++
		if actual > count {
			return fail("content_schema_invalid")
		}
		raw := line[:len(line)-1]
		value, err := canonicaljson.DecodeStrictBounded(raw, 64)
		if err != nil || !validPackNumbers(value) || !contentShapes[definition.Key+"/"+kind].Matches(value) {
			return fail("content_schema_invalid")
		}
		canonical, err := canonicaljson.Marshal(value)
		if err != nil || !bytes.Equal(raw, canonical) {
			return fail("content_schema_invalid")
		}
		if semantics {
			if err = s.row(value.(map[string]any), kind); err != nil {
				return err
			}
		}
	}
	if actual != count {
		return fail("content_schema_invalid")
	}
	return nil
}
func (s *state) row(m map[string]any, kind string) error {
	c := &semanticChecks{}
	c.require(nullableText(m["description"], 8192, true), ".description")
	extensions, ok := object(m["extensions"])
	c.require(ok && len(extensions) == 0, ".extensions")
	previous := ""
	for _, v := range m["source_refs"].([]any) {
		ref := v.(map[string]any)
		index, ok := integer(ref["artifact_index"], 0, int64(len(s.manifest.Artifacts))-1)
		locator := ref["locator"].(string)
		order := strconv.FormatInt(index+100, 10) + "\x00" + locator
		c.require(ok && singleLine(locator, 1024) && order > previous, ".source_refs")
		previous = order
	}
	if kind == "relationship" {
		typ, source, target, id := m["relationship_type"].(string), m["source_object_id"].(string), m["target_object_id"].(string), m["relationship_id"].(string)
		expected, err := RelationshipID(definition.ID, typ, source, target, m["source_refs"])
		if err != nil {
			return err
		}
		order := typ + "\x00" + source + "\x00" + target + "\x00" + id
		c.require(id == expected && s.ids[source] && s.ids[target] && order > s.previous, "")
		s.previous = order
		if s.hooks.relationship != nil {
			s.hooks.relationship(m, c)
		}
		return c.finish("content_semantic_invalid")
	}
	idName, replacementName := "entry_id", "replacement_entry_id"
	if kind == "object" {
		idName, replacementName = "object_id", "replacement_object_id"
	}
	id := m[idName].(string)
	c.require(id > s.previous && !s.ids[id], ".identity")
	s.previous = id
	s.ids[id] = true
	c.require(singleLine(m["display_label"].(string), 256), ".display_label")
	if replacement := m[replacementName]; replacement != nil {
		c.require(m["deprecated"].(bool) && replacement.(string) != id, ".replacement")
		s.replacements[id] = replacement.(string)
	}
	c.ordered(m["aliases"], ".aliases", asciiLower, commonLine)
	for _, v := range m["aliases"].([]any) {
		key := asciiLower(v.(string))
		c.require(!s.aliases[key], ".aliases")
		s.aliases[key] = true
	}
	if kind == "object" {
		previous = ""
		for _, v := range m["external_refs"].([]any) {
			ref := v.(map[string]any)
			name, id := ref["source_name"].(string), ref["external_id"].(string)
			url := ""
			if ref["url"] != nil {
				url = ref["url"].(string)
				c.require(httpsReference(url), ".external_refs")
			}
			order := name + "\x00" + id + "\x00" + url
			c.require(singleLine(name, 256) && singleLine(id, 256) && order > previous, ".external_refs")
			previous = order
		}
		if s.hooks.object != nil {
			s.hooks.object(m, c)
		}
	} else {
		entry(m, c, s.manifest)
	}
	return c.finish("content_semantic_invalid")
}
func (s *state) replacementsValid() error {
	for id := range s.replacements {
		current := id
		seen := map[string]bool{}
		for n := 0; current != ""; n++ {
			if n == 32 || seen[current] || !s.ids[current] {
				return fail("content_semantic_invalid")
			}
			seen[current] = true
			current = s.replacements[current]
		}
	}
	return nil
}
func RelationshipID(profile, typ, source, target string, refs any) (string, error) {
	b, err := canonicaljson.Marshal(refs)
	if err != nil {
		return "", err
	}
	return "rpr_" + Digest(append([]byte(profile+"\x00"+typ+"\x00"+source+"\x00"+target+"\x00"), b...)), nil
}
func ContentSchemaFindings(ctx context.Context, m Manifest, source Source, emit FindingSink) error {
	// Producer diagnostics are independent of application retention; preserve the
	// safe closed finding surface for the assembler's explicit audit failure.
	err := ValidateContent(ctx, m, source, nil)
	if err == nil {
		return nil
	}
	var failure *envelope.Failure
	if !errors.As(err, &failure) {
		return err
	}
	return emit(Finding{Path: "$"})
}

func equalJSON(a, b any) bool {
	left, e1 := canonicaljson.Marshal(a)
	right, e2 := canonicaljson.Marshal(b)
	return e1 == nil && e2 == nil && bytes.Equal(left, right)
}
