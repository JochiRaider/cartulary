package server

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestEvidenceComposition_ServerOwnsNarrowRuntime(t *testing.T) {
	runtimeType := reflect.TypeOf(Runtime{})
	forbidden := []reflect.Type{
		reflect.TypeOf((*evidence.OwnerRuntime)(nil)), reflect.TypeOf((*evidence.SourceRuntime)(nil)),
		reflect.TypeOf((*postgres.DB)(nil)).Elem(), reflect.TypeOf((*objectstore.TypedStore)(nil)).Elem(),
	}
	if violations := runtimeFieldBoundaryViolations(runtimeType, forbidden...); len(violations) != 0 {
		t.Fatalf("server retains exposed storage or construction capabilities: %v", violations)
	}
	// Renaming or extracting private lifecycle state is harmless; exposing state
	// or retaining a generic persistence/construction capability is not.
	if got := runtimeFieldBoundaryViolations(reflect.TypeOf(struct {
		renamed   int
		extracted struct{ state bool }
	}{}), forbidden...); len(got) != 0 {
		t.Fatal("private implementation extraction rejected", got)
	}
	for _, candidate := range []any{struct{ Exposed int }{}, struct{ db postgres.DB }{}, struct{ owner *evidence.OwnerRuntime }{}} {
		if got := runtimeFieldBoundaryViolations(reflect.TypeOf(candidate), forbidden...); len(got) == 0 {
			t.Fatal("prohibited runtime capability escaped", reflect.TypeOf(candidate))
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
	if violations := runtimeFieldBoundaryViolations(ownerType); len(violations) != 0 {
		t.Fatalf("Evidence owner runtime exposes private state: %v", violations)
	}
	wantOwnerMethods := map[string]reflect.Type{
		"RouteRegistrar":                 reflect.TypeOf((httpapi.RouteRegistrar)(nil)),
		"MutationContribution":           reflect.TypeOf((*evidence.MutationContribution)(nil)).Elem(),
		"TimelineAttachmentContribution": reflect.TypeOf((*evidence.TimelineAttachmentContribution)(nil)).Elem(),
		"ImportCreateFacade":             reflect.TypeOf((*ownerfacade.ImportOwnerCreateFacade)(nil)).Elem(),
		"CleanupDispatcher":              reflect.TypeOf((*evidence.CleanupDispatcher)(nil)),
	}
	ownerPointer := reflect.PointerTo(ownerType)
	if ownerPointer.NumMethod() != len(wantOwnerMethods) {
		t.Fatal("Evidence owner public capability set changed")
	}
	for name, want := range wantOwnerMethods {
		method, ok := ownerPointer.MethodByName(name)
		if !ok || method.Type.NumOut() != 1 || method.Type.Out(0) != want {
			t.Fatalf("Evidence owner capability %s no longer exposes its narrow interface", name)
		}
	}
}

func runtimeFieldBoundaryViolations(runtime reflect.Type, forbidden ...reflect.Type) []string {
	var violations []string
	for index := 0; index < runtime.NumField(); index++ {
		field := runtime.Field(index)
		invalid := field.IsExported()
		for _, capability := range forbidden {
			invalid = invalid || field.Type == capability
		}
		if invalid {
			violations = append(violations, field.Name)
		}
	}
	return violations
}
