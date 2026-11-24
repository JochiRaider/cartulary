package performancefixture

import (
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
)

func TestRowFixtureRouting(t *testing.T) {
	t.Parallel()
	var profile, other performancefixtureprofile.Profile
	for _, candidate := range performancefixtureprofile.Profiles() {
		if candidate.Status != "active" || len(candidate.VerificationBindings) == 0 {
			continue
		}
		if profile.FixtureProfileID == "" {
			profile = candidate
		} else if other.FixtureProfileID == "" {
			other = candidate
		}
	}
	if other.FixtureProfileID == "" {
		t.Fatal("routing needs two active profiles with bound verifications")
	}
	binding := profile.VerificationBindings[0]

	fixture, ok, err := ResolveRow(profile.FixtureProfileID, []string{"unbound.verification", binding.VerificationID})
	if err != nil || !ok || fixture.Profile.FixtureProfileID != profile.FixtureProfileID || !slices.Equal(fixture.Predicates, []string{binding.PredicateID}) {
		t.Fatalf("bound row = %+v, %v, %v; want profile %s predicate %s", fixture.Predicates, ok, err, profile.FixtureProfileID, binding.PredicateID)
	}
	if _, ok, err := ResolveRow("", []string{"unbound.verification"}); err != nil || ok {
		t.Fatalf("unbound row = %v, %v; want no fixture", ok, err)
	}
	for name, row := range map[string]struct {
		profile       string
		verifications []string
	}{
		"profile without a binding":  {profile.FixtureProfileID, []string{"unbound.verification"}},
		"binding without a profile":  {"", []string{binding.VerificationID}},
		"binding of another profile": {other.FixtureProfileID, []string{binding.VerificationID}},
		"bindings of two profiles":   {profile.FixtureProfileID, []string{binding.VerificationID, other.VerificationBindings[0].VerificationID}},
	} {
		if _, _, err := ResolveRow(row.profile, row.verifications); err == nil {
			t.Errorf("%s: routing succeeded", name)
		}
	}
}
