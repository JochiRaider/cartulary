package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type artifact struct {
	Path   string
	JSON   string
	SHA256 string
}

type family struct {
	Dir                       string
	GoName                    string
	TSName                    string
	TypeScriptRuntimePrefixes []string
	Artifacts                 []artifact
	OutputOrder               int
}

func main() {
	root, err := repoRoot()
	if err != nil {
		fatal(err)
	}
	if outputPath := os.Getenv("CARTULARY_EXTENSION_TRACEABILITY_OUTPUT"); outputPath != "" {
		document, err := os.ReadFile(filepath.Join(root, "docs", "extension-subsystem-nlspec.md"))
		if err != nil {
			fatal(err)
		}
		output, err := marshalExtensionTraceabilityMappingSource(document)
		if err != nil {
			fatal(err)
		}
		if err := os.WriteFile(outputPath, output, 0o600); err != nil {
			fatal(err)
		}
		return
	}
	if outputRoot := os.Getenv("CARTULARY_EXTENSION_MANIFEST_OUTPUT"); outputRoot != "" {
		if err := refreshExtensionOwnerInputs(root, outputRoot); err != nil {
			fatal(err)
		}
		return
	}

	families, err := loadFamilies(root)
	if err != nil {
		fatal(err)
	}

	for index := range families {
		artifacts, err := collectArtifacts(root, families[index].Dir)
		if err != nil {
			fatal(err)
		}
		families[index].Artifacts = artifacts
	}

	if err := writeGo(root, families); err != nil {
		fatal(err)
	}
	if err := writeNetworkFlowMappingRegistryGo(root, families); err != nil {
		fatal(err)
	}
	if err := writeTypeScript(root, families); err != nil {
		fatal(err)
	}
}

type networkFlowMappingRegistry struct {
	SchemaID            string                                    `json:"schema_id"`
	ProfileID           string                                    `json:"profile_id"`
	DocumentVersion     string                                    `json:"document_version"`
	TargetKind          string                                    `json:"target_kind"`
	TargetTableSchemaID string                                    `json:"target_table_schema_id"`
	SystemDerivations   []networkFlowMappingSystemDerivation      `json:"system_derivations"`
	SourceProfiles      []networkFlowMappingRegistrySourceProfile `json:"source_profiles"`
}

type networkFlowMappingSystemDerivation struct {
	FieldKey      string `json:"field_key"`
	DerivationID  string `json:"derivation_id"`
	Combinability string `json:"combinability"`
}

type networkFlowMappingRegistrySourceProfile struct {
	SourceProfileID                string                             `json:"source_profile_id"`
	DisplayName                    string                             `json:"display_name"`
	ConformanceStatus              string                             `json:"conformance_status"`
	ParserProfileID                string                             `json:"parser_profile_id"`
	DefaultUnknownColumnPolicy     string                             `json:"default_unknown_column_policy"`
	SupportedUnknownColumnPolicies []string                           `json:"supported_unknown_column_policies"`
	DefaultTimestampProfile        networkFlowMappingTimestampProfile `json:"default_timestamp_profile"`
	SupportedTimestampModes        []string                           `json:"supported_timestamp_modes"`
	Fields                         []networkFlowMappingRegistryField  `json:"fields"`
}

type networkFlowMappingTimestampProfile struct {
	SchemaID                 string  `json:"schema_id"`
	Mode                     string  `json:"mode"`
	Precision                string  `json:"precision"`
	Timezone                 *string `json:"timezone"`
	TimezoneRulesetID        *string `json:"timezone_ruleset_id"`
	AmbiguousLocalTimePolicy string  `json:"ambiguous_local_time_policy"`
	LocalTimeGapPolicy       string  `json:"local_time_gap_policy"`
}

type networkFlowMappingRegistryField struct {
	FieldKey         string   `json:"field_key"`
	Requirement      string   `json:"requirement"`
	TransformID      *string  `json:"transform_id"`
	EmptyValuePolicy *string  `json:"empty_value_policy"`
	Aliases          []string `json:"aliases"`
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repo root from %q", dir)
		}
		dir = parent
	}
}

func collectArtifacts(root, familyDir string) ([]artifact, error) {
	baseDir := filepath.Join(root, "contracts", familyDir)
	contractRoot, err := os.OpenRoot(baseDir)
	if err != nil {
		return nil, fmt.Errorf("open contract root %s: %w", baseDir, err)
	}
	defer contractRoot.Close()

	var artifacts []artifact

	if err := filepath.WalkDir(baseDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		name := entry.Name()
		if name == ".gitkeep" || !supportedContractFile(name) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("stat %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if !pathWithinRoot(baseDir, path) {
			return fmt.Errorf("contract input path %s escapes %s", path, baseDir)
		}

		relativeInputPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return fmt.Errorf("relativize contract input %s: %w", path, err)
		}
		data, err := contractRoot.ReadFile(relativeInputPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		decoded, err := decodeContract(data)
		if err != nil {
			return fmt.Errorf("decode %s: %w", path, err)
		}
		if err := validateContractInput(familyDir, filepath.ToSlash(relativeInputPath), decoded); err != nil {
			return fmt.Errorf("validate %s: %w", path, err)
		}
		canonicalJSON, err := canonicalizeDecoded(decoded)
		if err != nil {
			return fmt.Errorf("canonicalize %s: %w", path, err)
		}
		if familyDir == "extensions" {
			canonicalJSON += "\n"
		}

		relativePath, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("relativize %s: %w", path, err)
		}

		hash := sha256.Sum256([]byte(canonicalJSON))
		artifacts = append(artifacts, artifact{
			Path:   filepath.ToSlash(relativePath),
			JSON:   canonicalJSON,
			SHA256: hex.EncodeToString(hash[:]),
		})
		return nil
	}); err != nil {
		return nil, err
	}

	if err := validateContractFamily(root, familyDir); err != nil {
		return nil, err
	}
	if familyDir == "extensions" {
		generated, err := deriveExtensionArtifacts(root)
		if err != nil {
			return nil, fmt.Errorf("derive extension artifacts: %w", err)
		}
		artifacts = append(artifacts, generated...)
	}

	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].Path < artifacts[j].Path
	})

	return artifacts, nil
}

func supportedContractFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".json", ".yaml", ".yml":
		return true
	default:
		return false
	}
}

func pathWithinRoot(root string, target string) bool {
	cleanRoot := filepath.Clean(root)
	cleanTarget := filepath.Clean(target)
	relative, err := filepath.Rel(cleanRoot, cleanTarget)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func decodeContract(data []byte) (any, error) {
	var decoded any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode contract input as JSON or JSON-compatible YAML: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return nil, fmt.Errorf("contract input must contain exactly one JSON document")
	}

	return normalize(decoded), nil
}

func canonicalizeDecoded(normalized any) (string, error) {
	canonicalJSON, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}

	return string(canonicalJSON), nil
}

func normalize(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		normalized := make(map[string]any, len(typed))
		for key, item := range typed {
			normalized[key] = normalize(item)
		}
		return normalized
	case map[any]any:
		normalized := make(map[string]any, len(typed))
		for key, item := range typed {
			normalized[fmt.Sprint(key)] = normalize(item)
		}
		return normalized
	case []any:
		items := make([]any, len(typed))
		for index, item := range typed {
			items[index] = normalize(item)
		}
		return items
	default:
		return typed
	}
}

func writeGo(root string, families []family) error {
	var buffer bytes.Buffer

	buffer.WriteString("// Code generated by tools/contractgen; DO NOT EDIT.\n\n")
	buffer.WriteString("package contracts\n\n")
	buffer.WriteString("type Artifact struct {\n")
	buffer.WriteString("\tPath string `json:\"path\"`\n")
	buffer.WriteString("\tJSON string `json:\"json\"`\n")
	buffer.WriteString("\tSHA256 string `json:\"sha256\"`\n")
	buffer.WriteString("}\n\n")
	buffer.WriteString("func indexArtifacts(artifacts []Artifact) map[string]Artifact {\n")
	buffer.WriteString("\tindexed := make(map[string]Artifact, len(artifacts))\n")
	buffer.WriteString("\tfor _, artifact := range artifacts {\n")
	buffer.WriteString("\t\tindexed[artifact.Path] = artifact\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\treturn indexed\n")
	buffer.WriteString("}\n\n")
	buffer.WriteString("func mergeArtifactIndexes(artifactSets ...[]Artifact) map[string]Artifact {\n")
	buffer.WriteString("\ttotal := 0\n")
	buffer.WriteString("\tfor _, artifacts := range artifactSets {\n")
	buffer.WriteString("\t\ttotal += len(artifacts)\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\tindexed := make(map[string]Artifact, total)\n")
	buffer.WriteString("\tfor _, artifacts := range artifactSets {\n")
	buffer.WriteString("\t\tfor _, artifact := range artifacts {\n")
	buffer.WriteString("\t\t\tindexed[artifact.Path] = artifact\n")
	buffer.WriteString("\t\t}\n")
	buffer.WriteString("\t}\n")
	buffer.WriteString("\treturn indexed\n")
	buffer.WriteString("}\n\n")

	for _, current := range families {
		buffer.WriteString("var ")
		buffer.WriteString(current.GoName)
		buffer.WriteString(" = []Artifact{\n")
		for _, currentArtifact := range current.Artifacts {
			buffer.WriteString("\t{\n")
			buffer.WriteString("\t\tPath: ")
			buffer.WriteString(strconv.Quote(currentArtifact.Path))
			buffer.WriteString(",\n")
			buffer.WriteString("\t\tJSON: ")
			buffer.WriteString(strconv.Quote(currentArtifact.JSON))
			buffer.WriteString(",\n")
			buffer.WriteString("\t\tSHA256: ")
			buffer.WriteString(strconv.Quote(currentArtifact.SHA256))
			buffer.WriteString(",\n")
			buffer.WriteString("\t},\n")
		}
		buffer.WriteString("}\n\n")

		buffer.WriteString("var ")
		buffer.WriteString(current.GoName)
		buffer.WriteString("Index = indexArtifacts(")
		buffer.WriteString(current.GoName)
		buffer.WriteString(")\n\n")
	}

	buffer.WriteString("var ContractArtifactIndex = mergeArtifactIndexes(\n")
	for _, current := range families {
		buffer.WriteString("\t")
		buffer.WriteString(current.GoName)
		buffer.WriteString(",\n")
	}
	buffer.WriteString(")\n")

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("format generated Go: %w", err)
	}

	outputPath := filepath.Join(root, "internal", "gen", "contracts", "contracts_gen.go")
	return os.WriteFile(outputPath, formatted, 0o644) // #nosec G306 -- generated repo source files intentionally keep normal source permissions.
}

func writeNetworkFlowMappingRegistryGo(root string, families []family) error {
	const registryPath = "contracts/network-flow/mapping-registry.v1.json"
	var registryJSON string
	for _, current := range families {
		for _, currentArtifact := range current.Artifacts {
			if currentArtifact.Path == registryPath {
				registryJSON = currentArtifact.JSON
				break
			}
		}
	}
	if registryJSON == "" {
		return fmt.Errorf("missing %s", registryPath)
	}

	var registry networkFlowMappingRegistry
	if err := json.Unmarshal([]byte(registryJSON), &registry); err != nil {
		return fmt.Errorf("decode %s for Go registry: %w", registryPath, err)
	}
	if registry.SchemaID == "" || registry.TargetKind == "" || len(registry.SourceProfiles) == 0 {
		return fmt.Errorf("%s is incomplete", registryPath)
	}

	var buffer bytes.Buffer
	buffer.WriteString("// Code generated by tools/contractgen; DO NOT EDIT.\n\n")
	buffer.WriteString("package networkflowmapping\n\n")
	buffer.WriteString("type TimestampProfile struct {\n")
	buffer.WriteString("\tSchemaID string\n\tMode string\n\tPrecision string\n\tTimezone string\n\tTimezoneRulesetID string\n\tAmbiguousLocalTimePolicy string\n\tLocalTimeGapPolicy string\n}\n\n")
	buffer.WriteString("type Field struct {\n")
	buffer.WriteString("\tFieldKey string\n\tRequirement string\n\tTransformID string\n\tEmptyValuePolicy string\n\tAliases []string\n}\n\n")
	buffer.WriteString("type SourceProfile struct {\n")
	buffer.WriteString("\tSourceProfileID string\n\tDisplayName string\n\tConformanceStatus string\n\tParserProfileID string\n\tDefaultUnknownColumnPolicy string\n\tSupportedUnknownColumnPolicies []string\n\tDefaultTimestampProfile TimestampProfile\n\tSupportedTimestampModes []string\n\tFields []Field\n}\n\n")
	buffer.WriteString("type SystemDerivation struct {\n\tFieldKey string\n\tDerivationID string\n\tCombinability string\n}\n\n")
	buffer.WriteString("type MappingRegistry struct {\n")
	buffer.WriteString("\tSchemaID string\n\tProfileID string\n\tDocumentVersion string\n\tTargetKind string\n\tTargetTableSchemaID string\n\tSystemDerivations []SystemDerivation\n\tSourceProfiles []SourceProfile\n}\n\n")
	buffer.WriteString("var Registry = MappingRegistry{\n")
	fmt.Fprintf(&buffer, "\tSchemaID: %s,\n", strconv.Quote(registry.SchemaID))
	fmt.Fprintf(&buffer, "\tProfileID: %s,\n", strconv.Quote(registry.ProfileID))
	fmt.Fprintf(&buffer, "\tDocumentVersion: %s,\n", strconv.Quote(registry.DocumentVersion))
	fmt.Fprintf(&buffer, "\tTargetKind: %s,\n", strconv.Quote(registry.TargetKind))
	fmt.Fprintf(&buffer, "\tTargetTableSchemaID: %s,\n", strconv.Quote(registry.TargetTableSchemaID))
	buffer.WriteString("\tSystemDerivations: []SystemDerivation{\n")
	for _, derivation := range registry.SystemDerivations {
		fmt.Fprintf(&buffer, "\t\t{FieldKey: %s, DerivationID: %s, Combinability: %s},\n", strconv.Quote(derivation.FieldKey), strconv.Quote(derivation.DerivationID), strconv.Quote(derivation.Combinability))
	}
	buffer.WriteString("\t},\n\tSourceProfiles: []SourceProfile{\n")
	for _, profile := range registry.SourceProfiles {
		buffer.WriteString("\t\t{\n")
		fmt.Fprintf(&buffer, "\t\t\tSourceProfileID: %s,\n", strconv.Quote(profile.SourceProfileID))
		fmt.Fprintf(&buffer, "\t\t\tDisplayName: %s,\n", strconv.Quote(profile.DisplayName))
		fmt.Fprintf(&buffer, "\t\t\tConformanceStatus: %s,\n", strconv.Quote(profile.ConformanceStatus))
		fmt.Fprintf(&buffer, "\t\t\tParserProfileID: %s,\n", strconv.Quote(profile.ParserProfileID))
		fmt.Fprintf(&buffer, "\t\t\tDefaultUnknownColumnPolicy: %s,\n", strconv.Quote(profile.DefaultUnknownColumnPolicy))
		writeGoStringSlice(&buffer, "\t\t\tSupportedUnknownColumnPolicies", profile.SupportedUnknownColumnPolicies)
		buffer.WriteString("\t\t\tDefaultTimestampProfile: TimestampProfile{\n")
		fmt.Fprintf(&buffer, "\t\t\t\tSchemaID: %s, Mode: %s, Precision: %s,\n", strconv.Quote(profile.DefaultTimestampProfile.SchemaID), strconv.Quote(profile.DefaultTimestampProfile.Mode), strconv.Quote(profile.DefaultTimestampProfile.Precision))
		fmt.Fprintf(&buffer, "\t\t\t\tTimezone: %s, TimezoneRulesetID: %s,\n", strconv.Quote(stringPointerValue(profile.DefaultTimestampProfile.Timezone)), strconv.Quote(stringPointerValue(profile.DefaultTimestampProfile.TimezoneRulesetID)))
		fmt.Fprintf(&buffer, "\t\t\t\tAmbiguousLocalTimePolicy: %s, LocalTimeGapPolicy: %s,\n", strconv.Quote(profile.DefaultTimestampProfile.AmbiguousLocalTimePolicy), strconv.Quote(profile.DefaultTimestampProfile.LocalTimeGapPolicy))
		buffer.WriteString("\t\t\t},\n")
		writeGoStringSlice(&buffer, "\t\t\tSupportedTimestampModes", profile.SupportedTimestampModes)
		buffer.WriteString("\t\t\tFields: []Field{\n")
		for _, field := range profile.Fields {
			fmt.Fprintf(&buffer, "\t\t\t\t{FieldKey: %s, Requirement: %s, TransformID: %s, EmptyValuePolicy: %s, Aliases: []string{", strconv.Quote(field.FieldKey), strconv.Quote(field.Requirement), strconv.Quote(stringPointerValue(field.TransformID)), strconv.Quote(stringPointerValue(field.EmptyValuePolicy)))
			for index, alias := range field.Aliases {
				if index > 0 {
					buffer.WriteString(", ")
				}
				buffer.WriteString(strconv.Quote(alias))
			}
			buffer.WriteString("}},\n")
		}
		buffer.WriteString("\t\t\t},\n\t\t},\n")
	}
	buffer.WriteString("\t},\n}\n")

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("format generated Network Flow mapping registry: %w", err)
	}
	outputDir := filepath.Join(root, "internal", "gen", "networkflowmapping")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create generated Network Flow mapping registry directory: %w", err)
	}
	return os.WriteFile(filepath.Join(outputDir, "registry_gen.go"), formatted, 0o644) // #nosec G306 -- generated repo source files intentionally keep normal source permissions.
}

func writeGoStringSlice(buffer *bytes.Buffer, field string, values []string) {
	buffer.WriteString(field)
	buffer.WriteString(": []string{")
	for index, value := range values {
		if index > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(strconv.Quote(value))
	}
	buffer.WriteString("},\n")
}

func stringPointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func writeTypeScript(root string, families []family) error {
	contractsPath := filepath.Join(root, "packages", "protocol-ts", "src", "generated", "contracts.ts")
	indexPath := filepath.Join(root, "packages", "protocol-ts", "src", "generated", "index.ts")

	var buffer bytes.Buffer
	buffer.WriteString("// Code generated by tools/contractgen; DO NOT EDIT.\n\n")
	buffer.WriteString("export type Artifact = {\n")
	buffer.WriteString("  readonly path: string;\n")
	buffer.WriteString("  readonly json: string;\n")
	buffer.WriteString("  readonly sha256: string;\n")
	buffer.WriteString("};\n\n")
	buffer.WriteString("function indexArtifacts(\n")
	buffer.WriteString("  artifacts: readonly Artifact[],\n")
	buffer.WriteString("): Readonly<Record<string, Artifact>> {\n")
	buffer.WriteString("  return Object.freeze(\n")
	buffer.WriteString("    Object.fromEntries(\n")
	buffer.WriteString("      artifacts.map((artifact) => [artifact.path, artifact]),\n")
	buffer.WriteString("    ) as Record<string, Artifact>,\n")
	buffer.WriteString("  );\n")
	buffer.WriteString("}\n\n")

	for _, current := range families {
		typeScriptArtifacts, err := typeScriptRuntimeArtifacts(current)
		if err != nil {
			return err
		}
		buffer.WriteString("export const ")
		buffer.WriteString(current.TSName)
		buffer.WriteString(": readonly Artifact[] = [\n")
		for _, currentArtifact := range typeScriptArtifacts {
			buffer.WriteString("  {\n")
			buffer.WriteString("    path: ")
			buffer.WriteString(marshalJSONString(currentArtifact.Path))
			buffer.WriteString(",\n")
			buffer.WriteString("    json: ")
			buffer.WriteString(marshalJSONString(currentArtifact.JSON))
			buffer.WriteString(",\n")
			buffer.WriteString("    sha256: ")
			buffer.WriteString(marshalJSONString(currentArtifact.SHA256))
			buffer.WriteString(",\n")
			buffer.WriteString("  },\n")
		}
		buffer.WriteString("];\n\n")

		buffer.WriteString("export const ")
		buffer.WriteString(current.TSName)
		buffer.WriteString("Index = indexArtifacts(")
		buffer.WriteString(current.TSName)
		buffer.WriteString(");\n\n")
	}

	buffer.WriteString("export const contractArtifacts = {\n")
	sortedFamilies := append([]family(nil), families...)
	sort.Slice(sortedFamilies, func(i, j int) bool {
		return sortedFamilies[i].OutputOrder < sortedFamilies[j].OutputOrder
	})
	for _, current := range sortedFamilies {
		buffer.WriteString("  ")
		buffer.WriteString(current.TSName)
		buffer.WriteString(",\n")
	}
	buffer.WriteString("} as const;\n")
	buffer.WriteString("\n")
	buffer.WriteString("export const contractArtifactIndex = Object.freeze({\n")
	for _, current := range sortedFamilies {
		buffer.WriteString("  ...")
		buffer.WriteString(current.TSName)
		buffer.WriteString("Index,\n")
	}
	buffer.WriteString("}) as Readonly<Record<string, Artifact>>;\n")

	indexContent := "// Code generated by tools/contractgen; DO NOT EDIT.\n\nexport * from \"./contracts.js\";\n"

	if err := os.WriteFile(contractsPath, buffer.Bytes(), 0o644); err != nil { // #nosec G306 -- generated repo source files intentionally keep normal source permissions.
		return err
	}
	return os.WriteFile(indexPath, []byte(indexContent), 0o644) // #nosec G306 -- generated repo source files intentionally keep normal source permissions.
}

func typeScriptRuntimeArtifacts(current family) ([]artifact, error) {
	selected := make([]artifact, 0, len(current.Artifacts))
	matched := make([]bool, len(current.TypeScriptRuntimePrefixes))
	for _, currentArtifact := range current.Artifacts {
		for index, prefix := range current.TypeScriptRuntimePrefixes {
			if currentArtifact.Path == strings.TrimSuffix(prefix, "/") || strings.HasPrefix(currentArtifact.Path, prefix) {
				selected = append(selected, currentArtifact)
				matched[index] = true
				break
			}
		}
	}
	for index, ok := range matched {
		if !ok {
			return nil, fmt.Errorf("contract family %s TypeScript runtime prefix %s matches no artifact", current.Dir, current.TypeScriptRuntimePrefixes[index])
		}
	}
	return selected, nil
}

func marshalJSONString(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		fatal(fmt.Errorf("marshal JSON string: %w", err))
	}
	return string(encoded)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "contractgen: %v\n", err)
	os.Exit(1)
}
