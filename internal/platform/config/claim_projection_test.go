package config

import (
	"reflect"
	"testing"
)

func TestBooleanValuesAtPathsProjectsOnlyRequestedClaims_Unit(t *testing.T) {
	cfg := Config{}
	cfg.Import.Claimed = true
	cfg.ReferencePack.Claimed = false
	values, err := BooleanValuesAtPaths(cfg, []string{"import.claimed", "reference_pack.claimed"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"import.claimed": true, "reference_pack.claimed": false}
	if !reflect.DeepEqual(values, want) {
		t.Fatalf("values = %#v, want %#v", values, want)
	}
}

func TestBooleanValuesAtPathsRejectsUnknownAndDuplicatePaths_Unit(t *testing.T) {
	for name, paths := range map[string][]string{
		"unknown":   {"future_profile.claimed"},
		"duplicate": {"import.claimed", "import.claimed"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := BooleanValuesAtPaths(Config{}, paths); err == nil {
				t.Fatal("expected projection rejection")
			}
		})
	}
}
