package incidentbundle

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type blobPortability struct {
	store   objectstore.TypedStore
	staging stagedobjects.ResourceAllocator
}

var ErrBlobMetadata = errors.New("finalized blob metadata does not describe exported bytes")

func NewBlobPortability(store objectstore.TypedStore, staging stagedobjects.ResourceAllocator) *blobPortability {
	return &blobPortability{store: store, staging: staging}
}

func (portability *blobPortability) ExportBlobFiles(
	ctx context.Context,
	queryer incidentportability.Queryer,
	incidentID uuid.UUID,
	members incidentportability.MemberWriter,
) error {
	rows, err := queryer.Query(ctx, `
SELECT storage_key, observed_sha256_hex, byte_size, observed_size, expected_sha256_hex
  FROM object_blobs
 WHERE incident_id = $1
   AND upload_state IN ('available', 'quarantined')
 ORDER BY observed_sha256_hex
`, incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	previousDigest := ""
	var previousSize int64
	for rows.Next() {
		var storageKey string
		var observedDigest, expectedDigest *string
		var size int64
		var observedSize *int64
		if err := rows.Scan(&storageKey, &observedDigest, &size, &observedSize, &expectedDigest); err != nil {
			return err
		}
		if observedDigest == nil || len(*observedDigest) != 64 || strings.ToLower(*observedDigest) != *observedDigest ||
			size < 0 || observedSize == nil || *observedSize != size ||
			(expectedDigest != nil && *expectedDigest != *observedDigest) {
			return ErrBlobMetadata
		}
		sha := *observedDigest
		if _, err := hex.DecodeString(sha); err != nil {
			return ErrBlobMetadata
		}
		if sha == previousDigest {
			if size != previousSize {
				return ErrBlobMetadata
			}
			continue
		}
		reader, _, err := portability.store.Get(ctx, objectstore.GetObjectRequest{
			Key:     storageKey,
			Purpose: objectstore.PurposeMigrationCopy,
		})
		if err != nil {
			return &incidentportability.VerificationFailure{ReasonCode: "missing_required_blob"}
		}
		member, writeErr := members.WriteMember(ctx, "blobs/sha256/"+sha, reader)
		closeErr := reader.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		if member.SHA256 != sha {
			return &incidentportability.VerificationFailure{ReasonCode: "blob_hash_mismatch"}
		}
		if member.ByteSize != size {
			return ErrBlobMetadata
		}
		previousDigest = sha
		previousSize = member.ByteSize
	}
	return rows.Err()
}

func (portability *blobPortability) RewriteAndStageObjectBlobs(
	ctx context.Context,
	members incidentportability.MemberReader,
	destination incidentportability.MemberWriter,
	incidentID uuid.UUID,
	operationID string,
) (prepared *stagedobjects.PreparedResources, failure error) {
	scope, err := stagedobjects.NewScope(operationID, "evidence", portability.staging)
	if err != nil {
		return nil, err
	}
	defer func() {
		if prepared == nil {
			failure = errors.Join(failure, scope.Abandon(context.WithoutCancel(ctx)))
		}
	}()
	publications := []stagedobjects.Publication{}
	err = incidentportability.WriteMemberFrom(ctx, destination, "data/object_blobs.ndjson", func(ctx context.Context, output io.Writer) error {
		return incidentportability.WalkPath(ctx, members, "data/object_blobs.ndjson", func(row map[string]any) error {
			objectBlobText, _ := row["object_blob_id"].(string)
			objectBlobID, err := uuid.Parse(objectBlobText)
			if err != nil {
				return &incidentportability.VerificationFailure{ReasonCode: "malformed_manifest"}
			}
			storageKey, err := blobref.ObjectBlobStorageKey(incidentID, objectBlobID, uuid.New())
			if err != nil {
				return &incidentportability.VerificationFailure{ReasonCode: "malformed_manifest"}
			}
			row["storage_key"] = storageKey
			state, _ := row["upload_state"].(string)
			if state == "available" || state == "quarantined" {
				sha, _ := row["observed_sha256_hex"].(string)
				bound, err := incidentportability.BindMembers(members, []string{"blobs/sha256/" + sha})
				if err != nil {
					return &incidentportability.VerificationFailure{ReasonCode: "missing_required_blob"}
				}
				member := bound.Members()[0]
				if member.SHA256 != sha {
					return &incidentportability.VerificationFailure{ReasonCode: "blob_hash_mismatch"}
				}
				reader, err := incidentportability.OpenVerifiedMember(ctx, bound, member)
				if err != nil {
					return err
				}
				resource := stagedobjects.ResourceBinding{Kind: "object_blob", ID: incidentID.String() + ":" + objectBlobID.String()}
				sealed, identity, writeErr := scope.AllocateResourceStream(ctx, operationID, resource, reader, member.ByteSize)
				if writeErr != nil {
					_ = reader.Close()
					return writeErr
				}
				if sealed.ByteSize != member.ByteSize || sealed.SHA256 != sha {
					_ = reader.Close()
					return &incidentportability.VerificationFailure{ReasonCode: "blob_hash_mismatch"}
				}
				row["storage_key"] = identity
				publications = append(publications, stagedobjects.Publication{StagingID: sealed.StagingID, ResourceKind: resource.Kind, ResourceID: resource.ID, ByteSize: sealed.ByteSize, SHA256: sealed.SHA256})
				var probe [1]byte
				n, readErr := reader.Read(probe[:])
				closeErr := reader.Close()
				if n != 0 || readErr != io.EOF {
					return &incidentportability.VerificationFailure{ReasonCode: "blob_hash_mismatch"}
				}
				if closeErr != nil {
					return closeErr
				}
			}
			return incidentportability.WriteNDJSONRecord(output, row)
		})
	})
	if err != nil {
		return nil, err
	}
	return scope.PrepareResources(publications)
}
