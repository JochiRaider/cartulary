package evidence

// portableBlobPath holds portable Object Blob rows. Their shape and lifecycle
// tuple are decided by the shared Blob authority (internal/blob), the same
// predicate live Blob locks verify (REQ-02-189a).
const portableBlobPath = "data/object_blobs.ndjson"
