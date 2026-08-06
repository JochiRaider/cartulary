package server

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
)

func TestEvidenceComposition_ServerOwnsNarrowRuntime(t *testing.T) {
	runtimeType := reflect.TypeOf(Runtime{})
	wantFields := map[string]struct{}{
		"handler":                 {},
		"stagedJanitor":           {},
		"jobRunner":               {},
		"collaborationDispatcher": {},
		"processLease":            {},
		"servingLease":            {},
		"lifecycle":               {},
		"publication":             {},
		"publicHTTP":              {},
		"shutdownDrainTimeout":    {},
		"reconciliationTimeout":   {},
		"stagedObjectSweepPeriod": {},
		"closeOnce":               {},
		"publicationOnce":         {},
		"cleanups":                {},
		"stagedJanitorContext":    {},
	}
	if runtimeType.NumField() != len(wantFields) {
		t.Fatalf("server Runtime field count = %d, want %d", runtimeType.NumField(), len(wantFields))
	}
	for index := 0; index < runtimeType.NumField(); index++ {
		field := runtimeType.Field(index)
		if field.IsExported() {
			t.Fatalf("server Runtime field %q must remain private", field.Name)
		}
		if _, present := wantFields[field.Name]; !present {
			t.Fatalf("server Runtime retains construction-only field %q", field.Name)
		}
	}
	wantMethods := map[string]struct{}{
		"ActivatePublication": {},
		"Close":               {},
		"HTTPHandler":         {},
	}
	pointerType := reflect.TypeOf((*Runtime)(nil))
	if pointerType.NumMethod() != len(wantMethods) {
		t.Fatalf("server Runtime exported method count = %d, want %d", pointerType.NumMethod(), len(wantMethods))
	}
	for methodName := range wantMethods {
		if _, present := pointerType.MethodByName(methodName); !present {
			t.Fatalf("server Runtime is missing lifecycle method %q", methodName)
		}
	}
	if _, exposed := runtimeType.FieldByName("EvidenceStore"); exposed {
		t.Fatal("server Runtime unexpectedly exposes a production Evidence store")
	}
	if _, retained := runtimeType.FieldByName("evidenceOwner"); retained {
		t.Fatal("server Runtime retains the construction-only Evidence owner")
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
