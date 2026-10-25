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
	lease, err := f.production.Acquire(ctx, nativeFixtureRequest(r))
	if err != nil {
		return nil, err
	}
	return &fixtureLease{ID: lease.ID, Environment: lease.Environment, Release: lease.Release}, nil
}

// Admission and acquisition must use the same physical fixture identity.
func nativeFixtureRequest(r fixtureRequest) harnessfixtures.Request {
	env := r.Environment
	profile := env["CARTULARY_BROWSER_RUNTIME_PROFILE_ID"]
	if profile == "" {
		profile = "default"
	}
	request := harnessfixtures.Request{PostgresIsolation: r.PostgresIsolation, Services: r.Services, UnitID: r.UnitID, Capability: r.Capability, Affinity: r.Affinity,
		BrowserStage: env["CARTULARY_BROWSER_STAGE"], RuntimeProfile: profile,
		FixtureProfile: env["CARTULARY_FIXTURE_PROFILE_ID"], SnapshotKey: env["CARTULARY_FIXTURE_SNAPSHOT_KEY"],
		BuilderUnit: env["CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"], RowID: env["CARTULARY_FIXTURE_ROW_ID"], PredicateID: env["CARTULARY_FIXTURE_PREDICATE_ID"]}
	if label := env["CARTULARY_BROWSER_PREPARATION_LABEL"]; label != "" {
		request.Preparation = &harnessfixtures.BrowserPreparation{Label: label, Target: env["CARTULARY_TEST_TARGET"],
			DrainTimeout: env["CARTULARY_BROWSER_RESET_DRAIN_TIMEOUT_MS"], Timeout: env["CARTULARY_BROWSER_PREPARATION_TIMEOUT_MS"]}
	}
	return request
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
