package incidentbundles

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"unicode"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
)

const (
	bundleFormat              = "cartulary.incident_bundle"
	bundleVersion             = 6
	sourceBoundaryTokenPrefix = "cartulary.source_boundary.v1:"
	tarTypeRegA               = byte(0)
)

var incidentBundleOptionalSectionTokens = map[string]struct{}{
	"reference_packs": {},
	"snapshots":       {},
}

// requiredStructuredFiles is the current export surface. It remains a named
// value because tests and manifest construction intentionally prove that the
// current source-file registry is closed.
var requiredStructuredFiles = []string{
	"data/incident.json",
	"data/actors.ndjson",
	"data/committed_sources.ndjson",
	"data/records.ndjson",
	"data/timeline_time_profiles.ndjson",
	"data/timeline_records.ndjson",
	"data/timeline_source_provenance.ndjson",
	"data/parties.ndjson",
	"data/entity_mentions.ndjson",
	"data/hosts.ndjson",
	"data/identities.ndjson",
	"data/entity_preserved_identifiers.ndjson",
	"data/entity_aliases.ndjson",
	"data/indicators.ndjson",
	"data/indicator_observations.ndjson",
	"data/indicator_state_intervals.ndjson",
	"data/artifacts.ndjson",
	"data/artifact_findings.ndjson",
	"data/artifact_investigative_queries.ndjson",
	"data/artifact_forensic_keywords.ndjson",
	"data/handoff_risk_refs.ndjson",
	"data/task_requests.ndjson",
	"data/decisions.ndjson",
	"data/evidence_records.ndjson",
	"data/evidence_custody_events.ndjson",
	"data/object_blobs.ndjson",
	"data/compromise_assessments.ndjson",
	"data/record_links.ndjson",
	"data/tags.ndjson",
	"data/record_tags.ndjson",
	"data/change_sets.ndjson",
	"data/change_set_mutations.ndjson",
	"data/record_revisions.ndjson",
	"data/saved_views.ndjson",
	"data/reference_pack_refs.json",
}

type manifestInput struct {
	Limits               *Limits
	BundleID             string
	IncidentID           string
	IncidentKey          string
	ExportedAt           string
	ReferencePackMode    string
	OptionalSections     []string
	RequiredCapabilities []string
}

type bundleArchive struct {
	members        incidentportability.MemberReader
	context        context.Context
	ManifestSHA256 string
	ChecksumLines  []string
	Manifest       bundleManifest
}

type bundleManifest struct {
	BundleFormat                 string         `json:"bundle_format"`
	BundleVersion                int            `json:"bundle_version"`
	BundleID                     string         `json:"bundle_id"`
	IncidentID                   string         `json:"incident_id"`
	IncidentKey                  string         `json:"incident_key"`
	ExportedAt                   string         `json:"exported_at"`
	SourceChangeSetHighWatermark string         `json:"source_change_set_high_watermark"`
	HistoryMode                  string         `json:"history_mode"`
	BlobMode                     string         `json:"blob_mode"`
	ReferencePackMode            string         `json:"reference_pack_mode"`
	OptionalSections             []string       `json:"optional_sections"`
	RequiredCapabilities         []string       `json:"required_capabilities"`
	SigningKeyID                 *string        `json:"signing_key_id,omitempty"`
	Files                        []manifestFile `json:"files"`
}

type manifestFile struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes"`
	Required  bool   `json:"required"`
}

var errExtensionCapabilityNotSupported = errors.New("extension_capability_not_supported")

type verifiedBundle struct {
	Manifest       bundleManifest
	ManifestSHA256 string
	Members        incidentportability.MemberReader
	Checksums      map[string]string
}

type verificationError struct {
	ReasonCode     string
	SourceFamilyID string
	InvariantID    string
}

func (e *verificationError) Error() string {
	return "incident bundle verification failed: " + e.ReasonCode
}

func buildBundleArchive(ctx context.Context, input manifestInput, members *memberStore) (bundleArchive, error) {
	for _, required := range requiredStructuredFiles {
		if _, ok := members.members[required]; !ok {
			return bundleArchive{}, fmt.Errorf("%s is required", required)
		}
	}
	manifestFiles := make([]manifestFile, 0, len(members.members))
	checksumLines := make([]string, 0, len(members.members))
	for _, member := range members.Members() {
		if member.Path == "manifest.json" || strings.HasPrefix(member.Path, "integrity/") {
			return bundleArchive{}, &verificationError{ReasonCode: "malformed_manifest"}
		}
		manifestFiles = append(manifestFiles, manifestFile{Path: member.Path, SHA256: "sha256:" + member.SHA256, SizeBytes: member.ByteSize, Required: !strings.HasPrefix(member.Path, "ext/")})
		checksumLines = append(checksumLines, member.SHA256+"  "+member.Path)
	}
	boundary := sha256.New()
	_, _ = io.WriteString(boundary, "[")
	boundaryBytes := int64(2)
	checksumBytes := int64(0)
	for index, file := range manifestFiles {
		encodedFile, err := json.Marshal(file)
		if err != nil {
			return bundleArchive{}, err
		}
		if index > 0 {
			_, _ = io.WriteString(boundary, ",")
			boundaryBytes++
		}
		boundaryBytes += int64(len(encodedFile))
		checksumBytes += int64(64 + 2 + len(file.Path) + 1)
		if boundaryBytes > PortabilityParticipantByteLimit || checksumBytes > PortabilityParticipantByteLimit {
			return bundleArchive{}, &verificationError{ReasonCode: "malformed_manifest"}
		}
		_, _ = boundary.Write(encodedFile)
	}
	_, _ = io.WriteString(boundary, "]")
	manifest := bundleManifest{
		BundleFormat: bundleFormat, BundleVersion: bundleVersion,
		BundleID: input.BundleID, IncidentID: input.IncidentID, IncidentKey: input.IncidentKey, ExportedAt: input.ExportedAt,
		SourceChangeSetHighWatermark: sourceBoundaryTokenPrefix + hex.EncodeToString(boundary.Sum(nil)),
		HistoryMode:                  historyModeFull, BlobMode: blobModeFull, ReferencePackMode: input.ReferencePackMode,
		OptionalSections: canonicalStringSet(input.OptionalSections), RequiredCapabilities: canonicalStringSet(input.RequiredCapabilities), Files: manifestFiles,
	}
	if manifest.ReferencePackMode == "" {
		manifest.ReferencePackMode = referencePackModeRefsOnly
	}
	header := manifest
	header.Files = nil
	encodedHeader, err := json.Marshal(header)
	if err != nil {
		return bundleArchive{}, err
	}
	// Replace the four-byte null Files value with its bounded array and LF.
	if int64(len(encodedHeader))-4+boundaryBytes+1 > PortabilityParticipantByteLimit {
		return bundleArchive{}, &verificationError{ReasonCode: "malformed_manifest"}
	}
	encoded, err := canonicalJSONString(manifest)
	if err != nil {
		return bundleArchive{}, err
	}
	if int64(len(encoded)) > PortabilityParticipantByteLimit {
		return bundleArchive{}, &verificationError{ReasonCode: "malformed_manifest"}
	}
	if _, err := members.WriteMember(ctx, "manifest.json", bytes.NewReader(encoded)); err != nil {
		return bundleArchive{}, err
	}
	checksums := strings.Join(checksumLines, "\n") + "\n"
	if int64(len(checksums)) > PortabilityParticipantByteLimit {
		return bundleArchive{}, &verificationError{ReasonCode: "malformed_manifest"}
	}
	if _, err := members.WriteMember(ctx, "integrity/checksums.sha256", strings.NewReader(checksums)); err != nil {
		return bundleArchive{}, err
	}
	members.Seal()
	return bundleArchive{members: members, context: ctx, ManifestSHA256: hashHex(encoded), ChecksumLines: checksumLines, Manifest: manifest}, nil
}

func verifyBundle(ctx context.Context, members incidentportability.MemberReader) (verifiedBundle, error) {
	files := make(map[string]incidentportability.Member)
	for _, member := range members.Members() {
		if _, duplicate := files[member.Path]; duplicate {
			return verifiedBundle{}, &verificationError{ReasonCode: "malformed_manifest"}
		}
		files[member.Path] = member
	}
	manifestBytes, err := readControlMember(ctx, members, "manifest.json", PortabilityParticipantByteLimit)
	if err != nil {
		return verifiedBundle{}, err
	}
	sanitizedManifestBytes, capabilityActivationRequested, err := sanitizeRequiredCapabilities(manifestBytes)
	if err != nil {
		return verifiedBundle{}, err
	}
	bundleVersion, err := parseBundleVersion(manifestBytes)
	if err != nil {
		return verifiedBundle{}, err
	}
	var manifest bundleManifest
	if err := json.Unmarshal(sanitizedManifestBytes, &manifest); err != nil {
		return verifiedBundle{}, &verificationError{ReasonCode: "malformed_manifest"}
	}
	manifest.BundleVersion = bundleVersion
	if manifest.BundleFormat != bundleFormat ||
		manifest.HistoryMode != historyModeFull ||
		manifest.BlobMode != blobModeFull {
		return verifiedBundle{}, &verificationError{ReasonCode: "malformed_manifest"}
	}
	if !strings.HasPrefix(manifest.SourceChangeSetHighWatermark, sourceBoundaryTokenPrefix) {
		return verifiedBundle{}, &verificationError{ReasonCode: "malformed_manifest"}
	}
	if err := validateManifestVocabularies(manifest); err != nil {
		return verifiedBundle{}, err
	}
	if signatureBytes, ok := files["integrity/signature.ed25519"]; ok {
		if signatureBytes.ByteSize == 0 || manifest.SigningKeyID == nil || strings.TrimSpace(*manifest.SigningKeyID) == "" {
			return verifiedBundle{}, &verificationError{ReasonCode: "signature_mismatch"}
		}
		return verifiedBundle{}, &verificationError{ReasonCode: "signature_mismatch"}
	}
	requiredPaths, err := requiredStructuredFilesForVersion(manifest.BundleVersion)
	if err != nil {
		return verifiedBundle{}, err
	}
	if err := validateClosedBundlePaths(files, requiredPaths); err != nil {
		return verifiedBundle{}, err
	}
	for _, pathName := range requiredPaths {
		if _, ok := files[pathName]; !ok {
			return verifiedBundle{}, &verificationError{ReasonCode: "missing_required_file"}
		}
	}
	_, ok := files["integrity/checksums.sha256"]
	if !ok {
		return verifiedBundle{}, &verificationError{ReasonCode: "checksum_mismatch"}
	}
	checksumBytes, err := readControlMember(ctx, members, "integrity/checksums.sha256", PortabilityParticipantByteLimit)
	if err != nil {
		return verifiedBundle{}, err
	}
	checksums, err := parseChecksumInventory(string(checksumBytes))
	if err != nil {
		return verifiedBundle{}, &verificationError{ReasonCode: "checksum_mismatch"}
	}
	for pathName, content := range files {
		if pathName == "manifest.json" || strings.HasPrefix(pathName, "integrity/") {
			continue
		}
		want, ok := checksums[pathName]
		if !ok {
			return verifiedBundle{}, &verificationError{ReasonCode: "checksum_mismatch"}
		}
		if content.SHA256 != want {
			return verifiedBundle{}, &verificationError{ReasonCode: "checksum_mismatch"}
		}
		if strings.HasPrefix(pathName, "blobs/sha256/") {
			if blobSHA := strings.TrimPrefix(pathName, "blobs/sha256/"); blobSHA != want {
				return verifiedBundle{}, &verificationError{ReasonCode: "blob_hash_mismatch"}
			}
		}
	}
	for listedPath := range checksums {
		if listedPath == "manifest.json" || strings.HasPrefix(listedPath, "integrity/") {
			return verifiedBundle{}, &verificationError{ReasonCode: "checksum_mismatch"}
		}
		if _, ok := files[listedPath]; !ok {
			if strings.HasPrefix(listedPath, "blobs/sha256/") {
				return verifiedBundle{}, &verificationError{ReasonCode: "missing_required_blob"}
			}
			return verifiedBundle{}, &verificationError{ReasonCode: "missing_required_file"}
		}
	}
	if !canonicalManifestFilesMatch(files, manifest.Files) {
		return verifiedBundle{}, &verificationError{ReasonCode: "malformed_manifest"}
	}
	if !bundleOptionalSectionsAllowed(files, manifest) {
		return verifiedBundle{}, &verificationError{ReasonCode: "malformed_manifest"}
	}
	if capabilityActivationRequested {
		return verifiedBundle{}, errExtensionCapabilityNotSupported
	}
	return verifiedBundle{
		Manifest:       manifest,
		ManifestSHA256: hashHex(manifestBytes),
		Members:        members,
		Checksums:      checksums,
	}, nil
}

func sanitizeRequiredCapabilities(manifestBytes []byte) ([]byte, bool, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(manifestBytes, &raw); err != nil || raw == nil {
		return nil, false, &verificationError{ReasonCode: "malformed_manifest"}
	}
	value, ok := raw["required_capabilities"]
	if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return nil, false, &verificationError{ReasonCode: "malformed_manifest"}
	}
	var items []json.RawMessage
	if err := json.Unmarshal(value, &items); err != nil {
		return nil, false, &verificationError{ReasonCode: "malformed_manifest"}
	}
	for _, item := range items {
		var token string
		if err := json.Unmarshal(item, &token); err != nil {
			return nil, false, &verificationError{ReasonCode: "malformed_manifest"}
		}
	}
	raw["required_capabilities"] = json.RawMessage("[]")
	sanitized, err := json.Marshal(raw)
	if err != nil {
		return nil, false, &verificationError{ReasonCode: "malformed_manifest"}
	}
	return sanitized, len(items) > 0, nil
}

func parseBundleVersion(manifestBytes []byte) (int, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(manifestBytes, &raw); err != nil {
		return 0, &verificationError{ReasonCode: "malformed_manifest"}
	}
	value, ok := raw["bundle_version"]
	if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return 0, &verificationError{ReasonCode: "malformed_manifest"}
	}
	var version int
	decoder := json.NewDecoder(bytes.NewReader(value))
	if err := decoder.Decode(&version); err != nil {
		return 0, &verificationError{ReasonCode: "malformed_manifest"}
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return 0, &verificationError{ReasonCode: "malformed_manifest"}
	}
	switch version {
	case bundleVersion:
		return version, nil
	default:
		return 0, &verificationError{ReasonCode: "unsupported_bundle_version"}
	}
}

func requiredStructuredFilesForVersion(version int) ([]string, error) {
	switch version {
	case bundleVersion:
		return requiredStructuredFiles, nil
	default:
		return nil, &verificationError{ReasonCode: "unsupported_bundle_version"}
	}
}

func validateClosedBundlePaths(files map[string]incidentportability.Member, requiredPaths []string) error {
	required := make(map[string]struct{}, len(requiredPaths))
	for _, filePath := range requiredPaths {
		required[filePath] = struct{}{}
	}
	for filePath := range files {
		if !strings.HasPrefix(filePath, "data/") {
			continue
		}
		if _, ok := required[filePath]; !ok {
			return &verificationError{ReasonCode: "malformed_manifest"}
		}
	}
	return nil
}

func canonicalManifestFilesMatch(files map[string]incidentportability.Member, manifestFiles []manifestFile) bool {
	expected := make([]manifestFile, 0, len(files))
	for path, member := range files {
		if path == "manifest.json" || strings.HasPrefix(path, "integrity/") {
			continue
		}
		expected = append(expected, manifestFile{Path: path, SHA256: "sha256:" + member.SHA256, SizeBytes: member.ByteSize, Required: !strings.HasPrefix(path, "ext/")})
	}
	sort.Slice(expected, func(i, j int) bool { return expected[i].Path < expected[j].Path })
	if len(expected) != len(manifestFiles) {
		return false
	}
	for idx := range expected {
		if expected[idx] != manifestFiles[idx] {
			return false
		}
	}
	return true
}

func validateManifestVocabularies(manifest bundleManifest) error {
	if manifest.ReferencePackMode != referencePackModeRefsOnly && manifest.ReferencePackMode != referencePackModeEmbedded {
		return &verificationError{ReasonCode: "malformed_manifest"}
	}
	if !canonicalKnownTokenSet(manifest.OptionalSections, incidentBundleOptionalSectionTokens) {
		return &verificationError{ReasonCode: "malformed_manifest"}
	}
	return nil
}

func canonicalKnownTokenSet(values []string, allowed map[string]struct{}) bool {
	previous := ""
	seen := map[string]struct{}{}
	for _, value := range values {
		if _, ok := allowed[value]; !ok {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		if previous != "" && value < previous {
			return false
		}
		seen[value] = struct{}{}
		previous = value
	}
	return true
}

func bundleOptionalSectionsAllowed(files map[string]incidentportability.Member, manifest bundleManifest) bool {
	declared := map[string]struct{}{}
	for _, section := range manifest.OptionalSections {
		declared[section] = struct{}{}
	}
	if manifest.ReferencePackMode == referencePackModeEmbedded {
		declared["reference_packs"] = struct{}{}
	}
	for pathName := range files {
		if !strings.HasPrefix(pathName, "ext/") {
			continue
		}
		switch {
		case pathName == reference_data.IncidentPackContentPath:
			// Exact requirements are transport metadata, including refs-only exports.
			continue
		case strings.HasPrefix(pathName, "ext/reference_packs/"):
			if _, ok := declared["reference_packs"]; !ok {
				return false
			}
		case strings.HasPrefix(pathName, "ext/snapshots/"):
			if _, ok := declared["snapshots"]; !ok {
				return false
			}
		case strings.HasPrefix(pathName, "ext/extensions/"):
			// Extension payload admission is owned by the immutable
			// portability catalog after archive integrity verification.
			continue
		default:
			return false
		}
	}
	return true
}

type archiveMemberKind int

const (
	archiveMemberFile archiveMemberKind = iota
	archiveMemberDirectory
)

func classifyZipMember(member *zip.File) (archiveMemberKind, error) {
	isDirectory := member.FileInfo().IsDir() || strings.HasSuffix(member.Name, "/")
	if !safeArchiveMemberPath(member.Name, isDirectory) {
		return archiveMemberFile, &verificationError{ReasonCode: "invalid_member_path"}
	}
	modeType := member.FileInfo().Mode().Type()
	if isDirectory {
		if modeType != 0 && !member.FileInfo().IsDir() {
			return archiveMemberFile, &verificationError{ReasonCode: "unsupported_member_type"}
		}
		return archiveMemberDirectory, nil
	}
	if modeType != 0 {
		return archiveMemberFile, &verificationError{ReasonCode: "unsupported_member_type"}
	}
	return archiveMemberFile, nil
}

func classifyTarMember(header *tar.Header) (archiveMemberKind, error) {
	isDirectory := header.Typeflag == tar.TypeDir
	if !safeArchiveMemberPath(header.Name, isDirectory) {
		return archiveMemberFile, &verificationError{ReasonCode: "invalid_member_path"}
	}
	switch header.Typeflag {
	case tar.TypeDir:
		return archiveMemberDirectory, nil
	case tar.TypeReg, tarTypeRegA:
		return archiveMemberFile, nil
	default:
		return archiveMemberFile, &verificationError{ReasonCode: "unsupported_member_type"}
	}
}

func checkMemberCount(count int, limits Limits) error {
	max := limits.Archives.MaxMembers
	if max <= 0 {
		max = defaultArchiveMaxMembers
	}
	if int64(count) > max {
		return &verificationError{ReasonCode: "archive_member_count_exceeded"}
	}
	return nil
}

func checkCompressionRatio(extracted int64, compressed int64, limits Limits) error {
	max := limits.Archives.MaxCompressionRatio
	if max <= 0 {
		max = defaultArchiveMaxCompressionRatio
	}
	if compressed <= 0 {
		compressed = 1
	}
	const maxInt64 = int64(^uint64(0) >> 1)
	if compressed <= maxInt64/max && extracted > compressed*max {
		return &verificationError{ReasonCode: "archive_compression_ratio_exceeded"}
	}
	return nil
}

func parseChecksumInventory(content string) (map[string]string, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, fmt.Errorf("empty checksum inventory")
	}
	checksums := map[string]string{}
	previous := ""
	for _, line := range lines {
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 || len(parts[0]) != sha256.Size*2 {
			return nil, fmt.Errorf("malformed checksum line")
		}
		if _, err := hex.DecodeString(parts[0]); err != nil {
			return nil, err
		}
		if previous != "" && parts[1] < previous {
			return nil, fmt.Errorf("checksum inventory not sorted")
		}
		previous = parts[1]
		checksums[parts[1]] = parts[0]
	}
	return checksums, nil
}

func (archive bundleArchive) Write(destination io.Writer) error {
	return writeZIPMembers(archive.context, destination, archive.members)
}

func safeBundlePath(pathName string) bool {
	if validateStorageReference(pathName) != nil || strings.IndexFunc(pathName, unicode.IsControl) >= 0 || strings.HasPrefix(pathName, "/") || strings.Contains(pathName, "\\") {
		return false
	}
	clean := path.Clean(pathName)
	if clean == "." || clean != pathName {
		return false
	}
	for _, segment := range strings.Split(pathName, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func safeArchiveMemberPath(pathName string, isDirectory bool) bool {
	if !isDirectory {
		return safeBundlePath(pathName)
	}
	return safeBundlePath(strings.TrimSuffix(pathName, "/"))
}

func canonicalStringSet(values []string) []string {
	seen := map[string]struct{}{}
	for _, value := range values {
		if value != "" {
			seen[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
