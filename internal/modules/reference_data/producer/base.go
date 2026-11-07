package producer

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type BaseReissueInput struct {
	SchemaID        string `json:"schema_id"`
	PreviousRelease string `json:"previous_release"`
	OutputFile      string `json:"output_file"`
	BuiltAt         string `json:"built_at"`
	BuilderSHA256   string `json:"builder_sha256"`
}

// ReissueBase is build-time production from the preserved project-owned source
// registries. It is never linked into a runtime compatibility/read path.
func ReissueBase(ctx context.Context, b []byte, builder Builder) (map[string]any, error) {
	var in BaseReissueInput
	if err := DecodeClosed(b, &in); err != nil {
		return nil, err
	}
	if in.SchemaID != "cartulary.base_registry_reissue_input.v1" || in.BuilderSHA256 != builder.SHA256 || !validBuildTime(in.BuiltAt) {
		return nil, errors.New("invalid Base reissue input")
	}
	previous, err := readRegular(in.PreviousRelease)
	if err != nil {
		return nil, err
	}
	v, err := canonicaljson.DecodeStrict(previous)
	if err != nil {
		return nil, err
	}
	release, ok := v.(map[string]any)
	if !ok || release["schema_id"] != "cartulary.reference_pack_builtin_release.v1" {
		return nil, errors.New("invalid source Base release")
	}
	packs, ok := release["packs"].([]any)
	if !ok || len(packs) != 3 {
		return nil, errors.New("three Base registries required")
	}
	setMembers := []packformat.SetMember{}
	for _, value := range packs {
		p := value.(map[string]any)
		members := p["members"].(map[string]any)
		var m packformat.Manifest
		if err = json.Unmarshal([]byte(members["manifest.json"].(string)), &m); err != nil {
			return nil, err
		}
		m.Contract = packformat.ContractID
		m.Version = "base-2026-11-06.1"
		m.Sequence = 2
		m.BuiltAt = in.BuiltAt
		m.Builder.ID = builder.ID
		m.Builder.Version = builder.Version
		m.Builder.SHA256 = builder.SHA256
		manifest, err := canonicaljson.Marshal(m)
		if err != nil {
			return nil, err
		}
		if _, err = packformat.DecodeManifest(manifest, false); err != nil {
			return nil, err
		}
		content := packageSource{}
		for name, value := range members {
			content[name] = []byte(value.(string))
		}
		if err = packformat.ValidateContent(ctx, m, content, nil); err != nil {
			return nil, err
		}
		for _, f := range m.Files {
			if int64(len(content[f.Path])) != f.Size || Digest(content[f.Path]) != f.SHA256 {
				return nil, errors.New("Base registry bytes changed")
			}
		}
		payload, err := packformat.PayloadDigest(m.Files)
		if err != nil {
			return nil, err
		}
		members["manifest.json"] = string(manifest)
		binding := p["binding"].(map[string]any)
		binding["manifest_sha256"] = Digest(manifest)
		binding["pack_version"] = m.Version
		binding["pack_release_sequence"] = m.Sequence
		setMembers = append(setMembers, packformat.SetMember{Key: m.Key, Version: m.Version, ManifestSHA256: Digest(manifest), PayloadSHA256: payload, Contract: m.Contract, ProfileID: m.ProfileID, ProfileVersion: m.ProfileVersion})
	}
	release["application_release_id"] = "cartulary.base_registries.2026-11-06.v2"
	encoded, err := canonicaljson.Marshal(release)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(in.OutputFile), 0755); err != nil {
		return nil, err
	}
	out, err := os.OpenRoot(filepath.Dir(in.OutputFile))
	if err != nil {
		return nil, err
	}
	defer out.Close()
	if err = Retain(out, filepath.Base(in.OutputFile), encoded, 0644); err != nil {
		return nil, err
	}
	set, err := packformat.BuildSet(setMembers)
	if err != nil {
		return nil, err
	}
	return map[string]any{"schema_id": "cartulary.base_registry_reissue_report.v1", "release_sha256": Digest(encoded), "previous_release_sha256": Digest(previous), "builder": builder, "pack_set": set}, nil
}
