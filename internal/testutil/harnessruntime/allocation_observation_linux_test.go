//go:build linux

package harnessruntime

import (
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"testing"
)

func testAllocationObservationBounds(t *testing.T) {
	t.Run("allocation_diagnostics_have_independent_bounds", func(t *testing.T) {
		s := &executionSession{enabled: true, truncated: true, lifecycle: newAllocationLifecycleSection(), policy: map[string]any{"maximum_record_bytes": float64(65536), "maximum_allocation_lifecycle_records": float64(1), "maximum_allocation_lifecycle_bytes": float64(4194304)}}
		v := harnessfixtures.AllocationLifecycle{AllocationRef: "allocation:a", Borrows: []harnessfixtures.AllocationBorrow{}, ClockIdentity: "clock", Acquisition: harnessfixtures.AllocationInterval{Start: 0, End: 1, Duration: 1}, Retirement: harnessfixtures.AllocationInterval{Start: 3, End: 5, Duration: 2}, Reason: "run_finalization", Settlement: "settled", Coverage: "complete"}
		s.beginLifecycle(v.AllocationRef)
		s.recordLifecycle(&v)
		s.recordLifecycle(&v)
		if s.lifecycle.document()["status"] != "complete" || len(s.lifecycle.records) != 1 {
			t.Fatal("verbose truncation or repetition changed lifecycle")
		}
		v.Settlement = "unresolved"
		s.recordLifecycle(&v)
		if !s.lifecycle.partial || s.lifecycle.omissions != 1 {
			t.Fatal("conflicting terminal observation accepted")
		}
		v.AllocationRef = "allocation:b"
		s.beginLifecycle(v.AllocationRef)
		s.recordLifecycle(&v)
		if !s.lifecycle.truncated || !s.lifecycle.omitted[v.AllocationRef] {
			t.Fatal("missing explicit allocation omission")
		}
		s.published = true
		v.AllocationRef = "allocation:c"
		s.recordLifecycle(&v)
		if s.lifecycle.omissions != 2 {
			t.Fatal("publication reopened")
		}
	})
	t.Run("missing_terminal_survives_verbose_truncation", func(t *testing.T) {
		s := &executionSession{enabled: true, records: map[string]map[string]any{}, missing: map[string]bool{}, lifecycle: newAllocationLifecycleSection(), policy: map[string]any{"maximum_record_bytes": float64(65536), "maximum_execution_records": float64(1), "maximum_metadata_bytes": float64(524288), "maximum_allocation_lifecycle_records": float64(2), "maximum_allocation_lifecycle_bytes": float64(4194304)}}
		v := harnessfixtures.Observation{Kind: "allocation", Ref: "allocation:missing", AllocationRef: "allocation:missing", UnitID: "unit:a", Capability: "browser_stack", Outcome: "incomplete"}
		s.fixture(v)
		s.fixture(v)
		if !s.truncated || len(s.records) != 0 {
			t.Fatal("fixture did not exhaust the independent verbose bound")
		}
		for range 2 {
			d := s.lifecycle.document()
			refs := d["omitted_refs"].([]string)
			if d["status"] != "partial" || d["omitted_records"] != 1 || len(refs) != 1 || refs[0] != v.AllocationRef || len(s.lifecycle.records) != 0 {
				t.Fatal("missing terminal was concealed, duplicated or fabricated", d)
			}
		}
	})
}
