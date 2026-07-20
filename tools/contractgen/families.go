package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	contractFamilyRegistrySchemaID = "cartulary.contract_family_registry.v1"
	contractFamilyRegistryID       = "cartulary.contract_families.v1"
)

type contractFamilyRegistry struct {
	Schema     string                `json:"$schema"`
	SchemaID   string                `json:"schema_id"`
	RegistryID string                `json:"registry_id"`
	Note       string                `json:"note"`
	Families   []contractFamilyEntry `json:"families"`
}

type contractFamilyEntry struct {
	FamilyID                  string   `json:"family_id"`
	ContractRoot              string   `json:"contract_root"`
	GenerationStatus          string   `json:"generation_status"`
	GoName                    string   `json:"go_name"`
	TSName                    string   `json:"ts_name"`
	OutputOrder               int      `json:"output_order"`
	OwnerDocument             string   `json:"owner_document"`
	OwnerSections             []string `json:"owner_sections"`
	GeneratedOutputs          []string `json:"generated_outputs"`
	TypeScriptRuntimePrefixes []string `json:"typescript_runtime_artifact_prefixes"`
	ActivationDependencyIDs   []string `json:"activation_dependency_ids"`
	Description               string   `json:"description"`
}

func loadFamilies(root string) ([]family, error) {
	registryPath := filepath.Join(root, "contracts", "index.json")
	file, err := os.Open(registryPath)
	if err != nil {
		return nil, fmt.Errorf("open contract family registry: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var registry contractFamilyRegistry
	if err := decoder.Decode(&registry); err != nil {
		return nil, fmt.Errorf("decode contract family registry: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return nil, fmt.Errorf("contract family registry must contain exactly one JSON document")
	}

	return activeFamiliesFromRegistry(root, registry)
}

func activeFamiliesFromRegistry(root string, registry contractFamilyRegistry) ([]family, error) {
	if registry.SchemaID != contractFamilyRegistrySchemaID {
		return nil, fmt.Errorf("contracts/index.json schema_id must be %s", contractFamilyRegistrySchemaID)
	}
	if registry.RegistryID != contractFamilyRegistryID {
		return nil, fmt.Errorf("contracts/index.json registry_id must be %s", contractFamilyRegistryID)
	}
	if strings.TrimSpace(registry.Note) == "" {
		return nil, fmt.Errorf("contracts/index.json note is required")
	}
	if len(registry.Families) == 0 {
		return nil, fmt.Errorf("contracts/index.json families must not be empty")
	}

	seenFamilyIDs := map[string]struct{}{}
	seenRoots := map[string]string{}
	seenGoNames := map[string]string{}
	seenTSNames := map[string]string{}
	seenOutputOrders := map[int]string{}
	active := []family{}

	for index, entry := range registry.Families {
		label := fmt.Sprintf("contracts/index.json.families[%d]", index+1)
		if strings.TrimSpace(entry.FamilyID) == "" {
			return nil, fmt.Errorf("%s.family_id is required", label)
		}
		if _, exists := seenFamilyIDs[entry.FamilyID]; exists {
			return nil, fmt.Errorf("duplicate contract family id %s", entry.FamilyID)
		}
		seenFamilyIDs[entry.FamilyID] = struct{}{}

		dir, err := familyDir(entry.ContractRoot, label)
		if err != nil {
			return nil, err
		}
		if previous, exists := seenRoots[dir]; exists {
			return nil, fmt.Errorf("%s.contract_root duplicates %s", label, previous)
		}
		seenRoots[dir] = entry.FamilyID
		if _, err := os.Stat(filepath.Join(root, "contracts", dir)); err != nil {
			return nil, fmt.Errorf("%s.contract_root %s is not an existing directory: %w", label, entry.ContractRoot, err)
		}
		if strings.TrimSpace(entry.GoName) == "" {
			return nil, fmt.Errorf("%s.go_name is required", label)
		}
		if previous, exists := seenGoNames[entry.GoName]; exists {
			return nil, fmt.Errorf("%s.go_name duplicates family %s", label, previous)
		}
		seenGoNames[entry.GoName] = entry.FamilyID
		if strings.TrimSpace(entry.TSName) == "" {
			return nil, fmt.Errorf("%s.ts_name is required", label)
		}
		if previous, exists := seenTSNames[entry.TSName]; exists {
			return nil, fmt.Errorf("%s.ts_name duplicates family %s", label, previous)
		}
		seenTSNames[entry.TSName] = entry.FamilyID
		if previous, exists := seenOutputOrders[entry.OutputOrder]; exists {
			return nil, fmt.Errorf("%s.output_order duplicates family %s", label, previous)
		}
		seenOutputOrders[entry.OutputOrder] = entry.FamilyID
		if strings.TrimSpace(entry.OwnerDocument) == "" {
			return nil, fmt.Errorf("%s.owner_document is required", label)
		}
		if len(entry.OwnerSections) == 0 {
			return nil, fmt.Errorf("%s.owner_sections must not be empty", label)
		}
		if len(entry.GeneratedOutputs) == 0 {
			return nil, fmt.Errorf("%s.generated_outputs must not be empty", label)
		}
		if len(entry.TypeScriptRuntimePrefixes) == 0 {
			return nil, fmt.Errorf("%s.typescript_runtime_artifact_prefixes must not be empty", label)
		}
		seenTypeScriptPrefixes := map[string]struct{}{}
		for _, prefix := range entry.TypeScriptRuntimePrefixes {
			if _, duplicate := seenTypeScriptPrefixes[prefix]; duplicate {
				return nil, fmt.Errorf("%s.typescript_runtime_artifact_prefixes contains duplicate %s", label, prefix)
			}
			seenTypeScriptPrefixes[prefix] = struct{}{}
			if !strings.HasPrefix(prefix, entry.ContractRoot+"/") || strings.Contains(prefix, "..") {
				return nil, fmt.Errorf("%s.typescript_runtime_artifact_prefixes entry %s must stay within %s", label, prefix, entry.ContractRoot)
			}
		}
		if strings.TrimSpace(entry.Description) == "" {
			return nil, fmt.Errorf("%s.description is required", label)
		}

		switch entry.GenerationStatus {
		case "active":
			active = append(active, family{
				Dir:                       dir,
				GoName:                    entry.GoName,
				TSName:                    entry.TSName,
				TypeScriptRuntimePrefixes: append([]string(nil), entry.TypeScriptRuntimePrefixes...),
				OutputOrder:               entry.OutputOrder,
			})
		case "planned":
			if len(entry.ActivationDependencyIDs) == 0 {
				return nil, fmt.Errorf("%s.activation_dependency_ids must not be empty for planned families", label)
			}
		default:
			return nil, fmt.Errorf("%s.generation_status must be active or planned", label)
		}
	}

	sort.Slice(active, func(i, j int) bool {
		return active[i].OutputOrder < active[j].OutputOrder
	})
	for index, current := range active {
		if current.OutputOrder != index {
			return nil, fmt.Errorf("active contract family output_order values must be contiguous from 0")
		}
	}
	return active, nil
}

func familyDir(contractRoot, label string) (string, error) {
	if !strings.HasPrefix(contractRoot, "contracts/") {
		return "", fmt.Errorf("%s.contract_root must start with contracts/", label)
	}
	cleanRoot := filepath.ToSlash(filepath.Clean(contractRoot))
	if !strings.HasPrefix(cleanRoot, "contracts/") {
		return "", fmt.Errorf("%s.contract_root escapes contracts/", label)
	}
	relative := strings.TrimPrefix(cleanRoot, "contracts/")
	if relative == "" || strings.HasPrefix(relative, "../") || relative == ".." {
		return "", fmt.Errorf("%s.contract_root escapes contracts/", label)
	}
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("%s.contract_root must be repository-relative", label)
	}
	return relative, nil
}
