package incidentportability

// PreparedVerdicts records bundle-only failures that an owner's precedence
// ranks after a database-side check (Core 01 REQ-01-641). Preparation records
// them; the final transaction reports At(rank) at that rank. For each rank it
// keeps the failure with the smallest owner-supplied selection key, so the
// reported failure is independent of archive or row order (REQ-01-642).
type PreparedVerdicts struct {
	byRank map[int]preparedVerdict
}

type preparedVerdict struct {
	key     string
	failure error
}

// Record keeps failure for rank unless a failure with a smaller key exists.
func (v *PreparedVerdicts) Record(rank int, key string, failure error) {
	if v == nil || failure == nil {
		return
	}
	if v.byRank == nil {
		v.byRank = map[int]preparedVerdict{}
	}
	if existing, ok := v.byRank[rank]; ok && existing.key <= key {
		return
	}
	v.byRank[rank] = preparedVerdict{key: key, failure: failure}
}

// At returns the recorded failure for rank, or nil.
func (v *PreparedVerdicts) At(rank int) error {
	if v == nil {
		return nil
	}
	return v.byRank[rank].failure
}

// Empty reports whether preparation recorded no deferred verdict.
func (v *PreparedVerdicts) Empty() bool { return v == nil || len(v.byRank) == 0 }
