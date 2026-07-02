package extensions

import "github.com/JochiRaider/cartulary/internal/platform/httpapi"

func BuildResource(profile httpapi.ExtensionProfile) map[string]any {
	return map[string]any{
		"profile_id":     profile.ProfileID,
		"claimed":        profile.Claimed,
		"route_families": append([]string(nil), profile.RouteFamilies...),
	}
}

func BuildResponseData(profiles []httpapi.ExtensionProfile) map[string]any {
	extensions := make([]map[string]any, 0, len(profiles))
	for _, profile := range profiles {
		extensions = append(extensions, BuildResource(profile))
	}
	return map[string]any{"extensions": extensions}
}
