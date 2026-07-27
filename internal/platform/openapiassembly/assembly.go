package openapiassembly

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	manifestSchemaID = "cartulary.openapi_source_manifest.v3"
	policySchemaID   = "cartulary.openapi_assembly_policy.v1"
	unitRootRole     = "root"
	unitOwnerRole    = "owner"
)

var operationKeys = map[string]struct{}{
	"delete":  {},
	"get":     {},
	"head":    {},
	"options": {},
	"patch":   {},
	"post":    {},
	"put":     {},
	"trace":   {},
}

var requiredSecuritySchemeNames = []string{
	"bearerSession",
	"credentialBootstrapBearer",
	"csrfCookie",
	"csrfHeader",
	"sessionCookie",
}

var ownerIDPattern = regexp.MustCompile(`^(module|platform|app|web|package|harness)\.[a-z][a-z0-9_]{0,62}$`)

type manifest struct {
	SchemaID string      `json:"schema_id"`
	Target   string      `json:"target"`
	Policy   string      `json:"policy"`
	UnitRoot string      `json:"unit_root"`
	Units    []unitEntry `json:"units"`
}

type assemblyPolicy struct {
	SchemaID        string         `json:"schema_id"`
	Indent          string         `json:"indent"`
	TrailingNewline bool           `json:"trailing_newline"`
	Limits          resourceLimits `json:"limits"`
}

type resourceLimits struct {
	MaxUnits           int   `json:"max_units"`
	MaxUnitBytes       int64 `json:"max_unit_bytes"`
	MaxTotalInputBytes int64 `json:"max_total_input_bytes"`
	MaxJSONDepth       int   `json:"max_json_depth"`
	MaxPaths           int   `json:"max_paths"`
	MaxOperations      int   `json:"max_operations"`
	MaxNamedComponents int   `json:"max_named_components"`
}

type unitEntry struct {
	OwnerID string `json:"owner_id"`
	Path    string `json:"path"`
	Role    string `json:"role"`
}

type valueKind uint8

const (
	nullKind valueKind = iota
	boolKind
	numberKind
	stringKind
	arrayKind
	objectKind
)

type orderedMember struct {
	name  string
	value *orderedValue
}

type orderedValue struct {
	kind   valueKind
	scalar any
	array  []*orderedValue
	object []orderedMember
}

func Assemble(manifestPath string) ([]byte, string, error) {
	manifestAbsolute, err := filepath.Abs(manifestPath)
	if err != nil {
		return nil, "", fmt.Errorf("resolve manifest: %w", err)
	}
	repositoryRoot, err := findRepositoryRoot(manifestAbsolute)
	if err != nil {
		return nil, "", err
	}
	repositoryFiles, err := os.OpenRoot(repositoryRoot)
	if err != nil {
		return nil, "", fmt.Errorf("open repository root: %w", err)
	}
	defer repositoryFiles.Close()
	manifestRelative, err := relativeWithinRoot(repositoryRoot, manifestAbsolute)
	if err != nil {
		return nil, "", fmt.Errorf("manifest: %w", err)
	}
	manifestInfo, err := repositoryFiles.Lstat(manifestRelative)
	if err != nil {
		return nil, "", fmt.Errorf("stat manifest: %w", err)
	}
	if manifestInfo.Mode()&os.ModeSymlink != 0 {
		return nil, "", errors.New("manifest must not be a symlink")
	}
	manifestBytes, err := repositoryFiles.ReadFile(manifestRelative)
	if err != nil {
		return nil, "", fmt.Errorf("read manifest: %w", err)
	}
	if _, err := parseOrderedJSON(manifestBytes, 128); err != nil {
		return nil, "", fmt.Errorf("parse manifest: %w", err)
	}
	var sourceManifest manifest
	decoder := json.NewDecoder(bytes.NewReader(manifestBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&sourceManifest); err != nil {
		return nil, "", fmt.Errorf("decode manifest: %w", err)
	}
	if _, err := safePath(repositoryRoot, sourceManifest.Policy); err != nil {
		return nil, "", fmt.Errorf("policy: %w", err)
	}
	policy, err := loadAssemblyPolicy(repositoryFiles, sourceManifest.Policy)
	if err != nil {
		return nil, "", err
	}
	if err := validateManifest(sourceManifest, policy); err != nil {
		return nil, "", err
	}

	target, err := safePath(repositoryRoot, sourceManifest.Target)
	if err != nil {
		return nil, "", fmt.Errorf("target: %w", err)
	}
	unitRoot, err := safePath(repositoryRoot, sourceManifest.UnitRoot)
	if err != nil {
		return nil, "", fmt.Errorf("unit_root: %w", err)
	}
	if err := validateUnitInventory(unitRoot, repositoryRoot, sourceManifest.Units); err != nil {
		return nil, "", err
	}

	var aggregate *orderedValue
	var totalInputBytes int64
	rootCount := 0
	units := append([]unitEntry(nil), sourceManifest.Units...)
	sort.Slice(units, func(left, right int) bool {
		if units[left].Role != units[right].Role {
			return units[left].Role == unitRootRole
		}
		if units[left].OwnerID != units[right].OwnerID {
			return units[left].OwnerID < units[right].OwnerID
		}
		return units[left].Path < units[right].Path
	})
	for index, entry := range units {
		ownerPrefix := strings.TrimSuffix(sourceManifest.UnitRoot, "/") + "/" + entry.OwnerID + "/"
		if !strings.HasPrefix(entry.Path, ownerPrefix) {
			return nil, "", fmt.Errorf(
				"unit %q must be stored below its declared owner directory %q",
				entry.Path,
				ownerPrefix,
			)
		}
		if entry.Role == unitRootRole {
			rootCount++
			if entry.OwnerID != "platform.openapi" {
				return nil, "", fmt.Errorf("root unit owner must be platform.openapi, got %q", entry.OwnerID)
			}
		}
		if _, err := safePath(repositoryRoot, entry.Path); err != nil {
			return nil, "", fmt.Errorf("unit %d path: %w", index+1, err)
		}
		info, err := repositoryFiles.Lstat(entry.Path)
		if err != nil {
			return nil, "", fmt.Errorf("stat unit %q: %w", entry.Path, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, "", fmt.Errorf("unit %q must not be a symlink", entry.Path)
		}
		if !info.Mode().IsRegular() {
			return nil, "", fmt.Errorf("unit %q must be a regular file", entry.Path)
		}
		if info.Size() > policy.Limits.MaxUnitBytes {
			return nil, "", fmt.Errorf("unit %q exceeds max_unit_bytes", entry.Path)
		}
		totalInputBytes += info.Size()
		if totalInputBytes > policy.Limits.MaxTotalInputBytes {
			return nil, "", errors.New("unit inputs exceed max_total_input_bytes")
		}
		content, err := repositoryFiles.ReadFile(entry.Path)
		if err != nil {
			return nil, "", fmt.Errorf("read unit %q: %w", entry.Path, err)
		}
		unit, err := parseOrderedJSON(content, policy.Limits.MaxJSONDepth)
		if err != nil {
			return nil, "", fmt.Errorf("parse unit %q: %w", entry.Path, err)
		}
		if err := validateUnitShape(unit, entry); err != nil {
			return nil, "", fmt.Errorf("unit %q: %w", entry.Path, err)
		}
		if aggregate == nil {
			aggregate = &orderedValue{kind: objectKind}
		}
		if err := mergeUnit(aggregate, unit, entry); err != nil {
			return nil, "", fmt.Errorf("unit %q: %w", entry.Path, err)
		}
	}
	if rootCount != 1 {
		return nil, "", fmt.Errorf("manifest must contain exactly one root unit, got %d", rootCount)
	}
	if aggregate == nil {
		return nil, "", errors.New("assembly produced no document")
	}
	if err := validateAggregate(aggregate, policy.Limits); err != nil {
		return nil, "", err
	}
	sortCanonicalObjectMembers(aggregate, true)
	var output bytes.Buffer
	writeOrderedJSON(&output, aggregate, policy.Indent, 0)
	if policy.TrailingNewline {
		output.WriteByte('\n')
	}
	return output.Bytes(), target, nil
}

func validateManifest(sourceManifest manifest, policy assemblyPolicy) error {
	if sourceManifest.SchemaID != manifestSchemaID {
		return fmt.Errorf("manifest schema_id must be %q", manifestSchemaID)
	}
	if sourceManifest.Target == "" ||
		sourceManifest.Policy == "" ||
		sourceManifest.UnitRoot == "" {
		return errors.New("target, policy, and unit_root are required")
	}
	limits := policy.Limits
	if limits.MaxUnits != 256 ||
		limits.MaxUnitBytes != 2*1024*1024 ||
		limits.MaxTotalInputBytes != 16*1024*1024 ||
		limits.MaxJSONDepth != 128 ||
		limits.MaxPaths != 2048 ||
		limits.MaxOperations != 4096 ||
		limits.MaxNamedComponents != 16384 {
		return errors.New("manifest resource limits must match the platform.openapi contract")
	}
	if len(sourceManifest.Units) == 0 || len(sourceManifest.Units) > limits.MaxUnits {
		return fmt.Errorf("units must contain between 1 and %d entries", limits.MaxUnits)
	}
	seenPaths := make(map[string]struct{}, len(sourceManifest.Units))
	seenOwners := make(map[string]struct{}, len(sourceManifest.Units))
	rootCount := 0
	for index, entry := range sourceManifest.Units {
		if entry.OwnerID == "" || entry.Path == "" {
			return fmt.Errorf("unit %d requires owner_id and path", index+1)
		}
		if !ownerIDPattern.MatchString(entry.OwnerID) {
			return fmt.Errorf("unit %d has invalid owner_id %q", index+1, entry.OwnerID)
		}
		if entry.Role != unitRootRole && entry.Role != unitOwnerRole {
			return fmt.Errorf("unit %d has invalid role %q", index+1, entry.Role)
		}
		if _, duplicate := seenPaths[entry.Path]; duplicate {
			return fmt.Errorf("unit path %q is listed more than once", entry.Path)
		}
		seenPaths[entry.Path] = struct{}{}
		if _, duplicate := seenOwners[entry.OwnerID]; duplicate {
			return fmt.Errorf("owner %q has more than one source unit", entry.OwnerID)
		}
		seenOwners[entry.OwnerID] = struct{}{}
		if entry.Role == unitRootRole {
			rootCount++
		}
	}
	if rootCount != 1 {
		return fmt.Errorf("manifest must contain exactly one root unit, got %d", rootCount)
	}
	return nil
}

func safePath(base, candidate string) (string, error) {
	if candidate == "" || filepath.IsAbs(candidate) || strings.Contains(candidate, "\\") {
		return "", fmt.Errorf("%q must be a non-empty relative slash path", candidate)
	}
	clean := filepath.Clean(filepath.FromSlash(candidate))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%q escapes its base directory", candidate)
	}
	resolved := filepath.Join(base, clean)
	relative, err := filepath.Rel(base, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%q escapes its base directory", candidate)
	}
	return resolved, nil
}

func relativeWithinRoot(root, target string) (string, error) {
	relative, err := filepath.Rel(root, target)
	if err != nil ||
		relative == "." ||
		relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%q is not a file below repository root", target)
	}
	return relative, nil
}

func findRepositoryRoot(manifestPath string) (string, error) {
	for candidate := filepath.Dir(manifestPath); ; candidate = filepath.Dir(candidate) {
		info, err := os.Lstat(filepath.Join(candidate, "go.mod"))
		if err == nil && info.Mode().IsRegular() {
			return candidate, nil
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			return "", errors.New("manifest is not inside a repository containing go.mod")
		}
	}
}

func loadAssemblyPolicy(repositoryFiles *os.Root, path string) (assemblyPolicy, error) {
	var policy assemblyPolicy
	info, err := repositoryFiles.Lstat(path)
	if err != nil {
		return policy, fmt.Errorf("stat assembly policy: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return policy, errors.New("assembly policy must not be a symlink")
	}
	content, err := repositoryFiles.ReadFile(path)
	if err != nil {
		return policy, fmt.Errorf("read assembly policy: %w", err)
	}
	if _, err := parseOrderedJSON(content, 128); err != nil {
		return policy, fmt.Errorf("parse assembly policy: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&policy); err != nil {
		return policy, fmt.Errorf("decode assembly policy: %w", err)
	}
	if policy.SchemaID != policySchemaID {
		return policy, fmt.Errorf("assembly policy schema_id must be %q", policySchemaID)
	}
	if policy.Indent != "  " || !policy.TrailingNewline {
		return policy, errors.New("serialization profile must use two-space indentation and one trailing newline")
	}
	return policy, nil
}

func validateUnitInventory(unitRoot, manifestDir string, units []unitEntry) error {
	rootInfo, err := os.Lstat(unitRoot)
	if err != nil {
		return fmt.Errorf("stat unit root: %w", err)
	}
	if rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return errors.New("unit_root must be a real directory, not a symlink")
	}
	listed := make(map[string]struct{}, len(units))
	for _, entry := range units {
		path, err := safePath(manifestDir, entry.Path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(unitRoot, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("unit %q is outside unit_root", entry.Path)
		}
		listed[filepath.Clean(path)] = struct{}{}
	}
	found := make(map[string]struct{})
	err = filepath.WalkDir(unitRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("unit inventory contains symlink %q", path)
		}
		if entry.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unit inventory contains non-regular file %q", path)
		}
		found[filepath.Clean(path)] = struct{}{}
		return nil
	})
	if err != nil {
		return err
	}
	for path := range found {
		if _, ok := listed[path]; !ok {
			return fmt.Errorf("orphan unit %q is not listed in the manifest", path)
		}
	}
	for path := range listed {
		if _, ok := found[path]; !ok {
			return fmt.Errorf("listed unit %q is missing from unit_root", path)
		}
	}
	return nil
}

func parseOrderedJSON(content []byte, maxDepth int) (*orderedValue, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	value, err := parseValue(decoder, 0, maxDepth)
	if err != nil {
		return nil, err
	}
	if token, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unexpected trailing token %v", token)
	}
	return value, nil
}

func parseValue(decoder *json.Decoder, depth, maxDepth int) (*orderedValue, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch typed := token.(type) {
	case nil:
		return &orderedValue{kind: nullKind}, nil
	case bool:
		return &orderedValue{kind: boolKind, scalar: typed}, nil
	case json.Number:
		return &orderedValue{kind: numberKind, scalar: typed}, nil
	case string:
		return &orderedValue{kind: stringKind, scalar: typed}, nil
	case json.Delim:
		if depth+1 > maxDepth {
			return nil, fmt.Errorf("JSON depth exceeds %d", maxDepth)
		}
		switch typed {
		case '{':
			result := &orderedValue{kind: objectKind}
			seen := make(map[string]struct{})
			for decoder.More() {
				rawName, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				name, ok := rawName.(string)
				if !ok {
					return nil, errors.New("object member name must be a string")
				}
				if _, duplicate := seen[name]; duplicate {
					return nil, fmt.Errorf("duplicate object key %q", name)
				}
				seen[name] = struct{}{}
				child, err := parseValue(decoder, depth+1, maxDepth)
				if err != nil {
					return nil, err
				}
				result.object = append(result.object, orderedMember{name: name, value: child})
			}
			if _, err := decoder.Token(); err != nil {
				return nil, err
			}
			return result, nil
		case '[':
			result := &orderedValue{kind: arrayKind}
			for decoder.More() {
				child, err := parseValue(decoder, depth+1, maxDepth)
				if err != nil {
					return nil, err
				}
				result.array = append(result.array, child)
			}
			if _, err := decoder.Token(); err != nil {
				return nil, err
			}
			return result, nil
		default:
			return nil, fmt.Errorf("unexpected delimiter %q", typed)
		}
	default:
		return nil, fmt.Errorf("unsupported JSON token %T", typed)
	}
}

func validateUnitShape(unit *orderedValue, entry unitEntry) error {
	if unit.kind != objectKind {
		return errors.New("unit must be an object")
	}
	if len(unit.object) == 0 {
		return errors.New("unit must not be empty")
	}
	for _, member := range unit.object {
		switch member.name {
		case "openapi", "info":
			if entry.Role != unitRootRole || entry.OwnerID != "platform.openapi" {
				return fmt.Errorf("%q may be contributed only by the platform.openapi root unit", member.name)
			}
		case "tags":
			if member.value.kind != arrayKind {
				return errors.New("tags must be an array")
			}
		case "paths", "components":
			if member.value.kind != objectKind {
				return fmt.Errorf("%s must be an object", member.name)
			}
		default:
			return fmt.Errorf("unknown unit top-level key %q", member.name)
		}
	}
	return nil
}

func mergeUnit(aggregate, unit *orderedValue, entry unitEntry) error {
	for _, member := range unit.object {
		existing, ok := objectMember(aggregate, member.name)
		if !ok {
			aggregate.object = append(aggregate.object, cloneMember(member))
			continue
		}
		switch member.name {
		case "openapi", "info":
			return fmt.Errorf("aggregate metadata key %q is contributed more than once", member.name)
		case "tags":
			existing.array = append(existing.array, cloneValues(member.value.array)...)
		case "paths":
			if err := mergePaths(existing, member.value); err != nil {
				return err
			}
		case "components":
			if err := mergeComponents(existing, member.value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("cannot merge top-level key %q for owner %q", member.name, entry.OwnerID)
		}
	}
	return nil
}

func mergePaths(target, source *orderedValue) error {
	for _, pathMember := range source.object {
		existingPathItem, ok := objectMember(target, pathMember.name)
		if !ok {
			target.object = append(target.object, cloneMember(pathMember))
			continue
		}
		if existingPathItem.kind != objectKind || pathMember.value.kind != objectKind {
			return fmt.Errorf("path %q must be an object", pathMember.name)
		}
		for _, pathItemMember := range pathMember.value.object {
			if _, collision := objectMember(existingPathItem, pathItemMember.name); collision {
				return fmt.Errorf("path/method collision at %s %s", strings.ToUpper(pathItemMember.name), pathMember.name)
			}
			existingPathItem.object = append(existingPathItem.object, cloneMember(pathItemMember))
		}
	}
	return nil
}

func mergeComponents(target, source *orderedValue) error {
	for _, categoryMember := range source.object {
		existingCategory, ok := objectMember(target, categoryMember.name)
		if !ok {
			target.object = append(target.object, cloneMember(categoryMember))
			continue
		}
		if existingCategory.kind != objectKind || categoryMember.value.kind != objectKind {
			return fmt.Errorf("component category %q must be an object", categoryMember.name)
		}
		for _, namedComponent := range categoryMember.value.object {
			if _, collision := objectMember(existingCategory, namedComponent.name); collision {
				return fmt.Errorf("component collision at %s.%s", categoryMember.name, namedComponent.name)
			}
			existingCategory.object = append(existingCategory.object, cloneMember(namedComponent))
		}
	}
	return nil
}

func validateAggregate(
	document *orderedValue,
	limits resourceLimits,
) error {
	openapi, ok := objectMember(document, "openapi")
	if !ok || openapi.kind != stringKind {
		return errors.New("assembled document requires string openapi metadata")
	}
	info, ok := objectMember(document, "info")
	if !ok || info.kind != objectKind {
		return errors.New("assembled document requires info metadata")
	}
	paths, ok := objectMember(document, "paths")
	if !ok || paths.kind != objectKind {
		return errors.New("assembled document requires paths")
	}
	if len(paths.object) > limits.MaxPaths {
		return fmt.Errorf("assembled document exceeds max_paths %d", limits.MaxPaths)
	}
	operationIDs := make(map[string]string)
	operationCount := 0
	declaredSecuritySchemes, err := validateSecuritySchemeInventory(document)
	if err != nil {
		return err
	}
	usedSecuritySchemes := make(map[string]struct{})
	for _, pathMember := range paths.object {
		if !strings.HasPrefix(pathMember.name, "/") || pathMember.value.kind != objectKind {
			return fmt.Errorf("invalid path item %q", pathMember.name)
		}
		for _, pathItemMember := range pathMember.value.object {
			if _, ok := operationKeys[pathItemMember.name]; !ok {
				continue
			}
			operationCount++
			if operationCount > limits.MaxOperations {
				return fmt.Errorf("assembled document exceeds max_operations %d", limits.MaxOperations)
			}
			if pathItemMember.value.kind != objectKind {
				return fmt.Errorf("%s %s must be an object", strings.ToUpper(pathItemMember.name), pathMember.name)
			}
			operationIDValue, ok := objectMember(pathItemMember.value, "operationId")
			if !ok || operationIDValue.kind != stringKind || operationIDValue.scalar.(string) == "" {
				return fmt.Errorf("%s %s requires operationId", strings.ToUpper(pathItemMember.name), pathMember.name)
			}
			operationID := operationIDValue.scalar.(string)
			key := strings.ToUpper(pathItemMember.name) + " " + pathMember.name
			if existing, duplicate := operationIDs[operationID]; duplicate {
				return fmt.Errorf("duplicate operationId %q at %s and %s", operationID, existing, key)
			}
			operationIDs[operationID] = key
			missingParameters, err := validatePathParameters(document, pathMember.name, pathMember.value, pathItemMember.value)
			if err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
			if len(missingParameters) > 0 {
				return fmt.Errorf(
					"%s path placeholders without exact parameter declarations: %s",
					key,
					strings.Join(missingParameters, ", "),
				)
			}
			if _, declared := objectMember(pathItemMember.value, "responses"); !declared {
				return fmt.Errorf("%s: operation requires responses", key)
			} else if err := validateOperationResponses(pathItemMember.value); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
			if _, declared := objectMember(pathItemMember.value, "security"); !declared {
				return fmt.Errorf("%s: operation requires explicit security", key)
			} else if err := validateOperationSecurity(pathItemMember.value, declaredSecuritySchemes, usedSecuritySchemes); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
		}
	}
	if components, ok := objectMember(document, "components"); ok {
		if components.kind != objectKind {
			return errors.New("components must be an object")
		}
		namedComponentCount := 0
		for _, category := range components.object {
			if category.value.kind != objectKind {
				return fmt.Errorf("component category %q must be an object", category.name)
			}
			namedComponentCount += len(category.value.object)
		}
		if namedComponentCount > limits.MaxNamedComponents {
			return fmt.Errorf("assembled document exceeds max_named_components %d", limits.MaxNamedComponents)
		}
	}
	var refs []string
	collectRefs(document, &refs)
	sort.Strings(refs)
	for _, ref := range refs {
		if !strings.HasPrefix(ref, "#/") {
			return fmt.Errorf("external reference %q is not allowed", ref)
		}
		if _, ok := resolvePointer(document, ref); !ok {
			return fmt.Errorf("unresolved reference %q", ref)
		}
	}
	for schemeName := range declaredSecuritySchemes {
		if _, used := usedSecuritySchemes[schemeName]; !used {
			return fmt.Errorf("unused security scheme %q", schemeName)
		}
	}
	return nil
}

func validateSecuritySchemeInventory(document *orderedValue) (map[string]struct{}, error) {
	components, ok := objectMember(document, "components")
	if !ok || components.kind != objectKind {
		return nil, errors.New("assembled document requires components")
	}
	rawSchemes, ok := objectMember(components, "securitySchemes")
	if !ok || rawSchemes.kind != objectKind {
		return nil, errors.New("assembled document requires components.securitySchemes")
	}
	schemes := make(map[string]struct{}, len(rawSchemes.object))
	for _, member := range rawSchemes.object {
		schemes[member.name] = struct{}{}
	}
	actual := make([]string, 0, len(schemes))
	for name := range schemes {
		actual = append(actual, name)
	}
	sort.Strings(actual)
	if strings.Join(actual, "\x00") != strings.Join(requiredSecuritySchemeNames, "\x00") {
		return nil, fmt.Errorf(
			"security scheme inventory must be exactly %s, got %s",
			strings.Join(requiredSecuritySchemeNames, ", "),
			strings.Join(actual, ", "),
		)
	}
	return schemes, nil
}

func validateOperationResponses(operation *orderedValue) error {
	responses, ok := objectMember(operation, "responses")
	if !ok {
		return errors.New("operation requires responses")
	}
	if responses.kind != objectKind || len(responses.object) == 0 {
		return errors.New("operation responses must be a nonempty object")
	}
	for _, response := range responses.object {
		if !validResponseKey(response.name) {
			return fmt.Errorf("invalid response status key %q", response.name)
		}
		if response.value.kind != objectKind {
			return fmt.Errorf("response %q must be an object", response.name)
		}
		if ref, referenced := objectMember(response.value, "$ref"); referenced {
			if ref.kind != stringKind || ref.scalar.(string) == "" {
				return fmt.Errorf("response %q has invalid $ref", response.name)
			}
			continue
		}
		description, described := objectMember(response.value, "description")
		if !described || description.kind != stringKind || strings.TrimSpace(description.scalar.(string)) == "" {
			return fmt.Errorf("response %q requires a description or $ref", response.name)
		}
	}
	return nil
}

func validResponseKey(value string) bool {
	if value == "default" {
		return true
	}
	if len(value) != 3 || value[0] < '1' || value[0] > '5' {
		return false
	}
	for _, digit := range value[1:] {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func validateOperationSecurity(
	operation *orderedValue,
	declaredSchemes map[string]struct{},
	usedSchemes map[string]struct{},
) error {
	security, ok := objectMember(operation, "security")
	if !ok {
		return errors.New("operation requires explicit security")
	}
	if security.kind != arrayKind {
		return errors.New("operation security must be an array")
	}
	for _, requirement := range security.array {
		if requirement.kind != objectKind {
			return errors.New("security requirement must be an object")
		}
		for _, scheme := range requirement.object {
			if _, declared := declaredSchemes[scheme.name]; !declared {
				return fmt.Errorf("security requirement references unknown scheme %q", scheme.name)
			}
			if scheme.value.kind != arrayKind || len(scheme.value.array) != 0 {
				return fmt.Errorf("security requirement %q scopes must be an empty array", scheme.name)
			}
			usedSchemes[scheme.name] = struct{}{}
		}
	}
	return nil
}

func validatePathParameters(
	document *orderedValue,
	path string,
	pathItem, operation *orderedValue,
) ([]string, error) {
	placeholders := pathPlaceholders(path)
	if len(placeholders) == 0 {
		return nil, nil
	}
	declared := make(map[string]bool)
	for _, container := range []*orderedValue{pathItem, operation} {
		parameters, ok := objectMember(container, "parameters")
		if !ok {
			continue
		}
		if parameters.kind != arrayKind {
			return nil, errors.New("parameters must be an array")
		}
		for _, parameter := range parameters.array {
			resolved := parameter
			if refValue, ok := objectMember(parameter, "$ref"); ok && refValue.kind == stringKind {
				var found bool
				resolved, found = resolvePointer(document, refValue.scalar.(string))
				if !found {
					continue
				}
			}
			if resolved.kind != objectKind {
				continue
			}
			name, hasName := objectMember(resolved, "name")
			in, hasIn := objectMember(resolved, "in")
			required, hasRequired := objectMember(resolved, "required")
			if hasName && hasIn && hasRequired &&
				name.kind == stringKind && in.kind == stringKind && required.kind == boolKind &&
				in.scalar.(string) == "path" {
				declared[name.scalar.(string)] = required.scalar.(bool)
			}
		}
	}
	for _, placeholder := range placeholders {
		if required, ok := declared[placeholder]; !ok {
			continue
		} else if !required {
			return nil, fmt.Errorf("path parameter %q must be required", placeholder)
		}
	}
	missing := make([]string, 0, len(placeholders))
	for _, placeholder := range placeholders {
		if _, ok := declared[placeholder]; !ok {
			missing = append(missing, placeholder)
		}
	}
	sort.Strings(missing)
	return missing, nil
}

func pathPlaceholders(path string) []string {
	var placeholders []string
	for start := 0; start < len(path); {
		open := strings.IndexByte(path[start:], '{')
		if open < 0 {
			break
		}
		open += start
		closeOffset := strings.IndexByte(path[open+1:], '}')
		if closeOffset < 0 {
			break
		}
		closeIndex := open + 1 + closeOffset
		if closeIndex > open+1 {
			placeholders = append(placeholders, path[open+1:closeIndex])
		}
		start = closeIndex + 1
	}
	return placeholders
}

func collectRefs(value *orderedValue, refs *[]string) {
	switch value.kind {
	case objectKind:
		for _, member := range value.object {
			if member.name == "$ref" && member.value.kind == stringKind {
				*refs = append(*refs, member.value.scalar.(string))
			}
			collectRefs(member.value, refs)
		}
	case arrayKind:
		for _, child := range value.array {
			collectRefs(child, refs)
		}
	}
}

func resolvePointer(root *orderedValue, pointer string) (*orderedValue, bool) {
	if pointer == "#" {
		return root, true
	}
	if !strings.HasPrefix(pointer, "#/") {
		return nil, false
	}
	current := root
	for _, rawPart := range strings.Split(strings.TrimPrefix(pointer, "#/"), "/") {
		if current.kind != objectKind {
			return nil, false
		}
		part := strings.ReplaceAll(strings.ReplaceAll(rawPart, "~1", "/"), "~0", "~")
		var ok bool
		current, ok = objectMember(current, part)
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func objectMember(object *orderedValue, name string) (*orderedValue, bool) {
	if object == nil || object.kind != objectKind {
		return nil, false
	}
	for _, member := range object.object {
		if member.name == name {
			return member.value, true
		}
	}
	return nil, false
}

func cloneMember(member orderedMember) orderedMember {
	return orderedMember{name: member.name, value: cloneValue(member.value)}
}

func cloneValues(values []*orderedValue) []*orderedValue {
	result := make([]*orderedValue, 0, len(values))
	for _, value := range values {
		result = append(result, cloneValue(value))
	}
	return result
}

func cloneValue(value *orderedValue) *orderedValue {
	cloned := &orderedValue{kind: value.kind, scalar: value.scalar}
	cloned.array = cloneValues(value.array)
	for _, member := range value.object {
		cloned.object = append(cloned.object, cloneMember(member))
	}
	return cloned
}

func sortCanonicalObjectMembers(value *orderedValue, documentRoot bool) {
	switch value.kind {
	case arrayKind:
		for _, child := range value.array {
			sortCanonicalObjectMembers(child, false)
		}
	case objectKind:
		for _, member := range value.object {
			sortCanonicalObjectMembers(member.value, false)
		}
		if documentRoot {
			rank := map[string]int{
				"openapi":           0,
				"info":              1,
				"jsonSchemaDialect": 2,
				"servers":           3,
				"tags":              4,
				"paths":             5,
				"webhooks":          6,
				"components":        7,
				"security":          8,
				"externalDocs":      9,
			}
			sort.SliceStable(value.object, func(left, right int) bool {
				leftRank, leftKnown := rank[value.object[left].name]
				rightRank, rightKnown := rank[value.object[right].name]
				if leftKnown != rightKnown {
					return leftKnown
				}
				if leftKnown && leftRank != rightRank {
					return leftRank < rightRank
				}
				return value.object[left].name < value.object[right].name
			})
			return
		}
		sort.SliceStable(value.object, func(left, right int) bool {
			return value.object[left].name < value.object[right].name
		})
	}
}

func writeOrderedJSON(output *bytes.Buffer, value *orderedValue, indent string, depth int) {
	switch value.kind {
	case nullKind:
		output.WriteString("null")
	case boolKind:
		output.WriteString(strconv.FormatBool(value.scalar.(bool)))
	case numberKind:
		output.WriteString(value.scalar.(json.Number).String())
	case stringKind:
		encoded, _ := json.Marshal(value.scalar.(string))
		output.Write(encoded)
	case arrayKind:
		if len(value.array) == 0 {
			output.WriteString("[]")
			return
		}
		output.WriteString("[\n")
		for index, child := range value.array {
			output.WriteString(strings.Repeat(indent, depth+1))
			writeOrderedJSON(output, child, indent, depth+1)
			if index < len(value.array)-1 {
				output.WriteByte(',')
			}
			output.WriteByte('\n')
		}
		output.WriteString(strings.Repeat(indent, depth))
		output.WriteByte(']')
	case objectKind:
		if len(value.object) == 0 {
			output.WriteString("{}")
			return
		}
		output.WriteString("{\n")
		for index, member := range value.object {
			output.WriteString(strings.Repeat(indent, depth+1))
			encodedName, _ := json.Marshal(member.name)
			output.Write(encodedName)
			output.WriteString(": ")
			writeOrderedJSON(output, member.value, indent, depth+1)
			if index < len(value.object)-1 {
				output.WriteByte(',')
			}
			output.WriteByte('\n')
		}
		output.WriteString(strings.Repeat(indent, depth))
		output.WriteByte('}')
	}
}

func CheckTarget(target string, expected []byte) error {
	directory := filepath.Dir(target)
	directoryFiles, err := os.OpenRoot(directory)
	if err != nil {
		return fmt.Errorf("open target directory: %w", err)
	}
	defer directoryFiles.Close()
	targetName := filepath.Base(target)
	info, err := directoryFiles.Lstat(targetName)
	if err != nil {
		return fmt.Errorf("stat target: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return errors.New("target must be a regular file, not a symlink")
	}
	actual, err := directoryFiles.ReadFile(targetName)
	if err != nil {
		return fmt.Errorf("read target: %w", err)
	}
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("assembled output differs from %s; run the repository generate target", target)
	}
	return nil
}

func WriteTargetAtomically(target string, content []byte) error {
	directory := filepath.Dir(target)
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}
	directoryFiles, err := os.OpenRoot(directory)
	if err != nil {
		return fmt.Errorf("open target directory: %w", err)
	}
	defer directoryFiles.Close()
	targetName := filepath.Base(target)
	mode := fs.FileMode(0o644)
	if info, err := directoryFiles.Lstat(targetName); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return errors.New("target must be a regular file, not a symlink")
		}
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat target: %w", err)
	}
	temporary, err := os.CreateTemp(directory, "."+filepath.Base(target)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary target: %w", err)
	}
	temporaryPath := temporary.Name()
	temporaryName := filepath.Base(temporaryPath)
	removeTemporary := true
	defer func() {
		if removeTemporary {
			_ = directoryFiles.Remove(temporaryName)
		}
	}()
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("chmod temporary target: %w", err)
	}
	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary target: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary target: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary target: %w", err)
	}
	if err := directoryFiles.Rename(temporaryName, targetName); err != nil {
		return fmt.Errorf("replace target: %w", err)
	}
	removeTemporary = false
	directoryHandle, err := directoryFiles.Open(".")
	if err == nil {
		_ = directoryHandle.Sync()
		_ = directoryHandle.Close()
	}
	return nil
}
