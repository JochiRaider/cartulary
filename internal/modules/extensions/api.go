package extensions

import "github.com/JochiRaider/cartulary/internal/platform/httpapi"

func BuildResource(profile httpapi.ExtensionProfile) map[string]any {
	var contractMajor any
	if profile.ContractMajor != nil {
		contractMajor = *profile.ContractMajor
	}
	return map[string]any{
		"profile_id":     profile.ProfileID,
		"claimable":      profile.Claimable,
		"claimed":        profile.Claimed,
		"contract_major": contractMajor,
		"route_families": presentStrings(profile.RouteFamilies),
		"workspace_keys": presentStrings(profile.WorkspaceKeys),
		"capabilities":   presentStrings(profile.Capabilities),
	}
}

func presentStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func BuildResponseData(profiles []httpapi.ExtensionProfile) map[string]any {
	extensions := make([]map[string]any, 0, len(profiles))
	for _, profile := range profiles {
		extensions = append(extensions, BuildResource(profile))
	}
	return map[string]any{"extensions": extensions}
}
