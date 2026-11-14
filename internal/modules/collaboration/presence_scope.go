package collaboration

import (
	"errors"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

// ClaimedWorkspace is projected by application assembly from the admitted
// extension publication plan. Collaboration does not resolve profile claims.
type ClaimedWorkspace struct {
	ExtensionProfileID string
	WorkspaceKey       string
}

type claimedWorkspaces map[ClaimedWorkspace]struct{}

func newClaimedWorkspaces(inputs []ClaimedWorkspace) (claimedWorkspaces, error) {
	result := make(claimedWorkspaces, len(inputs))
	for _, input := range inputs {
		if input.ExtensionProfileID == "" || input.WorkspaceKey == "" || strings.TrimSpace(input.ExtensionProfileID) != input.ExtensionProfileID || strings.TrimSpace(input.WorkspaceKey) != input.WorkspaceKey {
			return nil, errors.New("collaboration claimed workspace requires exact nonempty tokens")
		}
		if _, duplicate := result[input]; duplicate {
			return nil, errors.New("collaboration claimed workspace is duplicated")
		}
		result[input] = struct{}{}
	}
	return result, nil
}

func (scopes claimedWorkspaces) validate(input protocol.PresenceInput) error {
	if input.SheetRef["kind"] != "extension_workspace" {
		return nil
	}
	key := ClaimedWorkspace{ExtensionProfileID: input.SheetRef["extension_profile_id"], WorkspaceKey: input.SheetRef["workspace_key"]}
	if _, claimed := scopes[key]; !claimed {
		return errors.New("presence extension workspace is not claimed")
	}
	return nil
}
