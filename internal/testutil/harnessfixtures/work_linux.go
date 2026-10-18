//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

type WorkRequirement struct {
	Request     Request
	Renderer    string
	Physical    map[string]map[string]int64
	Active      map[string]int64
	Exclusive   bool
	Diagnostics bool
}
type admittedWork struct {
	requirement             WorkRequirement
	slot                    *rendererSlot
	host                    *harnessadmission.Lease
	preservedStack, settled bool
	mu                      sync.Mutex
}
type WorkLease struct {
	ID          string
	Environment map[string]string
	Release     func(context.Context, bool, bool) (bool, error)
}

// ConfigureAdmission binds this invocation to the existing host admission
// owner. Profile quantities come from the compiled graph, not fixture guesses.
func (p *Production) ConfigureAdmission(root string, parent *string, capacities map[string]int64, requirements []WorkRequirement) {
	r := p.Options.residency
	if root == "" {
		root = harnessadmission.DefaultRoot()
	}
	r.root = root
	r.parent = parent
	r.capacities = quantities(capacities)
	p.serviceClaims = map[string]int64{}
	services := map[string]bool{}
	for _, work := range requirements {
		for _, service := range []string{"postgres", "object_store"} {
			if claims, ok := work.Physical[service]; ok && !services[service] {
				for key, value := range claims {
					p.serviceClaims[key] += value
				}
				services[service] = true
			}
		}
		if work.Renderer != "none" && work.Renderer != "" {
			p.renderers.remaining[work.Request.RuntimeProfile]++
		}
	}
}

func (p *Production) Admit(ctx context.Context, requirement WorkRequirement) (*harnessadmission.Lease, func(context.Context) error, error) {
	r := p.Options.residency
	work := &admittedWork{requirement: requirement}
	if requirement.Exclusive {
		if err := p.Options.withResidentActivity(ctx, func(ctx context.Context) error {
			return errors.Join(p.renderers.evictIdle(ctx, nil), p.Broker.RetireIdle(ctx))
		}); err != nil {
			return nil, nil, err
		}
	}
	releaseStack, err := p.Broker.reserveStack(ctx, requirement.Request)
	if err != nil {
		return nil, nil, err
	}
	if requirement.Renderer != "none" && requirement.Renderer != "" {
		work.slot, err = p.renderers.reserve(ctx, requirement.Request.UnitID, requirement.Request.RuntimeProfile, requirement.Renderer, requirement.Diagnostics)
		if err != nil {
			releaseStack()
			return nil, nil, err
		}
	}
	records := []*residentRecord{}
	add := func(key string, claims map[string]int64) error {
		record, err := r.ensure(key, claims)
		if err != nil {
			return err
		}
		records = append(records, record)
		return nil
	}
	// Attached services retain the persistent owner's reservation. Attachment
	// validates that capability; an invocation never recharges borrowed services.
	if p.Options.Environment["CARTULARY_TEST_SERVICES_MODE"] != "attach" && (len(requirement.Physical["postgres"]) != 0 || len(requirement.Physical["object_store"]) != 0) {
		err = add("services", p.serviceClaims)
	}
	if err == nil && requirement.Request.Capability == "browser_stack" {
		err = add(stackKey(requirement.Request), requirement.Physical["browser_stack"])
	}
	if err == nil && work.slot != nil {
		err = add("renderer:"+work.slot.allocation, requirement.Physical["renderer"])
		if err == nil {
			work.slot.resident = records[len(records)-1]
		}
	}
	residents := []harnessadmission.Resident{}
	for _, record := range records {
		residents = append(residents, record.Allocation)
	}
	waiting := func(ctx context.Context, pending *harnessadmission.Lease) error {
		retired, err := p.renderers.evictForAdmission(ctx, work.slot, pending)
		if err != nil || retired > 0 {
			return err
		}
		// Retained stacks occupy ports even when every renderer has retired.
		// Release only an idle stack whose own claims relieve persistent pressure.
		_, err = p.Broker.retireIdle(ctx, 1, func(request Request) (bool, error) {
			r.mu.Lock()
			record := r.allocations[stackKey(request)]
			r.mu.Unlock()
			if record == nil {
				return false, nil
			}
			return pending.NeedsRetirement(ctx, record.Allocation.Claims)
		}, func(ctx context.Context, e *entry) error {
			return p.Options.withResidentActivity(ctx, func(ctx context.Context) error {
				return p.Broker.clean(ctx, e)
			})
		})
		return err
	}
	for err == nil {
		if requirement.Renderer != "pooled" {
			work.host, err = harnessadmission.AcquireWithResidents(ctx, r.root, quantities(requirement.Active), r.capacities, r.parent, requirement.Exclusive, residents, waiting)
			break
		}
		work.host, err = harnessadmission.TryAcquireWithResidents(ctx, r.root, quantities(requirement.Active), r.capacities, r.parent, residents, waiting)
		if err != nil || work.host != nil {
			break
		}
		// The host request is withdrawn before the logical slot is returned.
		// Another consumer may fit by reusing an already resident stack.
		p.renderers.unreserve(work.slot)
		work.slot = nil
		timer := time.NewTimer(50 * time.Millisecond)
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case <-timer.C:
		}
		timer.Stop()
		if err != nil {
			break
		}
		work.slot, err = p.renderers.reserve(ctx, requirement.Request.UnitID, requirement.Request.RuntimeProfile, requirement.Renderer, requirement.Diagnostics)
		if err != nil {
			break
		}
		records = records[:len(records)-1]
		err = add("renderer:"+work.slot.allocation, requirement.Physical["renderer"])
		if err != nil {
			break
		}
		work.slot.resident = records[len(records)-1]
		residents = residents[:len(residents)-1]
		residents = append(residents, work.slot.resident.Allocation)
	}
	finish := func(ctx context.Context) error {
		defer releaseStack()
		work.mu.Lock()
		defer work.mu.Unlock()
		var result error
		if work.host != nil {
			ctx = WithActivity(ctx, work.host.Token)
		}
		if !work.settled && work.slot != nil {
			result = p.renderers.retire(ctx, work.slot)
		}
		if !work.preservedStack && requirement.Request.Capability == "browser_stack" {
			result = errors.Join(result, r.close(ctx, stackKey(requirement.Request)))
		}
		p.workMu.Lock()
		delete(p.work, requirement.Request.UnitID)
		p.workMu.Unlock()
		return result
	}
	if err != nil {
		releaseStack()
		// Queue cancellation cannot leave a reserved slot occupied. Unacquired
		// capacity files remain journal-owned and close with the invocation.
		if work.slot != nil {
			p.renderers.mu.Lock()
			work.slot.borrower = ""
			if work.slot.physical != nil {
				work.slot.state = "ready"
			} else {
				p.renderers.slots = slices.DeleteFunc(p.renderers.slots, func(slot *rendererSlot) bool { return slot == work.slot })
			}
			p.renderers.signal()
			p.renderers.mu.Unlock()
		}
		return nil, nil, err
	}
	for _, record := range records {
		if err := p.Options.Runtime.RecordResource("resident_allocation", record.Allocation.Reference, "acquired"); err != nil {
			return work.host, finish, err
		}
	}
	p.workMu.Lock()
	p.work[requirement.Request.UnitID] = work
	p.workMu.Unlock()
	return work.host, finish, nil
}

func (p *Production) Acquire(ctx context.Context, request Request) (*WorkLease, error) {
	p.workMu.Lock()
	work := p.work[request.UnitID]
	p.workMu.Unlock()
	if work != nil {
		ctx = WithActivity(ctx, work.host.Token)
	}
	var lease *Lease
	var err error
	environment := maps.Clone(p.Options.Environment)
	id := "renderer:" + request.UnitID
	if request.Capability != "none" {
		lease, err = p.Broker.Acquire(ctx, request)
		if err != nil {
			return nil, err
		}
		environment = lease.Environment()
		id = lease.Record().ID
	}
	var rendererRelease func(context.Context, bool) (bool, error)
	if work != nil && work.slot != nil {
		attached, release, err := p.renderers.borrow(ctx, work.slot, environment)
		if err != nil {
			var closeErr error
			if lease != nil {
				_, closeErr = lease.Release(ctx, false, false)
			}
			return nil, errors.Join(err, closeErr)
		}
		maps.Copy(environment, attached)
		dependencies := []string{}
		if value := environment["CARTULARY_HARNESS_RESOURCE_DEPENDENCIES"]; value != "" {
			if err := json.Unmarshal([]byte(value), &dependencies); err != nil {
				return nil, err
			}
		}
		dependencies = append(dependencies, attached["CARTULARY_RENDERER_RESOURCE_REF"])
		environment["CARTULARY_HARNESS_RESOURCE_DEPENDENCIES"] = resourceDependencies(dependencies)
		rendererRelease = release
	}
	var once sync.Once
	var retained bool
	var closeErr error
	release := func(ctx context.Context, healthy, retain bool) (bool, error) {
		once.Do(func() {
			if work != nil {
				ctx = WithActivity(ctx, work.host.Token)
			}
			rendererRetained := false
			if rendererRelease != nil {
				rendererRetained, closeErr = rendererRelease(ctx, healthy)
			}
			var err error
			if lease != nil {
				retained, err = lease.Release(ctx, healthy && closeErr == nil, retain)
			}
			closeErr = errors.Join(closeErr, err)
			if work != nil {
				work.mu.Lock()
				work.preservedStack = retained
				work.settled = true
				work.mu.Unlock()
			}
			retained = retained || rendererRetained
		})
		return retained, closeErr
	}
	return &WorkLease{ID: id, Environment: environment, Release: release}, nil
}

// ConfigureDedicated projects the same authored physical claims for a native
// review owner. Its one borrower lives until review settlement, not group exit.
func (p *Production) ConfigureDedicated(parent *string, capacities map[string]int64, request Request, diagnostics bool) (WorkRequirement, error) {
	bytes, err := os.ReadFile(filepath.Join(p.Options.Root, "tools/execution_topology_manifest.json"))
	if err != nil {
		return WorkRequirement{}, err
	}
	topology, err := harnessschema.Decode(bytes)
	if err != nil || harnessschema.Validate("cartulary.execution_topology.v14", topology) != nil {
		return WorkRequirement{}, errors.New("invalid physical resource topology")
	}
	data, err := json.Marshal(topology.(map[string]any)["physical_resource_claims"])
	if err != nil {
		return WorkRequirement{}, err
	}
	physical := map[string]map[string]int64{}
	if err := json.Unmarshal(data, &physical); err != nil {
		return WorkRequirement{}, err
	}
	work := WorkRequirement{Request: request, Renderer: "dedicated", Diagnostics: diagnostics, Physical: physical, Active: map[string]int64{"cpu": 1, "io": 1, "memory_mb": 512, "process": 1}}
	p.ConfigureAdmission("", parent, capacities, []WorkRequirement{work})
	return work, nil
}
