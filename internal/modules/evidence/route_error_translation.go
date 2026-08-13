package evidence

import (
	"net/http"

	"github.com/google/uuid"

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

func pathUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	value, err := uuid.Parse(r.PathValue(key))
	if err != nil {
		http.NotFound(w, r)
		return uuid.UUID{}, false
	}
	return value, true
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusInternalServerError, Code: "internal_error", Message: err.Error(), Details: map[string]any{}}
}

func PersistedObjectBlobStorageKeyErrorReason(err error) (string, bool) {
	return evidencepolicy.PersistedObjectBlobStorageKeyErrorReason(err)
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

func objectUploadRejected(status int, reasonCode string, details map[string]any) *httpapi.APIError {
	if details == nil {
		details = map[string]any{}
	}
	details["reason_code"] = reasonCode
	return &httpapi.APIError{Status: status, Code: "object_upload_rejected", Details: details}
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

func incidentNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
}

func evidenceRecordNotFound() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "evidence_record_not_found", Details: map[string]any{}}
}
