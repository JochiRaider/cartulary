package extensionassembly

import (
	"context"
	"errors"
	"sort"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
)

// IncidentPortabilityPolicies is the application composition edge from the
// immutable Extensions catalog and serving claim epoch to the Incident Bundles
// owner's local policy model.
func IncidentPortabilityPolicies(catalog []extensions.PortabilityPolicy, claims extensions.ResolvedClaimSet) []incidentbundles.ExtensionPolicy {
	claimed := make(map[string]struct{})
	for _, profileID := range claims.ProfileIDs() {
		claimed[profileID] = struct{}{}
	}
	result := make([]incidentbundles.ExtensionPolicy, len(catalog))
	for index, policy := range catalog {
		claimState := incidentbundles.ClaimStateRecognizedUnclaimable
		if policy.Claimable {
			claimState = incidentbundles.ClaimStateUnclaimed
			if _, ok := claimed[policy.ProfileID]; ok {
				claimState = incidentbundles.ClaimStateClaimed
			}
		}
		result[index] = incidentbundles.ExtensionPolicy{
			ProfileID: policy.ProfileID, ClaimState: claimState,
			ContractMajor: policy.ContractMajor, Mode: policy.Mode,
			ParticipantID: policy.ParticipantID, ParticipantSHA256: policy.ParticipantSHA256,
			ParticipantSchemaID: policy.ParticipantSchemaID,
			MaximumInputBytes:   policy.MaximumInputBytes, MaximumOutputBytes: policy.MaximumOutputBytes,
			AuthoritativeFamilyIDs: append([]string(nil), policy.AuthoritativeFamilyIDs...),
			BlockingFamilyIDs:      append([]string(nil), policy.BlockingFamilyIDs...),
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ProfileID < result[j].ProfileID })
	return result
}

// IncidentPortabilityStatePresence dispatches only admitted logical family
// identities to owner-provided declarative storage bindings.
type IncidentPortabilityStatePresence struct {
	networkFlow *networkflow.PortabilityStateBinding
}

func NewIncidentPortabilityStatePresence(networkFlow *networkflow.PortabilityStateBinding) (*IncidentPortabilityStatePresence, error) {
	if networkFlow == nil {
		return nil, errors.New("incident portability state presence binding incomplete")
	}
	return &IncidentPortabilityStatePresence{networkFlow: networkFlow}, nil
}

func (p *IncidentPortabilityStatePresence) AuthoritativeStatePresent(ctx context.Context, query incidentbundles.StatePresenceQuery, incidentID uuid.UUID, profileID string, familyIDs []string) (bool, error) {
	if p == nil || p.networkFlow == nil {
		return false, errors.New("incident portability state presence unavailable")
	}
	switch profileID {
	case networkflow.ProfileID:
		return p.networkFlow.RetainedAuthoritativeStatePresentTx(ctx, query, incidentID, familyIDs)
	default:
		return false, errors.New("incident portability state presence profile scope invalid")
	}
}

// IncidentPortabilityReads refines the same export snapshot and the admitted
// source members into owner-specific capabilities. No profile sees the bundle
// coordinator's transaction or another owner's members.
type IncidentPortabilityReads struct{}

func (IncidentPortabilityReads) ForSnapshot(query incidentbundles.ExportSnapshotQuery, incidentID uuid.UUID, profileID string) (incidentbundles.ExportReadCapability, error) {
	if profileID != networkflow.ProfileID {
		return nil, incidentbundles.ErrPortabilityUnavailable
	}
	return networkflow.NewPortabilityExportAccess(query, incidentID, incidentbundles.PortableSourceAttributions())
}
func (IncidentPortabilityReads) ForImport(ctx context.Context, input incidentbundles.PortabilityImportContext, members incidentportability.MemberReader, profileID string) (incidentbundles.ImportDependencyAccess, error) {
	if profileID != networkflow.ProfileID {
		return nil, incidentbundles.ErrPortabilityUnavailable
	}
	sources, err := imports.PrepareCommittedSources(ctx, members, input.IncidentID, input.Actors)
	if err != nil {
		return nil, err
	}
	indicatorReferences, err := indicators.NewPortableReferences(members, input.IncidentID)
	if err != nil {
		return nil, err
	}
	return networkflow.NewPortabilityDependencies(input.IncidentID, sources, indicatorReferences)
}
