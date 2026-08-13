package evidence

import (
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// These bridges exist only in the package's test variant. They let the
// black-box integration suite exercise owner-private capabilities without
// restoring production constructors or teaching shared application support
// how to reconstruct a partial Evidence topology.
type BlobLifecycleDependencies = blobLifecycleDependencies
type BlobLifecycleService = blobLifecycleService
type RouteOperations = routeOperations
type AttachBlobRequest = attachBlobRequest
type AttachRecordChange = attachRecordChange
type BlobCreateRequest = blobCreateRequest
type BlobSlotParams = blobSlotParams
type BlobSlotResult = blobSlotResult
type CleanupSweepResult = cleanupSweepResult
type CleanupObjectDeleter = cleanupObjectDeleter
type EvidenceAccessRecord = evidenceAccessRecord
type HandleRecord = handleRecord
type ObservedObject = observedObject
type UploadLeaseCreateParams = uploadLeaseCreateParams
type WorkbookCreateParams = createParams
type WorkbookLifecyclePatchChange = lifecyclePatchChange

var AttachBlobRequestHash = attachBlobRequestHash
var BlobCreateRequestHash = blobCreateRequestHash
var DecodeAttachBlobRequest = decodeAttachBlobRequest
var DecodeBlobCreateRequest = decodeBlobCreateRequest
var DecodeHandleIssueRequest = decodeHandleIssueRequest
var ValidateWorkbookCreateParams = validateCreateParams

func NewBlobLifecycleService(dependencies blobLifecycleDependencies) (*blobLifecycleService, error) {
	return newBlobLifecycleService(dependencies)
}

func NewAccessHandleService(pool postgres.DB) (*accessHandleService, error) {
	return newAccessHandleService(pool)
}

func NewCleanupService(pool postgres.DB) (*cleanupService, error) {
	return newCleanupService(pool)
}

func NewRouteOperations(blobs *blobLifecycleService, access *accessHandleService) (*routeOperations, error) {
	return newRouteOperations(blobs, access)
}

func NewCleanupObjectDeleter(store objectstore.TypedStore) (cleanupObjectDeleter, error) {
	return newCleanupObjectDeleter(store)
}

func NewCleanupDispatcher(
	sweeper cleanupSweeper,
	deleter cleanupObjectDeleter,
	observer CleanupObserver,
	now func() time.Time,
) (*CleanupDispatcher, error) {
	return newCleanupDispatcher(sweeper, deleter, observer, now)
}
