package startup

import (
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// WorkspaceResolver is the narrow startup seam for extension-owned surfaces.
// It resolves only claim, declaration, and shell visibility; extension data
// stores are intentionally outside this boundary.
type WorkspaceResolver interface {
	ResolveWorkspace(ref SheetRef, role string) string
}

type WorkspaceRegistry struct {
	profiles map[string]workspaceProfile
}

type workspaceProfile struct {
	claimed    bool
	workspaces map[string]string
}

func NewWorkspaceRegistry(profiles []httpapi.ExtensionProfile) *WorkspaceRegistry {
	registry := &WorkspaceRegistry{profiles: map[string]workspaceProfile{}}
	for _, profile := range httpapi.ResolveExtensionProfiles(profiles) {
		profileID := strings.TrimSpace(profile.ProfileID)
		if profileID == "" {
			continue
		}
		entry := workspaceProfile{
			claimed:    profile.Claimed,
			workspaces: map[string]string{},
		}
		for _, workspace := range profile.Workspaces {
			workspaceKey := strings.TrimSpace(workspace.WorkspaceKey)
			minimumRole := strings.TrimSpace(workspace.MinimumRole)
			if workspaceKey == "" || minimumRole == "" {
				continue
			}
			entry.workspaces[workspaceKey] = minimumRole
		}
		registry.profiles[profileID] = entry
	}
	return registry
}

func (r *WorkspaceRegistry) ResolveWorkspace(ref SheetRef, role string) string {
	if !validExtensionToken(ref.ExtensionProfileID) {
		return "invalid_extension_profile_id"
	}
	if !validExtensionToken(ref.WorkspaceKey) {
		return "invalid_extension_workspace_key"
	}
	if r == nil {
		return "extension_profile_not_claimed"
	}
	profile, ok := r.profiles[ref.ExtensionProfileID]
	if !ok || !profile.claimed {
		return "extension_profile_not_claimed"
	}
	minimumRole, ok := profile.workspaces[ref.WorkspaceKey]
	if !ok {
		return "extension_workspace_unavailable"
	}
	if !roleAtLeast(role, minimumRole) {
		return "extension_workspace_not_visible"
	}
	return ""
}

func roleAtLeast(role string, minimumRole string) bool {
	rank := map[string]int{
		"viewer":   1,
		"editor":   2,
		"reviewer": 3,
		"admin":    4,
	}
	actual, actualOK := rank[strings.TrimSpace(role)]
	minimum, minimumOK := rank[strings.TrimSpace(minimumRole)]
	return actualOK && minimumOK && actual >= minimum
}
