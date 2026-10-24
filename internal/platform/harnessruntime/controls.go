package harnessruntime

import "github.com/JochiRaider/cartulary/internal/platform/httpapi"

type ControlContribution struct {
	Routes []httpapi.RouteRegistrar
}

func RegisterRoutes(clock *httpapi.TestClock, contributions ...ControlContribution) []httpapi.RouteRegistrar {
	routes := make([]httpapi.RouteRegistrar, 0, 2+len(contributions))
	if clock != nil {
		routes = append(routes, httpapi.RegisterTestClockRoutes(clock))
	}
	for _, contribution := range contributions {
		routes = append(routes, contribution.Routes...)
	}
	return append(routes, RegisterTestRuntimeIdentityRoute())
}
