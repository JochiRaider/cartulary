package reference_data

import "crypto/sha256"

const (
	ProfileID = "reference_pack"

	PacksRouteContributionID = "reference_pack.packs_route"
	LifecycleWorkerKind      = "reference_pack.lifecycle_worker_v2"

	ImportJobKind     = "reference_pack.import_v2"
	ReverifyJobKind   = "reference_pack.reverify_v2"
	RefreshJobKind    = "reference_pack.refresh_v2"
	ImportOperation   = "reference_pack.import"
	ReverifyOperation = "reference_pack.reverify"
	RefreshOperation  = "reference_pack.refresh"

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
}

func optionalString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func hashBytes(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

func referencePackRoute(packKey string, packVersion string) string {
	return "/api/v1/reference-packs/" + packKey + "/" + packVersion
}

// RequestRejection identifies an invalid semantic request without selecting a transport response.
type RequestRejection struct{ Field, Reason string }

func (r *RequestRejection) Error() string { return "reference pack request: " + r.Reason }

// ActivationRejection retains the distinct activation failure contract.
type ActivationRejection struct{ Reason string }

func (r *ActivationRejection) Error() string { return "reference pack activation: " + r.Reason }
