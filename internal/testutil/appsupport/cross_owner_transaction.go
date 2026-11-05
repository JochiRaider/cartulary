package appsupport

import (
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
)

// CrossOwnerTransactionOptions composes the production registry and validation
// policy for owner tests. The caller supplies the transaction backend.
func CrossOwnerTransactionOptions(t testing.TB, networkFlow *networkflow.Module) crossownertransaction.Options {
	t.Helper()
	registry, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	return crossownertransaction.Options{
		Limits:           extensionassembly.CrossOwnerLimits(registry),
		ValidationPolicy: extensionassembly.CrossOwnerValidationPolicy{Coordinator: registry, NetworkFlow: networkFlow},
		Catalog:          extensionassembly.CrossOwnerDescriptors(registry.ParticipantContracts()),
		Timeout:          time.Minute,
		FatalSink:        func(err error) { t.Errorf("unexpected fatal transaction: %v", err) },
	}
}
