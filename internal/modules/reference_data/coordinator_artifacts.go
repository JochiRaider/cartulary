package reference_data

import (
	"context"
	"errors"
	"io"
	"slices"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// retainAttemptObject records ownership immediately after an immutable byte
// object is complete. These preparation references are backup/cleanup inputs,
// never evidence that content has passed or become consumer-visible.
func retainAttemptObject(ctx context.Context, pool postgres.DB, storage ArtifactStorage, operationID, attemptID uuid.UUID, path, digest string, size int64, reader io.Reader) (preparedObject, error) {
	ref, lease, err := storage.PublishStream(ctx, digest, size, reader)
	if err != nil {
		return preparedObject{}, err
	}
	defer lease.Close()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return preparedObject{}, errors.Join(err, storage.RemovePublished(ref))
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockByteRetentionTx(ctx, tx, false); err != nil {
		return preparedObject{}, errors.Join(err, storage.RemovePublished(ref))
	}
	object := preparedObject{ID: uuid.New(), Path: path, Digest: digest, Size: size, Reference: ref}
	if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_objects(object_id,sha256,storage_ref,size_bytes,generation) VALUES($1,$2,$3,$4,1)`, object.ID, digest, ref.String(), size); err != nil {
		return preparedObject{}, errors.Join(err, storage.RemovePublished(ref))
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_object_refs(owner_kind,owner_id,logical_path,object_id) VALUES('operation',$1,$2,$3)`, operationID.String(), attemptID.String()+"/"+path, object.ID); err != nil {
		return preparedObject{}, errors.Join(err, storage.RemovePublished(ref))
	}
	if err := tx.Commit(ctx); err != nil {
		// An acknowledgement failure can leave this preparation reference
		// committed. Retain bytes until a later authoritative inventory sweep.
		return preparedObject{}, err
	}
	return object, nil
}

func prepareVerifiedObjects(ctx context.Context, pool postgres.DB, storage ArtifactStorage, operationID, attemptID uuid.UUID, attempt verificationAttempt, content *verifiedContent, indexID uuid.UUID) (preparedVersion, error) {
	p := preparedVersion{Content: content, IndexID: indexID}
	paths := []string{"manifest.json"}
	for _, file := range content.Manifest.Files {
		paths = append(paths, file.Path)
	}
	slices.Sort(paths)
	for _, path := range paths {
		member, exists := content.Inventory[path]
		if !exists {
			return p, errors.New("reference pack: verified inventory changed")
		}
		reader, err := content.workspace.Open(ctx, path)
		if err != nil {
			return p, err
		}
		object, writeErr := retainAttemptObject(ctx, pool, storage, operationID, attemptID, versionObjectID(content.Manifest.Key, content.Manifest.Version)+"/"+path, member.SHA256, member.Size, reader)
		closeErr := reader.Close()
		if writeErr != nil || closeErr != nil {
			return p, errors.Join(writeErr, closeErr)
		}
		object.Path = path
		p.Objects = append(p.Objects, object)
	}
	if attempt.Retained == nil {
		return p, errors.New("reference pack: successful input is not retained")
	}
	object := preparedObject{Path: "container", Reference: *attempt.Retained}
	// The immutable input was fully hashed by verification and is already
	// protected by an operation or successful envelope. Reuse its identity;
	// renewing an envelope does not duplicate a potentially large container.
	if err := pool.QueryRow(ctx, `SELECT object_id,sha256,size_bytes FROM reference_pack_objects WHERE storage_ref=$1 AND available`, attempt.Retained.String()).Scan(&object.ID, &object.Digest, &object.Size); err != nil {
		return p, err
	}
	if object.Digest != content.ContainerSHA256 || object.Size != content.ContainerBytes {
		return p, errors.New("reference pack: retained container binding mismatch")
	}
	p.Objects = append(p.Objects, object)
	reference := object.Reference.String()
	digest := content.ContainerSHA256
	expiry := content.Trust.ValidUntil
	snapshot := content.TrustSnapshot
	p.Envelope = successfulEnvelope{SchemaID: "cartulary.reference_pack_successful_envelope.v1", OperationID: operationID, PackKey: content.Manifest.Key, PackVersion: content.Manifest.Version, DistributionKind: "operator_imported", ManifestSHA256: content.ManifestSHA256, PayloadSHA256: content.PayloadSHA256, ContainerSHA256: &digest, ContainerRef: &reference, VerifiedAt: content.VerifiedAt, TrustValidUntil: &expiry, TrustSnapshot: &snapshot, TrustProposal: &content.Trust}
	return p, nil
}
