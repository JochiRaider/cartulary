package performancefixture

import (
	"fmt"
	"sync"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
)

// RowFixture is the populated fixture a catalog row's verifications bind: the
// row's active profile and the predicate each bound verification names, in
// verification order.
type RowFixture struct {
	Profile    performancefixtureprofile.Profile
	Predicates []string
}

type boundVerification struct{ profileID, predicateID string }

// verificationBindings indexes every registered verification binding by
// verification ID. A verification bound by two profiles is an invalid
// registry.
var verificationBindings = sync.OnceValues(func() (map[string]boundVerification, error) {
	bindings := map[string]boundVerification{}
	for _, profile := range performancefixtureprofile.Profiles() {
		for _, binding := range profile.VerificationBindings {
			if _, ok := bindings[binding.VerificationID]; ok {
				return nil, fmt.Errorf("verification %s is bound by more than one fixture profile", binding.VerificationID)
			}
			bindings[binding.VerificationID] = boundVerification{profile.FixtureProfileID, binding.PredicateID}
		}
	}
	return bindings, nil
})

// ResolveRow routes a catalog row to its populated fixture through the
// registry's verification bindings. A row declares a profile exactly when one
// of its verifications is bound, and every bound verification names that
// active profile; the profile is never inferred from anything else. ok is
// false for a row with no bound verification.
func ResolveRow(fixtureProfileID string, verificationIDs []string) (fixture RowFixture, ok bool, err error) {
	bindings, err := verificationBindings()
	if err != nil {
		return RowFixture{}, false, err
	}
	for _, id := range verificationIDs {
		bound, found := bindings[id]
		if !found {
			continue
		}
		if bound.profileID != fixtureProfileID {
			return RowFixture{}, false, fmt.Errorf("verification %s binds fixture profile %s, not %q", id, bound.profileID, fixtureProfileID)
		}
		fixture.Predicates = append(fixture.Predicates, bound.predicateID)
	}
	if len(fixture.Predicates) == 0 {
		if fixtureProfileID != "" {
			return RowFixture{}, false, fmt.Errorf("fixture profile %s is declared without a bound verification", fixtureProfileID)
		}
		return RowFixture{}, false, nil
	}
	profile, found := performancefixtureprofile.Lookup(fixtureProfileID)
	if !found || profile.Status != "active" {
		return RowFixture{}, false, fmt.Errorf("fixture profile %s is unknown or inactive", fixtureProfileID)
	}
	fixture.Profile = profile
	return fixture, true, nil
}
