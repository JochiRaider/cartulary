package packformat

import (
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestBuiltinBindingClosesIdentityAndShape_Unit(t *testing.T) {
	var release struct {
		Packs []struct {
			Binding map[string]any    `json:"binding"`
			Members map[string]string `json:"members"`
		} `json:"packs"`
	}
	if err := json.Unmarshal(projection("builtins/release.v1.json"), &release); err != nil {
		t.Fatal(err)
	}
	for _, pack := range release.Packs {
		data, err := canonicaljson.Marshal(pack.Binding)
		if err != nil {
			t.Fatal(err)
		}
		binding, err := DecodeBuiltinReleaseBinding(data)
		if err != nil {
			t.Fatal(err)
		}
		manifest, err := DecodeManifest([]byte(pack.Members["manifest.json"]), false)
		if err != nil {
			t.Fatal(err)
		}
		payload, err := PayloadDigest(manifest.Files)
		if err != nil || !binding.Matches(manifest, Digest([]byte(pack.Members["manifest.json"])), payload) {
			t.Fatal("release identity mismatch", err)
		}
		for name := range pack.Binding {
			for _, value := range []any{nil, true} {
				var changed map[string]any
				if err := json.Unmarshal(data, &changed); err != nil {
					t.Fatal(err)
				}
				changed[name] = value
				encoded, err := canonicaljson.Marshal(changed)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := DecodeBuiltinReleaseBinding(encoded); err == nil {
					t.Fatal("invalid binding member admitted", name)
				}
				delete(changed, name)
				encoded, err = canonicaljson.Marshal(changed)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := DecodeBuiltinReleaseBinding(encoded); err == nil {
					t.Fatal("omitted binding member admitted", name)
				}
			}
		}
		pack.Binding["unknown"] = "hostile"
		encoded, err := canonicaljson.Marshal(pack.Binding)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeBuiltinReleaseBinding(encoded); err == nil {
			t.Fatal("unknown binding member admitted")
		}
		changed := manifest
		changed.Sequence++
		if binding.Matches(changed, binding.ManifestSHA256, binding.PayloadSHA256) {
			t.Fatal("sequence binding ignored")
		}
		changed = manifest
		changed.ProfileID = "cartulary.reference_pack.enrichment.tor.v1"
		if binding.Matches(changed, binding.ManifestSHA256, binding.PayloadSHA256) {
			t.Fatal("profile binding ignored")
		}
	}
}
