package contractadmission

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestRetainedArtifactsClosedAdmission_Unit(t *testing.T) {
	read := func(path string) []byte {
		t.Helper()
		b, err := os.ReadFile("../../../../" + path)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	encode := func(v any) []byte {
		t.Helper()
		b, err := canonicaljson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	var release struct {
		Packs []struct {
			Members map[string]string `json:"members"`
		} `json:"packs"`
	}
	if err := json.Unmarshal(read("contracts/reference-packs/builtins/release.v3.json"), &release); err != nil {
		t.Fatal(err)
	}
	vector := struct{ Members []packformat.SetMember }{currentBaseMembers(t)}
	set, err := packformat.BuildSet(vector.Members)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("stored set binding", func(t *testing.T) {
		raw := encode(set)
		if _, err := packformat.DecodeBoundSet(raw, set.ID, set.SHA256); err != nil {
			t.Fatal("valid bound set", err)
		}
		for _, binding := range [][2]string{{"rpset_" + strings.Repeat("a", 64), set.SHA256}, {set.ID, strings.Repeat("a", 64)}} {
			if _, err := packformat.DecodeBoundSet(raw, binding[0], binding[1]); err == nil {
				t.Fatal("valid set admitted under another identity or digest")
			}
		}
	})
	for _, tc := range []struct {
		kind     string
		raw      []byte
		required string
	}{
		{"manifest:packaged_builtin", []byte(release.Packs[0].Members["manifest.json"]), "source_identifier"},
		{"set", encode(set), "members"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			if err := validate(tc.kind, tc.raw); err != nil {
				t.Fatal("valid current artifact", err)
			}
			for _, mode := range []string{"unknown", "missing", "null", "type", "duplicate JSON", "trailing", "noncanonical", "schema"} {
				t.Run(mode, func(t *testing.T) {
					var obj map[string]any
					if err := json.Unmarshal(tc.raw, &obj); err != nil {
						t.Fatal(err)
					}
					switch mode {
					case "unknown":
						obj["extra"] = true
					case "missing":
						delete(obj, tc.required)
					case "null":
						obj[tc.required] = nil
					case "type":
						obj[tc.required] = 42
					case "schema":
						obj["schema_id"] = "unsupported"
					}
					raw := encode(obj)
					switch mode {
					case "duplicate JSON":
						raw = append([]byte(`{"schema_id":"duplicate",`), raw[1:]...)
					case "trailing":
						raw = append(raw, []byte(`{}`)...)
					case "noncanonical":
						raw = append(raw, '\n')
					}
					if !IsIncompatible(validate(tc.kind, raw)) {
						t.Fatal("malformed artifact admitted", mode)
					}
				})
			}
		})
	}
	if !IsIncompatible(validate("manifest:operator_imported", []byte(release.Packs[0].Members["manifest.json"]))) {
		t.Fatal("stored distribution kind ignored")
	}
	for _, mode := range []string{"order", "duplicate", "required registry", "identity", "nested unknown", "nested missing", "nested null"} {
		t.Run(mode, func(t *testing.T) {
			var obj map[string]any
			if err := json.Unmarshal(encode(set), &obj); err != nil {
				t.Fatal(err)
			}
			members := obj["members"].([]any)
			switch mode {
			case "order":
				members[0], members[1] = members[1], members[0]
			case "duplicate":
				members[1] = members[0]
			case "required registry":
				obj["members"] = members[:len(members)-1]
			case "identity":
				obj["pack_set_sha256"] = strings.Repeat("f", 64)
			case "nested unknown":
				members[0].(map[string]any)["extra"] = true
			case "nested missing":
				delete(members[0].(map[string]any), "pack_version")
			case "nested null":
				members[0].(map[string]any)["pack_version"] = nil
			}
			if !IsIncompatible(validate("set", encode(obj))) {
				t.Fatal("malformed set admitted", mode)
			}
		})
	}
}

func currentBaseMembers(t testing.TB) []packformat.SetMember {
	t.Helper()
	raw, err := os.ReadFile("../../../../contracts/reference-packs/builtins/release.v3.json")
	if err != nil {
		t.Fatal(err)
	}
	var release struct {
		Packs []struct {
			Members map[string]string `json:"members"`
		} `json:"packs"`
	}
	if err := json.Unmarshal(raw, &release); err != nil {
		t.Fatal(err)
	}
	members := []packformat.SetMember{}
	for _, p := range release.Packs {
		raw := []byte(p.Members["manifest.json"])
		m, err := packformat.DecodeManifest(raw, false)
		if err != nil {
			t.Fatal(err)
		}
		payload, err := packformat.PayloadDigest(m.Files)
		if err != nil {
			t.Fatal(err)
		}
		members = append(members, packformat.SetMember{Key: m.Key, Version: m.Version, ManifestSHA256: packformat.Digest(raw), PayloadSHA256: payload, Contract: m.Contract, ProfileID: m.ProfileID, ProfileVersion: m.ProfileVersion})
	}
	return members
}
