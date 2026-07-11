package scenariotest

import (
	"testing"

	apptestsupport "github.com/JochiRaider/cartulary/internal/app/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

type RuntimeHarness struct {
	*apptestsupport.Runtime
}

type ServerHarness = apptestsupport.ServerHarness

func StartRuntime(t testing.TB) *RuntimeHarness {
	t.Helper()
	return &RuntimeHarness{Runtime: apptestsupport.StartRuntime(t)}
}

func (h *RuntimeHarness) StartServer(t testing.TB, prefix string) *ServerHarness {
	t.Helper()
	return h.Runtime.StartServer(t, apptestsupport.ServerOptions{
		Prefix:        prefix,
		Dependencies:  httpapi.DependencySet{},
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
}
