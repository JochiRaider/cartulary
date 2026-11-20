package evidence

import (
	"context"

	evidenceblob "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/blob"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// Source metadata must describe the immutable member before staging begins.
// The staging capability separately verifies the actual bytes it consumes.
func validateIncidentBundleBlobObservations(ctx context.Context, rows, bundle incidentportability.MemberReader, descriptor sourceport.Descriptor) error {
	return incidentportability.WalkPath(ctx, rows, "data/object_blobs.ndjson", func(row map[string]any) error {
		if row["upload_state"] != "available" && row["upload_state"] != "quarantined" {
			return nil
		}
		size, sizeOK := evidenceblob.ParseSize(row["byte_size"])
		observedSize, observedSizeOK := evidenceblob.ParseSize(row["observed_size"])
		digest, digestOK := evidenceblob.ParseDigest(row["observed_sha256_hex"])
		expected, expectedPresent := row["expected_sha256_hex"]
		if !sizeOK || !observedSizeOK || size != observedSize || !digestOK || !expectedPresent {
			return descriptor.DeclaredFailure("evidence.byte_size_digest_agree")
		}
		if expected != nil {
			value, valid := evidenceblob.ParseDigest(expected)
			if !valid || value != digest {
				return descriptor.DeclaredFailure("evidence.byte_size_digest_agree")
			}
		}
		bound, err := incidentportability.BindMembers(bundle, []string{"blobs/sha256/" + digest})
		if err != nil {
			return &incidentportability.VerificationFailure{ReasonCode: "missing_required_blob"}
		}
		member := bound.Members()[0]
		if member.SHA256 != digest {
			return &incidentportability.VerificationFailure{ReasonCode: "blob_hash_mismatch"}
		}
		if member.ByteSize != size {
			return descriptor.DeclaredFailure("evidence.byte_size_digest_agree")
		}
		return nil
	})
}
