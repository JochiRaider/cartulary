package reference_data

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	ProfileID = "reference_pack"

	PackContractVersionV1 = "cartulary.reference_pack.v1"

	MediaTypeZip         = "application/zip"
	MediaTypeTar         = "application/x-tar"
	MediaTypeGzip        = "application/gzip"
	MediaTypeXGzip       = "application/x-gzip"
	MediaTypeOctetStream = "application/octet-stream"

	ConditionStaged            = "staged"
	ConditionVerifiedAvailable = "verified_available"
	ConditionDisabled          = "disabled"
	ConditionFailed            = "failed"
	ConditionMissing           = "missing"

	StoredStatusStaged    = "staged"
	StoredStatusAvailable = "available"
	StoredStatusDisabled  = "disabled"
	StoredStatusFailed    = "failed"
	StoredStatusMissing   = "missing"

	VerificationPending = "pending"
	VerificationPassed  = "passed"
	VerificationFailed  = "failed"

	ResultReferencePackImported   = "reference_pack_imported"
	ResultReferencePackReverified = "reference_pack_reverified"
	ResultReferencePacksRefreshed = "reference_packs_refreshed"
)

var ReferencePackFileContentTypes = []string{
	MediaTypeZip,
	MediaTypeTar,
	MediaTypeGzip,
	MediaTypeXGzip,
	MediaTypeOctetStream,
}

type ImportMetadataRequest struct {
	ClientTxnID      string
	ActivationPolicy string
	Normalized       []byte
}

type ActionRequest struct {
	ClientTxnID string
	Reason      *string
	Normalized  []byte
}

type RefreshRequest struct {
	ClientTxnID      string
	PackKeysProvided bool
	PackKeys         []string
	ResolvedPackKeys []string
	Normalized       []byte
}

type VersionRecord struct {
	PackKey             string
	PackKind            string
	PackVersion         string
	StoredStatus        string
	Active              bool
	SourceIdentifier    *string
	ManifestSHA256      string
	PayloadSHA256       string
	PackContractVersion string
	VerificationMethod  string
	VerificationResult  string
	SignerKeyID         *string
	PreviousActive      *string
	ImportedByUserID    *string
	ImportedAt          time.Time
	ActivatedByUserID   *string
	ActivatedAt         *time.Time
	BundleSHA256        string
	BundleStoragePath   string
}

type apiError struct {
	apiErr *auth.APIError
}

func (e apiError) Error() string {
	if e.apiErr == nil {
		return "reference pack api error"
	}
	return e.apiErr.Code
}

func wrapAPIError(apiErr *auth.APIError) error {
	if apiErr == nil {
		return nil
	}
	return apiError{apiErr: apiErr}
}

func (r VersionRecord) Resource() map[string]any {
	return map[string]any{
		"pack_key":                r.PackKey,
		"pack_kind":               r.PackKind,
		"pack_version":            r.PackVersion,
		"pack_version_state":      publicCondition(r.StoredStatus, r.VerificationResult),
		"active":                  r.Active,
		"source_identifier":       optionalString(r.SourceIdentifier),
		"manifest_sha256":         r.ManifestSHA256,
		"payload_sha256":          r.PayloadSHA256,
		"pack_contract_version":   r.PackContractVersion,
		"verification_method":     r.VerificationMethod,
		"verification_result":     r.VerificationResult,
		"signer_key_id":           optionalString(r.SignerKeyID),
		"previous_active_version": optionalString(r.PreviousActive),
		"imported_by_user_id":     optionalString(r.ImportedByUserID),
		"imported_at":             r.ImportedAt,
		"activated_by_user_id":    optionalString(r.ActivatedByUserID),
		"activated_at":            optionalTime(r.ActivatedAt),
	}
}

func publicCondition(status string, verificationResult string) string {
	switch status {
	case StoredStatusAvailable:
		if verificationResult == VerificationPassed {
			return ConditionVerifiedAvailable
		}
		if verificationResult == VerificationFailed {
			return ConditionFailed
		}
		return ConditionStaged
	case StoredStatusDisabled:
		return ConditionDisabled
	case StoredStatusFailed:
		return ConditionFailed
	case StoredStatusMissing:
		return ConditionMissing
	default:
		return ConditionStaged
	}
}

func optionalString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func optionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func DecodeImportMetadata(envelope httpapi.UploadEnvelope) (ImportMetadataRequest, *auth.APIError) {
	allowed := map[string]struct{}{
		"client_txn_id":     {},
		"activation_policy": {},
	}
	for key := range envelope.Metadata {
		if _, ok := allowed[key]; !ok {
			return ImportMetadataRequest{}, invalidReferencePackRequest(key, "unknown_field")
		}
	}
	clientTxnID, apiErr := requiredStringField(envelope.Metadata, "client_txn_id")
	if apiErr != nil {
		return ImportMetadataRequest{}, apiErr
	}
	activationPolicy := "staged_only"
	if value, ok := envelope.Metadata["activation_policy"]; ok {
		if bytesEqualJSONNull(value) {
			return ImportMetadataRequest{}, invalidReferencePackRequest("activation_policy", "field_not_nullable")
		}
		var parsed string
		if err := json.Unmarshal(value, &parsed); err != nil {
			return ImportMetadataRequest{}, invalidReferencePackRequest("activation_policy", "invalid_activation_policy")
		}
		if parsed != "staged_only" {
			return ImportMetadataRequest{}, invalidReferencePackRequest("activation_policy", "auto_activation_not_supported")
		}
		activationPolicy = parsed
	}
	normalized, err := json.Marshal(map[string]any{
		"activation_policy": activationPolicy,
		"file_sha256":       envelope.FileSHA256Hex,
	})
	if err != nil {
		return ImportMetadataRequest{}, internalAPIError(err)
	}
	return ImportMetadataRequest{
		ClientTxnID:      clientTxnID,
		ActivationPolicy: activationPolicy,
		Normalized:       normalized,
	}, nil
}

func DecodeActionRequest(reader io.Reader) (ActionRequest, *auth.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return ActionRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"client_txn_id": {},
		"reason":        {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ActionRequest{}, invalidReferencePackRequest(key, "unknown_field")
		}
	}
	clientTxnID, apiErr := requiredStringField(raw, "client_txn_id")
	if apiErr != nil {
		return ActionRequest{}, apiErr
	}
	var reason *string
	if value, ok := raw["reason"]; ok && !bytesEqualJSONNull(value) {
		var parsed string
		if err := json.Unmarshal(value, &parsed); err != nil {
			return ActionRequest{}, invalidReferencePackRequest("reason", "request_not_object")
		}
		if normalized, ok := fieldnorm.NormalizeNote(parsed); ok {
			reason = &normalized
		}
	}
	normalized, err := json.Marshal(map[string]any{
		"client_txn_id": clientTxnID,
		"reason":        optionalString(reason),
	})
	if err != nil {
		return ActionRequest{}, internalAPIError(err)
	}
	return ActionRequest{ClientTxnID: clientTxnID, Reason: reason, Normalized: normalized}, nil
}

func DecodeRefreshRequest(reader io.Reader) (RefreshRequest, *auth.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return RefreshRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"client_txn_id": {},
		"pack_keys":     {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return RefreshRequest{}, invalidReferencePackRequest(key, "unknown_field")
		}
	}
	clientTxnID, apiErr := requiredStringField(raw, "client_txn_id")
	if apiErr != nil {
		return RefreshRequest{}, apiErr
	}
	request := RefreshRequest{ClientTxnID: clientTxnID}
	if value, ok := raw["pack_keys"]; ok {
		request.PackKeysProvided = true
		if bytesEqualJSONNull(value) {
			return RefreshRequest{}, invalidReferencePackRequest("pack_keys", "field_not_nullable")
		}
		var values []any
		if err := json.Unmarshal(value, &values); err != nil {
			return RefreshRequest{}, invalidReferencePackRequest("pack_keys", "invalid_pack_keys")
		}
		if len(values) == 0 {
			return RefreshRequest{}, invalidReferencePackRequest("pack_keys", "empty_pack_keys")
		}
		seen := map[string]struct{}{}
		for _, value := range values {
			packKey, ok := value.(string)
			if !ok || strings.TrimSpace(packKey) == "" {
				return RefreshRequest{}, invalidReferencePackRequest("pack_keys", "invalid_pack_keys")
			}
			seen[packKey] = struct{}{}
		}
		request.PackKeys = make([]string, 0, len(seen))
		for packKey := range seen {
			request.PackKeys = append(request.PackKeys, packKey)
		}
		sort.Strings(request.PackKeys)
	}
	return request, nil
}

func NormalizeRefreshRequest(request RefreshRequest, resolved []string) (RefreshRequest, error) {
	request.ResolvedPackKeys = append([]string(nil), resolved...)
	sort.Strings(request.ResolvedPackKeys)
	normalized, err := json.Marshal(map[string]any{
		"client_txn_id":      request.ClientTxnID,
		"resolved_pack_keys": request.ResolvedPackKeys,
	})
	if err != nil {
		return RefreshRequest{}, err
	}
	request.Normalized = normalized
	return request, nil
}

func ValidateRefreshPackKeys(request RefreshRequest, visible []string) ([]string, *auth.APIError) {
	visibleSet := map[string]struct{}{}
	for _, packKey := range visible {
		visibleSet[packKey] = struct{}{}
	}
	if !request.PackKeysProvided {
		return append([]string(nil), visible...), nil
	}
	for _, packKey := range request.PackKeys {
		if _, ok := visibleSet[packKey]; !ok {
			return nil, invalidReferencePackRequest("pack_keys", "invalid_pack_keys")
		}
	}
	return append([]string(nil), request.PackKeys...), nil
}

func decodeJSONObject(reader io.Reader) (map[string]json.RawMessage, *auth.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, &auth.APIError{Status: http.StatusBadRequest, Code: "invalid_reference_pack_request", Details: map[string]any{"reason_code": "request_not_object"}}
	}
	if raw == nil {
		return nil, &auth.APIError{Status: http.StatusBadRequest, Code: "invalid_reference_pack_request", Details: map[string]any{"reason_code": "request_not_object"}}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, &auth.APIError{Status: http.StatusBadRequest, Code: "invalid_reference_pack_request", Details: map[string]any{"reason_code": "request_not_object"}}
	}
	return raw, nil
}

func requiredStringField(raw map[string]json.RawMessage, field string) (string, *auth.APIError) {
	value, ok := raw[field]
	if !ok {
		return "", invalidReferencePackRequest(field, "missing_required_field")
	}
	if bytesEqualJSONNull(value) {
		return "", invalidReferencePackRequest(field, "field_not_nullable")
	}
	var parsed string
	if err := json.Unmarshal(value, &parsed); err != nil || strings.TrimSpace(parsed) == "" {
		return "", invalidReferencePackRequest(field, "missing_required_field")
	}
	return parsed, nil
}

func invalidReferencePackRequest(field string, reasonCode string) *auth.APIError {
	return &auth.APIError{
		Status: http.StatusBadRequest,
		Code:   "invalid_reference_pack_request",
		Details: map[string]any{
			"field":       field,
			"reason_code": reasonCode,
		},
	}
}

func uploadEnvelopeAPIError(apiErr *httpapi.UploadEnvelopeError) *auth.APIError {
	if apiErr == nil {
		return nil
	}
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_reference_pack_request",
		Message: fmt.Sprintf("invalid reference pack request: %s", apiErr.ReasonCode),
		Details: apiErr.Details(),
	}
}

func referencePackNotFound() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "reference_pack_not_found", Details: map[string]any{}}
}

func referencePackActivationRejected(reasonCode string) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "reference_pack_activation_rejected", Details: map[string]any{"reason_code": reasonCode}}
}

func referencePackStateConflict(reasonCode string) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "reference_pack_state_conflict", Details: map[string]any{"reason_code": reasonCode}}
}

func referencePackVerificationFailed(reasonCode string) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "reference_pack_verification_failed", Details: map[string]any{"reason_code": reasonCode}}
}

func clientTxnConflict(clientTxnID string) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "client_txn_conflict", Details: map[string]any{"client_txn_id": clientTxnID}}
}

func invalidPaginationRequest(reasonCode string) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_pagination_request",
		Message: "invalid pagination request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *auth.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteErrorWithConflict(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details, apiErr.Conflict)
}

func internalAPIError(err error) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}

func bytesEqualJSONNull(value json.RawMessage) bool {
	return strings.TrimSpace(string(value)) == "null"
}

func hashBytes(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func referencePackResourceRef(packKey string, packVersion string) struct {
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	Route string `json:"route"`
} {
	route := referencePackRoute(packKey, packVersion)
	return struct {
		Kind  string `json:"kind"`
		ID    string `json:"id"`
		Route string `json:"route"`
	}{Kind: "reference_pack_version", ID: route, Route: route}
}

func referencePackRoute(packKey string, packVersion string) string {
	return "/api/v1/reference-packs/" + packKey + "/" + packVersion
}

func isValidVerificationFailureReason(reason string) bool {
	return slices.Contains([]string{
		"checksum_mismatch",
		"signature_mismatch",
		"missing_integrity_metadata",
		"contract_incompatible",
		"path_traversal",
		"disallowed_content",
		"payload_missing",
		"archive_extracted_bytes_exceeded",
		"archive_compression_ratio_exceeded",
		"archive_member_count_exceeded",
	}, reason)
}
