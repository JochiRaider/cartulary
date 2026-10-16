package reporting

import "sort"

// Party membership comes from frozen source-record identity. Display strings,
// authored profile permissions and field partition claims cannot create a Party.
func snapshotPartyPartitions(snapshot SnapshotContent) ([]string, error) {
	refs := []string{}
	seen := map[string]bool{}
	for _, record := range snapshot.Records {
		if record.RecordType != "party" || record.SourceRecordRef.SourceFamily != "party" {
			continue
		}
		ref := "party:" + record.SourceRecordRef.SourceRecordID
		if !validPartyPartitionRef(ref) || seen[ref] {
			return nil, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
		}
		seen[ref] = true
		refs = append(refs, ref)
	}
	return orderedDisclosurePartitions(refs), nil
}

// Effective permissions are a bound on a profile, never supplied by it. In
// particular, no profile can admit blocked content or an unselected recipient.
func effectiveDisclosurePartitions(snapshot SnapshotContent, scope string, recipients []string) ([]string, error) {
	parties, err := snapshotPartyPartitions(snapshot)
	if err != nil {
		return nil, err
	}
	switch scope {
	case ReleaseScopeInternalDraft, ReleaseScopeInternalReview:
		if len(recipients) != 0 {
			return nil, newRenderValidationError("invalid_release_request", "recipient_partitions_not_allowed")
		}
		return orderedDisclosurePartitions(append([]string{"public", "internal_only"}, parties...)), nil
	case ReleaseScopeExternal:
		known, seen := map[string]bool{}, map[string]bool{}
		for _, ref := range parties {
			known[ref] = true
		}
		for _, ref := range recipients {
			if !validPartyPartitionRef(ref) || seen[ref] {
				return nil, newRenderValidationError("invalid_release_request", "invalid_recipient_partition_ref")
			}
			if !known[ref] {
				return nil, newRenderValidationError("invalid_release_request", "unknown_recipient_partition")
			}
			seen[ref] = true
		}
		return orderedDisclosurePartitions(append([]string{"public"}, recipients...)), nil
	default:
		return nil, newRenderValidationError("invalid_release_request", "unsupported_release_scope")
	}
}

func disclosurePartitionOrder(ref string) int {
	switch ref {
	case "public":
		return 0
	case "internal_only":
		return 90
	case "blocked":
		return 99
	default:
		if validPartyPartitionRef(ref) {
			return 10
		}
		return -1
	}
}

// Do not trim or rewrite identifiers. The caller still validates the closed
// vocabulary; ordering never turns an invalid ref into an admitted one.
func orderedDisclosurePartitions(refs []string) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, ref := range refs {
		if !seen[ref] {
			result = append(result, ref)
			seen[ref] = true
		}
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := disclosurePartitionOrder(result[i]), disclosurePartitionOrder(result[j])
		if left != right {
			return left < right
		}
		return result[i] < result[j]
	})
	return result
}

func validateDisclosurePartitionSet(refs []string) error {
	previousOrder, previous := -1, ""
	for _, ref := range refs {
		order := disclosurePartitionOrder(ref)
		if order < 0 || order < previousOrder || order == previousOrder && ref <= previous {
			return profilePolicyError()
		}
		previousOrder, previous = order, ref
	}
	return nil
}
