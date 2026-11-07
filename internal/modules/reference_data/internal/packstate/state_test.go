package packstate

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

func fixtureVersion(key string) Version {
	return Version{Member: packformat.SetMember{Key: key, Version: "1", PayloadSHA256: "digest"}, Manifest: packformat.Manifest{Key: key}, Health: Available, Envelope: &Envelope{ID: "envelope-1", VerifiedAt: time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)}}
}
func TestDependencyFallbackPrunesToFixedPointAndRejectsCycles_Unit(t *testing.T) {
	t.Run("condition preserves unhealthy state while disabled", func(t *testing.T) {
		for _, health := range []Health{Available, Failed, Missing} {
			for _, disabled := range []bool{false, true} {
				want := string(health)
				if health == Available && disabled {
					want = "disabled"
				}
				if got := Condition(Version{Health: health, Disabled: disabled}); got != want {
					t.Fatalf("health=%s disabled=%t: %s", health, disabled, got)
				}
			}
		}
	})
	base := []Version{}
	for _, key := range []string{"type_registry.evidence", "type_registry.host", "type_registry.indicator"} {
		v := fixtureVersion(key)
		v.Builtin = true
		base = append(base, v)
	}
	t.Run("effective set owns immutable input copies", func(t *testing.T) {
		selected := fixtureVersion("enrichment.tor")
		selected.Manifest.Extensions = map[string]any{"example.org.fixture": []any{"original"}}
		got, _, err := ResolveEffectiveSet([]Version{selected}, base)
		if err != nil {
			t.Fatal(err)
		}
		index := slices.IndexFunc(got, func(v Version) bool { return v.Member.Key == selected.Member.Key })
		if index < 0 {
			t.Fatal("healthy optional member omitted")
		}
		selected.Envelope.ID = "mutated"
		selected.Manifest.Extensions["example.org.fixture"].([]any)[0] = "mutated"
		if got[index].Envelope.ID != "envelope-1" || got[index].Manifest.Extensions["example.org.fixture"].([]any)[0] != "original" {
			t.Fatal("effective set aliases its input")
		}
	})
	custom := fixtureVersion("type_registry.host")
	custom.Member.Version = "custom"
	custom.Health = Failed
	a := fixtureVersion("framework.attack")
	a.Manifest.Dependencies = []packformat.Dependency{{Key: custom.Member.Key, Version: "custom", SHA256: "digest"}}
	b := fixtureVersion("framework.d3fend")
	b.Manifest.Dependencies = []packformat.Dependency{{Key: a.Member.Key, Version: "1", SHA256: "digest"}}
	for _, selected := range [][]Version{{b, a, custom}, {custom, a, b}} {
		result, removed, err := ResolveEffectiveSet(selected, base)
		if err != nil || len(result) != 3 || !slices.Equal(removed, []string{"framework.attack", "framework.d3fend", "type_registry.host"}) {
			t.Fatalf("fallback failed: %#v %v %v", result, removed, err)
		}
		for _, v := range result {
			if !v.Builtin {
				t.Fatal("failed descriptor survived fallback")
			}
		}
	}
	t.Run("unavailable Base cannot erase a healthy replacement", func(t *testing.T) {
		lost := slices.Clone(base)
		lost[1].Health = Missing
		healthyCustom := custom
		healthyCustom.Health = Available
		got, _, err := ResolveEffectiveSet([]Version{healthyCustom}, lost)
		if err != nil || len(got) != 3 || got[1].Member.Version != "custom" {
			t.Fatal("healthy replacement lost", got, err)
		}
		if _, _, err := ResolveEffectiveSet([]Version{custom}, lost); !errors.Is(err, ErrRequiredRegistryUnavailable) {
			t.Fatal("required loss not distinguished", err)
		}
	})
	a.Manifest.Dependencies = []packformat.Dependency{{Key: b.Member.Key, Version: "1", SHA256: "digest"}}
	_, _, err := ResolveEffectiveSet([]Version{a, b}, base)
	var rejection *Rejection
	if !errors.As(err, &rejection) || rejection.Reason != "dependency_cycle" {
		t.Fatalf("cycle admitted: %v", err)
	}
	a.Manifest.Dependencies = nil
	a.Manifest.Conflicts = []packformat.Conflict{{Key: "type_registry.host"}}
	if _, _, err := ResolveEffectiveSet([]Version{a}, base); err == nil {
		t.Fatal("optional conflict against mandatory registry admitted")
	}
}
