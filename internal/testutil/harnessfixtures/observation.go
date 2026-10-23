package harnessfixtures

// Retirement is supplied at the ownership decision, never reconstructed by a
// diagnostic reader from an allocation name or a timing coincidence.
type Retirement struct{ Reason, UnitID string }
type AllocationInterval struct {
	Start    float64 `json:"start_ms"`
	End      float64 `json:"end_ms"`
	Duration float64 `json:"duration_ms"`
}
type AllocationBorrow struct {
	LeaseRef string `json:"lease_ref"`
	UnitID   string `json:"unit_id"`
}
type AllocationLifecycle struct {
	AllocationRef string             `json:"allocation_ref"`
	Borrows       []AllocationBorrow `json:"borrows"`
	ClockIdentity string             `json:"clock_identity"`
	Acquisition   AllocationInterval `json:"acquisition"`
	Retirement    AllocationInterval `json:"retirement"`
	Reason        string             `json:"retirement_reason"`
	Trigger       *string            `json:"triggering_unit_id"`
	Settlement    string             `json:"settlement"`
	Coverage      string             `json:"coverage"`
}

func (b *Broker) elapsed() float64 {
	return float64(b.options.Clock().Sub(b.epoch).Nanoseconds()) / 1e6
}
func (b *Broker) newLifecycle(request Request, start float64) AllocationLifecycle {
	end := b.elapsed()
	return AllocationLifecycle{AllocationRef: request.AllocationRef, Borrows: []AllocationBorrow{}, ClockIdentity: "fixture-broker-monotonic", Acquisition: AllocationInterval{start, end, end - start}, Coverage: "complete"}
}
func (b *Broker) observeLifecycle(request Request, value *AllocationLifecycle, start float64, retirement Retirement, settlement string) {
	if request.Capability != "browser_stack" {
		return
	}
	end := b.elapsed()
	value.Retirement = AllocationInterval{start, end, end - start}
	value.Reason = retirement.Reason
	if retirement.UnitID != "" {
		unit := retirement.UnitID
		value.Trigger = &unit
	}
	value.Settlement = settlement
	b.observe(Observation{Kind: "allocation_lifecycle", AllocationRef: request.AllocationRef, Lifecycle: value})
}

func (b *Broker) retirementDecision(value Retirement) Retirement {
	if b.options.Cancelled != nil && b.options.Cancelled() {
		value.Reason = "invocation_cancellation"
	}
	return value
}
