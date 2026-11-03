package extensions

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
)

// admittedPackage is a private snapshot of the exact build inventory. Later
// catalog readers cannot request unlisted files or observe a changing adapter.
type admittedPackage struct {
	validateFinding      func(any) error
	validationConditions map[string]ValidationCondition
	requiredSchemas      map[string][]string
	artifacts            map[string]PackagedArtifact
	objects              map[string]map[string]any
	paths                []string
	integrity            string
}

func (p *admittedPackage) Artifact(path string) (PackagedArtifact, bool) {
	artifact, ok := p.artifacts[path]
	artifact.JSON = append([]byte(nil), artifact.JSON...)
	return artifact, ok
}
func (p *admittedPackage) ArtifactPaths() []string { return append([]string(nil), p.paths...) }
func (p *admittedPackage) IntegritySHA256() string { return p.integrity }

func registryLimit(reason string, limit, actual int) error {
	cause := validationFailure(Finding{Code: "extension_registry_limit_exceeded", Phase: "registry_generation", Actual: reason})
	return startupFailure(cause, StartupFinding{Path: "$", ReasonCode: "extension_registry_limit_exceeded", Message: "An extension registry limit was exceeded.", Details: map[string]any{"phase": "registry_generation", "limit": limit, "actual": actual}})
}
func registryInvalid(reason string) error {
	return invalidArtifact("registry_integrity", errors.New(reason))
}

func packagedArtifactByteLimit(path string) int {
	switch path {
	case generatedExtensionsRoot + "profile-registry.json", generatedExtensionsRoot + "owner-input-registry.json":
		return registryfacts.RegistryByteLimit
	case generatedExtensionsRoot + "registry-integrity.json":
		return 8 << 20
	case generatedExtensionsRoot + "validation-condition-registry.json", generatedExtensionsRoot + "client-asset-set-manifest.json":
		return 16 << 20
	}
	if strings.HasPrefix(path, generatedExtensionsRoot+"descriptors/") {
		return 256 << 10
	}
	if strings.HasPrefix(path, generatedExtensionsRoot) {
		for _, prefix := range []string{"implementation-bindings/", "job-contracts/", "worker-runtime-contracts/"} {
			if strings.HasPrefix(path, generatedExtensionsRoot+prefix) {
				return 1 << 20
			}
		}
		if path == generatedExtensionsRoot+"client-support-registry.json" {
			return 1 << 20
		}
		return 64 << 20
	}
	if strings.HasPrefix(path, "contracts/extensions/fragments/") {
		return registryfacts.OwnerFragmentByteLimit
	}
	if strings.HasPrefix(path, "contracts/extensions/profiles/") || strings.HasPrefix(path, "contracts/extensions/build/") {
		return 1 << 20
	}
	return 64 << 20
}

func validSourceRef(value string) bool {
	if len(value) < 1 || len(value) > 512 || !utf8.ValidString(value) || strings.ContainsAny(value, "\\\x00") || strings.HasPrefix(value, "/") {
		return false
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}
func lowerDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, c := range value {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

var packagedSchemaIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,159}$`)

func asciiIdentity(value string, max int) bool {
	if len(value) == 0 || len(value) > max {
		return false
	}
	for _, c := range value {
		if c < 33 || c > 126 {
			return false
		}
	}
	return true
}

// admitRegistryPackage validates the integrity contract against the explicit
// build inventory. Source provenance is syntax-only here: this package never
// resolves a source path and works without a repository or development files.
func admitRegistryPackage(source ArtifactSource) (*admittedPackage, error) {
	rootPath := generatedExtensionsRoot + "registry-integrity.json"
	root, artifact, err := readArtifactObject(source, rootPath)
	if err != nil {
		return nil, err
	}
	expected := source.IntegritySHA256()
	if !lowerDigest(expected) || artifact.SHA256 != expected {
		return nil, registryInvalid("build_bound_integrity_mismatch")
	}
	if requireExactKeys(root, "schema_id", "canonicalization_algorithm_id", "dependency_snapshot_sha256", "owner_input_registry_sha256", "registry_schema_id", "registry_sha256", "descriptor_digests", "owner_fragment_digests", "implementation_binding_digests", "supporting_contract_artifact_digests", "generator_id", "generator_sources", "generated_schema_digests") != nil ||
		root["schema_id"] != "cartulary.extension_registry_integrity.v2" || root["canonicalization_algorithm_id"] != "extension_registry_canonical_json_v1" || root["registry_schema_id"] != "cartulary.extension_profile_registry.v1" || !asciiIdentity(stringValue(root["generator_id"]), 160) {
		return nil, registryInvalid("integrity_contract_invalid")
	}
	for _, key := range []string{"dependency_snapshot_sha256", "owner_input_registry_sha256", "registry_sha256"} {
		if !lowerDigest(stringValue(root[key])) {
			return nil, registryInvalid("integrity_digest_invalid")
		}
	}
	sources, ok := objectSlice(root["generator_sources"])
	if !ok || len(sources) == 0 {
		return nil, registryInvalid("generator_sources_missing")
	}
	if len(sources) > 1024 {
		return nil, registryLimit("generator_sources", 1024, len(sources))
	}
	previous := ""
	for _, row := range sources {
		ref := stringValue(row["source_ref"])
		if requireExactKeys(row, "source_ref", "source_sha256") != nil || !validSourceRef(ref) || ref <= previous || !lowerDigest(stringValue(row["source_sha256"])) {
			return nil, registryInvalid("generator_source_invalid")
		}
		previous = ref
	}
	p := &admittedPackage{artifacts: map[string]PackagedArtifact{}, objects: map[string]map[string]any{}, integrity: expected}
	paths := source.ArtifactPaths()
	// Support entries have the largest declared set. Fixed roots, descriptors,
	// bindings, fragments and generated schemas are separately checked below.
	if len(paths) > 65536+65536+512+512+4 {
		return nil, registryLimit("packaged_artifacts", 65536+65536+512+512+4, len(paths))
	}
	previous = ""
	for _, path := range paths {
		if !validSourceRef(path) || !strings.HasPrefix(path, "contracts/extensions/") || !strings.HasSuffix(path, ".json") || path <= previous {
			return nil, registryInvalid("package_inventory_invalid")
		}
		previous = path
		object, current, readErr := readArtifactObject(source, path)
		if readErr != nil {
			return nil, readErr
		}
		if path == rootPath && (current.SHA256 != expected || !equalCanonicalObjects(object, root)) {
			return nil, registryInvalid("package_changed_during_admission")
		}
		p.artifacts[path] = current
		p.objects[path] = object
	}
	p.paths = append([]string(nil), paths...)
	if _, ok := p.artifacts[rootPath]; !ok {
		return nil, registryInvalid("integrity_missing_from_inventory")
	}
	expectedSets := map[string]map[string]string{
		"descriptor_digests": {}, "owner_fragment_digests": {}, "implementation_binding_digests": {}, "supporting_contract_artifact_digests": {}, "generated_schema_digests": {},
	}
	supportSchemas := map[string]string{}
	fixed := map[string]string{"dependency-snapshot.json": "dependency_snapshot_sha256", "owner-input-registry.json": "owner_input_registry_sha256", "profile-registry.json": "registry_sha256"}
	for _, path := range paths {
		object, current := p.objects[path], p.artifacts[path]
		relative := strings.TrimPrefix(path, generatedExtensionsRoot)
		if path == rootPath {
			continue
		}
		if field, ok := fixed[relative]; ok {
			if root[field] != current.SHA256 {
				return nil, registryInvalid("root_artifact_digest_mismatch")
			}
			delete(fixed, relative)
			continue
		}
		category, identity := "supporting_contract_artifact_digests", ""
		switch {
		case strings.HasPrefix(path, generatedExtensionsRoot+"descriptors/"):
			category, identity = "descriptor_digests", stringValue(object["profile_id"])
			if path != generatedExtensionsRoot+"descriptors/"+identity+".json" || object["schema_id"] != "cartulary.extension_profile_descriptor.v3" {
				return nil, registryInvalid("descriptor_identity_invalid")
			}
		case strings.HasPrefix(path, generatedExtensionsRoot+"implementation-bindings/"):
			category, identity = "implementation_binding_digests", stringValue(object["profile_id"])
			if path != generatedExtensionsRoot+"implementation-bindings/"+identity+".json" || object["schema_id"] != "cartulary.extension_implementation_binding.v2" {
				return nil, registryInvalid("binding_identity_invalid")
			}
		case strings.HasPrefix(path, generatedExtensionsRoot+"schemas/"):
			category, identity = "generated_schema_digests", stringValue(object["$id"])
			if path != generatedExtensionsRoot+"schemas/"+identity+".schema.json" || validatePackagedSchema(object) != nil {
				return nil, registryInvalid("generated_schema_invalid")
			}
		case object["schema_id"] == "cartulary.extension_owner_fragment.v3":
			category, identity = "owner_fragment_digests", stringValue(object["owner_fragment_id"])
		default:
			if strings.HasPrefix(path, generatedExtensionsRoot) {
				identity = strings.TrimSuffix(relative, ".json")
			} else {
				identity = strings.TrimSuffix(strings.ReplaceAll(strings.TrimPrefix(path, "contracts/extensions/"), "/", "."), ".json")
			}
			schemaID := stringValue(object["schema_id"])
			if !packagedSchemaIDPattern.MatchString(schemaID) {
				return nil, registryInvalid("support_schema_invalid")
			}
			supportSchemas[identity] = schemaID
		}
		if identity == "" || expectedSets[category][identity] != "" {
			return nil, registryInvalid("packaged_identity_duplicate")
		}
		expectedSets[category][identity] = current.SHA256
	}
	if len(fixed) != 0 {
		return nil, registryInvalid("required_root_artifact_missing")
	}
	for _, set := range []struct {
		name, identity, digest string
		maximum                int
	}{
		{"descriptor_digests", "profile_id", "descriptor_sha256", 256}, {"owner_fragment_digests", "owner_fragment_id", "owner_fragment_sha256", 512},
		{"implementation_binding_digests", "profile_id", "binding_sha256", 256}, {"supporting_contract_artifact_digests", "artifact_id", "artifact_sha256", 65536},
		{"generated_schema_digests", "schema_id", "schema_sha256", 65536},
	} {
		rows, ok := objectSlice(root[set.name])
		if !ok {
			return nil, registryInvalid("integrity_set_not_array")
		}
		if len(rows) > set.maximum {
			return nil, registryLimit(set.name, set.maximum, len(rows))
		}
		if len(rows) != len(expectedSets[set.name]) {
			return nil, registryInvalid("integrity_set_incomplete_or_extra")
		}
		previous = ""
		for _, row := range rows {
			keys := []string{set.identity, set.digest}
			if set.name == "supporting_contract_artifact_digests" {
				keys = append(keys, "schema_id")
			}
			id, digest := stringValue(row[set.identity]), stringValue(row[set.digest])
			if requireExactKeys(row, keys...) != nil || id <= previous || !lowerDigest(digest) || expectedSets[set.name][id] != digest {
				return nil, registryInvalid("integrity_set_identity_or_digest_invalid")
			}
			if set.name == "supporting_contract_artifact_digests" && row["schema_id"] != supportSchemas[id] {
				return nil, registryInvalid("supporting_schema_mismatch")
			}
			previous = id
		}
	}
	if err := validateRegistrySchemaBundle(p); err != nil {
		return nil, err
	}
	if err := p.validateOwnerInputs(); err != nil {
		return nil, err
	}
	return p, nil
}

func validatePackagedSchema(object map[string]any) error {
	if !packagedSchemaIDPattern.MatchString(stringValue(object["$id"])) || object["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		return registryInvalid("schema_identity_invalid")
	}
	return nil
}
