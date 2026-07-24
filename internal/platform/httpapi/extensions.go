package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"sync"

	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

type ExtensionProfile struct {
	ProfileID     string               `json:"profile_id"`
	Claimable     bool                 `json:"claimable"`
	Claimed       bool                 `json:"claimed"`
	ContractMajor *int                 `json:"contract_major"`
	RouteFamilies []string             `json:"route_families"`
	WorkspaceKeys []string             `json:"workspace_keys"`
	Capabilities  []string             `json:"capabilities"`
	Workspaces    []ExtensionWorkspace `json:"-"`
}

// ExtensionWorkspace is application-composition authorization metadata. Public
// discovery emits only its workspace key.
type ExtensionWorkspace struct {
	WorkspaceKey string
	MinimumRole  string
}

type ExtensionClaim struct {
	ProfileID string
	Claimed   bool
}

type ExtensionRoute struct {
	ProfileID   string
	RouteFamily string
	Claimed     bool
}

type ExtensionWorkspacePublication struct {
	ProfileID    string
	WorkspaceKey string
	MinimumRole  string
}

type ReservedExtensionMatch struct {
	ProfileID   string
	Claimed     bool
	RouteFamily string
}

// ExtensionEpochProvider is the read-only platform view of the application-owned
// serving epoch. Production composition supplies a provider backed by the
// installed immutable publication plan.
type ExtensionEpochProvider interface {
	ExtensionProfiles() []ExtensionProfile
}

type ExtensionDiscoveryProvider interface {
	ExtensionDiscoveryProfiles() []ExtensionProfile
}

type ExtensionClaimProvider interface {
	ExtensionClaims() []ExtensionClaim
}

type ExtensionRouteProvider interface {
	ExtensionRoutes() []ExtensionRoute
}

type ExtensionWorkspaceProvider interface {
	ExtensionWorkspaces() []ExtensionWorkspacePublication
}

type StaticExtensionEpochProvider struct {
	profiles []ExtensionProfile
}

func NewStaticExtensionEpochProvider(profiles []ExtensionProfile) StaticExtensionEpochProvider {
	return StaticExtensionEpochProvider{profiles: cloneExtensionProfiles(profiles)}
}

func (p StaticExtensionEpochProvider) ExtensionProfiles() []ExtensionProfile {
	return cloneExtensionProfiles(p.profiles)
}

func (p StaticExtensionEpochProvider) ExtensionDiscoveryProfiles() []ExtensionProfile {
	return cloneExtensionProfiles(p.profiles)
}

func (p StaticExtensionEpochProvider) ExtensionClaims() []ExtensionClaim {
	claims := make([]ExtensionClaim, 0, len(p.profiles))
	for _, profile := range p.profiles {
		claims = append(claims, ExtensionClaim{ProfileID: profile.ProfileID, Claimed: profile.Claimed})
	}
	return claims
}

func (p StaticExtensionEpochProvider) ExtensionRoutes() []ExtensionRoute {
	routes := []ExtensionRoute{}
	for _, profile := range p.profiles {
		for _, routeFamily := range profile.RouteFamilies {
			routes = append(routes, ExtensionRoute{
				ProfileID: profile.ProfileID, RouteFamily: routeFamily, Claimed: profile.Claimed,
			})
		}
	}
	return routes
}

func (p StaticExtensionEpochProvider) ExtensionWorkspaces() []ExtensionWorkspacePublication {
	workspaces := []ExtensionWorkspacePublication{}
	for _, profile := range p.profiles {
		if !profile.Claimed {
			continue
		}
		for _, workspace := range profile.Workspaces {
			workspaces = append(workspaces, ExtensionWorkspacePublication{
				ProfileID: profile.ProfileID, WorkspaceKey: workspace.WorkspaceKey, MinimumRole: workspace.MinimumRole,
			})
		}
	}
	return workspaces
}

func ExtensionProfilesFromEpoch(provider ExtensionEpochProvider) []ExtensionProfile {
	if provider == nil {
		return nil
	}
	return cloneExtensionProfiles(provider.ExtensionProfiles())
}

func ExtensionProfileClaimedBy(provider ExtensionEpochProvider, profileID string) bool {
	return ExtensionProfileClaimedIn(ExtensionProfilesFromEpoch(provider), profileID)
}

func ExtensionProfileClaimedInProjection(claims []ExtensionClaim, profileID string) bool {
	for _, claim := range claims {
		if claim.ProfileID == profileID {
			return claim.Claimed
		}
	}
	return false
}

var (
	extensionProfilesMu      = sync.RWMutex{}
	currentProfileExtensions = mustGeneratedExtensionProfiles()
)

func mustGeneratedExtensionProfiles() []ExtensionProfile {
	artifact, ok := contractsgen.ExtensionArtifactsIndex["contracts/extensions/generated/profile-registry.json"]
	if !ok {
		panic("generated extension profile registry is not packaged")
	}
	digest := sha256.Sum256([]byte(artifact.JSON))
	if hex.EncodeToString(digest[:]) != artifact.SHA256 {
		panic("generated extension profile registry digest mismatch")
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
	if err := json.Unmarshal([]byte(artifact.JSON), &registry); err != nil || registry.SchemaID != "cartulary.extension_profile_registry.v1" {
		panic("generated extension profile registry is invalid")
	}
	profiles := make([]ExtensionProfile, 0, len(registry.Profiles))
	for _, descriptor := range registry.Profiles {
		claimed := descriptor.ProfileID != "enterprise_authentication" && descriptor.ProfileID != "network_flow_activity"
		workspaces := make([]ExtensionWorkspace, 0, len(descriptor.WorkspaceKeys))
		for _, workspaceKey := range descriptor.WorkspaceKeys {
			workspaces = append(workspaces, ExtensionWorkspace{WorkspaceKey: workspaceKey, MinimumRole: "viewer"})
		}
		profiles = append(profiles, ExtensionProfile{
			ProfileID: descriptor.ProfileID, Claimable: descriptor.Claimable, Claimed: claimed,
			ContractMajor: cloneInt(descriptor.ContractMajor), RouteFamilies: descriptor.RouteFamilies,
			WorkspaceKeys: descriptor.WorkspaceKeys, Capabilities: descriptor.CapabilityIDs, Workspaces: workspaces,
		})
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ProfileID < profiles[j].ProfileID })
	return profiles
}

func CurrentExtensionProfiles() []ExtensionProfile {
	extensionProfilesMu.RLock()
	defer extensionProfilesMu.RUnlock()
	return cloneExtensionProfiles(currentProfileExtensions)
}

func ResolveExtensionProfiles(profiles []ExtensionProfile) []ExtensionProfile {
	if profiles == nil {
		return CurrentExtensionProfiles()
	}
	return cloneExtensionProfiles(profiles)
}

func ExtensionProfileClaimed(profileID string) bool {
	extensionProfilesMu.RLock()
	defer extensionProfilesMu.RUnlock()
	for _, profile := range currentProfileExtensions {
		if profile.ProfileID == profileID {
			return profile.Claimed
		}
	}
	return false
}

func ExtensionProfileClaimedIn(profiles []ExtensionProfile, profileID string) bool {
	for _, profile := range ResolveExtensionProfiles(profiles) {
		if profile.ProfileID == profileID {
			return profile.Claimed
		}
	}
	return false
}

func MatchReservedExtensionFamily(path string) (ReservedExtensionMatch, bool) {
	extensionProfilesMu.RLock()
	defer extensionProfilesMu.RUnlock()
	for _, profile := range currentProfileExtensions {
		for _, routeFamily := range profile.RouteFamilies {
			if routeFamilyMatchesPath(routeFamily, path) {
				return ReservedExtensionMatch{
					ProfileID:   profile.ProfileID,
					Claimed:     profile.Claimed,
					RouteFamily: routeFamily,
				}, true
			}
		}
	}
	return ReservedExtensionMatch{}, false
}

func MatchReservedExtensionFamilyIn(profiles []ExtensionProfile, path string) (ReservedExtensionMatch, bool) {
	for _, profile := range ResolveExtensionProfiles(profiles) {
		for _, routeFamily := range profile.RouteFamilies {
			if routeFamilyMatchesPath(routeFamily, path) {
				return ReservedExtensionMatch{
					ProfileID:   profile.ProfileID,
					Claimed:     profile.Claimed,
					RouteFamily: routeFamily,
				}, true
			}
		}
	}
	return ReservedExtensionMatch{}, false
}

func MatchReservedExtensionRouteIn(routes []ExtensionRoute, path string) (ReservedExtensionMatch, bool) {
	for _, route := range routes {
		if routeFamilyMatchesPath(route.RouteFamily, path) {
			return ReservedExtensionMatch{
				ProfileID: route.ProfileID, Claimed: route.Claimed, RouteFamily: route.RouteFamily,
			}, true
		}
	}
	return ReservedExtensionMatch{}, false
}

func SetCurrentExtensionProfilesForTesting(profiles []ExtensionProfile) func() {
	extensionProfilesMu.Lock()
	previous := cloneExtensionProfiles(currentProfileExtensions)
	currentProfileExtensions = cloneExtensionProfiles(profiles)
	extensionProfilesMu.Unlock()

	return func() {
		extensionProfilesMu.Lock()
		currentProfileExtensions = previous
		extensionProfilesMu.Unlock()
	}
}

func routeFamilyMatchesPath(routeFamily string, path string) bool {
	pathSegments := splitRouteSegments(path)
	familySegments := splitRouteSegments(routeFamily)
	if len(pathSegments) < len(familySegments) {
		return false
	}

	for index := range familySegments {
		familySegment := familySegments[index]
		if isRouteTemplateSegment(familySegment) {
			if pathSegments[index] == "" {
				return false
			}
			continue
		}
		if familySegment != pathSegments[index] {
			return false
		}
	}

	return true
}

func splitRouteSegments(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func isRouteTemplateSegment(value string) bool {
	return strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")
}

func cloneExtensionProfiles(profiles []ExtensionProfile) []ExtensionProfile {
	cloned := make([]ExtensionProfile, 0, len(profiles))
	for _, profile := range profiles {
		cloned = append(cloned, ExtensionProfile{
			ProfileID:     profile.ProfileID,
			Claimable:     profile.Claimable,
			Claimed:       profile.Claimed,
			ContractMajor: cloneInt(profile.ContractMajor),
			RouteFamilies: append([]string(nil), profile.RouteFamilies...),
			WorkspaceKeys: append([]string(nil), profile.WorkspaceKeys...),
			Capabilities:  append([]string(nil), profile.Capabilities...),
			Workspaces:    append([]ExtensionWorkspace(nil), profile.Workspaces...),
		})
	}
	return cloned
}

func cloneExtensionClaims(claims []ExtensionClaim) []ExtensionClaim {
	return append([]ExtensionClaim(nil), claims...)
}

func cloneExtensionRoutes(routes []ExtensionRoute) []ExtensionRoute {
	return append([]ExtensionRoute(nil), routes...)
}

func cloneExtensionWorkspaces(workspaces []ExtensionWorkspacePublication) []ExtensionWorkspacePublication {
	return append([]ExtensionWorkspacePublication(nil), workspaces...)
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}
