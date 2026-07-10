package harnessruntime

import "github.com/JochiRaider/cartulary/internal/platform/httpapi"

type Controls struct {
	PublicErrorFaults          *PublicErrorFaultRegistry
	NetworkFlowFaults          *NetworkFlowFaultRegistry
	NetworkFlowRandomness      *NetworkFlowRandomnessRegistry
	NetworkFlowTransitions     *NetworkFlowAuthTransitionRegistry
	NetworkFlowAuditAssertions *NetworkFlowAuditAssertionRegistry
}

func NewControls() *Controls {
	return &Controls{
		PublicErrorFaults:          NewPublicErrorFaultRegistry(),
		NetworkFlowFaults:          NewNetworkFlowFaultRegistry(),
		NetworkFlowRandomness:      NewNetworkFlowRandomnessRegistry(),
		NetworkFlowTransitions:     NewNetworkFlowAuthTransitionRegistry(),
		NetworkFlowAuditAssertions: NewNetworkFlowAuditAssertionRegistry(),
	}
}

func (c *Controls) Clear() {
	if c == nil {
		return
	}
	if c.PublicErrorFaults != nil {
		c.PublicErrorFaults.Clear()
	}
	if c.NetworkFlowFaults != nil {
		c.NetworkFlowFaults.Clear()
	}
	if c.NetworkFlowRandomness != nil {
		c.NetworkFlowRandomness.Clear()
	}
	if c.NetworkFlowTransitions != nil {
		c.NetworkFlowTransitions.Clear()
	}
	if c.NetworkFlowAuditAssertions != nil {
		c.NetworkFlowAuditAssertions.Clear()
	}
}

func RegisterRoutes(controls *Controls, clock *httpapi.TestClock) []httpapi.RouteRegistrar {
	if controls == nil {
		controls = NewControls()
	}
	resetHooks := []func(){controls.Clear}
	routes := make([]httpapi.RouteRegistrar, 0, 7)
	if clock != nil {
		routes = append(routes, httpapi.RegisterTestClockRoutes(clock))
		resetHooks = append(resetHooks, func() {
			_ = clock.Reset()
		})
	}
	routes = append(routes,
		RegisterTestRuntimeResetRoute(resetHooks...),
		RegisterPublicErrorFaultRoutes(controls.PublicErrorFaults),
		RegisterNetworkFlowFaultRoutes(controls.NetworkFlowFaults),
		RegisterNetworkFlowRandomnessRoutes(controls.NetworkFlowRandomness),
		RegisterNetworkFlowAuthTransitionRoutes(controls.NetworkFlowTransitions),
		RegisterNetworkFlowAuditAssertionRoutes(controls.NetworkFlowAuditAssertions),
	)
	return routes
}
