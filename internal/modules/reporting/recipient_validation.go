package reporting

import "strings"

// Validate against the policy retained by admission. A worker must not resolve
// the selector again and substitute a current registry policy during replay.
func validateRenderRecipientPartitions(releaseScope string, recipients []string, model SnapshotContent, profile RedactionProfile) *renderValidationError {
	if releaseScope != ReleaseScopeExternal {
		return nil
	}
	if len(recipients) == 0 {
		return newRenderValidationError("redaction_manifest_invalid", "recipient_partition_profile_mismatch")
	}
	knownPartyPartitions := map[string]struct{}{}
	refs, err := snapshotPartyPartitions(model)
	if err != nil {
		return newRenderValidationError("redaction_manifest_invalid", "unknown_recipient_partition")
	}
	for _, ref := range refs {
		knownPartyPartitions[ref] = struct{}{}
	}
	for _, ref := range recipients {
		if !validPartyPartitionRef(ref) {
			return newRenderValidationError("redaction_manifest_invalid", "invalid_recipient_partition_ref")
		}
		if _, ok := knownPartyPartitions[ref]; !ok {
			return newRenderValidationError("redaction_manifest_invalid", "unknown_recipient_partition")
		}
	}
	if !profileSupportsScope(RedactionProfileView{ProfileStatus: profile.ProfileStatus, AllowedReleaseScopes: profile.AllowedReleaseScopes}, releaseScope) {
		return newRenderValidationError("redaction_manifest_invalid", "recipient_partition_profile_mismatch")
	}
	profilePartyRefs := map[string]struct{}{}
	for _, ref := range profile.AllowedDisclosurePartitionRefs {
		if strings.HasPrefix(ref, "party:") {
			profilePartyRefs[ref] = struct{}{}
		}
	}
	if len(profilePartyRefs) != len(recipients) {
		return newRenderValidationError("redaction_manifest_invalid", "recipient_partition_profile_mismatch")
	}
	for _, ref := range recipients {
		if _, ok := profilePartyRefs[ref]; !ok {
			return newRenderValidationError("redaction_manifest_invalid", "recipient_partition_profile_mismatch")
		}
	}
	return nil
}
