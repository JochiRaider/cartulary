package packstate

import (
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

func fixtureVersion(key string) Version {
	return Version{Member: packformat.SetMember{Key: key, Version: "1", PayloadSHA256: "digest"}, Manifest: packformat.Manifest{Key: key}, Health: Available, Envelope: &Envelope{ID: "envelope-1", VerifiedAt: time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)}}
}
func TestVerificationTransitionsPreserveDisablementAndSuccessfulHistory_Unit(t *testing.T) {
	previous := fixtureVersion("enrichment.tor")
	previous.Disabled = true
	for _, verdict := range []Verdict{ContentFailure, MissingPayload, Aborted} {
		got, err := ApplyVerification(previous, Renewal, verdict, nil)
		if err != nil || !reflect.DeepEqual(got, previous) {
			t.Fatalf("failed renewal changed retained state: %#v %v", got, err)
		}
	}
	next := &Envelope{ID: "envelope-2", VerifiedAt: previous.Envelope.VerifiedAt.Add(time.Hour)}
	got, err := ApplyVerification(previous, Reverify, Success, next)
	if err != nil || !got.Disabled || got.Health != Available || got.Envelope.ID != "envelope-2" || previous.Envelope.ID != "envelope-1" {
		t.Fatal("successful reverify changed administrative intent or old envelope")
	}
	activated, err := Activate(got)
	if err != nil || activated.Disabled || !got.Disabled {
		t.Fatal("explicit activation did not clear disablement independently")
	}
	missing, err := ApplyVerification(previous, Reverify, MissingPayload, nil)
	if err != nil || missing.Health != Missing || missing.MissingReason == nil || missing.Envelope.ID != "envelope-1" || !missing.Disabled {
		t.Fatal("missing reverify lost successful history")
	}
	restored, err := ApplyVerification(missing, Reverify, Success, next)
	if err != nil || restored.MissingReason != nil || restored.Health != Available {
		t.Fatal("successful verification did not clear missing reason")
	}
	candidate := fixtureVersion("enrichment.tor")
	candidate.Envelope = nil
	candidate.Health = Staged
	for _, verdict := range []Verdict{Aborted, ContentFailure, MissingPayload} {
		failed, err := ApplyVerification(candidate, Import, verdict, nil)
		want := Failed
		if verdict == MissingPayload {
			want = Missing
		}
		if err != nil || failed.Health != want || failed.Envelope != nil {
			t.Fatal("initial rejection fabricated success or remained staged")
		}
		if verdict == MissingPayload && (failed.MissingReason == nil || *failed.MissingReason != "staging_loss") {
			t.Fatal("staging loss not distinguished from retained storage loss")
		}
	}
	if _, err := AdmitReverify(candidate, false); err == nil {
		t.Fatal("never-successful candidate admitted for reverify")
	}
	_, err = AdmitReverify(candidate, true)
	var rejection *Rejection
	if !errors.As(err, &rejection) || rejection.Reason != "verification_pending" {
		t.Fatal("pending work must precede successful-envelope eligibility")
	}
	if Condition(got) != "disabled" || Condition(missing) != "missing" || Condition(restored) != "disabled" {
		t.Fatal("condition projection lost health or administrative disablement")
	}
	next.ID = "mutated"
	if got.Envelope.ID != "envelope-2" {
		t.Fatal("envelope input was not frozen")
	}
}
func TestRefreshAdmissionFreezesExactSuccessfulCohort_Unit(t *testing.T) {
	a := fixtureVersion("enrichment.tor")
	a.Manifest.Extensions = map[string]any{"example.org.fixture": []any{"original"}}
	b := fixtureVersion("enrichment.cisa_kev")
	b.Health = Failed
	b.Disabled = true
	never := fixtureVersion("enrichment.lolbas")
	never.Envelope = nil
	never.Health = Failed
	removed := fixtureVersion("enrichment.lolesxi")
	removed.Removed = true
	builtin := fixtureVersion("type_registry.host")
	builtin.Builtin = true
	cohort, err := FreezeRefresh([]Version{a, never, b, removed, builtin}, nil)
	if err != nil || len(cohort) != 2 || cohort[0].Member.Key != "enrichment.cisa_kev" {
		t.Fatal("wrong refresh cohort")
	}
	a.Manifest.Extensions["example.org.fixture"].([]any)[0] = "mutated"
	a.Envelope.ID = "mutated"
	if cohort[1].Envelope.ID != "envelope-1" || cohort[1].Manifest.Extensions["example.org.fixture"].([]any)[0] != "original" {
		t.Fatal("cohort input remained mutable")
	}
	_, err = FreezeRefresh([]Version{a, b}, map[string]bool{b.Member.Key: true})
	var rejection *Rejection
	if !errors.As(err, &rejection) || rejection.Reason != "verification_pending" {
		t.Fatalf("pending cohort work: %v", err)
	}
	_, err = FreezeRefresh([]Version{never, removed}, map[string]bool{never.Member.Key: true})
	if !errors.As(err, &rejection) || rejection.Reason != "verification_pending" {
		t.Fatal("selected key with never-successful pending work did not block admission")
	}
	cohort, err = FreezeRefresh([]Version{never, removed, builtin}, nil)
	if err != nil || len(cohort) != 0 {
		t.Fatal("ineligible retained versions entered empty refresh")
	}
}
func TestDependencyFallbackPrunesToFixedPointAndRejectsCycles_Unit(t *testing.T) {
	base := []Version{}
	for _, key := range []string{"type_registry.evidence", "type_registry.host", "type_registry.indicator"} {
		v := fixtureVersion(key)
		v.Builtin = true
		base = append(base, v)
	}
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
func TestPublicationRevisionComparisonIgnoresUnrelatedState_Unit(t *testing.T) {
	captured := map[string]int64{"trust/repo": 2, "set/current": 4, "usage/host": 7}
	current := map[string]int64{"trust/repo": 2, "set/current": 4, "usage/host": 7, "unrelated": 99}
	if err := CheckCapturedRevisions(captured, current); err != nil {
		t.Fatal(err)
	}
	current["trust/repo"]++
	if err := CheckCapturedRevisions(captured, current); err == nil {
		t.Fatal("shared repository race admitted")
	}
	current["trust/repo"] = 2
	delete(current, "usage/host")
	if err := CheckCapturedRevisions(captured, current); err == nil {
		t.Fatal("missing usage guard admitted")
	}
}
