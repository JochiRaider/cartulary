package revisionsupport

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

type Composition struct {
	Runtime       *revisionassembly.Runtime
	RecordChanges collaboration.RecordChangedAppender
}

func NewComposition() (Composition, error) {
	recordChanges := collaborationsupport.NewRecordChangedAppender()
	contributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		return Composition{}, err
	}
	runtime, err := revisionassembly.Build(contributions...)
	if err != nil {
		return Composition{}, err
	}
	return Composition{Runtime: runtime, RecordChanges: recordChanges}, nil
}

func NewRuntime() (*revisionassembly.Runtime, error) {
	composition, err := NewComposition()
	return composition.Runtime, err
}

func MustComposition(t testing.TB) Composition {
	t.Helper()
	composition, err := NewComposition()
	if err != nil {
		t.Fatalf("compose test Revisions runtime: %v", err)
	}
	return composition
}

func MustRuntime(t testing.TB) *revisionassembly.Runtime {
	t.Helper()
	return MustComposition(t).Runtime
}

func MustAppender(t testing.TB) *revisions.Appender {
	t.Helper()
	return MustRuntime(t).Appender()
}

func NewConflictFieldResolver() (conflicts.FieldResolver, error) {
	runtime, err := NewRuntime()
	if err != nil {
		return nil, err
	}
	return runtime.ConflictFieldResolver(), nil
}

func MustConflictFieldResolver(t testing.TB) conflicts.FieldResolver {
	t.Helper()
	resolver, err := NewConflictFieldResolver()
	if err != nil {
		t.Fatalf("compose test Revisions conflict-field resolver: %v", err)
	}
	return resolver
}

func MustTargetSemanticsCatalog(t testing.TB) *revisions.TargetSemanticsCatalog {
	t.Helper()
	contributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		t.Fatalf("compose current Revisions provider contributions: %v", err)
	}
	catalog, err := revisions.NewTargetSemanticsCatalog(contributions)
	if err != nil {
		t.Fatalf("compose test Revisions target-semantics catalog: %v", err)
	}
	return catalog
}
