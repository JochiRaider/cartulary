//go:build linux

package harnessadmission

import (
	"context"
	"path/filepath"
)

func validResident(r Resident) bool {
	if !tokenPattern.MatchString(r.Token) || !filepath.IsAbs(r.Reference) || filepath.Clean(r.Reference) != r.Reference || !quantities(r.Claims) || !quantities(r.Capacities) {
		return false
	}
	if r.Parent != nil && !tokenPattern.MatchString(*r.Parent) {
		return false
	}
	for _, key := range resources {
		if r.Claims[key] > r.Capacities[key] {
			return false
		}
	}
	// Resident capacity never substitutes for active CPU or I/O admission.
	return r.Claims["cpu"] == 0 && r.Claims["io"] == 0
}

// ReleaseResident is called only by the physical owner's dependency-ordered
// cleanup, after exact physical removal. Process liveness never releases it.
func ReleaseResident(ctx context.Context, root string, allocation Resident) error {
	if !validResident(allocation) {
		return ErrConflict
	}
	_, err := Transact(ctx, Request{Root: root, Operation: "release_resident", Token: allocation.Token, Reference: allocation.Reference})
	return err
}
