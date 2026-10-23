package pgtest

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

func emptyPublication(status string) suiteservices.PostgresPublicationObservation {
	return suiteservices.PostgresPublicationObservation{Status: status, RequestedPorts: []string{}, EffectivePorts: []string{}, Bindings: []suiteservices.PostgresPublicationBinding{}}
}

// The observer only reads the final create request. Invalid or repeated input
// loses diagnostic coverage; it never changes provisioning or cleanup.
func (d *startupDiagnostic) publicationObserver(requested []string) testcontainersx.CreateObserver {
	requested = slices.Clone(requested)
	called := false
	return func(options client.ContainerCreateOptions) {
		d.Publication = emptyPublication("unavailable")
		if called {
			return
		}
		called = true
		if options.Config == nil || options.HostConfig == nil || len(requested) > 16 || len(options.Config.ExposedPorts) > 16 || len(options.HostConfig.PortBindings) > 32 {
			return
		}
		o := emptyPublication("observed")
		for _, raw := range requested {
			port, err := network.ParsePort(raw)
			if err != nil || !observablePort(port) || port.String() != raw {
				return
			}
			o.RequestedPorts = append(o.RequestedPorts, raw)
		}
		for port := range options.Config.ExposedPorts {
			if !observablePort(port) {
				return
			}
			o.EffectivePorts = append(o.EffectivePorts, port.String())
		}
		for port, bindings := range options.HostConfig.PortBindings {
			if !observablePort(port) || len(o.Bindings)+len(bindings) > 32 {
				return
			}
			for _, binding := range bindings {
				scope, kind := "unspecified", "ephemeral"
				if binding.HostIP.IsLoopback() {
					scope = "loopback"
				} else if binding.HostIP.IsValid() && !binding.HostIP.IsUnspecified() {
					scope = "specific"
				}
				if binding.HostPort != "" && binding.HostPort != "0" {
					value, err := strconv.ParseUint(binding.HostPort, 10, 16)
					if err != nil || value == 0 || strconv.FormatUint(value, 10) != binding.HostPort {
						return
					}
					kind = "fixed"
				}
				o.Bindings = append(o.Bindings, suiteservices.PostgresPublicationBinding{ContainerPort: port.String(), AddressScope: scope, PortKind: kind})
			}
		}
		slices.Sort(o.RequestedPorts)
		o.RequestedPorts = slices.Compact(o.RequestedPorts)
		slices.Sort(o.EffectivePorts)
		slices.SortFunc(o.Bindings, func(a, b suiteservices.PostgresPublicationBinding) int {
			return cmp.Or(cmp.Compare(a.ContainerPort, b.ContainerPort), cmp.Compare(a.AddressScope, b.AddressScope), cmp.Compare(a.PortKind, b.PortKind))
		})
		publishAll := options.HostConfig.PublishAllPorts
		o.PublishAllPorts = &publishAll
		d.Publication = o
	}
}

func observablePort(port network.Port) bool {
	return port.IsValid() && port.Num() != 0 && (port.Proto() == network.TCP || port.Proto() == network.UDP)
}
