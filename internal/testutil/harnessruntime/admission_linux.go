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
		ctx, cancel := context.WithTimeout(caller, time.Duration(u.TimeoutMS)*time.Millisecond)
		defer cancel()
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
			if lease != nil {
				cleanup, done := context.WithTimeout(context.Background(), 2*time.Minute)
				if finish != nil {
					err = errors.Join(err, finish(cleanup))
				}
				err = errors.Join(err, lease.Release(cleanup))
				done()
			}
			if errors.Is(err, harnessadmission.ErrRelease) {
				return nil, failed("harness", "cleanup_error")
			}
			if caller.Err() != nil {
				return nil, caller.Err()
			}
			return nil, failed("infra", "resource_conflict")
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
	env := u.Command.Environment
	profile := env["CARTULARY_BROWSER_RUNTIME_PROFILE_ID"]
	if profile == "" {
		profile = "default"
	}
	return harnessfixtures.WorkRequirement{Request: harnessfixtures.Request{UnitID: u.ID, Capability: u.Fixture, Affinity: fixtureAffinity(u), RuntimeProfile: profile, FixtureProfile: env["CARTULARY_FIXTURE_PROFILE_ID"], SnapshotKey: env["CARTULARY_FIXTURE_SNAPSHOT_KEY"]}, Renderer: u.Renderer, Physical: u.Physical, Active: u.Claims, Exclusive: slices.Contains(u.ExclusiveLocks, "host_activity")}
}
