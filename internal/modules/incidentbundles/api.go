package incidentbundles

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	ProfileID                  = "incident_portability"
	BundlesRouteContributionID = "incident_portability.bundles_route"
	BundleWorkerKind           = "incident_portability.bundle_worker_v1"
	ExportJobKind              = "incident_portability.export_v1"
	ImportJobKind              = "incident_portability.import_v1"
	ExportOperationKind        = "incident_portability.export"
	ImportOperationKind        = "incident_portability.import"

	HistoryModeFull = "full"
	BlobModeFull    = "full"

	ReferencePackModeRefsOnly = "refs_only"
	ReferencePackModeEmbedded = "embedded"

	ResultIncidentBundleExported = "incident_bundle_exported"
	ResultIncidentBundleImported = "incident_bundle_imported"

	MediaTypeZip         = "application/zip"
	MediaTypeTar         = "application/x-tar"
	MediaTypeGzip        = "application/gzip"
	MediaTypeXGzip       = "application/x-gzip"
	MediaTypeOctetStream = "application/octet-stream"
)

var IncidentBundleFileContentTypes = []string{
	MediaTypeZip,
	MediaTypeTar,
	MediaTypeGzip,
	MediaTypeXGzip,
	MediaTypeOctetStream,
}

type ExportRequest struct {
	IncidentID                    uuid.UUID
	ClientTxnID                   string
	ReferencePackMode             string
	OptionalSections              []string
	RequiredCapabilities          []string
	CapabilityActivationRequested bool
	HistoryMode                   string
	BlobMode                      string
	Normalized                    []byte
}

type ImportMetadataRequest struct {
	ClientTxnID string
	Normalized  []byte
}

func DecodeExportRequest(reader io.Reader) (ExportRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return ExportRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"incident_id":           {},
		"client_txn_id":         {},
		"reference_pack_mode":   {},
		"optional_sections":     {},
		"required_capabilities": {},
		"history_mode":          {},
		"blob_mode":             {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ExportRequest{}, invalidIncidentBundleRequest(key, "unknown_field")
		}
	}
	if _, ok := raw["history_mode"]; ok {
		return ExportRequest{}, invalidIncidentBundleRequest("history_mode", "history_mode_not_supported")
	}
	if _, ok := raw["blob_mode"]; ok {
		return ExportRequest{}, invalidIncidentBundleRequest("blob_mode", "blob_mode_not_supported")
	}
	incidentIDText, apiErr := requiredStringField(raw, "incident_id")
	if apiErr != nil {
		return ExportRequest{}, apiErr
	}
	incidentID, err := uuid.Parse(incidentIDText)
	if err != nil {
		return ExportRequest{}, invalidIncidentBundleRequest("incident_id", "invalid_value")
	}
	clientTxnID, apiErr := requiredStringField(raw, "client_txn_id")
	if apiErr != nil {
		return ExportRequest{}, apiErr
	}
	referencePackMode := ReferencePackModeRefsOnly
	if value, ok := raw["reference_pack_mode"]; ok {
		if bytesEqualJSONNull(value) {
			return ExportRequest{}, invalidIncidentBundleRequest("reference_pack_mode", "field_not_nullable")
		}
		if err := json.Unmarshal(value, &referencePackMode); err != nil {
			return ExportRequest{}, invalidIncidentBundleRequest("reference_pack_mode", "invalid_reference_pack_mode")
		}
		if referencePackMode != ReferencePackModeRefsOnly && referencePackMode != ReferencePackModeEmbedded {
			return ExportRequest{}, invalidIncidentBundleRequest("reference_pack_mode", "invalid_reference_pack_mode")
		}
	}
	optionalSections, apiErr := canonicalOptionalSections(raw)
	if apiErr != nil {
		return ExportRequest{}, apiErr
	}
	capabilityActivationRequested, apiErr := decodeRequiredCapabilities(raw)
	if apiErr != nil {
		return ExportRequest{}, apiErr
	}
	requiredCapabilities := []string{}
	normalized, err := json.Marshal(map[string]any{
		"blob_mode":             BlobModeFull,
		"history_mode":          HistoryModeFull,
		"incident_id":           incidentID.String(),
		"optional_sections":     optionalSections,
		"reference_pack_mode":   referencePackMode,
		"required_capabilities": requiredCapabilities,
	})
	if err != nil {
		return ExportRequest{}, internalAPIError(err)
	}
	return ExportRequest{
		IncidentID:                    incidentID,
		ClientTxnID:                   clientTxnID,
		ReferencePackMode:             referencePackMode,
		OptionalSections:              optionalSections,
		RequiredCapabilities:          requiredCapabilities,
		CapabilityActivationRequested: capabilityActivationRequested,
		HistoryMode:                   HistoryModeFull,
		BlobMode:                      BlobModeFull,
		Normalized:                    normalized,
	}, nil
}

func DecodeImportMetadata(envelope httpapi.UploadEnvelope) (ImportMetadataRequest, *httpapi.APIError) {
	allowed := map[string]struct{}{
		"client_txn_id": {},
	}
	for key := range envelope.Metadata {
		if _, ok := allowed[key]; !ok {
			return ImportMetadataRequest{}, invalidIncidentBundleRequest(key, "unknown_field")
		}
	}
	clientTxnID, apiErr := requiredStringField(envelope.Metadata, "client_txn_id")
	if apiErr != nil {
		return ImportMetadataRequest{}, apiErr
	}
	normalized, err := json.Marshal(map[string]any{
		"file_sha256": envelope.FileSHA256Hex,
	})
	if err != nil {
		return ImportMetadataRequest{}, internalAPIError(err)
	}
	return ImportMetadataRequest{ClientTxnID: clientTxnID, Normalized: normalized}, nil
}

func decodeJSONObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalidIncidentBundleRequest("", "request_not_object")
	}
	if raw == nil {
		return nil, invalidIncidentBundleRequest("", "request_not_object")
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return nil, invalidIncidentBundleRequest("", "request_not_object")
	}
	return raw, nil
}

func requiredStringField(raw map[string]json.RawMessage, field string) (string, *httpapi.APIError) {
	value, ok := raw[field]
	if !ok {
		return "", invalidIncidentBundleRequest(field, "missing_required_field")
	}
	if bytesEqualJSONNull(value) {
		return "", invalidIncidentBundleRequest(field, "field_not_nullable")
	}
	var parsed string
	if err := json.Unmarshal(value, &parsed); err != nil || strings.TrimSpace(parsed) == "" {
		return "", invalidIncidentBundleRequest(field, "invalid_value")
	}
	return parsed, nil
}

func canonicalOptionalSections(raw map[string]json.RawMessage) ([]string, *httpapi.APIError) {
	const field = "optional_sections"
	value, ok := raw[field]
	if !ok {
		return []string{}, nil
	}
	if bytesEqualJSONNull(value) {
		return nil, invalidIncidentBundleRequest(field, "field_not_nullable")
	}
	var items []string
	if err := json.Unmarshal(value, &items); err != nil {
		return nil, invalidIncidentBundleRequest(field, "invalid_optional_sections")
	}
	seen := map[string]struct{}{}
	for _, item := range items {
		if item != "reference_packs" && item != "snapshots" {
			return nil, invalidIncidentBundleRequest(field, "invalid_optional_sections")
		}
		seen[item] = struct{}{}
	}
	canonical := make([]string, 0, len(seen))
	for item := range seen {
		canonical = append(canonical, item)
	}
	sort.Strings(canonical)
	return canonical, nil
}

func decodeRequiredCapabilities(raw map[string]json.RawMessage) (bool, *httpapi.APIError) {
	const field = "required_capabilities"
	value, ok := raw[field]
	if !ok {
		return false, nil
	}
	if bytesEqualJSONNull(value) {
		return false, invalidIncidentBundleRequest(field, "field_not_nullable")
	}
	var items []string
	if err := json.Unmarshal(value, &items); err != nil {
		return false, invalidIncidentBundleRequest(field, "invalid_required_capabilities")
	}
	return len(items) > 0, nil
}

func invalidIncidentBundleRequest(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{"reason_code": reasonCode}
	if field != "" {
		details["field"] = field
	}
	return &httpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_incident_bundle_request", Details: details}
}

func extensionCapabilityNotSupported() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "extension_capability_not_supported",
		Details: map[string]any{"profile_id": ProfileID},
	}
}

func incidentBundleNotFound() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_bundle_not_found", Details: map[string]any{}}
}

func clientTxnConflict(clientTxnID string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "client_txn_conflict", Details: map[string]any{"client_txn_id": clientTxnID}}
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusInternalServerError, Code: "internal_error", Message: err.Error(), Details: map[string]any{}}
}

func uploadEnvelopeAPIError(err *httpapi.UploadEnvelopeError) *httpapi.APIError {
	if err == nil {
		return nil
	}
	return &httpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_incident_bundle_request", Details: err.Details()}
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	if apiErr == nil {
		apiErr = internalAPIError(fmt.Errorf("missing api error"))
	}
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteErrorWithConflict(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details, apiErr.Conflict)
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

func canonicalJSONString(value any) ([]byte, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}
