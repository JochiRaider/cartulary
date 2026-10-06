package recoveryassembly

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestOperatorReleaseIdentity_Unit(t *testing.T) {
	root := t.TempDir()
	executable := filepath.Join(root, "operator")
	if err := os.WriteFile(executable, []byte("exact executable"), 0700); err != nil {
		t.Fatal(err)
	}
	binaryHash := sha256.Sum256([]byte("exact executable"))
	digest := hex.EncodeToString(binaryHash[:])
	manifestPath := filepath.Join(root, "release-manifest.json")
	if got, err := operatorReleaseIdentity(executable, manifestPath); err != nil || got != digest {
		t.Fatal("unpackaged executable identity", got, err)
	}
	manifest := map[string]any{"schema_id": "cartulary.local_release_manifest.v1", "platform": "linux/amd64", "assets": []any{}, "images": []any{}, "binaries": map[string]any{"operator": map[string]string{"sha256": digest, "receipt": "receipts/operator.crypto.json"}, "server": map[string]string{}, "migrate": map[string]string{}}}
	body, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, body, 0600); err != nil {
		t.Fatal(err)
	}
	manifestHash := sha256.Sum256(body)
	if got, err := operatorReleaseIdentity(executable, manifestPath); err != nil || got != hex.EncodeToString(manifestHash[:]) {
		t.Fatal("whole release identity", got, err)
	}
	if err := os.WriteFile(executable, []byte("substitution"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := operatorReleaseIdentity(executable, manifestPath); err == nil {
		t.Fatal("substituted executable admitted")
	}
	if err := os.WriteFile(executable, []byte("exact executable"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(manifestPath); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(root, "other")
	if err := os.WriteFile(other, body, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, manifestPath); err != nil {
		t.Fatal(err)
	}
	if _, err := operatorReleaseIdentity(executable, manifestPath); err == nil {
		t.Fatal("linked manifest admitted")
	}
}
