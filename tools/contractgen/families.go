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
	contractFamilyRegistrySchemaID = "cartulary.contract_family_registry.v5"
	contractFamilyRegistryID       = "cartulary.contract_families.v1"
	jsonValueProjection            = "json_value"
)

type contractFamilyRegistry struct {
	Schema     string                `json:"$schema"`
	SchemaID   string                `json:"schema_id"`
	RegistryID string                `json:"registry_id"`
	Note       string                `json:"note"`
	Families   []contractFamilyEntry `json:"families"`
}

type contractFamilyEntry struct {
	FamilyID                string                      `json:"family_id"`
	ContractRoot            string                      `json:"contract_root"`
	GenerationStatus        string                      `json:"generation_status"`
	GoName                  string                      `json:"go_name"`
	OutputOrder             int                         `json:"output_order"`
	GeneratedOutputs        []string                    `json:"generated_outputs"`
	TypeScriptProjections   []typeScriptProjectionEntry `json:"typescript_projections"`
	ActivationDependencyIDs []string                    `json:"activation_dependency_ids"`
	Description             string                      `json:"description"`
}

type typeScriptProjectionEntry struct {
	ProjectionKind string `json:"projection_kind"`
	OutputPath     string `json:"output_path"`
	Identifier     string `json:"identifier"`
	ArtifactPath   string `json:"artifact_path"`
}

type generationPlan struct {
	Families []family
}

func loadGenerationPlan(root string) (generationPlan, error) {
	registryPath := filepath.Join(root, "contracts", "index.json")
	file, err := os.Open(registryPath)
	if err != nil {
		return generationPlan{}, fmt.Errorf("open contract family registry: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var registry contractFamilyRegistry
	if err := decoder.Decode(&registry); err != nil {
		return generationPlan{}, fmt.Errorf("decode contract family registry: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return generationPlan{}, fmt.Errorf("contract family registry must contain exactly one JSON document")
	}

	return generationPlanFromRegistry(root, registry)
}

func loadFamilies(root string) ([]family, error) {
	plan, err := loadGenerationPlan(root)
	if err != nil {
		return nil, err
	}
	return plan.Families, nil
}

func generationPlanFromRegistry(root string, registry contractFamilyRegistry) (generationPlan, error) {
	if registry.Schema != contractDraft202012Schema {
		return generationPlan{}, fmt.Errorf("contracts/index.json $schema must be %s", contractDraft202012Schema)
	}
	if registry.SchemaID != contractFamilyRegistrySchemaID {
		return generationPlan{}, fmt.Errorf("contracts/index.json schema_id must be %s", contractFamilyRegistrySchemaID)
	}
	if registry.RegistryID != contractFamilyRegistryID {
		return generationPlan{}, fmt.Errorf("contracts/index.json registry_id must be %s", contractFamilyRegistryID)
	}
	if strings.TrimSpace(registry.Note) == "" {
		return generationPlan{}, fmt.Errorf("contracts/index.json note is required")
	}
	if len(registry.Families) == 0 {
		return generationPlan{}, fmt.Errorf("contracts/index.json families must not be empty")
	}
	seenFamilyIDs := map[string]struct{}{}
	seenRoots := map[string]string{}
	seenGoNames := map[string]string{}
	seenOutputOrders := map[int]string{}
	seenGeneratedOutputs := map[string]string{}
	seenProjectionOutputs := map[string]string{}
	seenProjectionIdentifiers := map[string]string{}
	active := []family{}

	for index, entry := range registry.Families {
		label := fmt.Sprintf("contracts/index.json.families[%d]", index+1)
		if strings.TrimSpace(entry.FamilyID) == "" {
			return generationPlan{}, fmt.Errorf("%s.family_id is required", label)
		}
		if _, exists := seenFamilyIDs[entry.FamilyID]; exists {
			return generationPlan{}, fmt.Errorf("duplicate contract family id %s", entry.FamilyID)
		}
		seenFamilyIDs[entry.FamilyID] = struct{}{}

		dir, err := familyDir(entry.ContractRoot, label)
		if err != nil {
			return generationPlan{}, err
		}
		if previous, exists := seenRoots[dir]; exists {
			return generationPlan{}, fmt.Errorf("%s.contract_root duplicates %s", label, previous)
		}
		seenRoots[dir] = entry.FamilyID
		if _, err := os.Stat(filepath.Join(root, "contracts", dir)); err != nil {
			return generationPlan{}, fmt.Errorf("%s.contract_root %s is not an existing directory: %w", label, entry.ContractRoot, err)
		}
		if strings.TrimSpace(entry.GoName) == "" {
			return generationPlan{}, fmt.Errorf("%s.go_name is required", label)
		}
		if previous, exists := seenGoNames[entry.GoName]; exists {
			return generationPlan{}, fmt.Errorf("%s.go_name duplicates family %s", label, previous)
		}
		seenGoNames[entry.GoName] = entry.FamilyID
		if previous, exists := seenOutputOrders[entry.OutputOrder]; exists {
			return generationPlan{}, fmt.Errorf("%s.output_order duplicates family %s", label, previous)
		}
		seenOutputOrders[entry.OutputOrder] = entry.FamilyID
		if len(entry.GeneratedOutputs) == 0 {
			return generationPlan{}, fmt.Errorf("%s.generated_outputs must not be empty", label)
		}
		familyOutputs := map[string]struct{}{}
		for outputIndex, output := range entry.GeneratedOutputs {
			outputLabel := fmt.Sprintf("%s.generated_outputs[%d]", label, outputIndex+1)
			if err := validateRepositoryPath(output, outputLabel); err != nil {
				return generationPlan{}, err
			}
			if _, duplicate := familyOutputs[output]; duplicate {
				return generationPlan{}, fmt.Errorf("%s contains duplicate %s", label+".generated_outputs", output)
			}
			familyOutputs[output] = struct{}{}
			if previous, duplicate := seenGeneratedOutputs[output]; duplicate {
				return generationPlan{}, fmt.Errorf("%s duplicates generated output owned by family %s", outputLabel, previous)
			}
			seenGeneratedOutputs[output] = entry.FamilyID
		}

		projections := make([]typeScriptProjection, 0, len(entry.TypeScriptProjections))
		for projectionIndex, projection := range entry.TypeScriptProjections {
			projectionLabel := fmt.Sprintf("%s.typescript_projections[%d]", label, projectionIndex+1)
			if err := validateGeneratedTypeScriptPath(projection.OutputPath, projectionLabel+".output_path"); err != nil {
				return generationPlan{}, err
			}
			if _, declared := familyOutputs[projection.OutputPath]; !declared {
				return generationPlan{}, fmt.Errorf("%s.output_path %s must be declared in generated_outputs", projectionLabel, projection.OutputPath)
			}
			if previous, duplicate := seenProjectionOutputs[projection.OutputPath]; duplicate {
				return generationPlan{}, fmt.Errorf("%s.output_path duplicates projection owned by family %s", projectionLabel, previous)
			}
			seenProjectionOutputs[projection.OutputPath] = entry.FamilyID
			if err := registerProjectionIdentifier(seenProjectionIdentifiers, projection.Identifier, entry.FamilyID, projectionLabel+".identifier"); err != nil {
				return generationPlan{}, err
			}

			if projection.ProjectionKind != jsonValueProjection {
				return generationPlan{}, fmt.Errorf("%s.projection_kind must be json_value", projectionLabel)
			}
			if projection.ArtifactPath == "" {
				return generationPlan{}, fmt.Errorf("%s json_value requires artifact_path", projectionLabel)
			}
			if err := validateArtifactSelection(projection.ArtifactPath, entry.ContractRoot, projectionLabel+".artifact_path"); err != nil {
				return generationPlan{}, err
			}
			validated := typeScriptProjection{
				Kind:         projection.ProjectionKind,
				OutputPath:   projection.OutputPath,
				Identifier:   projection.Identifier,
				ArtifactPath: projection.ArtifactPath,
			}
			projections = append(projections, validated)
		}
		if strings.TrimSpace(entry.Description) == "" {
			return generationPlan{}, fmt.Errorf("%s.description is required", label)
		}

		switch entry.GenerationStatus {
		case "active":
			active = append(active, family{
				Dir:                   dir,
				ContractRoot:          entry.ContractRoot,
				GoName:                entry.GoName,
				GeneratedOutputs:      append([]string(nil), entry.GeneratedOutputs...),
				TypeScriptProjections: projections,
				OutputOrder:           entry.OutputOrder,
			})
		case "planned":
			if len(entry.ActivationDependencyIDs) == 0 {
				return generationPlan{}, fmt.Errorf("%s.activation_dependency_ids must not be empty for planned families", label)
			}
		default:
			return generationPlan{}, fmt.Errorf("%s.generation_status must be active or planned", label)
		}
	}

	sort.Slice(active, func(i, j int) bool {
		return active[i].OutputOrder < active[j].OutputOrder
	})
	for index, current := range active {
		if current.OutputOrder != index {
			return generationPlan{}, fmt.Errorf("active contract family output_order values must be contiguous from 0")
		}
	}
	return generationPlan{Families: active}, nil
}

func registerProjectionIdentifier(seen map[string]string, identifier, familyID, label string) error {
	if strings.TrimSpace(identifier) == "" {
		return fmt.Errorf("%s is required", label)
	}
	if previous, duplicate := seen[identifier]; duplicate {
		return fmt.Errorf("%s duplicates projection identifier owned by family %s", label, previous)
	}
	seen[identifier] = familyID
	return nil
}

func validateArtifactSelection(selection, contractRoot, label string) error {
	if strings.HasSuffix(selection, "/") {
		return fmt.Errorf("%s entry %s must select one exact artifact", label, selection)
	}
	if err := validateRepositoryPath(selection, label); err != nil {
		return err
	}
	if !strings.HasPrefix(selection, contractRoot+"/") || strings.Contains(selection, "..") {
		return fmt.Errorf("%s entry %s must stay within %s", label, selection, contractRoot)
	}
	return nil
}

func validateGeneratedTypeScriptPath(path, label string) error {
	if err := validateRepositoryPath(path, label); err != nil {
		return err
	}
	if !strings.HasPrefix(path, "packages/protocol-ts/src/generated/") || !strings.HasSuffix(path, ".ts") {
		return fmt.Errorf("%s must be a TypeScript file under packages/protocol-ts/src/generated", label)
	}
	return nil
}

func validateRepositoryPath(path, label string) error {
	if path == "" || filepath.IsAbs(path) || filepath.ToSlash(filepath.Clean(path)) != path || strings.HasPrefix(path, "../") || path == ".." {
		return fmt.Errorf("%s must be a normalized repository-relative path", label)
	}
	return nil
}

func familyDir(contractRoot, label string) (string, error) {
	if !strings.HasPrefix(contractRoot, "contracts/") {
		return "", fmt.Errorf("%s.contract_root must start with contracts/", label)
	}
	cleanRoot := filepath.ToSlash(filepath.Clean(contractRoot))
	if cleanRoot != contractRoot || !strings.HasPrefix(cleanRoot, "contracts/") {
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
