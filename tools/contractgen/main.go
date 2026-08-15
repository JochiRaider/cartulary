package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
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
	Dir                   string
	ContractRoot          string
	GoName                string
	GeneratedOutputs      []string
	TypeScriptProjections []typeScriptProjection
	Artifacts             []artifact
	OutputOrder           int
}

type typeScriptProjection struct {
	Kind         string
	OutputPath   string
	Identifier   string
	ArtifactPath string
}

type stagedGeneratedFile struct {
	data []byte
	mode os.FileMode
}

var (
	stagedGeneratedFiles   = map[string]stagedGeneratedFile{}
	stagedGeneratedDeletes = map[string]struct{}{}
)

func main() {
	root, err := repoRoot()
	if err != nil {
		fatal(err)
	}

	plan, err := loadGenerationPlan(root)
	if err != nil {
		fatal(err)
	}
	families := plan.Families

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
	if err := writeGraphProjectionRestoreBindingGo(root, families); err != nil {
		fatal(err)
	}
	if err := writeNetworkFlowMappingRegistryGo(root, families); err != nil {
		fatal(err)
	}
	if err := writeAdministrativeAuditRegistryGo(root, families); err != nil {
		fatal(err)
	}
	if err := writeImportTargetRegistryGo(root, families); err != nil {
		fatal(err)
	}
	if err := writeViewSchemaSourceTypesGo(root); err != nil {
		fatal(err)
	}
	if err := writeArtifactSourceCatalogGo(root); err != nil {
		fatal(err)
	}
	plan.Families = families
	if err := writeTypeScript(root, plan); err != nil {
		fatal(err)
	}
	if err := writeViewContractProjectionTypeScript(root); err != nil {
		fatal(err)
	}
	if err := writeImportTargetRegistryTypeScript(root, families); err != nil {
		fatal(err)
	}
	if err := commitGeneratedFiles(); err != nil {
		fatal(err)
	}
}

type networkFlowMappingRegistry struct {
	SchemaID            string                                    `json:"schema_id"`
	ProfileID           string                                    `json:"profile_id"`
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

type administrativeAuditRegistry struct {
	SchemaID                    string                             `json:"schema_id"`
	ScopeKinds                  []string                           `json:"scope_kinds"`
	ActorKinds                  []string                           `json:"actor_kinds"`
	Sources                     []string                           `json:"sources"`
	ValueStates                 []string                           `json:"value_states"`
	Actions                     []administrativeAuditActionBinding `json:"actions"`
	ForbiddenVisibleFieldTokens []string                           `json:"forbidden_visible_field_tokens"`
}

type administrativeAuditActionBinding struct {
	ActionCode string `json:"action_code"`
	ScopeKind  string `json:"scope_kind"`
	TargetKind string `json:"target_kind"`
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
	if familyDir == "imports" {
		generated, err := deriveImportTargetArtifacts(root)
		if err != nil {
			return nil, fmt.Errorf("derive import-target artifacts: %w", err)
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
	if err := rejectDuplicateJSONMembers(data); err != nil {
		return nil, err
	}
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

func rejectDuplicateJSONMembers(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanUniqueJSONValue(decoder, "$"); err != nil {
		return fmt.Errorf("contract input contains invalid or duplicate JSON members: %w", err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("contract input must contain exactly one JSON document")
		}
		return fmt.Errorf("contract input trailing content: %w", err)
	}
	return nil
}

func scanUniqueJSONValue(decoder *json.Decoder, path string) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("%s object member is not a string", path)
			}
			if _, duplicate := seen[key]; duplicate {
				return fmt.Errorf("%s has duplicate member %q", path, key)
			}
			seen[key] = struct{}{}
			if err := scanUniqueJSONValue(decoder, path+"."+key); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return fmt.Errorf("%s object has invalid closing delimiter", path)
		}
	case '[':
		index := 0
		for decoder.More() {
			if err := scanUniqueJSONValue(decoder, fmt.Sprintf("%s[%d]", path, index)); err != nil {
				return err
			}
			index++
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return fmt.Errorf("%s array has invalid closing delimiter", path)
		}
	default:
		return fmt.Errorf("%s has unexpected delimiter %q", path, delimiter)
	}
	return nil
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
	for _, current := range families {
		var buffer bytes.Buffer
		packageName := contractFamilyPackage(current.Dir)
		buffer.WriteString("// Code generated by tools/contractgen; DO NOT EDIT.\n\n")
		buffer.WriteString("package ")
		buffer.WriteString(packageName)
		buffer.WriteString("\n\n")
		buffer.WriteString("type Artifact struct {\n")
		buffer.WriteString("\tPath string `json:\"path\"`\n")
		buffer.WriteString("\tJSON string `json:\"json\"`\n")
		buffer.WriteString("\tSHA256 string `json:\"sha256\"`\n")
		buffer.WriteString("}\n\n")
		buffer.WriteString("var Artifacts = []Artifact{\n")
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
		buffer.WriteString("var Index = func() map[string]Artifact {\n")
		buffer.WriteString("\tindexed := make(map[string]Artifact, len(Artifacts))\n")
		buffer.WriteString("\tfor _, artifact := range Artifacts {\n")
		buffer.WriteString("\t\tindexed[artifact.Path] = artifact\n")
		buffer.WriteString("\t}\n")
		buffer.WriteString("\treturn indexed\n")
		buffer.WriteString("}()\n")

		formatted, err := format.Source(buffer.Bytes())
		if err != nil {
			return fmt.Errorf("format generated Go family %s: %w", current.Dir, err)
		}
		outputDir := filepath.Join(root, "internal", "gen", packageName)
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("create generated Go family %s: %w", current.Dir, err)
		}
		if err := stageGeneratedFile(filepath.Join(outputDir, "artifacts_gen.go"), formatted, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func contractFamilyPackage(familyDir string) string {
	return "contract" + strings.NewReplacer("-", "", "_", "", ".", "").Replace(familyDir)
}

func writeGraphProjectionRestoreBindingGo(root string, families []family) error {
	const (
		registryPath = "contracts/recovery/fixtures/graph-projection-restore-source-registry.v1.json"
		bindingPath  = "contracts/recovery/fixtures/graph-projection-restore-implementation-binding.v1.json"
	)
	var registryArtifact artifact
	var bindingArtifact artifact
	for _, current := range families {
		if current.Dir != "recovery" {
			continue
		}
		for _, currentArtifact := range current.Artifacts {
			switch currentArtifact.Path {
			case registryPath:
				registryArtifact = currentArtifact
			case bindingPath:
				bindingArtifact = currentArtifact
			}
		}
	}
	if registryArtifact.Path == "" || bindingArtifact.Path == "" {
		return fmt.Errorf("missing current Graph Projection restore registry or binding artifact")
	}
	var legacyBinding map[string]any
	if err := json.Unmarshal([]byte(bindingArtifact.JSON), &legacyBinding); err != nil {
		return fmt.Errorf("decode current Graph Projection restore binding for historical projection: %w", err)
	}
	legacyBinding["binding_id"] = "graphprojection.restore_rebuild.legacy_empty_registry.v1"
	legacyBinding["graph_projection_contract_id"] = "cartulary.graph_projection_nlspec.v1.1.0"
	legacyJSONBytes, err := json.Marshal(legacyBinding)
	if err != nil {
		return fmt.Errorf("encode historical Graph Projection restore binding: %w", err)
	}
	legacySum := sha256.Sum256(legacyJSONBytes)
	legacySHA256 := hex.EncodeToString(legacySum[:])

	var buffer bytes.Buffer
	buffer.WriteString("// Code generated by tools/contractgen; DO NOT EDIT.\n\n")
	buffer.WriteString("package contractrecovery\n\n")
	fmt.Fprintf(&buffer, "const CurrentGraphProjectionRestoreSourceRegistryJSON = %s\n", strconv.Quote(registryArtifact.JSON))
	fmt.Fprintf(&buffer, "const CurrentGraphProjectionRestoreSourceRegistrySHA256 = %s\n\n", strconv.Quote(registryArtifact.SHA256))
	fmt.Fprintf(&buffer, "const CurrentGraphProjectionRestoreImplementationBindingJSON = %s\n", strconv.Quote(bindingArtifact.JSON))
	fmt.Fprintf(&buffer, "const CurrentGraphProjectionRestoreImplementationBindingSHA256 = %s\n\n", strconv.Quote(bindingArtifact.SHA256))
	fmt.Fprintf(&buffer, "const LegacyGraphProjectionRestoreImplementationBindingJSON = %s\n", strconv.Quote(string(legacyJSONBytes)))
	fmt.Fprintf(&buffer, "const LegacyGraphProjectionRestoreImplementationBindingSHA256 = %s\n", strconv.Quote(legacySHA256))

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("format generated Graph Projection restore binding: %w", err)
	}
	return stageGeneratedFile(
		filepath.Join(root, "internal", "gen", "contractrecovery", "graph_projection_restore_binding_gen.go"),
		formatted,
		0o644,
	)
}

func writeNetworkFlowMappingRegistryGo(root string, families []family) error {
	const registryPath = "contracts/network-flow/mapping-registry.v2.json"
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
	buffer.WriteString("\tSchemaID string\n\tProfileID string\n\tTargetKind string\n\tTargetTableSchemaID string\n\tSystemDerivations []SystemDerivation\n\tSourceProfiles []SourceProfile\n}\n\n")
	buffer.WriteString("var Registry = MappingRegistry{\n")
	fmt.Fprintf(&buffer, "\tSchemaID: %s,\n", strconv.Quote(registry.SchemaID))
	fmt.Fprintf(&buffer, "\tProfileID: %s,\n", strconv.Quote(registry.ProfileID))
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
	return stageGeneratedFile(filepath.Join(outputDir, "registry_gen.go"), formatted, 0o644)
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

func writeAdministrativeAuditRegistryGo(root string, families []family) error {
	const registryPath = "contracts/audit/index.json"
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

	var registry administrativeAuditRegistry
	if err := json.Unmarshal([]byte(registryJSON), &registry); err != nil {
		return fmt.Errorf("decode %s for Go registry: %w", registryPath, err)
	}
	if registry.SchemaID == "" || len(registry.Actions) == 0 || len(registry.ForbiddenVisibleFieldTokens) == 0 {
		return fmt.Errorf("%s is incomplete", registryPath)
	}

	var buffer bytes.Buffer
	buffer.WriteString("// Code generated by tools/contractgen; DO NOT EDIT.\n\n")
	buffer.WriteString("package administrativeauditregistry\n\n")
	buffer.WriteString("const SchemaID = ")
	buffer.WriteString(strconv.Quote(registry.SchemaID))
	buffer.WriteString("\n\nconst (\n")
	writeAuditConstants(&buffer, "Scope", registry.ScopeKinds)
	writeAuditConstants(&buffer, "Actor", registry.ActorKinds)
	writeAuditConstants(&buffer, "Source", registry.Sources)
	writeAuditConstants(&buffer, "Value", registry.ValueStates)
	actionCodes := make([]string, 0, len(registry.Actions))
	targetKinds := make([]string, 0, len(registry.Actions))
	for _, binding := range registry.Actions {
		actionCodes = append(actionCodes, binding.ActionCode)
		targetKinds = append(targetKinds, binding.TargetKind)
	}
	writeAuditConstants(&buffer, "Action", sortedUnique(actionCodes))
	writeAuditConstants(&buffer, "Target", sortedUnique(targetKinds))
	buffer.WriteString(")\n\ntype ActionBinding struct {\n\tActionCode string\n\tScopeKind string\n\tTargetKind string\n}\n\n")
	buffer.WriteString("var ActionBindings = []ActionBinding{\n")
	for _, binding := range registry.Actions {
		fmt.Fprintf(&buffer, "\t{ActionCode: %s, ScopeKind: %s, TargetKind: %s},\n", strconv.Quote(binding.ActionCode), strconv.Quote(binding.ScopeKind), strconv.Quote(binding.TargetKind))
	}
	buffer.WriteString("}\n\nvar ForbiddenVisibleFieldTokens = []string{\n")
	for _, token := range registry.ForbiddenVisibleFieldTokens {
		fmt.Fprintf(&buffer, "\t%s,\n", strconv.Quote(token))
	}
	buffer.WriteString("}\n")

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("format generated administrative audit registry: %w", err)
	}
	outputDir := filepath.Join(root, "internal", "gen", "administrativeauditregistry")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create generated administrative audit registry directory: %w", err)
	}
	return stageGeneratedFile(filepath.Join(outputDir, "registry_gen.go"), formatted, 0o644)
}

func writeAuditConstants(buffer *bytes.Buffer, prefix string, values []string) {
	for _, value := range values {
		fmt.Fprintf(buffer, "\t%s%s = %s\n", prefix, goExportedToken(value), strconv.Quote(value))
	}
}

func sortedUnique(values []string) []string {
	unique := map[string]struct{}{}
	for _, value := range values {
		unique[value] = struct{}{}
	}
	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func goExportedToken(value string) string {
	parts := strings.Split(value, "_")
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func writeViewSchemaSourceTypesGo(root string) error {
	source := `// Code generated by tools/contractgen from cartulary.view_schema_source.v1; DO NOT EDIT.

package viewschemasource

type Field struct {
	FieldKey string ` + "`json:\"field_key\"`" + `
	Label string ` + "`json:\"label\"`" + `
	DefaultHidden bool ` + "`json:\"default_hidden\"`" + `
	Sortable bool ` + "`json:\"sortable\"`" + `
	Groupable bool ` + "`json:\"groupable\"`" + `
	ReadKind string ` + "`json:\"read_kind\"`" + `
	ReadModel string ` + "`json:\"read_model\"`" + `
	WriteKind string ` + "`json:\"write_kind\"`" + `
	GridEditable bool ` + "`json:\"grid_editable\"`" + `
	Writable bool ` + "`json:\"writable\"`" + `
	CreateWritable bool ` + "`json:\"create_writable\"`" + `
	HeaderSortFieldKey *string ` + "`json:\"header_sort_field_key\"`" + `
	FilterOps []string ` + "`json:\"filter_ops\"`" + `
	ConflictResolutionClass string ` + "`json:\"conflict_resolution_class\"`" + `
	EntityBindingMode *string ` + "`json:\"entity_binding_mode\"`" + `
	StringContractID *string ` + "`json:\"string_contract_id\"`" + `
	DirectScalarContractID *string ` + "`json:\"direct_scalar_contract_id\"`" + `
	DirectReferenceContractID *string ` + "`json:\"direct_reference_contract_id\"`" + `
	WriteTarget *string ` + "`json:\"write_target\"`" + `
	WriteAction *string ` + "`json:\"write_action\"`" + `
	Clearable bool ` + "`json:\"clearable\"`" + `
	EnumValues []string ` + "`json:\"enum_values\"`" + `
}

type SortEntry struct {
	FieldKey string ` + "`json:\"field_key\"`" + `
	Direction string ` + "`json:\"direction\"`" + `
}

type InlineCreate struct {
	MinimumCreateFieldSets [][]string ` + "`json:\"minimum_create_field_sets\"`" + `
	PermitsZeroFieldCreate bool ` + "`json:\"permits_zero_field_create\"`" + `
}

type SyntheticFilterPredicate struct {
	FieldKey string ` + "`json:\"field_key\"`" + `
	Label string ` + "`json:\"label\"`" + `
	FilterOps []string ` + "`json:\"filter_ops\"`" + `
}

type CanonicalSourceFilter struct {
	Kind string ` + "`json:\"kind\"`" + `
	Field string ` + "`json:\"field\"`" + `
	Value string ` + "`json:\"value\"`" + `
}

type InspectorSubjectBinding struct { Kind string ` + "`json:\"kind\"`" + ` }
type InspectorPanel struct {
	PanelID string ` + "`json:\"panel_id\"`" + `
	Label string ` + "`json:\"label\"`" + `
}
type InspectorRouteBinding struct {
	Kind string ` + "`json:\"kind\"`" + `
	Owner string ` + "`json:\"owner\"`" + `
	TargetViewSchemaID string ` + "`json:\"target_view_schema_id,omitempty\"`" + `
	ActionKey string ` + "`json:\"action_key,omitempty\"`" + `
}
type InspectorSeedSource struct {
	Kind string ` + "`json:\"kind\"`" + `
	SourceFieldKey string ` + "`json:\"source_field_key,omitempty\"`" + `
	Value any ` + "`json:\"value,omitempty\"`" + `
}
type InspectorSeedBinding struct {
	TargetFieldKey string ` + "`json:\"target_field_key\"`" + `
	Source InspectorSeedSource ` + "`json:\"source\"`" + `
}
type InspectorFeatureGroup struct {
	FeatureGroupKey string ` + "`json:\"feature_group_key\"`" + `
	PanelID string ` + "`json:\"panel_id\"`" + `
	Label string ` + "`json:\"label\"`" + `
	MinimumIncidentRole *string ` + "`json:\"minimum_incident_role\"`" + `
	Mutates bool ` + "`json:\"mutates\"`" + `
	RequiresConfirmation bool ` + "`json:\"requires_confirmation\"`" + `
	RouteBinding InspectorRouteBinding ` + "`json:\"route_binding\"`" + `
	SeedBindings []InspectorSeedBinding ` + "`json:\"seed_bindings\"`" + `
	DisabledWhen []string ` + "`json:\"disabled_when\"`" + `
	SuccessResultBehavior string ` + "`json:\"success_result_behavior\"`" + `
	FailureResultBehavior string ` + "`json:\"failure_result_behavior\"`" + `
}
type InspectorConfig struct {
	InspectorConfigSchemaID string ` + "`json:\"inspector_config_schema_id\"`" + `
	ViewSchemaID string ` + "`json:\"view_schema_id\"`" + `
	DefaultOpen bool ` + "`json:\"default_open\"`" + `
	SubjectBinding InspectorSubjectBinding ` + "`json:\"subject_binding\"`" + `
	NoRowState string ` + "`json:\"no_row_state\"`" + `
	UnsupportedFeatureBehavior string ` + "`json:\"unsupported_feature_behavior\"`" + `
	Panels []InspectorPanel ` + "`json:\"panels\"`" + `
	FeatureGroups []InspectorFeatureGroup ` + "`json:\"feature_groups\"`" + `
}
type CreateInput struct {
	InputKey string ` + "`json:\"input_key\"`" + `
	ValueContractID string ` + "`json:\"value_contract_id\"`" + `
	Required bool ` + "`json:\"required\"`" + `
	Nullable bool ` + "`json:\"nullable\"`" + `
}

type Document struct {
	Schema string ` + "`json:\"$schema\"`" + `
	SchemaID string ` + "`json:\"schema_id\"`" + `
	ViewSchemaID string ` + "`json:\"view_schema_id\"`" + `
	Title string ` + "`json:\"title\"`" + `
	SurfaceKind string ` + "`json:\"surface_kind\"`" + `
	SourceRecordTypes []string ` + "`json:\"source_record_types\"`" + `
	BaseProjection string ` + "`json:\"base_projection\"`" + `
	CanonicalSourceFilter *CanonicalSourceFilter ` + "`json:\"canonical_source_filter\"`" + `
	TechnicalFields []string ` + "`json:\"technical_fields\"`" + `
	RequiredReferencePackKeys []string ` + "`json:\"required_reference_pack_keys\"`" + `
	DefaultVisibleFields []string ` + "`json:\"default_visible_fields\"`" + `
	DefaultHiddenFields []string ` + "`json:\"default_hidden_fields\"`" + `
	DefaultSort []SortEntry ` + "`json:\"default_sort\"`" + `
	SortFields []string ` + "`json:\"sort_fields\"`" + `
	SortNullOrder string ` + "`json:\"sort_null_order\"`" + `
	FilterFields []string ` + "`json:\"filter_fields\"`" + `
	SyntheticFilterPredicates []SyntheticFilterPredicate ` + "`json:\"synthetic_filter_predicates\"`" + `
	GroupingFields []string ` + "`json:\"grouping_fields\"`" + `
	CreateCapable bool ` + "`json:\"create_capable\"`" + `
	CreateInputs []CreateInput ` + "`json:\"create_inputs\"`" + `
	InlineCreate InlineCreate ` + "`json:\"inline_create\"`" + `
	InspectorConfig InspectorConfig ` + "`json:\"inspector_config\"`" + `
	Fields []Field ` + "`json:\"fields\"`" + `
}

type RegistryIndex struct {
	Schema string ` + "`json:\"$schema\"`" + `
	RegistryID string ` + "`json:\"registry_id\"`" + `
	Note string ` + "`json:\"note\"`" + `
	ViewSchemas []RegistryIndexEntry ` + "`json:\"view_schemas\"`" + `
}
type RegistryIndexEntry struct {
	ViewSchemaID string ` + "`json:\"view_schema_id\"`" + `
	Title string ` + "`json:\"title\"`" + `
	SurfaceKind string ` + "`json:\"surface_kind\"`" + `
	SurfaceStatus string ` + "`json:\"surface_status\"`" + `
	SourceRecordTypes []string ` + "`json:\"source_record_types\"`" + `
	RequiredReferencePackKeys []string ` + "`json:\"required_reference_pack_keys\"`" + `
	ArtifactPath string ` + "`json:\"artifact_path\"`" + `
}
`
	formatted, err := format.Source([]byte(source))
	if err != nil {
		return fmt.Errorf("format generated view-schema source types: %w", err)
	}
	outputDir := filepath.Join(root, "internal", "gen", "viewschemasource")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create generated view-schema source directory: %w", err)
	}
	return stageGeneratedFile(filepath.Join(outputDir, "types_gen.go"), formatted, 0o644)
}

func writeTypeScript(root string, plan generationPlan) error {
	writtenOutputs := map[string]struct{}{}
	for _, current := range plan.Families {
		for _, projection := range current.TypeScriptProjections {
			if projection.Kind != jsonValueProjection {
				return fmt.Errorf("unsupported TypeScript projection kind %s", projection.Kind)
			}
			source, err := jsonValueSource(current, projection)
			if err != nil {
				return err
			}
			if err := stageGeneratedFile(filepath.Join(root, filepath.FromSlash(projection.OutputPath)), source, 0o644); err != nil {
				return err
			}
			writtenOutputs[projection.OutputPath] = struct{}{}
		}
	}
	for _, output := range contractgenManagedTypeScriptOutputs {
		if _, written := writtenOutputs[output]; written {
			continue
		}
		if err := stageGeneratedDelete(filepath.Join(root, filepath.FromSlash(output))); err != nil {
			return err
		}
	}
	return nil
}

var contractgenManagedTypeScriptOutputs = []string{
	"packages/protocol-ts/src/generated/artifact.ts",
	"packages/protocol-ts/src/generated/audit-artifacts.ts",
	"packages/protocol-ts/src/generated/error-registry.ts",
	"packages/protocol-ts/src/generated/errors-artifacts.ts",
	"packages/protocol-ts/src/generated/extension-client-support-registry.ts",
	"packages/protocol-ts/src/generated/extension-profile-registry.ts",
	"packages/protocol-ts/src/generated/extensions-artifacts.ts",
	"packages/protocol-ts/src/generated/index.ts",
	"packages/protocol-ts/src/generated/network-flow-artifacts.ts",
	"packages/protocol-ts/src/generated/network-flow-error-registry.ts",
	"packages/protocol-ts/src/generated/revisions-artifacts.ts",
	"packages/protocol-ts/src/generated/view-schema-registry.ts",
	"packages/protocol-ts/src/generated/view-schema-source-types.ts",
	"packages/protocol-ts/src/generated/view-schemas-artifacts.ts",
	"packages/protocol-ts/src/generated/ws-artifacts.ts",
}

func jsonValueSource(current family, projection typeScriptProjection) ([]byte, error) {
	var selected *artifact
	for index := range current.Artifacts {
		if current.Artifacts[index].Path == projection.ArtifactPath {
			selected = &current.Artifacts[index]
			break
		}
	}
	if selected == nil {
		return nil, fmt.Errorf("contract family %s TypeScript json_value path %s matches no artifact", current.Dir, projection.ArtifactPath)
	}
	var decoded any
	if err := json.Unmarshal([]byte(selected.JSON), &decoded); err != nil {
		return nil, fmt.Errorf("decode TypeScript json_value path %s: %w", projection.ArtifactPath, err)
	}
	encoded, err := json.MarshalIndent(decoded, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode TypeScript json_value path %s: %w", projection.ArtifactPath, err)
	}
	var buffer bytes.Buffer
	buffer.WriteString("// Code generated by tools/contractgen; DO NOT EDIT.\n\n")
	buffer.WriteString("function deepFreeze<Value>(value: Value): Value {\n")
	buffer.WriteString("  if (value !== null && typeof value === \"object\" && !Object.isFrozen(value)) {\n")
	buffer.WriteString("    for (const child of Object.values(value)) {\n")
	buffer.WriteString("      deepFreeze(child);\n")
	buffer.WriteString("    }\n")
	buffer.WriteString("    Object.freeze(value);\n")
	buffer.WriteString("  }\n")
	buffer.WriteString("  return value;\n")
	buffer.WriteString("}\n\nconst value = ")
	buffer.Write(encoded)
	buffer.WriteString(" as const;\n\nexport const ")
	buffer.WriteString(projection.Identifier)
	buffer.WriteString(" = deepFreeze(value);\n")
	return buffer.Bytes(), nil
}

func stageGeneratedFile(path string, data []byte, mode os.FileMode) error {
	cleanPath := filepath.Clean(path)
	if _, duplicate := stagedGeneratedFiles[cleanPath]; duplicate {
		return fmt.Errorf("generated output %s was staged more than once", cleanPath)
	}
	if _, deleting := stagedGeneratedDeletes[cleanPath]; deleting {
		return fmt.Errorf("generated output %s cannot be staged for write and deletion", cleanPath)
	}
	stagedGeneratedFiles[cleanPath] = stagedGeneratedFile{
		data: append([]byte(nil), data...),
		mode: mode,
	}
	return nil
}

func stageGeneratedDelete(path string) error {
	cleanPath := filepath.Clean(path)
	if _, writing := stagedGeneratedFiles[cleanPath]; writing {
		return fmt.Errorf("generated output %s cannot be staged for deletion and write", cleanPath)
	}
	stagedGeneratedDeletes[cleanPath] = struct{}{}
	return nil
}

type generatedFilePublication struct {
	path         string
	temporary    string
	previousData []byte
	previousMode os.FileMode
	existed      bool
}

func commitGeneratedFiles() error {
	paths := make([]string, 0, len(stagedGeneratedFiles)+len(stagedGeneratedDeletes))
	for path := range stagedGeneratedFiles {
		paths = append(paths, path)
	}
	for path := range stagedGeneratedDeletes {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	publications := make([]generatedFilePublication, 0, len(paths))
	cleanupTemporaries := func() {
		for _, publication := range publications {
			if publication.temporary != "" {
				_ = os.Remove(publication.temporary)
			}
		}
	}

	for _, path := range paths {
		publication := generatedFilePublication{path: path}
		info, err := os.Lstat(path)
		switch {
		case err == nil:
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				cleanupTemporaries()
				return fmt.Errorf("generated output %s must be a regular file when it exists", path)
			}
			publication.previousData, err = os.ReadFile(path)
			if err != nil {
				cleanupTemporaries()
				return fmt.Errorf("read previous generated output %s: %w", path, err)
			}
			publication.previousMode = info.Mode().Perm()
			publication.existed = true
		case os.IsNotExist(err):
		default:
			cleanupTemporaries()
			return fmt.Errorf("stat generated output %s: %w", path, err)
		}

		if staged, writing := stagedGeneratedFiles[path]; writing {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				cleanupTemporaries()
				return fmt.Errorf("create generated output directory for %s: %w", path, err)
			}
			temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
			if err != nil {
				cleanupTemporaries()
				return fmt.Errorf("create temporary generated output for %s: %w", path, err)
			}
			publication.temporary = temporary.Name()
			if err := temporary.Chmod(staged.mode); err != nil {
				_ = temporary.Close()
				_ = os.Remove(publication.temporary)
				cleanupTemporaries()
				return fmt.Errorf("chmod temporary generated output for %s: %w", path, err)
			}
			if _, err := temporary.Write(staged.data); err != nil {
				_ = temporary.Close()
				_ = os.Remove(publication.temporary)
				cleanupTemporaries()
				return fmt.Errorf("write temporary generated output for %s: %w", path, err)
			}
			if err := temporary.Sync(); err != nil {
				_ = temporary.Close()
				_ = os.Remove(publication.temporary)
				cleanupTemporaries()
				return fmt.Errorf("sync temporary generated output for %s: %w", path, err)
			}
			if err := temporary.Close(); err != nil {
				_ = os.Remove(publication.temporary)
				cleanupTemporaries()
				return fmt.Errorf("close temporary generated output for %s: %w", path, err)
			}
		}
		publications = append(publications, publication)
	}

	restore := func(published int) error {
		var restoreErrors []error
		for index := published - 1; index >= 0; index-- {
			publication := publications[index]
			if publication.existed {
				if err := restoreGeneratedFile(
					publication.path,
					publication.previousData,
					publication.previousMode,
				); err != nil {
					restoreErrors = append(restoreErrors, err)
				}
				continue
			}
			if err := os.Remove(publication.path); err != nil && !os.IsNotExist(err) {
				restoreErrors = append(restoreErrors, fmt.Errorf("remove newly published output %s: %w", publication.path, err))
			}
		}
		return errors.Join(restoreErrors...)
	}

	for index := range publications {
		publication := &publications[index]
		var err error
		if publication.temporary != "" {
			err = os.Rename(publication.temporary, publication.path)
			if err == nil {
				publication.temporary = ""
			}
		} else if publication.existed {
			err = os.Remove(publication.path)
		}
		if err != nil {
			rollbackErr := restore(index)
			cleanupTemporaries()
			if rollbackErr != nil {
				return fmt.Errorf("publish generated output %s: %w; rollback failed: %v", publication.path, err, rollbackErr)
			}
			return fmt.Errorf("publish generated output %s: %w", publication.path, err)
		}
	}
	cleanupTemporaries()
	return nil
}

func restoreGeneratedFile(path string, data []byte, mode os.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".rollback-*")
	if err != nil {
		return fmt.Errorf("create rollback output for %s: %w", path, err)
	}
	temporaryName := temporary.Name()
	defer func() {
		_ = os.Remove(temporaryName)
	}()
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("chmod rollback output for %s: %w", path, err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write rollback output for %s: %w", path, err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync rollback output for %s: %w", path, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close rollback output for %s: %w", path, err)
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return fmt.Errorf("restore generated output %s: %w", path, err)
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "contractgen: %v\n", err)
	os.Exit(1)
}
