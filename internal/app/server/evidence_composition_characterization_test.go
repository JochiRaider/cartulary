package server

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
)

func TestEvidenceComposition_ServerOwnsNarrowRuntime(t *testing.T) {
	runtimeType := reflect.TypeOf(Runtime{})
	if _, exposed := runtimeType.FieldByName("EvidenceStore"); exposed {
		t.Fatal("server Runtime unexpectedly exposes a production Evidence store")
	}
	ownerField, present := runtimeType.FieldByName("evidenceOwner")
	if !present || ownerField.Type != reflect.TypeOf((*evidence.OwnerRuntime)(nil)) {
		t.Fatalf("server Runtime Evidence owner field = %#v, want private *evidence.OwnerRuntime", ownerField)
	}

	timelineBundleType := reflect.TypeOf(timelineassembly.Bundle{})
	if _, exposed := timelineBundleType.FieldByName("EvidenceStore"); exposed {
		t.Fatal("Timeline bundle exposes an Evidence store")
	}
	ownerType := reflect.TypeOf(evidence.OwnerRuntime{})
	for index := 0; index < ownerType.NumField(); index++ {
		if ownerType.Field(index).Type == reflect.TypeOf((*evidence.Store)(nil)) {
			t.Fatalf("Evidence owner runtime field %q exposes *evidence.Store", ownerType.Field(index).Name)
		}
	}
}
