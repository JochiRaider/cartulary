package httpapi

import (
	"strings"
	"sync"
)

type ExtensionProfile struct {
	ProfileID     string   `json:"profile_id"`
	Claimed       bool     `json:"claimed"`
	RouteFamilies []string `json:"route_families"`
}

type ReservedExtensionMatch struct {
	ProfileID   string
	Claimed     bool
	RouteFamily string
}

var (
	extensionProfilesMu      = sync.RWMutex{}
	currentProfileExtensions = []ExtensionProfile{
		{
			ProfileID: "enterprise_authentication",
			Claimed:   false,
			RouteFamilies: []string{
				"/api/v1/auth/oidc",
				"/api/v1/auth/providers",
				"/api/v1/auth/saml",
				"/api/v1/users/{user_id}/auth-bindings",
			},
		},
		{
			ProfileID: "import",
			Claimed:   true,
			RouteFamilies: []string{
				"/api/v1/import-sessions",
			},
		},
		{
			ProfileID: "incident_portability",
			Claimed:   true,
			RouteFamilies: []string{
				"/api/v1/incident-bundles",
			},
		},
		{
			ProfileID: "reference_pack",
			Claimed:   true,
			RouteFamilies: []string{
				"/api/v1/reference-packs",
			},
		},
		{
			ProfileID: "snapshot_reporting",
			Claimed:   true,
			RouteFamilies: []string{
				"/api/v1/releases",
				"/api/v1/snapshots",
			},
		},
	}
)

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
			Claimed:       profile.Claimed,
			RouteFamilies: append([]string(nil), profile.RouteFamilies...),
		})
	}
	return cloned
}
