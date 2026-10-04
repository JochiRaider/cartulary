package packformat

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type manifestVector struct {
	Canonical   string `json:"canonical_manifest"`
	ManifestSHA string `json:"manifest_sha256"`
	PayloadSHA  string `json:"payload_sha256"`
}

func manifestFixture(t *testing.T) manifestVector {
	t.Helper()
	data, err := os.ReadFile("../../../../../contracts/reference-packs/fixtures/manifest.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var v manifestVector
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatal(err)
	}
	return v
}
func TestManifestIndependentIdentityVector_Unit(t *testing.T) {
	t.Run("complete manifest and path limits", testCompleteManifestAndPathLimits)
	v := manifestFixture(t)
	m, err := DecodeManifest([]byte(v.Canonical), true)
	if err != nil {
		t.Fatal(err)
	}
	if Digest([]byte(v.Canonical)) != v.ManifestSHA {
		t.Fatal("manifest digest differs from independent vector")
	}
	payload, err := PayloadDigest(m.Files)
	if err != nil || payload != v.PayloadSHA {
		t.Fatalf("payload digest=%s err=%v", payload, err)
	}
	if _, err := DecodeManifest([]byte(v.Canonical+"\n"), true); err == nil {
		t.Fatal("accepted noncanonical manifest")
	}
	value, err := canonicaljson.DecodeStrict([]byte(v.Canonical))
	if err != nil {
		t.Fatal(err)
	}
	value.(map[string]any)["built_at"] = "2026-02-30T00:00:00Z"
	invalid, err := canonicaljson.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	_, err = DecodeManifest(append(invalid, '\n'), true)
	var failure *Failure
	if !errors.As(err, &failure) || failure.Code != "manifest_schema_invalid" {
		t.Fatalf("canonicality masked invalid manifest schema: %v", err)
	}
}

func TestManifestClosesNestedOmissionsAndNulls_Unit(t *testing.T) {
	v := manifestFixture(t)
	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"missing scalar", func(m map[string]any) { delete(m, "source_version") }},
		{"missing nullable", func(m map[string]any) { delete(m, "source_as_of") }},
		{"null required", func(m map[string]any) { m["source_version"] = nil }},
		{"null array", func(m map[string]any) { m["dependencies"] = nil }},
		{"unknown top member", func(m map[string]any) { m["extra"] = true }},
		{"unknown extension", func(m map[string]any) { m["extensions"] = map[string]any{"org.example.field": "hostile-secret"} }},
		{"nested omission", func(m map[string]any) { delete(m["builder"].(map[string]any), "builder_version") }},
		{"nested unknown", func(m map[string]any) { m["builder"].(map[string]any)["extra"] = true }},
		{"nested null", func(m map[string]any) { m["compatibility"].(map[string]any)["required_capabilities"] = nil }},
		{"source omission", func(m map[string]any) { delete(m["source_artifacts"].([]any)[0].(map[string]any), "source_version") }},
		{"file omission", func(m map[string]any) { delete(m["files"].([]any)[0].(map[string]any), "size_bytes") }},
		{"wrong source time", func(m map[string]any) { m["source_as_of"] = "2026-10-03T00:00:00Z" }},
		{"invalid calendar", func(m map[string]any) { m["built_at"] = "2026-02-30T00:00:00Z" }},
		{"unknown profile", func(m map[string]any) { m["content_profile_id"] = "cartulary.missing.v1" }},
		{"wrong payload path", func(m map[string]any) { m["files"].([]any)[0].(map[string]any)["path"] = "payload/other.ndjson" }},
		{"unknown license", func(m map[string]any) { m["license"].(map[string]any)["expression"] = "Unregistered-1.0" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := canonicaljson.DecodeStrict([]byte(v.Canonical))
			if err != nil {
				t.Fatal(err)
			}
			m := value.(map[string]any)
			test.mutate(m)
			data, err := canonicaljson.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeManifest(data, true); err == nil {
				t.Fatal("accepted invalid manifest")
			}
		})
	}
}

func TestLicenseGrammarAndNoticeBindings_Unit(t *testing.T) {
	for _, text := range []string{"License\n", "é\n", strings.Repeat("line\n", 20000) + "end\n", "e\u0301\n", "\ufeffLicense\n", "invalid\x80\n", "two\n\n", "no final newline", ""} {
		want := ValidateNotice([]byte(text)) == nil
		for _, reader := range []bool{false, true} {
			source := strings.NewReader(text)
			var err error
			if reader {
				err = ValidateNoticeStream(context.Background(), iotest.OneByteReader(source))
			} else {
				err = ValidateNoticeStream(context.Background(), source)
			}
			if (err == nil) != want {
				t.Fatalf("streaming notice parity (one byte=%v, bytes=%d): %v", reader, len(text), err)
			}
		}
	}
	for _, expression := range []string{"MIT", "mIt", "MIT OR Apache-2.0 AND BSD-3-Clause", "(MIT or Apache-2.0) AND GPL-2.0+", "GPL-2.0-only WITH Classpath-exception-2.0"} {
		if !validLicense(License{Expression: expression}) {
			t.Fatalf("rejected %s", expression)
		}
	}
	for _, expression := range []string{"", "MIT OR", "MIT And Apache-2.0", "(MIT) WITH Classpath-exception-2.0", "MIT+ +", "MIT WITH AdditionRef-extra", "LicenseRef-unbound", "DocumentRef-x:LicenseRef-y", "MIT OR( )"} {
		if validLicense(License{Expression: expression}) {
			t.Fatalf("accepted %s", expression)
		}
	}
	license := License{Expression: "LicenseRef-local", Notices: []string{"notices/license.txt"}, Bindings: []LicenseBinding{{Ref: "LicenseRef-local", Path: "notices/license.txt"}}}
	if !validLicense(license) {
		t.Fatal("valid bound license rejected")
	}
	license.Expression = "LicenseRef-local OR LicenseRef-LOCAL"
	if validLicense(license) {
		t.Fatal("ambiguous spelling accepted")
	}
}

func TestPackSetIdentityIsContentOnlyAndRequiresBase_Unit(t *testing.T) {
	testContentCountReachability(t)
	if !admittedSetCardinality(64) || admittedSetCardinality(65) || admittedSetCardinality(2) {
		t.Fatal("set cardinality guard drifted")
	}
	for _, boundary := range []struct {
		kind string
		at   int64
	}{{"entry", 2000000}, {"object", 2000000}, {"relationship", 5000000}} {
		if !admittedContentCount(boundary.kind, boundary.at) || admittedContentCount(boundary.kind, boundary.at+1) || admittedContentCount(boundary.kind, -1) {
			t.Fatal("content count guard drifted", boundary.kind)
		}
	}
	var complete struct {
		Members  []SetMember `json:"members"`
		Expected string      `json:"expected_set_id"`
	}
	raw, err := os.ReadFile("../../../../../contracts/reference-packs/fixtures/complete-set.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &complete); err != nil {
		t.Fatal(err)
	}
	got, err := BuildSet(complete.Members)
	if err != nil || got.ID != complete.Expected || len(got.Members) != 16 {
		t.Fatalf("complete reachable set = %#v, %v", got, err)
	}

	members := []SetMember{}
	for _, p := range Profiles() {
		if p.Required {
			members = append(members, SetMember{Key: p.Key, Version: "1", ManifestSHA256: strings.Repeat("0", 64), PayloadSHA256: strings.Repeat("1", 64), Contract: ContractID, ProfileID: p.ID, ProfileVersion: p.Version})
		}
	}
	first, err := BuildSet(members)
	if err != nil {
		t.Fatal(err)
	}
	members[0], members[2] = members[2], members[0]
	second, err := BuildSet(members)
	if err != nil || first.ID != second.ID {
		t.Fatal("input iteration order changed set identity")
	}
	members[0].PayloadSHA256 = strings.Repeat("2", 64)
	changed, err := BuildSet(members)
	if err != nil || changed.ID == first.ID {
		t.Fatal("payload change did not change identity")
	}
	if _, err := BuildSet(members[:2]); err == nil {
		t.Fatal("missing Base registry accepted")
	}
}

func TestManifestDiagnosticEnumerationAndPrecedence_Unit(t *testing.T) {
	t.Run("manifest prerequisite closure", func(t *testing.T) {
		for _, tc := range []struct {
			data string
			size int64
			code string
		}{
			{"", -1, ""},
			{"", 1048577, "manifest_schema_invalid"},
			{"", 0, "manifest_json_invalid"},
			{`[{"x":1,"x":2}]`, 15, "manifest_json_invalid"},
			{`{"x":"\ud800","x":2}`, 20, "manifest_encoding_invalid"},
			{`{"x":1,"x":2,"bad":}`, 22, "manifest_json_invalid"},
		} {
			program, _ := ManifestChecks([]byte(tc.data), tc.size, true)
			err := runCheckRange(context.Background(), "manifest_encoding", "manifest_canonical", nil, program, true)
			var failure *Failure
			if tc.code == "" {
				if err != nil {
					t.Fatal(err)
				}
				continue
			}
			if !errors.As(err, &failure) || failure.Code != tc.code || failure.Summary == nil || failure.Summary.Total != 1 {
				t.Fatalf("prerequisite %q: %v", tc.data, err)
			}
		}
	})
	t.Run("complete duplicate findings and syntax precedence", func(t *testing.T) {
		input := []byte(`{"pack_key":"secret","pack_key":"other","builder":{"builder_id":"secret","builder_id":"secret"},"source_artifacts":[{"sha256":"secret","sha256":"secret"}],"secret":{"a":{"deep_secret":1,"deep_secret":2},"a":false}}`)
		_, err := DecodeManifestWithDiagnostics(context.Background(), input, true, newDiagnosticTestScratch())
		var failure *Failure
		if !errors.As(err, &failure) || failure.Code != "duplicate_object_member" || failure.Summary == nil || failure.Summary.Total != 4 {
			t.Fatalf("duplicate findings: %#v %v", failure, err)
		}
		want := []string{"$.builder.builder_id", "$.pack_key", "$.source_artifacts[0].sha256", `$["<unknown>"]`}
		for i, issue := range failure.Summary.Issues {
			if issue.Path != want[i] {
				t.Fatalf("duplicate path: %#v", issue)
			}
		}
		for _, malformed := range [][]byte{append(input, []byte(` {}`)...), append(input[:len(input)-1], []byte(`,"bad":}`)...)} {
			_, err := DecodeManifestWithDiagnostics(context.Background(), malformed, true, nil)
			if !errors.As(err, &failure) || failure.Code != "manifest_json_invalid" {
				t.Fatal("syntax did not precede duplicate", err)
			}
		}
	})
	v := manifestFixture(t)
	decode := func(mutate func(map[string]any)) *Failure {
		t.Helper()
		value, err := canonicaljson.DecodeStrict([]byte(v.Canonical))
		if err != nil {
			t.Fatal(err)
		}
		mutate(value.(map[string]any))
		data, err := canonicaljson.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		// A later noncanonical fault must not suppress enumerable schema faults.
		data = append(data, '\n')
		_, err = DecodeManifestWithDiagnostics(context.Background(), data, true, newDiagnosticTestScratch())
		var failure *Failure
		if !errors.As(err, &failure) || failure.Code != "manifest_schema_invalid" || failure.Summary == nil {
			t.Fatalf("missing complete schema verdict: %v", err)
		}
		if _, binaryErr := DecodeManifestSchema(data, true); binaryErr == nil {
			t.Fatal("binary reader disagreed with diagnostics")
		}
		return failure
	}
	check := func(f *Failure, paths []string) {
		t.Helper()
		if f.Summary.Total != len(paths) || f.Summary.Retained != len(paths) {
			t.Fatalf("findings: %#v", f.Summary)
		}
		for i, path := range paths {
			if f.Summary.Issues[i].Path != path {
				t.Fatalf("issue %d: %#v", i, f.Summary.Issues[i])
			}
		}
		data, err := canonicaljson.Marshal(f.Summary)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "secret") {
			t.Fatal("hostile input entered summary")
		}
		if _, err := DecodeValidationSummary(data); err != nil {
			t.Fatal(err)
		}
	}
	check(decode(func(m map[string]any) {
		delete(m["builder"].(map[string]any), "builder_version")
		m["source_identifier"] = nil
		m["private_secret"] = true
		m["other_secret"] = false
	}), []string{"$.builder.builder_version", "$.source_identifier", `$["<unknown>"]`})
	check(decode(func(m map[string]any) {
		m["builder"].(map[string]any)["builder_id"] = " private secret "
		m["source_identifier"] = " private secret "
		m["source_version"] = " private secret "
		m["source_artifacts"].([]any)[0].(map[string]any)["source_ref"] = " private secret "
	}), []string{"$.builder.builder_id", "$.source_artifacts[0].source_ref", "$.source_identifier", "$.source_version"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := DecodeManifestWithDiagnostics(ctx, []byte(v.Canonical), true, nil); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation became content rejection", err)
	}
	_, err := DecodeHintWithDiagnostics(context.Background(), []byte(`{"schema_id":null,"trust_repository_id":null,"private_secret":false}`), nil)
	var failure *Failure
	if !errors.As(err, &failure) || failure.Summary == nil || failure.Summary.Total != 3 {
		t.Fatal("hint schema lost enumerable findings", err)
	}
}

// Closed row schemas already make the entry/object/relationship counters
// unreachable at their production ceiling within one payload member. Keep the
// isolated at/one-over guards in TestPackSetIdentityIsContentOnlyAndRequiresBase.
// This conservative byte proof is computed only from executable projections.
func testContentCountReachability(t *testing.T) {
	t.Helper()
	for key, shape := range contentShapes {
		count := int64(2000000)
		if strings.HasSuffix(key, "/relationship") {
			count = 5000000
		}
		lower := minimumShapeJSONBytes(shape) + 1 // required NDJSON LF
		if lower <= 0 || count*lower <= 268435456 {
			t.Fatal("count reachability changed; add a complete equality fixture", key, lower)
		}
	}
}

func minimumShapeJSONBytes(s *shape) int64 {
	if s.hasConstant {
		raw, _ := canonicaljson.Marshal(s.constant)
		return int64(len(raw))
	}
	if len(s.enum) > 0 {
		shortest := int64(1 << 62)
		for _, value := range s.enum {
			raw, _ := canonicaljson.Marshal(value)
			shortest = min(shortest, int64(len(raw)))
		}
		return shortest
	}
	var n int64
	switch s.kind {
	case "string":
		n = 2
		if s.minLength != nil {
			n += int64(*s.minLength)
		}
	case "boolean":
		n = 4
	case "integer", "number":
		n = 1
	case "array":
		n = 2
		if s.minItems != nil && *s.minItems > 0 {
			n += int64(*s.minItems-1) + int64(*s.minItems)*minimumShapeJSONBytes(s.items)
		}
	case "object":
		n = 2 + int64(max(0, len(s.required)-1))
		for _, key := range s.required {
			raw, _ := canonicaljson.Marshal(key)
			n += int64(len(raw)) + 1 + minimumShapeJSONBytes(s.properties[key])
		}
	}
	if s.nullable {
		n = min(n, 4)
	}
	// Ignoring conditional and pattern restrictions only weakens a lower bound.
	return n
}
