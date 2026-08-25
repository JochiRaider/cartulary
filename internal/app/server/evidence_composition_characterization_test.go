package server

import (
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestEvidenceComposition_ServerOwnsNarrowRuntime(t *testing.T) {
	runtimeType := reflect.TypeOf(Runtime{})
	wantFields := map[string]struct{}{
		"handler":                      {},
		"stagedJanitor":                {},
		"jobRunner":                    {},
		"collaborationRuntime":         {},
		"evidenceCleanupDispatcher":    {},
		"networkFlowCleanupDispatcher": {},
		"processLease":                 {},
		"servingLease":                 {},
		"lifecycle":                    {},
		"publication":                  {},
		"publicHTTP":                   {},
		"shutdownDrainTimeout":         {},
		"reconciliationTimeout":        {},
		"stagedObjectSweepPeriod":      {},
		"closeOnce":                    {},
		"publicationOnce":              {},
		"cleanups":                     {},
		"stagedJanitorContext":         {},
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
	fixtureField, present := timelineBundleType.FieldByName("PerformanceFixture")
	if !present {
		t.Fatal("Timeline bundle does not expose its isolated performance fixture contribution")
	}
	if got, want := fixtureField.Type.String(), "*timeline.PerformanceFixtureContribution"; got != want {
		t.Fatalf("Timeline performance fixture contribution type = %s, want %s", got, want)
	}
	ownerType := reflect.TypeOf(evidence.OwnerRuntime{})
	wantOwnerFields := map[string]reflect.Type{
		"postgres":     reflect.TypeOf((*postgres.DB)(nil)).Elem(),
		"objectStore":  reflect.TypeOf((*objectstore.TypedStore)(nil)).Elem(),
		"now":          reflect.TypeOf((func() time.Time)(nil)),
		"routes":       nil,
		"workbook":     reflect.TypeOf((*evidence.MutationContribution)(nil)).Elem(),
		"attachments":  reflect.TypeOf((*evidence.TimelineAttachmentContribution)(nil)).Elem(),
		"importCreate": reflect.TypeOf((*ownerfacade.ImportOwnerCreateFacade)(nil)).Elem(),
		"cleanup":      reflect.TypeOf((*evidence.CleanupDispatcher)(nil)),
	}
	if ownerType.NumField() != len(wantOwnerFields) {
		t.Fatalf("Evidence owner runtime field count = %d, want %d", ownerType.NumField(), len(wantOwnerFields))
	}
	for fieldName, wantType := range wantOwnerFields {
		field, present := ownerType.FieldByName(fieldName)
		if !present {
			t.Fatalf("Evidence owner runtime is missing narrow capability %q", fieldName)
		}
		if field.IsExported() {
			t.Fatalf("Evidence owner runtime field %q must remain private", fieldName)
		}
		if wantType == nil && field.Type.Kind() != reflect.Interface {
			t.Fatalf("Evidence owner runtime field %q type = %v, want private interface", fieldName, field.Type)
		}
		if wantType != nil && field.Type != wantType {
			t.Fatalf("Evidence owner runtime field %q type = %v, want %v", fieldName, field.Type, wantType)
		}
	}
}
