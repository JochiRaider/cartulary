package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"sort"
	"strings"

	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contractextensions"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const extensionGeneratedPrefix = "contracts/extensions/generated/"

// This is the final packaging step. Its root describes the exact static-byte
// manifest, while the contracts themselves stay outside that static mapping.
func writeExtensionIntegrity(repositoryRoot, manifestPath, supportPath, bootstrapPath, output, pinOutput string) error {
	base, err := canonicaljson.DecodeExtension([]byte(contractsgen.Index[extensionGeneratedPrefix+"registry-integrity.json"].JSON))
	if err != nil {
		return err
	}
	root, ok := base.(map[string]any)
	if !ok {
		return errors.New("generated integrity is not an object")
	}
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	support, err := os.ReadFile(supportPath)
	if err != nil {
		return err
	}
	bootstrap, err := os.ReadFile(bootstrapPath)
	if err != nil {
		return err
	}
	current, err := canonicaljson.DecodeExtension(support)
	if err != nil {
		return err
	}
	source, err := canonicaljson.DecodeExtension([]byte(contractsgen.Index[extensionGeneratedPrefix+"client-support-registry.json"].JSON))
	if err != nil {
		return err
	}
	currentObject, currentOK := current.(map[string]any)
	sourceObject, sourceOK := source.(map[string]any)
	if !currentOK || !sourceOK {
		return errors.New("client support is not an object")
	}
	currentProfiles, err := canonicaljson.MarshalExtension(currentObject["profiles"])
	if err != nil {
		return err
	}
	sourceProfiles, err := canonicaljson.MarshalExtension(sourceObject["profiles"])
	if err != nil {
		return err
	}
	if string(currentProfiles) != string(sourceProfiles) || currentObject["client_build_class"] != sourceObject["client_build_class"] {
		return errors.New("browser support disagrees with the generated profile catalog")
	}
	contracts, ok := root["supporting_contract_artifact_digests"].([]any)
	if !ok {
		return errors.New("generated supporting contracts are missing")
	}
	rows := make([]any, 0, len(contracts)+1)
	found := false
	for _, raw := range contracts {
		row, ok := raw.(map[string]any)
		if !ok {
			return errors.New("generated contract digest row is invalid")
		}
		switch row["artifact_id"] {
		case "client-support-registry":
			row["artifact_sha256"] = byteDigest(support)
			found = true
		case "client-asset-set-manifest", "client-bootstrap-template":
			return errors.New("base registry already contains a final browser manifest")
		}
		rows = append(rows, row)
	}
	if !found {
		return errors.New("generated client support digest is missing")
	}
	rows = append(rows, map[string]any{"artifact_id": "client-asset-set-manifest", "schema_id": "cartulary.client_asset_set_manifest.v1", "artifact_sha256": byteDigest(manifest)})
	rows = append(rows, map[string]any{"artifact_id": "client-bootstrap-template", "schema_id": "cartulary.client_bootstrap_template.v1", "artifact_sha256": byteDigest(bootstrap)})
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].(map[string]any)["artifact_id"].(string) < rows[j].(map[string]any)["artifact_id"].(string)
	})
	root["supporting_contract_artifact_digests"] = rows
	root["generator_id"] = "cartulary.embedwebassets.extensions.v1"
	sources, ok := root["generator_sources"].([]any)
	if !ok {
		return errors.New("generated source provenance is missing")
	}
	capability, err := os.OpenRoot(repositoryRoot)
	if err != nil {
		return err
	}
	defer capability.Close()
	for _, ref := range []string{"tools/embedwebassets/main.go", "tools/embedwebassets/registry.go", "tools/harness/readiness/embed-web-assets.sh"} {
		parts := strings.Split(ref, "/")
		for i := range parts {
			info, err := capability.Lstat(strings.Join(parts[:i+1], "/"))
			if err != nil {
				return err
			}
			if info.Mode()&os.ModeSymlink != 0 || i < len(parts)-1 && !info.IsDir() || i == len(parts)-1 && !info.Mode().IsRegular() {
				return errors.New("packaging source must be a regular confined file")
			}
		}
		body, err := capability.ReadFile(ref)
		if err != nil {
			return err
		}
		sources = append(sources, map[string]any{"source_ref": ref, "source_sha256": byteDigest(body)})
	}
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].(map[string]any)["source_ref"].(string) < sources[j].(map[string]any)["source_ref"].(string)
	})
	for i := 1; i < len(sources); i++ {
		if sources[i].(map[string]any)["source_ref"] == sources[i-1].(map[string]any)["source_ref"] {
			return errors.New("duplicate packaging provenance")
		}
	}
	root["generator_sources"] = sources
	body, err := canonicaljson.MarshalExtension(root)
	if err != nil {
		return err
	}
	if len(body) > 8<<20 {
		return errors.New("final registry integrity exceeds byte limit")
	}
	if err := writeCanonicalOutput(output, body); err != nil {
		return err
	}
	return writeCanonicalOutput(pinOutput, []byte(byteDigest(body)+"\n"))
}

func byteDigest(body []byte) string { sum := sha256.Sum256(body); return hex.EncodeToString(sum[:]) }
