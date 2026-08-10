package server

import (
	"reflect"
	"testing"
)

func TestSnapshotProcessEnvironmentCopiesExactEntries(t *testing.T) {
	entries := []string{
		"CARTULARY_SECRET_ALPHA=first",
		"CARTULARY_SECRET_ALPHA=second",
		"CARTULARY_SECRET_BETA=value=with=equals",
		"",
		"MISSING_SEPARATOR",
		"=missing-key",
	}
	want := map[string]string{
		"CARTULARY_SECRET_ALPHA": "second",
		"CARTULARY_SECRET_BETA":  "value=with=equals",
	}
	got := snapshotProcessEnvironment(entries)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("environment snapshot = %#v, want %#v", got, want)
	}

	entries[1] = "CARTULARY_SECRET_ALPHA=mutated"
	if got["CARTULARY_SECRET_ALPHA"] != "second" {
		t.Fatalf("environment snapshot changed after source mutation: %#v", got)
	}
}
