package packformat

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// Fresh deterministic test-only signatures keep this boundary test independent
// of the verifier's signing/threshold logic. Separate checked-in vectors own
// cross-language cryptographic expected bytes.
func testCompleteTUFByteBoundaries(t *testing.T) {
	vector := trustVectors(t)[0]

	canonical := func(value any) []byte {
		data, err := canonicaljson.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	keys := map[string]ed25519.PrivateKey{}
	publicKeys := map[string]any{}
	keyIDs := []string{}
	for i := 0; i < 64; i++ {
		key := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{byte(0x37 + i)}, ed25519.SeedSize))
		public := map[string]any{"keytype": "ed25519", "scheme": "ed25519", "keyval": map[string]any{"public": hex.EncodeToString(key.Public().(ed25519.PublicKey))}}
		id := Digest(canonical(public))
		keys[id], publicKeys[id] = key, public
		keyIDs = append(keyIDs, id)
	}
	slices.Sort(keyIDs)
	sign := func(signed map[string]any) []byte {
		ids := keyIDs
		payload := canonical(signed)
		signatures := []any{}
		for _, id := range ids {
			signatures = append(signatures, map[string]any{"keyid": id, "sig": hex.EncodeToString(ed25519.Sign(keys[id], payload))})
		}
		return canonical(map[string]any{"signed": signed, "signatures": signatures})
	}
	signedObject := func(data []byte) map[string]any {
		value, err := canonicaljson.DecodeStrict(data)
		if err != nil {
			t.Fatal(err)
		}
		return value.(map[string]any)["signed"].(map[string]any)
	}
	root := signedObject([]byte(vector.Bootstrap))
	root["keys"] = publicKeys
	roles := map[string]any{}
	for _, role := range []string{"root", "targets", "snapshot", "timestamp"} {
		roles[role] = map[string]any{"keyids": keyIDs, "threshold": 1}
	}
	roles["root"] = map[string]any{"keyids": keyIDs, "threshold": 2}
	root["roles"] = roles
	bootstrap := sign(root)
	trust, err := AdmitBootstrap(bootstrap)
	if err != nil {
		t.Fatal(err)
	}
	// The sixty-fifth independent key fails shape admission even when its
	// key ID and public bytes agree. It must not become a trust-policy failure.
	extraRoot := signedObject(bootstrap)
	extraKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x90}, ed25519.SeedSize))
	extraPublic := map[string]any{"keytype": "ed25519", "scheme": "ed25519", "keyval": map[string]any{"public": hex.EncodeToString(extraKey.Public().(ed25519.PublicKey))}}
	extraRoot["keys"].(map[string]any)[Digest(canonical(extraPublic))] = extraPublic
	if _, err := decodeMetadataShape(sign(extraRoot), "root", 128); err == nil {
		t.Fatal("sixty-fifth root key admitted")
	}
	for _, boundary := range []struct {
		raw     []byte
		kind    string
		maximum int
	}{
		{sign(signedObject([]byte(vector.Metadata["metadata/targets.json"]))), "targets", 64},
		{[]byte(trustVectors(t)[2].Metadata["metadata/2.root.json"]), "root", 128},
	} {
		value, err := canonicaljson.DecodeStrict(boundary.raw)
		if err != nil {
			t.Fatal(err)
		}
		envelope := value.(map[string]any)
		signatures := envelope["signatures"].([]any)
		if len(signatures) != boundary.maximum {
			t.Fatal("signature boundary recipe")
		}
		envelope["signatures"] = append(signatures, signatures[0])
		if _, err := decodeMetadataShape(canonical(envelope), boundary.kind, boundary.maximum); err == nil {
			t.Fatal("one-over signature count admitted", boundary.kind)
		}
	}
	const metadataLimit = 2097152
	pad := func(signed map[string]any) []byte {
		signed["org.cartulary.test_boundary"] = ""
		length := len(sign(signed))
		if length >= metadataLimit {
			t.Fatal("invalid metadata recipe")
		}
		signed["org.cartulary.test_boundary"] = strings.Repeat("a", metadataLimit-length)
		result := sign(signed)
		if len(result) != metadataLimit {
			t.Fatal("metadata recipe does not reach equality")
		}
		return result
	}
	// A canonical four-repository deployment file reaches the independent
	// eight MiB bootstrap limit while each embedded root stays within two MiB.
	repositories := []any{}
	for i := 0; i < 4; i++ {
		item := signedObject(bootstrap)
		id := fmt.Sprintf("test.repo%02d", i)
		item["cartulary"].(map[string]any)["trust_repository_id"] = id
		raw := pad(item)
		value, err := canonicaljson.DecodeStrict(raw)
		if err != nil {
			t.Fatal(err)
		}
		repositories = append(repositories, map[string]any{"repository_id": id, "trusted_root": value, "trusted_root_sha256": Digest(raw)})
	}
	deployment := map[string]any{"schema_id": "cartulary.reference_pack_trust_bootstrap.v1", "repositories": repositories}
	last := repositories[3].(map[string]any)
	lastSigned := last["trusted_root"].(map[string]any)["signed"].(map[string]any)
	excess := len(canonical(deployment)) - 8388608
	padding := lastSigned["org.cartulary.test_boundary"].(string)
	lastSigned["org.cartulary.test_boundary"] = padding[:len(padding)-excess]
	for _, over := range []bool{false, true} {
		if over {
			lastSigned["org.cartulary.test_boundary"] = lastSigned["org.cartulary.test_boundary"].(string) + "a"
		}
		raw := sign(lastSigned)
		value, err := canonicaljson.DecodeStrict(raw)
		if err != nil {
			t.Fatal(err)
		}
		last["trusted_root"], last["trusted_root_sha256"] = value, Digest(raw)
		data := canonical(deployment)
		expected := 8388608
		if over {
			expected++
		}
		if len(data) != expected {
			t.Fatal("bootstrap byte recipe", len(data))
		}
		roots, err := DecodeBootstrap(data)
		if (!over && (err != nil || len(roots) != 4)) || (over && err == nil) {
			t.Fatal("bootstrap byte boundary", over, err)
		}
	}
	root["version"] = 2
	files := map[string][]byte{"metadata/2.root.json": pad(root)}
	files["metadata/targets.json"] = pad(signedObject([]byte(vector.Metadata["metadata/targets.json"])))
	descriptor := func(data []byte) map[string]any {
		return map[string]any{"version": 1, "length": len(data), "hashes": map[string]any{"sha256": Digest(data)}}
	}
	snapshot := signedObject([]byte(vector.Metadata["metadata/snapshot.json"]))
	snapshot["meta"] = map[string]any{"targets.json": descriptor(files["metadata/targets.json"])}
	files["metadata/snapshot.json"] = pad(snapshot)
	timestamp := signedObject([]byte(vector.Metadata["metadata/timestamp.json"]))
	timestamp["meta"] = map[string]any{"snapshot.json": descriptor(files["metadata/snapshot.json"])}
	files["metadata/timestamp.json"] = pad(timestamp)
	proposal, err := VerifyTrust(files, vector.Inventory, "test.repo", vector.VerificationTime, trust)
	if err != nil || len(proposal.RootTransitions) != 1 || len(proposal.Signers["targets"]) != 64 {
		t.Fatal("complete eight MiB trust envelope", err)
	}
	// Depth counts open containers, including the metadata envelope and signed
	// object. This complete authenticated extension reaches 10,000, independently
	// of byte and signature bounds. Signed extension values remain authenticated.
	var limits struct {
		Fixed map[string]int `json:"fixed"`
	}
	if err := json.Unmarshal(projection("limits.v1.json"), &limits); err != nil || limits.Fixed["max_json_nesting_depth"] != 10000 {
		t.Fatal("JSON nesting projection drift", err)
	}
	deep := strings.Repeat("[", 9998) + "0" + strings.Repeat("]", 9998)
	deepTimestamp := signedObject(files["metadata/timestamp.json"])
	delete(deepTimestamp, "org.cartulary.test_boundary")
	deepTimestamp["org.cartulary.nesting"] = json.RawMessage(deep)
	deepFiles := maps.Clone(files)
	deepFiles["metadata/timestamp.json"] = sign(deepTimestamp)
	if _, err := VerifyTrust(deepFiles, vector.Inventory, "test.repo", vector.VerificationTime, trust); err != nil {
		t.Fatal("complete authenticated JSON nesting boundary", err)
	}
	deepFiles["metadata/timestamp.json"] = bytes.Replace(deepFiles["metadata/timestamp.json"], []byte(deep), []byte("["+deep+"]"), 1)
	var failure *Failure
	if _, err := VerifyTrust(deepFiles, vector.Inventory, "test.repo", vector.VerificationTime, trust); !errors.As(err, &failure) || failure.Code != "tuf_metadata_invalid" {
		t.Fatal("JSON depth did not reject at original admission", err)
	}
	// All per-file bounds remain valid; an exact historical-root replay exceeds
	// the aggregate only. Nothing in the supplied trust snapshot is mutated.
	files["metadata/1.root.json"] = bootstrap
	shorter := signedObject(files["metadata/timestamp.json"])
	padding = shorter["org.cartulary.test_boundary"].(string)
	shorter["org.cartulary.test_boundary"] = padding[:len(padding)-len(bootstrap)+1]
	files["metadata/timestamp.json"] = sign(shorter)
	total := 0
	for _, data := range files {
		total += len(data)
	}
	if total != 8388609 {
		t.Fatal("aggregate one-over recipe", total)
	}
	summary, err := TrustSchemaSummary(context.Background(), files, nil)
	if err != nil || summary.Result != "failed" || summary.Issues[0].Details.LimitID == nil || *summary.Issues[0].Details.LimitID != "max_metadata_total_bytes" {
		t.Fatal("aggregate metadata bound", summary, err)
	}
	delete(files, "metadata/1.root.json")
	// One byte over a per-file bound is rejected before canonical/signature work.
	raw := append(bytes.Clone(files["metadata/targets.json"]), ' ')
	var findings []Finding
	_, err = metadataSchemaFindings(raw, "targets", 64, "$.metadata[0]", func(f Finding) error { findings = append(findings, f); return nil })
	if err == nil || len(findings) != 1 || findings[0].Details.LimitID == nil || *findings[0].Details.LimitID != "max_metadata_file_bytes" {
		t.Fatal("individual metadata bound", findings, err)
	}
}
