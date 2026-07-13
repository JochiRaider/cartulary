package graphprojection

import (
	"encoding/hex"
	"encoding/json"
	"os"
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
	verifyUnitFixture(t, "GP-FIX-014")
}

func TestGPFIX015IntegerLexicalBoundaries(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-015")
}

func TestGPFIX016TimestampCalendarValidity(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-016")
}

func TestGPFIX017IdentifierUnicodeWhitespace(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-017")
}

func TestGPFIX036DistinctSortedArrayCanonicalOrder(t *testing.T) {
	verifyUnitFixture(t, "GP-FIX-036")
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
