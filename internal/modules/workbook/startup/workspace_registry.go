package startup

import (
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// WorkspaceResolver is the narrow startup seam for extension-owned surfaces.
// It resolves only claim, declaration, and shell visibility; extension data
// stores are intentionally outside this boundary.
type WorkspaceResolver interface {
	ResolveWorkspace(ref SheetRef, role string) string
	AvailableWorkspaces(role string) []ExtensionWorkspaceAvailabilityRow
}

const ExtensionWorkspaceAvailabilitySchemaID = "cartulary.extension_workspace_availability.v1"

type ExtensionWorkspaceAvailabilityRow struct {
	ExtensionProfileID string `json:"extension_profile_id"`
	WorkspaceKey       string `json:"workspace_key"`
}

type ExtensionWorkspaceAvailability struct {
	SchemaID   string                              `json:"schema_id"`
	IncidentID string                              `json:"incident_id"`
	Workspaces []ExtensionWorkspaceAvailabilityRow `json:"workspaces"`
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

func NewWorkspaceRegistryFromPublication(workspaces []httpapi.ExtensionWorkspacePublication) *WorkspaceRegistry {
	registry := &WorkspaceRegistry{profiles: map[string]workspaceProfile{}}
	for _, workspace := range workspaces {
		profileID := strings.TrimSpace(workspace.ProfileID)
		workspaceKey := strings.TrimSpace(workspace.WorkspaceKey)
		minimumRole := strings.TrimSpace(workspace.MinimumRole)
		if profileID == "" || workspaceKey == "" || minimumRole == "" {
			continue
		}
		entry := registry.profiles[profileID]
		if entry.workspaces == nil {
			entry = workspaceProfile{claimed: true, workspaces: map[string]string{}}
		}
		entry.workspaces[workspaceKey] = minimumRole
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

func (r *WorkspaceRegistry) AvailableWorkspaces(role string) []ExtensionWorkspaceAvailabilityRow {
	if r == nil {
		return []ExtensionWorkspaceAvailabilityRow{}
	}
	rows := make([]ExtensionWorkspaceAvailabilityRow, 0)
	for profileID, profile := range r.profiles {
		if !profile.claimed {
			continue
		}
		for workspaceKey, minimumRole := range profile.workspaces {
			if roleAtLeast(role, minimumRole) {
				rows = append(rows, ExtensionWorkspaceAvailabilityRow{
					ExtensionProfileID: profileID,
					WorkspaceKey:       workspaceKey,
				})
			}
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ExtensionProfileID != rows[j].ExtensionProfileID {
			return rows[i].ExtensionProfileID < rows[j].ExtensionProfileID
		}
		return rows[i].WorkspaceKey < rows[j].WorkspaceKey
	})
	return rows
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
