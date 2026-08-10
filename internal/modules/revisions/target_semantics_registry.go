package revisions

import (
	"encoding/json"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/gen/contractrevisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

const targetSemanticsRegistrySchemaID = "cartulary.revisions_target_semantics_registry.v1"

type targetSemanticsRequirement struct {
	TargetKind            string
	SourceOwnerID         string
	DispatchClass         rollbackcontract.DispatchClass
	AdmittedRecordTypes   []string
	HistoryRecordIDFields []string
	Addressability        HistoryAddressability
}

type targetSemanticsRegistry struct {
	SchemaID        string                         `json:"schema_id"`
	RegistryVersion int                            `json:"registry_version"`
	Targets         []targetSemanticsRegistryEntry `json:"targets"`
}

type targetSemanticsRegistryEntry struct {
	TargetKind            string                         `json:"target_kind"`
	SourceOwner           string                         `json:"source_owner"`
	DispatchClass         rollbackcontract.DispatchClass `json:"dispatch_class"`
	AdmittedRecordTypes   []string                       `json:"admitted_record_types"`
	HistoryRecordIDFields []string                       `json:"history_record_id_fields"`
	Addressability        HistoryAddressability          `json:"addressability"`
}

func parseTargetSemanticsRequirements(data []byte) ([]targetSemanticsRequirement, error) {
	var registry targetSemanticsRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("%w: decode registry: %v", ErrInvalidTargetSemantics, err)
	}
	if registry.SchemaID != targetSemanticsRegistrySchemaID || registry.RegistryVersion != 1 || len(registry.Targets) == 0 {
		return nil, fmt.Errorf("%w: registry identity", ErrInvalidTargetSemantics)
	}
	requirements := make([]targetSemanticsRequirement, 0, len(registry.Targets))
	for _, entry := range registry.Targets {
		requirements = append(requirements, targetSemanticsRequirement{
			TargetKind:            entry.TargetKind,
			SourceOwnerID:         entry.SourceOwner,
			DispatchClass:         entry.DispatchClass,
			AdmittedRecordTypes:   append([]string(nil), entry.AdmittedRecordTypes...),
			HistoryRecordIDFields: append([]string(nil), entry.HistoryRecordIDFields...),
			Addressability:        entry.Addressability,
		})
	}
	return requirements, nil
}

func currentTargetSemanticsRequirements() ([]targetSemanticsRequirement, error) {
	artifact, ok := contractrevisions.Index["contracts/revisions/target-semantics-registry.v1.json"]
	if !ok {
		return nil, ErrMissingTargetSemantics
	}
	return parseTargetSemanticsRequirements([]byte(artifact.JSON))
}
