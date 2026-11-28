//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
)

func nativeAdmission(capacities map[string]int64, parent *string, root string, production *harnessfixtures.Production) func(context.Context, unit) (*hostLease, error) {
	return func(caller context.Context, u unit) (*hostLease, error) {
		// Capacity waiting is invocation-owned queue residence. The process
		// runner applies the unit watchdog only after admission and setup.
		ctx := caller
		claims, available := harnessadmission.Quantities{}, harnessadmission.Quantities{}
		for k, v := range u.Claims {
			claims[k] = float64(v)
		}
		for k, v := range capacities {
			available[k] = float64(v)
		}
		var lease *harnessadmission.Lease
		var finish func(context.Context) error
		var err error
		if production == nil {
			lease, err = harnessadmission.Acquire(ctx, root, claims, available, parent, u.quiet())
		} else {
			lease, finish, err = production.Admit(ctx, workRequirement(u))
		}
		if err != nil {
			primary := classifyAdmissionFailure(caller, err)
			if lease != nil {
				cleanup, done := context.WithTimeout(context.Background(), 2*time.Minute)
				if finish != nil {
					err = errors.Join(err, finish(cleanup))
				}
				err = errors.Join(err, lease.Release(cleanup))
				done()
			}
			return nil, errors.Join(primary, err)
		}
		return &hostLease{Token: lease.Token, Bind: lease.Bind, Settle: finish, Release: lease.Release}, nil
	}
}

func workRequirement(u unit) harnessfixtures.WorkRequirement {
	request := nativeFixtureRequest(fixtureRequest{PostgresIsolation: u.PostgresIsolation, Services: u.Services, UnitID: u.ID, Capability: u.Fixture, Affinity: fixtureAffinity(u), Environment: u.Command.Environment})
	return harnessfixtures.WorkRequirement{Request: request, Renderer: u.Renderer, Residents: u.Residents, Active: u.Claims, Exclusive: u.quiet()}
}

func classifyAdmissionFailure(caller context.Context, err error) error {
	var typed interface{ HarnessFailure() (string, string) }
	if errors.As(err, &typed) {
		return failureOf(err)
	}
	if caller.Err() != nil {
		return failureOf(caller.Err())
	}
	if errors.Is(err, harnessadmission.ErrRelease) {
		return failed("harness", "cleanup_error")
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, harnessadmission.ErrConflict) || errors.Is(err, harnessadmission.ErrUnsatisfiable) {
		return failed("infra", "resource_conflict")
	}
	return failureOf(err)
}

// nativeAdmissionPort makes the host ledger the graph's admission authority
// for physical resources (G30): the scheduler evaluates its ranked ready work
// against one ledger view per decision and transacts only for admissible
// work, so no unit holds a token or a queue position while it waits.
type nativeAdmissionPort struct {
	ctx        context.Context
	root       string
	production *harnessfixtures.Production
	observe    func(context.Context, string) context.Context
	physical   map[string]bool
	changes    <-chan struct{}
}

func newNativeAdmissionPort(ctx context.Context, root string, production *harnessfixtures.Production, observe func(context.Context, string) context.Context) *nativeAdmissionPort {
	physical := map[string]bool{}
	for _, resource := range harnessadmission.Resources() {
		physical[resource] = true
	}
	return &nativeAdmissionPort{ctx: ctx, root: root, production: production, observe: observe, physical: physical}
}

func (p *nativeAdmissionPort) View(ctx context.Context) (admissionView, error) {
	view, err := harnessadmission.ReadView(ctx, p.root)
	if err != nil {
		return nil, err
	}
	// The ledger directory exists once it has been read.
	if p.changes == nil {
		if changes, err := harnessadmission.Watch(p.ctx, p.root); err == nil {
			p.changes = changes
		}
	}
	return &nativeAdmissionView{view: view, production: p.production, memo: map[string]verdict{}}, nil
}

func (p *nativeAdmissionPort) TryAdmit(ctx context.Context, u unit) (*hostLease, bool, error) {
	if p.observe != nil {
		ctx = p.observe(ctx, u.ID)
	}
	lease, finish, granted, err := p.production.TryAdmit(ctx, workRequirement(u))
	if err != nil || !granted {
		return nil, false, err
	}
	return &hostLease{Token: lease.Token, Bind: lease.Bind, Settle: finish, Release: lease.Release}, true, nil
}

func (p *nativeAdmissionPort) Changed() <-chan struct{} { return p.changes }
func (p *nativeAdmissionPort) Physical(resource string) bool {
	return p.physical[resource]
}

// nativeAdmissionView evaluates work against one ledger reading. Work with
// the same claims and resident identities has the same outcome, so each
// shape is evaluated once per view.
type nativeAdmissionView struct {
	view       *harnessadmission.View
	production *harnessfixtures.Production
	memo       map[string]verdict
}

func (v *nativeAdmissionView) Evaluate(u unit) (verdict, error) {
	requirement := workRequirement(u)
	shape, err := json.Marshal(map[string]any{"active": requirement.Active, "residents": requirement.Residents, "renderer": requirement.Renderer,
		"renderer_claims": requirement.RendererClaims, "exclusive": requirement.Exclusive})
	if err != nil {
		return verdict{}, err
	}
	// Unit-scoped residents are new for every unit; their keys differ but
	// their claims decide the outcome.
	key := string(shape)
	unitScoped := false
	for _, resident := range requirement.Residents {
		unitScoped = unitScoped || resident.Scope == "unit"
	}
	if unitScoped {
		normalized := map[string]any{"active": requirement.Active, "renderer": requirement.Renderer, "exclusive": requirement.Exclusive}
		residents := []any{}
		for _, resident := range requirement.Residents {
			if resident.Scope == "unit" {
				residents = append(residents, map[string]any{"scope": "unit", "claims": resident.Claims})
			} else {
				residents = append(residents, resident)
			}
		}
		normalized["residents"] = residents
		bytes, err := json.Marshal(normalized)
		if err != nil {
			return verdict{}, err
		}
		key = string(bytes)
	}
	if outcome, ok := v.memo[key]; ok {
		return outcome, nil
	}
	evaluation, err := v.production.Evaluate(v.view, requirement)
	if err != nil {
		return verdict{}, err
	}
	deficits := []string{}
	for resource, amount := range evaluation.Deficits {
		if amount > 0 {
			deficits = append(deficits, resource)
		}
	}
	slices.Sort(deficits)
	outcome := verdict{admissible: evaluation.Admissible, deficits: deficits}
	v.memo[key] = outcome
	return outcome, nil
}
