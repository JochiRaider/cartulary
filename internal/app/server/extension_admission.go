package server

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func networkFlowAdmissionPorts(coordinator *extensions.Coordinator, store *extensionstore.Store, configuration networkflow.Configuration, definitions []jobs.Definition) (map[string]extensions.ProfileConfigurationView, map[string]extensions.AdmissionAlgorithm, error) {
	views := map[string]extensions.ProfileConfigurationView{}
	if configuration.Claimed {
		values := []extensions.ProfileConfigurationValue{}
		for _, value := range configassembly.NetworkFlowAdmissionConfiguration(configuration) {
			values = append(values, extensions.ProfileConfigurationValue{Key: value.Key, Source: value.Source, Value: value.Value})
		}
		view, err := coordinator.ConfigurationView("network_flow_activity", values)
		if err != nil {
			return nil, nil, err
		}
		views["network_flow_activity"] = view
	}
	var graphDefinition jobs.Definition
	for _, definition := range definitions {
		if definition.JobKind == networkflow.GraphViewMaterializationJobKind {
			graphDefinition = definition
		}
	}
	algorithms := map[string]extensions.AdmissionAlgorithm{
		networkflow.SavedGraphCutoverAlgorithmID: func(ctx context.Context, input extensions.AdmissionValidationContext) (extensions.AdmissionValidationResult, error) {
			err := store.WithAdmissionRead(ctx, func(reader extensionstore.Querier) error {
				return networkflow.ValidateSavedGraphAdmission(ctx, reader, graphDefinition)
			})
			return extensions.ValidAdmissionResult(input, networkflow.SavedGraphCutoverAlgorithmID), err
		},
		"network_flow_activity.validate_state_v4": func(ctx context.Context, input extensions.AdmissionValidationContext) (extensions.AdmissionValidationResult, error) {
			err := store.WithAdmissionRead(ctx, func(reader extensionstore.Querier) error {
				return networkflow.ValidateRetainedExtensionState(ctx, reader)
			})
			return extensions.ValidAdmissionResult(input, "network_flow_activity.validate_state_v4"), err
		},
	}
	return views, algorithms, nil
}
