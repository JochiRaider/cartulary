package packformat

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func runTrustManifest(t *testing.T, input []byte, expected map[string]any) {
	t.Helper()
	value, err := canonicaljson.DecodeStrict(input)
	if err != nil || !compileProjection("trust_fixture.v1.schema.json").matches(value) {
		t.Fatal("invalid trust fixture", err)
	}
	var fixture struct {
		trustVector
		Versions      map[string]int64    `json:"retained_versions"`
		ForgetHistory bool                `json:"forget_root_history"`
		ConsumeRoots  bool                `json:"consume_roots"`
		Signers       map[string][]string `json:"expected_role_signers"`
		Transitions   []RootTransition    `json:"expected_root_transitions"`
	}
	if err := json.Unmarshal(input, &fixture); err != nil {
		t.Fatal(err)
	}
	trust, err := AdmitBootstrap([]byte(fixture.Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	files := vectorMetadata(fixture.trustVector)
	if fixture.ConsumeRoots {
		// This models an exact retained history, not a second bootstrap or a
		// newly authorized root. The subsequent replay must produce no rotation.
		for version := int64(2); version <= fixture.RootVersion; version++ {
			root := files["metadata/2.root.json"]
			trust.RootHistory[version] = bytes.Clone(root)
			trust.Root = bytes.Clone(root)
		}
		for _, role := range []string{"targets", "snapshot", "timestamp"} {
			var envelope struct {
				Signed struct {
					Version int64 `json:"version"`
				} `json:"signed"`
			}
			data := files["metadata/"+role+".json"]
			if err := json.Unmarshal(data, &envelope); err != nil {
				t.Fatal(err)
			}
			trust.Highest[role] = RetainedMetadata{Version: envelope.Signed.Version, Bytes: bytes.Clone(data)}
		}
	}
	if fixture.ForgetHistory {
		trust.RootHistory = map[int64][]byte{}
	}
	for role, version := range fixture.Versions {
		trust.Highest[role] = RetainedMetadata{Version: version, Bytes: []byte("retained higher revision")}
	}
	before, err := json.Marshal(trust)
	if err != nil {
		t.Fatal(err)
	}
	proposal, verdict := VerifyTrustWithDiagnostics(context.Background(), files, fixture.Inventory, "test.repo", fixture.VerificationTime, trust, newDiagnosticTestScratch())
	after, err := json.Marshal(trust)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("verification mutated retained trust", err)
	}
	assertManifestVerdict(t, expected, verdict)
	if verdict == nil {
		root, err := decodeRoot(proposal.Root)
		if err != nil || root.version != fixture.RootVersion || !proposal.ValidUntil.Equal(fixture.ValidUntil) || !reflect.DeepEqual(proposal.Signers, fixture.Signers) || !reflect.DeepEqual(proposal.RootTransitions, fixture.Transitions) {
			t.Fatalf("trusted result changed: %#v, %v", proposal.RootTransitions, err)
		}
	}
}
