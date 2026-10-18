//go:build linux

package harnessruntime

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
)

// nativeFixtures adapts graph facts to the single fixture policy owner. It does
// not repeat lease, reuse, acquisition, or recovery decisions.
type nativeFixtures struct{ production *harnessfixtures.Production }

func (f *nativeFixtures) Acquire(ctx context.Context, r fixtureRequest) (*fixtureLease, error) {
	env := r.Environment
	lease, err := f.production.Acquire(ctx, harnessfixtures.Request{UnitID: r.UnitID, Capability: r.Capability, Affinity: r.Affinity,
		BrowserStage: env["CARTULARY_BROWSER_STAGE"], RuntimeProfile: env["CARTULARY_BROWSER_RUNTIME_PROFILE_ID"],
		FixtureProfile: env["CARTULARY_FIXTURE_PROFILE_ID"], SnapshotKey: env["CARTULARY_FIXTURE_SNAPSHOT_KEY"],
		BuilderUnit: env["CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"], RowID: env["CARTULARY_FIXTURE_ROW_ID"], PredicateID: env["CARTULARY_FIXTURE_PREDICATE_ID"]})
	if err != nil {
		return nil, err
	}
	return &fixtureLease{ID: lease.ID, Environment: lease.Environment, Release: lease.Release}, nil
}
func (f *nativeFixtures) Close(ctx context.Context) error { return f.production.CloseFixtures(ctx) }
func (f *nativeFixtures) Unresolved() bool {
	if f.production.Broker.Unresolved() {
		return true
	}
	resources, err := f.production.Options.Runtime.Resources(4096)
	if err != nil {
		return true
	}
	for _, r := range resources {
		if r.Kind == "browser_stack" || r.Kind == "object_store_proxy" || r.Kind == "browser_renderer" {
			return true
		}
	}
	return false
}
