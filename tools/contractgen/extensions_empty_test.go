package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func checkEmptyExtensionGeneration(t *testing.T, root string) {
	t.Helper()
	inputs, err := readExtensionInputCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	body, err := readExtensionSourceFile(root, "contracts/extensions/index.json")
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeContract(body)
	if err != nil {
		t.Fatal(err)
	}
	catalog := decoded.(map[string]any)
	kept := []any{}
	refs := map[string]bool{}
	for _, ref := range extensionProducerSources {
		refs[ref] = true
	}
	for _, raw := range catalog["artifacts"].([]any) {
		row := raw.(map[string]any)
		path := stringValue(row["path"])
		if strings.HasPrefix(path, "profiles/") || strings.HasPrefix(path, "fragments/") {
			continue
		}
		if strings.HasPrefix(path, "build/") && path != "build/base-route-reservations.json" && path != "build/client-support.json" && path != "build/implementation-bindings.json" {
			continue
		}
		kept = append(kept, row)
		refs["contracts/extensions/"+path] = true
	}
	catalog["artifacts"] = kept
	for ref := range refs {
		data, err := readExtensionSourceFile(root, ref)
		if err != nil {
			t.Fatal(err)
		}
		var object map[string]any
		switch ref {
		case "contracts/extensions/dependencies.json":
			object = cloneObject(inputs["dependencies.json"])
			for _, raw := range object["dependencies"].([]any) {
				row := raw.(map[string]any)
				for _, key := range []string{"imported_schema_ids", "imported_algorithm_ids", "imported_artifacts"} {
					row[key] = []any{}
				}
			}
		case "contracts/extensions/index.json":
			object = catalog
		case "contracts/extensions/build/implementation-bindings.json":
			object = cloneObject(inputs["build/implementation-bindings.json"])
			object["bindings"] = []any{}
		case "contracts/extensions/build/client-support.json":
			object = cloneObject(inputs["build/client-support.json"])
			object["rows"] = []any{}
		}
		if object != nil {
			data, err = json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
		}
		path := filepath.Join(directory, filepath.FromSlash(ref))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	emptyInputs, err := readExtensionInputCatalog(directory)
	if err != nil {
		t.Fatal(err)
	}
	for path, object := range emptyInputs {
		if err := validateExtensionArtifactShape(object, path); err != nil {
			t.Fatal(err)
		}
	}
	if err := validateExtensionOperationalBindings(emptyInputs); err != nil {
		t.Fatal(err)
	}
	artifacts, err := deriveExtensionArtifacts(directory)
	if err != nil {
		t.Fatalf("generate empty catalog: %v", err)
	}
	expected := map[string]string{"profile-registry.json": "profiles", "owner-input-registry.json": "owner_fragments", "state-registry.json": "profiles", "backup-registry.json": "profiles", "participant-registry.json": "participants", "client-support-registry.json": "profiles"}
	for _, artifact := range artifacts {
		relative := strings.TrimPrefix(artifact.Path, generatedExtensionPrefix)
		if strings.HasPrefix(relative, "descriptors/") || strings.HasPrefix(relative, "implementation-bindings/") || strings.HasPrefix(relative, "job-contracts/") || strings.HasPrefix(relative, "worker-runtime-contracts/") {
			t.Fatalf("empty catalog invented %s", relative)
		}
		if field, ok := expected[relative]; ok {
			decoded, err := decodeContract([]byte(artifact.JSON))
			if err != nil {
				t.Fatal(err)
			}
			values, ok := decoded.(map[string]any)[field].([]any)
			if !ok || len(values) != 0 {
				t.Fatalf("%s.%s is not an empty array", relative, field)
			}
			delete(expected, relative)
		}
	}
	if len(expected) != 0 {
		t.Fatalf("missing empty companion artifacts: %v", expected)
	}
}
