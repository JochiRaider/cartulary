package harnessruntime

import "github.com/JochiRaider/cartulary/internal/platform/httpapi"

type Controls struct {
	PublicErrorFaults *PublicErrorFaultRegistry
}

type ControlContribution struct {
	Routes []httpapi.RouteRegistrar
}

func NewControls() *Controls {
	return &Controls{
		PublicErrorFaults: NewPublicErrorFaultRegistry(),
	}
}

func (c *Controls) Clear() {
	if c == nil {
		return
	}
	if c.PublicErrorFaults != nil {
		c.PublicErrorFaults.Clear()
	}
}

func RegisterRoutes(controls *Controls, clock *httpapi.TestClock, contributions ...ControlContribution) []httpapi.RouteRegistrar {
	if controls == nil {
		controls = NewControls()
	}
	routes := make([]httpapi.RouteRegistrar, 0, 3+len(contributions))
	if clock != nil {
		routes = append(routes, httpapi.RegisterTestClockRoutes(clock))
	}
	for _, contribution := range contributions {
		routes = append(routes, contribution.Routes...)
	}
	routes = append(routes,
		RegisterTestRuntimeIdentityRoute(),
		RegisterPublicErrorFaultRoutes(controls.PublicErrorFaults),
	)
	return routes
}
