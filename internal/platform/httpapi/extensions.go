package httpapi

import (
	"strings"
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

func ExtensionProfileClaimedInProjection(claims []ExtensionClaim, profileID string) bool {
	for _, claim := range claims {
		if claim.ProfileID == profileID {
			return claim.Claimed
		}
	}
	return false
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
