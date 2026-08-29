package revisionsupport

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
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
