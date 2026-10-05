package packformat

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"reflect"
	"slices"
	"testing"
)

func TestDiagnosticRegistrySelectsOneCheckAndStableDeduplicatedPrefix_Unit(t *testing.T) {
	for _, count := range []int{1000, 1001} {
		summary, err := CheckSummary(context.Background(), "content_schema", newDiagnosticTestScratch(), func(_ context.Context, emit FindingSink) error {
			for i := 0; i < count; i++ {
				if err := emit(Finding{Path: fmt.Sprintf("$.entries[%d]", i)}); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil || summary.Total != count || summary.Retained != 1000 || summary.Truncated != (count > 1000) {
			t.Fatal("exact diagnostic retention boundary", summary, err)
		}
	}
	program := map[string]CheckFunc{}
	pass := func(context.Context, FindingSink) error { return nil }
	for _, check := range Checks() {
		program[check.ID] = pass
	}
	findings := []Finding{}
	for i := 40001; i >= 0; i-- {
		f := Finding{Path: fmt.Sprintf("$.files[%d]", i)}
		findings = append(findings, f)
	}
	// Repeats cross many spill runs, rather than only coalescing in one chunk.
	findings = append(findings, slices.Clone(findings)...)
	later := false
	program["manifest_schema"] = func(_ context.Context, emit FindingSink) error {
		for _, f := range findings {
			if err := emit(f); err != nil {
				return err
			}
		}
		return nil
	}
	program["content_schema"] = func(_ context.Context, emit FindingSink) error { later = true; return emit(Finding{Path: "$"}) }
	scratch := newDiagnosticTestScratch()
	summary, err := RunChecks(context.Background(), scratch, program)
	if err != nil {
		t.Fatal(err)
	}
	if later || summary.Total != 40002 || summary.Retained != 1000 || !summary.Truncated || summary.PrimaryIssueID == nil || *summary.PrimaryIssueID != summary.Issues[0].ID || summary.Issues[0].CheckID != "manifest_schema" {
		t.Fatalf("incorrect first-check summary: %#v", summary)
	}
	slices.Reverse(findings)
	again, err := RunChecks(context.Background(), scratch, program)
	if err != nil || !reflect.DeepEqual(summary, again) {
		t.Fatal("traversal order changed summary")
	}
	program["manifest_schema"] = pass
	program["content_schema"] = pass
	summary, err = RunChecks(context.Background(), scratch, program)
	if err != nil || summary.Result != "succeeded" || summary.PrimaryIssueID != nil || summary.Total != 0 || summary.Issues == nil {
		t.Fatal("invalid success summary")
	}
	abort := errors.New("operational abort")
	program["archive_structure"] = func(context.Context, FindingSink) error { return abort }
	summary, err = RunChecks(context.Background(), scratch, program)
	if !errors.Is(err, abort) || summary != nil {
		t.Fatal("operational abort fabricated a content verdict")
	}
	delete(program, "archive_structure")
	if _, err := RunChecks(context.Background(), scratch, program); err == nil {
		t.Fatal("incomplete check program admitted")
	}
	program["archive_structure"] = nil
	if _, err := RunChecks(context.Background(), scratch, program); err == nil {
		t.Fatal("nil check disabled required verification")
	}
	if len(scratch.files) != 0 || scratch.maximumOpen > 2 || scratch.open != 0 || scratch.maximumFiles > 8 {
		t.Fatalf("unbounded or retained scratch: files=%d maximum_files=%d readers=%d maximum_readers=%d", len(scratch.files), scratch.maximumFiles, scratch.open, scratch.maximumOpen)
	}

	t.Run("deferred owner programs", testDeferredCheckPrograms)
	t.Run("safe vocabulary and stable owned values", testDiagnosticSafety)
	t.Run("storage abort and cancellation", testDiagnosticFaults)
	t.Run("content schema complete findings", testContentSchemaFindings)
	t.Run("closed retained summary", testValidationSummaryAdmission)
	t.Run("safe numeric limit findings", testLimitDiagnostics)
}

type diagnosticTestScratch struct {
	files                           map[string][]byte
	open, maximumOpen, maximumFiles int
	failure                         string
}

func newDiagnosticTestScratch() *diagnosticTestScratch {
	return &diagnosticTestScratch{files: map[string][]byte{}}
}
func (s *diagnosticTestScratch) Write(_ context.Context, name string, write func(io.Writer) error) error {
	if s.failure == "write" {
		return io.ErrClosedPipe
	}
	if _, exists := s.files[name]; exists {
		return fs.ErrExist
	}
	var b bytes.Buffer
	if err := write(&b); err != nil {
		return err
	}
	s.files[name] = b.Bytes()
	s.maximumFiles = max(s.maximumFiles, len(s.files))
	return nil
}
func (s *diagnosticTestScratch) Open(_ context.Context, name string) (io.ReadCloser, error) {
	if s.failure == "open" {
		return nil, io.ErrClosedPipe
	}
	data, ok := s.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	if s.failure == "truncated" {
		data = data[:len(data)-1]
	}
	if s.failure == "corrupted" {
		data = slices.Clone(data)
		data[0] ^= 1
	}
	s.open++
	s.maximumOpen = max(s.maximumOpen, s.open)
	return &diagnosticTestReader{Reader: bytes.NewReader(data), scratch: s}, nil
}
func (s *diagnosticTestScratch) Remove(_ context.Context, name string) error {
	if s.failure == "remove" {
		return io.ErrClosedPipe
	}
	if _, ok := s.files[name]; !ok {
		return fs.ErrNotExist
	}
	delete(s.files, name)
	return nil
}

type diagnosticTestReader struct {
	*bytes.Reader
	scratch *diagnosticTestScratch
}

func (r *diagnosticTestReader) Close() error { r.scratch.open--; return nil }

func testDiagnosticSafety(t *testing.T) {
	check := Checks()[30]
	value := "string"
	collector := newDiagnosticCollector(context.Background(), check, nil)
	if err := collector.add(Finding{Path: "$.pack_key", Details: SafeDetails{ExpectedToken: &value}}); err != nil {
		t.Fatal(err)
	}
	value = "password"
	summary, err := collector.finish()
	if err != nil || *summary.Issues[0].Details.ExpectedToken != "string" {
		t.Fatal("caller mutated retained diagnostic")
	}
	// Fixture-only schema fields must never enlarge the runtime vocabulary.
	for _, path := range []string{"$.fixture_id", "$.expected_public_error", "$.test_name", "/srv/private/secret", "$.attacker_private_value", "$.files[01]", "$.files[-1]", "$.files[9007199254740992]", "$[\"attacker_private_value\"]", "$\n"} {
		if _, err := makeIssue(check, Finding{Path: path}); err == nil {
			t.Fatalf("unsafe path admitted %q", path)
		}
	}
	for _, path := range []string{"$", "$.files[0].sha256", `$.files[0]["<unknown>"]`, "metadata/2.root.json"} {
		if _, err := makeIssue(check, Finding{Path: path}); err != nil {
			t.Fatalf("declared path rejected %q: %v", path, err)
		}
	}
	for _, detail := range []SafeDetails{{ActualToken: &value}, {ExpectedToken: &value}, {LimitID: &value}, {RelatedPackVersion: stringPointer("secret/path")}, {RelatedPackKey: stringPointer("/etc/secret")}} {
		if _, err := makeIssue(check, Finding{Path: "$", Details: detail}); err == nil {
			t.Fatal("hostile details admitted")
		}
	}
	for _, entry := range []string{"", "raw\nvalue", string(bytes.Repeat([]byte("x"), 513))} {
		if _, err := makeIssue(check, Finding{Path: "$", EntryID: &entry}); err == nil {
			t.Fatal("unbounded or invalid identity admitted")
		}
	}
}

func testDiagnosticFaults(t *testing.T) {
	for _, fault := range []string{"write", "open", "remove", "truncated", "corrupted"} {
		t.Run(fault, func(t *testing.T) {
			scratch := newDiagnosticTestScratch()
			scratch.failure = fault
			collector := newDiagnosticCollector(context.Background(), Checks()[30], scratch)
			for i := range diagnosticChunk * 2 {
				_ = collector.add(Finding{Path: fmt.Sprintf("$.files[%d]", i)}) // A buggy caller cannot suppress an emitter failure.
			}
			if summary, err := collector.finish(); err == nil || summary != nil {
				t.Fatal("storage failure fabricated a content verdict")
			}
			if scratch.open != 0 {
				t.Fatal("failed merge leaked readers")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	collector := newDiagnosticCollector(ctx, Checks()[30], nil)
	if err := collector.add(Finding{Path: "$"}); err != nil {
		t.Fatal(err)
	}
	cancel()
	if summary, err := collector.finish(); !errors.Is(err, context.Canceled) || summary != nil {
		t.Fatal("cancellation fabricated a content verdict")
	}
}

func stringPointer(value string) *string { return &value }

func testDeferredCheckPrograms(t *testing.T) {
	ctx := context.Background()
	pass := func(context.Context, FindingSink) error { return nil }
	for _, tc := range []struct {
		name  string
		build func(context.Context) (map[string]CheckFunc, error)
	}{
		{"nil", nil},
		{"missing", func(context.Context) (map[string]CheckFunc, error) {
			return map[string]CheckFunc{"bundle_shape": pass}, nil
		}},
		{"nil callback", func(context.Context) (map[string]CheckFunc, error) {
			return map[string]CheckFunc{"bundle_shape": pass, "bundle_canonical": nil}, nil
		}},
		{"unknown", func(context.Context) (map[string]CheckFunc, error) {
			return map[string]CheckFunc{"bundle_shape": pass, "unknown": pass}, nil
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			program, err := DeferredChecks("bundle_shape", "bundle_canonical", tc.build)
			if err == nil {
				err = program["bundle_shape"](ctx, func(Finding) error { t.Fatal("incomplete program emitted verdict"); return nil })
			}
			if err == nil {
				t.Fatal("incomplete program accepted")
			}
		})
	}
	for _, bound := range [][2]string{{"unknown", "bundle_canonical"}, {"bundle_shape", "unknown"}, {"bundle_canonical", "bundle_shape"}} {
		if _, err := DeferredChecks(bound[0], bound[1], func(context.Context) (map[string]CheckFunc, error) { return nil, nil }); err == nil {
			t.Fatal("invalid range accepted", bound)
		}
	}
	if _, err := ComposeChecks(map[string]CheckFunc{"bundle_shape": pass}, map[string]CheckFunc{"bundle_shape": pass}); err == nil {
		t.Fatal("overlapping check owners accepted")
	}
	if _, err := ComposeChecks(map[string]CheckFunc{"bundle_shape": nil}); err == nil {
		t.Fatal("nil check accepted")
	}
}
