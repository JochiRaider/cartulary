package evidence

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

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
		size, sizeOK := portableBlobSize(row["byte_size"])
		observedSize, observedSizeOK := portableBlobSize(row["observed_size"])
		digest, digestOK := portableBlobDigest(row["observed_sha256_hex"])
		expected, expectedPresent := row["expected_sha256_hex"]
		if !sizeOK || !observedSizeOK || size != observedSize || !digestOK || !expectedPresent {
			return descriptor.DeclaredFailure("evidence.byte_size_digest_agree")
		}
		if expected != nil {
			value, valid := portableBlobDigest(expected)
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

func portableBlobSize(raw any) (int64, bool) {
	number, ok := raw.(json.Number)
	if !ok {
		return 0, false
	}
	value, err := number.Int64()
	return value, err == nil && value >= 0 && strconv.FormatInt(value, 10) == number.String()
}

func portableBlobDigest(raw any) (string, bool) {
	value, ok := raw.(string)
	if !ok || len(value) != 64 || strings.ToLower(value) != value {
		return "", false
	}
	_, err := hex.DecodeString(value)
	return value, err == nil
}
