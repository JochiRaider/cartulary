package configassembly

import (
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
)

// AdmissionConfigurationValue is the configuration owner's normalized projection.
// Application composition adapts it to the Extensions validation port.
type AdmissionConfigurationValue struct {
	Key    string
	Source string
	Value  any
}

// NetworkFlowAdmissionConfiguration projects normalized owner values. The file
// handle names the configuration key; the absolute path stays inside Core's
// process-local file-availability boundary.
func NetworkFlowAdmissionConfiguration(configuration networkflow.Configuration) []AdmissionConfigurationValue {
	values := []AdmissionConfigurationValue{{Key: "network_flow_activity.key_ring_manifest_path", Source: "explicit", Value: map[string]any{"kind": "regular_file_ref", "key": "network_flow_activity.key_ring_manifest_path"}}}
	if configuration.ResourceLimits != nil {
		values = append(values, AdmissionConfigurationValue{Key: "network_flow_activity.resource_limits", Source: "explicit", Value: configuration.ResourceLimits})
	}
	return values
}
