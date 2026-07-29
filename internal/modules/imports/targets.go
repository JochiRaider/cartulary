package imports

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/gen/importtargetregistry"
)

const (
	applyStatusSupported              = "supported"
	applyStatusSupportedWhenAvailable = "supported_when_implemented"
	applyStatusSupportedWhenClaimed   = "supported_when_claimed"
)

type importTarget struct {
	TargetKind         string
	ViewSchemaID       string
	ExtensionProfileID string
	ApplyStatus        string
	CreateFacade       string
	ApplyFacade        string
	AllowRawCapture    bool
	AllowCustomAttrs   bool
}

func (target importTarget) importable(profileClaimed func(string) bool) bool {
	switch target.ApplyStatus {
	case applyStatusSupported:
		return true
	case applyStatusSupportedWhenClaimed:
		return profileClaimed != nil && profileClaimed(target.ExtensionProfileID)
	default:
		return false
	}
}

func (target importTarget) readyCheckImportable() bool {
	return target.ApplyStatus == applyStatusSupported ||
		target.ApplyStatus == applyStatusSupportedWhenClaimed
}

func (target importTarget) ownerCreateFacadeAvailable() bool {
	return target.CreateFacade != ""
}

func (target importTarget) ownerApplyFacadeAvailable() bool {
	return target.ApplyFacade != ""
}

func lookupImportTarget(viewSchemaID string) (importTarget, bool) {
	target, ok := importTargets[viewSchemaID]
	return target, ok
}

func lookupApprovedImportTarget(mapping ApprovedMapping) (importTarget, bool) {
	switch mapping.targetKindOrDefault() {
	case ImportTargetKindViewSchema:
		return lookupImportTarget(mapping.TargetViewSchemaID)
	case ImportTargetKindNetworkFlowTable:
		target, ok := analyticalImportTargets[analyticalImportTargetKey{
			TargetKind:         mapping.TargetKind,
			ExtensionProfileID: mapping.ExtensionProfileID,
		}]
		return target, ok
	default:
		return importTarget{}, false
	}
}

type analyticalImportTargetKey struct {
	TargetKind         string
	ExtensionProfileID string
}

var importTargets, analyticalImportTargets = mustGeneratedImportTargets()

func mustGeneratedImportTargets() (
	map[string]importTarget,
	map[analyticalImportTargetKey]importTarget,
) {
	viewTargets := make(map[string]importTarget)
	analyticalTargets := make(map[analyticalImportTargetKey]importTarget)
	for _, generated := range importtargetregistry.Targets {
		status := generatedImportApplyStatus(generated.AvailabilityKind)
		switch generated.TargetKind {
		case ImportTargetKindViewSchema:
			if generated.TargetViewSchemaID == nil {
				panic(fmt.Sprintf(
					"generated import target %s has no view schema id",
					generated.TargetID,
				))
			}
			target := importTarget{
				TargetKind:       generated.TargetKind,
				ViewSchemaID:     *generated.TargetViewSchemaID,
				ApplyStatus:      status,
				CreateFacade:     generatedString(generated.FacadeID),
				AllowRawCapture:  generated.DefaultUnknownColumnPolicy == "preserve_raw_capture",
				AllowCustomAttrs: false,
			}
			if _, exists := viewTargets[target.ViewSchemaID]; exists {
				panic(fmt.Sprintf(
					"generated import target registry duplicated %s",
					target.ViewSchemaID,
				))
			}
			viewTargets[target.ViewSchemaID] = target
		case ImportTargetKindNetworkFlowTable:
			if generated.ExtensionProfileID == nil {
				panic(fmt.Sprintf(
					"generated analytical import target %s has no extension profile id",
					generated.TargetID,
				))
			}
			target := importTarget{
				TargetKind:         generated.TargetKind,
				ExtensionProfileID: *generated.ExtensionProfileID,
				ApplyStatus:        status,
				ApplyFacade:        generatedString(generated.FacadeID),
			}
			key := analyticalImportTargetKey{
				TargetKind:         target.TargetKind,
				ExtensionProfileID: target.ExtensionProfileID,
			}
			if _, exists := analyticalTargets[key]; exists {
				panic(fmt.Sprintf(
					"generated import target registry duplicated %s/%s",
					key.TargetKind,
					key.ExtensionProfileID,
				))
			}
			analyticalTargets[key] = target
		default:
			panic(fmt.Sprintf(
				"generated import target %s has unsupported kind %s",
				generated.TargetID,
				generated.TargetKind,
			))
		}
	}
	return viewTargets, analyticalTargets
}

func generatedImportApplyStatus(availability string) string {
	switch availability {
	case "enabled":
		return applyStatusSupported
	case "reserved":
		return applyStatusSupportedWhenAvailable
	case "claim_gated":
		return applyStatusSupportedWhenClaimed
	default:
		panic(fmt.Sprintf(
			"generated import target has unsupported availability %s",
			availability,
		))
	}
}

func generatedString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
