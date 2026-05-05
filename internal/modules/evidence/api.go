package evidence

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
)

const (
	blobCreateRouteKey = "object_blobs.create"
	blobAttachRouteKey = "evidence.attach_blob"

	evidenceViewSchemaID = "cartulary.view.evidence.v1"
)

var sha256HexPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

var blobCreateServerManagedFields = map[string]struct{}{
	"object_blob_id":           {},
	"record_id":                {},
	"evidence_record_id":       {},
	"storage_key":              {},
	"storage_backend":          {},
	"bucket":                   {},
	"key":                      {},
	"upload_state":             {},
	"target_expires_at":        {},
	"pending_expires_at":       {},
	"upload_target":            {},
	"accepted_contract":        {},
	"observed_size":            {},
	"observed_content_type":    {},
	"observed_sha256_hex":      {},
	"finalized_at":             {},
	"terminal_reason":          {},
	"failed_at":                {},
	"finalize_attempt_count":   {},
	"cleanup_due_at":           {},
	"cleaned_up_at":            {},
	"created_at":               {},
	"updated_at":               {},
	"created_by_user_id":       {},
	"preview_kind":             {},
	"preview_intent":           {},
	"download_intent":          {},
	"release_state":            {},
	"workflow_state":           {},
	"evidence_lifecycle_state": {},
}

type BlobCreateRequest struct {
	IncidentID       uuid.UUID
	ClientTxnID      string
	ByteSize         int64
	FilenameHint     *string
	ContentTypeHint  *string
	SHA256Hex        *string
	AcceptedContract map[string]any
}

type AttachBlobRequest struct {
	ObjectBlobID   uuid.UUID
	BaseRowVersion int64
	ClientTxnID    string
}

func DecodeBlobCreateRequest(reader io.Reader, maxByteSize int64) (BlobCreateRequest, *auth.APIError) {
	raw, apiErr := decodeStrictObject(reader, "invalid_blob_create_request")
	if apiErr != nil {
		return BlobCreateRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"incident_id": {}, "client_txn_id": {}, "byte_size": {},
		"filename_hint": {}, "content_type_hint": {}, "sha256_hex": {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			if _, managed := blobCreateServerManagedFields[key]; managed {
				return BlobCreateRequest{}, invalidBlobCreate(key, "server_managed_field")
			}
			return BlobCreateRequest{}, invalidBlobCreate(key, "unknown_field")
		}
	}
	incidentID, apiErr := requiredBlobUUID(raw, "incident_id")
	if apiErr != nil {
		return BlobCreateRequest{}, apiErr
	}
	clientTxnID, apiErr := requiredBlobString(raw, "client_txn_id")
	if apiErr != nil {
		return BlobCreateRequest{}, apiErr
	}
	byteSize, apiErr := requiredBlobInt64(raw, "byte_size")
	if apiErr != nil {
		return BlobCreateRequest{}, apiErr
	}
	if byteSize < 0 {
		return BlobCreateRequest{}, invalidBlobCreate("byte_size", "invalid_byte_size")
	}
	if byteSize > maxByteSize {
		return BlobCreateRequest{}, &auth.APIError{
			Status: http.StatusRequestEntityTooLarge,
			Code:   "blob_create_rejected",
			Details: map[string]any{
				"reason_code":            "byte_size_exceeds_limit",
				"requested_byte_size":    byteSize,
				"configured_limit_bytes": maxByteSize,
				"field":                  "byte_size",
			},
		}
	}
	filenameHint, apiErr := optionalBlobTrimmedString(raw, "filename_hint")
	if apiErr != nil {
		return BlobCreateRequest{}, apiErr
	}
	contentTypeHint, apiErr := optionalBlobTrimmedString(raw, "content_type_hint")
	if apiErr != nil {
		return BlobCreateRequest{}, apiErr
	}
	sha256Hex, apiErr := optionalBlobTrimmedString(raw, "sha256_hex")
	if apiErr != nil {
		return BlobCreateRequest{}, apiErr
	}
	if sha256Hex != nil && !sha256HexPattern.MatchString(*sha256Hex) {
		return BlobCreateRequest{}, invalidBlobCreate("sha256_hex", "invalid_sha256_hex")
	}
	accepted := map[string]any{
		"incident_id":       incidentID.String(),
		"byte_size":         byteSize,
		"filename_hint":     nullableString(filenameHint),
		"content_type_hint": nullableString(contentTypeHint),
		"sha256_hex":        nullableString(sha256Hex),
	}
	return BlobCreateRequest{
		IncidentID: incidentID, ClientTxnID: clientTxnID, ByteSize: byteSize,
		FilenameHint: filenameHint, ContentTypeHint: contentTypeHint, SHA256Hex: sha256Hex,
		AcceptedContract: accepted,
	}, nil
}

func DecodeAttachBlobRequest(reader io.Reader) (AttachBlobRequest, *auth.APIError) {
	raw, apiErr := decodeStrictObject(reader, "invalid_mutation_payload")
	if apiErr != nil {
		return AttachBlobRequest{}, apiErr
	}
	allowed := map[string]struct{}{"object_blob_id": {}, "base_row_version": {}, "client_txn_id": {}}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return AttachBlobRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	objectBlobID, apiErr := requiredUUID(raw, "object_blob_id", "invalid_mutation_payload")
	if apiErr != nil {
		return AttachBlobRequest{}, apiErr
	}
	baseRowVersion, apiErr := requiredInt64(raw, "base_row_version", "invalid_mutation_payload")
	if apiErr != nil {
		return AttachBlobRequest{}, apiErr
	}
	if baseRowVersion < 1 {
		return AttachBlobRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	clientTxnID, apiErr := requiredString(raw, "client_txn_id", "invalid_mutation_payload")
	if apiErr != nil {
		return AttachBlobRequest{}, apiErr
	}
	return AttachBlobRequest{ObjectBlobID: objectBlobID, BaseRowVersion: baseRowVersion, ClientTxnID: clientTxnID}, nil
}

func DecodeHandleIssueRequest(reader io.Reader) *auth.APIError {
	raw, apiErr := decodeStrictObject(reader, "invalid_evidence_handle_request")
	if apiErr != nil {
		return apiErr
	}
	for key := range raw {
		return invalidEvidenceHandleRequest(key, "unknown_field")
	}
	return nil
}

func BlobCreateRequestHash(request BlobCreateRequest) []byte {
	return hashRequestPayload(map[string]any{
		"byte_size":         request.ByteSize,
		"filename_hint":     nullableString(request.FilenameHint),
		"content_type_hint": nullableString(request.ContentTypeHint),
		"sha256_hex":        nullableString(request.SHA256Hex),
	})
}

func AttachBlobRequestHash(request AttachBlobRequest) []byte {
	return hashRequestPayload(map[string]any{
		"object_blob_id":   request.ObjectBlobID.String(),
		"base_row_version": request.BaseRowVersion,
	})
}

func hashRequestPayload(payload any) []byte {
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return append([]byte(nil), sum[:]...)
}

func decodeStrictObject(reader io.Reader, code string) (map[string]json.RawMessage, *auth.APIError) {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	var raw map[string]json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalidRequest(code, "", "request_not_object")
	}
	if raw == nil {
		return nil, invalidRequest(code, "", "request_not_object")
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return nil, invalidRequest(code, "", "request_not_object")
	}
	return raw, nil
}

func requiredBlobString(raw map[string]json.RawMessage, field string) (string, *auth.APIError) {
	value, ok := raw[field]
	if !ok {
		return "", invalidBlobCreate(field, "missing_required_field")
	}
	if string(value) == "null" {
		return "", invalidBlobCreate(field, "field_not_nullable")
	}
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		return "", invalidBlobCreate(field, "field_empty_after_normalization")
	}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", invalidBlobCreate(field, "field_empty_after_normalization")
	}
	return trimmed, nil
}

func requiredBlobUUID(raw map[string]json.RawMessage, field string) (uuid.UUID, *auth.APIError) {
	text, apiErr := requiredBlobString(raw, field)
	if apiErr != nil {
		return uuid.UUID{}, apiErr
	}
	id, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, invalidBlobCreate(field, "")
	}
	return id, nil
}

func requiredBlobInt64(raw map[string]json.RawMessage, field string) (int64, *auth.APIError) {
	value, ok := raw[field]
	if !ok {
		return 0, invalidBlobCreate(field, "missing_required_field")
	}
	if string(value) == "null" {
		return 0, invalidBlobCreate(field, "field_not_nullable")
	}
	var number json.Number
	if err := json.Unmarshal(value, &number); err != nil {
		return 0, invalidBlobCreate(field, "invalid_byte_size")
	}
	integer, err := number.Int64()
	if err != nil || number.String() != fmt.Sprintf("%d", integer) {
		return 0, invalidBlobCreate(field, "invalid_byte_size")
	}
	return integer, nil
}

func optionalBlobTrimmedString(raw map[string]json.RawMessage, field string) (*string, *auth.APIError) {
	value, ok := raw[field]
	if !ok || string(value) == "null" {
		return nil, nil
	}
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		if field == "sha256_hex" {
			return nil, invalidBlobCreate(field, "invalid_sha256_hex")
		}
		return nil, invalidBlobCreate(field, "")
	}
	trimmed := strings.TrimSpace(text)
	if field == "filename_hint" {
		trimmed = strings.ReplaceAll(trimmed, "\x00", "")
		trimmed = strings.TrimSpace(trimmed)
	}
	if trimmed == "" {
		return nil, nil
	}
	return &trimmed, nil
}

func requiredString(raw map[string]json.RawMessage, field string, code string) (string, *auth.APIError) {
	value, ok := raw[field]
	if !ok || string(value) == "null" {
		return "", invalidRequest(code, field, "missing_required_field")
	}
	var text string
	if err := json.Unmarshal(value, &text); err != nil || text == "" {
		return "", invalidRequest(code, field, "invalid_value")
	}
	return text, nil
}

func requiredUUID(raw map[string]json.RawMessage, field string, code string) (uuid.UUID, *auth.APIError) {
	text, apiErr := requiredString(raw, field, code)
	if apiErr != nil {
		return uuid.UUID{}, apiErr
	}
	id, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, invalidRequest(code, field, "invalid_value")
	}
	return id, nil
}

func requiredInt64(raw map[string]json.RawMessage, field string, code string) (int64, *auth.APIError) {
	value, ok := raw[field]
	if !ok || string(value) == "null" {
		return 0, invalidRequest(code, field, "missing_required_field")
	}
	var number json.Number
	if err := json.Unmarshal(value, &number); err != nil {
		return 0, invalidRequest(code, field, "invalid_value")
	}
	integer, err := number.Int64()
	if err != nil || number.String() != fmt.Sprintf("%d", integer) {
		return 0, invalidRequest(code, field, "invalid_value")
	}
	return integer, nil
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func invalidRequest(code string, field string, reasonCode string) *auth.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &auth.APIError{Status: http.StatusBadRequest, Code: code, Message: code, Details: details}
}

func invalidBlobCreate(field string, reasonCode string) *auth.APIError {
	return invalidRequest("invalid_blob_create_request", field, reasonCode)
}

func invalidMutationPayload(field string, reasonCode string) *auth.APIError {
	return invalidRequest("invalid_mutation_payload", field, reasonCode)
}

func invalidEvidenceHandleRequest(field string, reasonCode string) *auth.APIError {
	return invalidRequest("invalid_evidence_handle_request", field, reasonCode)
}

func clientTxnConflict(clientTxnID string) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "client_txn_conflict", Details: map[string]any{"client_txn_id": clientTxnID}}
}

func rowVersionConflict(recordID uuid.UUID, base int64, current int64) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "row_version_conflict", Details: map[string]any{
		"record_id": recordID.String(), "base_row_version": base, "current_row_version": current,
	}}
}

func evidenceAccessUnavailable(reasonCode string) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "evidence_access_unavailable", Details: map[string]any{"reason_code": reasonCode}}
}

func handleNotFoundOrRevoked() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "handle_not_found_or_revoked", Details: map[string]any{}}
}

func randomToken(prefix string) (string, error) {
	var bytes [20]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(bytes[:]), nil
}

func sortedChangedKeys(before map[string]any, after map[string]any) []string {
	beforeCells, _ := before["cells"].(map[string]any)
	afterCells, _ := after["cells"].(map[string]any)
	keys := make([]string, 0)
	for key, afterValue := range afterCells {
		if fmt.Sprintf("%#v", beforeCells[key]) != fmt.Sprintf("%#v", afterValue) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func sanitizeFilename(input string, recordID uuid.UUID, contentType string) string {
	name := strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', 0, '\r', '\n':
			return -1
		default:
			return r
		}
	}, strings.TrimSpace(input))
	name = strings.Trim(name, ". ")
	if name == "" || name == "." || name == ".." || strings.Contains(name, "..") {
		name = "evidence-" + recordID.String() + extensionForContentType(contentType)
	}
	return name
}

func formatContentDisposition(disposition string, filename string) string {
	ascii := asciiHeaderFilename(filename)
	if ascii == "" {
		ascii = "download"
	}
	formatted := mime.FormatMediaType(disposition, map[string]string{"filename": ascii})
	return formatted + "; filename*=UTF-8''" + url.PathEscape(filename)
}

func asciiHeaderFilename(filename string) string {
	var builder strings.Builder
	for _, r := range filename {
		switch {
		case r >= 0x20 && r <= 0x7e && r != '"' && r != '\\' && r != ';':
			builder.WriteRune(r)
		case r == ' ':
			builder.WriteByte(' ')
		default:
			builder.WriteByte('_')
		}
	}
	return strings.TrimSpace(builder.String())
}

func nullableStringEqual(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func extensionForContentType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])) {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "application/pdf":
		return ".pdf"
	case "text/plain":
		return ".txt"
	default:
		return ""
	}
}

func classifyMedia(contentType string) (mediaClass string, previewKind *string) {
	base := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch {
	case base == "text/html" || base == "application/xhtml+xml":
		return "text", nil
	case base == "image/svg+xml":
		return "image", nil
	case base == "image/png" || base == "image/jpeg" || base == "image/gif" || base == "image/webp" || base == "image/bmp" || base == "image/tiff":
		kind := "image_inline"
		return "image", &kind
	case base == "application/pdf":
		kind := "pdf_inline"
		return "pdf", &kind
	case base == "text/plain" || base == "text/csv" || base == "text/tab-separated-values" || base == "text/markdown" || base == "application/json" || base == "application/x-ndjson":
		kind := "text_inline"
		return "text", &kind
	case strings.HasPrefix(base, "text/"):
		return "text", nil
	case strings.HasPrefix(base, "audio/"):
		return "audio", nil
	case strings.HasPrefix(base, "video/"):
		return "video", nil
	case strings.Contains(base, "zip") || strings.Contains(base, "tar"):
		return "archive", nil
	default:
		return "binary", nil
	}
}

func formatHTTPTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
