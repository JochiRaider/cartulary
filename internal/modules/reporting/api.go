package reporting

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	ProfileID = "snapshot_reporting"

	DerivationVersion = "cartulary.snapshot_export_model.v2"

	DefaultTemplateID      = "cartulary.report.default"
	DefaultTemplateVersion = "1"

	OutputKindHTML        = "html"
	OutputKindMarkdown    = "markdown"
	OutputKindSlidev      = "slidev"
	OutputKindMermaid     = "mermaid"
	OutputKindReenactment = "reenactment"

	ReleaseScopeInternalDraft  = "internal_draft"
	ReleaseScopeInternalReview = "internal_review"
	ReleaseScopeExternal       = "external_release"

	ReleaseStatePendingApproval = "pending_approval"
	ReleaseStateApproved        = "approved"
	ReleaseStatePublished       = "published"
	ReleaseStateInvalidated     = "invalidated"
	ReleaseStateRenderFailed    = "render_failed"
)

type CreateSnapshotRequest struct {
	IncidentID                   uuid.UUID
	ClientTxnID                  string
	SourceChangeSetHighWatermark *string
	Normalized                   []byte
}

type CreateReleaseRequest struct {
	SnapshotID              uuid.UUID
	ClientTxnID             string
	TemplateID              string
	TemplateVersion         string
	RedactionProfileID      string
	RedactionProfileVersion string
	OutputKind              string
	ReleaseScope            string
	RecipientPartitionRefs  []string
	Normalized              []byte
}

type ReleaseActionRequest struct {
	ClientTxnID string
	Reason      *string
	Normalized  []byte
}

func DecodeCreateSnapshotRequest(reader io.Reader) (CreateSnapshotRequest, *auth.APIError) {
	raw, apiErr := decodeJSONObject(reader, "invalid_snapshot_request")
	if apiErr != nil {
		return CreateSnapshotRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"incident_id":                      {},
		"client_txn_id":                    {},
		"source_change_set_high_watermark": {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return CreateSnapshotRequest{}, invalidSnapshotRequest(key, "unknown_field")
		}
	}
	var request CreateSnapshotRequest
	if value, ok := raw["incident_id"]; !ok {
		return CreateSnapshotRequest{}, invalidSnapshotRequest("incident_id", "missing_required_field")
	} else if bytesEqualJSONNull(value) {
		return CreateSnapshotRequest{}, invalidSnapshotRequest("incident_id", "field_not_nullable")
	} else {
		var rawID string
		if err := json.Unmarshal(value, &rawID); err != nil {
			return CreateSnapshotRequest{}, invalidSnapshotRequest("incident_id", "invalid_value")
		}
		parsed, err := uuid.Parse(rawID)
		if err != nil {
			return CreateSnapshotRequest{}, invalidSnapshotRequest("incident_id", "invalid_value")
		}
		request.IncidentID = parsed
	}
	clientTxnID, apiErr := requiredStringField(raw, "client_txn_id", "invalid_snapshot_request")
	if apiErr != nil {
		return CreateSnapshotRequest{}, apiErr
	}
	request.ClientTxnID = clientTxnID
	if value, ok := raw["source_change_set_high_watermark"]; ok {
		if bytesEqualJSONNull(value) {
			return CreateSnapshotRequest{}, invalidSnapshotRequest("source_change_set_high_watermark", "field_not_nullable")
		}
		var watermark string
		if err := json.Unmarshal(value, &watermark); err != nil || strings.TrimSpace(watermark) == "" {
			return CreateSnapshotRequest{}, invalidSnapshotRequest("source_change_set_high_watermark", "invalid_value")
		}
		request.SourceChangeSetHighWatermark = &watermark
	}
	normalized, err := normalizeSnapshotRequest(request.IncidentID, request.ClientTxnID, request.SourceChangeSetHighWatermark)
	if err != nil {
		return CreateSnapshotRequest{}, internalAPIError(err)
	}
	request.Normalized = normalized
	return request, nil
}

func DecodeCreateReleaseRequest(reader io.Reader) (CreateReleaseRequest, *auth.APIError) {
	raw, apiErr := decodeJSONObject(reader, "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"snapshot_id":               {},
		"client_txn_id":             {},
		"template_id":               {},
		"template_version":          {},
		"redaction_profile_id":      {},
		"redaction_profile_version": {},
		"output_kind":               {},
		"release_scope":             {},
		"recipient_partition_refs":  {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return CreateReleaseRequest{}, invalidReleaseRequest(key, "unknown_field")
		}
	}
	var request CreateReleaseRequest
	if value, ok := raw["snapshot_id"]; !ok {
		return CreateReleaseRequest{}, invalidReleaseRequest("snapshot_id", "missing_required_field")
	} else if bytesEqualJSONNull(value) {
		return CreateReleaseRequest{}, invalidReleaseRequest("snapshot_id", "field_not_nullable")
	} else {
		var rawID string
		if err := json.Unmarshal(value, &rawID); err != nil {
			return CreateReleaseRequest{}, invalidReleaseRequest("snapshot_id", "invalid_value")
		}
		parsed, err := uuid.Parse(rawID)
		if err != nil {
			return CreateReleaseRequest{}, invalidReleaseRequest("snapshot_id", "invalid_value")
		}
		request.SnapshotID = parsed
	}
	clientTxnID, apiErr := requiredStringField(raw, "client_txn_id", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.ClientTxnID = clientTxnID
	request.ReleaseScope = ReleaseScopeInternalDraft
	templateID, apiErr := requiredStringField(raw, "template_id", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.TemplateID = templateID
	templateVersion, apiErr := requiredStringField(raw, "template_version", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.TemplateVersion = templateVersion
	outputKind, apiErr := requiredStringField(raw, "output_kind", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.OutputKind = outputKind
	if value, ok := raw["release_scope"]; ok {
		value, apiErr := optionalNonNullString(value, "release_scope", "invalid_release_request")
		if apiErr != nil {
			return CreateReleaseRequest{}, apiErr
		}
		request.ReleaseScope = value
	}
	if !validOutputKind(request.OutputKind) {
		return CreateReleaseRequest{}, invalidReleaseRequest("output_kind", "unsupported_output_kind")
	}
	if !validReleaseScope(request.ReleaseScope) {
		return CreateReleaseRequest{}, invalidReleaseRequest("release_scope", "unsupported_release_scope")
	}
	if value, ok := raw["recipient_partition_refs"]; ok {
		refs, apiErr := optionalStringSet(value, "recipient_partition_refs", "invalid_release_request")
		if apiErr != nil {
			return CreateReleaseRequest{}, apiErr
		}
		request.RecipientPartitionRefs = refs
	}
	if request.RecipientPartitionRefs == nil {
		request.RecipientPartitionRefs = []string{}
	}
	if request.ReleaseScope != ReleaseScopeExternal && len(request.RecipientPartitionRefs) > 0 {
		return CreateReleaseRequest{}, invalidReleaseRequest("recipient_partition_refs", "recipient_partitions_not_allowed")
	}
	redactionProfileID, apiErr := requiredStringField(raw, "redaction_profile_id", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.RedactionProfileID = redactionProfileID
	redactionProfileVersion, apiErr := requiredStringField(raw, "redaction_profile_version", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.RedactionProfileVersion = redactionProfileVersion
	normalized, err := json.Marshal(map[string]any{
		"snapshot_id":               request.SnapshotID.String(),
		"client_txn_id":             request.ClientTxnID,
		"template_id":               request.TemplateID,
		"template_version":          request.TemplateVersion,
		"redaction_profile_id":      request.RedactionProfileID,
		"redaction_profile_version": request.RedactionProfileVersion,
		"output_kind":               request.OutputKind,
		"release_scope":             request.ReleaseScope,
		"recipient_partition_refs":  request.RecipientPartitionRefs,
	})
	if err != nil {
		return CreateReleaseRequest{}, internalAPIError(err)
	}
	request.Normalized = normalized
	return request, nil
}

func DecodeReleaseActionRequest(reader io.Reader) (ReleaseActionRequest, *auth.APIError) {
	raw, apiErr := decodeJSONObject(reader, "invalid_release_request")
	if apiErr != nil {
		return ReleaseActionRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"client_txn_id": {},
		"reason":        {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ReleaseActionRequest{}, invalidReleaseRequest(key, "unknown_field")
		}
	}
	clientTxnID, apiErr := requiredStringField(raw, "client_txn_id", "invalid_release_request")
	if apiErr != nil {
		return ReleaseActionRequest{}, apiErr
	}
	var reason *string
	if value, ok := raw["reason"]; ok {
		if bytesEqualJSONNull(value) {
			reason = nil
		} else {
			parsed, apiErr := optionalNonNullString(value, "reason", "invalid_release_request")
			if apiErr != nil {
				return ReleaseActionRequest{}, apiErr
			}
			trimmed := strings.TrimSpace(parsed)
			if trimmed != "" {
				reason = &trimmed
			}
		}
	}
	normalized, err := json.Marshal(map[string]any{
		"client_txn_id": clientTxnID,
		"reason":        optionalStringForHash(reason),
	})
	if err != nil {
		return ReleaseActionRequest{}, internalAPIError(err)
	}
	return ReleaseActionRequest{ClientTxnID: clientTxnID, Reason: reason, Normalized: normalized}, nil
}

func decodeJSONObject(reader io.Reader, errorCode string) (map[string]json.RawMessage, *auth.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return nil, &auth.APIError{Status: http.StatusBadRequest, Code: errorCode, Details: map[string]any{"reason_code": "request_not_object"}}
	}
	if raw == nil {
		return nil, &auth.APIError{Status: http.StatusBadRequest, Code: errorCode, Details: map[string]any{"reason_code": "request_not_object"}}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, &auth.APIError{Status: http.StatusBadRequest, Code: errorCode, Details: map[string]any{"reason_code": "request_not_object"}}
	}
	return raw, nil
}

func requiredStringField(raw map[string]json.RawMessage, field string, errorCode string) (string, *auth.APIError) {
	value, ok := raw[field]
	if !ok {
		return "", invalidRequest(errorCode, field, "missing_required_field")
	}
	if bytesEqualJSONNull(value) {
		return "", invalidRequest(errorCode, field, "field_not_nullable")
	}
	var parsed string
	if err := json.Unmarshal(value, &parsed); err != nil || strings.TrimSpace(parsed) == "" {
		return "", invalidRequest(errorCode, field, "invalid_value")
	}
	return parsed, nil
}

func optionalNonNullString(raw json.RawMessage, field string, errorCode string) (string, *auth.APIError) {
	if bytesEqualJSONNull(raw) {
		return "", invalidRequest(errorCode, field, "field_not_nullable")
	}
	var parsed string
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", invalidRequest(errorCode, field, "invalid_value")
	}
	return parsed, nil
}

func optionalStringSet(raw json.RawMessage, field string, errorCode string) ([]string, *auth.APIError) {
	if bytesEqualJSONNull(raw) {
		return nil, invalidRequest(errorCode, field, "field_not_nullable")
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, invalidRequest(errorCode, field, "invalid_value")
	}
	seen := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return nil, invalidRequest(errorCode, field, "invalid_value")
		}
		seen[value] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for value := range seen {
		out = append(out, value)
	}
	sort.Strings(out)
	return out, nil
}

func normalizeSnapshotRequest(incidentID uuid.UUID, clientTxnID string, watermark *string) ([]byte, error) {
	return json.Marshal(map[string]any{
		"incident_id":                      incidentID.String(),
		"client_txn_id":                    clientTxnID,
		"source_change_set_high_watermark": optionalStringForHash(watermark),
	})
}

func optionalStringForHash(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func validOutputKind(kind string) bool {
	switch kind {
	case OutputKindHTML, OutputKindMarkdown, OutputKindSlidev, OutputKindMermaid, OutputKindReenactment:
		return true
	default:
		return false
	}
}

func validReleaseScope(scope string) bool {
	switch scope {
	case ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal:
		return true
	default:
		return false
	}
}

func isSupportedRedactionProfileSelector(id string, version string) bool {
	switch id {
	case InternalRedactionProfileID, ExternalRedactionProfileID:
		return version == "1"
	default:
		return false
	}
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func invalidSnapshotRequest(field string, reasonCode string) *auth.APIError {
	return invalidRequest("invalid_snapshot_request", field, reasonCode)
}

func invalidReleaseRequest(field string, reasonCode string) *auth.APIError {
	return invalidRequest("invalid_release_request", field, reasonCode)
}

func invalidRequest(code string, field string, reasonCode string) *auth.APIError {
	return &auth.APIError{
		Status: http.StatusBadRequest,
		Code:   code,
		Details: map[string]any{
			"field":       field,
			"reason_code": reasonCode,
		},
	}
}

func clientTxnConflict(clientTxnID string) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "client_txn_conflict", Details: map[string]any{"client_txn_id": clientTxnID}}
}

func bytesEqualJSONNull(value json.RawMessage) bool {
	return strings.TrimSpace(string(value)) == "null"
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *auth.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func internalAPIError(err error) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}

func releaseStateConflict(reasonCode string) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "release_state_conflict", Details: map[string]any{"reason_code": reasonCode}}
}

func releaseApprovalRejected(reasonCode string) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "release_approval_rejected", Details: map[string]any{"reason_code": reasonCode}}
}

func unsupportedTemplateError(id string, version string) *auth.APIError {
	return &auth.APIError{
		Status: http.StatusBadRequest,
		Code:   "invalid_release_request",
		Details: map[string]any{
			"field":            "template_id",
			"reason_code":      "unsupported_template",
			"template_id":      id,
			"template_version": version,
		},
	}
}

func unsupportedRedactionProfileError(id string, version string) *auth.APIError {
	return &auth.APIError{
		Status: http.StatusBadRequest,
		Code:   "invalid_release_request",
		Details: map[string]any{
			"field":                     "redaction_profile_id",
			"reason_code":               "unsupported_redaction_profile",
			"redaction_profile_id":      id,
			"redaction_profile_version": version,
		},
	}
}

func snapshotBoundaryMismatch(expected string, actual string) *auth.APIError {
	return &auth.APIError{
		Status: http.StatusConflict,
		Code:   "snapshot_source_boundary_conflict",
		Details: map[string]any{
			"expected_source_change_set_high_watermark": expected,
			"current_source_change_set_high_watermark":  actual,
		},
	}
}
