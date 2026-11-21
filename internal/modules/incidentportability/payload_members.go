package incidentportability

import (
	"context"
	"errors"
	"io"
)

// ErrPayloadNotAdmitted reports an open of a bundle member that preparation
// did not admit as an opaque payload, or whose descriptor differs.
var ErrPayloadNotAdmitted = errors.New("incident portability payload was not admitted by preparation")

// PayloadMembers is the opaque-payload view of a sealed bundle (Core 01
// REQ-01-641). While preparing, an owner admits a member whose digest it
// verified; preparation copies no payload bytes. Afterwards the owner opens
// only admitted payloads: each open streams the sealed bundle member through
// a reader that verifies the seal and hashes the bytes once. It is never a
// row reader and re-parses nothing.
type PayloadMembers interface {
	AdmitPayload(ctx context.Context, path string) (Member, error)
	OpenPayload(ctx context.Context, member Member) (io.ReadCloser, error)
}
