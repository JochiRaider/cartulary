//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"sync"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
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
	for _, work := range requirements {
		if key := clusterKey(work.Request); key != "" {
			p.clusterRemaining[key]++
		}
		if work.Renderer != "none" && work.Renderer != "" {
			p.renderers.remaining[work.Request.RuntimeProfile]++
		}
	}
}

func (p *Production) Admit(ctx context.Context, requirement WorkRequirement) (*harnessadmission.Lease, func(context.Context) error, error) {
	if err := validateServiceRequest(requirement.Request); err != nil {
		return nil, nil, err
	}
	r := p.Options.residency
	work := &admittedWork{requirement: requirement}
	if requirement.Exclusive {
		rendererErr := p.renderers.evictIdle(ctx, nil)
		_, stackErr := p.retireIdleStacks(ctx, Retirement{"exclusive_admission", requirement.Request.UnitID}, 0, nil)
		if err := errors.Join(rendererErr, stackErr); err != nil {
			return nil, nil, err
		}
	}
	releaseStack := func() {}
	records := []*residentRecord{}
	residents := []harnessadmission.Resident{}
	reserve := func() error {
		var err error
		// A request waiting for a renderer must not pin an idle stack that
		// another admitted request needs to retire to make progress.
		if requirement.Renderer != "none" && requirement.Renderer != "" {
			work.slot, err = p.renderers.reserve(ctx, requirement.Request.UnitID, requirement.Request.RuntimeProfile, requirement.Renderer, requirement.Diagnostics)
			if err != nil {
				return err
			}
		}
		unpin, err := p.Broker.reserveStack(ctx, requirement.Request)
		if err != nil {
			return err
		}
		releaseStack = unpin
		// Every miss yields logical ownership. An intervening retirement can
		// replace any generation, so rebuild all resident identities together.
		records = records[:0]
		residents = residents[:0]
		add := func(key string, claims map[string]int64) error {
			record, err := r.ensure(key, claims)
			if err != nil {
				return err
			}
			records = append(records, record)
			residents = append(residents, record.Allocation)
			return nil
		}
		// Attached services retain the persistent owner's reservation.
		attached := p.Options.Environment["CARTULARY_TEST_SERVICES_MODE"] == "attach"
		if len(requirement.Physical["object_store"]) != 0 && !attached {
			err = add("object_store:shared", requirement.Physical["object_store"])
		}
		if err == nil && len(requirement.Physical["postgres"]) != 0 && (!attached || requirement.Request.PostgresIsolation == "exclusive") {
			err = add(clusterKey(requirement.Request), requirement.Physical["postgres"])
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
		return err
	}
	yieldReservations := func() {
		if work.slot != nil {
			p.renderers.unreserve(work.slot)
			work.slot = nil
		}
		releaseStack()
		releaseStack = func() {}
	}
	waiting := func(ctx context.Context, pending *harnessadmission.Lease) error {
		// Pending priority survives, but logical reservations must not prevent
		// pressure retirement or another waiter from reacquiring a slot.
		yieldReservations()
		retired, err := p.renderers.evictForAdmission(ctx, work.slot, pending)
		if err != nil || retired > 0 {
			return err
		}
		// Retained stacks occupy ports even when every renderer has retired.
		// Release only an idle stack whose own claims relieve persistent pressure.
		_, err = p.retireIdleStacks(ctx, Retirement{"capacity_pressure", requirement.Request.UnitID}, 1, func(request Request) (bool, error) {
			r.mu.Lock()
			record := r.allocations[stackKey(request)]
			r.mu.Unlock()
			if record == nil {
				return false, nil
			}
			claims := maps.Clone(record.Allocation.Claims)
			r.mu.Lock()
			cluster := r.allocations[clusterKey(request)]
			r.mu.Unlock()
			if cluster != nil {
				for key, value := range cluster.Allocation.Claims {
					claims[key] += value
				}
			}
			return pending.NeedsRetirement(ctx, claims)
		})
		return err
	}
	var err error
	work.host, err = harnessadmission.AcquireWithPreparedResidents(ctx, r.root, quantities(requirement.Active), r.capacities, r.parent, requirement.Exclusive, func(context.Context) ([]harnessadmission.Resident, error) {
		if err := reserve(); err != nil {
			return nil, err
		}
		return residents, nil
	}, waiting)
	clusterFinished := false
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
		if key := clusterKey(requirement.Request); key != "" && !clusterFinished {
			clusterFinished = true
			p.servicesMu.Lock()
			p.clusterRemaining[key]--
			last := p.clusterRemaining[key] == 0
			p.servicesMu.Unlock()
			if last {
				result = errors.Join(result, p.retireCluster(ctx, key))
			}
		}
		p.workMu.Lock()
		delete(p.work, requirement.Request.UnitID)
		p.workMu.Unlock()
		return result
	}
	if err != nil {
		// The admission owner settles its pending request before return. Keep
		// unstarted slots reusable; their proofs settle with the existing owner.
		yieldReservations()
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
	if err != nil || harnessschema.Validate("cartulary.execution_topology.v15", topology) != nil {
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

// retireIdleStacks reserves cleanup activity against the selected stack alone.
// The broker pins its retirement before this callback, so its resident identity
// cannot be replaced by another borrower. Other residents stay charged by the
// admission owner, but independent retirement cannot invalidate this request.
func (p *Production) retireIdleStacks(ctx context.Context, retirement Retirement, maximum int, needed func(Request) (bool, error)) (int, error) {
	return p.Broker.retireIdle(ctx, maximum, needed, func(ctx context.Context, e *entry) error {
		r := p.Options.residency
		r.mu.Lock()
		record := r.allocations[stackKey(e.request)]
		r.mu.Unlock()
		if record == nil {
			return harnesssession.AtCleanupStage("resource_inventory", harnesslocal.ErrUnsafe)
		}
		return r.drain(ctx, []*residentRecord{record}, func(ctx context.Context) error { return p.Broker.clean(ctx, e, retirement) })
	})
}
