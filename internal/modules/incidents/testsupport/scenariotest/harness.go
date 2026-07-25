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

	return h.startServer(t, prefix, httpapi.DependencySet{}, httptestx.TestRouteModeDisabled)
}

func (h *RuntimeHarness) StartServerWithRoutes(t testing.TB, prefix string, routes ...httpapi.RouteRegistrar) *ServerHarness {
	t.Helper()

	return h.startServer(t, prefix, httpapi.DependencySet{}, httptestx.TestRouteModeHarnessOwned, routes...)
}

func (h *RuntimeHarness) StartServerWithHarnessControls(t testing.TB, prefix string) *ServerHarness {
	t.Helper()

	return h.startServer(t, prefix, httpapi.DependencySet{}, httptestx.TestRouteModeHarnessOwned)
}

func (h *RuntimeHarness) StartServerWithDependencies(t testing.TB, prefix string, deps httpapi.DependencySet) *ServerHarness {
	t.Helper()

	return h.startServer(t, prefix, deps, httptestx.TestRouteModeDisabled)
}

func (h *RuntimeHarness) StartServerWithDependenciesAndEnv(t testing.TB, prefix string, deps httpapi.DependencySet, env map[string]string) *ServerHarness {
	t.Helper()

	return h.Runtime.StartServer(t, appsupport.ServerOptions{
		Prefix:        prefix,
		Env:           env,
		Dependencies:  deps,
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
}

func (h *RuntimeHarness) StartServerWithTestDependencies(t testing.TB, prefix string, deps httpapi.DependencySet) *ServerHarness {
	t.Helper()

	return h.startServer(t, prefix, deps, httptestx.TestRouteModeHarnessOwned)
}

func (h *RuntimeHarness) startServer(t testing.TB, prefix string, deps httpapi.DependencySet, routeMode httptestx.TestRouteMode, routes ...httpapi.RouteRegistrar) *ServerHarness {
	t.Helper()

	return h.Runtime.StartServer(t, appsupport.ServerOptions{
		Prefix:           prefix,
		Dependencies:     deps,
		AdditionalRoutes: routes,
		TestRouteMode:    routeMode,
	})
}
