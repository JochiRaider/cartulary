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
	"github.com/JochiRaider/cartulary/internal/testutil/runcontext"
)

type WorkRequirement struct {
	Request  Request
	Renderer string
	// Residents are the physical allocations the work holds, by the stable
	// keys its plan names. Attached residents belong to a persistent service
	// session and are not charged.
	Residents []Resident
	// RendererClaims are the claims of the renderer slot admission assigns;
	// nil uses the topology's renderer claims.
	RendererClaims map[string]int64
	Active         map[string]int64
	Exclusive      bool
	Diagnostics    bool
}

// dedicatedResidents are the residents of the review owner's single
// borrower, which no plan names: the shared object store, its PostgreSQL
// generation and its browser stack, under the topology's claims. An attached
// session holds the shared allocations.
func (p *Production) dedicatedResidents(request Request, physical map[string]map[string]int64) []Resident {
	attached := p.Options.Environment["CARTULARY_TEST_SERVICES_MODE"] == "attach"
	var residents []Resident
	shared := func(key string, claims map[string]int64) {
		scope := ""
		if attached {
			scope = ResidentAttached
		}
		residents = append(residents, Resident{Key: key, Scope: scope, Claims: claims})
	}
	if claims := physical["object_store"]; len(claims) != 0 {
		shared("object_store:shared", claims)
	}
	if claims := physical["postgres"]; len(claims) != 0 {
		if request.PostgresIsolation == "exclusive" {
			residents = append(residents, Resident{Key: clusterKey(request), Claims: claims})
		} else {
			shared(clusterKey(request), claims)
		}
	}
	if request.Capability == "browser_stack" {
		residents = append(residents, Resident{Key: stackKey(request), Claims: physical["browser_stack"]})
	}
	return residents
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
	Release     func(context.Context, bool) (bool, error)
}

// ConfigureAdmission binds this invocation to the existing host admission
// owner. Profile quantities come from the compiled graph, not fixture guesses.
func (p *Production) ConfigureAdmission(root string, parent *string, capacities map[string]int64) {
	r := p.Options.residency
	if root == "" {
		root = harnessadmission.DefaultRoot()
	}
	r.root = root
	r.parent = parent
	r.capacities = quantities(capacities)

}

// Admit waits for admission: dedicated owners and nested callers that hold
// no scheduler of their own.
func (p *Production) Admit(ctx context.Context, requirement WorkRequirement) (*harnessadmission.Lease, func(context.Context) error, error) {
	lease, finish, _, err := p.admit(ctx, requirement, false)
	return lease, finish, err
}

// TryAdmit makes one admission attempt for a graph scheduler that has ranked
// this work and found it admissible on a ledger view (G30). Denied work
// returns its logical reservations and keeps no queue position; granted is
// false and the error nil.
func (p *Production) TryAdmit(ctx context.Context, requirement WorkRequirement) (*harnessadmission.Lease, func(context.Context) error, bool, error) {
	return p.admit(ctx, requirement, true)
}

// Evaluate decides requirement against one ledger view without reserving a
// slot or creating a resident record: charged residents count once, the rest
// are prospective, and a renderer requirement adds one renderer slot.
func (p *Production) Evaluate(view *harnessadmission.View, requirement WorkRequirement) (harnessadmission.Evaluation, error) {
	r := p.Options.residency
	residents := []harnessadmission.Resident{}
	prospective := func(claims map[string]int64) error {
		resident, err := harnessadmission.ProspectiveResident(quantities(claims), r.capacities, r.parent)
		if err == nil {
			residents = append(residents, resident)
		}
		return err
	}
	for _, resident := range requirement.Residents {
		if resident.Scope == ResidentAttached {
			continue
		}
		r.mu.Lock()
		record := r.allocations[resident.Key]
		r.mu.Unlock()
		if record != nil {
			residents = append(residents, record.Allocation)
			continue
		}
		if err := prospective(resident.Claims); err != nil {
			return harnessadmission.Evaluation{}, err
		}
	}
	if requirement.Renderer != "none" && requirement.Renderer != "" {
		claims := requirement.RendererClaims
		if len(claims) == 0 {
			physical, err := p.physicalClaims()
			if err != nil {
				return harnessadmission.Evaluation{}, err
			}
			claims = physical["renderer"]
		}
		if err := prospective(claims); err != nil {
			return harnessadmission.Evaluation{}, err
		}
	}
	return view.Evaluate(quantities(requirement.Active), r.capacities, r.parent, requirement.Exclusive, residents)
}

// AdmissionRoot is the ledger this production owner admits against.
func (p *Production) AdmissionRoot() string { return p.Options.residency.root }

func (p *Production) admit(ctx context.Context, requirement WorkRequirement, try bool) (*harnessadmission.Lease, func(context.Context) error, bool, error) {
	if err := validateServiceRequest(requirement.Request); err != nil {
		return nil, nil, false, err
	}
	if err := p.validateConsumer(requirement); err != nil {
		return nil, nil, false, err
	}
	r := p.Options.residency
	work := &admittedWork{requirement: requirement}
	if requirement.Exclusive {
		endObservation := harnessadmission.Phase(ctx, "pressure_retirement")
		p.retireMu.Lock()
		_, _, err := p.retireAllocations(ctx, nil, false, Retirement{"exclusive_admission", requirement.Request.UnitID})
		p.retireMu.Unlock()
		endObservation()
		if err != nil {
			return nil, nil, false, err
		}
	}
	releaseStack := func() {}
	records := []*residentRecord{}
	residents := []harnessadmission.Resident{}
	held := requirement.Residents
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
		for _, resident := range held {
			if err == nil && resident.Scope != ResidentAttached {
				err = add(resident.Key, resident.Claims)
			}
		}
		if err == nil && work.slot != nil {
			claims := requirement.RendererClaims
			if len(claims) == 0 {
				var physical map[string]map[string]int64
				if physical, err = p.physicalClaims(); err == nil {
					claims = physical["renderer"]
				}
			}
			if err == nil {
				err = add("renderer:"+work.slot.allocation, claims)
			}
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
	waiting := func(ctx context.Context, pending *harnessadmission.Lease) (bool, error) {
		// Pending priority survives, but logical reservations must not prevent
		// pressure retirement or another waiter from reacquiring a slot.
		yieldReservations()
		p.retireMu.Lock()
		defer p.retireMu.Unlock()
		retired, deferred, err := p.retireAllocations(ctx, pending, false, Retirement{"capacity_pressure", requirement.Request.UnitID})
		return retired+deferred > 0, err
	}

	var err error
	prepare := func(context.Context) ([]harnessadmission.Resident, error) {
		if err := reserve(); err != nil {
			return nil, err
		}
		return residents, nil
	}
	granted := true
	if try {
		work.host, granted, err = harnessadmission.TryAcquireWithPreparedResidents(ctx, r.root, quantities(requirement.Active), r.capacities, r.parent, requirement.Exclusive, prepare)
	} else {
		work.host, err = harnessadmission.AcquireWithPreparedResidents(ctx, r.root, quantities(requirement.Active), r.capacities, r.parent, requirement.Exclusive, prepare, waiting)
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
	if err != nil || !granted {
		// The admission owner settles its pending request before return. Keep
		// unstarted slots reusable; their proofs settle with the existing owner.
		yieldReservations()
		return nil, nil, false, err
	}
	for _, record := range records {
		if err := p.Options.Runtime.RecordResource("resident_allocation", record.Allocation.Reference, "acquired"); err != nil {
			return work.host, finish, true, err
		}
	}
	p.workMu.Lock()
	p.work[requirement.Request.UnitID] = work
	p.workMu.Unlock()
	return work.host, finish, true, nil
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
		maps.Copy(environment, runcontext.ResourceDependencies(dependencies))
		rendererRelease = release
	}
	var once sync.Once
	var retained bool
	var closeErr error
	release := func(ctx context.Context, healthy bool) (bool, error) {
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
				retained, err = lease.Release(ctx, healthy && closeErr == nil, p.stackConsumers(sharedKey(request), request.UnitID) > 0)
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
	physical, err := p.physicalClaims()
	if err != nil {
		return WorkRequirement{}, err
	}
	work := WorkRequirement{Request: request, Renderer: "dedicated", Diagnostics: diagnostics, Residents: p.dedicatedResidents(request, physical), Active: map[string]int64{"cpu": 1, "io": 1, "memory_mb": 512, "process": 1}}
	p.ConfigureAdmission("", parent, capacities)
	return work, nil
}

// retireIdleStacks reserves cleanup activity against the selected stack alone.
// The broker pins its retirement before this callback, so its resident identity
// cannot be replaced by another borrower. Other residents stay charged by the
// admission owner, but independent retirement cannot invalidate this request.
func (p *Production) retireIdleStacks(ctx context.Context, retirement Retirement, maximum int, needed func(Request) (bool, error)) (int, error) {
	return p.Broker.retireIdle(ctx, maximum, needed, func(ctx context.Context, e *entry) error {
		// Terminal notifications may arrive between candidate selection and
		// the pinned eligibility decision. Attribute mandatory retirement at
		// that boundary rather than reporting unproved capacity relief.
		decision := retirement
		if decision.Reason == "capacity_pressure" && p.stackConsumers(e.key, "") == 0 {
			decision = Retirement{"terminal_borrower_release", p.lastTerminal(func(v fixtureConsumer) bool { return v.stack == e.key })}
		}
		r := p.Options.residency
		records := p.stackRetirementRecords(e.request)
		if len(records) == 0 {
			return harnesssession.AtCleanupStage("resource_inventory", harnesslocal.ErrUnsafe)
		}
		released := []harnessadmission.Resident{}
		for _, record := range records {
			released = append(released, record.Allocation)
		}
		observe := harnessadmission.Retirement(ctx, released, decision.Reason)
		err := r.drain(ctx, records, func(ctx context.Context) error { return p.Broker.clean(ctx, e, decision) })
		observe(err)
		return err
	})
}

// physicalClaims are the topology's resident claims per physical allocation
// kind, read once. A plan names its residents but leaves the renderer slot,
// which admission assigns, to these claims.
func (p *Production) physicalClaims() (map[string]map[string]int64, error) {
	p.physicalOnce.Do(func() {
		bytes, err := os.ReadFile(filepath.Join(p.Options.Root, "tools/execution_topology_manifest.json"))
		if err != nil {
			p.physicalErr = err
			return
		}
		topology, err := harnessschema.Decode(bytes)
		if err != nil || harnessschema.Validate("cartulary.execution_topology.v16", topology) != nil {
			p.physicalErr = errors.New("invalid physical resource topology")
			return
		}
		data, err := json.Marshal(topology.(map[string]any)["physical_resource_claims"])
		if err == nil {
			err = json.Unmarshal(data, &p.physical)
		}
		p.physicalErr = err
	})
	return p.physical, p.physicalErr
}
