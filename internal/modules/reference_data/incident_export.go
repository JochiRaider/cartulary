package reference_data

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// IncidentReferenceExportRequest supplies an already selected exact catalog.
// The sink receives only a format-owned bundle path and a bounded stream, never
// storage locators. Its bytes remain private until the outer export succeeds.
type IncidentReferenceExportRequest struct {
	IncidentID     uuid.UUID
	References     []byte
	Embed          bool
	WriteContainer func(context.Context, string, int64, io.Reader) error
}

func (r *incidentReferences) ExportContentTx(ctx context.Context, tx pgx.Tx, request IncidentReferenceExportRequest) ([]byte, error) {
	if tx == nil || request.IncidentID == uuid.Nil || request.Embed && request.WriteContainer == nil {
		return nil, errors.New("reference pack: invalid content export")
	}
	refs, err := DecodeIncidentBundleReferences(request.References)
	if err != nil {
		return nil, err
	}
	content, err := r.retainedContentTx(ctx, tx, request.IncidentID)
	if err != nil {
		return nil, err
	}
	// Discard source descriptors. Only bytes actually admitted and written by
	// this export can receive an embedded-container inventory row.
	content.Containers = []packformat.PortableContainer{}
	if request.Embed {
		for _, version := range refs.Versions {
			if version.DistributionKind != "operator_imported" {
				continue
			}
			container, err := r.exportContainerTx(ctx, tx, version, request.WriteContainer)
			if err != nil {
				return nil, err
			}
			if container != nil {
				content.Containers = append(content.Containers, *container)
			}
		}
	}
	if len(content.Containers) == 0 && len(content.RequiredMembers) == 0 {
		return nil, nil
	}
	return packformat.EncodePortableContent(content, refs.format())
}

func (r *incidentReferences) exportContainerTx(ctx context.Context, tx pgx.Tx, reference IncidentBundleVersionReference, write func(context.Context, string, int64, io.Reader) error) (result *packformat.PortableContainer, resultErr error) {
	var manifestBytes, envelopeBytes []byte
	var manifestDigest, payloadDigest string
	err := tx.QueryRow(ctx, `SELECT v.manifest_sha256,v.payload_sha256,v.manifest_bytes,e.canonical_envelope
 FROM reference_pack_versions v JOIN reference_pack_candidates c USING(pack_key,pack_version)
 JOIN reference_pack_envelopes e ON e.envelope_id=c.current_envelope_id
 WHERE v.pack_key=$1 AND v.pack_version=$2`, reference.Key, reference.Version).Scan(&manifestDigest, &payloadDigest, &manifestBytes, &envelopeBytes)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if manifestDigest != reference.ManifestSHA256 || payloadDigest != reference.PayloadSHA256 || packformat.Digest(manifestBytes) != manifestDigest {
		return nil, errHistoricalIntegrity
	}
	envelope, err := decodeSuccessfulEnvelope(envelopeBytes)
	if err != nil {
		return nil, errHistoricalIntegrity
	}
	manifest, err := packformat.DecodeManifest(manifestBytes, true)
	if err != nil {
		return nil, errHistoricalIntegrity
	}
	if envelope.PackKey != reference.Key || envelope.PackVersion != reference.Version || envelope.DistributionKind != "operator_imported" || envelope.ManifestSHA256 != manifestDigest || envelope.PayloadSHA256 != payloadDigest ||
		manifest.Key != reference.Key || manifest.Version != reference.Version || manifest.Contract != reference.Contract || manifest.ProfileID != reference.ProfileID || manifest.ProfileVersion != reference.ProfileVersion ||
		manifest.SourceProfileID != reference.SourceProfileID || manifest.SourceProfileSHA256 != reference.SourceProfileSHA256 {
		return nil, errHistoricalIntegrity
	}
	if !portableEmbeddingAllowed(manifest) {
		return nil, nil
	}
	if envelope.ContainerRef == nil || envelope.ContainerSHA256 == nil {
		return nil, errHistoricalIntegrity
	}
	var expectedSize int64
	var expectedDigest string
	var available bool
	if err := tx.QueryRow(ctx, `SELECT o.size_bytes,o.sha256,o.available FROM reference_pack_objects o JOIN reference_pack_object_refs r USING(object_id) WHERE r.owner_kind='envelope' AND r.owner_id=$2 AND r.logical_path='container' AND o.storage_ref=$1`, *envelope.ContainerRef, "rpenv_"+packformat.Digest(envelopeBytes)).Scan(&expectedSize, &expectedDigest, &available); err != nil {
		return nil, err
	}
	if expectedDigest != *envelope.ContainerSHA256 || expectedSize < 1 {
		return nil, errHistoricalIntegrity
	}
	if !available {
		return nil, nil
	}
	ref, err := ParseStorageRef(*envelope.ContainerRef)
	if err != nil {
		return nil, errHistoricalIntegrity
	}
	source, size, err := r.storage.OpenPublished(ctx, ref)
	if ctx.Err() != nil {
		if source != nil {
			_ = source.Close()
		}
		return nil, ctx.Err()
	}
	if errors.Is(err, ErrArtifactUnavailable) || errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, errors.New("reference pack: invalid export reader")
	}
	defer func() {
		if closeErr := source.Close(); closeErr != nil {
			result = nil
			resultErr = closeErr
		}
	}()
	if size != expectedSize {
		return nil, nil
	}
	// First pass classifies unavailable optional bytes before invoking the
	// sink. A second hash prevents changed bytes from entering a successful
	// export between preflight and transport copying.
	digest := sha256.New()
	n, err := io.CopyBuffer(digest, &attemptReader{ctx: ctx, source: io.NewSectionReader(source, 0, size)}, make([]byte, 32768))
	if err != nil {
		return nil, err
	}
	if n != size || hex.EncodeToString(digest.Sum(nil)) != expectedDigest {
		return nil, nil
	}
	path, err := packformat.PortableContainerPath(reference.ManifestSHA256)
	if err != nil {
		return nil, err
	}
	digest.Reset()
	reader := io.NewSectionReader(source, 0, size)
	if err := write(ctx, path, size, io.TeeReader(&attemptReader{ctx: ctx, source: reader}, digest)); err != nil {
		return nil, err
	}
	consumed, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	if consumed != size || hex.EncodeToString(digest.Sum(nil)) != expectedDigest {
		return nil, errors.New("reference pack: incomplete or changed export stream")
	}
	return &packformat.PortableContainer{ManifestSHA256: manifestDigest, ContainerSHA256: expectedDigest, SizeBytes: size}, nil
}

func portableEmbeddingAllowed(manifest packformat.Manifest) bool {
	return manifest.License.Redistribution == "allowed"
}
