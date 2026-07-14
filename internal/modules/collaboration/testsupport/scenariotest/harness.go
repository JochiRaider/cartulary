package scenariotest

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

type RuntimeHarness struct {
	*appsupport.Runtime
}

type ServerHarness = appsupport.ServerHarness

func StartRuntime(t testing.TB) *RuntimeHarness {
	t.Helper()
	return &RuntimeHarness{Runtime: appsupport.StartRuntime(t)}
}

func (h *RuntimeHarness) StartServer(t testing.TB, prefix string) *ServerHarness {
	t.Helper()
	return h.Runtime.StartServer(t, appsupport.ServerOptions{
		Prefix:        prefix,
		Dependencies:  httpapi.DependencySet{},
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
}
