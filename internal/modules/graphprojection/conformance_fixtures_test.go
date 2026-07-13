package graphprojection

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/fixturetest"
)

type fixtureDigests struct {
	GraphViewID            string `json:"graph_view_id"`
	ProjectionConfigDigest string `json:"projection_config_digest"`
	ProjectionSourceDigest string `json:"projection_source_digest"`
}

func TestGPFIX022MinimalGraphDigestTranscript(t *testing.T) {
	assertFixtureDigests(t, "GP-FIX-022")
}

func TestGPFIX023HostGraphDigestTranscript(t *testing.T) {
	assertFixtureDigests(t, "GP-FIX-023")
}

func TestGPFIX014CanonicalJSONString(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-014")
	actual, err := CanonicalJSON(map[string]any{"b": "quote\" slash/ reverse\\ control\u0001\u2028", "a": []any{1, true, nil}})
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"a\":[1,true,null],\"b\":\"quote\\\" slash/ reverse\\\\ control\\u0001\u2028\"}"
	if err := fixturetest.CompareBytes("canonical JSON", actual, []byte(want)); err != nil {
		t.Fatal(err)
	}
}

func TestGPFIX015IntegerLexicalBoundaries(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-015")
	for _, value := range []string{"0", "-1", "9223372036854775807"} {
		if !finiteIntegerPattern.MatchString(value) {
			t.Fatalf("valid integer %q rejected", value)
		}
	}
	for _, value := range []string{"01", "-0", "+1", "1.0", "1e2", ""} {
		if finiteIntegerPattern.MatchString(value) {
			t.Fatalf("invalid integer %q accepted", value)
		}
	}
}

func TestGPFIX016TimestampCalendarValidity(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-016")
	if _, err := parseTimestamp("2024-02-29T23:59:59Z"); err != nil {
		t.Fatalf("valid leap timestamp rejected: %v", err)
	}
	for _, value := range []string{"2025-02-29T00:00:00Z", "2026-13-01T00:00:00Z", "2026-05-30T00:00:00+01:00"} {
		if _, err := parseTimestamp(value); err == nil {
			t.Fatalf("invalid timestamp %q accepted", value)
		}
	}
}

func TestGPFIX017IdentifierUnicodeWhitespace(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-017")
	for _, value := range []string{" host", "host ", "\u00a0host", "host\u2028", "\thost"} {
		if validIdentifier(value) {
			t.Fatalf("identifier with boundary whitespace %q accepted", value)
		}
	}
	if !validIdentifier("host_01") {
		t.Fatal("ordinary identifier rejected")
	}
}

func TestGPFIX036DistinctSortedArrayCanonicalOrder(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-036")
	values := []string{canonicalValueKey("a\\"), canonicalValueKey("a\""), canonicalValueKey("a/")}
	sort.Strings(values)
	want := []string{`"a/"`, `"a\""`, `"a\\"`}
	for index := range want {
		if values[index] != want[index] {
			t.Fatalf("canonical sort = %#v want %#v", values, want)
		}
	}
}

func assertFixtureDigests(t *testing.T, fixtureID string) {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := fixturetest.RepoRoot(workingDirectory)
	if err != nil {
		t.Fatal(err)
	}
	manifest, directory, err := fixturetest.Load(root, fixtureID)
	if err != nil {
		t.Fatalf("load %s: %v", fixtureID, err)
	}
	input, err := fixturetest.ReadArtifact(directory, manifest.Steps[0].InputArtifact)
	if err != nil {
		t.Fatal(err)
	}
	expectedBody, err := fixturetest.ReadArtifact(directory, manifest.Steps[0].ExpectedArtifact)
	if err != nil {
		t.Fatal(err)
	}
	var expected fixtureDigests
	if err := json.Unmarshal(expectedBody, &expected); err != nil {
		t.Fatal(err)
	}
	run, err := AdmitRetainedProjection(input, manifest.Determinism.Nonce, time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("admit %s: %v", fixtureID, err)
	}
	if run.GraphViewID != expected.GraphViewID || run.ProjectionConfigDigest != expected.ProjectionConfigDigest || run.ProjectionSourceDigest != expected.ProjectionSourceDigest {
		t.Fatalf("%s digests = %#v; expected %#v", fixtureID, fixtureDigests{run.GraphViewID, run.ProjectionConfigDigest, run.ProjectionSourceDigest}, expected)
	}
	if fixtureID != "GP-FIX-022" && fixtureID != "GP-FIX-023" {
		return
	}
	actualConfig, actualSource, err := ProjectionDigestTranscripts(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, transcript := range []struct {
		name   string
		path   string
		actual []byte
	}{
		{name: "config tuple", path: "expected.config-tuple.hex", actual: actualConfig},
		{name: "source tuple", path: "expected.source-tuple.hex", actual: actualSource},
	} {
		expectedHex, err := fixturetest.ReadArtifact(directory, transcript.path)
		if err != nil {
			t.Fatal(err)
		}
		expectedBytes, err := hex.DecodeString(string(expectedHex[:len(expectedHex)-1]))
		if err != nil {
			t.Fatal(err)
		}
		if err := fixturetest.CompareBytes(transcript.name, transcript.actual, expectedBytes); err != nil {
			t.Fatal(err)
		}
	}
}

func assertFixtureLoaded(t *testing.T, fixtureID string) {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := fixturetest.RepoRoot(workingDirectory)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := fixturetest.Load(root, fixtureID); err != nil {
		t.Fatalf("load %s: %v", fixtureID, err)
	}
}
