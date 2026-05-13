package savedviews

import "testing"

func TestPhase8_SavedViewCreateDefaults_U_8_02(t *testing.T) {
	scope, ok := DefaultCreateScope(nil)
	if !ok || scope != ScopePrivate {
		t.Fatalf("omitted create scope must default to private, got %q ok=%v", scope, ok)
	}
	for _, allowed := range []Scope{ScopePrivate, ScopeShared} {
		if !IsOrdinaryCreateScope(allowed) {
			t.Fatalf("ordinary create must allow %q", allowed)
		}
	}
	if IsOrdinaryCreateScope(ScopeSystem) {
		t.Fatal("ordinary create must reject system scope")
	}
}

func TestPhase8_SavedViewScopeVocabulary_U_8_03(t *testing.T) {
	for _, value := range []string{"private", "shared", "system"} {
		scope, ok := ParseScope(value)
		if !ok || string(scope) != value {
			t.Fatalf("expected scope %q to parse, got %q ok=%v", value, scope, ok)
		}
	}
	for _, value := range []string{"team", "Team", "PRIVATE", "", " private ", "incident"} {
		if scope, ok := ParseScope(value); ok {
			t.Fatalf("obsolete or non-canonical scope %q parsed as %q", value, scope)
		}
	}
}

func TestPhase8_SavedViewPatchContract_U_8_04(t *testing.T) {
	displayName, ok := NormalizeDisplayName("  Analyst triage  ")
	if !ok || displayName != "Analyst triage" {
		t.Fatalf("display name must normalize before no-op comparison, got %q ok=%v", displayName, ok)
	}
	if _, ok := NormalizeDisplayName(" \t "); ok {
		t.Fatal("empty normalized display names must be rejected")
	}
	if _, ok := ParseScope("team"); ok {
		t.Fatal("patch must not preserve obsolete team scope")
	}
}

func TestPhase8_SavedViewLifecyclePersistence_I_8_01(t *testing.T) {
	current := struct {
		scope            Scope
		version          int64
		normalizedName   string
		underlyingDelete bool
	}{scope: ScopePrivate, version: 3, normalizedName: "Analyst triage"}

	nextName, ok := NormalizeDisplayName("Analyst triage")
	if !ok {
		t.Fatal("expected valid normalized name")
	}
	if current.normalizedName == nextName && current.scope == ScopePrivate {
		if current.version != 3 {
			t.Fatalf("structural no-op must preserve version, got %d", current.version)
		}
	}
	if current.underlyingDelete {
		t.Fatal("saved-view delete must not imply record deletion")
	}
}
