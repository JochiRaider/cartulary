//go:build linux

package harnessfixtures

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
)

type retirementCandidate struct {
	key    string
	unused bool
	slot   *rendererSlot
}

// retireAllocations is called with retireMu held. It orders stacks and renderer
// slots together, then each physical owner pins/revalidates its exact generation.
// Pressure retires at most one allocation before native admission reevaluates.
func (p *Production) retireAllocations(ctx context.Context, pending *harnessadmission.Lease, unusedOnly bool, reason Retirement) (int, error) {
	candidates := []retirementCandidate{}
	p.renderers.mu.Lock()
	for _, s := range p.renderers.slots {
		if s.borrower == "" && (s.state == "ready" || s.state == "unallocated") && (pending == nil || s.physical != nil) {
			candidates = append(candidates, retirementCandidate{key: fmt.Sprintf("renderer:%04d", s.id), unused: p.renderers.future(s, "") == 0, slot: s})
		}
	}
	p.renderers.mu.Unlock()
	p.Broker.mu.Lock()
	for _, e := range p.Broker.allocations {
		if e.request.Capability == "browser_stack" && e.refs == 0 && p.Broker.reserved[e.key] == 0 && !e.settled && p.Broker.shared[e.key] == e {
			candidates = append(candidates, retirementCandidate{key: e.key, unused: p.stackConsumers(e.key, "") == 0})
		}
	}
	p.Broker.mu.Unlock()
	slices.SortFunc(candidates, func(a, b retirementCandidate) int {
		if a.unused != b.unused {
			if a.unused {
				return -1
			}
			return 1
		}
		return strings.Compare(a.key, b.key)
	})
	retired := 0
	for _, c := range candidates {
		if unusedOnly && !c.unused {
			continue
		}
		disposition := reason
		if c.unused && disposition.Reason == "capacity_pressure" {
			disposition.Reason = "terminal_borrower_release"
		}
		if disposition.Reason == "terminal_borrower_release" && c.slot == nil {
			disposition.UnitID = p.lastTerminal(func(v fixtureConsumer) bool { return v.stack == c.key })
		}
		count := 0
		var err error
		if c.slot != nil {
			var did bool
			did, err = p.renderers.retireSlotIfIdle(ctx, c.slot, unusedOnly, pending, disposition.Reason)
			if did {
				count = 1
			}
		} else {
			count, err = p.retireIdleStacks(ctx, disposition, 1, func(request Request) (bool, error) {
				if sharedKey(request) != c.key {
					return false, nil
				}
				unused := p.stackConsumers(c.key, "") == 0
				if unusedOnly && !unused {
					return false, nil
				}
				if pending == nil || unused {
					return true, nil
				}
				records := p.stackRetirementRecords(request)
				if len(records) == 0 {
					return false, nil
				}
				released := []harnessadmission.Resident{}
				for _, record := range records {
					released = append(released, record.Allocation)
				}
				return pending.NeedsRetirement(ctx, released)
			})
		}
		if err != nil {
			return retired, err
		}
		retired += count
		if pending != nil && count > 0 {
			return retired, nil
		}
	}
	return retired, nil
}

func (p *Production) stackRetirementRecords(request Request) []*residentRecord {
	r := p.Options.residency
	r.mu.Lock()
	defer r.mu.Unlock()
	record := r.allocations[stackKey(request)]
	if record == nil {
		return nil
	}
	records := []*residentRecord{record}
	if request.PostgresIsolation == "exclusive" && request.FixtureProfile == "" {
		if cluster := r.allocations[clusterKey(request)]; cluster != nil {
			records = append(records, cluster)
		}
	}
	return records
}
