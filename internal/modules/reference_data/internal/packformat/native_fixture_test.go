package packformat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type nativeFixtureSource struct {
	root    string
	members map[string]string
}

func (s nativeFixtureSource) Open(_ context.Context, name string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(s.root, s.members[name]))
}

func runNativeFixture(t *testing.T, root string, input []byte, expected map[string]any) {
	t.Helper()
	var f struct {
		SchemaID string            `json:"schema_id"`
		Key      string            `json:"pack_key"`
		Manifest string            `json:"manifest"`
		Members  map[string]string `json:"members"`
	}
	decoder := json.NewDecoder(bytes.NewReader(input))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&f); err != nil || f.SchemaID != "cartulary.attack_native_fixture.v1" || f.Key != "framework.attack" {
		t.Fatal("native fixture descriptor", err)
	}
	b, err := os.ReadFile(filepath.Join(root, f.Manifest))
	if err != nil {
		t.Fatal(err)
	}
	if Digest(b) != expected["expected_manifest_sha256"] {
		t.Fatal("native fixture manifest identity")
	}
	m, err := DecodeManifest(b, true)
	if err != nil {
		t.Fatal(err)
	}
	source := nativeFixtureSource{root, f.Members}
	verdict := ValidateContent(context.Background(), m, source, nil)
	public, rejected := expected["expected_public_error"].(map[string]any)
	if rejected {
		var failure *Failure
		if !errors.As(verdict, &failure) || failure.Code != public["reason_code"] {
			t.Fatal("native derivative verdict", verdict)
		}
		// Earlier malformed recipes intentionally alter bytes without reissuing
		// their inventory. Endpoint derivatives bind their corrected inventory.
		if expected["expected_payload_sha256"] == nil {
			return
		}
	} else if verdict != nil {
		t.Fatal(verdict)
	}
	inventory := Inventory{}
	for _, file := range m.Files {
		b, err := os.ReadFile(filepath.Join(root, f.Members[file.Path]))
		if err != nil {
			t.Fatal(err)
		}
		inventory[file.Path] = Member{Path: file.Path, Size: int64(len(b)), SHA256: Digest(b)}
	}
	// The manifest is outside the framed payload but inside the admitted inventory.
	inventory["manifest.json"] = Member{Path: "manifest.json", Size: int64(len(b)), SHA256: Digest(b)}
	if err := ValidateInventory(m, inventory, false); err != nil {
		t.Fatal(err)
	}
	payload, err := PayloadDigest(m.Files)
	if err != nil || payload != expected["expected_payload_sha256"] {
		t.Fatal("native fixture payload", err)
	}
}
