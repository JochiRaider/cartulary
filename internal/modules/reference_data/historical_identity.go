package reference_data

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

// Restore authenticates the successful historical binding, including its
// original release sequence. Current rollback counters cannot invalidate a
// previously successful, pinned historical envelope.
type retainedVerificationIdentity struct {
	db                      retainedContentQuery
	storage                 VerificationStorage
	manifest                packformat.Manifest
	manifestSHA, payloadSHA string
}

// Historical validation checks the retained member inventory before invoking
// the container verifier. It does not consult mutable live health or attempt
// to repair the already authenticated historical content.
func (retainedVerificationIdentity) checkRetainedPresence(context.Context) error  { return nil }
func (retainedVerificationIdentity) checkRetainedLength(context.Context) error    { return nil }
func (retainedVerificationIdentity) checkRetainedIntegrity(context.Context) error { return nil }

func (r retainedVerificationIdentity) checkReleaseSequence(_ context.Context, m packformat.Manifest, _, _ string) error {
	if m.Key != r.manifest.Key || m.Version != r.manifest.Version || m.Sequence != r.manifest.Sequence {
		return errHistoricalIntegrity
	}
	return nil
}

func (r retainedVerificationIdentity) checkLogicalVersion(_ context.Context, _ packformat.Manifest, manifestSHA, payloadSHA string) error {
	if manifestSHA != r.manifestSHA || payloadSHA != r.payloadSHA {
		return errHistoricalIntegrity
	}
	return nil
}

func (r retainedVerificationIdentity) resolveDependency(ctx context.Context, d packformat.Dependency) (packformat.Manifest, bool, error) {
	return resolveRetainedDependency(ctx, r.db, r.storage, d, true)
}
