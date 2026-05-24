package reference_data

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/config"
)

type VerificationInput struct {
	Bundle          []byte
	ContentType     string
	ArchiveLimits   config.ArchiveLimits
	ReferenceLimits config.ReferencePackLimits
}

type VerificationResult struct {
	PackKey             string
	PackKind            string
	PackVersion         string
	SourceIdentifier    *string
	ManifestSHA256      string
	PayloadSHA256       string
	PackContractVersion string
	VerificationMethod  string
	SignerKeyID         *string
	BundleSHA256        string
	Metadata            map[string]any
}

type VerificationError struct {
	ReasonCode string
}

func (e *VerificationError) Error() string {
	return "reference pack verification failed: " + e.ReasonCode
}

type bundleManifest struct {
	PackKey             string             `json:"pack_key"`
	PackKind            string             `json:"pack_kind"`
	PackVersion         string             `json:"pack_version"`
	SourceIdentifier    *string            `json:"source_identifier"`
	PackContractVersion string             `json:"pack_contract_version"`
	VerificationMethod  string             `json:"verification_method"`
	SignerKeyID         *string            `json:"signer_key_id"`
	Payloads            []manifestPayload  `json:"payloads"`
	Signature           *manifestSignature `json:"signature"`
}

type manifestPayload struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type manifestSignature struct {
	PayloadSHA256 string `json:"payload_sha256"`
}

type archiveMember struct {
	Path           string
	Data           []byte
	CompressedSize int64
}

func VerifyBundle(input VerificationInput) (VerificationResult, error) {
	if len(input.Bundle) == 0 {
		return VerificationResult{}, verificationFailure("payload_missing")
	}
	members, err := readArchiveMembers(input)
	if err != nil {
		var verificationErr *VerificationError
		if errors.As(err, &verificationErr) {
			return VerificationResult{}, err
		}
		return VerificationResult{}, verificationFailure("payload_missing")
	}
	if len(members) == 0 {
		return VerificationResult{}, verificationFailure("payload_missing")
	}
	if err := enforceArchiveLimits(members, int64(len(input.Bundle)), input); err != nil {
		return VerificationResult{}, err
	}

	memberByPath := map[string]archiveMember{}
	for _, member := range members {
		clean, ok := cleanArchivePath(member.Path)
		if !ok {
			return VerificationResult{}, verificationFailure("path_traversal")
		}
		if !contentAllowed(clean) && clean != "manifest.json" {
			return VerificationResult{}, verificationFailure("disallowed_content")
		}
		member.Path = clean
		memberByPath[clean] = member
	}
	manifestMember, ok := memberByPath["manifest.json"]
	if !ok {
		return VerificationResult{}, verificationFailure("missing_integrity_metadata")
	}
	manifestSHA := hashHex(manifestMember.Data)
	manifest, err := decodeManifest(manifestMember.Data)
	if err != nil {
		return VerificationResult{}, err
	}
	if manifest.PackContractVersion != PackContractVersionV1 {
		return VerificationResult{}, verificationFailure("contract_incompatible")
	}
	if len(manifest.Payloads) == 0 {
		return VerificationResult{}, verificationFailure("missing_integrity_metadata")
	}

	payloadDigests := make([]manifestPayload, 0, len(manifest.Payloads))
	seenPayloads := map[string]struct{}{}
	for _, payload := range manifest.Payloads {
		clean, ok := cleanArchivePath(payload.Path)
		if !ok || clean == "manifest.json" {
			return VerificationResult{}, verificationFailure("path_traversal")
		}
		if !contentAllowed(clean) {
			return VerificationResult{}, verificationFailure("disallowed_content")
		}
		if strings.TrimSpace(payload.SHA256) == "" || !isHexSHA256(payload.SHA256) {
			return VerificationResult{}, verificationFailure("missing_integrity_metadata")
		}
		member, ok := memberByPath[clean]
		if !ok {
			return VerificationResult{}, verificationFailure("payload_missing")
		}
		actual := hashHex(member.Data)
		if !strings.EqualFold(actual, payload.SHA256) {
			return VerificationResult{}, verificationFailure("checksum_mismatch")
		}
		if _, ok := seenPayloads[clean]; ok {
			return VerificationResult{}, verificationFailure("missing_integrity_metadata")
		}
		seenPayloads[clean] = struct{}{}
		payloadDigests = append(payloadDigests, manifestPayload{Path: clean, SHA256: strings.ToLower(payload.SHA256)})
	}
	payloadSHA, err := canonicalPayloadSHA256(payloadDigests)
	if err != nil {
		return VerificationResult{}, err
	}
	if manifest.VerificationMethod == "signed_manifest_v1" {
		if manifest.Signature == nil || manifest.Signature.PayloadSHA256 == "" {
			return VerificationResult{}, verificationFailure("missing_integrity_metadata")
		}
		if !strings.EqualFold(manifest.Signature.PayloadSHA256, payloadSHA) {
			return VerificationResult{}, verificationFailure("signature_mismatch")
		}
	}
	verificationMethod := manifest.VerificationMethod
	if strings.TrimSpace(verificationMethod) == "" {
		verificationMethod = "manifest_sha256_v1"
	}

	return VerificationResult{
		PackKey:             manifest.PackKey,
		PackKind:            manifest.PackKind,
		PackVersion:         manifest.PackVersion,
		SourceIdentifier:    emptyStringAsNil(manifest.SourceIdentifier),
		ManifestSHA256:      manifestSHA,
		PayloadSHA256:       payloadSHA,
		PackContractVersion: manifest.PackContractVersion,
		VerificationMethod:  verificationMethod,
		SignerKeyID:         emptyStringAsNil(manifest.SignerKeyID),
		BundleSHA256:        hashHex(input.Bundle),
		Metadata: map[string]any{
			"payload_count": len(payloadDigests),
		},
	}, nil
}

func decodeManifest(data []byte) (bundleManifest, error) {
	var manifest bundleManifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&manifest); err != nil {
		return bundleManifest{}, verificationFailure("missing_integrity_metadata")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return bundleManifest{}, verificationFailure("missing_integrity_metadata")
	}
	if strings.TrimSpace(manifest.PackKey) == "" ||
		strings.TrimSpace(manifest.PackKind) == "" ||
		strings.TrimSpace(manifest.PackVersion) == "" ||
		strings.TrimSpace(manifest.PackContractVersion) == "" {
		return bundleManifest{}, verificationFailure("missing_integrity_metadata")
	}
	return manifest, nil
}

func readArchiveMembers(input VerificationInput) ([]archiveMember, error) {
	switch input.ContentType {
	case MediaTypeZip:
		return readZip(input.Bundle)
	case MediaTypeTar:
		return readTar(input.Bundle)
	case MediaTypeGzip, MediaTypeXGzip:
		return readGzipTar(input.Bundle)
	case MediaTypeOctetStream:
		return sniffArchive(input.Bundle)
	default:
		return nil, verificationFailure("payload_missing")
	}
}

func sniffArchive(data []byte) ([]archiveMember, error) {
	if len(data) >= 4 && bytes.Equal(data[:4], []byte{'P', 'K', 0x03, 0x04}) {
		return readZip(data)
	}
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		return readGzipTar(data)
	}
	if members, err := readTar(data); err == nil {
		return members, nil
	}
	return nil, verificationFailure("payload_missing")
}

func readZip(data []byte) ([]archiveMember, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, verificationFailure("payload_missing")
	}
	members := make([]archiveMember, 0, len(reader.File))
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return nil, verificationFailure("payload_missing")
		}
		payload, readErr := io.ReadAll(rc)
		closeErr := rc.Close()
		if readErr != nil || closeErr != nil {
			return nil, verificationFailure("payload_missing")
		}
		members = append(members, archiveMember{
			Path:           file.Name,
			Data:           payload,
			CompressedSize: int64(file.CompressedSize64),
		})
	}
	return members, nil
}

func readTar(data []byte) ([]archiveMember, error) {
	reader := tar.NewReader(bytes.NewReader(data))
	return readTarMembers(reader, int64(len(data)))
}

func readGzipTar(data []byte) ([]archiveMember, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, verificationFailure("payload_missing")
	}
	defer func() { _ = gzipReader.Close() }()
	return readTarMembers(tar.NewReader(gzipReader), int64(len(data)))
}

func readTarMembers(reader *tar.Reader, compressedSize int64) ([]archiveMember, error) {
	var members []archiveMember
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, verificationFailure("payload_missing")
		}
		if header.FileInfo().IsDir() {
			continue
		}
		if header.Typeflag != tar.TypeReg {
			return nil, verificationFailure("disallowed_content")
		}
		payload, err := io.ReadAll(reader)
		if err != nil {
			return nil, verificationFailure("payload_missing")
		}
		members = append(members, archiveMember{
			Path:           header.Name,
			Data:           payload,
			CompressedSize: compressedSize,
		})
	}
	return members, nil
}

func enforceArchiveLimits(members []archiveMember, bundleSize int64, input VerificationInput) error {
	maxMembers := input.ArchiveLimits.MaxMembers
	if maxMembers > 0 && int64(len(members)) > maxMembers {
		return verificationFailure("archive_member_count_exceeded")
	}
	maxExtracted := input.ReferenceLimits.MaxExtractedBytes
	if maxExtracted <= 0 {
		maxExtracted = input.ArchiveLimits.DefaultMaxExtractedBytes
	}
	var extracted int64
	var compressed int64
	for _, member := range members {
		extracted += int64(len(member.Data))
		compressed += member.CompressedSize
	}
	if maxExtracted > 0 && extracted > maxExtracted {
		return verificationFailure("archive_extracted_bytes_exceeded")
	}
	if compressed <= 0 {
		compressed = bundleSize
	}
	maxRatio := input.ArchiveLimits.MaxCompressionRatio
	if maxRatio > 0 && compressed > 0 && extracted > compressed*maxRatio {
		return verificationFailure("archive_compression_ratio_exceeded")
	}
	return nil
}

func cleanArchivePath(raw string) (string, bool) {
	if raw == "" || strings.HasPrefix(raw, "/") || strings.Contains(raw, "\\") {
		return "", false
	}
	clean := path.Clean(raw)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", false
	}
	return clean, true
}

func contentAllowed(memberPath string) bool {
	ext := strings.ToLower(path.Ext(memberPath))
	return slices.Contains([]string{".json", ".csv", ".tsv", ".txt", ".md", ".yaml", ".yml"}, ext)
}

func canonicalPayloadSHA256(payloads []manifestPayload) (string, error) {
	sort.Slice(payloads, func(i int, j int) bool {
		return payloads[i].Path < payloads[j].Path
	})
	payload, err := json.Marshal(payloads)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func isHexSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func emptyStringAsNil(value *string) *string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	cloned := *value
	return &cloned
}

func verificationFailure(reasonCode string) error {
	if !isValidVerificationFailureReason(reasonCode) {
		panic(fmt.Sprintf("invalid reference pack verification reason %q", reasonCode))
	}
	return &VerificationError{ReasonCode: reasonCode}
}
