package reference_data

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/referencepacks/format/dataset"
	"github.com/jackc/pgx/v5"
)

// Descriptor reads the exact signed member for this immutable logical tuple.
// The consumer has already checked complete version availability; this read
// additionally binds the bytes it returns, including a concurrent storage loss.
func (r *canonicalRepository) Descriptor(ctx context.Context, member PackSetMember) (json.RawMessage, error) {
	if member.ProfileID != dataset.ProfileID {
		return nil, nil
	}
	var manifestBytes []byte
	var reference, digest string
	var size int64
	err := r.pool.QueryRow(ctx, `SELECT v.manifest_bytes,o.storage_ref,o.sha256,o.size_bytes
 FROM reference_pack_versions v
 JOIN reference_pack_object_refs f ON f.owner_kind='version' AND f.owner_id=$5 AND f.logical_path='payload/descriptor.json'
 JOIN reference_pack_objects o USING(object_id)
 WHERE v.pack_key=$1 AND v.pack_version=$2 AND v.manifest_sha256=$3 AND v.payload_sha256=$4 AND o.available`, member.Key, member.Version, member.ManifestSHA256, member.PayloadSHA256, versionObjectID(member.Key, member.Version)).Scan(&manifestBytes, &reference, &digest, &size)
	unavailable := func() (json.RawMessage, error) {
		if r.invalidate != nil {
			if err := r.invalidate(ctx, member.Key, member.Version); err != nil {
				return nil, errors.Join(consumerError("pack_unavailable"), err)
			}
		}
		return nil, consumerError("pack_unavailable")
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return unavailable()
	}
	if err != nil {
		return nil, err
	}
	if packformat.Digest(manifestBytes) != member.ManifestSHA256 || size < 0 || size > dataset.MaxDescriptorBytes {
		return unavailable()
	}
	m, err := packformat.DecodeManifestSchema(manifestBytes, true)
	if err != nil {
		return unavailable()
	}
	bound := false
	for _, file := range m.Files {
		if file.Path == dataset.DescriptorPath && file.Size == size && file.SHA256 == digest {
			bound = true
		}
	}
	if !bound || m.Key != member.Key || m.Version != member.Version || m.ProfileID != dataset.ProfileID {
		return unavailable()
	}
	ref, err := ParseStorageRef(reference)
	if err != nil {
		return nil, err
	}
	reader, actual, err := r.storage.OpenPublished(ctx, ref)
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, ErrArtifactUnavailable) {
		return unavailable()
	}
	if err != nil {
		return nil, err
	}
	if actual != size {
		if err := reader.Close(); err != nil {
			return nil, err
		}
		return unavailable()
	}
	raw, readErr := io.ReadAll(&attemptReader{ctx: ctx, source: io.NewSectionReader(reader, 0, size)})
	closeErr := reader.Close()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if readErr != nil || closeErr != nil {
		return nil, errors.Join(readErr, closeErr)
	}
	if int64(len(raw)) != size || packformat.Digest(raw) != digest {
		return unavailable()
	}
	if _, err := dataset.DecodeDescriptor(raw); err != nil {
		return unavailable()
	}
	return json.RawMessage(raw), nil
}
