package reference_data

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/jackc/pgx/v5"
)

const (
	PortableReferenceSetLimit                   = 1024
	IncidentPackContentPath                     = packformat.PortableContentPath
	IncidentBundleReferenceDegradationInvariant = "reference_pack_refs.degradation_bounded"
	IncidentBundleReferenceExactShapeInvariant  = "reference_pack_refs.exact_shape"
	IncidentBundleReferenceIdentityInvariant    = "reference_pack_refs.identity_exact"
)

type IncidentBundleReferenceValidationError struct {
	InvariantID string
}

func (e *IncidentBundleReferenceValidationError) Error() string {
	return "incident bundle reference-pack references violate " + e.InvariantID
}

func IncidentBundleReferenceInvariant(err error) (string, bool) {
	var validationErr *IncidentBundleReferenceValidationError
	if !errors.As(err, &validationErr) {
		return "", false
	}
	return validationErr.InvariantID, true
}

// ValidateIncidentBundleReferences is the Reference Pack owner's closed,
// non-mutating validator for the Incident Bundle reference catalog.

func DecodeIncidentBundleReferences(payload []byte) (IncidentBundleReferences, error) {
	refs, err := packformat.DecodePortableReferences(payload)
	if err != nil {
		var failure *packformat.Failure
		if errors.As(err, &failure) {
			return IncidentBundleReferences{}, &IncidentBundleReferenceValidationError{InvariantID: failure.Code}
		}
		return IncidentBundleReferences{}, err
	}
	return portableReferencesFromFormat(refs), nil
}

func ValidateIncidentBundleReferences(payload []byte) error {
	_, err := DecodeIncidentBundleReferences(payload)
	return err
}

func EncodeIncidentBundleReferences(sets []PackSet, versions []IncidentBundleVersionReference) ([]byte, error) {
	refs := (IncidentBundleReferences{Sets: sets, Versions: versions}).format()
	return packformat.EncodePortableReferences(refs.Sets, refs.Versions)
}

// ExportReferencesTx materializes the exact reference graph selected by the
// artifact owners in the caller's consistent export snapshot. No current-set
// lookup is performed. Duplicate uses of the same set or version become one
// catalog row; a missing or altered retained binding fails explicitly.
func (r *retention) ExportReferencesTx(ctx context.Context, tx pgx.Tx, setIDs []string) ([]byte, error) {
	if tx == nil {
		return nil, consumerError("invalid_pack_request")
	}
	ids := slices.Clone(setIDs)
	for _, id := range ids {
		if !setIDPattern.MatchString(id) {
			return nil, consumerError("invalid_pack_request")
		}
	}
	slices.Sort(ids)
	ids = slices.Compact(ids)
	repository := canonicalRepository{pool: tx, storage: r.repository.storage}
	sets := make([]PackSet, 0, len(ids))
	versions := []IncidentBundleVersionReference{}
	seen := map[string]bool{}
	for _, id := range ids {
		set, err := repository.RetainedSet(ctx, id)
		if err != nil {
			return nil, err
		}
		sets = append(sets, set)
		for _, member := range set.Members {
			identity := member.Key + "\x00" + member.Version
			if seen[identity] {
				continue
			}
			anchor, err := repository.Provenance(ctx, set.ID, member.Key)
			if err != nil {
				return nil, err
			}
			distribution := "operator_imported"
			if anchor.VerificationMethod == "packaged_release_manifest_v1" {
				distribution = "packaged_builtin"
			}
			versions = append(versions, IncidentBundleVersionReference{PackSetMember: member, DistributionKind: distribution, VerificationMethod: anchor.VerificationMethod, SourceProfileID: anchor.SourceProfileID, SourceProfileSHA256: anchor.SourceProfileSHA256})
			seen[identity] = true
		}
	}
	return EncodeIncidentBundleReferences(sets, versions)
}

// ValidatePortableBinding checks historical source evidence against an admitted
// catalog. It does not compare source verification to destination trust or grant
// permission to read local content. Destination retention owns that decision.
func ValidatePortableBinding(refs IncidentBundleReferences, binding SetBinding) error {
	invalid := &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceIdentityInvariant}
	for _, set := range refs.Sets {
		if set.ID != binding.SetID {
			continue
		}
		if set.SHA256 != binding.SHA256 {
			return invalid
		}
		data, err := canonicaljson.Marshal(binding.Provenance)
		if err != nil || packformat.ValidateProvenance(data, set.format()) != nil {
			return invalid
		}
		versions := make(map[string]IncidentBundleVersionReference, len(refs.Versions))
		for _, version := range refs.Versions {
			versions[version.Key+"\x00"+version.Version] = version
		}
		for _, anchor := range binding.Provenance {
			version, ok := versions[anchor.PackKey+"\x00"+anchor.PackVersion]
			if !ok || anchor.SourceProfileID != version.SourceProfileID || anchor.SourceProfileSHA256 != version.SourceProfileSHA256 || anchor.VerificationMethod != version.VerificationMethod {
				return invalid
			}
		}
		return nil
	}
	return invalid
}
