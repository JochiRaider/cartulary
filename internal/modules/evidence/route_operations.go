package evidence

import "errors"

// routeOperations is the narrow transport facade composed from the blob and
// access capabilities. It does not expose cleanup or source mutation.
type routeOperations struct {
	*blobLifecycleService
	*accessHandleService
}

func newRouteOperations(blobs *blobLifecycleService, access *accessHandleService) (*routeOperations, error) {
	if blobs == nil {
		return nil, errors.New("compose Evidence routes: blob lifecycle is required")
	}
	if access == nil {
		return nil, errors.New("compose Evidence routes: access handles are required")
	}
	return &routeOperations{blobLifecycleService: blobs, accessHandleService: access}, nil
}
