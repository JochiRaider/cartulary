package restorecontract

import (
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

// GraphProjectionParticipant is the only Graph Projection runtime capability
// imported by Recovery. It remains separate from the workbook projection
// provider registry and its ProjectionRebuilder contract.
type GraphProjectionParticipant = graphprojection.RestoreParticipant

type GraphProjectionRebuildRequest = graphprojection.RestoreRebuildRequest
type GraphProjectionRebuildResult = graphprojection.RestoreRebuildResult
type GraphProjectionRecoveryCatalogRef = graphprojection.RestoreRecoveryCatalogRef
type GraphProjectionSourceRegistryRef = graphprojection.RestoreSourceRegistryRef
type GraphProjectionImplementationBindingRef = graphprojection.RestoreImplementationBindingRef

const GraphProjectionRestoreAlgorithmID = graphprojection.RestoreAlgorithmID

func GraphProjectionTableIDs() []string {
	return graphprojection.RestoreGraphTableIDs()
}

func CurrentGraphProjectionSourceRegistryRef() GraphProjectionSourceRegistryRef {
	registry := graphprojection.CurrentRestoreSourceRegistry()
	return graphprojection.RestoreSourceRegistryRef{Registry: registry, SHA256: registry.DigestSHA256()}
}

func CurrentGraphProjectionImplementationBinding() GraphProjectionImplementationBindingRef {
	return graphprojection.CurrentRestoreImplementationBinding()
}

// RestoredGraphProjectionSourceState is an opaque Recovery-owned capability;
// source-owner registrations hold their own narrow restored-state readers.
type RestoredGraphProjectionSourceState struct{}

func (RestoredGraphProjectionSourceState) GraphProjectionRestoreSourceState() {}

type GraphProjectionCompletionEvidence struct {
	TargetGenerationID          uuid.UUID                    `json:"target_generation_id"`
	RestoreOperationID          uuid.UUID                    `json:"restore_operation_id"`
	BackupSetID                 uuid.UUID                    `json:"backup_set_id"`
	ConsistencyPointAt          time.Time                    `json:"consistency_point_at"`
	RecoveryStateCatalogSHA256  string                       `json:"recovery_state_catalog_sha256"`
	SourceRegistrySHA256        string                       `json:"source_registry_sha256"`
	ImplementationBindingSHA256 string                       `json:"implementation_binding_sha256"`
	PostconditionSHA256         string                       `json:"postcondition_sha256"`
	ParticipantResult           GraphProjectionRebuildResult `json:"participant_result"`
}
