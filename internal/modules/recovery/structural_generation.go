package recovery

import "github.com/google/uuid"

// RecoveryJournalPayloadSchemaID participates in the current generation's codec
// identity as well as the canonical terminal envelope carried by application.
const RecoveryJournalPayloadSchemaID = "cartulary.operator_recovery_journal_payload.v7"

// RestoreGenerationEvidence authenticates structural completion independently
// of the availability or reconstruction of extension-derived projections.
type RestoreGenerationEvidence struct {
	TargetGenerationID         uuid.UUID `json:"target_generation_id"`
	RecoveryStateCatalogSHA256 string    `json:"recovery_state_catalog_sha256"`
	CodecRegistrySHA256        string    `json:"codec_registry_sha256"`
}

func (identity RecoveryGenerationIdentity) AdmitsRestoreGeneration(evidence RestoreGenerationEvidence) bool {
	return evidence.TargetGenerationID != uuid.Nil &&
		identity.RecoveryStateCatalogSHA256 == evidence.RecoveryStateCatalogSHA256 &&
		identity.CodecRegistrySHA256 == evidence.CodecRegistrySHA256
}
