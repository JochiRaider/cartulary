//go:build linux

package harnessruntime

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
)

func nativeAdmission(capacities map[string]int64, parent *string, root string) func(context.Context, unit) (*hostLease, error) {
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
		lease, err := harnessadmission.Acquire(ctx, root, claims, available, parent, slices.Contains(u.ExclusiveLocks, "host_activity"))
		if err != nil {
			if errors.Is(err, harnessadmission.ErrRelease) {
				return nil, failed("harness", "cleanup_error")
			}
			if caller.Err() != nil {
				return nil, caller.Err()
			}
			return nil, failed("infra", "resource_conflict")
		}
		return &hostLease{Token: lease.Token, Release: lease.Release, Bind: lease.Bind}, nil
	}
}
