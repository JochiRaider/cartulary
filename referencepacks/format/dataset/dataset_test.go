package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
	"reflect"
)

type memorySource map[string][]byte

func (s memorySource) Open(_ context.Context, name string) (io.ReadCloser, error) {
	v, ok := s[name]
	if !ok {
		return nil, errors.New("absent")
	}
	return io.NopCloser(bytes.NewReader(v)), nil
}

type recordingSink struct{ rows []Row }

func (s *recordingSink) Append(_ context.Context, row Row) error {
	s.rows = append(s.rows, row)
	return nil
}
func canonical(t *testing.T, v any) []byte {
	t.Helper()
	b, err := canonicaljson.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func object(id string) Object {
	return Object{SchemaID: ObjectSchema, ID: id, Type: "example", DisplayLabel: id, SourceRefs: []SourceRef{{ArtifactIndex: 0, Locator: "#"}}, Data: json.RawMessage(`{"native":"inert"}`), LookupValues: map[string][]json.RawMessage{}}
}
func scalar(name, normalization string, unique bool, coverage Coverage) Lookup {
	return Lookup{Name: name, InputType: "scalar", Components: []Component{{Name: "value", Type: "string", Normalization: normalization}}, Unique: unique, Coverage: coverage}
}
func source(t *testing.T, d Descriptor, objects []Object, relationships []Relationship) memorySource {
	t.Helper()
	s := memorySource{DescriptorPath: canonical(t, d), ObjectsPath: {}, RelationshipsPath: {}}
	for _, o := range objects {
		s[ObjectsPath] = append(s[ObjectsPath], canonical(t, o)...)
		s[ObjectsPath] = append(s[ObjectsPath], '\n')
	}
	for _, r := range relationships {
		s[RelationshipsPath] = append(s[RelationshipsPath], canonical(t, r)...)
		s[RelationshipsPath] = append(s[RelationshipsPath], '\n')
	}
	return s
}
func descriptor(lookups ...Lookup) Descriptor {
	if lookups == nil {
		lookups = []Lookup{}
	}
	return Descriptor{SchemaID: DescriptorSchema, Lookups: lookups}
}
func failureCode(err error) string {
	var f *Failure
	if errors.As(err, &f) {
		return f.Code
	}
	return ""
}

func TestGenericAdmissionAndLookup(t *testing.T) {
	d := descriptor(scalar("alias", "ascii_case_insensitive", false, Coverage{Objects: 2, Postings: 2, DistinctKeys: 1}), scalar("service_principal_name", "ascii_case_insensitive", true, Coverage{}))
	a, b := object("a"), object("b")
	a.LookupValues["alias"] = []json.RawMessage{json.RawMessage(`"EXAMPLE"`)}
	b.LookupValues["alias"] = []json.RawMessage{json.RawMessage(`"example"`)}
	s := source(t, d, []Object{a, b}, nil)
	sink := &recordingSink{}
	result, err := Validate(context.Background(), s, Expected{Objects: 2, SourceArtifacts: 1}, sink)
	if err != nil {
		t.Fatal(err)
	}
	if result.Objects != 2 || result.Postings != 2 || len(sink.rows) != 2 || sink.rows[0].ID != "a" || sink.rows[1].ID != "b" {
		t.Fatalf("unexpected index: %+v", result)
	}
	if !bytes.Equal(result.DescriptorBytes, s[DescriptorPath]) || sink.rows[0].Postings[0].Key != `["example"]` {
		t.Fatal("descriptor or normalized index changed")
	}
	key, err := Normalize(result.Descriptor, "service_principal_name", []byte(`"API://Valid-SPN"`))
	if err != nil || key != `["api://valid-spn"]` {
		t.Fatal("zero coverage is not unsupported", key, err)
	}
	if _, err = Normalize(result.Descriptor, "missing", []byte(`"x"`)); !errors.Is(err, ErrUnsupportedLookup) {
		t.Fatal(err)
	}
	if _, err = Normalize(result.Descriptor, "alias", []byte(`null`)); !errors.Is(err, ErrInvalidInput) {
		t.Fatal(err)
	}
}

func TestGenericSafetyAndAtomicIndex(t *testing.T) {
	d := descriptor(scalar("alias", "ascii_case_insensitive", true, Coverage{Objects: 1, Postings: 1, DistinctKeys: 1}))
	base := object("a")
	base.LookupValues["alias"] = []json.RawMessage{json.RawMessage(`"A"`)}
	for _, tc := range []struct {
		name   string
		mutate func(memorySource)
		code   string
	}{
		{"coverage", func(s memorySource) {
			s[DescriptorPath] = bytes.Replace(s[DescriptorPath], []byte(`"postings":1`), []byte(`"postings":0`), 1)
		}, "content_semantic_invalid"},
		{"unknown field", func(s memorySource) {
			s[ObjectsPath] = bytes.Replace(s[ObjectsPath], []byte(`"data":`), []byte(`"code":"run","data":`), 1)
		}, "content_schema_invalid"},
		{"unknown lookup", func(s memorySource) {
			s[ObjectsPath] = bytes.Replace(s[ObjectsPath], []byte(`"alias":[`), []byte(`"unknown":[`), 1)
		}, "content_semantic_invalid"},
		{"duplicate normalized", func(s memorySource) {
			s[ObjectsPath] = bytes.Replace(s[ObjectsPath], []byte(`["A"]`), []byte(`["A","a"]`), 1)
		}, "content_semantic_invalid"},
		{"executable descriptor", func(s memorySource) {
			s[DescriptorPath] = append([]byte(`{"expression":"eval",`), s[DescriptorPath][1:]...)
		}, "content_schema_invalid"},
		{"fractional inert number", func(s memorySource) {
			s[ObjectsPath] = bytes.Replace(s[ObjectsPath], []byte(`"native":"inert"`), []byte(`"native":1.5`), 1)
		}, "content_schema_invalid"},
		{"missing newline", func(s memorySource) { s[ObjectsPath] = bytes.TrimSuffix(s[ObjectsPath], []byte{'\n'}) }, "content_schema_invalid"},
		{"duplicate JSON member", func(s memorySource) {
			s[ObjectsPath] = bytes.Replace(s[ObjectsPath], []byte(`"deprecated":false`), []byte(`"deprecated":false,"deprecated":true`), 1)
		}, "content_schema_invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := source(t, d, []Object{base}, nil)
			tc.mutate(s)
			sink := &recordingSink{}
			_, err := Validate(context.Background(), s, Expected{Objects: 1, SourceArtifacts: 1}, sink)
			if failureCode(err) != tc.code || len(sink.rows) != 0 {
				t.Fatalf("error=%v private writes=%d", err, len(sink.rows))
			}
		})
	}
	a, b := object("a"), object("b")
	a.LookupValues["alias"] = []json.RawMessage{json.RawMessage(`"A"`)}
	b.LookupValues["alias"] = []json.RawMessage{json.RawMessage(`"a"`)}
	d.Lookups[0].Coverage = Coverage{Objects: 2, Postings: 2, DistinctKeys: 1}
	if _, err := Validate(context.Background(), source(t, d, []Object{a, b}, nil), Expected{Objects: 2, SourceArtifacts: 1}, nil); failureCode(err) != "content_semantic_invalid" {
		t.Fatal("unique lookup accepted duplicate", err)
	}
	rel := Relationship{SchemaID: RelationshipSchema, ID: "edge", Type: "related", Source: "a", Target: "missing", SourceRefs: []SourceRef{{Locator: "#"}}, Data: json.RawMessage(`{}`)}
	sink := &recordingSink{}
	_, err := Validate(context.Background(), source(t, descriptor(), []Object{object("a")}, []Relationship{rel}), Expected{Objects: 1, Relationships: 1, SourceArtifacts: 1}, sink)
	if failureCode(err) != "content_semantic_invalid" || len(sink.rows) != 0 {
		t.Fatal("dangling endpoint reached index", err)
	}
}

func TestGenericNormalizationAndLimits(t *testing.T) {
	for _, tc := range []struct{ normalizer, input, want string }{
		{"exact_nfc", `"e\u0301"`, `["é"]`},
		{"ascii_case_insensitive", `"API://A/B?Q=Ä"`, `["api://a/b?q=Ä"]`},
		{"exact_nfc", `"a\\b\"c"`, `["a\\b\"c"]`},
		{"uuid", `"AAAAAAAA-0000-0000-0000-000000000000"`, `["aaaaaaaa-0000-0000-0000-000000000000"]`},
	} {
		d := descriptor(scalar("lookup", tc.normalizer, false, Coverage{}))
		key, err := Normalize(d, "lookup", []byte(tc.input))
		if err != nil || key != tc.want {
			t.Fatalf("%s: %s %v", tc.normalizer, key, err)
		}
	}
	lookup := Lookup{Name: "version", InputType: "composite", Components: []Component{{Name: "event_id", Type: "integer", Normalization: "safe_integer"}, {Name: "provider_name", Type: "string", Normalization: "ascii_case_insensitive"}}, Coverage: Coverage{}}
	d := descriptor(lookup)
	key, err := Normalize(d, "version", []byte(`{"provider_name":"SECURITY","event_id":4624}`))
	if err != nil || key != `[4624,"security"]` {
		t.Fatal(key, err)
	}
	for _, raw := range []string{`{"event_id":4624}`, `{"event_id":4624,"provider_name":"x","event_version":null}`, `{"event_id":"4624","provider_name":"x"}`, `{"event_id":9007199254740992,"provider_name":"x"}`, `{"event_id":1.0,"provider_name":"x"}`} {
		if _, err := Normalize(d, "version", []byte(raw)); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("accepted %s: %v", raw, err)
		}
	}
	wide := Lookup{Name: "wide", InputType: "composite"}
	input := map[string]any{}
	for _, name := range []string{"a", "b", "c", "d"} {
		wide.Components = append(wide.Components, Component{Name: name, Type: "string", Normalization: "exact_nfc"})
		input[name] = strings.Repeat("é", 8192)
	}
	input["d"] = strings.Repeat("é", 8192-13) + strings.Repeat("x", 13)
	key, err = normalize(wide, input)
	if err != nil || len(key) != MaxKeyBytes {
		t.Fatalf("exact key bound: bytes=%d err=%v", len(key), err)
	}
	input["d"] = strings.Repeat("é", 8192-12) + strings.Repeat("x", 12)
	if _, err = normalize(wide, input); failureCode(err) != "content_semantic_invalid" {
		t.Fatal("one-over key admitted", err)
	}
	for _, remaining := range []int64{MaxTotalKeyBytes, 3} {
		key, err = normalizeBudget(Lookup{InputType: "scalar", Components: []Component{{Name: "value", Type: "integer", Normalization: "safe_integer"}}}, json.Number("0"), remaining)
		if err != nil || key != "[0]" {
			t.Fatal(key, err)
		}
	}
	if _, err = normalizeBudget(Lookup{InputType: "scalar", Components: []Component{{Name: "value", Type: "integer", Normalization: "safe_integer"}}}, json.Number("0"), 2); failureCode(err) != "content_semantic_invalid" {
		t.Fatal("aggregate byte excess admitted", err)
	}
}

func TestGenericDepthAndRowBoundaries(t *testing.T) {
	for _, depth := range []int{MaxDepth, MaxDepth + 1} {
		o := object("a")
		o.Data = json.RawMessage(`{"nested":` + strings.Repeat("[", depth-2) + `0` + strings.Repeat("]", depth-2) + `}`)
		_, err := Validate(context.Background(), source(t, descriptor(), []Object{o}, nil), Expected{Objects: 1, SourceArtifacts: 1}, nil)
		if depth == MaxDepth && err != nil || depth > MaxDepth && failureCode(err) != "content_schema_invalid" {
			t.Fatalf("depth %d: %v", depth, err)
		}
	}
	o := object("a")
	o.Data = json.RawMessage(`{"padding":""}`)
	base := source(t, descriptor(), []Object{o}, nil)
	padding := MaxRowBytes - len(base[ObjectsPath])
	for _, extra := range []int{0, 1} {
		o.Data = json.RawMessage(`{"padding":"` + strings.Repeat("x", padding+extra) + `"}`)
		s := source(t, descriptor(), []Object{o}, nil)
		_, err := Validate(context.Background(), s, Expected{Objects: 1, SourceArtifacts: 1}, nil)
		if extra == 0 && err != nil || extra == 1 && failureCode(err) != "content_schema_invalid" {
			t.Fatalf("row bytes=%d: %v", len(s[ObjectsPath]), err)
		}
	}
	for _, count := range []int{32, 33} {
		d := descriptor()
		for i := 0; i < count; i++ {
			d.Lookups = append(d.Lookups, scalar(fmt.Sprintf("lookup_%02d", i), "exact_nfc", false, Coverage{}))
		}
		_, err := DecodeDescriptor(canonical(t, d))
		if count == 32 && err != nil || count == 33 && failureCode(err) != "content_schema_invalid" {
			t.Fatalf("lookups=%d: %v", count, err)
		}
	}
}

func TestGenericReplacementBoundaries(t *testing.T) {
	for _, count := range []int{32, 33} {
		objects := make([]Object, count)
		for i := range objects {
			objects[i] = object(fmt.Sprintf("object-%02d", i))
			if i+1 < count {
				next := fmt.Sprintf("object-%02d", i+1)
				objects[i].Deprecated = true
				objects[i].Replacement = &next
			}
		}
		_, err := Validate(context.Background(), source(t, descriptor(), objects, nil), Expected{Objects: int64(count), SourceArtifacts: 1}, nil)
		if count == 32 && err != nil || count == 33 && failureCode(err) != "content_semantic_invalid" {
			t.Fatalf("replacement chain=%d: %v", count, err)
		}
	}
	a, b := object("a"), object("b")
	a.Deprecated = true
	b.Deprecated = true
	a.Replacement = &b.ID
	b.Replacement = &a.ID
	if _, err := Validate(context.Background(), source(t, descriptor(), []Object{a, b}, nil), Expected{Objects: 2, SourceArtifacts: 1}, nil); failureCode(err) != "content_semantic_invalid" {
		t.Fatal("cycle accepted", err)
	}
}

type brokenCloseSource struct{ memorySource }
type brokenClose struct{ io.Reader }

func (brokenClose) Close() error { return errors.New("private host path must not escape") }
func (s brokenCloseSource) Open(_ context.Context, name string) (io.ReadCloser, error) {
	return brokenClose{bytes.NewReader(s.memorySource[name])}, nil
}

func TestGenericCancellationAndCleanup(t *testing.T) {
	s := source(t, descriptor(), []Object{object("a")}, nil)
	sink := &recordingSink{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Validate(ctx, s, Expected{Objects: 1, SourceArtifacts: 1}, sink); !errors.Is(err, context.Canceled) || len(sink.rows) != 0 {
		t.Fatal(err)
	}
	if _, err := Validate(context.Background(), brokenCloseSource{s}, Expected{Objects: 1, SourceArtifacts: 1}, sink); !errors.Is(err, ErrStorage) || strings.Contains(err.Error(), "host path") || len(sink.rows) != 0 {
		t.Fatal(err)
	}
	// A later malformed relationship stream outranks an earlier semantic error.
	o := object("a")
	o.DisplayLabel = " leading"
	s = source(t, descriptor(), []Object{o}, nil)
	s[RelationshipsPath] = []byte("not-json\n")
	if _, err := Validate(context.Background(), s, Expected{Objects: 1, Relationships: 1, SourceArtifacts: 1}, sink); failureCode(err) != "content_schema_invalid" || len(sink.rows) != 0 {
		t.Fatal(err)
	}
}

func TestGenericProducerEncoding(t *testing.T) {
	d := []Lookup{scalar("alias", "ascii_case_insensitive", false, Coverage{Objects: 999})}
	a, b := object("b"), object("a")
	a.LookupValues["alias"] = []json.RawMessage{json.RawMessage(`"Z"`), json.RawMessage(`"A"`)}
	b.LookupValues["alias"] = []json.RawMessage{json.RawMessage(`"a"`)}
	members, result, err := Encode(context.Background(), d, []Object{a, b}, nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.Descriptor.Lookups[0].Coverage != (Coverage{Objects: 2, Postings: 3, DistinctKeys: 2}) {
		t.Fatal("coverage must describe actual normalized postings")
	}
	lines := bytes.Split(members[ObjectsPath], []byte{'\n'})
	if !bytes.Contains(lines[0], []byte(`"object_id":"a"`)) || !bytes.Contains(lines[1], []byte(`"alias":["A","Z"]`)) {
		t.Fatal("canonical identity and key ordering")
	}
	a.LookupValues["alias"] = []json.RawMessage{json.RawMessage(`"A"`), json.RawMessage(`"a"`)}
	if _, _, err = Encode(context.Background(), d, []Object{a}, nil, 1); err == nil {
		t.Fatal("source duplicate silently coalesced")
	}
}

func TestGenericCompleteDiagnostics(t *testing.T) {
	ctx := context.Background()
	a, b := object("a"), object("b")
	a.DisplayLabel = " secret-label "
	a.SourceRefs[0].ArtifactIndex = 4
	a.LookupValues["unrecognized_secret"] = []json.RawMessage{json.RawMessage(`"secret-input"`)}
	b.DisplayLabel = " bad-label "
	b.Description = new(string)
	*b.Description = " bad-description "
	rel := Relationship{SchemaID: RelationshipSchema, ID: "rel", Type: "link", Source: "missing-a", Target: "missing-b", SourceRefs: []SourceRef{{ArtifactIndex: 2, Locator: " bad-locator "}}, Data: json.RawMessage(`{}`)}
	src := source(t, descriptor(), []Object{a, b}, []Relationship{rel})
	expected := Expected{Objects: 2, Relationships: 1, SourceArtifacts: 1}
	if _, err := ValidateStructure(ctx, src, expected); err != nil {
		t.Fatal(err)
	}
	var first []schema.Finding
	for range 2 {
		var found []schema.Finding
		if err := SemanticsFindings(ctx, src, expected, func(f schema.Finding) error { found = append(found, f); return nil }); err != nil {
			t.Fatal(err)
		}
		if len(found) != 9 {
			t.Fatalf("incomplete semantic findings: %#v", found)
		}
		if first == nil {
			first = found
		} else if !reflect.DeepEqual(first, found) {
			t.Fatal("nondeterministic findings")
		}
		encoded := string(canonical(t, found))
		for _, secret := range []string{"secret-label", "unrecognized_secret", "secret-input", "missing-a", "bad-locator"} {
			if strings.Contains(encoded, secret) {
				t.Fatal("payload leaked into diagnostic", secret)
			}
		}
	}
	sink := &recordingSink{}
	if _, err := Validate(ctx, src, expected, sink); failureCode(err) != "content_semantic_invalid" || len(sink.rows) != 0 {
		t.Fatal("diagnostic failure permitted indexing", err)
	}
	// Multiple independently malformed rows remain enumerable; the shape walker
	// reports distinct fields without copying unknown property names.
	src[ObjectsPath] = []byte(`{"schema_id":"bad","object_id":1,"private-value":true}` + "\n" + `{"schema_id":"bad","object_id":false}` + "\n")
	var found []schema.Finding
	if err := StructureFindings(ctx, src, expected, func(f schema.Finding) error { found = append(found, f); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(found) < 20 || strings.Contains(string(canonical(t, found)), "private-value") {
		t.Fatalf("incomplete or unsafe structural findings: %#v", found)
	}
	sentinel := errors.New("sink unavailable")
	if err := StructureFindings(ctx, src, expected, func(schema.Finding) error { return sentinel }); !errors.Is(err, sentinel) {
		t.Fatal("sink failure lost", err)
	}
}
func TestGenericDiagnosticResourceStop(t *testing.T) {
	src := source(t, descriptor(), []Object{object("a")}, nil)
	src[ObjectsPath] = []byte(strings.Repeat(" ", MaxRowBytes) + "x\n")
	var found []schema.Finding
	if err := StructureFindings(context.Background(), src, Expected{Objects: 1, SourceArtifacts: 1}, func(f schema.Finding) error { found = append(found, f); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].Details.LimitID == nil || *found[0].Details.LimitID != "max_line_bytes" {
		t.Fatalf("guard finding: %#v", found)
	}
	d := descriptor(scalar("a", "exact_nfc", false, Coverage{}))
	o := object("a")
	o.LookupValues["a"] = []json.RawMessage{json.RawMessage(`"a"`)}
	for _, budget := range []Result{{Postings: MaxPostings}, {EncodedKeyBytes: MaxTotalKeyBytes}} {
		found = nil
		_, err := postingFindings(d, o, &budget, "$.objects[0]", func(f schema.Finding) error { found = append(found, f); return nil })
		if !errors.Is(err, stop) || len(found) != 1 || found[0].Details.LimitID == nil {
			t.Fatal("aggregate guard did not stop", found, err)
		}
	}
}
