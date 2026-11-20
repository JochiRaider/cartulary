package evidence

import (
	"net/http"

	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteErrorWithOptions(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details, httpapi.ErrorOptions{
		Conflict:  apiErr.Conflict,
		Retryable: apiErr.Retryable || (apiErr.Status == http.StatusConflict && apiErr.Code == "record_locked"),
	})
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusInternalServerError, Code: "internal_error", Message: err.Error(), Details: map[string]any{}}
}

func objectStoreDependencyAPIError(err error) *httpapi.APIError {
	if reasonCode, ok := evidencepolicy.PersistedObjectBlobStorageKeyErrorReason(err); ok {
		return objectStoreInvalidRequestAPIError(reasonCode)
	}
	adapterErr, ok := objectstore.AsAdapterError(err)
	if !ok {
		return nil
	}
	switch adapterErr.Code {
	case objectstore.ErrorCodeUnavailable:
		return objectStoreUnavailableAPIError(unavailableReason(adapterErr), true)
	case objectstore.ErrorCodeDeadlineExceeded, objectstore.ErrorCodeRetryExhausted:
		return objectStoreUnavailableAPIError("retry_exhausted", true)
	case objectstore.ErrorCodeAccessRejected:
		return &httpapi.APIError{
			Status: http.StatusServiceUnavailable,
			Code:   "object_store_access_rejected",
			Details: map[string]any{
				"reason_code": accessRejectedReason(adapterErr),
			},
		}
	default:
		return nil
	}
}

func objectStoreInvalidRequestAPIError(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusInternalServerError,
		Code:   "object_store_invalid_request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func objectStoreUnavailableAPIError(reasonCode string, retryable bool) *httpapi.APIError {
	return &httpapi.APIError{
		Status:    http.StatusServiceUnavailable,
		Code:      "object_store_unavailable",
		Retryable: retryable,
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func objectUploadNotFoundOrRevoked() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "object_upload_not_found_or_revoked", Details: map[string]any{}}
}

func objectUploadExpired(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusGone, Code: "object_upload_expired", Details: map[string]any{"reason_code": reasonCode}}
}

// The object-upload rejections below give each REQ-01-246 status its one
// registered code (REQ-01-234).

// objectUploadRejected is the 409 for a capability whose bound contract no
// longer matches live slot state.
func objectUploadRejected(reasonCode string) *httpapi.APIError {
	return objectUploadRejection(http.StatusConflict, "object_upload_rejected", reasonCode, nil)
}

// objectUploadInvalid is the 400 for a request that does not honour the
// issued upload contract.
func objectUploadInvalid(reasonCode string, details map[string]any) *httpapi.APIError {
	return objectUploadRejection(http.StatusBadRequest, "object_upload_invalid", reasonCode, details)
}

func objectUploadLengthRequired() *httpapi.APIError {
	return objectUploadRejection(http.StatusLengthRequired, "object_upload_length_required", "content_length_required", nil)
}

// objectUploadTooLarge is the 413 for a declared length beyond the contracted
// byte_size.
func objectUploadTooLarge(details map[string]any) *httpapi.APIError {
	return objectUploadRejection(http.StatusRequestEntityTooLarge, "request_body_too_large", "byte_size_exceeds_contract", details)
}

func objectUploadRejection(status int, code, reasonCode string, details map[string]any) *httpapi.APIError {
	if details == nil {
		details = map[string]any{}
	}
	details["reason_code"] = reasonCode
	return &httpapi.APIError{Status: status, Code: code, Details: details}
}

func unavailableReason(adapterErr *objectstore.AdapterError) string {
	switch adapterErr.Reason {
	case objectstore.ReasonBucketMissing:
		return "bucket_missing"
	case objectstore.ReasonRetryExhausted, objectstore.ReasonDeadlineExceeded:
		return "retry_exhausted"
	default:
		return "endpoint_unreachable"
	}
}

func accessRejectedReason(adapterErr *objectstore.AdapterError) string {
	switch adapterErr.Reason {
	case objectstore.ReasonCredentialDenied:
		return "credential_denied"
	case objectstore.ReasonCORSRejected:
		return "cors_rejected"
	default:
		return "capability_missing"
	}
}

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
}

// recordDeletedUseRestore is REQ-01-705 step 2: a soft-deleted record must be
// restored before it can be mutated (REQ-01-211).
func recordDeletedUseRestore() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "record_deleted_use_restore", Message: "record deleted use restore", Details: map[string]any{}}
}

func evidenceRecordNotFound() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "evidence_record_not_found", Details: map[string]any{}}
}
