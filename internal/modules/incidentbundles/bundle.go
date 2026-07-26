package incidentbundles

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"time"
)

const (
	BundleFormat              = "cartulary.incident_bundle"
	BundleVersion             = 2
	LegacyBundleVersion       = 1
	sourceBoundaryTokenPrefix = "cartulary.source_boundary.v1:"
	tarTypeRegA               = byte(0)
)

var incidentBundleOptionalSectionTokens = map[string]struct{}{
	"reference_packs": {},
	"snapshots":       {},
}

var requiredStructuredFilesV2 = []string{
	"data/incident.json",
	"data/actors.ndjson",
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

var requiredStructuredFilesV1 = func() []string {
	files := append([]string(nil), requiredStructuredFilesV2...)
	files = removeRequiredStructuredFile(files, "data/timeline_time_profiles.ndjson")
	files = removeRequiredStructuredFile(files, "data/timeline_records.ndjson")
	files = removeRequiredStructuredFile(files, "data/timeline_source_provenance.ndjson")
	files = append(files,
		"data/timeline_time_conversion_profiles.ndjson",
		"data/timeline_events.ndjson",
	)
	sort.Strings(files)
	return files
}()

// requiredStructuredFiles is the current export surface. It remains a named
// value because tests and manifest construction intentionally prove that the
// current source-file registry is closed.
var requiredStructuredFiles = requiredStructuredFilesV2

type ManifestInput struct {
	BundleID             string
	IncidentID           string
	IncidentKey          string
	ExportedAt           string
	ReferencePackMode    string
	OptionalSections     []string
	RequiredCapabilities []string
}

type BundleArchive struct {
	Bytes          []byte
	ManifestSHA256 string
	ChecksumLines  []string
	Manifest       BundleManifest
}

type BundleManifest struct {
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
	Files                        []ManifestFile `json:"files"`
}

type ManifestFile struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes"`
	Required  bool   `json:"required"`
}

type VerificationInput struct {
	Bundle []byte
	Limits Limits
}

type VerifiedBundle struct {
	Manifest       BundleManifest
	ManifestSHA256 string
	Files          map[string][]byte
	Checksums      map[string]string
}

type VerificationError struct {
	ReasonCode string
}

func (e *VerificationError) Error() string {
	return "incident bundle verification failed: " + e.ReasonCode
}

func BuildBundleArchive(input ManifestInput, files map[string][]byte) (BundleArchive, error) {
	normalizedFiles := map[string][]byte{}
	for pathName, content := range files {
		if !safeBundlePath(pathName) {
			return BundleArchive{}, fmt.Errorf("unsafe bundle path %q", pathName)
		}
		normalizedFiles[pathName] = append([]byte(nil), content...)
	}
	for _, pathName := range requiredStructuredFiles {
		if _, ok := normalizedFiles[pathName]; !ok {
			return BundleArchive{}, fmt.Errorf("%s is required", pathName)
		}
	}

	manifestFiles := manifestFilesFor(normalizedFiles, false)
	sourceBoundaryBytes, err := json.Marshal(manifestFiles)
	if err != nil {
		return BundleArchive{}, err
	}
	manifest := BundleManifest{
		BundleFormat:                 BundleFormat,
		BundleVersion:                BundleVersion,
		BundleID:                     input.BundleID,
		IncidentID:                   input.IncidentID,
		IncidentKey:                  input.IncidentKey,
		ExportedAt:                   input.ExportedAt,
		SourceChangeSetHighWatermark: sourceBoundaryTokenPrefix + hashHex(sourceBoundaryBytes),
		HistoryMode:                  HistoryModeFull,
		BlobMode:                     BlobModeFull,
		ReferencePackMode:            input.ReferencePackMode,
		OptionalSections:             canonicalStringSet(input.OptionalSections),
		RequiredCapabilities:         canonicalStringSet(input.RequiredCapabilities),
		Files:                        manifestFiles,
	}
	if manifest.ReferencePackMode == "" {
		manifest.ReferencePackMode = ReferencePackModeRefsOnly
	}
	manifestBytes, err := canonicalJSONString(manifest)
	if err != nil {
		return BundleArchive{}, err
	}
	manifestSHA := hashHex(manifestBytes)
	normalizedFiles["manifest.json"] = manifestBytes

	checksumLines := checksumLinesFor(normalizedFiles)
	normalizedFiles["integrity/checksums.sha256"] = []byte(strings.Join(checksumLines, "\n") + "\n")

	archiveBytes, err := zipFiles(normalizedFiles)
	if err != nil {
		return BundleArchive{}, err
	}
	return BundleArchive{
		Bytes:          archiveBytes,
		ManifestSHA256: manifestSHA,
		ChecksumLines:  checksumLines,
		Manifest:       manifest,
	}, nil
}

func VerifyBundle(input VerificationInput) (VerifiedBundle, error) {
	files, err := readBundleArchive(input.Bundle, input.Limits)
	if err != nil {
		return VerifiedBundle{}, err
	}
	manifestBytes, ok := files["manifest.json"]
	if !ok {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "missing_required_file"}
	}
	var manifest BundleManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "malformed_manifest"}
	}
	if manifest.BundleFormat != BundleFormat ||
		(manifest.BundleVersion != BundleVersion && manifest.BundleVersion != LegacyBundleVersion) ||
		manifest.HistoryMode != HistoryModeFull ||
		manifest.BlobMode != BlobModeFull {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "malformed_manifest"}
	}
	if !strings.HasPrefix(manifest.SourceChangeSetHighWatermark, sourceBoundaryTokenPrefix) {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "malformed_manifest"}
	}
	if err := validateManifestVocabularies(manifest); err != nil {
		return VerifiedBundle{}, err
	}
	if signatureBytes, ok := files["integrity/signature.ed25519"]; ok {
		if len(signatureBytes) == 0 || manifest.SigningKeyID == nil || strings.TrimSpace(*manifest.SigningKeyID) == "" {
			return VerifiedBundle{}, &VerificationError{ReasonCode: "signature_mismatch"}
		}
		return VerifiedBundle{}, &VerificationError{ReasonCode: "signature_mismatch"}
	}
	if len(manifest.RequiredCapabilities) > 0 {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "unsupported_required_capability"}
	}
	for _, pathName := range requiredStructuredFilesForVersion(manifest.BundleVersion) {
		if _, ok := files[pathName]; !ok {
			return VerifiedBundle{}, &VerificationError{ReasonCode: "missing_required_file"}
		}
	}
	checksumBytes, ok := files["integrity/checksums.sha256"]
	if !ok {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "checksum_mismatch"}
	}
	checksums, err := parseChecksumInventory(string(checksumBytes))
	if err != nil {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "checksum_mismatch"}
	}
	for pathName, content := range files {
		if pathName == "manifest.json" || strings.HasPrefix(pathName, "integrity/") {
			continue
		}
		want, ok := checksums[pathName]
		if !ok {
			return VerifiedBundle{}, &VerificationError{ReasonCode: "checksum_mismatch"}
		}
		if got := hashHex(content); got != want {
			return VerifiedBundle{}, &VerificationError{ReasonCode: "checksum_mismatch"}
		}
		if strings.HasPrefix(pathName, "blobs/sha256/") {
			if blobSHA := strings.TrimPrefix(pathName, "blobs/sha256/"); blobSHA != want {
				return VerifiedBundle{}, &VerificationError{ReasonCode: "blob_hash_mismatch"}
			}
		}
	}
	for listedPath := range checksums {
		if listedPath == "manifest.json" || strings.HasPrefix(listedPath, "integrity/") {
			return VerifiedBundle{}, &VerificationError{ReasonCode: "checksum_mismatch"}
		}
		if _, ok := files[listedPath]; !ok {
			if strings.HasPrefix(listedPath, "blobs/sha256/") {
				return VerifiedBundle{}, &VerificationError{ReasonCode: "missing_required_blob"}
			}
			return VerifiedBundle{}, &VerificationError{ReasonCode: "missing_required_file"}
		}
	}
	if !canonicalManifestFilesMatch(files, manifest.Files) {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "malformed_manifest"}
	}
	if !bundleOptionalSectionsAllowed(files, manifest) {
		return VerifiedBundle{}, &VerificationError{ReasonCode: "malformed_manifest"}
	}
	return VerifiedBundle{
		Manifest:       manifest,
		ManifestSHA256: hashHex(manifestBytes),
		Files:          files,
		Checksums:      checksums,
	}, nil
}

func requiredStructuredFilesForVersion(version int) []string {
	if version == LegacyBundleVersion {
		return requiredStructuredFilesV1
	}
	return requiredStructuredFilesV2
}

func removeRequiredStructuredFile(files []string, target string) []string {
	for index, file := range files {
		if file == target {
			return append(files[:index], files[index+1:]...)
		}
	}
	return files
}

func canonicalManifestFilesMatch(files map[string][]byte, manifestFiles []ManifestFile) bool {
	expected := manifestFilesFor(files, false)
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

func validateManifestVocabularies(manifest BundleManifest) error {
	if manifest.ReferencePackMode != ReferencePackModeRefsOnly && manifest.ReferencePackMode != ReferencePackModeEmbedded {
		return &VerificationError{ReasonCode: "malformed_manifest"}
	}
	if !canonicalKnownTokenSet(manifest.OptionalSections, incidentBundleOptionalSectionTokens) {
		return &VerificationError{ReasonCode: "malformed_manifest"}
	}
	if !canonicalKnownTokenSet(manifest.RequiredCapabilities, incidentBundleOptionalSectionTokens) {
		return &VerificationError{ReasonCode: "malformed_manifest"}
	}
	if len(manifest.RequiredCapabilities) > 0 {
		return &VerificationError{ReasonCode: "unsupported_required_capability"}
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

func bundleOptionalSectionsAllowed(files map[string][]byte, manifest BundleManifest) bool {
	declared := map[string]struct{}{}
	for _, section := range manifest.OptionalSections {
		declared[section] = struct{}{}
	}
	if manifest.ReferencePackMode == ReferencePackModeEmbedded {
		declared["reference_packs"] = struct{}{}
	}
	for pathName := range files {
		if !strings.HasPrefix(pathName, "ext/") {
			continue
		}
		switch {
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

func readBundleArchive(bundle []byte, limits Limits) (map[string][]byte, error) {
	if len(bundle) == 0 {
		return nil, &VerificationError{ReasonCode: "missing_required_file"}
	}
	if files, err := readZipArchive(bundle, limits); err == nil {
		return files, nil
	} else if isVerificationFailure(err) {
		return nil, err
	}
	if files, err := readTarArchive(bytes.NewReader(bundle), int64(len(bundle)), limits); err == nil {
		return files, nil
	} else if isVerificationFailure(err) {
		return nil, err
	}
	gz, err := gzip.NewReader(bytes.NewReader(bundle))
	if err == nil {
		defer gz.Close()
		if files, tarErr := readTarArchive(gz, int64(len(bundle)), limits); tarErr == nil {
			return files, nil
		} else if isVerificationFailure(tarErr) {
			return nil, tarErr
		}
	}
	return nil, &VerificationError{ReasonCode: "unsupported_member_type"}
}

func readZipArchive(bundle []byte, limits Limits) (map[string][]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	var memberCount int
	var extracted int64
	for _, member := range zr.File {
		memberCount++
		if err := checkMemberCount(memberCount, limits); err != nil {
			return nil, err
		}
		kind, err := classifyZipMember(member)
		if err != nil {
			return nil, err
		}
		if kind == archiveMemberDirectory {
			if member.UncompressedSize64 != 0 {
				return nil, &VerificationError{ReasonCode: "unsupported_member_type"}
			}
			continue
		}
		extracted += int64(member.UncompressedSize64)
		if err := checkExtractedSize(extracted, limits); err != nil {
			return nil, err
		}
		rc, err := member.Open()
		if err != nil {
			return nil, &VerificationError{ReasonCode: "missing_required_file"}
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, &VerificationError{ReasonCode: "missing_required_file"}
		}
		files[member.Name] = data
	}
	if err := checkCompressionRatio(extracted, int64(len(bundle)), limits); err != nil {
		return nil, err
	}
	return files, nil
}

func readTarArchive(reader io.Reader, compressedSize int64, limits Limits) (map[string][]byte, error) {
	tr := tar.NewReader(reader)
	files := map[string][]byte{}
	var memberCount int
	var extracted int64
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		memberCount++
		if err := checkMemberCount(memberCount, limits); err != nil {
			return nil, err
		}
		kind, err := classifyTarMember(header)
		if err != nil {
			return nil, err
		}
		if kind == archiveMemberDirectory {
			if header.Size != 0 {
				return nil, &VerificationError{ReasonCode: "unsupported_member_type"}
			}
			continue
		}
		extracted += header.Size
		if err := checkExtractedSize(extracted, limits); err != nil {
			return nil, err
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, &VerificationError{ReasonCode: "missing_required_file"}
		}
		files[header.Name] = data
	}
	if len(files) == 0 {
		return nil, &VerificationError{ReasonCode: "missing_required_file"}
	}
	if err := checkCompressionRatio(extracted, compressedSize, limits); err != nil {
		return nil, err
	}
	return files, nil
}

type archiveMemberKind int

const (
	archiveMemberFile archiveMemberKind = iota
	archiveMemberDirectory
)

func classifyZipMember(member *zip.File) (archiveMemberKind, error) {
	isDirectory := member.FileInfo().IsDir() || strings.HasSuffix(member.Name, "/")
	if !safeArchiveMemberPath(member.Name, isDirectory) {
		return archiveMemberFile, &VerificationError{ReasonCode: "invalid_member_path"}
	}
	modeType := member.FileInfo().Mode().Type()
	if isDirectory {
		if modeType != 0 && !member.FileInfo().IsDir() {
			return archiveMemberFile, &VerificationError{ReasonCode: "unsupported_member_type"}
		}
		return archiveMemberDirectory, nil
	}
	if modeType != 0 {
		return archiveMemberFile, &VerificationError{ReasonCode: "unsupported_member_type"}
	}
	return archiveMemberFile, nil
}

func classifyTarMember(header *tar.Header) (archiveMemberKind, error) {
	isDirectory := header.Typeflag == tar.TypeDir
	if !safeArchiveMemberPath(header.Name, isDirectory) {
		return archiveMemberFile, &VerificationError{ReasonCode: "invalid_member_path"}
	}
	switch header.Typeflag {
	case tar.TypeDir:
		return archiveMemberDirectory, nil
	case tar.TypeReg, tarTypeRegA:
		return archiveMemberFile, nil
	default:
		return archiveMemberFile, &VerificationError{ReasonCode: "unsupported_member_type"}
	}
}

func checkExtractedSize(extracted int64, limits Limits) error {
	max := limits.IncidentBundles.MaxExtractedBytes
	if max <= 0 {
		max = defaultIncidentBundleMaxExtractedBytes
	}
	if extracted > max {
		return &VerificationError{ReasonCode: "archive_extracted_bytes_exceeded"}
	}
	return nil
}

func checkMemberCount(count int, limits Limits) error {
	max := limits.Archives.MaxMembers
	if max <= 0 {
		max = defaultArchiveMaxMembers
	}
	if int64(count) > max {
		return &VerificationError{ReasonCode: "archive_member_count_exceeded"}
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
		return &VerificationError{ReasonCode: "archive_compression_ratio_exceeded"}
	}
	return nil
}

func manifestFilesFor(files map[string][]byte, includeIntegrity bool) []ManifestFile {
	paths := make([]string, 0, len(files))
	for pathName := range files {
		if pathName == "manifest.json" || (!includeIntegrity && strings.HasPrefix(pathName, "integrity/")) {
			continue
		}
		paths = append(paths, pathName)
	}
	sort.Strings(paths)
	result := make([]ManifestFile, 0, len(paths))
	for _, pathName := range paths {
		result = append(result, ManifestFile{
			Path:      pathName,
			SHA256:    "sha256:" + hashHex(files[pathName]),
			SizeBytes: int64(len(files[pathName])),
			Required:  !strings.HasPrefix(pathName, "ext/"),
		})
	}
	return result
}

func checksumLinesFor(files map[string][]byte) []string {
	paths := make([]string, 0, len(files))
	for pathName := range files {
		if pathName == "manifest.json" || strings.HasPrefix(pathName, "integrity/") {
			continue
		}
		paths = append(paths, pathName)
	}
	sort.Strings(paths)
	lines := make([]string, 0, len(paths))
	for _, pathName := range paths {
		lines = append(lines, hashHex(files[pathName])+"  "+pathName)
	}
	return lines
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

func zipFiles(files map[string][]byte) ([]byte, error) {
	paths := make([]string, 0, len(files))
	for pathName := range files {
		paths = append(paths, pathName)
	}
	sort.Strings(paths)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	modified := time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, pathName := range paths {
		header := &zip.FileHeader{
			Name:     pathName,
			Method:   zip.Store,
			Modified: modified,
		}
		header.SetMode(0o600)
		w, err := zw.CreateHeader(header)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(files[pathName]); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func safeBundlePath(pathName string) bool {
	if pathName == "" || strings.HasPrefix(pathName, "/") || strings.Contains(pathName, "\\") {
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

func isVerificationFailure(err error) bool {
	_, ok := err.(*VerificationError)
	return ok
}
