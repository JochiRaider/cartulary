package openapicompat

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	releaseRegistrySchemaID = "cartulary.openapi_release_registry.v1"
	changeSetSchemaID       = "cartulary.openapi_release_change_set.v2"
	reportSchemaID          = "cartulary.openapi_compatibility_report.v1"
)

type Classification string

const (
	Additive          Classification = "additive"
	Breaking          Classification = "breaking"
	NonBehavioral     Classification = "non_behavioral"
	SecuritySensitive Classification = "security_sensitive"
)

type Change struct {
	Fingerprint    string         `json:"fingerprint"`
	Classification Classification `json:"classification"`
	Kind           string         `json:"kind"`
	Pointer        string         `json:"pointer"`
	BeforeSHA256   string         `json:"before_sha256,omitempty"`
	AfterSHA256    string         `json:"after_sha256,omitempty"`
}

type Report struct {
	SchemaID        string   `json:"schema_id"`
	BaselineVersion string   `json:"baseline_version"`
	TargetVersion   string   `json:"target_version"`
	BaselineSHA256  string   `json:"baseline_sha256"`
	TargetSHA256    string   `json:"target_sha256"`
	Changes         []Change `json:"changes"`
}

type releaseRegistry struct {
	SchemaID              string          `json:"schema_id"`
	LatestReleasedVersion string          `json:"latest_released_version"`
	Releases              []releaseRecord `json:"releases"`
}

type releaseRecord struct {
	Version          string `json:"version"`
	DocumentPath     string `json:"document_path"`
	SHA256           string `json:"sha256"`
	ByteLength       int    `json:"byte_length"`
	SourceCommit     string `json:"source_commit"`
	PublicationState string `json:"publication_state"`
}

type changeSet struct {
	SchemaID        string              `json:"schema_id"`
	BaselineVersion string              `json:"baseline_version"`
	TargetVersion   string              `json:"target_version"`
	Changes         []approvedChangeRef `json:"changes"`
}

type approvedChangeRef struct {
	Fingerprint    string         `json:"fingerprint"`
	Classification Classification `json:"classification"`
	OwnerID        string         `json:"owner_id"`
	Rationale      string         `json:"rationale"`
}

// CheckRepository compares the current canonical contract with the immutable
// latest released contract and validates any non-empty diff against an exact
// release change set.
func CheckRepository(repositoryRoot string) (Report, error) {
	repository, err := os.OpenRoot(repositoryRoot)
	if err != nil {
		return Report{}, fmt.Errorf("open repository root: %w", err)
	}
	defer repository.Close()

	report, err := reportRepository(repository)
	if err != nil {
		return Report{}, err
	}
	if err := validateVersionPolicy(report); err != nil {
		return Report{}, err
	}
	if len(report.Changes) == 0 {
		return report, nil
	}
	changeSetPath := filepath.Join(
		"contracts",
		"openapi-releases",
		report.TargetVersion+".change-set.json",
	)
	var approved changeSet
	if err := decodeClosedFile(repository, changeSetPath, &approved); err != nil {
		return Report{}, fmt.Errorf("release change set: %w", err)
	}
	if err := validateChangeSet(report, approved); err != nil {
		return Report{}, err
	}
	return report, nil
}

// ReportRepository verifies immutable released input integrity and returns the
// exact semantic diff without applying version or approval policy. It exists
// so a failing gate can emit the fingerprints needed for an owner review.
func ReportRepository(repositoryRoot string) (Report, error) {
	repository, err := os.OpenRoot(repositoryRoot)
	if err != nil {
		return Report{}, fmt.Errorf("open repository root: %w", err)
	}
	defer repository.Close()
	return reportRepository(repository)
}

func reportRepository(repository *os.Root) (Report, error) {
	registryPath := filepath.Join("contracts", "openapi-releases", "index.json")
	var registry releaseRegistry
	if err := decodeClosedFile(repository, registryPath, &registry); err != nil {
		return Report{}, fmt.Errorf("release registry: %w", err)
	}
	if registry.SchemaID != releaseRegistrySchemaID {
		return Report{}, fmt.Errorf("release registry schema_id must be %q", releaseRegistrySchemaID)
	}
	release, err := latestRelease(registry)
	if err != nil {
		return Report{}, err
	}
	baselinePath, err := safeRepositoryPath(release.DocumentPath)
	if err != nil {
		return Report{}, fmt.Errorf("released document: %w", err)
	}
	baselineBytes, err := repository.ReadFile(baselinePath)
	if err != nil {
		return Report{}, fmt.Errorf("read released document: %w", err)
	}
	if len(baselineBytes) != release.ByteLength {
		return Report{}, fmt.Errorf(
			"released document byte length is %d, registry requires %d",
			len(baselineBytes),
			release.ByteLength,
		)
	}
	if digest(baselineBytes) != release.SHA256 {
		return Report{}, errors.New("released document SHA-256 does not match registry")
	}
	currentPath := filepath.Join("contracts", "openapi", "cartulary.openapi.yaml")
	currentBytes, err := repository.ReadFile(currentPath)
	if err != nil {
		return Report{}, fmt.Errorf("read current document: %w", err)
	}
	report, err := Compare(baselineBytes, currentBytes)
	if err != nil {
		return Report{}, err
	}
	if report.BaselineVersion != release.Version {
		return Report{}, fmt.Errorf(
			"released document info.version %q does not match registry version %q",
			report.BaselineVersion,
			release.Version,
		)
	}
	if report.BaselineVersion != registry.LatestReleasedVersion {
		return Report{}, errors.New("release registry latest version is internally inconsistent")
	}
	return report, nil
}

func Compare(baselineBytes, targetBytes []byte) (Report, error) {
	baseline, err := decodeDocument(baselineBytes)
	if err != nil {
		return Report{}, fmt.Errorf("decode baseline OpenAPI: %w", err)
	}
	target, err := decodeDocument(targetBytes)
	if err != nil {
		return Report{}, fmt.Errorf("decode target OpenAPI: %w", err)
	}
	baselineVersion, err := documentVersion(baseline)
	if err != nil {
		return Report{}, fmt.Errorf("baseline: %w", err)
	}
	targetVersion, err := documentVersion(target)
	if err != nil {
		return Report{}, fmt.Errorf("target: %w", err)
	}
	changes := make([]Change, 0)
	diffValue("", baseline, target, &changes)
	sort.Slice(changes, func(left, right int) bool {
		if changes[left].Pointer != changes[right].Pointer {
			return changes[left].Pointer < changes[right].Pointer
		}
		return changes[left].Kind < changes[right].Kind
	})
	return Report{
		SchemaID:        reportSchemaID,
		BaselineVersion: baselineVersion,
		TargetVersion:   targetVersion,
		BaselineSHA256:  digest(baselineBytes),
		TargetSHA256:    digest(targetBytes),
		Changes:         changes,
	}, nil
}

func diffValue(pointer string, before, after any, changes *[]Change) {
	if pointer == "/info/version" {
		return
	}
	beforeObject, beforeIsObject := before.(map[string]any)
	afterObject, afterIsObject := after.(map[string]any)
	if beforeIsObject && afterIsObject {
		keys := make(map[string]struct{}, len(beforeObject)+len(afterObject))
		for key := range beforeObject {
			keys[key] = struct{}{}
		}
		for key := range afterObject {
			keys[key] = struct{}{}
		}
		orderedKeys := make([]string, 0, len(keys))
		for key := range keys {
			orderedKeys = append(orderedKeys, key)
		}
		sort.Strings(orderedKeys)
		for _, key := range orderedKeys {
			childPointer := pointer + "/" + escapePointer(key)
			beforeChild, beforeOK := beforeObject[key]
			afterChild, afterOK := afterObject[key]
			switch {
			case !beforeOK:
				appendChange(changes, "add", childPointer, nil, afterChild)
			case !afterOK:
				appendChange(changes, "remove", childPointer, beforeChild, nil)
			default:
				diffValue(childPointer, beforeChild, afterChild, changes)
			}
		}
		return
	}
	beforeArray, beforeIsArray := before.([]any)
	afterArray, afterIsArray := after.([]any)
	if beforeIsArray && afterIsArray {
		if equalJSON(beforeArray, afterArray) {
			return
		}
		appendChange(changes, "change", pointer, beforeArray, afterArray)
		return
	}
	if !equalJSON(before, after) {
		appendChange(changes, "change", pointer, before, after)
	}
}

func appendChange(changes *[]Change, kind, pointer string, before, after any) {
	classification := classify(kind, pointer, before, after)
	change := Change{
		Classification: classification,
		Kind:           kind,
		Pointer:        pointer,
	}
	if before != nil {
		change.BeforeSHA256 = valueDigest(before)
	}
	if after != nil {
		change.AfterSHA256 = valueDigest(after)
	}
	fingerprintInput := strings.Join([]string{
		string(change.Classification),
		change.Kind,
		change.Pointer,
		change.BeforeSHA256,
		change.AfterSHA256,
	}, "\x00")
	change.Fingerprint = digest([]byte(fingerprintInput))
	*changes = append(*changes, change)
}

func classify(kind, pointer string, before, after any) Classification {
	if isNonBehavioralPointer(pointer) {
		return NonBehavioral
	}
	if strings.Contains(pointer, "/security") ||
		strings.HasPrefix(pointer, "/components/securitySchemes/") {
		return SecuritySensitive
	}
	if strings.HasSuffix(pointer, "/required") {
		return Breaking
	}
	if strings.HasSuffix(pointer, "/enum") {
		beforeValues := scalarSet(before)
		afterValues := scalarSet(after)
		for value := range beforeValues {
			if _, ok := afterValues[value]; !ok {
				return Breaking
			}
		}
		return Additive
	}
	if kind == "remove" {
		return Breaking
	}
	if kind == "add" {
		return Additive
	}
	return Breaking
}

func isNonBehavioralPointer(pointer string) bool {
	segments := strings.Split(pointer, "/")
	if len(segments) == 0 {
		return false
	}
	switch segments[len(segments)-1] {
	case "description", "summary", "title", "example", "examples", "externalDocs", "tags", "x-cartulary-availability":
		return true
	default:
		return false
	}
}

func validateVersionPolicy(report Report) error {
	baseline, err := parseVersion(report.BaselineVersion)
	if err != nil {
		return fmt.Errorf("baseline version: %w", err)
	}
	target, err := parseVersion(report.TargetVersion)
	if err != nil {
		return fmt.Errorf("target version: %w", err)
	}
	if len(report.Changes) == 0 {
		if baseline != target {
			return errors.New("OpenAPI version changed without a semantic contract change")
		}
		return nil
	}
	hasBreaking := false
	hasAdditive := false
	for _, change := range report.Changes {
		switch change.Classification {
		case Breaking, SecuritySensitive:
			hasBreaking = true
		case Additive:
			hasAdditive = true
		case NonBehavioral:
		default:
			return fmt.Errorf("unknown compatibility classification %q", change.Classification)
		}
	}
	switch {
	case hasBreaking && target.major <= baseline.major:
		return errors.New("breaking or security-sensitive OpenAPI change requires a major version increment")
	case hasAdditive && !hasBreaking &&
		target.major == baseline.major &&
		target.minor <= baseline.minor:
		return errors.New("additive OpenAPI change requires a minor or major version increment")
	case !hasBreaking && !hasAdditive &&
		target.major == baseline.major &&
		target.minor == baseline.minor &&
		target.patch <= baseline.patch:
		return errors.New("non-behavioral OpenAPI change requires a patch, minor, or major version increment")
	case compareVersion(target, baseline) <= 0:
		return errors.New("changed OpenAPI contract version must be greater than the released version")
	default:
		return nil
	}
}

func validateChangeSet(report Report, approved changeSet) error {
	if approved.SchemaID != changeSetSchemaID {
		return fmt.Errorf("release change-set schema_id must be %q", changeSetSchemaID)
	}
	if approved.BaselineVersion != report.BaselineVersion ||
		approved.TargetVersion != report.TargetVersion {
		return errors.New("release change-set versions do not match compatibility report")
	}
	actual := make(map[string]Classification, len(report.Changes))
	for _, change := range report.Changes {
		actual[change.Fingerprint] = change.Classification
	}
	seen := make(map[string]struct{}, len(approved.Changes))
	for _, change := range approved.Changes {
		if change.Fingerprint == "" ||
			change.OwnerID == "" ||
			strings.TrimSpace(change.Rationale) == "" {
			return errors.New("release change-set entries require fingerprint, owner, and rationale")
		}
		if _, duplicate := seen[change.Fingerprint]; duplicate {
			return fmt.Errorf("duplicate release change-set fingerprint %q", change.Fingerprint)
		}
		seen[change.Fingerprint] = struct{}{}
		classification, ok := actual[change.Fingerprint]
		if !ok {
			return fmt.Errorf("stale release change-set fingerprint %q", change.Fingerprint)
		}
		if classification != change.Classification {
			return fmt.Errorf(
				"release change-set fingerprint %q classification is %q, actual is %q",
				change.Fingerprint,
				change.Classification,
				classification,
			)
		}
	}
	if len(seen) != len(actual) {
		return fmt.Errorf(
			"release change set covers %d of %d actual changes",
			len(seen),
			len(actual),
		)
	}
	return nil
}

func latestRelease(registry releaseRegistry) (releaseRecord, error) {
	var latest *releaseRecord
	seenVersions := make(map[string]struct{}, len(registry.Releases))
	seenPaths := make(map[string]struct{}, len(registry.Releases))
	for index := range registry.Releases {
		release := &registry.Releases[index]
		if _, err := parseVersion(release.Version); err != nil {
			return releaseRecord{}, fmt.Errorf("release %d version: %w", index+1, err)
		}
		if _, duplicate := seenVersions[release.Version]; duplicate {
			return releaseRecord{}, fmt.Errorf("duplicate released version %q", release.Version)
		}
		seenVersions[release.Version] = struct{}{}
		if _, duplicate := seenPaths[release.DocumentPath]; duplicate {
			return releaseRecord{}, fmt.Errorf("duplicate released document path %q", release.DocumentPath)
		}
		seenPaths[release.DocumentPath] = struct{}{}
		if release.Version == registry.LatestReleasedVersion {
			latest = release
		}
	}
	if latest == nil {
		return releaseRecord{}, fmt.Errorf(
			"latest released version %q has no release record",
			registry.LatestReleasedVersion,
		)
	}
	return *latest, nil
}

func decodeClosedFile(repository *os.Root, path string, target any) error {
	content, err := repository.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.More() {
		return errors.New("file contains trailing JSON values")
	}
	return nil
}

func decodeDocument(content []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	var document map[string]any
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	if document == nil {
		return nil, errors.New("document must be an object")
	}
	return document, nil
}

func documentVersion(document map[string]any) (string, error) {
	info, ok := document["info"].(map[string]any)
	if !ok {
		return "", errors.New("info must be an object")
	}
	version, ok := info["version"].(string)
	if !ok || version == "" {
		return "", errors.New("info.version must be a non-empty string")
	}
	if _, err := parseVersion(version); err != nil {
		return "", fmt.Errorf("info.version: %w", err)
	}
	return version, nil
}

type semanticVersion struct {
	major int
	minor int
	patch int
}

func parseVersion(raw string) (semanticVersion, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return semanticVersion{}, fmt.Errorf("%q must be major.minor.patch", raw)
	}
	values := [3]int{}
	for index, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return semanticVersion{}, fmt.Errorf("%q is not canonical SemVer", raw)
		}
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return semanticVersion{}, fmt.Errorf("%q is not canonical SemVer", raw)
		}
		values[index] = value
	}
	return semanticVersion{major: values[0], minor: values[1], patch: values[2]}, nil
}

func compareVersion(left, right semanticVersion) int {
	switch {
	case left.major != right.major:
		return left.major - right.major
	case left.minor != right.minor:
		return left.minor - right.minor
	default:
		return left.patch - right.patch
	}
}

func safeRepositoryPath(candidate string) (string, error) {
	if candidate == "" || filepath.IsAbs(candidate) || strings.Contains(candidate, "\\") {
		return "", errors.New("path must be a non-empty relative slash path")
	}
	clean := filepath.Clean(filepath.FromSlash(candidate))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("path escapes repository root")
	}
	return clean, nil
}

func scalarSet(value any) map[string]struct{} {
	values, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make(map[string]struct{}, len(values))
	for _, entry := range values {
		encoded, err := json.Marshal(entry)
		if err == nil {
			result[string(encoded)] = struct{}{}
		}
	}
	return result
}

func equalJSON(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func valueDigest(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return digest(encoded)
}

func digest(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func escapePointer(value string) string {
	value = strings.ReplaceAll(value, "~", "~0")
	return strings.ReplaceAll(value, "/", "~1")
}
