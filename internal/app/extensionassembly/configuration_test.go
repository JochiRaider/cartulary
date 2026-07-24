package extensionassembly

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
)

func TestResolveClaimRequestUsesGeneratedClaimKeys_Unit(t *testing.T) {
	descriptors := []extensions.Descriptor{
		{ProfileID: "future_profile", ClaimConfigKey: "future_profile.claimed"},
		{ProfileID: "existing_profile", ClaimConfigKey: "existing_profile.claimed"},
	}
	paths, err := ClaimConfigurationPaths(descriptors)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"existing_profile.claimed", "future_profile.claimed"}; !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths = %#v, want %#v", paths, want)
	}
	claims, err := ResolveClaimRequest(descriptors, map[string]bool{
		"existing_profile.claimed": false,
		"future_profile.claimed":   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"future_profile"}; !reflect.DeepEqual(claims, want) {
		t.Fatalf("claims = %#v, want %#v", claims, want)
	}
}

func TestResolveClaimRequestRejectsIncompleteOrExtraProjection_Unit(t *testing.T) {
	descriptors := []extensions.Descriptor{{ProfileID: "future_profile", ClaimConfigKey: "future_profile.claimed"}}
	for name, values := range map[string]map[string]bool{
		"missing": {},
		"extra":   {"future_profile.claimed": true, "other.claimed": false},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ResolveClaimRequest(descriptors, values); err == nil {
				t.Fatal("expected projection rejection")
			}
		})
	}
}
