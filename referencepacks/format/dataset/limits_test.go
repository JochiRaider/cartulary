package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

type limitCase struct {
	Limit              string
	Maximum            int
	Phase, Code, Probe string
}

// Authored expected limits are independent of implementation constants. The
// aggregate cases use controlled near-limit state; the public streaming path
// exercises the same posting budget with single and multiple lookup names.
func TestGenericLimitMatrix(t *testing.T) {
	raw, err := os.ReadFile("limit_matrix.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []limitCase
	if err = json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	t.Run("interactions", testLimitInteractions)
	for _, tc := range cases {
		t.Run(tc.Probe, func(t *testing.T) {
			for _, extra := range []int{0, 1} {
				t.Run(fmt.Sprint(extra), func(t *testing.T) {
					n := tc.Maximum + extra
					var err error
					sink := &recordingSink{}
					d := descriptor()
					o := object("a")
					src := source(t, d, []Object{o}, nil)
					switch tc.Probe {
					case "descriptor_bytes":
						// Byte admission is separate from the closed descriptor shape. No padding
						// field is introduced to manufacture an otherwise impossible valid shape.
						src[DescriptorPath] = bytes.Repeat([]byte(" "), n)
						var got []byte
						got, err = descriptorBytes(context.Background(), src)
						if extra == 0 && len(got) != n {
							t.Fatal("equality not fully read")
						}
					case "lookup_count":
						for i := 0; i < n; i++ {
							d.Lookups = append(d.Lookups, scalar(fmt.Sprintf("lookup_%02d", i), "exact_nfc", false, Coverage{}))
						}
						src = source(t, d, []Object{o}, nil)
					case "components":
						l := Lookup{Name: "composite", InputType: "composite"}
						for i := 0; i < n; i++ {
							l.Components = append(l.Components, Component{Name: fmt.Sprintf("c%d", i), Type: "string", Normalization: "exact_nfc"})
						}
						src = source(t, descriptor(l), []Object{o}, nil)
					case "object_postings":
						for _, split := range []bool{false, true} {
							checkPostingBoundary(t, n, split, tc.Code, tc.Limit)
						}
						return
					case "depth":
						o.Data = json.RawMessage(`{"nested":` + strings.Repeat("[", n-2) + `0` + strings.Repeat("]", n-2) + `}`)
						src = source(t, d, []Object{o}, nil)
					case "row_bytes":
						o.Data = json.RawMessage(`{"padding":""}`)
						src = source(t, d, []Object{o}, nil)
						o.Data = json.RawMessage(`{"padding":"` + strings.Repeat("x", n-len(src[ObjectsPath])) + `"}`)
						src = source(t, d, []Object{o}, nil)
					case "key_bytes":
						l := Lookup{Name: "key", InputType: "composite", Coverage: Coverage{Objects: 1, Postings: 1, DistinctKeys: 1}}
						v := map[string]string{}
						for _, name := range []string{"a", "b", "c", "d"} {
							l.Components = append(l.Components, Component{Name: name, Type: "string", Normalization: "exact_nfc"})
							v[name] = strings.Repeat("é", 8192)
						}
						v["d"] = strings.Repeat("é", 8192-13+extra) + strings.Repeat("x", 13-extra)
						o.LookupValues["key"] = []json.RawMessage{canonical(t, v)}
						src = source(t, descriptor(l), []Object{o}, nil)
					case "aggregate_postings", "aggregate_key_bytes":
						l := scalar("a", "exact_nfc", false, Coverage{})
						d = descriptor(l)
						o.LookupValues["a"] = []json.RawMessage{json.RawMessage(`"a"`)}
						budget := Result{}
						if tc.Probe == "aggregate_postings" {
							budget.Postings = int64(n - 1)
						} else {
							budget.EncodedKeyBytes = int64(n - 5)
						}
						before := budget
						_, err = objectPostings(d, o, &budget)
						if extra == 0 && (tc.Probe == "aggregate_postings" && budget.Postings != int64(tc.Maximum) || tc.Probe == "aggregate_key_bytes" && budget.EncodedKeyBytes != int64(tc.Maximum)) {
							t.Fatal("equality did not consume exact budget")
						}
						if extra == 1 && (budget.Postings != before.Postings || budget.EncodedKeyBytes != before.EncodedKeyBytes) {
							t.Fatal("excess mutated budget")
						}
					default:
						t.Fatal("unimplemented matrix row", tc.Probe)
					}
					if tc.Probe != "descriptor_bytes" && !strings.HasPrefix(tc.Probe, "aggregate_") {
						_, err = Validate(context.Background(), src, Expected{Objects: 1, SourceArtifacts: 1}, sink)
					}
					if extra == 0 {
						if err != nil {
							t.Fatal("equality rejected", err)
						}
						return
					}
					assertLimitFailure(t, err, tc.Code, tc.Limit)
					if len(sink.rows) != 0 {
						t.Fatal("failed admission published index")
					}
				})
			}
		})
	}
}

// Cancellation after a member opens must close that stream, publish nothing,
// and remain an operational cancellation rather than an invalid-content verdict.
type cancelMemberSource struct {
	memorySource
	cancel context.CancelFunc
	closed int
}
type cancelMember struct {
	io.Reader
	owner *cancelMemberSource
}

func (r cancelMember) Read(p []byte) (int, error) { r.owner.cancel(); return r.Reader.Read(p) }
func (r cancelMember) Close() error               { r.owner.closed++; return nil }
func (s *cancelMemberSource) Open(_ context.Context, name string) (io.ReadCloser, error) {
	return cancelMember{bytes.NewReader(s.memorySource[name]), s}, nil
}

func testLimitInteractions(t *testing.T) {
	for _, coverage := range []Coverage{{Objects: 1000001}, {Postings: 8000001}, {DistinctKeys: 8000001}} {
		d := descriptor(scalar("a", "exact_nfc", false, coverage))
		sink := &recordingSink{}
		_, err := Validate(context.Background(), source(t, d, []Object{object("a")}, nil), Expected{Objects: 1, SourceArtifacts: 1}, sink)
		if failureCode(err) != "content_semantic_invalid" || len(sink.rows) != 0 {
			t.Fatal("coverage resource phase", err)
		}
	}
	// NFC expands U+0344 into two code points, while JSON escaping expands a quote.
	for _, input := range []string{"\u0344", "\""} {
		l := scalar("a", "exact_nfc", false, Coverage{Objects: 1, Postings: 1, DistinctKeys: 1})
		o := object("a")
		o.LookupValues["a"] = []json.RawMessage{canonical(t, input)}
		sink := &recordingSink{}
		result, err := Validate(context.Background(), source(t, descriptor(l), []Object{o}, nil), Expected{Objects: 1, SourceArtifacts: 1}, sink)
		if err != nil || len(sink.rows) != 1 || result.EncodedKeyBytes != int64(len(sink.rows[0].Postings[0].Key)) {
			t.Fatal("normalized encoded budget", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	src := &cancelMemberSource{memorySource: source(t, descriptor(), []Object{object("a")}, nil), cancel: cancel}
	sink := &recordingSink{}
	_, err := Validate(ctx, src, Expected{Objects: 1, SourceArtifacts: 1}, sink)
	if !errors.Is(err, context.Canceled) || src.closed != 1 || len(sink.rows) != 0 {
		t.Fatal("cancellation cleanup", src.closed, err)
	}
}

func assertLimitFailure(t *testing.T, err error, code, limit string) {
	t.Helper()
	var f *Failure
	if !errors.As(err, &f) || f.Code != code || f.Limit != limit || f.Observed <= f.Maximum {
		t.Fatalf("failure=%+v want %s/%s", err, code, limit)
	}
	if len(err.Error()) > 128 {
		t.Fatal("unbounded diagnostic")
	}
}
func checkPostingBoundary(t *testing.T, n int, split bool, code, limit string) {
	t.Helper()
	d := descriptor()
	o := object("a")
	for i := 0; i < n; i++ {
		name := "a"
		if split && i%2 == 1 {
			name = "b"
		}
		o.LookupValues[name] = append(o.LookupValues[name], json.RawMessage(fmt.Sprintf("\"%03d\"", i)))
	}
	for _, name := range []string{"a", "b"} {
		if v, ok := o.LookupValues[name]; ok {
			d.Lookups = append(d.Lookups, scalar(name, "exact_nfc", false, Coverage{Objects: 1, Postings: int64(len(v)), DistinctKeys: int64(len(v))}))
		}
	}
	src := source(t, d, []Object{o}, nil)
	sink := &recordingSink{}
	_, err := Validate(context.Background(), src, Expected{Objects: 1, SourceArtifacts: 1}, sink)
	if n == 128 {
		if err != nil || len(sink.rows) != 1 || len(sink.rows[0].Postings) != 128 {
			t.Fatal("posting equality", split, err)
		}
		return
	}
	assertLimitFailure(t, err, code, limit)
	if len(sink.rows) != 0 {
		t.Fatal("index published")
	}
	decoded, err := decodeObjectContext(context.Background(), bytes.TrimSuffix(src[ObjectsPath], []byte{'\n'}))
	if err != nil {
		t.Fatal(err)
	}
	retained := 0
	for _, inputs := range decoded.LookupValues {
		retained += len(inputs)
	}
	if retained != 128 || decoded.inputPostings != 129 {
		t.Fatal("excess inputs materialized", retained, decoded.inputPostings)
	}
	// A malformed scalar after the excess inputs, and a malformed later member,
	// both retain structural precedence over the already-discovered resource error.
	src[ObjectsPath] = bytes.Replace(src[ObjectsPath], []byte(`"128"`), []byte(`false`), 1)
	_, err = Validate(context.Background(), src, Expected{Objects: 1, SourceArtifacts: 1}, sink)
	if failureCode(err) != "content_schema_invalid" || len(sink.rows) != 0 {
		t.Fatal("late structural error masked", err)
	}
	src = source(t, d, []Object{o}, nil)
	src[RelationshipsPath] = []byte("not-json\n")
	_, err = Validate(context.Background(), src, Expected{Objects: 1, Relationships: 1, SourceArtifacts: 1}, sink)
	if failureCode(err) != "content_schema_invalid" || len(sink.rows) != 0 {
		t.Fatal("later member error masked", err)
	}
}
