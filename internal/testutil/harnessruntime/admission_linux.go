//go:build linux

package harnessruntime

import (
	"context"
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
			lease, err = harnessadmission.Acquire(ctx, root, claims, available, parent, slices.Contains(u.ExclusiveLocks, "host_activity"))
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
		return &hostLease{Token: lease.Token, Release: func(ctx context.Context) error {
			var err error
			if finish != nil {
				err = finish(ctx)
			}
			return errors.Join(err, lease.Release(ctx))
		}, Bind: lease.Bind}, nil
	}
}

func workRequirement(u unit) harnessfixtures.WorkRequirement {
	request := nativeFixtureRequest(fixtureRequest{PostgresIsolation: u.PostgresIsolation, Services: u.Services, UnitID: u.ID, Capability: u.Fixture, Affinity: fixtureAffinity(u), Environment: u.Command.Environment})
	return harnessfixtures.WorkRequirement{Request: request, Renderer: u.Renderer, Physical: u.Physical, Active: u.Claims, Exclusive: slices.Contains(u.ExclusiveLocks, "host_activity")}
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
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, harnessadmission.ErrConflict) {
		return failed("infra", "resource_conflict")
	}
	return failureOf(err)
}
