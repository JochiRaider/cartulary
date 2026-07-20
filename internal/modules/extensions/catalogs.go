package extensions

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

type StatePlan struct {
	ProfileID                      string
	ContractMajor                  int
	MigrationLineageID             string
	CurrentStateVersion            int
	MinimumMigratableStateVersion  int
	EmptyStatePolicy               string
	DatabaseFamilyIDs              []string
	ObjectReferenceFamilyIDs       []string
	InitializationKind             string
	InitializationDefinitionSHA256 string
	FinalValidationAlgorithmID     string
	PhysicalStateBindingSHA256     string
	StatePresenceManifestSHA256    string
	MigrationDefinitions           []MigrationDefinition
}

type MigrationDefinition struct {
	MigrationID      string
	FromVersion      int
	ToVersion        int
	DefinitionSHA256 string
}

func cloneStatePlan(plan StatePlan) StatePlan {
	plan.DatabaseFamilyIDs = append([]string(nil), plan.DatabaseFamilyIDs...)
	plan.ObjectReferenceFamilyIDs = append([]string(nil), plan.ObjectReferenceFamilyIDs...)
	plan.MigrationDefinitions = append([]MigrationDefinition(nil), plan.MigrationDefinitions...)
	return plan
}

type BackupBinding struct {
	BindingID                        string
	LogicalFamilyID                  string
	StorageKind                      string
	PhysicalRef                      string
	StateClass                       string
	BackupInclusion                  string
	RestoreOrderGroup                int
	BackupCodecID                    string
	BackupCodecSHA256                string
	PostRestoreValidationAlgorithmID string
	RebuildAlgorithmID               string
}

type BackupCodec struct {
	CodecID                string
	SHA256                 string
	BindingID              string
	StorageKind            string
	MaxItems               int
	MaxEntryBytes          int64
	MaxBindingBytes        int64
	HistoricalCodecDigests []string
}

type BackupPlan struct {
	ProfileID                  string
	PhysicalStateBindingSHA256 string
	Bindings                   []BackupBinding
	Codecs                     map[string]BackupCodec
}

func cloneBackupPlan(plan BackupPlan) BackupPlan {
	plan.Bindings = append([]BackupBinding(nil), plan.Bindings...)
	plan.Codecs = cloneBackupCodecs(plan.Codecs)
	return plan
}

func cloneBackupCodecs(source map[string]BackupCodec) map[string]BackupCodec {
	result := make(map[string]BackupCodec, len(source))
	for key, codec := range source {
		codec.HistoricalCodecDigests = append([]string(nil), codec.HistoricalCodecDigests...)
		result[key] = codec
	}
	return result
}

type ParticipantContract struct {
	ParticipantID         string
	OwnerProfileID        string
	ContractSHA256        string
	ContractKind          string
	InputSchemaID         string
	AlgorithmIDs          []string
	SerializationKeyKinds []string
	OwnedStateFamilyIDs   []string
}

func cloneParticipantContract(contract ParticipantContract) ParticipantContract {
	contract.AlgorithmIDs = append([]string(nil), contract.AlgorithmIDs...)
	contract.SerializationKeyKinds = append([]string(nil), contract.SerializationKeyKinds...)
	contract.OwnedStateFamilyIDs = append([]string(nil), contract.OwnedStateFamilyIDs...)
	return contract
}

func (c *Coordinator) StatePlan(profileID string) (StatePlan, bool) {
	if c == nil {
		return StatePlan{}, false
	}
	plan, ok := c.statePlans[profileID]
	return cloneStatePlan(plan), ok
}

func (c *Coordinator) BackupPlan(profileID string) (BackupPlan, bool) {
	if c == nil {
		return BackupPlan{}, false
	}
	plan, ok := c.backupPlans[profileID]
	return cloneBackupPlan(plan), ok
}

func (c *Coordinator) ParticipantContract(participantID string) (ParticipantContract, bool) {
	if c == nil {
		return ParticipantContract{}, false
	}
	contract, ok := c.participantContracts[participantID]
	return cloneParticipantContract(contract), ok
}

func (c *Coordinator) admitRuntimeCatalogs(source ArtifactSource, supportDigests map[string]string) error {
	stateRegistry, stateArtifact, err := readArtifactObject(source, generatedExtensionsRoot+"state-registry.json")
	if err != nil {
		return err
	}
	if stateRegistry["schema_id"] != "cartulary.extension_state_registry.v1" || supportDigests["state-registry"] != stateArtifact.SHA256 {
		return invalidArtifact("state_registry", errors.New("integrity digest mismatch"))
	}
	c.statePlans = map[string]StatePlan{}
	stateRows, ok := objectSlice(stateRegistry["profiles"])
	if !ok {
		return invalidArtifact("state_registry", errors.New("profiles must be an array"))
	}
	for _, row := range stateRows {
		plan, parseErr := parseStatePlan(row)
		if parseErr != nil {
			return invalidArtifact("state_registry", parseErr)
		}
		if _, duplicate := c.statePlans[plan.ProfileID]; duplicate {
			return collisionFailure("state_profile_id", plan.ProfileID)
		}
		record, recognized := c.profiles[plan.ProfileID]
		if !recognized || record.descriptor.ContractMajor != plan.ContractMajor {
			return unavailableBinding(plan.ProfileID, "state_contract_major_mismatch")
		}
		c.statePlans[plan.ProfileID] = plan
	}

	backupRegistry, backupArtifact, err := readArtifactObject(source, generatedExtensionsRoot+"backup-registry.json")
	if err != nil {
		return err
	}
	if backupRegistry["schema_id"] != "cartulary.extension_backup_registry.v1" || supportDigests["backup-registry"] != backupArtifact.SHA256 {
		return invalidArtifact("backup_registry", errors.New("integrity digest mismatch"))
	}
	c.backupPlans = map[string]BackupPlan{}
	backupRows, ok := objectSlice(backupRegistry["profiles"])
	if !ok {
		return invalidArtifact("backup_registry", errors.New("profiles must be an array"))
	}
	for _, row := range backupRows {
		plan, parseErr := parseBackupPlan(row)
		if parseErr != nil {
			return invalidArtifact("backup_registry", parseErr)
		}
		state, exists := c.statePlans[plan.ProfileID]
		if !exists || state.PhysicalStateBindingSHA256 != plan.PhysicalStateBindingSHA256 {
			return unavailableBinding(plan.ProfileID, "backup_state_binding_mismatch")
		}
		c.backupPlans[plan.ProfileID] = plan
	}

	participantRegistry, participantArtifact, err := readArtifactObject(source, generatedExtensionsRoot+"participant-registry.json")
	if err != nil {
		return err
	}
	if participantRegistry["schema_id"] != "cartulary.extension_participant_registry.v1" || supportDigests["participant-registry"] != participantArtifact.SHA256 {
		return invalidArtifact("participant_registry", errors.New("integrity digest mismatch"))
	}
	c.participantContracts = map[string]ParticipantContract{}
	participantRows, ok := objectSlice(participantRegistry["participants"])
	if !ok {
		return invalidArtifact("participant_registry", errors.New("participants must be an array"))
	}
	for _, row := range participantRows {
		contract, parseErr := parseParticipantContract(row)
		if parseErr != nil {
			return invalidArtifact("participant_registry", parseErr)
		}
		if _, duplicate := c.participantContracts[contract.ParticipantID]; duplicate {
			return collisionFailure("participant_id", contract.ParticipantID)
		}
		if _, recognized := c.profiles[contract.OwnerProfileID]; !recognized {
			return unavailableBinding(contract.OwnerProfileID, "participant_owner_unrecognized")
		}
		c.participantContracts[contract.ParticipantID] = contract
	}
	return nil
}

func parseStatePlan(row map[string]any) (StatePlan, error) {
	plan := StatePlan{
		ProfileID:                      stringValue(row["profile_id"]),
		ContractMajor:                  intValue(row["contract_major"]),
		MigrationLineageID:             stringValue(row["migration_lineage_id"]),
		CurrentStateVersion:            intValue(row["current_state_version"]),
		MinimumMigratableStateVersion:  intValue(row["minimum_migratable_state_version"]),
		EmptyStatePolicy:               stringValue(row["empty_state_policy"]),
		DatabaseFamilyIDs:              catalogStrings(row["database_family_ids"]),
		ObjectReferenceFamilyIDs:       catalogStrings(row["object_reference_family_ids"]),
		InitializationKind:             stringValue(row["initialization_kind"]),
		InitializationDefinitionSHA256: stringValue(row["initialization_definition_sha256"]),
		FinalValidationAlgorithmID:     stringValue(row["final_state_validation_algorithm_id"]),
		PhysicalStateBindingSHA256:     stringValue(row["physical_state_binding_sha256"]),
		StatePresenceManifestSHA256:    stringValue(row["state_presence_manifest_sha256"]),
	}
	if plan.ProfileID == "" || plan.ContractMajor < 1 || plan.CurrentStateVersion < 1 || plan.MinimumMigratableStateVersion < 1 || plan.MinimumMigratableStateVersion > plan.CurrentStateVersion || (plan.EmptyStatePolicy != "allowed" && plan.EmptyStatePolicy != "forbidden") || plan.MigrationLineageID == "" || plan.FinalValidationAlgorithmID == "" {
		return StatePlan{}, fmt.Errorf("incomplete state plan %s", plan.ProfileID)
	}
	if !sort.StringsAreSorted(plan.DatabaseFamilyIDs) || !sort.StringsAreSorted(plan.ObjectReferenceFamilyIDs) {
		return StatePlan{}, fmt.Errorf("state families are not sorted for %s", plan.ProfileID)
	}
	migrations, ok := objectSlice(row["migration_definitions"])
	if !ok {
		return StatePlan{}, fmt.Errorf("migration definitions are not an array for %s", plan.ProfileID)
	}
	for _, migration := range migrations {
		plan.MigrationDefinitions = append(plan.MigrationDefinitions, MigrationDefinition{
			MigrationID:      stringValue(migration["migration_id"]),
			FromVersion:      intValue(migration["from_state_version"]),
			ToVersion:        intValue(migration["to_state_version"]),
			DefinitionSHA256: stringValue(migration["migration_definition_sha256"]),
		})
	}
	return plan, nil
}

func parseBackupPlan(row map[string]any) (BackupPlan, error) {
	plan := BackupPlan{ProfileID: stringValue(row["profile_id"]), PhysicalStateBindingSHA256: stringValue(row["physical_state_binding_sha256"]), Codecs: map[string]BackupCodec{}}
	physical, ok := row["physical_state_binding"].(map[string]any)
	if !ok || stringValue(physical["profile_id"]) != plan.ProfileID {
		return BackupPlan{}, errors.New("backup physical binding owner mismatch")
	}
	bindings, ok := objectSlice(physical["bindings"])
	if !ok || len(bindings) == 0 {
		return BackupPlan{}, errors.New("backup bindings must be nonempty")
	}
	previousGroup := 0
	for _, binding := range bindings {
		row := BackupBinding{
			BindingID: stringValue(binding["binding_id"]), LogicalFamilyID: stringValue(binding["logical_family_id"]), StorageKind: stringValue(binding["storage_kind"]), PhysicalRef: stringValue(binding["physical_ref"]), StateClass: stringValue(binding["state_class"]), BackupInclusion: stringValue(binding["backup_inclusion"]), RestoreOrderGroup: intValue(binding["restore_order_group"]), BackupCodecID: stringValue(binding["backup_codec_id"]), BackupCodecSHA256: stringValue(binding["backup_codec_sha256"]), PostRestoreValidationAlgorithmID: stringValue(binding["post_restore_validation_algorithm_id"]), RebuildAlgorithmID: stringValue(binding["rebuild_algorithm_id"]),
		}
		if row.BindingID == "" || row.RestoreOrderGroup < previousGroup || row.BackupCodecID == "" || row.BackupCodecSHA256 == "" {
			return BackupPlan{}, errors.New("invalid backup binding order or identity")
		}
		previousGroup = row.RestoreOrderGroup
		plan.Bindings = append(plan.Bindings, row)
	}
	codecs, ok := objectSlice(row["codecs"])
	if !ok {
		return BackupPlan{}, errors.New("backup codecs must be an array")
	}
	for _, encoded := range codecs {
		body, ok := encoded["codec"].(map[string]any)
		if !ok {
			return BackupPlan{}, errors.New("backup codec body missing")
		}
		codec := BackupCodec{CodecID: stringValue(encoded["backup_codec_id"]), SHA256: stringValue(encoded["backup_codec_sha256"]), BindingID: stringValue(body["binding_id"]), StorageKind: stringValue(body["storage_kind"]), MaxItems: intValue(body["max_items"]), MaxEntryBytes: int64Value(body["max_entry_bytes"]), MaxBindingBytes: int64Value(body["max_binding_bytes"]), HistoricalCodecDigests: catalogStrings(body["historical_restore_codecs"])}
		if codec.CodecID == "" || codec.SHA256 == "" || codec.BindingID == "" {
			return BackupPlan{}, errors.New("backup codec is incomplete")
		}
		plan.Codecs[codec.CodecID] = codec
	}
	return plan, nil
}

func parseParticipantContract(row map[string]any) (ParticipantContract, error) {
	body, ok := row["contract"].(map[string]any)
	if !ok {
		return ParticipantContract{}, errors.New("participant contract body missing")
	}
	contract := ParticipantContract{ParticipantID: stringValue(row["participant_id"]), OwnerProfileID: stringValue(row["profile_id"]), ContractSHA256: stringValue(row["participant_contract_sha256"]), ContractKind: stringValue(body["schema_id"]), InputSchemaID: stringValue(body["participant_input_schema_id"]), AlgorithmIDs: catalogStrings(body["algorithm_ids"]), SerializationKeyKinds: catalogStrings(body["serialization_key_kinds"]), OwnedStateFamilyIDs: catalogStrings(body["owned_state_family_ids"])}
	if contract.ContractKind == "cartulary.extension_transaction_participant_contract.v1" {
		contract.AlgorithmIDs = []string{stringValue(body["prepare_algorithm_id"]), stringValue(body["validation_algorithm_id"]), stringValue(body["write_algorithm_id"])}
		sort.Strings(contract.AlgorithmIDs)
	}
	if contract.ParticipantID == "" || contract.OwnerProfileID == "" || contract.ContractSHA256 == "" {
		return ParticipantContract{}, errors.New("participant contract is incomplete")
	}
	return contract, nil
}

func intValue(value any) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return int(parsed)
		}
	default:
		return 0
	}
	return 0
}

func int64Value(value any) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return parsed
		}
	default:
		return 0
	}
	return 0
}

func catalogStrings(value any) []string {
	values, ok := stringSlice(value)
	if !ok {
		return nil
	}
	return values
}
