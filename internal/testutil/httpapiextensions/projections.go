package httpapiextensions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"testing"

	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contractextensions"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type Projections struct {
	discovery  []httpapi.ExtensionProfile
	claims     []httpapi.ExtensionClaim
	routes     []httpapi.ExtensionRoute
	workspaces []httpapi.ExtensionWorkspacePublication
}

func New(profiles []httpapi.ExtensionProfile) Projections {
	projection := Projections{discovery: cloneProfiles(profiles)}
	for _, profile := range projection.discovery {
		projection.claims = append(projection.claims, httpapi.ExtensionClaim{
			ProfileID: profile.ProfileID,
			Claimed:   profile.Claimed,
		})
		for _, routeFamily := range profile.RouteFamilies {
			projection.routes = append(projection.routes, httpapi.ExtensionRoute{
				ProfileID:   profile.ProfileID,
				RouteFamily: routeFamily,
				Claimed:     profile.Claimed,
			})
		}
		if !profile.Claimed {
			continue
		}
		for _, workspace := range profile.Workspaces {
			projection.workspaces = append(projection.workspaces, httpapi.ExtensionWorkspacePublication{
				ProfileID:    profile.ProfileID,
				WorkspaceKey: workspace.WorkspaceKey,
				MinimumRole:  workspace.MinimumRole,
			})
		}
	}
	return projection
}

func FromGeneratedRegistry(t testing.TB, claimedProfileIDs ...string) Projections {
	t.Helper()
	artifact, ok := contractsgen.Index["contracts/extensions/generated/profile-registry.json"]
	if !ok {
		t.Fatal("generated extension profile registry is not packaged")
	}
	digest := sha256.Sum256([]byte(artifact.JSON))
	if hex.EncodeToString(digest[:]) != artifact.SHA256 {
		t.Fatal("generated extension profile registry digest mismatch")
	}
	var registry struct {
		SchemaID string `json:"schema_id"`
		Profiles []struct {
			ProfileID     string   `json:"profile_id"`
			Claimable     bool     `json:"claimable"`
			ContractMajor *int     `json:"contract_major"`
			RouteFamilies []string `json:"route_families"`
			WorkspaceKeys []string `json:"workspace_keys"`
			CapabilityIDs []string `json:"capability_ids"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(artifact.JSON), &registry); err != nil ||
		registry.SchemaID != "cartulary.extension_profile_registry.v1" {
		t.Fatal("generated extension profile registry is invalid")
	}

	claimed := make(map[string]struct{}, len(claimedProfileIDs))
	for _, profileID := range claimedProfileIDs {
		if profileID == "" {
			t.Fatal("claimed extension profile fixture ID must not be empty")
		}
		if _, duplicate := claimed[profileID]; duplicate {
			t.Fatalf("duplicate claimed extension profile fixture ID %q", profileID)
		}
		claimed[profileID] = struct{}{}
	}

	profiles := make([]httpapi.ExtensionProfile, 0, len(registry.Profiles))
	for _, descriptor := range registry.Profiles {
		_, isClaimed := claimed[descriptor.ProfileID]
		delete(claimed, descriptor.ProfileID)
		workspaces := make([]httpapi.ExtensionWorkspace, 0, len(descriptor.WorkspaceKeys))
		for _, workspaceKey := range descriptor.WorkspaceKeys {
			workspaces = append(workspaces, httpapi.ExtensionWorkspace{
				WorkspaceKey: workspaceKey,
				MinimumRole:  "viewer",
			})
		}
		profiles = append(profiles, httpapi.ExtensionProfile{
			ProfileID:     descriptor.ProfileID,
			Claimable:     descriptor.Claimable,
			Claimed:       isClaimed,
			ContractMajor: cloneInt(descriptor.ContractMajor),
			RouteFamilies: append([]string(nil), descriptor.RouteFamilies...),
			WorkspaceKeys: append([]string(nil), descriptor.WorkspaceKeys...),
			Capabilities:  append([]string(nil), descriptor.CapabilityIDs...),
			Workspaces:    workspaces,
		})
	}
	if len(claimed) != 0 {
		unknown := make([]string, 0, len(claimed))
		for profileID := range claimed {
			unknown = append(unknown, profileID)
		}
		sort.Strings(unknown)
		t.Fatalf("unknown claimed extension profile fixture IDs %v", unknown)
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ProfileID < profiles[j].ProfileID })
	return New(profiles)
}

func (p Projections) Dependencies(deps httpapi.DependencySet) httpapi.DependencySet {
	deps.ExtensionDiscovery = p
	deps.ExtensionClaims = p
	deps.ExtensionRoutes = p
	deps.ExtensionWorkspaces = p
	return deps
}

func (p Projections) ExtensionDiscoveryProfiles() []httpapi.ExtensionProfile {
	return cloneProfiles(p.discovery)
}

func (p Projections) ExtensionClaims() []httpapi.ExtensionClaim {
	return append([]httpapi.ExtensionClaim(nil), p.claims...)
}

func (p Projections) ExtensionRoutes() []httpapi.ExtensionRoute {
	return append([]httpapi.ExtensionRoute(nil), p.routes...)
}

func (p Projections) ExtensionWorkspaces() []httpapi.ExtensionWorkspacePublication {
	return append([]httpapi.ExtensionWorkspacePublication(nil), p.workspaces...)
}

func cloneProfiles(profiles []httpapi.ExtensionProfile) []httpapi.ExtensionProfile {
	cloned := make([]httpapi.ExtensionProfile, 0, len(profiles))
	for _, profile := range profiles {
		cloned = append(cloned, httpapi.ExtensionProfile{
			ProfileID:     profile.ProfileID,
			Claimable:     profile.Claimable,
			Claimed:       profile.Claimed,
			ContractMajor: cloneInt(profile.ContractMajor),
			RouteFamilies: append([]string(nil), profile.RouteFamilies...),
			WorkspaceKeys: append([]string(nil), profile.WorkspaceKeys...),
			Capabilities:  append([]string(nil), profile.Capabilities...),
			Workspaces:    append([]httpapi.ExtensionWorkspace(nil), profile.Workspaces...),
		})
	}
	return cloned
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
