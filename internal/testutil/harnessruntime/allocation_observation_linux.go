//go:build linux

package harnessruntime

import (
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"maps"
	"regexp"
	"slices"
)

type allocationLifecycleSection struct {
	records            map[string]json.RawMessage
	registered         map[string]bool
	omitted            map[string]bool
	bytes, omissions   int
	maximumRefs        int
	partial, truncated bool
}

func newAllocationLifecycleSection() allocationLifecycleSection {
	return allocationLifecycleSection{records: map[string]json.RawMessage{}, registered: map[string]bool{}, omitted: map[string]bool{}}
}

var allocationObservationRef = regexp.MustCompile(`^[A-Za-z0-9_.:/-]{1,255}$`)

// Registration precedes acquisition and is independent of verbose records.
// Only the owner may supply a terminal observation; publication never invents it.
func (s *executionSession) beginLifecycle(ref string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.enabled || s.published {
		return
	}
	l := &s.lifecycle
	if l.records == nil {
		*l = newAllocationLifecycleSection()
	}
	l.maximumRefs = s.limit("maximum_allocation_lifecycle_records")
	if !allocationObservationRef.MatchString(ref) {
		l.partial = true
		l.omissions++
		return
	}
	if l.registered[ref] || l.omitted[ref] {
		return
	}
	if len(l.registered) >= s.limit("maximum_allocation_lifecycle_records") {
		l.truncated = true
		l.omissions++
		if len(l.omitted) < s.limit("maximum_allocation_lifecycle_records") {
			l.omitted[ref] = true
		}
		return
	}
	l.registered[ref] = true
}

func (s *executionSession) recordLifecycle(value *harnessfixtures.AllocationLifecycle) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.enabled || s.published {
		return
	}
	l := &s.lifecycle
	if l.records == nil {
		*l = newAllocationLifecycleSection()
	}
	if value == nil {
		l.partial = true
		return
	}
	ref := value.AllocationRef
	if !l.registered[ref] {
		// Quota-rejected registrations already account for their terminal record.
		if !l.truncated {
			l.partial = true
			l.omissions++
		}
		return
	}
	if l.omitted[ref] {
		return
	}
	encoded, err := documentBytes("cartulary.harness_allocation_lifecycle.v1", value)
	if err != nil || len(encoded) > s.limit("maximum_record_bytes") {
		l.partial = true
		l.omissions++
		if l.records[ref] == nil {
			l.omitted[ref] = true
		}
		return
	}
	if previous, exists := l.records[ref]; exists {
		if string(previous) != string(encoded) {
			l.partial = true
			l.omissions++
		}
		return
	}
	if len(l.records) >= s.limit("maximum_allocation_lifecycle_records") || l.bytes+len(encoded) > s.limit("maximum_allocation_lifecycle_bytes")-s.limit("maximum_allocation_lifecycle_records")*260-4096 {
		l.truncated = true
		l.omissions++
		if len(l.omitted) < s.limit("maximum_allocation_lifecycle_records") {
			l.omitted[ref] = true
		}
		return
	}
	l.records[ref] = encoded
	l.bytes += len(encoded)
}
func (l *allocationLifecycleSection) document() map[string]any {
	records := []json.RawMessage{}
	for _, ref := range slices.Sorted(maps.Keys(l.records)) {
		records = append(records, l.records[ref])
	}
	omitted := maps.Clone(l.omitted)
	omissions := l.omissions
	for ref := range l.registered {
		if l.records[ref] == nil && !omitted[ref] {
			omitted[ref] = true
			omissions++
		}
	}
	status := "complete"
	if l.partial || omissions > 0 {
		status = "partial"
	}
	if l.truncated {
		status = "truncated"
	}
	encoded, _ := json.Marshal(records)
	// Omission identities have their own metadata reservation within the same
	// section. Truncation keeps the count, even when not every identity fits.
	refs := slices.Sorted(maps.Keys(omitted))
	if len(refs) > l.maximumRefs {
		refs = refs[:l.maximumRefs]
	}
	return map[string]any{"status": status, "omitted_records": omissions, "omitted_refs": append([]string{}, refs...), "records": records, "records_digest": bytesDigest(encoded)}
}
