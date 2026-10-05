package packformat

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

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
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
			_, err = VerifyTrust(m, vector.Inventory, "test.repo", vector.VerificationTime, trust)
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
	_, err = VerifyTrust(m, v.Inventory, "test.repo", v.VerificationTime, trust)
	var failure *Failure
	if !errors.As(err, &failure) || failure.Code != "signature_threshold_not_met" {
		t.Fatalf("valid threshold with unauthorized extra: %v", err)
	}
}

func trustVectors(t *testing.T) []trustVector {
	t.Helper()
	data, err := os.ReadFile("../../../../../contracts/reference-pack-fixtures/fixtures/tuf.v1.json")
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
			proposal, err := VerifyTrust(vectorMetadata(v), v.Inventory, "test.repo", v.VerificationTime, trust)
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
			if _, err := VerifyTrust(vectorMetadata(v), v.Inventory, "test.repo", v.VerificationTime, trust); err != nil {
				t.Fatalf("consumed root replay: %v", err)
			}
		})
	}
}

func TestOfflineTUFFailsClosedOnTamperingAndExpiry_Unit(t *testing.T) {
	t.Run("complete metadata byte boundaries", testCompleteTUFByteBoundaries)
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
		_, err = VerifyTrust(metadata, v.Inventory, "test.repo", v.VerificationTime, trust)
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
			_, err = VerifyTrust(m, inventory, "test.repo", v.VerificationTime, trust)
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
	_, err = VerifyTrust(vectorMetadata(v), v.Inventory, "test.repo", v.ValidUntil, trust)
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
		if !rule.matches(base) || !descriptorComplete(base, metadataDescriptor) {
			t.Fatal("valid descriptor rejected")
		}
		for _, member := range []string{"length", "hashes", "version"} {
			original, present := base[member]
			if !present {
				continue
			}
			delete(base, member)
			if !rule.matches(base) || descriptorComplete(base, metadataDescriptor) {
				t.Fatalf("omitted %s must belong to missing-integrity check", member)
			}
			for _, invalid := range []any{nil, true, []any{}, "value", json.Number("-1")} {
				base[member] = invalid
				if rule.matches(base) {
					t.Fatalf("invalid %s admitted", member)
				}
			}
			base[member] = original
		}
		base["unknown"] = true
		if rule.matches(base) {
			t.Fatal("unknown descriptor member admitted")
		}
		delete(base, "unknown")
		for _, hashes := range []map[string]any{
			{"sha256": nil}, {"sha256": strings.Repeat("A", 64)}, {"sha256": strings.Repeat("a", 63)},
			{"sha256": strings.Repeat("a", 64), "sha512": strings.Repeat("b", 128)},
		} {
			base["hashes"] = hashes
			if rule.matches(base) {
				t.Fatal("invalid digest object admitted")
			}
		}
		base["hashes"] = map[string]any{}
		if !rule.matches(base) || descriptorComplete(base, metadataDescriptor) {
			t.Fatal("absent SHA-256 must belong to missing-integrity check")
		}
	}
}

func TestOfflineTUFCombinedFaultPrecedence_Unit(t *testing.T) {
	v := trustVectors(t)[0]
	trust, err := AdmitBootstrap([]byte(v.Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, code string
		change     func(map[string][]byte, Inventory, *TrustSnapshot)
	}{
		{"schema before signature", "tuf_metadata_invalid", func(m map[string][]byte, _ Inventory, _ *TrustSnapshot) {
			raw := bytes.Clone(m["metadata/timestamp.json"])
			at := bytes.Index(raw, []byte(`"sig":"`)) + 7
			raw[at] = '0'
			m["metadata/timestamp.json"] = raw
			m["metadata/targets.json"] = []byte(`{"signed":null,"signatures":[]}`)
		}},
		{"signature before rollback", "signature_threshold_not_met", func(m map[string][]byte, _ Inventory, s *TrustSnapshot) {
			s.Highest["timestamp"] = RetainedMetadata{Version: 2}
			raw := bytes.Clone(m["metadata/targets.json"])
			at := bytes.Index(raw, []byte(`"sig":"`)) + 7
			if raw[at] == '0' {
				raw[at] = '1'
			} else {
				raw[at] = '0'
			}
			m["metadata/targets.json"] = raw
		}},
		{"missing role before signature", "tuf_metadata_invalid", func(m map[string][]byte, _ Inventory, _ *TrustSnapshot) {
			delete(m, "metadata/snapshot.json")
			value, _ := canonicaljson.DecodeStrict(m["metadata/targets.json"])
			envelope := value.(map[string]any)
			envelope["signatures"].([]any)[0].(map[string]any)["sig"] = strings.Repeat("0", 128)
			m["metadata/targets.json"], _ = canonicaljson.Marshal(envelope)
		}},
		{"role schema before malformed root filename", "tuf_metadata_invalid", func(m map[string][]byte, _ Inventory, _ *TrustSnapshot) {
			m["metadata/01.root.json"] = []byte(v.Bootstrap)
			m["metadata/targets.json"] = []byte(`{"signed":null,"signatures":[]}`)
		}},
		{"canonical bytes before malformed root filename", "metadata_noncanonical", func(m map[string][]byte, _ Inventory, _ *TrustSnapshot) {
			m["metadata/01.root.json"] = []byte(v.Bootstrap)
			m["metadata/targets.json"] = append(m["metadata/targets.json"], '\n')
		}},
		{"nested target schema before signature", "tuf_metadata_invalid", func(m map[string][]byte, _ Inventory, _ *TrustSnapshot) {
			value, _ := canonicaljson.DecodeStrict(m["metadata/targets.json"])
			envelope := value.(map[string]any)
			targets := envelope["signed"].(map[string]any)["targets"].(map[string]any)
			targets["manifest.json"].(map[string]any)["length"] = nil
			m["metadata/targets.json"], _ = canonicaljson.Marshal(envelope)
		}},
		{"unknown descriptor member before signature", "tuf_metadata_invalid", func(m map[string][]byte, _ Inventory, _ *TrustSnapshot) {
			value, _ := canonicaljson.DecodeStrict(m["metadata/snapshot.json"])
			envelope := value.(map[string]any)
			meta := envelope["signed"].(map[string]any)["meta"].(map[string]any)
			meta["targets.json"].(map[string]any)["hostile-unknown-member"] = "payload-must-not-escape"
			m["metadata/snapshot.json"], _ = canonicaljson.Marshal(envelope)
		}},
		{"length before hash", "target_length_mismatch", func(_ map[string][]byte, i Inventory, _ *TrustSnapshot) {
			first := i["bundle.json"]
			first.SHA256 = strings.Repeat("f", 64)
			i["bundle.json"] = first
			last := i["payload/entries.ndjson"]
			last.Size++
			i["payload/entries.ndjson"] = last
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			m := vectorMetadata(v)
			inventory := Inventory{}
			for k, x := range v.Inventory {
				inventory[k] = x
			}
			s := trust
			s.Highest = map[string]RetainedMetadata{}
			test.change(m, inventory, &s)
			_, err := VerifyTrust(m, inventory, "test.repo", v.VerificationTime, s)
			var failure *Failure
			if !errors.As(err, &failure) || failure.Code != test.code {
				t.Fatalf("got %v; want %s", err, test.code)
			}
		})
	}
}

func TestOfflineTUFCompleteSchemaDiagnostics_Unit(t *testing.T) {
	t.Run("target binding omission differs from invalid present members", func(t *testing.T) {
		v := trustVectors(t)[0]
		for _, variant := range []string{"missing", "null", "unknown", "wrong_type"} {
			files := vectorMetadata(v)
			value, err := canonicaljson.DecodeStrict(files["metadata/targets.json"])
			if err != nil {
				t.Fatal(err)
			}
			signed := value.(map[string]any)["signed"].(map[string]any)
			switch variant {
			case "missing":
				delete(signed, "cartulary")
			case "null":
				signed["cartulary"] = nil
			case "unknown":
				signed["cartulary"].(map[string]any)["private_secret"] = true
			case "wrong_type":
				signed["cartulary"].(map[string]any)["pack_release_sequence"] = "1"
			}
			files["metadata/targets.json"], err = canonicaljson.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			summary, err := TrustSchemaSummary(context.Background(), files, nil)
			if err != nil || (summary.Result == "succeeded") != (variant == "missing") {
				t.Fatal("target binding shape precedence", variant, summary, err)
			}
		}
	})

	v := trustVectors(t)[0]
	trust, err := AdmitBootstrap([]byte(v.Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	files := vectorMetadata(v)
	change := func(name string, edit func(map[string]any)) {
		value, err := canonicaljson.DecodeStrict(files[name])
		if err != nil {
			t.Fatal(err)
		}
		edit(value.(map[string]any))
		files[name], err = canonicaljson.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
	}
	change("metadata/timestamp.json", func(m map[string]any) {
		m["signatures"].([]any)[0].(map[string]any)["sig"] = strings.Repeat("0", 128)
		m["signed"].(map[string]any)["private_secret"] = "allowed signed extension"
	})
	change("metadata/snapshot.json", func(m map[string]any) {
		delete(m["signatures"].([]any)[0].(map[string]any), "sig")
		m["private_secret"] = true
	})
	change("metadata/targets.json", func(m map[string]any) { m["signed"].(map[string]any)["delegations"] = nil })
	_, err = VerifyTrustWithDiagnostics(context.Background(), files, v.Inventory, "test.repo", v.VerificationTime, trust, newDiagnosticTestScratch())
	var failure *Failure
	if !errors.As(err, &failure) || failure.Code != "tuf_metadata_invalid" || failure.Summary == nil {
		t.Fatal("signature error masked schema enumeration", err)
	}
	want := []string{"$.metadata[0].signatures[0].sig", `$.metadata[0]["<unknown>"]`, "$.metadata[1].signed.delegations"}
	if failure.Summary.Total != len(want) {
		t.Fatalf("incomplete diagnostics: %#v", failure.Summary)
	}
	for i, path := range want {
		if failure.Summary.Issues[i].Path != path {
			t.Fatalf("wrong issue %d: %#v", i, failure.Summary.Issues[i])
		}
	}
	encoded, err := canonicaljson.Marshal(failure.Summary)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "private_secret") || strings.Contains(string(encoded), "allowed signed extension") {
		t.Fatal("metadata diagnostic leaked input")
	}
	if _, err := DecodeValidationSummary(encoded); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := TrustSchemaSummary(ctx, files, nil); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation became metadata verdict", err)
	}
}

// Mutations below leave authentication facts independent of the implementation:
// signed vectors come from the retained external producer, and expected check
// IDs/counts are literal obligations rather than generated verifier output.
func TestOfflineTUFRegistryProgramDiagnostics_Unit(t *testing.T) {
	vector := trustVectors(t)[0]
	cases := []struct {
		name, check string
		count       int
		mutate      func(map[string][]byte, Inventory, *TrustSnapshot, *time.Time)
	}{
		{"canonical before retained root", "tuf_canonical", 2, func(files map[string][]byte, _ Inventory, trust *TrustSnapshot, _ *time.Time) {
			files["metadata/targets.json"] = append(files["metadata/targets.json"], '\n')
			files["metadata/timestamp.json"] = append(files["metadata/timestamp.json"], '\n')
			trust.RootHistory = map[int64][]byte{}
		}},
		{"retained root before signature", "root_trust", 1, func(files map[string][]byte, _ Inventory, trust *TrustSnapshot, _ *time.Time) {
			trust.RootHistory = map[int64][]byte{}
			value, _ := canonicaljson.DecodeStrict(files["metadata/timestamp.json"])
			value.(map[string]any)["signatures"].([]any)[0].(map[string]any)["sig"] = strings.Repeat("0", 128)
			files["metadata/timestamp.json"], _ = canonicaljson.Marshal(value)
		}},
		{"every invalid role signature and threshold", "role_signature", 6, func(files map[string][]byte, _ Inventory, _ *TrustSnapshot, _ *time.Time) {
			for _, name := range []string{"timestamp", "snapshot", "targets"} {
				path := "metadata/" + name + ".json"
				value, _ := canonicaljson.DecodeStrict(files[path])
				value.(map[string]any)["signatures"].([]any)[0].(map[string]any)["sig"] = strings.Repeat("0", 128)
				files[path], _ = canonicaljson.Marshal(value)
			}
		}},
		{"every rollback scope", "metadata_rollback", 3, func(_ map[string][]byte, _ Inventory, trust *TrustSnapshot, _ *time.Time) {
			for _, name := range []string{"timestamp", "snapshot", "targets"} {
				trust.Highest[name] = RetainedMetadata{Version: 9007199254740991, Bytes: []byte("unusable")}
			}
		}},
		{"all expired roles", "metadata_expiry", 4, func(_ map[string][]byte, _ Inventory, trust *TrustSnapshot, at *time.Time) {
			root, _ := decodeRoot(trust.Root)
			*at = root.expires
		}},
		{"all undeclared inventory members", "target_missing", 2, func(_ map[string][]byte, inventory Inventory, _ *TrustSnapshot, _ *time.Time) {
			inventory["secret-a"] = Member{Path: "secret-a", Size: 1, SHA256: strings.Repeat("0", 64)}
			inventory["secret-b"] = Member{Path: "secret-b", Size: 1, SHA256: strings.Repeat("0", 64)}
		}},
		{"all missing target members", "target_extra", len(vector.Inventory), func(_ map[string][]byte, inventory Inventory, _ *TrustSnapshot, _ *time.Time) {
			clear(inventory)
		}},
		{"all target lengths before hashes", "target_length", len(vector.Inventory), func(_ map[string][]byte, inventory Inventory, _ *TrustSnapshot, _ *time.Time) {
			for path, member := range inventory {
				member.Size++
				member.SHA256 = strings.Repeat("0", 64)
				inventory[path] = member
			}
		}},
		{"all target hashes", "target_hash", len(vector.Inventory), func(_ map[string][]byte, inventory Inventory, _ *TrustSnapshot, _ *time.Time) {
			for path, member := range inventory {
				member.SHA256 = strings.Repeat("0", 64)
				inventory[path] = member
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			trust, err := AdmitBootstrap([]byte(vector.Bootstrap))
			if err != nil {
				t.Fatal(err)
			}
			files := vectorMetadata(vector)
			inventory := Inventory{}
			for path, member := range vector.Inventory {
				inventory[path] = member
			}
			at := vector.VerificationTime
			tc.mutate(files, inventory, &trust, &at)
			_, err = VerifyTrustWithDiagnostics(context.Background(), files, inventory, "test.repo", at, trust, nil)
			var failure *Failure
			if !errors.As(err, &failure) || failure.CheckID != tc.check || failure.Summary == nil || failure.Summary.Total != tc.count {
				t.Fatalf("check=%s count=%d: %#v / %v", tc.check, tc.count, failure, err)
			}
			first, err := canonicaljson.Marshal(failure.Summary)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeValidationSummary(first); err != nil {
				t.Fatal("invalid generated diagnostic", err)
			}
			if bytes.Contains(first, []byte("secret-")) {
				t.Fatal("hostile target leaked")
			}
			reversed := map[string][]byte{}
			paths := sortedTrustKeys(files)
			slices.Reverse(paths)
			for _, path := range paths {
				reversed[path] = files[path]
			}
			_, err = VerifyTrustWithDiagnostics(context.Background(), reversed, inventory, "test.repo", at, trust, nil)
			if !errors.As(err, &failure) {
				t.Fatal(err)
			}
			again, _ := canonicaljson.Marshal(failure.Summary)
			if !bytes.Equal(first, again) {
				t.Fatal("map traversal changed complete summary")
			}
			_, err = VerifyTrust(files, inventory, "test.repo", at, trust)
			if !errors.As(err, &failure) || failure.CheckID != tc.check {
				t.Fatalf("reader drift: %v", err)
			}
		})
	}
}
