package extensions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/JochiRaider/cartulary/internal/platform/extensioncodec"
)

type StatePlan struct {
	ProfileID                               string
	ContractMajor                           int
	MigrationLineageID                      string
	CurrentStateVersion                     int
	MinimumMigratableStateVersion           int
	EmptyStatePolicy                        string
	DatabaseFamilyIDs                       []string
	ObjectReferenceFamilyIDs                []string
	InitializationKind                      string
	InitializationDefinitionSHA256          string
	InitializationAlgorithmID               string
	InitializationAlgorithmDefinitionSHA256 string
	FinalValidationAlgorithmID              string
	PhysicalStateBindingSHA256              string
	StatePresenceManifestSHA256             string
	ImplementationBindingSHA256             string
	MigrationDefinitions                    []MigrationDefinition
	MigrationLedgerDefinitions              []MigrationLedgerDefinition
}

type MigrationDefinition struct {
	MigrationLineageID             string
	MigrationID                    string
	FromVersion                    int
	ToVersion                      int
	DefinitionSHA256               string
	ApplyAlgorithmID               string
	ValidationAlgorithmID          string
	ImplementationBindingProfileID string
	ImplementationBindingSHA256    string
}

// MigrationLedgerDefinition authenticates a committed ledger row without
// making the historical migration executable.
type MigrationLedgerDefinition struct {
	MigrationLineageID string
	MigrationID        string
	FromVersion        int
	ToVersion          int
	DefinitionSHA256   string
}

func cloneStatePlan(plan StatePlan) StatePlan {
	plan.DatabaseFamilyIDs = append([]string(nil), plan.DatabaseFamilyIDs...)
	plan.ObjectReferenceFamilyIDs = append([]string(nil), plan.ObjectReferenceFamilyIDs...)
	plan.MigrationDefinitions = append([]MigrationDefinition(nil), plan.MigrationDefinitions...)
	plan.MigrationLedgerDefinitions = append([]MigrationLedgerDefinition(nil), plan.MigrationLedgerDefinitions...)
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
	CodecID                    string
	SHA256                     string
	BindingID                  string
	StorageKind                string
	CodecContractRef           string
	LogicalIdentityAlgorithmID string
	ContentEncodingAlgorithmID string
	MaxItems                   int
	MaxEntryBytes              int64
	MaxBindingBytes            int64
	Postgres                   *extensioncodec.Postgres
}

type BackupPlan struct {
	ProfileID                   string
	PhysicalStateBindingSHA256  string
	ImplementationBindingSHA256 string
	Bindings                    []BackupBinding
	Codecs                      map[string]BackupCodec
}

func cloneBackupPlan(plan BackupPlan) BackupPlan {
	plan.Bindings = append([]BackupBinding(nil), plan.Bindings...)
	plan.Codecs = cloneBackupCodecs(plan.Codecs)
	return plan
}

func cloneBackupCodecs(source map[string]BackupCodec) map[string]BackupCodec {
	result := make(map[string]BackupCodec, len(source))
	for key, codec := range source {
		result[key] = codec
	}
	return result
}

type ReportingSourceContract struct {
	ProjectionSchemaID  string
	LabelSchemaID       string
	ReadAlgorithmID     string
	OrderingAlgorithmID string
	MaxVertices         int
	MaxEdges            int
	MaxLabelBytes       int64
}

type ParticipantContract struct {
	ParticipantID         string
	OwnerProfileID        string
	ContractSHA256        string
	ContractKind          string
	ParticipantKind       string
	InputSchemaID         string
	PrepareAlgorithmID    string
	ValidationAlgorithmID string
	WriteAlgorithmID      string
	AlgorithmIDs          []string
	SerializationKeyKinds []string
	OwnedStateFamilyIDs   []string
	ReportingSource       ReportingSourceContract
	Operations            []ParticipantOperation
}

type ParticipantOperation struct {
	OperationKind       string
	ResultSchemaID      string
	AlgorithmID         string
	OutputSchemaID      string
	OrderingAlgorithmID string
	StateFamilyIDs      []string
	MaxInputBytes       int64
	MaxOutputBytes      int64
	MaxItems            int
}

const (
	PortabilityNoAuthoritativeState = "no_authoritative_incident_state"
	PortabilityParticipant          = "participant"
	PortabilityBlockedWhenPresent   = "blocked_when_present"
)

// PortabilityPolicy is the immutable, transport-neutral projection consumed by
// the Incident Bundles owner. It contains declarations and admitted contract
// identities only; executable participants and physical state bindings are
// supplied by application composition.
type PortabilityPolicy struct {
	ProfileID              string
	Claimable              bool
	ContractMajor          int
	Mode                   string
	ParticipantID          string
	ParticipantSHA256      string
	ParticipantSchemaID    string
	MaximumInputBytes      int64
	MaximumOutputBytes     int64
	AuthoritativeFamilyIDs []string
	BlockingFamilyIDs      []string
}

func clonePortabilityPolicy(policy PortabilityPolicy) PortabilityPolicy {
	policy.AuthoritativeFamilyIDs = append([]string(nil), policy.AuthoritativeFamilyIDs...)
	policy.BlockingFamilyIDs = append([]string(nil), policy.BlockingFamilyIDs...)
	return policy
}

func cloneParticipantContract(contract ParticipantContract) ParticipantContract {
	contract.AlgorithmIDs = append([]string(nil), contract.AlgorithmIDs...)
	contract.SerializationKeyKinds = append([]string(nil), contract.SerializationKeyKinds...)
	contract.OwnedStateFamilyIDs = append([]string(nil), contract.OwnedStateFamilyIDs...)
	contract.Operations = append([]ParticipantOperation(nil), contract.Operations...)
	for index := range contract.Operations {
		contract.Operations[index].StateFamilyIDs = append([]string(nil), contract.Operations[index].StateFamilyIDs...)
	}
	return contract
}

func (c *Coordinator) StatePlan(profileID string) (StatePlan, bool) {
	if c == nil {
		return StatePlan{}, false
	}
	if c.profiles[profileID].bindingSHA256 == "" {
		return StatePlan{}, false
	}
	plan, ok := c.statePlans[profileID]
	return cloneStatePlan(plan), ok
}

// BackupPlans returns the immutable backup/restore catalog in profile order.
// Application composition translates this projection into Recovery's physical
// binding view; Recovery never imports the broad Extensions facade.
func (c *Coordinator) BackupPlans() []BackupPlan {
	if c == nil {
		return nil
	}
	result := make([]BackupPlan, 0, len(c.backupPlans))
	for _, profileID := range c.orderedProfileIDs {
		if plan, present := c.backupPlans[profileID]; present {
			result = append(result, cloneBackupPlan(plan))
		}
	}
	return result
}

// ParticipantContracts returns the immutable, owner-admitted participant
// catalog in participant-ID order. Application composition translates these
// facts into an owner-local protocol catalog; profile and shared-protocol
// packages never need the broad Extensions coordinator.
func (c *Coordinator) ParticipantContracts() []ParticipantContract {
	if c == nil {
		return nil
	}
	result := make([]ParticipantContract, 0, len(c.participantContracts))
	for _, contract := range c.participantContracts {
		result = append(result, cloneParticipantContract(contract))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ParticipantID < result[j].ParticipantID
	})
	return result
}

// PortabilityPolicies returns the owner-admitted policy catalog in profile-ID
// order. The result can be passed across application composition without
// exposing registry objects or canonical artifact bytes.
func (c *Coordinator) PortabilityPolicies() []PortabilityPolicy {
	if c == nil {
		return nil
	}
	result := make([]PortabilityPolicy, len(c.portabilityPolicies))
	for index, policy := range c.portabilityPolicies {
		result[index] = clonePortabilityPolicy(policy)
	}
	return result
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
			return registryInvalid("duplicate_state_profile_projection")
		}
		record, recognized := c.profiles[plan.ProfileID]
		if !recognized || record.descriptor.ContractMajor != plan.ContractMajor {
			return unavailableBinding(plan.ProfileID, record.descriptor.ContractMajor)
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
			return unavailableBinding(plan.ProfileID, c.profiles[plan.ProfileID].descriptor.ContractMajor)
		}
		record, exists := c.profiles[plan.ProfileID]
		if !exists {
			return unavailableBinding(plan.ProfileID, 0)
		}
		plan.ImplementationBindingSHA256 = record.bindingSHA256
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
			return registryInvalid("duplicate_participant_projection")
		}
		if _, recognized := c.profiles[contract.OwnerProfileID]; !recognized {
			return unavailableBinding(contract.OwnerProfileID, 0)
		}
		c.participantContracts[contract.ParticipantID] = contract
	}
	if err := c.admitPortabilityPolicies(source, supportDigests); err != nil {
		return err
	}
	return nil
}

func (c *Coordinator) admitPortabilityPolicies(source ArtifactSource, supportDigests map[string]string) error {
	c.portabilityPolicies = make([]PortabilityPolicy, 0, len(c.orderedProfileIDs))
	for _, profileID := range c.orderedProfileIDs {
		record := c.profiles[profileID]
		mode := stringValue(record.descriptorObject["incident_portability_mode"])
		policy := PortabilityPolicy{
			ProfileID:     profileID,
			Claimable:     record.descriptor.Claimable,
			ContractMajor: record.descriptor.ContractMajor,
			Mode:          mode,
		}
		if state, ok := c.statePlans[profileID]; ok {
			policy.AuthoritativeFamilyIDs = append([]string(nil), state.DatabaseFamilyIDs...)
			policy.AuthoritativeFamilyIDs = append(policy.AuthoritativeFamilyIDs, state.ObjectReferenceFamilyIDs...)
			sort.Strings(policy.AuthoritativeFamilyIDs)
		}
		switch mode {
		case PortabilityNoAuthoritativeState:
		case PortabilityParticipant:
			contributions, _ := objectSlice(record.descriptorObject["contributions"])
			for _, contribution := range contributions {
				if stringValue(contribution["kind"]) != "incident_portability_participant" {
					continue
				}
				if policy.ParticipantID != "" {
					return registryInvalid("duplicate_portability_policy_projection")
				}
				policy.ParticipantID = stringValue(contribution["participant_id"])
				policy.ParticipantSHA256 = stringValue(contribution["participant_contract_sha256"])
			}
			participant, ok := c.participantContracts[policy.ParticipantID]
			if !ok || participant.OwnerProfileID != profileID || participant.ContractSHA256 != policy.ParticipantSHA256 {
				return unavailableBinding(profileID, c.profiles[profileID].descriptor.ContractMajor)
			}
			policy.ParticipantSchemaID = participant.ContractKind
			policy.MaximumInputBytes = 64 * 1024 * 1024
			policy.MaximumOutputBytes = 64 * 1024 * 1024
		case PortabilityBlockedWhenPresent:
			path := "contracts/extensions/profiles/" + profileID + "/portability-blocking-predicate.json"
			predicate, artifact, err := readArtifactObject(source, path)
			if err != nil {
				return unavailableBinding(profileID, c.profiles[profileID].descriptor.ContractMajor)
			}
			digestID := "profiles." + profileID + ".portability-blocking-predicate"
			if supportDigests[digestID] != artifact.SHA256 ||
				stringValue(predicate["schema_id"]) != "cartulary.extension_state_blocking_predicate.v1" ||
				stringValue(predicate["kind"]) != "any_authoritative_state_present" {
				return unavailableBinding(profileID, c.profiles[profileID].descriptor.ContractMajor)
			}
			policy.BlockingFamilyIDs = catalogStrings(predicate["family_ids"])
			if len(policy.BlockingFamilyIDs) == 0 || !sort.StringsAreSorted(policy.BlockingFamilyIDs) ||
				!isSubset(policy.BlockingFamilyIDs, policy.AuthoritativeFamilyIDs) {
				return unavailableBinding(profileID, c.profiles[profileID].descriptor.ContractMajor)
			}
		default:
			return unavailableBinding(profileID, c.profiles[profileID].descriptor.ContractMajor)
		}
		c.portabilityPolicies = append(c.portabilityPolicies, policy)
	}
	return nil
}

func isSubset(subset, superset []string) bool {
	allowed := make(map[string]struct{}, len(superset))
	for _, value := range superset {
		allowed[value] = struct{}{}
	}
	previous := ""
	for _, value := range subset {
		if value == "" || value == previous {
			return false
		}
		if _, ok := allowed[value]; !ok {
			return false
		}
		previous = value
	}
	return true
}

func parseStatePlan(row map[string]any) (StatePlan, error) {
	plan := StatePlan{
		ProfileID:                               stringValue(row["profile_id"]),
		ContractMajor:                           intValue(row["contract_major"]),
		MigrationLineageID:                      stringValue(row["migration_lineage_id"]),
		CurrentStateVersion:                     intValue(row["current_state_version"]),
		MinimumMigratableStateVersion:           intValue(row["minimum_migratable_state_version"]),
		EmptyStatePolicy:                        stringValue(row["empty_state_policy"]),
		DatabaseFamilyIDs:                       catalogStrings(row["database_family_ids"]),
		ObjectReferenceFamilyIDs:                catalogStrings(row["object_reference_family_ids"]),
		InitializationKind:                      stringValue(row["initialization_kind"]),
		InitializationDefinitionSHA256:          stringValue(row["initialization_definition_sha256"]),
		InitializationAlgorithmID:               stringValue(row["initialization_algorithm_id"]),
		InitializationAlgorithmDefinitionSHA256: stringValue(row["initialization_algorithm_definition_sha256"]),
		FinalValidationAlgorithmID:              stringValue(row["final_state_validation_algorithm_id"]),
		PhysicalStateBindingSHA256:              stringValue(row["physical_state_binding_sha256"]),
		StatePresenceManifestSHA256:             stringValue(row["state_presence_manifest_sha256"]),
		ImplementationBindingSHA256:             stringValue(row["implementation_binding_sha256"]),
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
			MigrationLineageID:             stringValue(migration["migration_lineage_id"]),
			MigrationID:                    stringValue(migration["migration_id"]),
			FromVersion:                    intValue(migration["from_state_version"]),
			ToVersion:                      intValue(migration["to_state_version"]),
			DefinitionSHA256:               stringValue(migration["migration_definition_sha256"]),
			ApplyAlgorithmID:               stringValue(migration["apply_algorithm_id"]),
			ValidationAlgorithmID:          stringValue(migration["validation_algorithm_id"]),
			ImplementationBindingProfileID: stringValue(migration["implementation_binding_profile_id"]),
			ImplementationBindingSHA256:    stringValue(migration["implementation_binding_sha256"]),
		})
	}
	ledgerDefinitions, ok := objectSlice(row["migration_ledger_definitions"])
	if !ok {
		return StatePlan{}, fmt.Errorf("migration ledger definitions are not an array for %s", plan.ProfileID)
	}
	for _, definition := range ledgerDefinitions {
		plan.MigrationLedgerDefinitions = append(plan.MigrationLedgerDefinitions, MigrationLedgerDefinition{
			MigrationLineageID: stringValue(definition["migration_lineage_id"]),
			MigrationID:        stringValue(definition["migration_id"]),
			FromVersion:        intValue(definition["from_state_version"]),
			ToVersion:          intValue(definition["to_state_version"]),
			DefinitionSHA256:   stringValue(definition["migration_definition_sha256"]),
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
		if err := requireExactKeys(body, "schema_id", "backup_codec_id", "binding_id", "storage_kind", "codec_contract_ref", "logical_identity_algorithm_id", "content_encoding_algorithm_id", "max_items", "max_entry_bytes", "max_binding_bytes"); err != nil {
			return BackupPlan{}, errors.New("backup codec members are invalid")
		}
		if body["schema_id"] != "cartulary.extension_backup_binding_codec.v4" {
			return BackupPlan{}, errors.New("backup codec schema is not current")
		}
		canonical, err := canonicalJSON(body, true)
		if err != nil {
			return BackupPlan{}, errors.New("backup codec canonical bytes are invalid")
		}
		digest := sha256.Sum256(canonical)
		codec := BackupCodec{
			CodecID: stringValue(encoded["backup_codec_id"]), SHA256: stringValue(encoded["backup_codec_sha256"]),
			BindingID: stringValue(body["binding_id"]), StorageKind: stringValue(body["storage_kind"]),
			CodecContractRef:           stringValue(body["codec_contract_ref"]),
			LogicalIdentityAlgorithmID: stringValue(body["logical_identity_algorithm_id"]),
			ContentEncodingAlgorithmID: stringValue(body["content_encoding_algorithm_id"]),
			MaxItems:                   intValue(body["max_items"]), MaxEntryBytes: int64Value(body["max_entry_bytes"]),
			MaxBindingBytes: int64Value(body["max_binding_bytes"]),
		}
		if codec.CodecID == "" || codec.CodecID != stringValue(body["backup_codec_id"]) ||
			codec.SHA256 != hex.EncodeToString(digest[:]) || codec.BindingID == "" ||
			codec.CodecContractRef == "" || codec.LogicalIdentityAlgorithmID == "" || codec.ContentEncodingAlgorithmID == "" ||
			codec.MaxItems < 1 || codec.MaxItems > 1048576 || codec.MaxEntryBytes < 1 || codec.MaxBindingBytes < codec.MaxEntryBytes {
			return BackupPlan{}, errors.New("backup codec is incomplete")
		}
		if codec.StorageKind != "postgres" && codec.StorageKind != "object_store" && codec.StorageKind != "filesystem" {
			return BackupPlan{}, errors.New("backup codec storage kind is invalid")
		}
		if _, duplicate := plan.Codecs[codec.CodecID]; duplicate {
			return BackupPlan{}, errors.New("backup codec identity is duplicated")
		}
		plan.Codecs[codec.CodecID] = codec
	}
	for _, binding := range plan.Bindings {
		codec, present := plan.Codecs[binding.BackupCodecID]
		if !present || codec.SHA256 != binding.BackupCodecSHA256 || codec.BindingID != binding.BindingID || codec.StorageKind != binding.StorageKind {
			return BackupPlan{}, errors.New("backup binding codec identity mismatch")
		}
	}
	if err := resolvePostgresCodecs(&plan, row); err != nil {
		return BackupPlan{}, err
	}
	return plan, nil
}

func parseParticipantContract(row map[string]any) (ParticipantContract, error) {
	body, ok := row["contract"].(map[string]any)
	if !ok {
		return ParticipantContract{}, errors.New("participant contract body missing")
	}
	contract := ParticipantContract{
		ParticipantID:         stringValue(row["participant_id"]),
		OwnerProfileID:        stringValue(row["profile_id"]),
		ContractSHA256:        stringValue(row["participant_contract_sha256"]),
		ContractKind:          stringValue(body["schema_id"]),
		ParticipantKind:       stringValue(body["participant_kind"]),
		InputSchemaID:         stringValue(body["participant_input_schema_id"]),
		PrepareAlgorithmID:    stringValue(body["prepare_algorithm_id"]),
		ValidationAlgorithmID: stringValue(body["validation_algorithm_id"]),
		WriteAlgorithmID:      stringValue(body["write_algorithm_id"]),
		AlgorithmIDs:          catalogStrings(body["algorithm_ids"]),
		SerializationKeyKinds: catalogStrings(body["serialization_key_kinds"]),
		OwnedStateFamilyIDs:   catalogStrings(body["owned_state_family_ids"]),
	}
	if contract.ContractKind == "cartulary.extension_reporting_source_contract.v1" {
		contract.ReportingSource = ReportingSourceContract{
			ProjectionSchemaID:  stringValue(body["projection_schema_id"]),
			LabelSchemaID:       stringValue(body["label_schema_id"]),
			ReadAlgorithmID:     stringValue(body["read_algorithm_id"]),
			OrderingAlgorithmID: stringValue(body["ordering_algorithm_id"]),
			MaxVertices:         intValue(body["max_vertices"]), MaxEdges: intValue(body["max_edges"]),
			MaxLabelBytes: int64Value(body["max_label_bytes"]),
		}
		contract.OwnedStateFamilyIDs = catalogStrings(body["state_family_ids"])
		contract.AlgorithmIDs = []string{contract.ReportingSource.ReadAlgorithmID, contract.ReportingSource.OrderingAlgorithmID}
		sort.Strings(contract.AlgorithmIDs)
	}
	if contract.ContractKind == "cartulary.extension_transaction_participant_contract.v3" {
		contract.AlgorithmIDs = []string{contract.PrepareAlgorithmID, contract.ValidationAlgorithmID, contract.WriteAlgorithmID}
		sort.Strings(contract.AlgorithmIDs)
	}
	if contract.ContractKind == "cartulary.extension_participant_specialization.v5" {
		contract.InputSchemaID = stringValue(body["shared_context_schema_id"])
		operations, ok := objectSlice(body["operations"])
		if !ok || len(operations) == 0 {
			return ParticipantContract{}, errors.New("participant specialization operations are missing")
		}
		algorithmSet := map[string]struct{}{}
		stateFamilySet := map[string]struct{}{}
		for _, operation := range operations {
			parsed := ParticipantOperation{
				OperationKind:       stringValue(operation["operation_kind"]),
				ResultSchemaID:      stringValue(operation["result_schema_id"]),
				AlgorithmID:         stringValue(operation["algorithm_id"]),
				OutputSchemaID:      stringValue(operation["output_schema_id"]),
				OrderingAlgorithmID: stringValue(operation["ordering_algorithm_id"]),
				StateFamilyIDs:      catalogStrings(operation["state_family_ids"]),
				MaxInputBytes:       int64Value(operation["max_input_bytes"]),
				MaxOutputBytes:      int64Value(operation["max_output_bytes"]),
				MaxItems:            intValue(operation["max_items"]),
			}
			if parsed.OperationKind == "" || parsed.ResultSchemaID == "" || parsed.AlgorithmID == "" ||
				parsed.OutputSchemaID == "" || parsed.OrderingAlgorithmID == "" ||
				parsed.MaxInputBytes < 1 || parsed.MaxOutputBytes < 0 || parsed.MaxItems < 0 {
				return ParticipantContract{}, errors.New("participant specialization operation is incomplete")
			}
			algorithmSet[parsed.AlgorithmID] = struct{}{}
			algorithmSet[parsed.OrderingAlgorithmID] = struct{}{}
			for _, stateFamilyID := range parsed.StateFamilyIDs {
				stateFamilySet[stateFamilyID] = struct{}{}
			}
			contract.Operations = append(contract.Operations, parsed)
		}
		contract.AlgorithmIDs = make([]string, 0, len(algorithmSet))
		for algorithmID := range algorithmSet {
			contract.AlgorithmIDs = append(contract.AlgorithmIDs, algorithmID)
		}
		sort.Strings(contract.AlgorithmIDs)
		contract.OwnedStateFamilyIDs = make([]string, 0, len(stateFamilySet))
		for stateFamilyID := range stateFamilySet {
			contract.OwnedStateFamilyIDs = append(contract.OwnedStateFamilyIDs, stateFamilyID)
		}
		sort.Strings(contract.OwnedStateFamilyIDs)
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
