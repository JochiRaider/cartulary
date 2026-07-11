package harnesscontrol

import (
	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type Controls struct {
	Faults          *NetworkFlowFaultRegistry
	Randomness      *NetworkFlowRandomnessRegistry
	Transitions     *NetworkFlowAuthTransitionRegistry
	AuditAssertions *NetworkFlowAuditAssertionRegistry
}

func NewControls() *Controls {
	return &Controls{
		Faults:          NewNetworkFlowFaultRegistry(),
		Randomness:      NewNetworkFlowRandomnessRegistry(),
		Transitions:     NewNetworkFlowAuthTransitionRegistry(),
		AuditAssertions: NewNetworkFlowAuditAssertionRegistry(),
	}
}

func (c *Controls) Clear() {
	if c == nil {
		return
	}
	if c.Faults != nil {
		c.Faults.Clear()
	}
	if c.Randomness != nil {
		c.Randomness.Clear()
	}
	if c.Transitions != nil {
		c.Transitions.Clear()
	}
	if c.AuditAssertions != nil {
		c.AuditAssertions.Clear()
	}
}

func (c *Controls) Contribution() harnessruntime.ControlContribution {
	if c == nil {
		c = NewControls()
	}
	return harnessruntime.ControlContribution{
		Routes: []httpapi.RouteRegistrar{
			RegisterNetworkFlowFaultRoutes(c.Faults),
			RegisterNetworkFlowRandomnessRoutes(c.Randomness),
			RegisterNetworkFlowAuthTransitionRoutes(c.Transitions),
			RegisterNetworkFlowAuditAssertionRoutes(c.AuditAssertions),
		},
		ResetHooks: []func(){c.Clear},
	}
}
