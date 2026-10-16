package evidence

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
)

const evidenceMetadataRequestMaxBytes = 16 * 1024

func decodeMetadataQuery(reader io.Reader) ([]uuid.UUID, *httpapi.APIError) {
	const code = "invalid_evidence_metadata_request"
	body, err := io.ReadAll(io.LimitReader(reader, evidenceMetadataRequestMaxBytes+1))
	if err != nil || len(body) > evidenceMetadataRequestMaxBytes {
		return nil, invalidRequest(code, "", "request_not_object")
	}
	raw, err := httpapi.DecodeStrictJSONObject(bytes.NewReader(body))
	if err != nil {
		return nil, invalidRequest(code, "", "request_not_object")
	}
	for key := range raw {
		if key != "record_ids" {
			return nil, invalidRequest(code, key, "unknown_field")
		}
	}
	var ids []string
	if err := json.Unmarshal(raw["record_ids"], &ids); err != nil || len(ids) < 1 || len(ids) > 100 {
		return nil, invalidRequest(code, "record_ids", "invalid_record_ids")
	}
	parsed := make([]uuid.UUID, len(ids))
	seen := make(map[uuid.UUID]bool, len(ids))
	for i, text := range ids {
		id, err := uuid.Parse(text)
		if err != nil || id == uuid.Nil || id.String() != text || seen[id] {
			return nil, invalidRequest(code, "record_ids", "invalid_record_ids")
		}
		parsed[i], seen[id] = id, true
	}
	return parsed, nil
}

// This is metadata eligibility only. Issuance still verifies object presence,
// and redemption independently rechecks the current authorization and version.
func (s *service) previewMetadataBlocker(access evidenceAccessRecord) string {
	if reason := classifyEvidenceAccess(access, nil); reason != "" {
		return reason
	}
	if access.PreviewKind == nil {
		return "unsupported_preview"
	}
	if access.SizeBytes > s.previewMax || (*access.PreviewKind == "text_inline" && access.SizeBytes > s.textPreviewMax) {
		return "preview_payload_too_large"
	}
	return ""
}

func (s *service) handleMetadataQuery(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.admission.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	if _, apiErr := s.admission.visibleIncident(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	ids, apiErr := decodeMetadataQuery(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	records, err := s.operations.LoadEvidenceMetadata(r.Context(), incidentID, ids)
	if errors.Is(err, ErrEvidenceNotFound) {
		writeAPIError(w, r, evidenceRecordNotFound())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	items := make([]map[string]any, len(records))
	for i, access := range records {
		reason := s.previewMetadataBlocker(access)
		var uploadState, reasonCode any
		if access.ObjectBlobID != nil && access.BlobMetadataVisible {
			uploadState = access.UploadState
		}
		if reason != "" {
			reasonCode = reason
		}
		items[i] = map[string]any{
			"record_id": access.RecordID.String(), "row_version": access.RecordRowVersion,
			"title": access.Title, "lifecycle_state": access.EvidenceLifecycleState,
			"has_blob": access.ObjectBlobID != nil, "upload_state": uploadState,
			"preview_eligibility": map[string]any{"eligible": reason == "", "reason_code": reasonCode},
		}
	}
	if err := s.admission.slide(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{"incident_id": incidentID.String(), "items": items})
}
