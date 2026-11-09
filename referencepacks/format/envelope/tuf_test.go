package envelope

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

type trustVector struct {
	CaseID           string            `json:"case_id"`
	Bootstrap        string            `json:"bootstrap"`
	Metadata         map[string]string `json:"metadata"`
	Inventory        Inventory         `json:"inventory"`
	VerificationTime time.Time         `json:"verification_time"`
	ValidUntil       time.Time         `json:"expected_valid_until"`
	Signers          []string          `json:"expected_targets_signers"`
	RootVersion      int64             `json:"expected_root_version"`
}

func TestOfflineTUFRequiresBothRootThresholdsAndRejectsExtraSignatures_Unit(t *testing.T) {
	vector := trustVectors(t)[1]
	for _, side := range []string{"old only", "new only"} {
		t.Run(side, func(t *testing.T) {
			trust, err := AdmitBootstrap([]byte(vector.Bootstrap))
			if err != nil {
				t.Fatal(err)
			}
			old, _ := decodeRoot(trust.Root)
			m := vectorMetadata(vector)
			rootPath := "metadata/2.root.json"
			value, err := canonicaljson.DecodeStrict(m[rootPath])
			if err != nil {
				t.Fatal(err)
			}
			envelope := value.(map[string]any)
			kept := []any{}
			for _, v := range envelope["signatures"].([]any) {
				sig := v.(map[string]any)
				_, isOld := old.roles["root"].keys[sig["keyid"].(string)]
				if (side == "old only") == isOld {
					kept = append(kept, v)
				}
			}
			envelope["signatures"] = kept
			m[rootPath], err = canonicaljson.Marshal(envelope)
			if err != nil {
				t.Fatal(err)
			}
			_, err = VerifyTrust(context.Background(), m, vector.Inventory, "test.repo", vector.VerificationTime, trust)
			var failure *Failure
			if !errors.As(err, &failure) || failure.Code != "tuf_root_rotation_invalid" {
				t.Fatalf("single-threshold rotation: %v", err)
			}
		})
	}
	v := trustVectors(t)[0]
	trust, err := AdmitBootstrap([]byte(v.Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	m := vectorMetadata(v)
	value, _ := canonicaljson.DecodeStrict(m["metadata/targets.json"])
	envelope := value.(map[string]any)
	signatures := envelope["signatures"].([]any)
	root, _ := decodeRoot(trust.Root)
	unauthorized := ""
	for id := range root.roles["root"].keys {
		if _, ok := root.roles["targets"].keys[id]; !ok {
			unauthorized = id
			break
		}
	}
	signatures = append(signatures, map[string]any{"keyid": unauthorized, "sig": signatures[0].(map[string]any)["sig"]})
	slices.SortFunc(signatures, func(a, b any) int {
		return strings.Compare(a.(map[string]any)["keyid"].(string), b.(map[string]any)["keyid"].(string))
	})
	envelope["signatures"] = signatures
	m["metadata/targets.json"], _ = canonicaljson.Marshal(envelope)
	_, err = VerifyTrust(context.Background(), m, v.Inventory, "test.repo", v.VerificationTime, trust)
	var failure *Failure
	if !errors.As(err, &failure) || failure.Code != "signature_threshold_not_met" {
		t.Fatalf("valid threshold with unauthorized extra: %v", err)
	}
}

func trustVectors(t *testing.T) []trustVector {
	t.Helper()
	data, err := os.ReadFile("testdata/tuf.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []trustVector `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 3 {
		t.Fatal("missing independent trust vectors")
	}
	return fixture.Cases
}
func vectorMetadata(v trustVector) map[string][]byte {
	m := map[string][]byte{}
	for name, data := range v.Metadata {
		m[name] = []byte(data)
	}
	return m
}
func TestOfflineTUFIndependentVectorsAndRootReplay_Unit(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(map[string]any)
	}{
		{"unsupported key algorithm", func(key map[string]any) { key["keytype"] = "rsa" }},
		{"unsupported signature scheme", func(key map[string]any) { key["scheme"] = "ed25519ph" }},
		{"short public key", func(key map[string]any) { key["keyval"].(map[string]any)["public"] = strings.Repeat("00", 31) }},
		{"long public key", func(key map[string]any) { key["keyval"].(map[string]any)["public"] = strings.Repeat("00", 33) }},
		{"malformed public key", func(key map[string]any) { key["keyval"].(map[string]any)["public"] = strings.Repeat("gg", 32) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, err := canonicaljson.DecodeStrict([]byte(trustVectors(t)[0].Bootstrap))
			if err != nil {
				t.Fatal(err)
			}
			keys := value.(map[string]any)["signed"].(map[string]any)["keys"].(map[string]any)
			for _, key := range keys {
				tc.change(key.(map[string]any))
			}
			raw, err := canonicaljson.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			// Shape admission must reject before signature verification; an
			// invalidated self-signature alone would not prove key policy.
			if _, err := decodeRootShape(raw); err == nil {
				t.Fatal("unsupported root key admitted before signature verification")
			}
			if _, err := AdmitBootstrap(raw); err == nil {
				t.Fatal("unsupported bootstrap key admitted")
			}
			if _, err := VerifyHistoricalRoot(nil, raw, "test.repo"); err == nil {
				t.Fatal("unsupported historical key admitted")
			}
		})
	}
	for _, vector := range trustVectors(t) {
		prior := []byte(vector.Bootstrap)
		if _, err := VerifyHistoricalRoot(nil, prior, "test.repo"); err != nil {
			t.Fatal("historical bootstrap", err)
		}
		if next, ok := vector.Metadata["metadata/2.root.json"]; ok {
			if _, err := VerifyHistoricalRoot(prior, []byte(next), "test.repo"); err != nil {
				t.Fatal("historical rotation", err)
			}
			if _, err := VerifyHistoricalRoot([]byte(next), prior, "test.repo"); err == nil {
				t.Fatal("historical rollback accepted")
			}
			if _, err := VerifyHistoricalRoot(prior, []byte(next), "wrong.repo"); err == nil {
				t.Fatal("historical repository substitution accepted")
			}
		}
	}
	bootstrapRoot, err := canonicaljson.DecodeStrict([]byte(trustVectors(t)[0].Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	bootstrapObject := map[string]any{"schema_id": "cartulary.reference_pack_trust_bootstrap.v1", "repositories": []any{map[string]any{"repository_id": "test.repo", "trusted_root": bootstrapRoot, "trusted_root_sha256": Digest([]byte(trustVectors(t)[0].Bootstrap))}}}
	bootstrapBytes, err := canonicaljson.Marshal(bootstrapObject)
	if err != nil {
		t.Fatal(err)
	}
	if roots, err := DecodeBootstrap(bootstrapBytes); err != nil || len(roots) != 1 {
		t.Fatalf("bootstrap: %v", err)
	}
	for _, malformed := range [][]byte{append([]byte(" "), bootstrapBytes...), []byte(`{"schema_id":"cartulary.reference_pack_trust_bootstrap.v1","repositories":null}`), []byte(`{"schema_id":"cartulary.reference_pack_trust_bootstrap.v1","repositories":[]}`)} {
		if _, err := DecodeBootstrap(malformed); err == nil {
			t.Fatal("malformed bootstrap admitted")
		}
	}
	bootstrapObject["repositories"].([]any)[0].(map[string]any)["repository_id"] = "other.repo"
	bootstrapBytes, err = canonicaljson.Marshal(bootstrapObject)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeBootstrap(bootstrapBytes); err == nil {
		t.Fatal("bootstrap root rebound to another repository")
	}
	for _, v := range trustVectors(t) {
		t.Run(v.CaseID, func(t *testing.T) {
			trust, err := AdmitBootstrap([]byte(v.Bootstrap))
			if err != nil {
				t.Fatal(err)
			}
			proposal, err := VerifyTrust(context.Background(), vectorMetadata(v), v.Inventory, "test.repo", v.VerificationTime, trust)
			if err != nil {
				t.Fatal(err)
			}
			if !proposal.ValidUntil.Equal(v.ValidUntil) || len(proposal.Signers["targets"]) != 1 || proposal.Signers["targets"][0] != v.Signers[0] {
				t.Fatal("wrong trust result")
			}
			root, err := decodeRoot(proposal.Root)
			if err != nil || root.version != v.RootVersion {
				t.Fatal("wrong root proposal")
			}
			if !bytes.Equal(trust.Root, []byte(v.Bootstrap)) || len(trust.Highest) != 0 {
				t.Fatal("verification mutated trust input")
			}
			for version, raw := range proposal.RootHistory {
				trust.RootHistory[version] = raw
			}
			trust.Root = proposal.Root
			trust.Highest = proposal.Metadata
			if _, err := VerifyTrust(context.Background(), vectorMetadata(v), v.Inventory, "test.repo", v.VerificationTime, trust); err != nil {
				t.Fatalf("consumed root replay: %v", err)
			}
		})
	}
}

func TestOfflineTUFFailsClosedOnTamperingAndExpiry_Unit(t *testing.T) {
	v := trustVectors(t)[0]
	t.Run("signed extension remains authenticated", func(t *testing.T) {
		metadata := vectorMetadata(v)
		value, err := canonicaljson.DecodeStrict(metadata["metadata/timestamp.json"])
		if err != nil {
			t.Fatal(err)
		}
		value.(map[string]any)["signed"].(map[string]any)["org.example.information"] = "extension-value"
		metadata["metadata/timestamp.json"], err = canonicaljson.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		admitted, err := decodeMetadataShape(metadata["metadata/timestamp.json"], "timestamp", 64)
		if err != nil || !bytes.Contains(admitted.signedBytes, []byte(`"org.example.information":"extension-value"`)) {
			t.Fatalf("signed extension was rejected or omitted from authenticated bytes: %v", err)
		}
		trust, err := AdmitBootstrap([]byte(v.Bootstrap))
		if err != nil {
			t.Fatal(err)
		}
		_, err = VerifyTrust(context.Background(), metadata, v.Inventory, "test.repo", v.VerificationTime, trust)
		var failure *Failure
		if !errors.As(err, &failure) || failure.Code != "signature_threshold_not_met" {
			t.Fatalf("changed signed extension bypassed authentication: %v", err)
		}
	})
	for _, test := range []struct {
		name   string
		change func(map[string][]byte, Inventory, *TrustSnapshot)
		code   string
	}{
		{"signature", func(m map[string][]byte, _ Inventory, _ *TrustSnapshot) {
			raw := bytes.Clone(m["metadata/targets.json"])
			i := bytes.Index(raw, []byte(`"sig":"`)) + 7
			if raw[i] == 'a' {
				raw[i] = 'b'
			} else {
				raw[i] = 'a'
			}
			m["metadata/targets.json"] = raw
		}, "signature_threshold_not_met"},
		{"target hash", func(_ map[string][]byte, i Inventory, _ *TrustSnapshot) {
			member := i["manifest.json"]
			member.SHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			i["manifest.json"] = member
		}, "checksum_mismatch"},
		{"target length", func(_ map[string][]byte, i Inventory, _ *TrustSnapshot) {
			member := i["manifest.json"]
			member.Size++
			i["manifest.json"] = member
		}, "target_length_mismatch"},
		{"missing target", func(_ map[string][]byte, i Inventory, _ *TrustSnapshot) { i["notices/extra.txt"] = Member{} }, "target_not_declared"},
		{"unexpected target", func(_ map[string][]byte, i Inventory, _ *TrustSnapshot) { delete(i, "manifest.json") }, "unexpected_target"},
		{"rollback", func(_ map[string][]byte, _ Inventory, s *TrustSnapshot) {
			s.Highest["timestamp"] = RetainedMetadata{Version: 2}
		}, "metadata_rollback_detected"},
		{"same version changed", func(_ map[string][]byte, _ Inventory, s *TrustSnapshot) {
			s.Highest["timestamp"] = RetainedMetadata{Version: 1, Bytes: []byte("different")}
		}, "metadata_rollback_detected"},
		{"delegated metadata", func(m map[string][]byte, _ Inventory, _ *TrustSnapshot) { m["metadata/delegated.json"] = []byte("{}") }, "tuf_metadata_invalid"},
	} {
		t.Run(test.name, func(t *testing.T) {
			trust, err := AdmitBootstrap([]byte(v.Bootstrap))
			if err != nil {
				t.Fatal(err)
			}
			m := vectorMetadata(v)
			inventory := Inventory{}
			for name, member := range v.Inventory {
				inventory[name] = member
			}
			test.change(m, inventory, &trust)
			_, err = VerifyTrust(context.Background(), m, inventory, "test.repo", v.VerificationTime, trust)
			var rejected *Failure
			if !errors.As(err, &rejected) || rejected.Code != test.code {
				t.Fatalf("error=%v, want %s", err, test.code)
			}
		})
	}
	trust, err := AdmitBootstrap([]byte(v.Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	_, err = VerifyTrust(context.Background(), vectorMetadata(v), v.Inventory, "test.repo", v.ValidUntil, trust)
	var rejected *Failure
	if !errors.As(err, &rejected) || rejected.Code != "metadata_expired" {
		t.Fatalf("expiry equality=%v", err)
	}
}

func TestOfflineTUFDescriptorAdmission_Unit(t *testing.T) {
	for _, metadataDescriptor := range []bool{false, true} {
		name := "tuf_target_descriptor.v1.schema.json"
		if metadataDescriptor {
			name = "tuf_metadata_descriptor.v1.schema.json"
		}
		rule := compileProjection(name)
		base := map[string]any{"length": json.Number("123"), "hashes": map[string]any{"sha256": strings.Repeat("a", 64)}}
		if metadataDescriptor {
			base["version"] = json.Number("1")
		}
		if !rule.Matches(base) || !descriptorComplete(base, metadataDescriptor) {
			t.Fatal("valid descriptor rejected")
		}
		for _, member := range []string{"length", "hashes", "version"} {
			original, present := base[member]
			if !present {
				continue
			}
			delete(base, member)
			if !rule.Matches(base) || descriptorComplete(base, metadataDescriptor) {
				t.Fatalf("omitted %s must belong to missing-integrity check", member)
			}
			for _, invalid := range []any{nil, true, []any{}, "value", json.Number("-1")} {
				base[member] = invalid
				if rule.Matches(base) {
					t.Fatalf("invalid %s admitted", member)
				}
			}
			base[member] = original
		}
		base["unknown"] = true
		if rule.Matches(base) {
			t.Fatal("unknown descriptor member admitted")
		}
		delete(base, "unknown")
		for _, hashes := range []map[string]any{
			{"sha256": nil}, {"sha256": strings.Repeat("A", 64)}, {"sha256": strings.Repeat("a", 63)},
			{"sha256": strings.Repeat("a", 64), "sha512": strings.Repeat("b", 128)},
		} {
			base["hashes"] = hashes
			if rule.Matches(base) {
				t.Fatal("invalid digest object admitted")
			}
		}
		base["hashes"] = map[string]any{}
		if !rule.Matches(base) || descriptorComplete(base, metadataDescriptor) {
			t.Fatal("absent SHA-256 must belong to missing-integrity check")
		}
	}
}
