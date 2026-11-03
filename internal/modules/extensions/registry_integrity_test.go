package extensions

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/browserbuild"
)

type packageOverride struct {
	ArtifactSource
	paths []string
	root  string
}

func (s packageOverride) ArtifactPaths() []string {
	if s.paths != nil {
		return s.paths
	}
	return s.ArtifactSource.ArtifactPaths()
}
func (s packageOverride) IntegritySHA256() string {
	if s.root != "" {
		return s.root
	}
	return s.ArtifactSource.IntegritySHA256()
}

func checkRegistryPackageClosure(t *testing.T) {
	t.Run("independent build root", func(t *testing.T) {
		_, err := NewCoordinator(packageOverride{ArtifactSource: newMutableGeneratedSource(), root: strings.Repeat("0", 64)})
		requireFinding(t, err, "extension_registry_invalid", "registry_generation")
	})
	t.Run("no source checkout", func(t *testing.T) {
		t.Chdir(t.TempDir())
		if _, err := NewGeneratedCoordinator(); err != nil {
			t.Fatal(err)
		}
	})
	for _, name := range []string{"omitted", "duplicate", "extra"} {
		t.Run("inventory "+name, func(t *testing.T) {
			source := newMutableGeneratedSource()
			paths := source.ArtifactPaths()
			switch name {
			case "omitted":
				paths = paths[1:]
			case "duplicate":
				paths = append(paths, paths[0])
				sort.Strings(paths)
			case "extra":
				path := "contracts/extensions/build/unbound.json"
				source.replace(t, path, map[string]any{"schema_id": "cartulary.unbound.v1"})
				paths = source.ArtifactPaths()
			}
			_, err := NewCoordinator(packageOverride{ArtifactSource: source, paths: paths})
			requireFinding(t, err, "extension_registry_invalid", "registry_generation")
		})
	}
	for _, count := range []int{1024, 1025} {
		t.Run(fmt.Sprintf("source count %d", count), func(t *testing.T) {
			source := newMutableGeneratedSource()
			path := generatedExtensionsRoot + "registry-integrity.json"
			root := decodeMutableObject(t, source[path].JSON)
			rows := make([]any, count)
			for i := range rows {
				rows[i] = map[string]any{"source_ref": fmt.Sprintf("producer/source%04d.go", i), "source_sha256": strings.Repeat("0", 64)}
			}
			root["generator_sources"] = rows
			source.replace(t, path, root)
			_, err := NewCoordinator(source)
			if count == 1024 {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				requireFinding(t, err, "extension_registry_limit_exceeded", "registry_generation")
			}
		})
	}
	t.Run("artifact bytes before decoding", func(t *testing.T) {
		path := generatedExtensionsRoot + "descriptors/limit.json"
		for _, extra := range []int{0, 1} {
			raw := []byte(`{"x":"` + strings.Repeat("a", (256<<10)-9+extra) + `"}` + "\n")
			if len(raw) != (256<<10)+extra {
				t.Fatalf("fixture bytes %d", len(raw))
			}
			sum := sha256.Sum256(raw)
			source := mutableArtifactSource{path: {JSON: raw, SHA256: hex.EncodeToString(sum[:])}}
			_, _, err := readArtifactObject(source, path)
			if extra == 0 {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				requireFinding(t, err, "extension_registry_limit_exceeded", "registry_generation")
			}
		}
	})
	for _, name := range []string{"missing state", "missing backup", "duplicate backup", "missing participant", "state implementation digest", "state final algorithm", "embedded physical binding", "embedded participant"} {
		t.Run(name, func(t *testing.T) {
			source := newMutableGeneratedSource()
			slot := "state-registry"
			if strings.Contains(name, "backup") || name == "embedded physical binding" {
				slot = "backup-registry"
			}
			if strings.Contains(name, "participant") {
				slot = "participant-registry"
			}
			path := generatedExtensionsRoot + slot + ".json"
			object := decodeMutableObject(t, source[path].JSON)
			field := "profiles"
			if slot == "participant-registry" {
				field = "participants"
			}
			rows := object[field].([]any)
			switch name {
			case "missing state", "missing backup", "missing participant":
				object[field] = rows[1:]
			case "duplicate backup":
				object[field] = append(rows, rows[0])
			case "state implementation digest":
				rows[0].(map[string]any)["implementation_binding_sha256"] = strings.Repeat("0", 64)
			case "state final algorithm":
				rows[0].(map[string]any)["final_state_validation_algorithm_id"] = "network_flow_activity.unbound_v1"
			case "embedded physical binding":
				rows[0].(map[string]any)["physical_state_binding"].(map[string]any)["state_presence_manifest_sha256"] = strings.Repeat("0", 64)
			case "embedded participant":
				row := rows[0].(map[string]any)
				body := row["contract"].(map[string]any)
				body["maximum_input_bytes"] = 1
				row["participant_contract_sha256"] = canonicalDigest(body)
			}
			source.replace(t, path, object)
			rebindSupportingArtifact(t, source, slot, path)
			_, err := NewCoordinator(source)
			requireFinding(t, err, "extension_registry_invalid", "registry_generation")
		})
	}
}

func rebindSupportingArtifact(t *testing.T, source mutableArtifactSource, id, path string) {
	t.Helper()
	rootPath := generatedExtensionsRoot + "registry-integrity.json"
	root := decodeMutableObject(t, source[rootPath].JSON)
	found := false
	for _, raw := range root["supporting_contract_artifact_digests"].([]any) {
		row := raw.(map[string]any)
		if row["artifact_id"] == id {
			row["artifact_sha256"] = source[path].SHA256
			found = true
		}
	}
	if !found {
		t.Fatal("support artifact missing from fixture")
	}
	source.replace(t, rootPath, root)
}

func checkFinalBrowserPackage(t *testing.T) {
	for _, name := range []string{"valid", "missing pin", "wrong pin", "missing template", "stale support", "coherent unsupported profile", "missing manifest digest", "unknown root member", "source placeholder"} {
		t.Run(name, func(t *testing.T) {
			c := browserbuild.New(t).Contracts()
			pack := BrowserPackage{IntegrityJSON: c.IntegrityJSON, IntegritySHA256: c.IntegritySHA256, ClientSupportJSON: c.ClientSupportJSON, AssetManifestJSON: c.AssetManifestJSON, BootstrapTemplateJSON: c.BootstrapTemplateJSON}
			root := decodeMutableObject(t, pack.IntegrityJSON)
			switch name {
			case "missing pin":
				pack.IntegritySHA256 = ""
			case "wrong pin":
				pack.IntegritySHA256 = strings.Repeat("a", 64)
			case "missing template":
				pack.BootstrapTemplateJSON = nil
			case "stale support":
				pack.ClientSupportJSON = []byte("{}\n")
			case "source placeholder":
				root["generator_id"] = "cartulary.contractgen.extensions.v1"
			case "unknown root member":
				root["extra"] = true
			case "missing manifest digest":
				rows := root["supporting_contract_artifact_digests"].([]any)
				filtered := rows[:0]
				for _, raw := range rows {
					if raw.(map[string]any)["artifact_id"] != "client-asset-set-manifest" {
						filtered = append(filtered, raw)
					}
				}
				root["supporting_contract_artifact_digests"] = filtered
			case "coherent unsupported profile":
				support := decodeMutableObject(t, pack.ClientSupportJSON)
				support["profiles"].([]any)[0].(map[string]any)["supported_contract_majors"] = []any{999}
				var err error
				pack.ClientSupportJSON, err = canonicaljson.MarshalExtension(support)
				if err != nil {
					t.Fatal(err)
				}
				for _, raw := range root["supporting_contract_artifact_digests"].([]any) {
					row := raw.(map[string]any)
					if row["artifact_id"] == "client-support-registry" {
						row["artifact_sha256"] = browserbuild.SHA(pack.ClientSupportJSON)
					}
				}
			}
			if name == "source placeholder" || name == "unknown root member" || name == "missing manifest digest" || name == "coherent unsupported profile" {
				var err error
				pack.IntegrityJSON, err = canonicaljson.MarshalExtension(root)
				if err != nil {
					t.Fatal(err)
				}
				pack.IntegritySHA256 = browserbuild.SHA(pack.IntegrityJSON)
			}
			coordinator, err := NewPackagedCoordinator(pack)
			if name != "valid" {
				if err == nil {
					t.Fatal("invalid final browser package admitted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			resolution, err := coordinator.ResolveClaims([]string{})
			if err != nil {
				t.Fatal(err)
			}
			plan, err := coordinator.BuildPublicationPlan(resolution)
			if err != nil {
				t.Fatal(err)
			}
			if plan.Summary().ClientSupportRegistrySHA256 != browserbuild.SHA(c.ClientSupportJSON) {
				t.Fatal("publication lost admitted browser identity")
			}
			pack.ClientSupportJSON[0] = 'x'
			again, err := coordinator.BuildPublicationPlan(resolution)
			if err != nil || again.Summary().ClientSupportRegistrySHA256 != plan.Summary().ClientSupportRegistrySHA256 {
				t.Fatal("caller mutated admitted browser identity")
			}
		})
	}
}
