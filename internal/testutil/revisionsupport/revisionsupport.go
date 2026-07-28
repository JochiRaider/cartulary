package revisionsupport

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type Composition struct {
	Runtime *revisionassembly.Runtime
	Intents collaboration.IntentAppender
}

func NewComposition() (Composition, error) {
	intents := collaboration.NewIntentAppender()
	runtime, err := revisionassembly.Build(
		revisionassembly.Dependencies{
			HistoricalIntentPolicy: collaboration.NewHistoricalIntentPolicy(),
			IntentAppender:         intents,
		},
		revisionassembly.CurrentProviderContributions()...,
	)
	if err != nil {
		return Composition{}, err
	}
	return Composition{Runtime: runtime, Intents: intents}, nil
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
