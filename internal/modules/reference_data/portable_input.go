package reference_data

import (
	"context"
	"errors"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

type portableContainerInput struct {
	descriptor packformat.PortableContainer
	identity   importIdentity
	object     preparedObject
}

var errPortableContainerLength = errors.New("reference pack: portable container length mismatch")

// Transport denial establishes no content verdict. Only a safely attributable,
// eligible container becomes a durable verification candidate. A returned
// object is private until the coordinator retains it with the frozen cohort.
func (r *incidentReferences) preparePortableContainer(ctx context.Context, request IncidentReferenceImportRequest, descriptor packformat.PortableContainer, reference IncidentBundleVersionReference) (result *portableContainerInput, resultErr error) {
	if descriptor.SizeBytes < 1 || descriptor.ManifestSHA256 != reference.ManifestSHA256 || !isLowerSHA256(descriptor.ContainerSHA256) {
		return nil, portableIdentityError()
	}
	if descriptor.SizeBytes > r.verifier.limits.ReferencePacks.MaxContainerBytes {
		return nil, nil
	}
	if request.OpenContainer == nil {
		return nil, errors.New("reference pack: portable container reader required")
	}
	path, err := packformat.PortableContainerPath(descriptor.ManifestSHA256)
	if err != nil {
		return nil, err
	}
	source, err := request.OpenContainer(ctx, path)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, errors.New("reference pack: missing portable container stream")
	}
	input := &portableLengthReader{source: source, remaining: descriptor.SizeBytes}
	identity, object, prepareErr := r.verifier.prepareImportInput(ctx, input)
	closeErr := source.Close()
	if prepareErr != nil {
		if closeErr != nil {
			return nil, closeErr
		}
		if errors.Is(prepareErr, errPortableContainerLength) {
			return nil, portableIdentityError()
		}
		return nil, prepareErr
	}
	keep := false
	defer func() {
		if !keep {
			if err := errors.Join(r.storage.RemovePublished(object.Reference), object.releasePublication()); err != nil {
				result = nil
				resultErr = err
			}
		}
	}()
	if closeErr != nil {
		return nil, closeErr
	}
	if object.Size != descriptor.SizeBytes || object.Digest != descriptor.ContainerSHA256 {
		return nil, portableIdentityError()
	}
	if identity.Key == "" {
		return nil, nil
	}
	if identity.Key != reference.Key || identity.Version != reference.Version {
		return nil, portableIdentityError()
	}
	if identity.Manifest != nil {
		m := *identity.Manifest
		payload, err := packformat.PayloadDigest(m.Files)
		if identity.ManifestSHA256 != reference.ManifestSHA256 || reference.DistributionKind != "operator_imported" || reference.VerificationMethod != "tuf_1_0_35_offline_bundle_v1" ||
			reference.Contract != m.Contract || reference.ProfileID != m.ProfileID || reference.ProfileVersion != m.ProfileVersion || reference.SourceProfileID != m.SourceProfileID || reference.SourceProfileSHA256 != m.SourceProfileSHA256 || err == nil && payload != reference.PayloadSHA256 {
			return nil, portableIdentityError()
		}
		if !portableEmbeddingAllowed(m) {
			return nil, nil
		}
	}
	// Cohort scheduling needs only the bounded dependency declarations; retaining
	// every full lexical manifest would multiply the per-container working set.
	identity.Manifest = nil
	keep = true
	return &portableContainerInput{descriptor: descriptor, identity: identity, object: object}, nil
}

func portableIdentityError() error {
	return &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceIdentityInvariant}
}

// The descriptor is an exact transport binding, not permission to consume an
// unlimited source. Probe one byte past it without forwarding that byte to the
// staging writer. Underlying I/O errors stay operational.
type portableLengthReader struct {
	source    io.Reader
	remaining int64
}

func (r *portableLengthReader) Read(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	if r.remaining == 0 {
		var extra [1]byte
		n, err := r.source.Read(extra[:])
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n > 0 {
			return 0, errPortableContainerLength
		}
		return 0, err
	}
	if int64(len(data)) > r.remaining {
		data = data[:r.remaining]
	}
	n, err := r.source.Read(data)
	r.remaining -= int64(n)
	return n, err
}
